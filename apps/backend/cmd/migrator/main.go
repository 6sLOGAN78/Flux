// Package main runs the migrator process entry point.
package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/observability"
	"go.opentelemetry.io/otel/codes"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Stderr)
	cancel()
	os.Exit(code)
}

func run(ctx context.Context, output io.Writer) int {
	log := logger.NewLogger(config.DefaultObservabilityConfig(), output, nil)
	cfg, err := config.LoadConfigForRole(config.RoleMigrator)
	if err != nil {
		log.Error().Str("error", observability.SafeError(err)).Msg("configuration.validate")
		return 1
	}
	return runConfigured(ctx, cfg, output, migratorFactories{
		telemetry: func(ctx context.Context,
			settings observability.Settings,
			role string) (*observability.Telemetry,
			func(context.Context) error,
			error) {
			owner, err34 := observability.New(ctx, settings, role)
			if err34 != nil {
				return nil, nil, err34
			}
			return owner, owner.Shutdown, nil
		},
		migrate: func(ctx context.Context,
			cfg *config.Config,
			closeContext func() context.Context) (database.MigrationResult,
			error) {
			return database.MigrateWithResult(ctx, cfg, closeContext)
		},
	})
}

type migratorFactories struct {
	telemetry func(context.Context,
		observability.Settings,
		string) (*observability.Telemetry,
		func(context.Context) error,
		error)
	migrate func(context.Context, *config.Config, func() context.Context) (database.MigrationResult, error)
}

// PostgreSQL close and provider flush share one lazily started exit deadline.
//
//nolint:nonamedreturns // Provider-close failure must update the exit status after migration returns.
func runConfigured(ctx context.Context, cfg *config.Config, output io.Writer, f migratorFactories) (code int) {
	log := logger.NewLogger(cfg.Observability, output, nil)
	var exitOnce sync.Once
	var exitCtx context.Context
	var cancelExit context.CancelFunc
	closeContext := func() context.Context {
		exitOnce.Do(func() {
			//nolint:gosec,fatcontext // The outer defer cancels this shared exit budget after all resource closes.
			exitCtx,
				cancelExit = context.WithTimeout(context.Background(),
				cfg.Observability.HealthChecks.Timeout)
		})
		return exitCtx
	}
	defer func() {
		if cancelExit != nil {
			cancelExit()
		}
	}()
	owner, closeOwner, err := f.telemetry(ctx, cfg.Observability.TelemetrySettings(), string(config.RoleMigrator))
	if closeOwner != nil {
		defer func() {
			if err71 := closeOwner(closeContext()); err71 != nil {
				log.Error().Str("error", observability.SafeError(err71)).Msg("telemetry.shutdown")
				code = 1
			}
		}()
	}
	if err != nil || owner == nil {
		log.Error().Str("error", observability.SafeError(err)).Msg("telemetry.export")
		return 1
	}
	log = logger.NewLogger(cfg.Observability, output, owner.Logger)
	ctx, cancel := context.WithTimeout(ctx, cfg.Observability.HealthChecks.Timeout)
	defer cancel()
	ctx, span := owner.Tracer.Start(ctx, "database.query")
	defer span.End()
	result, err := f.migrate(ctx, cfg, closeContext)
	log = logger.WithContext(log, ctx)
	if err != nil {
		span.SetStatus(codes.Error, "")
		log.Error().Str("error", observability.SafeError(err)).Msg("database.query")
		return 1
	}
	log.Info().
		Str("outcome",
			"success").
		Int32("start_version",
			result.StartVersion).
		Int32("end_version",
			result.EndVersion).
		Msg("database.query")
	return 0
}
