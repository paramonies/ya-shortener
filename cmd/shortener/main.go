package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/paramonies/internal/config"
	"github.com/paramonies/internal/handlers"
	"github.com/paramonies/internal/middleware"
	"github.com/paramonies/internal/store"
)

func main() {
	var cfg config.Config
	err := cfg.Init()
	if err != nil {
		log.Fatal(err)
	}

	var db store.Repository
	if cfg.DatabaseDSN != "" {
		db, err = store.NewPostgresDB(cfg.DatabaseDSN)
		if err != nil {
			os.Exit(1)
		}
	} else if cfg.FileStorePath != "" {
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

func NewRouter(db store.Repository, cfg *config.Config) *chi.Mux {
	log.Println("creating new chi-router")
	r := chi.NewRouter()

	r.Use(middleware.GzipDECompressHandler, middleware.GzipCompressHandler)
	r.Use(middleware.CookieMiddleware)

	r.Post("/", handlers.CreateShortURLHadler(db, cfg.BaseURL))
	r.Post("/api/shorten", handlers.CreateShortURLFromJSONHandler(db, cfg.BaseURL))
	r.Get("/{ID}", handlers.GetURLByIDHandler(db))
	r.Get("/api/user/urls", handlers.GetListByUserIDHandler(db, cfg.BaseURL))
	r.Get("/ping", handlers.PingHandler(db))
	r.Post("/api/shorten/batch", handlers.CreateManyShortURLHadler(db, cfg.BaseURL))
	r.Delete("/api/user/urls", handlers.DeleteManyShortURLHadler(db))
	return r
}
