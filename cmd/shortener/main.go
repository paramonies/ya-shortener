package main

import (
	"log"
	"net/http"

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

	r, err := config.NewRepository(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	log.Printf("starting server on %s...\n", cfg.SrvAddr)
	log.Fatal(http.ListenAndServe(cfg.SrvAddr, NewRouter(r, &cfg)))
}

func NewRouter(db store.Repository, cfg *config.Config) *chi.Mux {
	log.Println("creating new chi-router")
	r := chi.NewRouter()

	r.Use(middleware.GzipDECompressHandler, middleware.GzipCompressHandler)
	r.Use(middleware.CookieMiddleware)

	r.Post("/", handlers.CreateShortURL(db, cfg.BaseURL))
	r.Post("/api/shorten", handlers.CreateShortURLFromJSON(db, cfg.BaseURL))
	r.Post("/api/shorten/batch", handlers.CreateManyShortURL(db, cfg.BaseURL))
	r.Get("/{ID}", handlers.GetURLByID(db))
	r.Get("/api/user/urls", handlers.GetListByUserID(db, cfg.BaseURL))
	r.Delete("/api/user/urls", handlers.DeleteManyShortURL(db))
	r.Get("/ping", handlers.Ping(db))
	return r
}
