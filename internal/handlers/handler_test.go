package handlers_test

import (
	"testing"

	"github.com/valyala/fasthttp"
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

func newPOSTCtx(body string) *fasthttp.RequestCtx {
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(fasthttp.MethodPost)
	ctx.Request.SetRequestURI("/")
	ctx.Request.SetBodyString(body)
	return ctx
}

func newGETCtx(path string) *fasthttp.RequestCtx {
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(fasthttp.MethodGet)
	ctx.Request.SetRequestURI(path)
	return ctx
}

func TestShortenURL_ValidBody(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080")

	ctx := newPOSTCtx("https://example.com")
	h.ShortenURL(ctx)

	if got := ctx.Response.StatusCode(); got != fasthttp.StatusCreated {
		t.Errorf("expected status %d, got %d", fasthttp.StatusCreated, got)
	}

	got := string(ctx.Response.Body())
	const want = "http://localhost:8080/testid12"
	if got != want {
		t.Errorf("expected body %q, got %q", want, got)
	}
}

func TestShortenURL_ContentType(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080")

	ctx := newPOSTCtx("https://example.com")
	h.ShortenURL(ctx)

	ct := string(ctx.Response.Header.Peek("Content-Type"))
	if ct != "text/plain" {
		t.Errorf("expected Content-Type text/plain, got %q", ct)
	}
}

func TestShortenURL_EmptyBody(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080")

	ctx := newPOSTCtx("")
	h.ShortenURL(ctx)

	if got := ctx.Response.StatusCode(); got != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, got)
	}
}

func TestShortenURL_WhitespaceOnlyBody(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080")

	ctx := newPOSTCtx("   \n  ")
	h.ShortenURL(ctx)

	if got := ctx.Response.StatusCode(); got != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, got)
	}
}

func TestRedirect_KnownID(t *testing.T) {
	const originalURL = "https://example.com/original"
	store := newMockStorage()
	store.data["testid12"] = originalURL
	h := handlers.New(store, "http://localhost:8080")

	ctx := newGETCtx("/testid12")
	h.Redirect(ctx)

	if got := ctx.Response.StatusCode(); got != fasthttp.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", fasthttp.StatusTemporaryRedirect, got)
	}

	location := string(ctx.Response.Header.Peek("Location"))
	if location != originalURL {
		t.Errorf("expected Location %q, got %q", originalURL, location)
	}
}

func TestRedirect_UnknownID(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080")

	ctx := newGETCtx("/doesnotexist")
	h.Redirect(ctx)

	if got := ctx.Response.StatusCode(); got != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, got)
	}
}

func TestRedirect_EmptyID(t *testing.T) {
	h := handlers.New(newMockStorage(), "http://localhost:8080")

	ctx := newGETCtx("/")
	h.Redirect(ctx)

	if got := ctx.Response.StatusCode(); got != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, got)
	}
}
