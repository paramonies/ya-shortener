package main

import (
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
)

var DB map[string]string
var srvAddr = "localhost:8080"

func URLHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		//path := r.URL.Path
		//id := strings.Split(path, "/")[1]
		//fmt.Println(id)
		//if _, ok := DB[id]; !ok {
		//	http.Error(w, "id not found", http.StatusBadRequest)
		//	return
		//}
		//w.Header().Set("Location", DB[id])
		//w.WriteHeader(http.StatusTemporaryRedirect)
		//w.Header().Set("Content-type", "text/plain; charset=utf-8")
		//w.Write([]byte("id found"))
		id := r.URL.Path[1:]
		if val, ok := DB[id]; ok {
			w.Header().Set("Location", val)
			w.WriteHeader(http.StatusTemporaryRedirect)
			//http.Redirect(w, r, val, http.StatusTemporaryRedirect)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Not found"))
		}
	case http.MethodPost:
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id := hash(string(b))
		DB[fmt.Sprintf("%d", id)] = string(b)
		w.Header().Set("Content-type", "text/plain; charset=utf-8")
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
