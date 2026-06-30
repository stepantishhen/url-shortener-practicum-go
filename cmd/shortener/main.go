package main

import (
	"net/http"

	"url-shortener-practicum-go/internal/handlers"
	"url-shortener-practicum-go/internal/server"
	"url-shortener-practicum-go/internal/storage"
)

func main() {
	repo := storage.NewMemoryStorage()
	h := handlers.New(repo, "http://localhost:8080")
	http.ListenAndServe(":8080", server.NewRouter(h))
}
