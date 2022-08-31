package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/paramonies/internal/config"
	"github.com/paramonies/internal/middleware"
)

func BenchmarkCreateShortURL(b *testing.B) {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := chi.NewRouter()
	cfg := config.Config{
		SrvAddr:       "localhost:8080",
		BaseURL:       "http://localhost:8080",
		TrustedSubnet: "192.168.0.1/24",
	}

	rep, err := config.NewRepository(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer rep.Close()
	h := New(rep, cfg.BaseURL, cfg.TrustedSubnet)

	userID, _ := middleware.GenerateToken(10)
	cookie := &http.Cookie{
		Name:   "user_id",
		Value:  userID,
		Secure: false,
	}

	b.ResetTimer() // reset all timers
	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop all timers
		st := "http://test_link_" + strconv.Itoa(i) + ".ru"
		r = strings.NewReader(st)
		request := httptest.NewRequest(http.MethodPost, "/", r)
		request.AddCookie(cookie)
		b.StartTimer() //
		rtr.HandleFunc("/", h.CreateShortURL())
		// запускаем сервер
		rtr.ServeHTTP(w, request)
		res := w.Result()
		b.StopTimer() // останавливаем таймер

		res.Body.Close()
	}
}

func BenchmarkHandler_GetURLByID(b *testing.B) {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := chi.NewRouter()
	cfg := config.Config{
		SrvAddr:       "localhost:8080",
		BaseURL:       "http://localhost:8080",
		TrustedSubnet: "192.168.0.1/24",
	}

	rep, err := config.NewRepository(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer rep.Close()
	h := New(rep, cfg.BaseURL, cfg.TrustedSubnet)

	userID, _ := middleware.GenerateToken(10)
	cookie := &http.Cookie{
		Name:   "user_id",
		Value:  userID,
		Secure: false,
	}

	// add url in repository
	st := "http://test_link.ru"
	r = strings.NewReader(st)
	request := httptest.NewRequest(http.MethodPost, "/", r)
	request.AddCookie(cookie)
	rtr.HandleFunc("/", h.CreateShortURL())
	rtr.ServeHTTP(w, request)
	res := w.Result()
	res.Body.Close()
	fmt.Println(rep.GetAllByID(userID))

	b.ResetTimer() // reset all timers

	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop all timers
		tg := "/1133590167"
		request := httptest.NewRequest(http.MethodGet, tg, r)

		b.StartTimer() //
		rtr.HandleFunc("/{ID}", h.GetURLByID())
		// запускаем сервер
		rtr.ServeHTTP(w, request)
		res := w.Result()

		b.StopTimer() // останавливаем таймер

		res.Body.Close()
	}
}

func BenchmarkHandler_GetListByUserID(b *testing.B) {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := chi.NewRouter()
	cfg := config.Config{
		SrvAddr:       "localhost:8080",
		BaseURL:       "http://localhost:8080",
		TrustedSubnet: "192.168.0.1/24",
	}

	rep, err := config.NewRepository(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer rep.Close()
	h := New(rep, cfg.BaseURL, cfg.TrustedSubnet)

	userID, _ := middleware.GenerateToken(10)
	cookie := &http.Cookie{
		Name:   "user_id",
		Value:  userID,
		Secure: false,
	}

	// add url in repository
	st := "http://test_link.ru"
	r = strings.NewReader(st)
	request := httptest.NewRequest(http.MethodPost, "/", r)
	request.AddCookie(cookie)
	rtr.HandleFunc("/", h.CreateShortURL())
	rtr.ServeHTTP(w, request)
	res := w.Result()
	res.Body.Close()
	fmt.Println(rep.GetAllByID(userID))

	b.ResetTimer() // reset all timers

	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop all timers

		request := httptest.NewRequest(http.MethodGet, "/api/user/urls", r)
		request.AddCookie(cookie)
		b.StartTimer() //
		rtr.HandleFunc("/api/user/urls", h.GetListByUserID())
		// запускаем сервер
		rtr.ServeHTTP(w, request)
		res := w.Result()

		res.Body.Close()
	}
}

func BenchmarkHandler_CreateShortURLFromJSON(b *testing.B) {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := chi.NewRouter()
	cfg := config.Config{
		SrvAddr:       "localhost:8080",
		BaseURL:       "http://localhost:8080",
		TrustedSubnet: "192.168.0.1/24",
	}

	rep, err := config.NewRepository(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer rep.Close()
	h := New(rep, cfg.BaseURL, cfg.TrustedSubnet)

	userID, _ := middleware.GenerateToken(10)
	cookie := &http.Cookie{
		Name:   "user_id",
		Value:  userID,
		Secure: false,
	}

	b.ResetTimer() // reset all timers
	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop all timers
		st := "{\"url\": \"http://test_link_" + strconv.Itoa(i) + ".ru\"}"
		r = strings.NewReader(st)
		request := httptest.NewRequest(http.MethodPost, "/api/shorten", r)
		request.AddCookie(cookie)
		b.StartTimer() //
		rtr.HandleFunc("/api/shorten", h.CreateShortURLFromJSON())
		// запускаем сервер
		rtr.ServeHTTP(w, request)
		res := w.Result()

		b.StopTimer() // останавливаем таймер

		res.Body.Close()
	}
}
