package main

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/truenuta/urlshortener/cmd/skill"
	"github.com/truenuta/urlshortener/internal/config"
	"github.com/truenuta/urlshortener/internal/handler"
	"github.com/truenuta/urlshortener/internal/logger"
	"github.com/truenuta/urlshortener/internal/repository"
	"github.com/truenuta/urlshortener/internal/service"
)

func main() {

	cfg := config.NewConfig()
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	repository := repository.NewStorage()
	service := service.NewURLServiсe(repository)
	h := handler.NewHandler(cfg.BaseShortURLAddress, service, zapLogger)
	r := chi.NewRouter()
	r.Use(logger.RequestLogger(zapLogger))
	r.Use(skill.GzipMiddleware)

	r.Post("/", h.ShortenURL)
	r.Post("/api/shorten", h.Shorten)
	r.Get("/{id}", h.GetOriginalURL)

	LaSerr := http.ListenAndServe(cfg.Address, r)
	if LaSerr != nil {
		zapLogger.Fatal("server failed", zap.Error(LaSerr))
	}

}
