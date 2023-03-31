package cmd

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-faster/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/go-faster/simon/internal/app"
	"github.com/go-faster/simon/internal/oas"
	"github.com/go-faster/simon/internal/server"
)

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
				h, err := oas.NewServer(server.Server{})
				if err != nil {
					return err
				}
				s := &http.Server{
					Addr:              addr,
					ReadHeaderTimeout: time.Second,
					WriteTimeout:      time.Second,
					ReadTimeout:       time.Second,
					Handler:           h,
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
