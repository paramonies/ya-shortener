package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/paramonies/internal/handlers"
	"github.com/paramonies/internal/middleware"
	"github.com/paramonies/internal/store"
	"log"
	"net/http"
)

type Config struct {
	SrvAddr       string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStorePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDNS   string `env:"DATABASE_DSN" envDefault:"postgresql://postgres:123456@localhost/shortener-api?connect_timeout=10&sslmode=disable"`
}

var cfg Config

func init() {
	flag.StringVar(&cfg.SrvAddr, "a", cfg.SrvAddr, "server host and port")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "URL for making http request")
	flag.StringVar(&cfg.FileStorePath, "f", cfg.FileStorePath, "path to DB-file on disk")
	flag.StringVar(&cfg.DatabaseDNS, "d", cfg.DatabaseDNS, "database dns")
}

func main() {
	log.Printf("starting")
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.Parse()

	var db store.Repository
	if cfg.DatabaseDNS != "" {
		db, err = store.NewPostgresDB(cfg.DatabaseDNS)
		if err != nil {
			log.Fatal(err)
		}
	} else if cfg.FileStorePath == "" {
		db, err = store.NewFileDB(cfg.FileStorePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		db = store.NewMapDB()
	}
	defer db.Close()

	log.Printf("starting server on %s...\n", cfg.SrvAddr)
	log.Fatal(http.ListenAndServe(cfg.SrvAddr, NewRouter(db, &cfg)))
}

func NewRouter(db store.Repository, cfg *Config) *chi.Mux {
	r := chi.NewRouter()
	log.Println(0)

	r.Use(middleware.GzipDECompressHandler, middleware.GzipCompressHandler)
	r.Use(middleware.CookieMiddleware)

	r.Post("/", handlers.CreateShortURLHadler(db, cfg.BaseURL))
	r.Post("/api/shorten", handlers.CreateShortURLFromJSONHandler(db, cfg.BaseURL))
	r.Get("/{ID}", handlers.GetURLByIDHandler(db))
	r.Get("/api/user/urls", handlers.GetListByUserIDHandler(db, cfg.BaseURL))
	r.Get("/ping", handlers.PingHandler(db, cfg.DatabaseDNS))
	return r
}
