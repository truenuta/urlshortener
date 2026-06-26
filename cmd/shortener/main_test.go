package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
			name:       "PUT returns 405",
			method:     http.MethodPut,
			body:       "/",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage = make(map[string]string)
			request := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()
			ShortenRequest(recorder, request)
			response := recorder.Result()
			assert.Equal(t, tt.wantStatus, response.StatusCode)
			defer response.Body.Close()

		})
	}
}
func TestRedirect(t *testing.T) {
	storage = make(map[string]string)
	testURL := "https://example.com"
	postReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testURL))
	recorderPost := httptest.NewRecorder()
	ShortenRequest(recorderPost, postReq)
	postResp := recorderPost.Result()
	response_body, _ := io.ReadAll(postResp.Body)

	reqURL := httptest.NewRequest(http.MethodGet, string(response_body), nil)
	recorderGet := httptest.NewRecorder()
	ShortenRequest(recorderGet, reqURL)
	getURL := recorderGet.Result()

	assert.Equal(t, getURL.StatusCode, http.StatusTemporaryRedirect)
	assert.Equal(t, testURL, getURL.Header.Get("Location"))

}
