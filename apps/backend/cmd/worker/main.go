package main

import (
	"context"
	"io"
	"os"

	"github.com/6sLOGAN78/flux/internal/app"
	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/rs/zerolog"
)

func main() {
	ctx, stop := app.NotifyContext(context.Background())
	code := run(ctx, os.Stderr)
	stop()
	os.Exit(code)
}

// run returns the process exit decision after releasing the role's resources.
func run(ctx context.Context, output io.Writer) int {
	log := zerolog.New(output).With().Str("role", string(config.RoleWorker)).Timestamp().Logger()
	cfg, err := config.LoadConfigForRole(config.RoleWorker)
	if err != nil {
		log.Error().Err(err).Msg("worker configuration failed")
		return 1
	}
	role, err := app.NewWorker(ctx, cfg)
	if err != nil {
		log.Error().Err(err).Msg("worker construction failed")
		return 1
	}
	if err := role.Run(ctx); err != nil {
		log.Error().Err(err).Msg("worker execution failed")
		return 1
	}
	return 0
}
