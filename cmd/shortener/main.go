package main

import (
	"github.com/valyala/fasthttp"
	"url-shortener-practicum-go/internal/handlers"
	"url-shortener-practicum-go/internal/server"
	"url-shortener-practicum-go/internal/storage"
)

func main() {
	repo := storage.NewMemoryStorage()
	h := handlers.New(repo, "http://localhost:8080")
	fasthttp.ListenAndServe(":8080", server.NewRequestHandler(h))
}
