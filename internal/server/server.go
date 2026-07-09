package server

import (
	"net/http"

	"url-shortener-practicum-go/internal/handlers"
)

// NewMux wires the handler methods to their routes.
func NewMux(h *handlers.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/":
			h.ShortenURL(w, r)
		case r.Method == http.MethodGet && r.URL.Path != "/":
			h.Redirect(w, r)
		default:
			http.Error(w, "bad request", http.StatusBadRequest)
		}
	})
	return mux
}
