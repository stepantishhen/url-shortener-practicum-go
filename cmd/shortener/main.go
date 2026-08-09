package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"url-shortener-practicum-go/internal/config"
	"url-shortener-practicum-go/internal/handlers"
	"url-shortener-practicum-go/internal/server"
	"url-shortener-practicum-go/internal/storage"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	cfg, err := config.New()
	if err != nil {
		logger.Fatal("Failed to initialize config", zap.Error(err))
	}

	var repo storage.URLRepository
	var pinger storage.Pinger

	if cfg.DatabaseDSN != "" {
		db, err := sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			logger.Fatal("Failed to open database", zap.Error(err))
		}
		defer db.Close()

		pgRepo, err := storage.NewPostgresStorage(db)
		if err != nil {
			logger.Fatal("Failed to initialize postgres storage", zap.Error(err))
		}
		repo = pgRepo
		pinger = pgRepo
		logger.Info("Using PostgreSQL storage")
	} else if cfg.FileStoragePath != "" {
		fileRepo, err := storage.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			logger.Fatal("Failed to initialize file storage", zap.Error(err))
		}
		repo = fileRepo
		pinger = fileRepo
		logger.Info("Using file storage", zap.String("path", cfg.FileStoragePath))
	} else {
		memRepo := storage.NewMemoryStorage()
		repo = memRepo
		pinger = memRepo
		logger.Info("Using memory storage")
	}

	h := handlers.New(repo, cfg.BaseURL, pinger)
	router := server.NewRouter(h, logger)

	logger.Info("Starting server",
		zap.String("address", cfg.ServerAddr),
		zap.String("base_url", cfg.BaseURL),
	)

	if err := http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
