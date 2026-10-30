package audit

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

func TestEmitter_NoListeners(t *testing.T) {
	emitter := NewEmitter()
	// Should not panic with no listeners.
	emitter.Emit(Event{Action: "shorten", URL: "https://example.com"})
}

func TestEmitter_AddAndNotify(t *testing.T) {
	emitter := NewEmitter()

	var received []Event
	listener := &mockListener{
		notify: func(e Event) {
			received = append(received, e)
		},
	}
	emitter.AddListener(listener)

	emitter.Emit(Event{Action: "shorten", URL: "https://example.com"})
	emitter.Emit(Event{Action: "follow", URL: "https://go.dev"})

	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}
	if received[0].Action != "shorten" {
		t.Errorf("expected first action 'shorten', got %q", received[0].Action)
	}
	if received[1].Action != "follow" {
		t.Errorf("expected second action 'follow', got %q", received[1].Action)
	}
}

func TestEmitter_MultipleListeners(t *testing.T) {
	emitter := NewEmitter()

	var received1, received2 []Event
	l1 := &mockListener{
		notify: func(e Event) { received1 = append(received1, e) },
	}
	l2 := &mockListener{
		notify: func(e Event) { received2 = append(received2, e) },
	}
	emitter.AddListener(l1)
	emitter.AddListener(l2)

	emitter.Emit(Event{Action: "shorten", URL: "https://example.com"})

	if len(received1) != 1 {
		t.Errorf("listener1: expected 1 event, got %d", len(received1))
	}
	if len(received2) != 1 {
		t.Errorf("listener2: expected 1 event, got %d", len(received2))
	}
}

func TestEmitter_ConcurrentAccess(t *testing.T) {
	emitter := NewEmitter()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			listener := &mockListener{
				notify: func(e Event) {},
			}
			emitter.AddListener(listener)
		}()
	}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			emitter.Emit(Event{Action: "test", URL: "https://example.com", Timestamp: int64(i)})
		}(i)
	}
	wg.Wait()
}

func TestFileListener(t *testing.T) {
	tmpFile := t.TempDir() + "/audit.log"

	fl, err := NewFileListener(tmpFile)
	if err != nil {
		t.Fatalf("failed to create file listener: %v", err)
	}
	defer fl.Close()

	event := Event{Action: "shorten", URL: "https://example.com", Timestamp: 12345}
	fl.Notify(event)

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read audit file: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal audit event: %v", err)
	}
	if decoded.Action != "shorten" {
		t.Errorf("expected action 'shorten', got %q", decoded.Action)
	}
	if decoded.URL != "https://example.com" {
		t.Errorf("expected URL 'https://example.com', got %q", decoded.URL)
	}
}

func TestFileListener_Append(t *testing.T) {
	tmpFile := t.TempDir() + "/audit.log"

	fl, err := NewFileListener(tmpFile)
	if err != nil {
		t.Fatalf("failed to create file listener: %v", err)
	}
	defer fl.Close()

	fl.Notify(Event{Action: "shorten", URL: "https://example.com"})
	fl.Notify(Event{Action: "follow", URL: "https://go.dev"})

	lines, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read audit file: %v", err)
	}

	// Should have two lines (json.Encoder.Encode adds a newline).
	lineCount := 0
	for _, b := range lines {
		if b == '\n' {
			lineCount++
		}
	}
	if lineCount != 2 {
		t.Errorf("expected 2 lines, got %d", lineCount)
	}
}

func TestNow(t *testing.T) {
	var before int64
	for i := 0; i < 10; i++ {
		e := Now("shorten", "user1", "https://example.com")
		if e.Timestamp < before {
			t.Errorf("timestamp should not decrease: %d < %d", e.Timestamp, before)
		}
		before = e.Timestamp
		if e.Action != "shorten" {
			t.Errorf("expected action 'shorten', got %q", e.Action)
		}
		if e.UserID != "user1" {
			t.Errorf("expected user_id 'user1', got %q", e.UserID)
		}
		if e.URL != "https://example.com" {
			t.Errorf("expected URL 'https://example.com', got %q", e.URL)
		}
	}
}

type mockListener struct {
	notify func(Event)
}

func (m *mockListener) Notify(e Event) {
	m.notify(e)
}

func TestHTTPListener_Success(t *testing.T) {
	var received Event
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if err := json.Unmarshal(body, &received); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	hl := NewHTTPListener(server.URL)
	event := Event{Action: "shorten", UserID: "user1", URL: "https://example.com"}
	hl.Notify(event)

	// Notify is async, wait for the goroutine to finish.
	time.Sleep(100 * time.Millisecond)

	if received.Action != "shorten" {
		t.Errorf("expected action 'shorten', got %q", received.Action)
	}
	if received.UserID != "user1" {
		t.Errorf("expected user_id 'user1', got %q", received.UserID)
	}
	if received.URL != "https://example.com" {
		t.Errorf("expected URL 'https://example.com', got %q", received.URL)
	}
}

func TestHTTPListener_ConnectionRefused(t *testing.T) {
	// Use a URL that will definitely fail to connect.
	hl := NewHTTPListener("http://127.0.0.1:1")
	event := Event{Action: "follow", URL: "https://example.com"}

	// Should not panic; error should be logged.
	hl.Notify(event)
	time.Sleep(100 * time.Millisecond)
}
