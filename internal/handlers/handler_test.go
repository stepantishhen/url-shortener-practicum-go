package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"url-shortener-practicum-go/internal/handlers"
	"url-shortener-practicum-go/internal/storage"
)

// mockStorage is an in-test fake that implements storage.URLRepository.
type mockStorage struct {
	data map[string]string
}

// Compile-time check that mockStorage satisfies the interface.
var _ storage.URLRepository = (*mockStorage)(nil)

func newMockStorage() *mockStorage {
	return &mockStorage{data: make(map[string]string)}
}

func (m *mockStorage) Save(originalURL string) (string, error) {
	const fixedID = "testid12"
	m.data[fixedID] = originalURL
	return fixedID, nil
}

func (m *mockStorage) Get(id string) (string, bool) {
	url, ok := m.data[id]
	return url, ok
}

// ---- ShortenURL ----

func TestShortenURL_ValidBody(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080")

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	w := httptest.NewRecorder()

	h.ShortenURL(w, r)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("expected %d, got %d", http.StatusCreated, res.StatusCode)
	}

	body, _ := io.ReadAll(res.Body)
	got := string(body)
	const want = "http://localhost:8080/testid12"
	if got != want {
		t.Errorf("expected body %q, got %q", want, got)
	}
}

func TestShortenURL_ContentType(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080")

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	w := httptest.NewRecorder()

	h.ShortenURL(w, r)

	ct := w.Result().Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected Content-Type text/plain, got %q", ct)
	}
}

func TestShortenURL_EmptyBody(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080")

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	w := httptest.NewRecorder()

	h.ShortenURL(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestShortenURL_WhitespaceOnlyBody(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080")

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("   \n  "))
	w := httptest.NewRecorder()

	h.ShortenURL(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, w.Result().StatusCode)
	}
}

// ---- Redirect ----

func TestRedirect_KnownID(t *testing.T) {
	store := newMockStorage()
	store.data["testid12"] = "https://example.com"
	h := handlers.New(store, "http://localhost:8080")

	r := httptest.NewRequest(http.MethodGet, "/testid12", nil)
	w := httptest.NewRecorder()

	h.Redirect(w, r)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("expected %d, got %d", http.StatusTemporaryRedirect, res.StatusCode)
	}

	location := res.Header.Get("Location")
	if location != "https://example.com" {
		t.Errorf("expected Location %q, got %q", "https://example.com", location)
	}
}

func TestRedirect_UnknownID(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080")

	r := httptest.NewRequest(http.MethodGet, "/doesnotexist", nil)
	w := httptest.NewRecorder()

	h.Redirect(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestRedirect_EmptyID(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080")

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.Redirect(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, w.Result().StatusCode)
	}
}
