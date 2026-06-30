package config

import "testing"

func TestDefaults(t *testing.T) {
	cfg, err := parse([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerAddr != "localhost:8080" {
		t.Errorf("expected ServerAddr %q, got %q", "localhost:8080", cfg.ServerAddr)
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("expected BaseURL %q, got %q", "http://localhost:8080", cfg.BaseURL)
	}
}

func TestFlagA(t *testing.T) {
	cfg, err := parse([]string{"-a", "localhost:9090"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerAddr != "localhost:9090" {
		t.Errorf("expected ServerAddr %q, got %q", "localhost:9090", cfg.ServerAddr)
	}
}

func TestFlagB(t *testing.T) {
	cfg, err := parse([]string{"-b", "http://example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "http://example.com" {
		t.Errorf("expected BaseURL %q, got %q", "http://example.com", cfg.BaseURL)
	}
}

func TestBothFlags(t *testing.T) {
	cfg, err := parse([]string{"-a", "0.0.0.0:9000", "-b", "http://example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerAddr != "0.0.0.0:9000" {
		t.Errorf("expected ServerAddr %q, got %q", "0.0.0.0:9000", cfg.ServerAddr)
	}
	if cfg.BaseURL != "http://example.com" {
		t.Errorf("expected BaseURL %q, got %q", "http://example.com", cfg.BaseURL)
	}
}

func TestUnknownFlagReturnsError(t *testing.T) {
	_, err := parse([]string{"-z", "value"})
	if err == nil {
		t.Error("expected error for unknown flag, got nil")
	}
}
