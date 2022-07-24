// Package middleware define middlewart for incoming request.
package middleware

import (
	"compress/gzip"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

// GzipWriter difine custom Writer.
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

// CookieMiddleware define user_id in cookie.
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("user_id")

		if errors.Is(err, http.ErrNoCookie) {
			userID, err := GenerateToken(10)
			if err != nil {
				log.Println("failed to generate token for \"user_id\" cookie: ", err)
			} else {
				cookie := &http.Cookie{
					Name:   "user_id",
					Value:  userID,
					Secure: false,
				}
				r.AddCookie(cookie)
				http.SetCookie(w, cookie)
			}
		}

		next.ServeHTTP(w, r)
	})
}
