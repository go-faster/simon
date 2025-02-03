// Package server implements HTTP server.
package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/go-faster/simon/internal/oas"
)

func NewServer(tracerProvider trace.TracerProvider) *Server {
	return &Server{
		trace: tracerProvider.Tracer("simon.server"),
	}
}

// Server implements oas.Handler.
type Server struct {
	trace trace.Tracer
}

func (s Server) UploadFile(ctx context.Context, req *oas.UploadFileReq) (*oas.UploadResponse, error) {
	ctx, span := s.trace.Start(ctx, "Server.UploadFile")
	defer span.End()

	iterations := req.Iterations.Or(1)
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, req.File.File); err != nil {
		return nil, errors.Wrap(err, "copy")
	}

	zctx.From(ctx).Info("UploadFile",
		zap.Int("iterations", iterations),
		zap.Int("size", buf.Len()),
	)

	h := sha256.New()
	for i := 0; i < iterations; i++ {
		if _, err := h.Write(buf.Bytes()); err != nil {
			return nil, errors.Wrap(err, "write")
		}
	}

	return &oas.UploadResponse{
		Hash: fmt.Sprintf("%x", h.Sum(nil)),
	}, nil
}

var _ oas.Handler = (*Server)(nil)

func (s Server) Status(ctx context.Context) (*oas.Status, error) {
	ctx, span := s.trace.Start(ctx, "Server.Status")
	defer span.End()
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
