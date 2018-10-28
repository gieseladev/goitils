package gitils

import (
	"errors"
	"net/http"
	"time"

	"github.com/gieseladev/lyricsfindergo/pkg"
	"github.com/gieseladev/lyricsfindergo/pkg/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type StoredLyrics struct {
	Query string
	Url   string `sql:",pk"`
	Title,
	Artist,
	Lyrics string
	ReleaseData time.Time
	OriginName,
	OriginUrl string
}

func findCachedLyrics(query string) (*StoredLyrics, error) {
	lyrics := new(StoredLyrics)
	err := db.Model(lyrics).
		Where("query LIKE ?", query).
		WhereOr("artist || '-' || title LIKE ?", query).
		Limit(1).
		Select()

	if err != nil {
		return nil, err
	} else if *lyrics == (StoredLyrics{}) {
		return nil, errors.New("no lyrics found in db")
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
	storedLyrics.Query = searchQuery

	storedLyrics.Url = lyrics.Url
	storedLyrics.Title = lyrics.Title
	storedLyrics.Artist = lyrics.Artist
	storedLyrics.Lyrics = lyrics.Lyrics

	storedLyrics.ReleaseData = lyrics.ReleaseDate

	storedLyrics.OriginName = lyrics.Origin.Name
	storedLyrics.OriginUrl = lyrics.Origin.Url

	return storedLyrics, nil
}

func getLyrics(w http.ResponseWriter, r *http.Request) {
	searchQuery := chi.URLParam(r, "query")

	lyrics, err := findCachedLyrics(searchQuery)
	if err != nil {
		lyrics, err = searchLyrics(searchQuery)
		defer db.Insert(lyrics)
	}

	if lyrics == nil || *lyrics == (StoredLyrics{}) {
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
