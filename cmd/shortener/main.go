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
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	var db *sql.DB
	var repo storage.URLRepository

	if cfg.DatabaseDSN != "" {
		db, err = sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()

		pgRepo, err := storage.NewPostgresStorage(db)
		if err != nil {
			log.Fatalf("Failed to initialize postgres storage: %v", err)
		}
		repo = pgRepo
		logger.Info("Using PostgreSQL storage")
	} else if cfg.FileStoragePath != "" {
		fileRepo, err := storage.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			log.Fatalf("Failed to initialize file storage: %v", err)
		}
		repo = fileRepo
		logger.Info("Using file storage", zap.String("path", cfg.FileStoragePath))
	} else {
		repo = storage.NewMemoryStorage()
		logger.Info("Using memory storage")
	}

	h := handlers.New(repo, cfg.BaseURL, db)
	router := server.NewRouter(h, logger)

	logger.Info("Starting server",
		zap.String("address", cfg.ServerAddr),
		zap.String("base_url", cfg.BaseURL),
	)

	if err := http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
