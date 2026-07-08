package main

import (
	"log"
	"net/http"

	"url-shortener-practicum-go/internal/config"
	"url-shortener-practicum-go/internal/handlers"
	"url-shortener-practicum-go/internal/server"
	"url-shortener-practicum-go/internal/storage"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	repo := storage.NewMemoryStorage()
	h := handlers.New(repo, cfg.BaseURL)
	router := server.NewRouter(h)
	log.Printf("Starting server on %s", cfg.ServerAddr)
	log.Printf("Base URL: %s", cfg.BaseURL)
	if err := http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}