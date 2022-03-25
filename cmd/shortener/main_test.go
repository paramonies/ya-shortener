package main

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/paramonies/internal/handlers"
	"github.com/paramonies/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
			name:   "POST OK",
			body:   "https://practicum.yandex.ru",
			method: http.MethodPost,
			path:   "/",
			want: want{
				status: http.StatusCreated,
				body:   fmt.Sprintf("http://localhost:8080/%d", handlers.Hash("https://practicum.yandex.ru")),
			},
		},
		{
			name:   "POST invalid URL",
			body:   "123",
			method: http.MethodPost,
			path:   "/",
			want: want{
				status: http.StatusBadRequest,
			},
		},
		{
			name:   "GET OK",
			method: http.MethodGet,
			path:   fmt.Sprintf("/%d", handlers.Hash("https://practicum.yandex.ru")),
			want: want{
				status:   http.StatusTemporaryRedirect,
				location: "https://practicum.yandex.ru",
			},
		},
		{
			name:   "GET ID not found",
			method: http.MethodGet,
			path:   fmt.Sprintf("/%s", "123"),
			want: want{
				status: http.StatusBadRequest,
				body:   "id not found\n",
			},
		},
		{
			name:   "POST JSON OK",
			body:   `{"url":"https://practicum.yandex.ru"}`,
			method: http.MethodPost,
			path:   "/api/shorten",
			want: want{
				status: http.StatusCreated,
				body:   `{"result":"http://localhost:8080/3353207204"}`,
			},
		},
		{
			name:   "POST JSON BAD REQUEST",
			body:   `{"url":"https://practicum.yandex.ru"`,
			method: http.MethodPost,
			path:   "/api/shorten",
			want: want{
				status: http.StatusBadRequest,
				body:   "id not found\n",
			},
		},
	}

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	r := NewRouter(store.NewMapDB(), &cfg)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, ts.URL+tt.path, strings.NewReader(tt.body))
			require.NoError(t, err)

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
