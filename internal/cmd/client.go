package cmd

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	"github.com/go-faster/simon/internal/app"
	"github.com/go-faster/simon/internal/oas"
)

func cmdClient() *cobra.Command {
	return &cobra.Command{
		Use:   "client",
		Short: "Run a HTTP client",
		Run: func(cmd *cobra.Command, args []string) {
			app.Run(func(ctx context.Context, logger *zap.Logger, m *app.Metrics) error {
				addr := os.Getenv("SERVER_ADDR")
				if addr == "" {
					addr = "http://localhost:8080"
				}
				spanNameFormatter := app.NewSpanNameFormatter(&oas.Server{})
				c, err := oas.NewClient(addr,
					oas.WithMeterProvider(m.MeterProvider()),
					oas.WithTracerProvider(m.TracerProvider()),
					oas.WithClient(&http.Client{
						Timeout: time.Second * 2,
						Transport: otelhttp.NewTransport(http.DefaultTransport,
							otelhttp.WithSpanNameFormatter(spanNameFormatter),
							otelhttp.WithMeterProvider(m.MeterProvider()),
							otelhttp.WithTracerProvider(m.TracerProvider()),
						),
					}),
				)
				if err != nil {
					return errors.Wrap(err, "client")
				}
				ticker := time.NewTicker(time.Second)
				tracer := m.TracerProvider().Tracer("")
				tick := func() {
					ctx, cancel := context.WithTimeout(ctx, time.Second)
					defer cancel()

					ctx, span := tracer.Start(ctx, "client.tick")
					defer span.End()

					lg := zctx.From(ctx)
					lg.Info("Sending request")

					status, err := c.Status(ctx)
					if err != nil {
						lg.Error("Request failed", zap.Error(err))
						return
					}
					lg.Info("Request succeeded", zap.String("message", status.Message))
				}
				tick()
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-ticker.C:
						tick()
					}
				}
			})
		},
	}
}
