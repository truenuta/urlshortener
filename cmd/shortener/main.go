package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/truenuta/urlshortener/internal/handler"
)

func main() {

	h := handler.NewHandler()
	r := chi.NewRouter()
	r.Post("/", h.ShortenURL)
	r.Get("/{id}", h.GetOriginalURL)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}

}
