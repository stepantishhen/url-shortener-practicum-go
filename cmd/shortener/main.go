package main

import (
	"log"
	"net/http"

	"go.uber.org/zap"
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

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	var repo storage.URLRepository
	if cfg.FileStoragePath != "" {
		fileRepo, err := storage.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			log.Fatalf("Failed to initialize file storage: %v", err)
		}
		repo = fileRepo
		logger.Info("Using file storage", zap.String("path", cfg.FileStoragePath))
	} else {
		repo = storage.NewMemoryStorage()
	}
	h := handlers.New(repo, cfg.BaseURL)
	router := server.NewRouter(h, logger)

	logger.Info("Starting server",
		zap.String("address", cfg.ServerAddr),
		zap.String("base_url", cfg.BaseURL),
	)

	if err := http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}