package cmd

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
			app.Run(func(ctx context.Context, lg *zap.Logger, m *app.Metrics) error {
				addr := os.Getenv("HTTP_ADDR")
				if addr == "" {
					addr = "localhost:8080"
				}
				h, err := oas.NewServer(server.Server{},
					oas.WithMeterProvider(m.MeterProvider()),
					oas.WithTracerProvider(m.TracerProvider()),
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
					otelhttp.WithMeterProvider(m.MeterProvider()),
					otelhttp.WithTracerProvider(m.TracerProvider()),
				)
				s := &http.Server{
					Addr:              addr,
					ReadHeaderTimeout: time.Second,
					WriteTimeout:      time.Second,
					ReadTimeout:       time.Second,
					Handler:           middleware.Wrap(instrumentedHandler, m.LogMiddleware()),
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
			})
		},
	}
}
