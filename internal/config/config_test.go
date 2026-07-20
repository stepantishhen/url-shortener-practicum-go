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

func TestEnvServerAddress(t *testing.T) {
	t.Setenv("SERVER_ADDRESS", "localhost:7777")
	cfg, err := parse([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerAddr != "localhost:7777" {
		t.Errorf("expected ServerAddr %q, got %q", "localhost:7777", cfg.ServerAddr)
	}
}

func TestEnvBaseURL(t *testing.T) {
	t.Setenv("BASE_URL", "http://env.example.com")
	cfg, err := parse([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "http://env.example.com" {
		t.Errorf("expected BaseURL %q, got %q", "http://env.example.com", cfg.BaseURL)
	}
}

func TestEnvOverridesFlag(t *testing.T) {
	t.Setenv("SERVER_ADDRESS", "localhost:9999")
	cfg, err := parse([]string{"-a", "localhost:1111"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerAddr != "localhost:9999" {
		t.Errorf("expected env value %q to override flag, got %q", "localhost:9999", cfg.ServerAddr)
	}
}

func TestDefaultFileStoragePath(t *testing.T) {
	cfg, err := parse([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FileStoragePath != "urls.json" {
		t.Errorf("expected FileStoragePath %q, got %q", "urls.json", cfg.FileStoragePath)
	}
}

func TestFlagF(t *testing.T) {
	cfg, err := parse([]string{"-f", "/tmp/storage.json"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FileStoragePath != "/tmp/storage.json" {
		t.Errorf("expected FileStoragePath %q, got %q", "/tmp/storage.json", cfg.FileStoragePath)
	}
}

func TestEnvFileStoragePath(t *testing.T) {
	t.Setenv("FILE_STORAGE_PATH", "/tmp/env-storage.json")
	cfg, err := parse([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FileStoragePath != "/tmp/env-storage.json" {
		t.Errorf("expected FileStoragePath %q, got %q", "/tmp/env-storage.json", cfg.FileStoragePath)
	}
}

func TestEnvFileStoragePathOverridesFlag(t *testing.T) {
	t.Setenv("FILE_STORAGE_PATH", "/tmp/env-storage.json")
	cfg, err := parse([]string{"-f", "/tmp/flag-storage.json"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FileStoragePath != "/tmp/env-storage.json" {
		t.Errorf("expected env value %q to override flag, got %q", "/tmp/env-storage.json", cfg.FileStoragePath)
	}
}
