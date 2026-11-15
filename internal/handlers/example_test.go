package handlers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"url-shortener-practicum-go/internal/audit"
	"url-shortener-practicum-go/internal/handlers"
	"url-shortener-practicum-go/internal/storage"
)

// ExampleHandler_ShortenURL demonstrates shortening a URL via the plain-text endpoint.
func ExampleHandler_ShortenURL() {
	store := storage.NewMemoryStorage()
	h := handlers.New(store, "http://localhost:8080", nil, nil, audit.NewEmitter())

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com/very/long/path"))
	w := httptest.NewRecorder()
	h.ShortenURL(w, r)

	res := w.Result()
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(strings.HasPrefix(string(body), "http://localhost:8080/"))
	// Output:
	// 201
	// true
}

// ExampleHandler_ShortenURLJSON demonstrates shortening a URL via the JSON endpoint.
func ExampleHandler_ShortenURLJSON() {
	store := storage.NewMemoryStorage()
	h := handlers.New(store, "http://localhost:8080", nil, nil, audit.NewEmitter())

	body := strings.NewReader(`{"url":"https://example.com/page"}`)
	r := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ShortenURLJSON(w, r)

	res := w.Result()
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)

	var resp struct {
		Result string `json:"result"`
	}
	json.Unmarshal(raw, &resp)

	fmt.Println(res.StatusCode)
	fmt.Println(strings.HasPrefix(resp.Result, "http://localhost:8080/"))
	// Output:
	// 201
	// true
}

// ExampleHandler_Redirect demonstrates following a short URL.
func ExampleHandler_Redirect() {
	store := storage.NewMemoryStorage()
	id, _ := store.Save("user1", "https://example.com/original")
	h := handlers.New(store, "http://localhost:8080", nil, nil, audit.NewEmitter())

	r := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	r = withURLParam(r, "id", id)
	w := httptest.NewRecorder()
	h.Redirect(w, r)

	res := w.Result()
	fmt.Println(res.StatusCode)
	fmt.Println(res.Header.Get("Location"))
	// Output:
	// 307
	// https://example.com/original
}

// ExampleHandler_ShortenBatch demonstrates batch URL shortening.
func ExampleHandler_ShortenBatch() {
	store := storage.NewMemoryStorage()
	h := handlers.New(store, "http://localhost:8080", nil, nil, audit.NewEmitter())

	payload := `[
		{"correlation_id":"a","original_url":"https://example.com/1"},
		{"correlation_id":"b","original_url":"https://example.com/2"}
	]`
	r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(payload))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ShortenBatch(w, r)

	res := w.Result()
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)

	var items []struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}
	json.Unmarshal(raw, &items)

	fmt.Println(res.StatusCode)
	fmt.Println(len(items))
	// Output:
	// 201
	// 2
}

// ExampleHandler_Ping demonstrates the health-check endpoint without a database.
func ExampleHandler_Ping() {
	h := handlers.New(storage.NewMemoryStorage(), "http://localhost:8080", nil, nil, audit.NewEmitter())

	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	h.Ping(w, r)

	fmt.Println(w.Result().StatusCode)
	// Output:
	// 500
}
