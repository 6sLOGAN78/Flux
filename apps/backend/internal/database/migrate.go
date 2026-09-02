package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"flux/apps/backend/internal/config"

	"github.com/jackc/pgx/v5"
	tern "github.com/jackc/tern/v2/migrate"
	"github.com/rs/zerolog"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate executes pending database migrations embedded in the binary via Tern using Config.
func Migrate(ctx context.Context, logger *zerolog.Logger, cfg *config.Config) error {
	if cfg == nil || cfg.GetDatabaseURL() == "" {
		return fmt.Errorf("invalid config or empty DatabaseURL")
	}
	return MigrateDSN(ctx, logger, cfg.GetDatabaseURL())
}

// MigrateDSN executes pending database migrations using a DSN string.
func MigrateDSN(ctx context.Context, logger *zerolog.Logger, dsn string) error {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database for migration: %w", err)
	}
	defer conn.Close(ctx)

	m, err := tern.NewMigrator(ctx, conn, "schema_version")
	if err != nil {
		return fmt.Errorf("constructing database migrator: %w", err)
	}

	subtree, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("retrieving database migrations subtree: %w", err)
	}

	if err := m.LoadMigrations(subtree); err != nil {
		return fmt.Errorf("loading database migrations: %w", err)
	}

	from, err := m.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("retrieving current database migration version: %w", err)
	}

	if err := m.Migrate(ctx); err != nil {
		return fmt.Errorf("executing database migrations: %w", err)
	}

	if from == int32(len(m.Migrations)) {
		if logger != nil {
			logger.Info().Msgf("database schema up to date, version %d", len(m.Migrations))
		}
	} else {
		if logger != nil {
			logger.Info().Msgf("migrated database schema, from %d to %d", from, len(m.Migrations))
		}
	}

	return nil
}
