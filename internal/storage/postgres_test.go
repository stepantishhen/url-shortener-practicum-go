package storage_test

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"url-shortener-practicum-go/internal/storage"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		t.Skip("DATABASE_DSN not set, skipping postgres integration test")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestPostgresStorage_SaveAndGet(t *testing.T) {
	db := openTestDB(t)
	repo, err := storage.NewPostgresStorage(db)
	if err != nil {
		t.Fatalf("NewPostgresStorage: %v", err)
	}

	const original = "https://example.com/postgres-test"
	id, err := repo.Save(original)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if len(id) == 0 {
		t.Fatal("Save returned empty id")
	}

	got, ok := repo.Get(id)
	if !ok {
		t.Fatalf("Get(%q) returned not found", id)
	}
	if got != original {
		t.Errorf("Get(%q) = %q, want %q", id, got, original)
	}
}

func TestPostgresStorage_GetUnknown(t *testing.T) {
	db := openTestDB(t)
	repo, err := storage.NewPostgresStorage(db)
	if err != nil {
		t.Fatalf("NewPostgresStorage: %v", err)
	}

	_, ok := repo.Get("no_such_")
	if ok {
		t.Error("Get on unknown id should return false")
	}
}

func TestPostgresStorage_MigrationsIdempotent(t *testing.T) {
	db := openTestDB(t)
	if _, err := storage.NewPostgresStorage(db); err != nil {
		t.Fatalf("first NewPostgresStorage: %v", err)
	}
	if _, err := storage.NewPostgresStorage(db); err != nil {
		t.Fatalf("second NewPostgresStorage (migrations must be idempotent): %v", err)
	}
}
