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
	err := fs.Parse(args)
	return cfg, err
}
