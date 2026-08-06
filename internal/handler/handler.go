package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/truenuta/urlshortener/internal/model"
	"github.com/truenuta/urlshortener/internal/repository"
	"github.com/truenuta/urlshortener/internal/service"
	"go.uber.org/zap"
)

var conflictErr *repository.ConflictError

type Handler struct {
	baseURL string
	service service.Service
	logger  *zap.Logger
	repo    repository.URLRepository
}

func NewHandler(BaseShortURLAddress string, service service.Service, logger *zap.Logger, repo repository.URLRepository) *Handler {
	return &Handler{
		baseURL: BaseShortURLAddress,
		service: service,
		logger:  logger,
		repo:    repo,
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
		return
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

func (h *Handler) Shorten(response http.ResponseWriter, request *http.Request) {
	var req model.Request

	dec := json.NewDecoder(request.Body)
	if err := dec.Decode(&req); err != nil {
		h.logger.Debug("cannot decode request JSON body", zap.Error(err))
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	shortID, err := h.service.Shorten(req.URL)
	if err != nil {
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	responseURL, err := url.JoinPath(h.baseURL, shortID)
	if errors.As(err, &conflictErr) {
		responseUrl, joinErr := url.JoinPath(h.baseURL, conflictErr.ShortURL)
		if joinErr != nil {
			h.logger.Error("cannot create response", zap.Error(joinErr))
			http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusConflict)
		json.NewEncoder(response).Encode(model.Response{Result: responseUrl})
		return
	}

	resp := model.Response{
		Result: responseURL,
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(response)
	if err := enc.Encode(resp); err != nil {
		h.logger.Debug("error encoding response", zap.Error(err))
		return
	}
}

func (h *Handler) PingBD(response http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := h.repo.Ping(ctx); err != nil {
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (h *Handler) ShortenBatch(response http.ResponseWriter, request *http.Request) {
	var req []model.BatchRequest

	dec := json.NewDecoder(request.Body)
	if err := dec.Decode(&req); err != nil {
		h.logger.Debug("cannot decode batch request JSON body", zap.Error(err))
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(req) == 0 {
		http.Error(response, "empty batch", http.StatusBadRequest)
		return
	}
	items, err := h.service.ShortenBatch(req)
	if err != nil {
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	for i := range items {
		resultShortURL, err := url.JoinPath(h.baseURL, items[i].ShortURL)
		if err != nil {
			h.logger.Error("cannot create response", zap.Error(err))
		}
		items[i].ShortURL = resultShortURL
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(response)
	if err := enc.Encode(items); err != nil {
		h.logger.Debug("error encoding batch response", zap.Error(err))
		return
	}

}
