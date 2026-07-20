package server

import (
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"url-shortener-practicum-go/internal/handlers"
	"url-shortener-practicum-go/internal/middleware"
)

func NewRouter(h *handlers.Handler, log *zap.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger(log))
	r.Post("/", h.ShortenURL)
	r.Get("/{id}", h.Redirect)
	return r
}
