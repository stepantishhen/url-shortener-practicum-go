package storage

import (
	"context"
	"testing"
)

func TestMemoryStorage_SaveAndGet(t *testing.T) {
	m := NewMemoryStorage()

	id, err := m.Save("user1", "https://example.com")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if id == "" {
		t.Fatal("Save returned empty id")
	}

	got, ok, deleted := m.Get(id)
	if !ok {
		t.Fatalf("Get(%q) returned not found", id)
	}
	if deleted {
		t.Errorf("Get(%q) returned deleted=true unexpectedly", id)
	}
	if got != "https://example.com" {
		t.Errorf("Get(%q) = %q, want %q", id, got, "https://example.com")
	}
}

func TestMemoryStorage_GetUnknown(t *testing.T) {
	m := NewMemoryStorage()

	_, ok, _ := m.Get("nonexistent")
	if ok {
		t.Error("Get on unknown id should return false")
	}
}

func TestMemoryStorage_SaveBatch(t *testing.T) {
	m := NewMemoryStorage()

	items := []BatchInput{
		{CorrelationID: "c1", OriginalURL: "https://example.com/1"},
		{CorrelationID: "c2", OriginalURL: "https://example.com/2"},
	}
	results, err := m.SaveBatch("user1", items)
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
		if res.ShortID == "" {
			t.Errorf("result[%d].ShortID is empty", i)
		}
		got, ok, _ := m.Get(res.ShortID)
		if !ok {
			t.Errorf("Get(%q) after SaveBatch returned not found", res.ShortID)
		}
		if got != items[i].OriginalURL {
			t.Errorf("Get(%q) = %q, want %q", res.ShortID, got, items[i].OriginalURL)
		}
	}
}

func TestMemoryStorage_GetByUser(t *testing.T) {
	m := NewMemoryStorage()

	id1, _ := m.Save("alice", "https://a.com")
	id2, _ := m.Save("alice", "https://b.com")
	m.Save("bob", "https://c.com")

	urls, err := m.GetByUser("alice")
	if err != nil {
		t.Fatalf("GetByUser: %v", err)
	}
	if len(urls) != 2 {
		t.Fatalf("expected 2 URLs for alice, got %d", len(urls))
	}

	ids := map[string]bool{id1: true, id2: true}
	for _, u := range urls {
		if !ids[u.ShortID] {
			t.Errorf("unexpected ShortID %q in result", u.ShortID)
		}
	}
}

func TestMemoryStorage_GetByUser_Empty(t *testing.T) {
	m := NewMemoryStorage()
	m.Save("other", "https://example.com")

	urls, err := m.GetByUser("nobody")
	if err != nil {
		t.Fatalf("GetByUser: %v", err)
	}
	if len(urls) != 0 {
		t.Errorf("expected 0 URLs, got %d", len(urls))
	}
}

func TestMemoryStorage_GetByUser_BatchSave(t *testing.T) {
	m := NewMemoryStorage()

	items := []BatchInput{
		{CorrelationID: "c1", OriginalURL: "https://x.com"},
		{CorrelationID: "c2", OriginalURL: "https://y.com"},
	}
	m.SaveBatch("carol", items)

	urls, err := m.GetByUser("carol")
	if err != nil {
		t.Fatalf("GetByUser: %v", err)
	}
	if len(urls) != 2 {
		t.Errorf("expected 2 URLs after SaveBatch, got %d", len(urls))
	}
}

func TestMemoryStorage_PingContext(t *testing.T) {
	m := NewMemoryStorage()
	if err := m.PingContext(context.Background()); err != nil {
		t.Errorf("PingContext returned unexpected error: %v", err)
	}
}
