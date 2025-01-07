package cmd

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-faster/errors"
	sdka "github.com/go-faster/sdk/app"
	"github.com/go-faster/sdk/zctx"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/go-faster/simon/internal/app"
	"github.com/go-faster/simon/internal/middleware"
	"github.com/go-faster/simon/internal/oas"
	"github.com/go-faster/simon/internal/server"
)

type zapCorsLogger struct {
	lg *zap.SugaredLogger
}

func (z zapCorsLogger) Printf(s string, i ...interface{}) {
	z.lg.Infof(strings.TrimSpace(s), i...)
}

func getEnvBool(k string) bool {
	v, _ := strconv.ParseBool(os.Getenv(k))
	return v
}

func cmdServer() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Run a HTTP server",
		Run: func(cmd *cobra.Command, args []string) {
			sdka.Run(func(ctx context.Context, lg *zap.Logger, t *sdka.Telemetry) error {
				ctx = zctx.WithOpenTelemetryZap(ctx)
				addr := os.Getenv("HTTP_ADDR")
				if addr == "" {
					addr = "localhost:8080"
				}
				lg.Info("Listening on", zap.String("addr", addr))
				h, err := oas.NewServer(server.Server{},
					oas.WithMeterProvider(t.MeterProvider()),
					oas.WithTracerProvider(t.TracerProvider()),
				)
				if err != nil {
					return err
				}

				allowedOrigins := []string{"*"}
				if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
					allowedOrigins = strings.Split(v, ",")
				}

				c := cors.New(cors.Options{
					AllowedOrigins:   allowedOrigins,
					AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS"),
					Debug:            getEnvBool("CORS_DEBUG"),
					MaxAge:           60, // seconds
				})

				c.Log = zapCorsLogger{lg: lg.Sugar()}

				spanNameFormatter := app.NewSpanNameFormatter(h)
				instrumentedHandler := otelhttp.NewHandler(c.Handler(h), "",
					otelhttp.WithSpanNameFormatter(spanNameFormatter),
					otelhttp.WithMeterProvider(t.MeterProvider()),
					otelhttp.WithTracerProvider(t.TracerProvider()),
				)
				s := &http.Server{
					Addr:              addr,
					ReadHeaderTimeout: time.Second,
					WriteTimeout:      time.Second,
					ReadTimeout:       time.Second,
					Handler:           middleware.Wrap(instrumentedHandler, app.LogMiddleware(lg)),
				}

				lg.Info("Starting HTTP server", zap.String("addr", addr))

				parentCtx := ctx
				g, ctx := errgroup.WithContext(ctx)
				g.Go(func() error {
					<-ctx.Done()
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()
					return s.Shutdown(ctx)
				})
				g.Go(func() error {
					if err := s.ListenAndServe(); err != nil {
						if errors.Is(err, http.ErrServerClosed) && parentCtx.Err() != nil {
							lg.Info("HTTP server closed gracefully")
							return nil
						}
						return errors.Wrap(err, "http server")
					}
					return nil
				})
				return g.Wait()
			},
				sdka.WithResource(func(ctx context.Context) (*resource.Resource, error) {
					return resource.New(ctx,
						resource.WithOS(),
						resource.WithFromEnv(),
						resource.WithTelemetrySDK(),
						resource.WithHost(),
						resource.WithProcess(),
						resource.WithAttributes(
							attribute.String("service.name", "simon.server"),
						),
					)
				}),
			)
		},
	}
}
