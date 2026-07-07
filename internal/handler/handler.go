package handler

import (
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/truenuta/urlshortener/internal/service"
)

type Handler struct {
	baseURL string
	service service.Service
}

func NewHandler(BaseShortURLAddress string, service service.Service) *Handler {
	return &Handler{
		baseURL: BaseShortURLAddress,
		service: service,
	}
}

func (h *Handler) ShortenURL(response http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return

	}
	if len(body) == 0 {
		http.Error(response, "bad request", http.StatusBadRequest)
		return
	}
	shortURL, err := h.service.Shorten(string(body))
	if err != nil {
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "text/plain")
	response.WriteHeader(http.StatusCreated)
	responseUrl, err := url.JoinPath(h.baseURL, shortURL)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
	}
	response.Write([]byte(responseUrl))
}

func (h *Handler) GetOriginalURL(response http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")
	originalURL, ok := h.service.GetURL(id)
	if !ok {
		http.Error(response, "bad request", http.StatusBadRequest)
		return
	}
	response.Header().Set("Location", originalURL)
	response.WriteHeader(http.StatusTemporaryRedirect)

}
