package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4"
	"github.com/paramonies/internal/store"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"net/url"
)

func CreateShortURLHadler(rep store.Repository, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("create short url from text/plain body")
		log.Printf("request url: %s %s", r.Method, r.URL)

		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("request body: %s", string(b))

		urlStr := string(b)
		log.Printf("original url: %s", urlStr)
		_, err = url.ParseRequestURI(urlStr)
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id := Hash(urlStr)
		log.Printf("hash for url: %d", id)
		cookie, err := r.Cookie("user_id")
		if errors.Is(err, http.ErrNoCookie) {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("cookie: %s=%s", cookie.Name, cookie.Value)

		shortURL := fmt.Sprintf("%s/%d", baseURL, id)
		log.Printf("short url: %s", shortURL)

		err = rep.Set(fmt.Sprintf("%d", id), urlStr, cookie.Value)
		if err != nil {
			log.Printf("error: %v", err)
			if errors.Is(err, store.ErrConstraintViolation) {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(shortURL))
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Println("save url info in repository")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))

		log.Printf("response body: %s", shortURL)
		log.Println("short url from text/plain body created")
	}
}

func CreateShortURLFromJSONHandler(rep store.Repository, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("create short URL from JSON")
		log.Printf("request url: %s %s", r.Method, r.URL)

		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("request body: %s", string(b))

		var reqBodyJSON struct {
			URL string `json:"url"`
		}
		err = json.Unmarshal(b, &reqBodyJSON)
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		URL := reqBodyJSON.URL
		log.Printf("original url from JSON: %s", URL)
		_, err = url.ParseRequestURI(URL)
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id := Hash(URL)
		log.Printf("hash for url: %d", id)

		cookie, err := r.Cookie("user_id")
		if errors.Is(err, http.ErrNoCookie) {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("cookie: %s=%s", cookie.Name, cookie.Value)

		shortURL := fmt.Sprintf("%s/%d", baseURL, id)
		log.Printf("short url: %s", shortURL)

		errSet := rep.Set(fmt.Sprintf("%d", id), URL, cookie.Value)

		resBodyJSON := struct {
			Result string `json:"result"`
		}{
			Result: shortURL,
		}

		resBody, err := json.Marshal(resBodyJSON)
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if errSet != nil {
			log.Printf("error: %v", errSet)
			if errors.Is(errSet, store.ErrConstraintViolation) {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusConflict)
				w.Write(resBody)
				return
			}
			http.Error(w, errSet.Error(), http.StatusBadRequest)
			return
		}
		log.Println("save url info in repository")

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write(resBody)

		log.Printf("response body: %s", string(resBody))
		log.Println("short URL from JSON created")
	}
}

func GetURLByIDHandler(rep store.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("get original URL by short ID")
		id := chi.URLParam(r, "ID")
		log.Printf("request url: %s %s", r.Method, r.URL)

		val, err := rep.Get(id)
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, "id not found", http.StatusBadRequest)
			return
		}
		log.Printf("load original url from repository: %s", val)

		http.Redirect(w, r, val, http.StatusTemporaryRedirect)
		w.Write([]byte("ID found"))

		log.Printf("original url %s for id %s found", val, id)
	}
}

func GetListByUserIDHandler(rep store.Repository, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("get list URLs for userID")
		log.Printf("request url: %s %s", r.Method, r.URL)

		cookie, err := r.Cookie("user_id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("cookie: %s=%s", cookie.Name, cookie.Value)

		userID := cookie.Value
		list, err := rep.GetAllByID(userID)

		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(list) == 0 {
			msg := fmt.Sprintf("No content for user with id %s", userID)
			log.Printf("No content for user with id %s", userID)
			http.Error(w, msg, http.StatusNoContent)
			return
		}

		log.Printf("load list URLs for userID %s from repository", userID)

		type data struct {
			ShortURL string `json:"short_url"`
			OrigURL  string `json:"original_url"`
		}

		var listURL []data

		for key, val := range list {
			shortURL := fmt.Sprintf("%s/%s", baseURL, key)
			listURL = append(listURL, data{ShortURL: shortURL, OrigURL: val})
			log.Printf("\t %s %s", shortURL, val)
		}

		listB, err := json.Marshal(listURL)
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(listB)

		log.Printf("response body: %s", string(listB))
		log.Printf("loaded list URLs for userID %s", userID)
	}
}

func PingHandler(rep store.Repository, dns string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("ping database")
		log.Printf("request url: %s %s", r.Method, r.URL)

		ctx, cancel := context.WithTimeout(context.Background(), store.DBConnectTimeout)
		defer cancel()

		conn, err := pgx.Connect(ctx, dns)
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = conn.Ping(context.Background())
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))

		log.Println("database is connected")
	}
}

func CreateManyShortURLHadler(rep store.Repository, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("create many short URLs from JSON")
		log.Printf("request url: %s %s", r.Method, r.URL)

		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		log.Printf("request body: %s", string(b))

		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		type inputData struct {
			CorrelationID string `json:"correlation_id"`
			OriginalURL   string `json:"original_url"`
		}

		var inputJSON []inputData
		err = json.Unmarshal(b, &inputJSON)
		if err != nil {
			msg := fmt.Sprintf("failed to unmarshal JSON: %s", err.Error())
			log.Println(msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		cookie, err := r.Cookie("user_id")
		if errors.Is(err, http.ErrNoCookie) {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("cookie: %s=%s", cookie.Name, cookie.Value)

		data := make(map[string]string)
		for _, row := range inputJSON {
			URL := row.OriginalURL
			_, err = url.ParseRequestURI(URL)
			if err != nil {
				msg := fmt.Sprintf("id %s not found", err.Error())
				log.Println(msg)
				http.Error(w, msg, http.StatusBadRequest)
				return
			}

			id := Hash(URL)
			log.Printf("\thash %d for URL %s", id, URL)
			err := rep.Set(fmt.Sprintf("%d", id), URL, cookie.Value)
			if err != nil {
				log.Printf("error: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			shortURL := fmt.Sprintf("%s/%d", baseURL, id)
			data[row.CorrelationID] = shortURL

			log.Printf("\tshort url %s for original %s saved in repository", shortURL, URL)
		}

		type outputData struct {
			CorrelationID string `json:"correlation_id"`
			ShortURL      string `json:"short_url"`
		}

		var outputJSON []outputData

		for key, val := range data {
			outputJSON = append(outputJSON, outputData{CorrelationID: key, ShortURL: val})
		}

		resBody, err := json.Marshal(outputJSON)
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write(resBody)

		log.Printf("response body: %s", string(resBody))
		log.Println("list of short URLs from JSON created")
	}
}

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
