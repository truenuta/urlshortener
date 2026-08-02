package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/truenuta/urlshortener/internal/db"
	"github.com/truenuta/urlshortener/internal/handler"
	"github.com/truenuta/urlshortener/internal/model"
	"github.com/truenuta/urlshortener/internal/repository"
	"github.com/truenuta/urlshortener/internal/service"
	"go.uber.org/zap"
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
			zapLogger, err := zap.NewDevelopment()
			if err != nil {
				t.Fatalf("logger did not init, %v", err)
			}
			repository, err := repository.NewStorage("")
			if err != nil {
				zapLogger.Fatal("failed to initialiaze storage", zap.Error(err))
			}
			defer repository.Close()

			database, err := db.NewDB("")
			if err != nil {
				zapLogger.Fatal("failed to connect to database", zap.Error(err))
			}
			defer database.Close()

			service := service.NewURLServiсe(repository)
			h := handler.NewHandler("http://localhost:8080", service, zapLogger, database)
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
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("logger did not init, %v", err)
	}
	repository, err := repository.NewStorage("")
	if err != nil {
		zapLogger.Fatal("failed to initialiaze storage", zap.Error(err))
	}
	defer repository.Close()

	database, err := db.NewDB("")
	if err != nil {
		zapLogger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	service := service.NewURLServiсe(repository)
	h := handler.NewHandler("http://localhost:8080", service, zapLogger, database)

	r := chi.NewRouter()
	r.Post("/", h.ShortenURL)
	r.Get("/{id}", h.GetOriginalURL)

	testURL := "https://example.com"
	postReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testURL))
	recorderPost := httptest.NewRecorder()
	r.ServeHTTP(recorderPost, postReq)
	postResp := recorderPost.Result()
	responseBody, _ := io.ReadAll(postResp.Body)

	path := strings.TrimPrefix(string(responseBody), "http://localhost:8080")
	reqURL := httptest.NewRequest(http.MethodGet, path, nil)
	recorderGet := httptest.NewRecorder()
	r.ServeHTTP(recorderGet, reqURL)
	getURL := recorderGet.Result()

	assert.Equal(t, http.StatusTemporaryRedirect, getURL.StatusCode)
	assert.Equal(t, testURL, getURL.Header.Get("Location"))

}

func TestAPIShorten(t *testing.T) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("logger did not init, %v", err)
	}
	repository, err := repository.NewStorage("")
	service := service.NewURLServiсe(repository)

	database, err := db.NewDB("")
	if err != nil {
		zapLogger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	h := handler.NewHandler("http://localhost:8080", service, zapLogger, database)

	r := chi.NewRouter()
	r.Post("/api/shorten", h.Shorten)

	var req model.Request
	req.URL = "https://example.com"

	body, _ := json.Marshal(req)

	postReq := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	recorderPost := httptest.NewRecorder()
	r.ServeHTTP(recorderPost, postReq)
	postResp := recorderPost.Result()

	assert.Equal(t, http.StatusCreated, postResp.StatusCode)
	assert.Equal(t, "application/json", postResp.Header.Get("Content-Type"))

	var resp model.Response
	err = json.NewDecoder(postResp.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("response decoded with an error, %v", err)
	}
	assert.NotEmpty(t, resp.Result)
	assert.Contains(t, resp.Result, "http://localhost:8080")
}
