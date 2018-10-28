package gitils

import (
	"log"
	"net/http"

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

func serve() error {
	router := AllRoutes()

	err := chi.Walk(router, func(method string, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route)
		return nil
	})
	if err != nil {
		return err
	}

	defer shutdown()
	log.Println("GiTils listening on", config.Address)
	return http.ListenAndServe(config.Address, router)
}

func shutdown() {
	if db != nil {
		db.Close()
	}
}

func Start(conf Config) error {
	config = conf
	log.Println("Using Config", config)

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

func setupDatabase() error {
	log.Print("Preparing Database")
	opt := &orm.CreateTableOptions{
		IfNotExists: true,
	}

	orm.SetTableNameInflector(func(s string) string {
		return "gitils." + inflection.Plural(s)
	})

	err := db.CreateTable(new(StoredLyrics), opt)
	return err
}
