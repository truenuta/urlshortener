package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/truenuta/urlshortener/internal/config"
	"github.com/truenuta/urlshortener/internal/handler"
	"github.com/truenuta/urlshortener/internal/repository"
	"github.com/truenuta/urlshortener/internal/service"
)

func main() {

	cfg := config.ParseFlags()
	repository := repository.NewStorage()
	service := service.NewURLServiсe(repository)
	h := handler.NewHandler(cfg.BaseShortURLAddress, service)
	r := chi.NewRouter()

	r.Post("/", h.ShortenURL)
	r.Get("/{id}", h.GetOriginalURL)

	err := http.ListenAndServe(cfg.Address, r)
	if err != nil {
		log.Fatal(err)
	}

}
