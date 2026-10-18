package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zaptest"
	"url-shortener-practicum-go/internal/middleware"
)

func TestLogger_PassesThroughResponse(t *testing.T) {
	log := zaptest.NewLogger(t)

	handler := middleware.Logger(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("hello"))
	}))

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	res := w.Result()
	if res.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", res.StatusCode)
	}
	if w.Body.String() != "hello" {
		t.Errorf("expected body %q, got %q", "hello", w.Body.String())
	}
}

func TestLogger_DefaultStatus200(t *testing.T) {
	log := zaptest.NewLogger(t)

	handler := middleware.Logger(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("expected default status 200, got %d", w.Result().StatusCode)
	}
}
