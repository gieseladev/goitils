package gitils

import (
	"errors"
	"net/http"
	"time"

	"github.com/gieseladev/lyricsfindergo/pkg"
	"github.com/gieseladev/lyricsfindergo/pkg/models"
	"github.com/go-chi/chi"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	log "github.com/sirupsen/logrus"
)

// TODO: use separate table for query -> url

type StoredLyrics struct {
	Url         string    `json:"url" sql:",pk"`
	Title       string    `json:"title"`
	Artist      string    `json:"artist"`
	Lyrics      string    `json:"lyrics"`
	ReleaseDate time.Time `json:"release_date"`
	OriginName  string    `json:"source_name"`
	OriginUrl   string    `json:"source_url"`
}

type LyricsQuery struct {
	Query     string `sql:",unique"`
	LyricsUrl string `sql:",notnull"`
	Lyrics    *StoredLyrics
}

func NewLyricsQuery(query string, lyrics *StoredLyrics) *LyricsQuery {
	return &LyricsQuery{query, lyrics.Url, lyrics}
}

func findStoredLyrics(query string) (*LyricsQuery, error) {
	lyrics := new(LyricsQuery)

	err := db.Model(lyrics).
		Where("query @@ plainto_tsquery(?)", query).
		Column("Lyrics").
		Limit(1).
		Select()

	if err != nil {
		if err != pg.ErrNoRows {
			log.Warningf("Couldn't find stored lyrics: %v", err)
		}
		return nil, err
	}

	return lyrics, nil
}

func searchLyrics(searchQuery string) (*StoredLyrics, error) {
	if searchQuery == "" {
		return nil, errors.New("no search query")
	}

	lyrics := lyricsfinder.SearchFirstLyrics(searchQuery, config.GoogleApiKey)

	if lyrics == (models.Lyrics{}) {
		return nil, errors.New("no lyrics found")
	}

	storedLyrics := new(StoredLyrics)

	storedLyrics.Url = lyrics.Url
	storedLyrics.Title = lyrics.Title
	storedLyrics.Artist = lyrics.Artist
	storedLyrics.Lyrics = lyrics.Lyrics

	storedLyrics.ReleaseDate = lyrics.ReleaseDate

	storedLyrics.OriginName = lyrics.Origin.Name
	storedLyrics.OriginUrl = lyrics.Origin.Url

	return storedLyrics, nil
}

func storeLyrics(lyrics *LyricsQuery) {
	if _, err := db.Model(lyrics.Lyrics).OnConflict("(url) DO NOTHING").Insert(); err != nil {
		log.Warning(err)
	}

	if err := db.Insert(lyrics); err != nil {
		log.Warning(err)
	}
}

func getLyrics(w http.ResponseWriter, r *http.Request) {
	searchQuery := chi.URLParam(r, "query")

	lyricsQuery, err := findStoredLyrics(searchQuery)
	if err != nil {
		log.Debugf("Searching lyrics for: %s", searchQuery)
		lyrics, err := searchLyrics(searchQuery)
		if err == nil {
			lyricsQuery = NewLyricsQuery(searchQuery, lyrics)
			defer storeLyrics(lyricsQuery)
		}
	}

	if lyricsQuery == nil {
		log.Warningf("No lyrics found: %v for query: %q", err, searchQuery)
		SendErrorResponse(w, r, 404, "No lyrics found", nil)
		return
	}

	SendJsonResponse(w, r, lyricsQuery.Lyrics)
}

func CreateLyricsTable() error {
	opt := &orm.CreateTableOptions{
		IfNotExists:   true,
		FKConstraints: true,
	}

	log.Trace("Creating lyrics table")

	if err := db.CreateTable(new(StoredLyrics), opt); err != nil {
		return err
	}

	if err := db.CreateTable(new(LyricsQuery), opt); err != nil {
		return err
	}

	_, err := db.Exec("CREATE INDEX IF NOT EXISTS query_search ON gitils.lyrics_queries USING GIN (to_tsvector('english', query))")

	return err
}

func LyricsRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/lyrics/{query}", getLyrics)

	return router
}
