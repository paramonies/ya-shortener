package main

import (
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
)

var DB map[string]string

func URLHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		strID := r.URL.Path
		fmt.Println(strID[1:])
		id := strID[1:]
		if _, ok := DB[id]; !ok {
			http.Error(w, "id not found", http.StatusBadRequest)
			return
		}
		//w.Header().Set("content-type", "text/plain")
		w.Header().Set("Location", DB[id])
		w.WriteHeader(http.StatusTemporaryRedirect)
		//w.WriteHeader(http.StatusOK)
		//w.Write([]byte(DB[id]))
	case http.MethodPost:
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println(string(b))
		id := uuid.New().String()[:8]
		fmt.Println(id)
		DB[id] = string(b)
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(id))
	default:
		http.Error(w, "method not found", http.StatusBadRequest)
		return
	}
}

func main() {
	DB = make(map[string]string)
	http.HandleFunc("/", URLHandler)
	http.ListenAndServe("localhost:8080", nil)
}
