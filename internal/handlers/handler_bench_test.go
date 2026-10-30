package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"url-shortener-practicum-go/internal/audit"
	"url-shortener-practicum-go/internal/handlers"
	"url-shortener-practicum-go/internal/storage"
)

func BenchmarkShortenURL(b *testing.B) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil, audit.NewEmitter())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com/page"))
		w := httptest.NewRecorder()
		h.ShortenURL(w, r)
	}
}

func BenchmarkShortenURLJSON(b *testing.B) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil, audit.NewEmitter())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := strings.NewReader(`{"url":"https://example.com/page"}`)
		r := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ShortenURLJSON(w, r)
	}
}

func BenchmarkRedirect(b *testing.B) {
	store := newMockStorage()
	store.data["testid12"] = "https://example.com/original"
	h := handlers.New(store, "http://localhost:8080", nil, nil, audit.NewEmitter())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodGet, "/testid12", nil)
		r = withURLParam(r, "id", "testid12")
		w := httptest.NewRecorder()
		h.Redirect(w, r)
	}
}

func BenchmarkShortenBatch(b *testing.B) {
	h := handlers.New(newMockStorage(), "http://localhost:8080", nil, nil, audit.NewEmitter())
	parts := make([]string, 20)
	for i := 0; i < 20; i++ {
		parts[i] = fmt.Sprintf(`{"correlation_id":"id%d","original_url":"https://example.com/%d"}`, i, i)
	}
	body := "[" + strings.Join(parts, ",") + "]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ShortenBatch(w, r)
	}
}

func BenchmarkGetUserURLs(b *testing.B) {
	const userID = "bench-user"
	store := newMockStorage()
	for i := 0; i < 50; i++ {
		store.userURLs[userID] = append(store.userURLs[userID], storage.UserURL{
			ShortID:     fmt.Sprintf("id%06d", i),
			OriginalURL: fmt.Sprintf("https://example.com/%d", i),
		})
	}
	h := handlers.New(store, "http://localhost:8080", nil, nil, audit.NewEmitter())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		r = withAuthCtx(r, userID, false)
		w := httptest.NewRecorder()
		h.GetUserURLs(w, r)
	}
}
