package main

import (
	"log"
	"net/http"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/truenuta/urlshortener/internal/config"
	"github.com/truenuta/urlshortener/internal/db"
	"github.com/truenuta/urlshortener/internal/handler"
	"github.com/truenuta/urlshortener/internal/logger"
	"github.com/truenuta/urlshortener/internal/middleware"
	"github.com/truenuta/urlshortener/internal/repository"
	"github.com/truenuta/urlshortener/internal/service"
)

func main() {

	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("logger did not init, %v", err)
	}
	cfg, err := config.NewConfig()
	if err != nil {
		zapLogger.Fatal("failed to get config", zap.Error(err))
	}
	repository, err := repository.NewStorage(cfg.FileStoragePath)
	if err != nil {
		zapLogger.Fatal("failed to initialiaze storage", zap.Error(err))
	}
	defer repository.Close()

	err = repository.Load()
	if err != nil {
		zapLogger.Fatal("failed to load storage", zap.Error(err))
	}
	database, err := db.NewDB(cfg.DataBaseDSN)
	if err != nil {
		zapLogger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer database.Close()
	service := service.NewURLServiсe(repository)
	h := handler.NewHandler(cfg.BaseShortURLAddress, service, zapLogger, database)
	r := chi.NewRouter()
	r.Use(logger.RequestLogger(zapLogger))
	r.Use(middleware.GzipMiddleware)

	r.Post("/", h.ShortenURL)
	r.Post("/api/shorten", h.Shorten)
	r.Get("/{id}", h.GetOriginalURL)
	r.Get("/ping", h.PingBD)

	LaSerr := http.ListenAndServe(cfg.Address, r)
	if LaSerr != nil {
		zapLogger.Fatal("server failed", zap.Error(LaSerr))
	}

}
