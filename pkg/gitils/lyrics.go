package gitils

import (
	"errors"
	"net/http"
	"time"

	"github.com/gieseladev/lyricsfindergo/pkg"
	"github.com/gieseladev/lyricsfindergo/pkg/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	log "github.com/sirupsen/logrus"
)

type StoredLyrics struct {
	Query       string    `json:"-"`
	Url         string    `json:"url" sql:",pk"`
	Title       string    `json:"title"`
	Artist      string    `json:"artist"`
	Lyrics      string    `json:"lyrics"`
	ReleaseDate time.Time `json:"release_date"`
	OriginName  string    `json:"source_name"`
	OriginUrl   string    `json:"source_url"`
}

func findStoredLyrics(query string) (*StoredLyrics, error) {
	lyrics := new(StoredLyrics)

	err := db.Model(lyrics).
		Where("query @@ plainto_tsquery(?)", query).
		WhereOr("artist || ' ' || title @@ plainto_tsquery(?)", query).
		Limit(1).
		Select()

	if err != nil {
		log.Warningf("Couldn't find stored lyrics: %v", err)
		return nil, err
	} else if *lyrics == (StoredLyrics{}) {
		return nil, errors.New("no lyrics found in db")
	}

	return lyrics, nil
}

func storeLyrics(lyrics *StoredLyrics) {
	_, err := db.Model(lyrics).
		OnConflict("(url) DO NOTHING").
		Insert()
	if err != nil {
		log.Warning(err)
	}
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
	storedLyrics.Query = searchQuery

	storedLyrics.Url = lyrics.Url
	storedLyrics.Title = lyrics.Title
	storedLyrics.Artist = lyrics.Artist
	storedLyrics.Lyrics = lyrics.Lyrics

	storedLyrics.ReleaseDate = lyrics.ReleaseDate

	storedLyrics.OriginName = lyrics.Origin.Name
	storedLyrics.OriginUrl = lyrics.Origin.Url

	return storedLyrics, nil
}

func getLyrics(w http.ResponseWriter, r *http.Request) {
	searchQuery := chi.URLParam(r, "query")

	lyrics, err := findStoredLyrics(searchQuery)
	if err != nil {
		log.Debugf("Searching lyrics for: %s", searchQuery)
		lyrics, err = searchLyrics(searchQuery)
		defer storeLyrics(lyrics)
	}

	if lyrics == nil || *lyrics == (StoredLyrics{}) {
		log.Warningf("No lyrics found: %v for query: %q", err, searchQuery)
		http.Error(w, "No lyrics found", 404)
		return
	}

	render.JSON(w, r, lyrics)
}

func LyricsRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/lyrics/{query}", getLyrics)

	return router
}
