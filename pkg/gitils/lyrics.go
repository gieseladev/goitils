package gitils

import (
	"encoding/json"
	"net/http"

	"github.com/gieseladev/lyricsfindergo/pkg"
	"github.com/gieseladev/lyricsfindergo/pkg/models"
)

func getLyrics(writer http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	searchQuery := query.Get("query")
	if searchQuery == "" {
		http.Error(writer, "Missing Search query", 400)
		return
	}
	lyrics := lyricsfinder.SearchFirstLyrics(searchQuery, config.GoogleApiKey)

	if lyrics == (models.Lyrics{}) {
		http.Error(writer, "No lyrics found", 404)
		return
	}

	writer.Header().Set("content-type", "application/json")
	err := json.NewEncoder(writer).Encode(lyrics)
	if err != nil {
		panic(err)
	}
}

func init() {
	http.HandleFunc("/lyrics/", getLyrics)
}
