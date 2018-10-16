package gitils

import (
	"log"
	"net/http"
)

var config Config

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request %s", r.URL.Path)
}

func serve() error {
	http.HandleFunc("/", indexHandler)

	log.Println("GiTils listening on", config.Address)
	return http.ListenAndServe(config.Address, nil)
}

func Start(conf Config) {
	config = conf
	log.Println("Using Config", config)
	log.Fatal(serve())
}
