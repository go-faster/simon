// Package middleware is http middleware.
package middleware

import "net/http"

// Middleware is http middleware.
type Middleware func(next http.Handler) http.Handler

// Wrap Middleware.
func Wrap(h http.Handler, mw Middleware) http.Handler {
	return mw(h)
}
