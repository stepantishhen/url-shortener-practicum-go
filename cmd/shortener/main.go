package main

import (
	"net/http"

	"url-shortener-practicum-go/internal/config"
	"url-shortener-practicum-go/internal/handlers"
	"url-shortener-practicum-go/internal/server"
	"url-shortener-practicum-go/internal/storage"
)

func main() {
	cfg := config.New()
	repo := storage.NewMemoryStorage()
	h := handlers.New(repo, cfg.BaseURL)
	http.ListenAndServe(cfg.ServerAddr, server.NewRouter(h))
}
