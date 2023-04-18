package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-faster/sdk/zctx"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

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

// TraceMiddleware returns new instrumented middleware.
func (m *Metrics) TraceMiddleware() middleware.Middleware {
	var (
		h  = oas.Server{}
		p  = m.TextMapPropagator()
		tp = m.TracerProvider()
	)
	return func(next http.Handler) http.Handler {
		t := tp.Tracer("http")
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			start := time.Now()
			w := &writerProxy{ResponseWriter: rw}

			operation := "(Unknown)"
			route, routeOk := h.FindRoute(r.Method, r.URL.Path)
			if routeOk {
				operation = route.OperationID()
			}

			ctx := p.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			ctx, span := t.Start(ctx, fmt.Sprintf("HTTP: %s", operation))
			defer span.End()
			spanCtx := span.SpanContext()

			// Use separate loggers for request logger (to log in defer) and
			// context logger, because context logger should not contain
			// trace_id and span_id to be able to change them for new
			// spans.
			lgCtx := m.lg
			fields := []zap.Field{
				zap.Stringer("trace_id", spanCtx.TraceID()),
				zap.Stringer("span_id", spanCtx.SpanID()),
			}
			if routeOk {
				f := zap.String("op", operation)
				fields = append(fields, f)
				lgCtx = lgCtx.With(f)
			}

			ctx = zctx.Base(ctx, lgCtx)
			ctx = zctx.With(ctx, fields...)
			lgReq := zctx.From(ctx)

			defer func() {
				if r := recover(); r != nil {
					lgReq.Error("Panic", zap.Stack("stack"))
					if w.status == 0 {
						w.WriteHeader(http.StatusInternalServerError)
					}
					span.AddEvent("Panic recovered",
						trace.WithStackTrace(true),
					)
					span.SetStatus(codes.Error, "Panic recovered")
				}
				lgReq.Debug("Request",
					zap.Duration("duration", time.Since(start)),
					zap.Int("http.status", w.status),
					zap.Int64("http.response.size", w.wrote),
					zap.String("http.path", r.URL.String()),
					zap.String("http.method", r.Method),
				)
			}()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
