package server

import (
	"github.com/valyala/fasthttp"
	"url-shortener-practicum-go/internal/handlers"
)

func NewRequestHandler(h *handlers.Handler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		switch {
		case ctx.IsPost() && string(ctx.Path()) == "/":
			h.ShortenURL(ctx)
		case ctx.IsGet() && string(ctx.Path()) != "/":
			h.Redirect(ctx)
		default:
			ctx.Error("bad request", fasthttp.StatusBadRequest)
		}
	}
}
