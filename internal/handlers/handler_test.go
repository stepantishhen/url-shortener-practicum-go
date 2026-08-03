package handlers_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"url-shortener-practicum-go/internal/handlers"
	"url-shortener-practicum-go/internal/storage"
)

type mockStorage struct {
	data map[string]string
}

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

func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestShortenURL_ValidBody(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	w := httptest.NewRecorder()
	h.ShortenURL(w, r)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, res.StatusCode)
	}

	body, _ := io.ReadAll(res.Body)
	const want = "http://localhost:8080/testid12"
	if got := string(body); got != want {
		t.Errorf("expected body %q, got %q", want, got)
	}
}

func TestShortenURL_ContentType(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	w := httptest.NewRecorder()
	h.ShortenURL(w, r)

	ct := w.Result().Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected Content-Type text/plain, got %q", ct)
	}
}

func TestShortenURL_EmptyBody(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	w := httptest.NewRecorder()
	h.ShortenURL(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestShortenURL_WhitespaceOnlyBody(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("   \n  "))
	w := httptest.NewRecorder()
	h.ShortenURL(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestRedirect_KnownID(t *testing.T) {
	store := newMockStorage()
	store.data["testid12"] = "https://example.com/original"
	h := handlers.New(store, "http://localhost:8080", nil)

	r := httptest.NewRequest(http.MethodGet, "/testid12", nil)
	r = withURLParam(r, "id", "testid12")
	w := httptest.NewRecorder()
	h.Redirect(w, r)

	res := w.Result()
	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, res.StatusCode)
	}

	location := res.Header.Get("Location")
	if location != "https://example.com/original" {
		t.Errorf("expected Location %q, got %q", "https://example.com/original", location)
	}
}

func TestRedirect_UnknownID(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil)

	r := httptest.NewRequest(http.MethodGet, "/doesnotexist", nil)
	r = withURLParam(r, "id", "doesnotexist")
	w := httptest.NewRecorder()
	h.Redirect(w, r)

	if w.Result().StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Result().StatusCode)
	}
}

func TestShortenURLJSON_ValidBody(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil)

	body := strings.NewReader(`{"url":"https://example.com"}`)
	r := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ShortenURLJSON(w, r)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, res.StatusCode)
	}

	respBody, _ := io.ReadAll(res.Body)
	const want = `{"result":"http://localhost:8080/testid12"}`
	if got := strings.TrimSpace(string(respBody)); got != want {
		t.Errorf("expected body %q, got %q", want, got)
	}
}

func TestShortenURLJSON_ContentType(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil)

	body := strings.NewReader(`{"url":"https://example.com"}`)
	r := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ShortenURLJSON(w, r)

	ct := w.Result().Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

func TestShortenURLJSON_EmptyURL(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil)

	body := strings.NewReader(`{"url":""}`)
	r := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ShortenURLJSON(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestShortenURLJSON_InvalidJSON(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil)

	body := strings.NewReader(`not json`)
	r := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ShortenURLJSON(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestRedirect_EmptyID(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.Redirect(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestPing_NoDB(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil)

	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	h.Ping(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Result().StatusCode)
	}
}
