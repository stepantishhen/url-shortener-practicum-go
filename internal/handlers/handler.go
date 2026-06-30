package handlers

import (
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"
	"url-shortener-practicum-go/internal/storage"
)

type Handler struct {
	repo    storage.URLRepository
	baseURL string
}

func New(repo storage.URLRepository, baseURL string) *Handler {
	return &Handler{repo: repo, baseURL: baseURL}
}

func (h *Handler) ShortenURL(ctx *fasthttp.RequestCtx) {
	body := strings.TrimSpace(string(ctx.PostBody()))
	if body == "" {
		ctx.Error("bad request", fasthttp.StatusBadRequest)
		return
	}

	id, err := h.repo.Save(body)
	if err != nil {
		ctx.Error("internal error", fasthttp.StatusBadRequest)
		return
	}

	ctx.Response.Header.Set("Content-Type", "text/plain")
	ctx.SetStatusCode(fasthttp.StatusCreated)
	fmt.Fprintf(ctx, "%s/%s", h.baseURL, id)
}

func (h *Handler) Redirect(ctx *fasthttp.RequestCtx) {
	id := strings.TrimPrefix(string(ctx.Path()), "/")
	if id == "" {
		ctx.Error("bad request", fasthttp.StatusBadRequest)
		return
	}

	originalURL, ok := h.repo.Get(id)
	if !ok {
		ctx.Error("not found", fasthttp.StatusBadRequest)
		return
	}

	ctx.Redirect(originalURL, fasthttp.StatusTemporaryRedirect)
}
