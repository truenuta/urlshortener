package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/truenuta/urlshortener/internal/handler"
)

func TestShortenRequest(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
	}{
		{
			name:       "POST valid URL returns 201",
			method:     http.MethodPost,
			body:       "https://example.com",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "POST empty body",
			method:     http.MethodPost,
			body:       "",
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handler.NewHandler()

			r := chi.NewRouter()
			r.Post("/", h.ShortenURL)
			r.Get("/{id}", h.GetOriginalURL)

			request := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, request)
			response := recorder.Result()
			assert.Equal(t, tt.wantStatus, response.StatusCode)
			defer response.Body.Close()

		})
	}
}
func TestRedirect(t *testing.T) {
	h := handler.NewHandler()

	r := chi.NewRouter()
	r.Post("/", h.ShortenURL)
	r.Get("/{id}", h.GetOriginalURL)

	testURL := "https://example.com"
	postReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testURL))
	recorderPost := httptest.NewRecorder()
	r.ServeHTTP(recorderPost, postReq)
	postResp := recorderPost.Result()
	response_body, _ := io.ReadAll(postResp.Body)

	path := strings.TrimPrefix(string(response_body), "http://localhost:8080")
	reqURL := httptest.NewRequest(http.MethodGet, path, nil)
	recorderGet := httptest.NewRecorder()
	r.ServeHTTP(recorderGet, reqURL)
	getURL := recorderGet.Result()

	assert.Equal(t, getURL.StatusCode, http.StatusTemporaryRedirect)
	assert.Equal(t, testURL, getURL.Header.Get("Location"))

}
