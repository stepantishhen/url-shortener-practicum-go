package storage

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestFileStorageSaveAndGet(t *testing.T) {
	f, err := os.CreateTemp("", "storage-*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	fs, err := NewFileStorage(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	const userID = "user1"
	id, err := fs.Save(userID, "http://example.com")
	if err != nil {
		t.Fatal(err)
	}

	url, ok, deleted := fs.Get(id)
	if !ok {
		t.Fatal("expected to find URL by id")
	}
	if deleted {
		t.Errorf("Get: expected deleted=false, got true")
	}
	if url != "http://example.com" {
		t.Errorf("Get: expected %q, got %q", "http://example.com", url)
	}

	urls, err := fs.GetByUser(userID)
	if err != nil {
		t.Fatalf("GetByUser: %v", err)
	}
	if len(urls) != 1 || urls[0].ShortID != id {
		t.Errorf("GetByUser: expected [{%s http://example.com}], got %v", id, urls)
	}
}

func TestFileStoragePersistsAcrossRestarts(t *testing.T) {
	f, err := os.CreateTemp("", "storage-*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	const userID = "user1"
	fs1, err := NewFileStorage(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	id, err := fs1.Save(userID, "http://example.com")
	if err != nil {
		t.Fatal(err)
	}

	fs2, err := NewFileStorage(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	url, ok, _ := fs2.Get(id)
	if !ok {
		t.Fatal("expected URL to survive restart")
	}
	if url != "http://example.com" {
		t.Errorf("expected %q, got %q", "http://example.com", url)
	}

	urls, err := fs2.GetByUser(userID)
	if err != nil {
		t.Fatalf("GetByUser after restart: %v", err)
	}
	if len(urls) != 1 || urls[0].ShortID != id {
		t.Errorf("GetByUser after restart: expected [{%s ...}], got %v", id, urls)
	}
}

func TestFileStorageJSONFormat(t *testing.T) {
	f, err := os.CreateTemp("", "storage-*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	fs, err := NewFileStorage(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	const userID = "user1"
	if _, err := fs.Save(userID, "http://example.com"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	var records []record
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].UUID == "" {
		t.Error("expected non-empty UUID")
	}
	if records[0].OriginalURL != "http://example.com" {
		t.Errorf("expected OriginalURL %q, got %q", "http://example.com", records[0].OriginalURL)
	}
	if records[0].UserID != userID {
		t.Errorf("expected UserID %q, got %q", userID, records[0].UserID)
	}
}

func TestFileStorageUUIDIncrement(t *testing.T) {
	f, err := os.CreateTemp("", "storage-*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	fs, err := NewFileStorage(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	const userID = "user1"
	for range 3 {
		if _, err := fs.Save(userID, "http://example.com"); err != nil {
			t.Fatal(err)
		}
	}

	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	var records []record
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool, len(records))
	for i, r := range records {
		if r.UUID == "" {
			t.Errorf("record %d: UUID is empty", i)
		}
		if seen[r.UUID] {
			t.Errorf("record %d: duplicate UUID %q", i, r.UUID)
		}
		seen[r.UUID] = true
		if r.UserID != userID {
			t.Errorf("record %d: expected UserID %q, got %q", i, userID, r.UserID)
		}
	}
}

func TestFileStorageSaveBatch(t *testing.T) {
	f, _ := os.CreateTemp("", "storage-*.json")
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	fs, err := NewFileStorage(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	items := []BatchInput{
		{CorrelationID: "c1", OriginalURL: "https://example.com/1"},
		{CorrelationID: "c2", OriginalURL: "https://example.com/2"},
	}
	results, err := fs.SaveBatch("user1", items)
	if err != nil {
		t.Fatalf("SaveBatch: %v", err)
	}
	if len(results) != len(items) {
		t.Fatalf("expected %d results, got %d", len(items), len(results))
	}
	for i, res := range results {
		if res.CorrelationID != items[i].CorrelationID {
			t.Errorf("result[%d].CorrelationID = %q, want %q", i, res.CorrelationID, items[i].CorrelationID)
		}
		got, ok, _ := fs.Get(res.ShortID)
		if !ok {
			t.Errorf("Get(%q) after SaveBatch returned not found", res.ShortID)
		}
		if got != items[i].OriginalURL {
			t.Errorf("Get(%q) = %q, want %q", res.ShortID, got, items[i].OriginalURL)
		}
	}
}

func TestFileStorageSaveBatch_PersistsAcrossRestarts(t *testing.T) {
	f, _ := os.CreateTemp("", "storage-*.json")
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	fs1, _ := NewFileStorage(f.Name())
	items := []BatchInput{
		{CorrelationID: "c1", OriginalURL: "https://batch.com/1"},
	}
	results, _ := fs1.SaveBatch("user1", items)

	fs2, err := NewFileStorage(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	got, ok, _ := fs2.Get(results[0].ShortID)
	if !ok {
		t.Fatal("expected batch URL to survive restart")
	}
	if got != items[0].OriginalURL {
		t.Errorf("got %q, want %q", got, items[0].OriginalURL)
	}
}

func TestFileStorageGetByUser(t *testing.T) {
	f, _ := os.CreateTemp("", "storage-*.json")
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	fs, _ := NewFileStorage(f.Name())
	id1, _ := fs.Save("alice", "https://a.com")
	id2, _ := fs.Save("alice", "https://b.com")
	fs.Save("bob", "https://c.com")

	urls, err := fs.GetByUser("alice")
	if err != nil {
		t.Fatalf("GetByUser: %v", err)
	}
	if len(urls) != 2 {
		t.Fatalf("expected 2 URLs for alice, got %d", len(urls))
	}
	ids := map[string]bool{id1: true, id2: true}
	for _, u := range urls {
		if !ids[u.ShortID] {
			t.Errorf("unexpected ShortID %q", u.ShortID)
		}
	}
}

func TestFileStorageGetByUser_PersistsAcrossRestarts(t *testing.T) {
	f, _ := os.CreateTemp("", "storage-*.json")
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	fs1, _ := NewFileStorage(f.Name())
	fs1.Save("alice", "https://a.com")
	fs1.Save("alice", "https://b.com")

	fs2, err := NewFileStorage(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	urls, err := fs2.GetByUser("alice")
	if err != nil {
		t.Fatalf("GetByUser after restart: %v", err)
	}
	if len(urls) != 2 {
		t.Errorf("expected 2 URLs after restart, got %d", len(urls))
	}
}

func TestFileStoragePingContext(t *testing.T) {
	f, _ := os.CreateTemp("", "storage-*.json")
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	fs, _ := NewFileStorage(f.Name())
	if err := fs.PingContext(context.Background()); err != nil {
		t.Errorf("PingContext returned unexpected error: %v", err)
	}
}

func TestConflictError_Error(t *testing.T) {
	err := &ConflictError{ShortID: "abc12345"}
	want := "url already exists: abc12345"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestFileStorageDeleteBatch(t *testing.T) {
	f, _ := os.CreateTemp("", "storage-*.json")
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	fs, _ := NewFileStorage(f.Name())
	id, _ := fs.Save("alice", "https://example.com")

	if err := fs.DeleteBatch("alice", []string{id}); err != nil {
		t.Fatalf("DeleteBatch: %v", err)
	}

	_, found, deleted := fs.Get(id)
	if !found {
		t.Fatal("URL should still exist after soft delete")
	}
	if !deleted {
		t.Error("expected deleted=true after DeleteBatch")
	}
}

func TestFileStorageDeleteBatch_OwnershipEnforced(t *testing.T) {
	f, _ := os.CreateTemp("", "storage-*.json")
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	fs, _ := NewFileStorage(f.Name())
	id, _ := fs.Save("alice", "https://example.com")

	// bob tries to delete alice's URL — must be ignored.
	if err := fs.DeleteBatch("bob", []string{id}); err != nil {
		t.Fatalf("DeleteBatch: %v", err)
	}

	_, _, deleted := fs.Get(id)
	if deleted {
		t.Error("bob must not be able to delete alice's URL")
	}
}

func TestFileStorageDeleteBatch_PersistsAcrossRestarts(t *testing.T) {
	f, _ := os.CreateTemp("", "storage-*.json")
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	fs1, _ := NewFileStorage(f.Name())
	id, _ := fs1.Save("alice", "https://example.com")
	if err := fs1.DeleteBatch("alice", []string{id}); err != nil {
		t.Fatalf("DeleteBatch: %v", err)
	}

	fs2, err := NewFileStorage(f.Name())
	if err != nil {
		t.Fatalf("NewFileStorage after restart: %v", err)
	}
	_, found, deleted := fs2.Get(id)
	if !found {
		t.Fatal("URL should still exist after restart")
	}
	if !deleted {
		t.Error("deleted flag must survive restart")
	}
}
