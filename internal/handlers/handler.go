package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"url-shortener-practicum-go/internal/storage"
)

// Handler holds dependencies for HTTP request handlers.
type Handler struct {
	repo    storage.URLRepository
	baseURL string
}

func New(repo storage.URLRepository, baseURL string) *Handler {
	return &Handler{repo: repo, baseURL: baseURL}
}

// ShortenURL handles POST / — stores a URL and returns its short form.
func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(strings.TrimSpace(string(body))) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(string(body))

	id, err := h.repo.Save(originalURL)
	if err != nil {
		http.Error(w, "internal error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", h.baseURL, id)
}

// Redirect handles GET /{id} — redirects to the original URL.
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")
	if id == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	originalURL, ok := h.repo.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
