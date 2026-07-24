package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr      string
	BaseURL         string
	FileStoragePath string
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
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	cfg.ServerAddr = envOr("SERVER_ADDRESS", cfg.ServerAddr)
	cfg.BaseURL = envOr("BASE_URL", cfg.BaseURL)
	cfg.FileStoragePath = envOr("FILE_STORAGE_PATH", cfg.FileStoragePath)
	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
