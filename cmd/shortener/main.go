package main

import (
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"strings"
)

var DB map[string]string
var srvAddr = "localhost:8080"

func URLHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		path := r.URL.Path
		id := strings.Split(path, "/")[1]
		fmt.Println(id)
		if _, ok := DB[id]; !ok {
			http.Error(w, "id not found", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Header().Set("Location", DB[id])
		w.Write([]byte("id found"))
	case http.MethodPost:
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id := hash(string(b))
		DB[fmt.Sprintf("%d", id)] = string(b)
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		shortURL := fmt.Sprintf("https://%s/%d", srvAddr, id)
		w.Write([]byte(shortURL))
	default:
		http.Error(w, "method not found", http.StatusBadRequest)
		return
	}
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func main() {
	DB = make(map[string]string)
	http.HandleFunc("/", URLHandler)
	http.ListenAndServe(srvAddr, nil)
}
