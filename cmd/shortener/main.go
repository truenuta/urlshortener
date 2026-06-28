package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/truenuta/urlshortener/internal/config"
	"github.com/truenuta/urlshortener/internal/handler"
)

func main() {
	cfg := config.ParseFlags()
	h := handler.NewHandler(cfg.BaseShortURLAddress)
	r := chi.NewRouter()

	r.Post("/", h.ShortenURL)
	r.Get("/{id}", h.GetOriginalURL)

	err := http.ListenAndServe(cfg.Address, r)
	if err != nil {
		panic(err)
	}

}
