package gitils

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/jinzhu/inflection"
)

var config Config
var db *pg.DB

func AllRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Logger,
		middleware.DefaultCompress,
		middleware.RedirectSlashes,
		middleware.Recoverer,
	)

	router.Mount("/", LyricsRoutes())
	return router
}

func setupDatabase() error {
	log.Info("Preparing Database")
	orm.SetTableNameInflector(func(s string) string {
		return "gitils." + inflection.Plural(s)
	})

	opt := &orm.CreateTableOptions{
		IfNotExists: true,
	}

	log.Trace("Creating lyrics table")
	err := db.CreateTable(new(StoredLyrics), opt)

	if err == nil {
		_, err = db.Exec("CREATE INDEX IF NOT EXISTS query_search ON gitils.stored_lyrics USING GIN (to_tsvector('english', query))")
	}

	return err
}

func serve() error {
	router := AllRoutes()

	log.Debug("Routes:")
	err := chi.Walk(router, func(method string, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		log.Debugf("%s %q", method, route)
		return nil
	})
	if err != nil {
		return err
	}

	defer shutdown()
	log.Infof("GiTils listening on %q", config.Address)
	return http.ListenAndServe(config.Address, router)
}

func shutdown() {
	if db != nil {
		db.Close()
	}
}

func Start(conf Config) error {
	config = conf
	log.Tracef("Using Config: %+v", config)

	options, err := pg.ParseURL(config.PostgresURL)
	if err != nil {
		return err
	}
	db = pg.Connect(options)

	if err := setupDatabase(); err != nil {
		return err
	}

	return serve()
}
