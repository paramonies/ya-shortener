package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/paramonies/internal/handlers"
	"github.com/paramonies/internal/store"
	"log"
	"net/http"
)

type Config struct {
	SrvAddr       string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStorePath string `env:"FILE_STORAGE_PATH"`
}

var cfg Config

func init() {
	flag.StringVar(&cfg.SrvAddr, "a", cfg.SrvAddr, "server host and port")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "URL for making http request")
	flag.StringVar(&cfg.FileStorePath, "f", cfg.FileStorePath, "path to DB-file on disk")
}

func main() {
	log.Printf("starting")
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.Parse()

	var db store.Repository
	if cfg.FileStorePath == "" {
		db = store.NewMapDB()
	} else {
		db, err = store.NewFileDB(cfg.FileStorePath)
		if err != nil {
			fmt.Println("!!!")
			log.Fatal(err)
		}
	}
	defer db.Close()

	log.Printf("starting server on %s...\n", cfg.SrvAddr)
	log.Fatal(http.ListenAndServe(cfg.SrvAddr, NewRouter(db, &cfg)))
}

func NewRouter(db store.Repository, cfg *Config) *chi.Mux {
	r := chi.NewRouter()
	log.Println(0)
	r.Post("/", handlers.CreateShortURLHadler(db, cfg.BaseURL))
	r.Post("/api/shorten", handlers.CreateShortURLFromJSONHandler(db, cfg.BaseURL))
	r.Get("/{ID}", handlers.GetURLByIDHandler(db))
	return r
}
