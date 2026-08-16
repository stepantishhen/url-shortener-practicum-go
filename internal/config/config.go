package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr      string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	SecretKey       string
}

func New() (*Config, error) {
	return parse(os.Args[1:])
}

func parse(args []string) (*Config, error) {
	cfg := &Config{}
	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)
	fs.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "HTTP server address")
	fs.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base URL for shortened URLs")
	fs.StringVar(&cfg.FileStoragePath, "f", "urls.json", "path to file storage")
	fs.StringVar(&cfg.DatabaseDSN, "d", "", "database DSN")
	fs.StringVar(&cfg.SecretKey, "s", "default-secret-key", "secret key for cookie signing")
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	cfg.ServerAddr = envOr("SERVER_ADDRESS", cfg.ServerAddr)
	cfg.BaseURL = envOr("BASE_URL", cfg.BaseURL)
	cfg.FileStoragePath = envOr("FILE_STORAGE_PATH", cfg.FileStoragePath)
	cfg.DatabaseDSN = envOr("DATABASE_DSN", cfg.DatabaseDSN)
	cfg.SecretKey = envOr("SECRET_KEY", cfg.SecretKey)
	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
