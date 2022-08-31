// Package handlers defines handlers for routes.
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"sort"

	"github.com/go-chi/chi/v5"

	"github.com/paramonies/internal/store"
)

const workersCount = 10

// Handler contains common info for handler methods.
type Handler struct {
	rep            store.Repository
	url            string
	trustedNetwork *net.IPNet
}

func New(rep store.Repository, url, trustedNetwork string) *Handler {
	return &Handler{rep, url, SetTrustedNetwork(trustedNetwork)}
}

// CreateShortURL create short URL for Post text/plain
func (h *Handler) CreateShortURL() http.HandlerFunc {
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

		shortURL := fmt.Sprintf("%s/%d", h.url, id)
		log.Printf("short url: %s", shortURL)

		err = h.rep.Set(fmt.Sprintf("%d", id), urlStr, cookie.Value)
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

// CreateShortURLFromJSON create short URL for Post JSON request.
func (h *Handler) CreateShortURLFromJSON() http.HandlerFunc {
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

		shortURL := fmt.Sprintf("%s/%d", h.url, id)
		log.Printf("short url: %s", shortURL)

		errSet := h.rep.Set(fmt.Sprintf("%d", id), URL, cookie.Value)

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

// CreateManyShortURL create many URL for POST request with many json records.
func (h *Handler) CreateManyShortURL() http.HandlerFunc {
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
			err := h.rep.Set(fmt.Sprintf("%d", id), URL, cookie.Value)
			if err != nil {
				log.Printf("error: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			shortURL := fmt.Sprintf("%s/%d", h.url, id)
			data[row.CorrelationID] = shortURL

			log.Printf("\tshort url %s for original %s saved in repository", shortURL, URL)
		}

		type outputData struct {
			CorrelationID string `json:"correlation_id"`
			ShortURL      string `json:"short_url"`
		}

		var outputJSON []outputData

		keys := make([]string, 0, len(data))
		for k := range data {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			outputJSON = append(outputJSON, outputData{CorrelationID: k, ShortURL: data[k]})
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

// GetURLByID get original URL by ID.
func (h *Handler) GetURLByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("get original URL by short ID")
		id := chi.URLParam(r, "ID")
		log.Printf("request url: %s %s", r.Method, r.URL)

		val, err := h.rep.Get(id)
		if err != nil {
			log.Printf("error: %v", err)

			if errors.Is(err, store.ErrGone) {
				http.Error(w, err.Error(), http.StatusGone)
				return
			}

			http.Error(w, "id not found", http.StatusBadRequest)
			return
		}
		log.Printf("load original url from repository: %s", val)

		http.Redirect(w, r, val, http.StatusTemporaryRedirect)
		w.Write([]byte("ID found"))

		log.Printf("original url %s for id %s found", val, id)
	}
}

// GetListByUserID get all saved URLs for user ID.
func (h *Handler) GetListByUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("get list URLs for userID")
		log.Printf("request url: %s %s", r.Method, r.URL)

		cookie, err := r.Cookie("user_id")
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("cookie: %s=%s", cookie.Name, cookie.Value)

		userID := cookie.Value
		list, err := h.rep.GetAllByID(userID)

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

		keys := make([]string, 0, len(list))
		for k := range list {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			val := list[k]
			shortURL := fmt.Sprintf("%s/%s", h.url, k)
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

// Ping verify service work.
func (h *Handler) Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("ping database")
		log.Printf("request url: %s %s", r.Method, r.URL)

		err := h.rep.Ping()
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

// GetStats return internal stats about count of shorted url and users in Service
func (h *Handler) GetStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.trustedNetwork == nil || !h.trustedNetwork.Contains(net.ParseIP(r.Header.Get("X-Real-IP"))) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		stats, err := h.rep.GetStats()
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		statsB, err := json.Marshal(stats)
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(statsB)
		log.Printf("internal stats for ip %s: URL count - %d, user count - %d", h.trustedNetwork, stats.URLNumber, stats.UserNumber)
	}
}

type item struct {
	URLID  string
	UserID string
}

type errorItem struct {
	item
	Err error
}

// DeleteManyShortURL delete many short URLs for User by IDs.
func (h *Handler) DeleteManyShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("delete many short URLs")
		log.Printf("request url: %s %s", r.Method, r.URL)

		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		log.Printf("request body: %s", string(b))

		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var ids []string
		err = json.Unmarshal(b, &ids)
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

		go execDelete(ids, cookie.Value, h.rep)

		resBody := "urls deleted"

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(resBody))

		log.Printf("response body: %s", resBody)
		log.Println(resBody)
	}
}

func execDelete(ids []string, userID string, rep store.Repository) {
	log.Println("async deleting many short URLs")
	inputCh := make(chan item)

	go func() {
		for _, id := range ids {
			inputCh <- item{URLID: id, UserID: userID}
		}
		close(inputCh)
	}()

	fanOutChs := fanOut(inputCh, workersCount)

	workerChs := make([]chan errorItem, 0, workersCount)
	for _, fanOutCh := range fanOutChs {
		workerCh := newWorker(fanOutCh, rep)
		workerChs = append(workerChs, workerCh)
	}

	for errItem := range fanIn(workerChs...) {
		var msg string
		if errItem.Err != nil {
			msg = fmt.Sprintf("Error deleting URL: %v", errItem.Err)
		} else {
			msg = "URL deleted"
		}
		log.Printf("short URL: %s, user: %s, msg: %s", errItem.URLID, errItem.UserID, msg)
	}
}
