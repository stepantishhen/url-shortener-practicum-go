package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"url-shortener-practicum-go/internal/config"
	"url-shortener-practicum-go/internal/deleter"
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

	delSvc := deleter.New(repo, logger)
	h := handlers.New(repo, cfg.BaseURL, pinger, delSvc)
	router := server.NewRouter(h, logger, cfg.SecretKey)

	srv := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		logger.Info("Starting server",
			zap.String("address", cfg.ServerAddr),
			zap.String("base_url", cfg.BaseURL),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan
	logger.Info("Received shutdown signal, stopping gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	delCtx, delCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer delCancel()

	done := make(chan struct{})
	go func() {
		if err := delSvc.Shutdown(15 * time.Second); err != nil {
			logger.Warn("Deleter service shutdown", zap.Error(err))
		} else {
			logger.Info("Deleter service stopped")
		}
		close(done)
	}()

	select {
	case <-done:
	case <-delCtx.Done():
		logger.Warn("Deleter service shutdown timeout")
	}

	logger.Info("Application stopped successfully")
}
