package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/acme/autocert"

	"github.com/paramonies/internal/config"
	"github.com/paramonies/internal/handlers"
	"github.com/paramonies/internal/routes"
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
	h := handlers.New(r, cfg.BaseURL, cfg.TrustedSubnet)

	//HTTP Server
	server := &http.Server{
		Addr:    cfg.SrvAddr,
		Handler: routes.New(h),
	}
	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-sigint
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("starting server on %s...\n", cfg.SrvAddr)

	if *cfg.EnableHTTPS {
		manager := &autocert.Manager{
			Cache:  autocert.DirCache("certs"),
			Prompt: autocert.AcceptTOS,
		}
		server.TLSConfig = manager.TLSConfig()

		log.Printf("HTTPs enable")
		if err := server.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			log.Fatalf("HTTPS server ListenAndServe: %v", err)
		}
	} else {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}

	// ждём завершения процедуры graceful shutdown
	<-idleConnsClosed
	fmt.Println("Server Shutdown gracefully")

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
