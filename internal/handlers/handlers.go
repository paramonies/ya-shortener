package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/paramonies/internal/store"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"net/url"
)

func CreateShortURLHadler(rep store.Repository, baseURL string) http.HandlerFunc {
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

		id := Hash(urlStr)
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

func CreateShortURLFromJSONHandler(rep store.Repository, baseURL string) http.HandlerFunc {
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

		id := Hash(URL)
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

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write(resBody)
	}
}

func GetURLByIDHandler(rep store.Repository) http.HandlerFunc {
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

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
