package handler

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	storage map[string]string
}

func NewHandler() *Handler {
	return &Handler{
		storage: make(map[string]string),
	}
}
func (h *Handler) shorten(body string) string {
	id := uuid.New().String() // заменить на нормальный метод
	h.storage[id] = body
	return id
}

func (h *Handler) ShortenURL(response http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)

	}
	if len(body) == 0 {
		http.Error(response, "bad request", http.StatusBadRequest)
		return
	}
	shortURL := h.shorten(string(body))
	response.Header().Set("Content-Type", "text/plain")
	response.WriteHeader(http.StatusCreated)
	response.Write([]byte("http://localhost:8080/" + shortURL))
}

func (h *Handler) GetOriginalURL(response http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")

	originalURL, ok := h.storage[id]
	if !ok {
		http.Error(response, "bad request", http.StatusBadRequest)
	}
	response.Header().Set("Location", originalURL)
	response.WriteHeader(http.StatusTemporaryRedirect)

}
