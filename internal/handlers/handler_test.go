package handlers_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"url-shortener-practicum-go/internal/handlers"
	"url-shortener-practicum-go/internal/middleware"
	"url-shortener-practicum-go/internal/storage"
)

type mockStorage struct {
	data     map[string]string
	userURLs map[string][]storage.UserURL
	deleted  map[string]bool
}

var _ storage.URLRepository = (*mockStorage)(nil)

func newMockStorage() *mockStorage {
	return &mockStorage{
		data:     make(map[string]string),
		userURLs: make(map[string][]storage.UserURL),
		deleted:  make(map[string]bool),
	}
}

func (m *mockStorage) Save(_, originalURL string) (string, error) {
	const fixedID = "testid12"
	m.data[fixedID] = originalURL
	return fixedID, nil
}

func (m *mockStorage) Get(id string) (string, bool, bool) {
	url, ok := m.data[id]
	return url, ok, m.deleted[id]
}

func (m *mockStorage) SaveBatch(_ string, items []storage.BatchInput) ([]storage.BatchOutput, error) {
	results := make([]storage.BatchOutput, len(items))
	for i, item := range items {
		id := fmt.Sprintf("batchid%02d", i)
		m.data[id] = item.OriginalURL
		results[i] = storage.BatchOutput{CorrelationID: item.CorrelationID, ShortID: id}
	}
	return results, nil
}

func (m *mockStorage) GetByUser(userID string) ([]storage.UserURL, error) {
	return m.userURLs[userID], nil
}

func (m *mockStorage) DeleteBatch(_ string, ids []string) error {
	for _, id := range ids {
		m.deleted[id] = true
	}
	return nil
}

// conflictMockStorage always returns ConflictError with a fixed existing ID.
type conflictMockStorage struct{ *mockStorage }

func (c *conflictMockStorage) Save(_, _ string) (string, error) {
	return "", &storage.ConflictError{ShortID: "existing1"}
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func withAuthCtx(r *http.Request, userID string, cookieInvalid bool) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.CookieInvalidKey, cookieInvalid)
	return r.WithContext(ctx)
}

func TestShortenURL_ValidBody(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

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
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	w := httptest.NewRecorder()
	h.ShortenURL(w, r)

	ct := w.Result().Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected Content-Type text/plain, got %q", ct)
	}
}

func TestShortenURL_EmptyBody(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	w := httptest.NewRecorder()
	h.ShortenURL(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestShortenURL_WhitespaceOnlyBody(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

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
	h := handlers.New(store, "http://localhost:8080", nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/testid12", nil)
	r = withURLParam(r, "id", "testid12")
	w := httptest.NewRecorder()
	h.Redirect(w, r)

	res := w.Result()
	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, res.StatusCode)
	}
	if location := res.Header.Get("Location"); location != "https://example.com/original" {
		t.Errorf("expected Location %q, got %q", "https://example.com/original", location)
	}
}

func TestRedirect_UnknownID(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/doesnotexist", nil)
	r = withURLParam(r, "id", "doesnotexist")
	w := httptest.NewRecorder()
	h.Redirect(w, r)

	if w.Result().StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Result().StatusCode)
	}
}

