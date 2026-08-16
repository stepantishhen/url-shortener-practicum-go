package storage_test

import (
	"database/sql"
	"errors"
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

// newTestRepo creates a PostgresStorage and truncates the urls table before the
// test runs so each test starts with an empty slate.
func newTestRepo(t *testing.T, db *sql.DB) *storage.PostgresStorage {
	t.Helper()
	repo, err := storage.NewPostgresStorage(db)
	if err != nil {
		t.Fatalf("NewPostgresStorage: %v", err)
	}
	if _, err := db.Exec(`TRUNCATE TABLE urls`); err != nil {
		t.Fatalf("truncate urls: %v", err)
	}
	return repo
}

func TestPostgresStorage_SaveAndGet(t *testing.T) {
	db := openTestDB(t)
	repo := newTestRepo(t, db)

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
	repo := newTestRepo(t, db)

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

func TestPostgresStorage_SaveConflict(t *testing.T) {
	db := openTestDB(t)
	repo := newTestRepo(t, db)

	const original = "https://example.com/conflict-test"
	firstID, err := repo.Save(original)
	if err != nil {
		t.Fatalf("first Save: %v", err)
	}

	_, err = repo.Save(original)
	if err == nil {
		t.Fatal("second Save with same URL should return ConflictError, got nil")
	}
	var conflictErr *storage.ConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("expected ConflictError, got %T: %v", err, err)
	}
	if conflictErr.ShortID != firstID {
		t.Errorf("ConflictError.ShortID = %q, want %q", conflictErr.ShortID, firstID)
	}
}

func TestPostgresStorage_SaveBatch(t *testing.T) {
	db := openTestDB(t)
	repo := newTestRepo(t, db)

	items := []storage.BatchInput{
		{CorrelationID: "corr1", OriginalURL: "https://example.com/batch1"},
		{CorrelationID: "corr2", OriginalURL: "https://example.com/batch2"},
		{CorrelationID: "corr3", OriginalURL: "https://example.com/batch3"},
	}

	results, err := repo.SaveBatch(items)
	if err != nil {
		t.Fatalf("SaveBatch: %v", err)
	}
	if len(results) != len(items) {
		t.Fatalf("SaveBatch returned %d results, want %d", len(results), len(items))
	}

	for i, res := range results {
		if res.CorrelationID != items[i].CorrelationID {
			t.Errorf("result[%d].CorrelationID = %q, want %q", i, res.CorrelationID, items[i].CorrelationID)
		}
		if len(res.ShortID) == 0 {
			t.Errorf("result[%d].ShortID is empty", i)
		}
		got, ok := repo.Get(res.ShortID)
		if !ok {
			t.Errorf("Get(%q) after SaveBatch returned not found", res.ShortID)
		}
		if got != items[i].OriginalURL {
			t.Errorf("Get(%q) = %q, want %q", res.ShortID, got, items[i].OriginalURL)
		}
	}
}
