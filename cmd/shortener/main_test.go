package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paramonies/internal/config"
	"github.com/paramonies/internal/handlers"
	"github.com/paramonies/internal/routes"
)

func TestMux(t *testing.T) {
	type want struct {
		status   int
		location string
		body     string
	}

	tests := []struct {
		name   string
		body   string
		method string
		path   string
		want   want
	}{
		{
			name:   "create short url from text/plain - OK",
			body:   "https://practicum.yandex.ru",
			method: http.MethodPost,
			path:   "/",
			want: want{
				status: http.StatusCreated,
				body:   fmt.Sprintf("http://localhost:8080/%d", handlers.Hash("https://practicum.yandex.ru")),
			},
		},
		{
			name:   "create short url from text/plain - invalid URL",
			body:   "123",
			method: http.MethodPost,
			path:   "/",
			want: want{
				status: http.StatusBadRequest,
			},
		},
		{
			name:   "get original URL by short ID - OK",
			method: http.MethodGet,
			path:   fmt.Sprintf("/%d", handlers.Hash("https://practicum.yandex.ru")),
			want: want{
				status:   http.StatusTemporaryRedirect,
				location: "https://practicum.yandex.ru",
			},
		},
		{
			name:   "get original URL by short ID - not found",
			method: http.MethodGet,
			path:   fmt.Sprintf("/%s", "123"),
			want: want{
				status: http.StatusBadRequest,
				body:   "id not found\n",
			},
		},
		{
			name:   "create short URL from JSON - OK",
			body:   `{"url":"https://practicum-1.yandex.ru"}`,
			method: http.MethodPost,
			path:   "/api/shorten",
			want: want{
				status: http.StatusCreated,
				body:   `{"result":"http://localhost:8080/3003527198"}`,
			},
		},
		{
			name:   "create short URL from JSON - BAD REQUEST",
			body:   `{"url":"https://practicum.yandex.ru"`,
			method: http.MethodPost,
			path:   "/api/shorten",
			want: want{
				status: http.StatusBadRequest,
				body:   "unexpected end of JSON input\n",
			},
		},
		{
			name:   "ping",
			method: http.MethodGet,
			path:   "/ping",
			want: want{
				status: http.StatusOK,
			},
		},
		{
			name:   "get list URLs for userID - OK",
			method: http.MethodGet,
			path:   "/api/user/urls",
			want: want{
				status: http.StatusOK,
				body:   `[{"short_url":"http://localhost:8080/3003527198","original_url":"https://practicum-1.yandex.ru"},{"short_url":"http://localhost:8080/3353207204","original_url":"https://practicum.yandex.ru"}]`,
			},
		},
		{
			name:   "create many short URLs from JSON - OK",
			body:   `[{"correlation_id": "first","original_url": "https://practicum-2.yandex.ru"},{"correlation_id": "second","original_url": "https://practicum-3.yandex.ru"}]`,
			method: http.MethodPost,
			path:   "/api/shorten/batch",
			want: want{
				status: http.StatusCreated,
				body:   `[{"correlation_id":"first","short_url":"http://localhost:8080/3159787651"},{"correlation_id":"second","short_url":"http://localhost:8080/740694524"}]`,
			},
		},
		{
			name:   "delete many short URLs Accepted",
			body:   `["3159787651", "740694524"]`,
			method: http.MethodDelete,
			path:   "/api/user/urls",
			want: want{
				status: http.StatusAccepted,
			},
		},
	}

	cfg := config.Config{
		SrvAddr: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}

	r, err := config.NewRepository(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	h := handlers.New(r, cfg.BaseURL)

	rtr := routes.New(h)
	ts := httptest.NewServer(rtr)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, ts.URL+tt.path, strings.NewReader(tt.body))
			require.NoError(t, err)

			cookie := &http.Cookie{
				Name:  "user_id",
				Value: "wSzPHUbHwQ/WKQ==",
			}
			req.AddCookie(cookie)

			client := &http.Client{}
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}

			resp, err := client.Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			assert.NoError(t, err)

			assert.Equal(t, tt.want.status, resp.StatusCode)

			if tt.want.body != "" {
				assert.Equal(t, tt.want.body, string(body))
			}

			if tt.want.location != "" {
				assert.Equal(t, tt.want.location, resp.Header.Get("Location"))
			}
		})
	}
}
