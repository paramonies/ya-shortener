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

		id := Hash(urlStr)
		cookie, err := r.Cookie("user_id")
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = rep.Set(fmt.Sprintf("%d", id), urlStr, cookie.Value)
		if err != nil {
			log.Printf("rep set %s...\n", err.Error())
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
			http.Error(w, "id not found", http.StatusBadRequest)
			return
		}
		URL := reqBodyJSON.URL
		log.Println(5)
		_, err = url.ParseRequestURI(URL)
		if err != nil {
			http.Error(w, "id not found", http.StatusBadRequest)
			return
		}

		id := Hash(URL)
		cookie, err := r.Cookie("user_id")
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rep.Set(fmt.Sprintf("%d", id), URL, cookie.Value)

		shortURL := fmt.Sprintf("%s/%d", baseURL, id)

		resBodyJSON := struct {
			Result string `json:"result"`
		}{
			Result: shortURL,
		}

		resBody, err := json.Marshal(resBodyJSON)
		if err != nil {
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

func GetListByUserIDHandler(rep store.Repository, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user_id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userID := cookie.Value
		list, err := rep.GetAllByID(userID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(list) == 0 {
			msg := fmt.Sprintf("No content for user with id %s", userID)
			http.Error(w, msg, http.StatusNoContent)
			return
		}

		type data struct {
			ShortURL string `json:"short_url"`
			OrigURL  string `json:"original_url"`
		}

		var listURL []data

		for key, val := range list {
			shortURL := fmt.Sprintf("%s/%s", baseURL, key)
			listURL = append(listURL, data{ShortURL: shortURL, OrigURL: val})
		}

		listB, err := json.Marshal(listURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(listB)
	}
}

func PingHandler(rep store.Repository, dns string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), store.DBConnectTimeout)
		defer cancel()
		conn, err := pgx.Connect(ctx, dns)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = conn.Ping(context.Background())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func CreateManyShortURLHadler(rep store.Repository, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
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
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		cookie, err := r.Cookie("user_id")
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data := make(map[string]string)
		for _, row := range inputJSON {
			URL := row.OriginalURL
			_, err = url.ParseRequestURI(URL)
			if err != nil {
				msg := fmt.Sprintf("id %s not found", err.Error())
				http.Error(w, msg, http.StatusBadRequest)
				return
			}
			id := Hash(URL)
			rep.Set(fmt.Sprintf("%d", id), URL, cookie.Value)
			shortURL := fmt.Sprintf("%s/%d", baseURL, id)
			data[row.CorrelationID] = shortURL
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write(resBody)
	}
}

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
