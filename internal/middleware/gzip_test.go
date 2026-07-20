package middleware_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"url-shortener-practicum-go/internal/middleware"
)

func jsonHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}
}

func htmlHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}
}

func textHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}
}

func echoBodyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}

func TestGzip_CompressesJSONResponse(t *testing.T) {
	const body = `{"result":"http://localhost:8080/testid12"}`

	handler := middleware.Gzip()(jsonHandler(body))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	res := w.Result()
	defer res.Body.Close()

	if res.Header.Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected Content-Encoding: gzip, got %q", res.Header.Get("Content-Encoding"))
	}

	rawBytes, _ := io.ReadAll(res.Body)

	if string(rawBytes) == body {
		t.Fatal("response body was not compressed: raw bytes equal original string")
	}

	gr, err := gzip.NewReader(bytes.NewReader(rawBytes))
	if err != nil {
		t.Fatalf("raw bytes are not valid gzip: %v", err)
	}
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to decompress response: %v", err)
	}

	if got := string(decompressed); got != body {
		t.Errorf("decompressed body mismatch: want %q, got %q", body, got)
	}
}

func TestGzip_CompressesHTMLResponse(t *testing.T) {
	const body = `<html><body><h1>Hello</h1></body></html>`

	handler := middleware.Gzip()(htmlHandler(body))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	res := w.Result()
	defer res.Body.Close()

	if res.Header.Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected Content-Encoding: gzip, got %q", res.Header.Get("Content-Encoding"))
	}

	rawBytes, _ := io.ReadAll(res.Body)

	if string(rawBytes) == body {
		t.Fatal("response body was not compressed: raw bytes equal original string")
	}

	gr, err := gzip.NewReader(bytes.NewReader(rawBytes))
	if err != nil {
		t.Fatalf("raw bytes are not valid gzip: %v", err)
	}
	defer gr.Close()

	decompressed, _ := io.ReadAll(gr)
	if got := string(decompressed); got != body {
		t.Errorf("decompressed body mismatch: want %q, got %q", body, got)
	}
}

func TestGzip_DoesNotCompressPlainText(t *testing.T) {
	const body = `hello world`

	handler := middleware.Gzip()(textHandler(body))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	res := w.Result()
	defer res.Body.Close()

	if enc := res.Header.Get("Content-Encoding"); enc != "" {
		t.Errorf("expected no Content-Encoding for text/plain, got %q", enc)
	}

	respBody, _ := io.ReadAll(res.Body)
	if got := string(respBody); got != body {
		t.Errorf("expected plain body %q, got %q", body, got)
	}
}

func TestGzip_NoCompressionWithoutAcceptEncoding(t *testing.T) {
	const body = `{"result":"http://localhost:8080/testid12"}`

	handler := middleware.Gzip()(jsonHandler(body))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	res := w.Result()
	defer res.Body.Close()

	if enc := res.Header.Get("Content-Encoding"); enc != "" {
		t.Errorf("expected no Content-Encoding without Accept-Encoding, got %q", enc)
	}

	respBody, _ := io.ReadAll(res.Body)
	if got := string(respBody); got != body {
		t.Errorf("expected plain body %q, got %q", body, got)
	}
}

func TestGzip_DecompressesRequestBody(t *testing.T) {
	const original = `{"url":"https://example.com"}`

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte(original))
	gz.Close()

	handler := middleware.Gzip()(echoBodyHandler())

	r := httptest.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}

	respBody, _ := io.ReadAll(res.Body)
	if got := string(respBody); got != original {
		t.Errorf("handler received wrong body: want %q, got %q", original, got)
	}
}

func TestGzip_InvalidGzipRequestBody(t *testing.T) {
	handler := middleware.Gzip()(jsonHandler("{}"))

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("this is not gzip data"))
	r.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid gzip body, got %d", w.Code)
	}
}
