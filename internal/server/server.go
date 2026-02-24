// Package server implements HTTP server.
package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"go.opentelemetry.io/otel/attribute"
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

func (s Server) getEnvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func (s Server) makeExternalRequest(ctx context.Context) error {
	// Make external request.
	ctx, span := s.trace.Start(ctx, "Server.makeExternalRequest")
	defer span.End()

	uri := s.getEnvDefault("EXTERNAL_URL", "https://www.google.com/")

	req, err := http.NewRequestWithContext(ctx, "GET", uri, http.NoBody)
	if err != nil {
		return errors.Wrap(err, "create external request")
	}

	span.AddEvent("Starting external request",
		trace.WithAttributes(
			attribute.String("url", req.URL.String()),
			attribute.String("method", req.Method),
		),
	)

	resp, err := http.DefaultClient.Do(req) // #nosec G704
	if err != nil {
		return errors.Wrap(err, "do external request")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		return errors.Wrap(err, "read external response")
	}

	zctx.From(ctx).Info("Request: external",
		zap.Int("status", resp.StatusCode),
		zap.Int("size", len(data)),
		zap.ByteString("data", data),
	)

	return nil
}

func (s Server) makeCurlRequest(ctx context.Context) error {
	ctx, span := s.trace.Start(ctx, "Server.makeCurlRequest")
	defer span.End()

	uri := s.getEnvDefault("CURL_URL", "https://ifconfig.me")

	bufErr := new(bytes.Buffer)
	buf := new(bytes.Buffer)
	cmd := exec.CommandContext(ctx, "curl", "-s", uri, "-o", "-", "--max-time", "5") // #nosec G204
	cmd.Stdout = buf
	cmd.Stderr = bufErr

	span.AddEvent("Starting curl command",
		trace.WithAttributes(
			attribute.StringSlice("args", cmd.Args),
		),
	)

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "curl: %s", bufErr.String())
	}

	zctx.From(ctx).Info("Request: curl",
		zap.Int("size", buf.Len()),
		zap.ByteString("data", buf.Bytes()),
	)

	return nil
}

func (s Server) makeShellCommand(ctx context.Context) error {
	ctx, span := s.trace.Start(ctx, "Server.makeShellCommand")
	defer span.End()

	bufErr := new(bytes.Buffer)
	buf := new(bytes.Buffer)
	cmd := exec.CommandContext(ctx, "sh", "-c", "echo hello && sleep 1 && echo world")
	cmd.Stdout = buf
	cmd.Stderr = bufErr

	span.AddEvent("Starting shell command",
		trace.WithAttributes(
			attribute.StringSlice("args", cmd.Args),
		),
	)

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "shell: %s", bufErr.String())
	}

	zctx.From(ctx).Info("Shell",
		zap.Int("size", buf.Len()),
		zap.ByteString("data", buf.Bytes()),
	)

	return nil
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

	if err := s.makeExternalRequest(ctx); err != nil {
		return nil, errors.Wrap(err, "external request")
	}
	if err := s.makeCurlRequest(ctx); err != nil {
		return nil, errors.Wrap(err, "curl request")
	}
	if err := s.makeShellCommand(ctx); err != nil {
		return nil, errors.Wrap(err, "shell command")
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
