package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

type batchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type batchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type Handler struct {
	repo    storage.URLRepository
	baseURL string
	db      *sql.DB
}

func New(repo storage.URLRepository, baseURL string, db *sql.DB) *Handler {
	return &Handler{repo: repo, baseURL: baseURL, db: db}
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
	if _, err := fmt.Fprintf(w, "%s/%s", h.baseURL, id); err != nil {
		log.Printf("ShortenURL: write response: %v", err)
	}
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
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("ShortenURLJSON: write response: %v", err)
	}
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

func (h *Handler) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	var items []batchRequest
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil || len(items) == 0 {
		http.Error(w, "bad request: invalid or empty batch", http.StatusBadRequest)
		return
	}

	batch := make([]storage.BatchInput, len(items))
	for i, item := range items {
		batch[i] = storage.BatchInput{
			CorrelationID: item.CorrelationID,
			OriginalURL:   item.OriginalURL,
		}
	}

	results, err := h.repo.SaveBatch(batch)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := make([]batchResponse, len(results))
	for i, res := range results {
		resp[i] = batchResponse{
			CorrelationID: res.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", h.baseURL, res.ShortID),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("ShortenBatch: write response: %v", err)
	}
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		http.Error(w, "database not configured", http.StatusInternalServerError)
		return
	}
	if err := h.db.PingContext(r.Context()); err != nil {
		http.Error(w, "database unavailable", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
