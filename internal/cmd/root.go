package cmd

import (
	"context"
	"time"

	"github.com/go-faster/sdk/app"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Root() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "simon",
		Short: "Simon is Observability Workloads Simulator",
		Run: func(cmd *cobra.Command, args []string) {
			app.Run(func(ctx context.Context, lg *zap.Logger, t *app.Telemetry) error {
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
		},
	}
	cmd.AddCommand(
		cmdServer(),
		cmdClient(),
	)
	return cmd
}
