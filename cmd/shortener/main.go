package main

import (
	"github.com/paramonies/internal/store"
	"log"
	"net/http"

	"github.com/paramonies/internal/config"
	"github.com/paramonies/internal/handlers"
	"github.com/paramonies/internal/routes"
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
			log.Fatal(err)
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
	h := handlers.New(db, cfg.BaseURL)

	log.Printf("starting server on %s...\n", cfg.SrvAddr)
	log.Fatal(http.ListenAndServe(cfg.SrvAddr, routes.New(h)))
}
