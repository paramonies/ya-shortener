package main

import (
	"github.com/paramonies/internal/config"
	"github.com/paramonies/internal/handlers"
	"github.com/paramonies/internal/routes"
	"log"
	"net/http"
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
	h := handlers.New(r, cfg.BaseURL)

	log.Printf("starting server on %s...\n", cfg.SrvAddr)
	log.Fatal(http.ListenAndServe(cfg.SrvAddr, routes.New(h)))
}
