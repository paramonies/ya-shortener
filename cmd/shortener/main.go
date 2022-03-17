package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

//var srvAddr = "localhost:8080"
//var baseUrl = "http://" + srvAddr + "/"

type Config struct {
	SrvAddr       string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStorePath string `env:"FILE_STORAGE_PATH" envDefault:""`
}

var cfg Config

func init() {
	flag.StringVar(&cfg.SrvAddr, "a", cfg.SrvAddr, "server host and port")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "URL for making http request")
	flag.StringVar(&cfg.FileStorePath, "f", cfg.FileStorePath, "path to DB-file on disk")
}

func main() {
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.Parse()

	var db Repository
	if cfg.FileStorePath == "" {
		db = NewMapDB()
	} else {
		db, err = NewFileDB(cfg.FileStorePath)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer db.Close()

	log.Printf("starting server on %s...\n", cfg.SrvAddr)
	log.Fatal(http.ListenAndServe(cfg.SrvAddr, NewRouter(db, &cfg)))
}

func NewRouter(db Repository, cfg *Config) *chi.Mux {
	r := chi.NewRouter()
	log.Println(0)
	r.Post("/", CreateShortURLHadler(db, cfg.BaseURL))
	r.Post("/api/shorten", CreateShortURLFromJSONHandler(db, cfg.BaseURL))
	r.Get("/{ID}", GetURLByIDHandler(db))
	return r
}

func CreateShortURLHadler(rep Repository, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			log.Printf("read body %s...\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		urlStr := string(b)
		_, err = url.ParseRequestURI(urlStr)
		if err != nil {
			log.Printf("parse url %s...\n", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id := hash(urlStr)
		err = rep.Set(fmt.Sprintf("%d", id), urlStr)
		if err != nil {
			log.Printf("rep set %s...\n", err.Error())
			log.Println(err)
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		shortURL := fmt.Sprintf("%s/%d", baseURL, id)
		w.Write([]byte(shortURL))
	}
}

func CreateShortURLFromJSONHandler(rep Repository, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(1)
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		log.Println(2)
		if err != nil {
			log.Printf("read body %s...\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println(3)
		var reqBodyJSON struct {
			URL string `json:"url"`
		}
		log.Println(4)
		err = json.Unmarshal(b, &reqBodyJSON)
		if err != nil {
			log.Printf("unmarshall body %s...\n", err.Error())
			http.Error(w, "id not found", http.StatusBadRequest)
			return
		}
		URL := reqBodyJSON.URL
		log.Println(5)
		_, err = url.ParseRequestURI(URL)
		if err != nil {
			log.Printf("parse url %s...\n", err.Error())
			http.Error(w, "id not found", http.StatusBadRequest)
			return
		}

		id := hash(URL)
		rep.Set(fmt.Sprintf("%d", id), URL)

		shortURL := fmt.Sprintf("%s/%d", baseURL, id)

		resBodyJSON := struct {
			Result string `json:"result"`
		}{
			Result: shortURL,
		}

		resBody, err := json.Marshal(resBodyJSON)
		if err != nil {
			log.Printf("parse url %s...\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resBody)
	}
}

func GetURLByIDHandler(rep Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "ID")

		val, err := rep.Get(id)
		if err != nil {
			log.Println(err)
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

type Repository interface {
	Set(key, val string) error
	Get(key string) (string, error)
	Close() error
}

type MapDB struct {
	DB map[string]string
}

func NewMapDB() *MapDB {
	return &MapDB{
		DB: make(map[string]string),
	}
}

func (db *MapDB) Set(key, val string) error {
	db.DB[key] = val
	return nil
}

func (db *MapDB) Get(key string) (string, error) {
	val, ok := db.DB[key]
	if !ok {
		return "", fmt.Errorf("key %s not found in database", key)
	}
	return val, nil
}

func (db *MapDB) Close() error {
	return nil
}

type Record struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type Records struct {
	Records []Record `json:"records"`
}

type FileDB struct {
	DB    *os.File
	Cache Records
}

func NewFileDB(path string) (*FileDB, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(file)
	//defer file.Close()
	if err != nil {
		return nil, err
	}

	var records []Record
	err = json.Unmarshal(data, &records)
	if err != nil {
		return nil, err
	}

	return &FileDB{DB: file, Cache: Records{records}}, nil
}

func (f *FileDB) Set(key, value string) error {
	for _, r := range f.Cache.Records {
		if r.ID == key {
			return nil
		}
	}

	r := Record{ID: key, URL: value}
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	f.Cache.Records = append(f.Cache.Records, r)
	_, err = f.DB.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (f *FileDB) Get(key string) (string, error) {
	for _, r := range f.Cache.Records {
		if r.ID == key {
			return r.URL, nil
		}
	}

	return "", fmt.Errorf("key %s not found in database", key)
}

func (f *FileDB) Close() error {
	return f.DB.Close()
}
