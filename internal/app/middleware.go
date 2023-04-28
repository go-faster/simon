package app

import (
	"net/http"

	"github.com/go-faster/sdk/zctx"

	"github.com/go-faster/simon/internal/middleware"
	"github.com/go-faster/simon/internal/oas"
)

type writerProxy struct {
	http.ResponseWriter

	wrote  int64
	status int
}

func (w *writerProxy) Write(bytes []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(bytes)
	w.wrote += int64(n)
	return n, err
}
func (w *writerProxy) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.status = statusCode
}

type Router interface {
	FindRoute(method, path string) (oas.Route, bool)
}

func NewSpanNameFormatter(h Router) func(operation string, r *http.Request) string {
	return func(operation string, r *http.Request) string {
		route, ok := h.FindRoute(r.Method, r.URL.Path)
		if !ok {
			return operation
		}
		return route.OperationID()
	}
}

// LogMiddleware adds logger via zctx.With to request context.
func (m *Metrics) LogMiddleware() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context()
			req := r.WithContext(zctx.Base(reqCtx, m.lg))
			next.ServeHTTP(w, req)
		})
	}
}
