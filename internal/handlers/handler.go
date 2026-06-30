package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"url-shortener-practicum-go/internal/storage"
)

type Handler struct {
	repo    storage.URLRepository
	baseURL string
}

func New(repo storage.URLRepository, baseURL string) *Handler {
	return &Handler{repo: repo, baseURL: baseURL}
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(strings.TrimSpace(string(body))) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	id, err := h.repo.Save(strings.TrimSpace(string(body)))
	if err != nil {
		http.Error(w, "internal error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", h.baseURL, id)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
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
