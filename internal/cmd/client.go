package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-faster/errors"
	sdka "github.com/go-faster/sdk/app"
	"github.com/go-faster/sdk/zctx"
	ohttp "github.com/ogen-go/ogen/http"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"

	"github.com/go-faster/simon/internal/app"
	"github.com/go-faster/simon/internal/oas"
)

func cmdClient() *cobra.Command {
	var arg struct {
		UploadRPS            int
		UploadHashIterations int
	}
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Run a HTTP client",
		Run: func(cmd *cobra.Command, args []string) {
			sdka.Run(func(ctx context.Context, logger *zap.Logger, t *sdka.Telemetry) error {
				ctx = zctx.WithOpenTelemetryZap(ctx)
				addr := os.Getenv("SERVER_ADDR")
				if addr == "" {
					addr = "http://localhost:8080"
				}
				spanNameFormatter := app.NewSpanNameFormatter(&oas.Server{})
				c, err := oas.NewClient(addr,
					oas.WithMeterProvider(t.MeterProvider()),
					oas.WithTracerProvider(t.TracerProvider()),
					oas.WithClient(&http.Client{
						Timeout: time.Second * 2,
						Transport: otelhttp.NewTransport(http.DefaultTransport,
							otelhttp.WithSpanNameFormatter(spanNameFormatter),
							otelhttp.WithMeterProvider(t.MeterProvider()),
							otelhttp.WithTracerProvider(t.TracerProvider()),
						),
					}),
				)
				if err != nil {
					return errors.Wrap(err, "client")
				}
				g, ctx := errgroup.WithContext(ctx)
				g.Go(func() error {
					ticker := time.NewTicker(time.Second)
					tracer := t.TracerProvider().Tracer("")
					tick := func() {
						ctx, cancel := context.WithTimeout(ctx, time.Millisecond*250)
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
						case <-t.ShutdownContext().Done():
							return ctx.Err()
						case <-ticker.C:
							tick()
						}
					}
				})
				g.Go(func() error {
					// Uploads.
					const burst = 1
					limiter := rate.NewLimiter(rate.Limit(arg.UploadRPS), burst)
					rnd := rand.New(rand.NewSource(10)) // #nosec G404

					tracer := t.TracerProvider().Tracer("")

					tick := func() error {
						if err := limiter.Wait(ctx); err != nil {
							return errors.Wrap(err, "limiter")
						}

						ctx, cancel := context.WithTimeout(ctx, time.Second*5)
						defer cancel()

						ctx, span := tracer.Start(ctx, "client.upload",
							trace.WithAttributes(
								attribute.Int("rps", arg.UploadRPS),
								attribute.Int("hash_iterations", arg.UploadHashIterations),
							),
						)
						defer span.End()

						lg := zctx.From(ctx)
						lg.Info("Uploading data")

						// Generate payload.
						const payloadSize = 1024 * 1024 * 1 // 1MB
						payload := make([]byte, payloadSize)
						if _, err := rnd.Read(payload); err != nil {
							return errors.Wrap(err, "gen payload")
						}

						msg, err := c.UploadFile(ctx, &oas.UploadFileReq{
							File: ohttp.MultipartFile{
								Name: "random.bin",
								Size: int64(len(payload)),
								File: bytes.NewReader(payload),
							},
							Iterations: oas.NewOptInt(arg.UploadHashIterations),
						})
						if err != nil {
							return errors.Wrap(err, "upload file")
						}

						lg.Info("Upload succeeded", zap.String("hash", msg.Hash))
						// Verifying hash.
						h := sha256.New()
						for i := 0; i < arg.UploadHashIterations; i++ {
							if _, err := h.Write(payload); err != nil {
								return errors.Wrap(err, "write")
							}
						}
						gotHash := fmt.Sprintf("%x", h.Sum(nil))
						span.AddEvent("Hash verification",
							trace.WithAttributes(
								attribute.String("expected", gotHash),
								attribute.String("got", msg.Hash),
								attribute.Bool("equal", gotHash == msg.Hash),
							),
						)

						return nil
					}

					lg := zctx.From(ctx)
					for {
						select {
						case <-t.ShutdownContext().Done():
							return ctx.Err()
						default:
							if err := tick(); err != nil {
								lg.Error("Upload failed", zap.Error(err))
							} else {
								lg.Info("Upload succeeded")
							}
						}
					}
				})
				return g.Wait()
			},
				sdka.WithServiceName("simon.client"),
			)
		},
	}

	cmd.Flags().IntVar(&arg.UploadRPS, "upload-rps", 1, "Upload requests per second")
	cmd.Flags().IntVar(&arg.UploadHashIterations, "upload-hash-iterations", 10, "Upload hash iterations")

	return cmd
}