func TestRedirect_Gone(t *testing.T) {
	store := newMockStorage()
	store.data["deleted1"] = "https://example.com"
	store.deleted["deleted1"] = true
	h := handlers.New(store, "http://localhost:8080", nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/deleted1", nil)
	r = withURLParam(r, "id", "deleted1")
	w := httptest.NewRecorder()
	h.Redirect(w, r)

	if w.Result().StatusCode != http.StatusGone {
		t.Errorf("expected 410 Gone, got %d", w.Result().StatusCode)
	}
}

func TestShortenURLJSON_ValidBody(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

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
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

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
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

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
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

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
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.Redirect(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestPing_NoDB(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	h.Ping(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Result().StatusCode)
	}
}

func TestShortenBatch_Valid(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

	body := strings.NewReader(`[
		{"correlation_id":"id1","original_url":"https://example.com/1"},
		{"correlation_id":"id2","original_url":"https://example.com/2"}
	]`)
	r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ShortenBatch(w, r)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	respBody, _ := io.ReadAll(res.Body)
	s := string(respBody)
	for _, want := range []string{
		`"correlation_id":"id1"`,
		`"correlation_id":"id2"`,
		`"short_url":"http://localhost:8080/batchid00"`,
		`"short_url":"http://localhost:8080/batchid01"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("response missing %q: %s", want, s)
		}
	}
}

func TestShortenBatch_EmptyBatch(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(`[]`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ShortenBatch(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestShortenURL_Conflict(t *testing.T) {
	h := handlers.New(&conflictMockStorage{newMockStorage()}, "http://localhost:8080", nil, nil)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	w := httptest.NewRecorder()
	h.ShortenURL(w, r)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	const want = "http://localhost:8080/existing1"
	if got := strings.TrimSpace(string(body)); got != want {
		t.Errorf("expected body %q, got %q", want, got)
	}
}

func TestShortenURLJSON_Conflict(t *testing.T) {
	h := handlers.New(&conflictMockStorage{newMockStorage()}, "http://localhost:8080", nil, nil)

	body := strings.NewReader(`{"url":"https://example.com"}`)
	r := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ShortenURLJSON(w, r)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, res.StatusCode)
	}
	respBody, _ := io.ReadAll(res.Body)
	const want = `{"result":"http://localhost:8080/existing1"}`
	if got := strings.TrimSpace(string(respBody)); got != want {
		t.Errorf("expected body %q, got %q", want, got)
	}
}

func TestShortenBatch_InvalidJSON(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(`not json`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ShortenBatch(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestGetUserURLs_InvalidCookie_Returns401(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	r = withAuthCtx(r, "new-generated-id", true)
	w := httptest.NewRecorder()
	h.GetUserURLs(w, r)

	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Result().StatusCode)
	}
}

func TestGetUserURLs_NoURLs_Returns204(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	r = withAuthCtx(r, "user-with-no-urls", false)
	w := httptest.NewRecorder()
	h.GetUserURLs(w, r)

	if w.Result().StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Result().StatusCode)
	}
}

func TestGetUserURLs_WithURLs_Returns200(t *testing.T) {
	const userID = "user-abc"
	store := newMockStorage()
	store.userURLs[userID] = []storage.UserURL{
		{ShortID: "abc12345", OriginalURL: "https://example.com"},
		{ShortID: "def67890", OriginalURL: "https://go.dev"},
	}
	h := handlers.New(store, "http://localhost:8080", nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	r = withAuthCtx(r, userID, false)
	w := httptest.NewRecorder()
	h.GetUserURLs(w, r)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	s := string(body)
	for _, want := range []string{
		`"short_url":"http://localhost:8080/abc12345"`,
		`"original_url":"https://example.com"`,
		`"short_url":"http://localhost:8080/def67890"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("response missing %q: %s", want, s)
		}
	}
}

func TestDeleteUserURLs_Returns202(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

	body := strings.NewReader(`["abc12345","def67890"]`)
	r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", body)
	r.Header.Set("Content-Type", "application/json")
	r = withAuthCtx(r, "user1", false)
	w := httptest.NewRecorder()
	h.DeleteUserURLs(w, r)

	if w.Result().StatusCode != http.StatusAccepted {
		t.Errorf("expected 202 Accepted, got %d", w.Result().StatusCode)
	}
}

func TestDeleteUserURLs_EmptyList_Returns400(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil)

	r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`[]`))
	r.Header.Set("Content-Type", "application/json")
	r = withAuthCtx(r, "user1", false)
	w := httptest.NewRecorder()
	h.DeleteUserURLs(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Result().StatusCode)
	}
}
