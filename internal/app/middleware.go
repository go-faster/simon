package app

import (
	"net/http"

	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"

	"github.com/go-faster/simon/internal/middleware"
	"github.com/go-faster/simon/internal/oas"
)

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
func LogMiddleware(lg *zap.Logger) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context()
			reqCtx = zctx.WithOpenTelemetryZap(reqCtx)
			req := r.WithContext(zctx.Base(reqCtx, lg))
			next.ServeHTTP(w, req)
		})
	}
}
