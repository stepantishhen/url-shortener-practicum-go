package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr string
	BaseURL    string
}

func New() (*Config, error) {
	return parse(os.Args[1:])
}

func parse(args []string) (*Config, error) {
	cfg := &Config{}
	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)
	fs.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "HTTP server address")
	fs.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base URL for shortened URLs")
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	if v := os.Getenv("SERVER_ADDRESS"); v != "" {
		cfg.ServerAddr = v
	}
	if v := os.Getenv("BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	return cfg, nil
}
