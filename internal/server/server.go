// Package server implements HTTP server.
package server

import (
	"context"

	"github.com/go-faster/simon/internal/oas"
	"github.com/go-faster/simon/sdk/zctx"
)

// Server implements oas.Handler.
type Server struct{}

var _ oas.Handler = (*Server)(nil)

func (s Server) Status(ctx context.Context) (*oas.Status, error) {
	zctx.From(ctx).Info("Status")
	return &oas.Status{Message: "ok"}, nil
}

func (s Server) NewError(_ context.Context, err error) *oas.ErrorStatusCode {
	return &oas.ErrorStatusCode{
		StatusCode: 500,
		Response: oas.Error{
			Message: err.Error(),
		},
	}
}
