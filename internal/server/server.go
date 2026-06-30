package server

import (
	"github.com/go-chi/chi"
	"url-shortener-practicum-go/internal/handlers"
)

func NewRouter(h *handlers.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", h.ShortenURL)
	r.Get("/{id}", h.Redirect)
	return r
}
