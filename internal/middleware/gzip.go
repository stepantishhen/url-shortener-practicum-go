package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	gz          *gzip.Writer
	wroteHeader bool
	acceptGzip  bool
}

func (g *gzipResponseWriter) WriteHeader(code int) {
	if g.wroteHeader {
		return
	}
	g.wroteHeader = true

	if g.acceptGzip && isCompressible(g.Header().Get("Content-Type")) {
		gz, err := gzip.NewWriterLevel(g.ResponseWriter, gzip.BestSpeed)
		if err == nil {
			g.gz = gz
			g.Header().Set("Content-Encoding", "gzip")
			g.Header().Del("Content-Length")
		}
	}
	g.ResponseWriter.WriteHeader(code)
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if !g.wroteHeader {
		if g.Header().Get("Content-Type") == "" {
			g.Header().Set("Content-Type", http.DetectContentType(b))
		}
		g.WriteHeader(http.StatusOK)
	}
	if g.gz != nil {
		return g.gz.Write(b)
	}
	return g.ResponseWriter.Write(b)
}

func (g *gzipResponseWriter) close() {
	if g.gz != nil {
		g.gz.Close()
	}
}

func isCompressible(contentType string) bool {
	switch {
	case strings.HasPrefix(contentType, "application/json"),
		strings.HasPrefix(contentType, "text/html"):
		return true
	default:
		return false
	}
}

func Gzip() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Encoding") == "gzip" {
				gr, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "invalid gzip body", http.StatusBadRequest)
					return
				}
				defer gr.Close()
				r.Body = io.NopCloser(gr)
				r.Header.Del("Content-Encoding")
			}

			grw := &gzipResponseWriter{
				ResponseWriter: w,
				acceptGzip:     strings.Contains(r.Header.Get("Accept-Encoding"), "gzip"),
			}
			defer grw.close()

			next.ServeHTTP(grw, r)
		})
	}
}
