//package main
//
//import (
//	"fmt"
//	"hash/fnv"
//	"io"
//	"net/http"
//)
//
//var DB map[string]string
//var srvAddr = "localhost:8080"
//
//func URLHandler(w http.ResponseWriter, r *http.Request) {
//	switch r.Method {
//	case http.MethodGet:
//		//path := r.URL.Path
//		//id := strings.Split(path, "/")[1]
//		//fmt.Println(id)
//		//if _, ok := DB[id]; !ok {
//		//	http.Error(w, "id not found", http.StatusBadRequest)
//		//	return
//		//}
//		//w.Header().Set("Location", DB[id])
//		//w.WriteHeader(http.StatusTemporaryRedirect)
//		//w.Header().Set("Content-type", "text/plain; charset=utf-8")
//		//w.Write([]byte("id found"))
//		id := r.URL.Path[1:]
//		if val, ok := DB[id]; ok {
//			w.Header().Set("Location", val)
//			w.WriteHeader(http.StatusTemporaryRedirect)
//			//http.Redirect(w, r, val, http.StatusTemporaryRedirect)
//		} else {
//			w.WriteHeader(http.StatusBadRequest)
//			w.Write([]byte("Not found"))
//		}
//	case http.MethodPost:
//		b, err := io.ReadAll(r.Body)
//		defer r.Body.Close()
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusInternalServerError)
//			return
//		}
//
//		id := hash(string(b))
//		DB[fmt.Sprintf("%d", id)] = string(b)
//		w.Header().Set("Content-type", "text/plain; charset=utf-8")
//		w.WriteHeader(http.StatusCreated)
//		shortURL := fmt.Sprintf("https://%s/%d", srvAddr, id)
//		w.Write([]byte(shortURL))
//	default:
//		http.Error(w, "method not found", http.StatusBadRequest)
//		return
//	}
//}
//
//func hash(s string) uint32 {
//	h := fnv.New32a()
//	h.Write([]byte(s))
//	return h.Sum32()
//}
//
//func main() {
//	DB = make(map[string]string)
//	http.HandleFunc("/", URLHandler)
//	http.ListenAndServe(srvAddr, nil)
//}

package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
)

var DB = make(map[string]string)

func HashURL(url []byte, short bool) string {
	hash := fmt.Sprintf("%x", md5.Sum(url))
	if short {
		return hash[:6]
	}
	return hash
}

func IndexHandle(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		id := r.URL.Path[1:]
		if val, ok := DB[id]; ok {
			rw.Header().Set("Location", val)
			rw.WriteHeader(http.StatusTemporaryRedirect)
			//http.Redirect(w, r, val, http.StatusTemporaryRedirect)
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Not found"))
		}
	case http.MethodPost:
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		body := string(b)
		key := HashURL(b, true)
		DB[key] = body
		rw.WriteHeader(http.StatusCreated)
		rw.Write([]byte(key))
	default:
		http.Error(rw, "Wrong", http.StatusNotFound)
		return
	}
}

func main() {
	http.HandleFunc("/", IndexHandle)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
