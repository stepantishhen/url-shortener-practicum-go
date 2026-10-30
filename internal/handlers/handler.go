package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"url-shortener-practicum-go/internal/audit"
	"url-shortener-practicum-go/internal/middleware"
	"url-shortener-practicum-go/internal/storage"
)

// DeleteService is the interface for submitting async URL deletions.
type DeleteService interface {
	Submit(userID string, ids []string)
}

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

type userURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// Handler holds the dependencies for the HTTP handlers.
type Handler struct {
	repo         storage.URLRepository
	baseURL      string
	pinger       storage.Pinger
	deleter      DeleteService
	auditEmitter *audit.Emitter
}

// New creates a new Handler with the given dependencies.
func New(repo storage.URLRepository, baseURL string, pinger storage.Pinger, deleter DeleteService, emitter *audit.Emitter) *Handler {
	return &Handler{repo: repo, baseURL: baseURL, pinger: pinger, deleter: deleter, auditEmitter: emitter}
}

// ShortenURL handles POST / — accepts a plain-text URL and returns its short form.
func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request: empty or invalid body", http.StatusBadRequest)
		return
	}
	originalURL := string(bytes.TrimSpace(body))
	if originalURL == "" {
		http.Error(w, "bad request: empty or invalid body", http.StatusBadRequest)
		return
	}

	id, err := h.repo.Save(middleware.UserID(r.Context()), originalURL)

	var conflictErr *storage.ConflictError
	if errors.As(err, &conflictErr) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusConflict)
		if _, werr := io.WriteString(w, h.baseURL+"/"+conflictErr.ShortID); werr != nil {
			log.Printf("ShortenURL: write conflict response: %v", werr)
		}
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	if _, werr := io.WriteString(w, h.baseURL+"/"+id); werr != nil {
		log.Printf("ShortenURL: write response: %v", werr)
	}
	h.auditEmitter.Emit(audit.Now("shorten", middleware.UserID(r.Context()), originalURL))
}

// ShortenURLJSON handles POST /api/shorten — accepts JSON {url} and returns JSON {result}.
func (h *Handler) ShortenURLJSON(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.URL) == "" {
		http.Error(w, "bad request: invalid or missing url field", http.StatusBadRequest)
		return
	}

	id, err := h.repo.Save(middleware.UserID(r.Context()), req.URL)

	var conflictErr *storage.ConflictError
	if errors.As(err, &conflictErr) {
		resp := shortenResponse{Result: h.baseURL + "/" + conflictErr.ShortID}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("ShortenURLJSON: write conflict response: %v", err)
		}
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := shortenResponse{Result: h.baseURL + "/" + id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("ShortenURLJSON: write response: %v", err)
	}
	h.auditEmitter.Emit(audit.Now("shorten", middleware.UserID(r.Context()), req.URL))
}

// Redirect handles GET /{id} — looks up the short ID and redirects to the original URL.
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "bad request: missing ID", http.StatusBadRequest)
		return
	}

	originalURL, found, deleted := h.repo.Get(id)
	if !found {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if deleted {
		http.Error(w, "gone", http.StatusGone)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
	h.auditEmitter.Emit(audit.Now("follow", middleware.UserID(r.Context()), originalURL))
}

// ShortenBatch handles POST /api/shorten/batch — shortens multiple URLs in one request.
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

	results, err := h.repo.SaveBatch(middleware.UserID(r.Context()), batch)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := make([]batchResponse, len(results))
	for i, res := range results {
		resp[i] = batchResponse{
			CorrelationID: res.CorrelationID,
			ShortURL:      h.baseURL + "/" + res.ShortID,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("ShortenBatch: write response: %v", err)
	}
}

// GetUserURLs handles GET /api/user/urls — returns all URLs shortened by the authenticated user.
func (h *Handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	cookieInvalid, _ := r.Context().Value(middleware.CookieInvalidKey).(bool)
	if cookieInvalid {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID := middleware.UserID(r.Context())
	urls, err := h.repo.GetByUser(userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]userURLResponse, len(urls))
	for i, u := range urls {
		resp[i] = userURLResponse{
			ShortURL:    h.baseURL + "/" + u.ShortID,
			OriginalURL: u.OriginalURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("GetUserURLs: write response: %v", err)
	}
}

// DeleteUserURLs handles DELETE /api/user/urls — schedules deletion of the given short IDs.
func (h *Handler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	var ids []string
	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil || len(ids) == 0 {
		http.Error(w, "bad request: invalid or empty list", http.StatusBadRequest)
		return
	}

	userID := middleware.UserID(r.Context())
	if h.deleter != nil {
		h.deleter.Submit(userID, ids)
		w.WriteHeader(http.StatusAccepted)
	} else {
		log.Println("DeleteUserURLs: deleter service is nil, cannot delete")
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// Ping handles GET /ping — checks database connectivity.
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if h.pinger == nil {
		http.Error(w, "database not configured", http.StatusInternalServerError)
		return
	}
	if err := h.pinger.PingContext(r.Context()); err != nil {
		http.Error(w, "database unavailable", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
