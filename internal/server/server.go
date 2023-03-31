// Package server implements HTTP server.
package server

import (
	"context"

	"github.com/go-faster/simon/internal/oas"
)

// Server implements oas.Handler.
type Server struct{}

var _ oas.Handler = (*Server)(nil)

func (s Server) Status(ctx context.Context) (*oas.Status, error) {
	return &oas.Status{
		Message: "ok",
	}, nil

}

func (s Server) NewError(ctx context.Context, err error) *oas.ErrorStatusCode {
	return &oas.ErrorStatusCode{
		StatusCode: 500,
		Response: oas.Error{
			Message: err.Error(),
		},
	}
}
