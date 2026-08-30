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
	"github.com/truenuta/urlshortener/internal/deleter"
	"github.com/truenuta/urlshortener/internal/middleware"
	"github.com/truenuta/urlshortener/internal/model"
	"github.com/truenuta/urlshortener/internal/repository"
	"github.com/truenuta/urlshortener/internal/service"
	"go.uber.org/zap"
)

type Service interface {
	Shorten(url, userID string) (string, error)
	GetURL(id string) (URL string, isDeleted bool, ok bool)
	ShortenBatch(items []model.BatchRequest, userID string) ([]model.BatchResponse, error)
	GetUserURLs(userID string) ([]model.UserURL, error)
}

type Handler struct {
	baseURL string
	service Service
	logger  *zap.Logger
	repo    repository.URLRepository
	deleter *deleter.Deleter
}

func NewHandler(BaseShortURLAddress string, service Service, logger *zap.Logger, repo repository.URLRepository, deleter *deleter.Deleter) *Handler {
	return &Handler{
		baseURL: BaseShortURLAddress,
		service: service,
		logger:  logger,
		repo:    repo,
		deleter: deleter,
	}
}

func (h *Handler) writeConflictResponse(response http.ResponseWriter, shortURL string) (string, bool) {
	responseUrl, joinErr := url.JoinPath(h.baseURL, shortURL)
	if joinErr != nil {
		h.logger.Error("can not create response", zap.Error(joinErr))
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return "", false
	}
	return responseUrl, true
}

func (h *Handler) ShortenURL(response http.ResponseWriter, request *http.Request) {
	var conflictErr *service.ConflictError
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return

	}
	if len(body) == 0 {
		http.Error(response, "bad request", http.StatusBadRequest)
		return
	}
	userID, _ := middleware.UserIDFromContext(request.Context())
	shortURL, err := h.service.Shorten(string(body), userID)
	if err != nil {
		if errors.As(err, &conflictErr) {
			if responseUrl, ok := h.writeConflictResponse(response, conflictErr.ShortURL); ok {
				response.Header().Set("Content-Type", "text/plain")
				response.WriteHeader(http.StatusConflict)
				response.Write([]byte(responseUrl))
				return
			}
			return
		}
	}
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	responseUrl, err := url.JoinPath(h.baseURL, shortURL)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}
	response.Header().Set("Content-Type", "text/plain")
	response.WriteHeader(http.StatusCreated)
	response.Write([]byte(responseUrl))
}

func (h *Handler) GetOriginalURL(response http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")
	originalURL, deleted, ok := h.service.GetURL(id)
	if !ok {
		http.Error(response, "not found", http.StatusNotFound)
		return
	}
	if deleted {
		response.WriteHeader(http.StatusGone)
		return
	}
	response.Header().Set("Location", originalURL)
	response.WriteHeader(http.StatusTemporaryRedirect)

}

func (h *Handler) Shorten(response http.ResponseWriter, request *http.Request) {
	var conflictErr *service.ConflictError
	var req model.Request

	dec := json.NewDecoder(request.Body)
	if err := dec.Decode(&req); err != nil {
		h.logger.Debug("cannot decode request JSON body", zap.Error(err))
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, _ := middleware.UserIDFromContext(request.Context())
	shortID, err := h.service.Shorten(req.URL, userID)
	if errors.As(err, &conflictErr) {
		if responseUrl, ok := h.writeConflictResponse(response, conflictErr.ShortURL); ok {
			response.Header().Set("Content-Type", "application/json")
			response.WriteHeader(http.StatusConflict)
			json.NewEncoder(response).Encode(model.Response{Result: responseUrl})
			return
		}
		return
	}
	if err != nil {
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseURL, err := url.JoinPath(h.baseURL, shortID)

	if err != nil {
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
	userID, _ := middleware.UserIDFromContext(request.Context())
	items, err := h.service.ShortenBatch(req, userID)
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

func (h *Handler) GetUserURLs(response http.ResponseWriter, request *http.Request) {
	if middleware.AuthFailedFromContext(request.Context()) {
		response.WriteHeader(http.StatusUnauthorized)
		return
	}
	userID, _ := middleware.UserIDFromContext(request.Context())
	urls, err := h.service.GetUserURLs(userID)
	if err != nil {
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(urls) == 0 {
		response.WriteHeader(http.StatusNoContent)
		return
	}
	for i := range urls {
		full, err := url.JoinPath(h.baseURL, urls[i].ShortURL)
		if err != nil {
			h.logger.Error("cannot build short url", zap.Error(err))
		}
		urls[i].ShortURL = full
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(urls)
}

func (h *Handler) DeleteUserURLs(response http.ResponseWriter, request *http.Request) {
	if middleware.AuthFailedFromContext(request.Context()) {
		response.WriteHeader(http.StatusUnauthorized)
		return
	}
	userID, _ := middleware.UserIDFromContext(request.Context())
	var shortURLs []string
	if err := json.NewDecoder(request.Body).Decode(&shortURLs); err != nil {
		h.logger.Debug("cannot decode delete request JSON body", zap.Error(err))
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	h.deleter.ScheduleDeletion(request.Context(), userID, shortURLs)
	response.WriteHeader(http.StatusAccepted)

}
