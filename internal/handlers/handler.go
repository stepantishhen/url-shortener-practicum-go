package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"url-shortener-practicum-go/internal/storage"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

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
		http.Error(w, "bad request: empty or invalid body", http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(string(body))
	id, err := h.repo.Save(originalURL)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", h.baseURL, id)
}

func (h *Handler) ShortenURLJSON(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.URL) == "" {
		http.Error(w, "bad request: invalid or missing url field", http.StatusBadRequest)
		return
	}

	id, err := h.repo.Save(req.URL)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := shortenResponse{Result: fmt.Sprintf("%s/%s", h.baseURL, id)}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "bad request: missing ID", http.StatusBadRequest)
		return
	}

	originalURL, ok := h.repo.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
