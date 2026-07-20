package storage

import (
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

	fs, err := NewFileStorage(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	id, err := fs.Save("http://example.com")
	if err != nil {
		t.Fatal(err)
	}

	url, ok := fs.Get(id)
	if !ok {
		t.Fatal("expected to find URL")
	}
	if url != "http://example.com" {
		t.Errorf("expected %q, got %q", "http://example.com", url)
	}

	os.Remove(f.Name())
}

func TestFileStoragePersistsAcrossRestarts(t *testing.T) {
	f, err := os.CreateTemp("", "storage-*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	fs1, err := NewFileStorage(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	id, err := fs1.Save("http://example.com")
	if err != nil {
		t.Fatal(err)
	}

	fs2, err := NewFileStorage(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	url, ok := fs2.Get(id)
	if !ok {
		t.Fatal("expected URL to survive restart")
	}
	if url != "http://example.com" {
		t.Errorf("expected %q, got %q", "http://example.com", url)
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
	if _, err := fs.Save("http://example.com"); err != nil {
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
	if records[0].UUID != "1" {
		t.Errorf("expected UUID %q, got %q", "1", records[0].UUID)
	}
	if records[0].OriginalURL != "http://example.com" {
		t.Errorf("expected OriginalURL %q, got %q", "http://example.com", records[0].OriginalURL)
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
	for range 3 {
		if _, err := fs.Save("http://example.com"); err != nil {
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
	for i, r := range records {
		expected := string(rune('1' + i))
		if r.UUID != expected {
			t.Errorf("record %d: expected UUID %q, got %q", i, expected, r.UUID)
		}
	}
}
