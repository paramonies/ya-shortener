// Package routes difines routes for user request and pprof.
package routes

import (
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"

	"github.com/paramonies/internal/handlers"
	"github.com/paramonies/internal/middleware"
)

func New(h *handlers.Handler) *chi.Mux {
	log.Println("creating new chi-routes")
	r := chi.NewRouter()

	r.Use(middleware.GzipDECompressHandler, middleware.GzipCompressHandler)
	r.Use(middleware.CookieMiddleware)

	r.Post("/", h.CreateShortURL())
	r.Post("/api/shorten", h.CreateShortURLFromJSON())
	r.Post("/api/shorten/batch", h.CreateManyShortURL())
	r.Get("/{ID}", h.GetURLByID())
	r.Get("/api/user/urls", h.GetListByUserID())
	r.Delete("/api/user/urls", h.DeleteManyShortURL())
	r.Get("/ping", h.Ping())
	r.Get("/api/internal/stats", h.GetStats())

	r.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	r.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	r.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	r.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	r.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	r.Handle("/debug/pprof/{cmd}", http.HandlerFunc(pprof.Index))

	return r
}
