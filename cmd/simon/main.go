package main

import (
	"context"
	"os"
	"time"

	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.uber.org/zap"

	"github.com/go-faster/simon/internal/app"
	"github.com/go-faster/simon/internal/otelenv"
)

func main() {
	if os.Getenv("OTEL_RESOURCE_ATTRIBUTES") == "" && os.Getenv("OTEL_SERVICE_NAME") == "" {
		// Set default service name and namespace.
		otelenv.Set(
			semconv.ServiceName("simon"),
			semconv.ServiceNamespace("go-faster"),
		)
	}
	app.Run(func(ctx context.Context, lg *zap.Logger, m *app.Metrics) error {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				lg.Info("Hello, world!")
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})
}
