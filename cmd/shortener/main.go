package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"net/url"
)

var srvAddr = "localhost:8080"

func main() {
	log.Fatal(http.ListenAndServe(srvAddr, NewRouter()))
}

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	db := NewDB()
	r.Post("/", CreateShortURLHadler(db))
	r.Get("/{ID}", GetURLByIDHandler(db))
	return r
}

func CreateShortURLHadler(rep Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		urlStr := string(b)
		_, err = url.ParseRequestURI(urlStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id := hash(urlStr)
		rep.Set(fmt.Sprintf("%d", id), urlStr)

		w.Header().Set("Content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		shortURL := fmt.Sprintf("https://%s/%d", srvAddr, id)
		w.Write([]byte(shortURL))
	}
}

func GetURLByIDHandler(rep Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "ID")

		val, err := rep.Get(id)
		if err != nil {
			http.Error(w, "id not found", http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, val, http.StatusTemporaryRedirect)
		w.Write([]byte("ID found"))
	}
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

type DB struct {
	mapDB map[string]string
}

type Repository interface {
	Set(key, val string)
	Get(key string) (string, error)
}

func NewDB() *DB {
	return &DB{
		mapDB: make(map[string]string),
	}
}

func (db *DB) Set(key, val string) {
	db.mapDB[key] = val
}

func (db *DB) Get(key string) (string, error) {
	if _, ok := db.mapDB[key]; !ok {
		return "", fmt.Errorf("key %s not found in database", key)
	}
	return db.mapDB[key], nil
}
