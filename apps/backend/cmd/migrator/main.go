package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/rs/zerolog"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Stderr)
	cancel()
	os.Exit(code)
}

func run(ctx context.Context, output io.Writer) int {
	log := zerolog.New(output).With().Str("role", string(config.RoleMigrator)).Timestamp().Logger()
	cfg, err := config.LoadConfigForRole(config.RoleMigrator)
	if err != nil {
		log.Error().Err(err).Msg("migrator configuration failed")
		return 1
	}
	level, err := zerolog.ParseLevel(cfg.Observability.Logging.Level)
	if err != nil {
		log.Error().Msg("migrator logging configuration failed")
		return 1
	}
	if cfg.Observability.Logging.Format == "console" {
		log = log.Output(zerolog.ConsoleWriter{Out: output})
	}
	log = log.Level(level)
	ctx, cancel := context.WithTimeout(ctx, cfg.Observability.HealthChecks.Timeout)
	defer cancel()
	result, err := database.MigrateWithResult(ctx, cfg)
	if err != nil {
		log.Error().Err(err).Msg("database migration failed")
		return 1
	}
	log.Info().Int32("start_version", result.StartVersion).Int32("end_version", result.EndVersion).Msg("database migration completed")
	return 0
}
