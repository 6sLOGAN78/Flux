package main

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/6sLOGAN78/flux/internal/app"
	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/observability"
)

// The legacy flux command is an API compatibility shim. Use cmd/worker for
// consumers and cmd/migrator for explicit migrations.
func main() {
	ctx, stop := app.NotifyContext(context.Background())
	code := run(ctx, os.Stderr)
	stop()
	os.Exit(code)
}

// run returns the process exit decision after releasing the role's resources.
func run(ctx context.Context, output io.Writer) int {
	log := logger.NewLogger(config.DefaultObservabilityConfig(), output, nil)
	cfg, err := config.LoadConfigForRole(config.RoleAPI)
	if err != nil {
		log.Error().Str("error", observability.SafeError(err)).Msg("configuration.validate")
		return 1
	}
	role, err := app.NewAPI(ctx, cfg)
	if err != nil {
		log.Error().Str("error", observability.SafeError(err)).Msg("database.connect")
		return 1
	}
	if err := role.Run(ctx); err != nil {
		event := log.Error().Str("error", observability.SafeError(err))
		if errors.Is(err, context.DeadlineExceeded) {
			event.Str("error.stage", "deadline")
		}
		event.Msg("telemetry.shutdown")
		return 1
	}
	return 0
}
