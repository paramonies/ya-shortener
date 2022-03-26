package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type GzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (gz GzipWriter) Write(p []byte) (int, error) {
	return gz.Writer.Write(p)
}

func NewGzipWriter(rw http.ResponseWriter, w io.Writer) GzipWriter {
	return GzipWriter{ResponseWriter: rw, Writer: w}
}

func GzipCompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gzipw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}

		defer gzipw.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(NewGzipWriter(w, gzipw), r)
	})
}

func GzipDECompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reader io.Reader
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			reader = gzipr
			defer gzipr.Close()
		} else {
			reader = r.Body
		}

		b, err := io.ReadAll(reader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(strings.NewReader(string(b)))
		r.ContentLength = int64(len(b))
		next.ServeHTTP(w, r)
	})
}
