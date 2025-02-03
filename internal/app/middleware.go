package app

import (
	"net/http"

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
