package main

import (
	"fmt"
	"github.com/paramonies/internal/config"
	"github.com/paramonies/internal/handlers"
	"github.com/paramonies/internal/routes"
	"log"
	"net/http"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()

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

func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}

	if buildDate == "" {
		buildDate = "N/A"
	}

	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
