// Package database owns PostgreSQL pools, safe query tracing, and explicit migrations.
package database

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/lifecycle"
	"github.com/jackc/pgx/v5"
	tern "github.com/jackc/tern/v2/migrate"
	"github.com/rs/zerolog"
)

const (
	migrationCloseTimeout = 5 * time.Second
)

//go:embed migrations/*.sql
var migrations embed.FS

// MigrationResult reports observed schema versions without provider diagnostics.
type MigrationResult struct {
	StartVersion int32
	EndVersion   int32
}

// MigrationError exposes a stable operation while preserving its private cause.
type MigrationError struct {
	cause     error
	Operation string
}

func (e *MigrationError) Error() string { return "database migration " + e.Operation + " failed" }
func (e *MigrationError) Unwrap() error { return e.cause }

// Migrate preserves the existing error-only API for callers awaiting role extraction.
func Migrate(ctx context.Context, logger *zerolog.Logger, cfg *config.Config) error {
	result, err := MigrateWithResult(ctx, cfg)
	if err == nil && logger != nil {
		logger.Info().
			Int32("start_version",
				result.StartVersion).
			Int32("end_version",
				result.EndVersion).
			Msg("database migration completed")
	}
	return err
}

// MigrateWithResult applies all embedded migrations and closes its one-shot connection.
//
//nolint:nonamedreturns // Deferred connection cleanup joins errors into the returned migration result.
func MigrateWithResult(ctx context.Context,
	cfg *config.Config,
	closeContexts ...func() context.Context) (result MigrationResult,
	err error) {
	if cfg == nil {
		return result, &MigrationError{Operation: "config", cause: errors.New("configuration required")}
	}
	conn, err := pgx.Connect(ctx, databaseURL(cfg.Database))
	if err != nil {
		return result, &MigrationError{Operation: "connect", cause: err}
	}
	var cleanup lifecycle.Cleanup
	closeConnection := conn.Close
	if len(closeContexts) > 0 && closeContexts[0] != nil {
		closeConnection = func(context.Context) error { return conn.Close(closeContexts[0]()) }
	}
	if err64 := cleanup.Push("migration_connection", closeConnection); err64 != nil {
		closeCtx, cancel := context.WithTimeout(context.Background(), migrationCloseTimeout)
		defer cancel()
		return result, &MigrationError{Operation: "ownership", cause: errors.Join(err64, conn.Close(closeCtx))}
	}
	defer func() {
		if len(closeContexts) > 0 && closeContexts[0] != nil {
			if cause := cleanup.Close(closeContexts[0]()); cause != nil {
				err = errors.Join(err, &MigrationError{Operation: "close", cause: cause})
			}
		} else {
			err = errors.Join(err, closeMigration(&cleanup))
		}
	}()
	m, err := tern.NewMigrator(ctx, conn, "schema_version")
	if err != nil {
		return result, &MigrationError{Operation: "construct", cause: err}
	}
	return applyMigrations(ctx, m)
}

// databaseURL encodes userinfo, the database path, and query values consistently
// for long-lived API pools and the one-shot migration connection.
func databaseURL(db config.DatabaseConfig) string {
	dsn := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(db.User, db.Password),
		Host:     net.JoinHostPort(db.Host, strconv.Itoa(db.Port)),
		Path:     "/" + db.Name,
		RawPath:  "/" + url.PathEscape(db.Name),
		RawQuery: url.Values{"sslmode": []string{db.SSLMode}}.Encode(),
	}
	return dsn.String()
}

func closeMigration(cleanup *lifecycle.Cleanup) error {
	// A canceled migration still needs an independent bounded close attempt.
	closeCtx, cancel := context.WithTimeout(context.Background(), migrationCloseTimeout)
	defer cancel()
	if err := cleanup.Close(closeCtx); err != nil {
		return &MigrationError{Operation: "close", cause: err}
	}
	return nil
}

type migrationRunner interface {
	LoadMigrations(fs.FS) error
	GetCurrentVersion(context.Context) (int32, error)
	Migrate(context.Context) error
}

// applyMigrations isolates fallible migration stages for failure injection.
func applyMigrations(ctx context.Context, m migrationRunner) (MigrationResult, error) {
	var result MigrationResult
	subtree, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return result, &MigrationError{Operation: "subtree", cause: err}
	}
	if err107 := m.LoadMigrations(subtree); err107 != nil {
		return result, &MigrationError{Operation: "load", cause: err107}
	}
	result.StartVersion, err = m.GetCurrentVersion(ctx)
	if err != nil {
		return result, &MigrationError{Operation: "start_version", cause: err}
	}
	if err114 := m.Migrate(ctx); err114 != nil {
		return result, &MigrationError{Operation: "apply", cause: err114}
	}
	result.EndVersion, err = m.GetCurrentVersion(ctx)
	if err != nil {
		return result, &MigrationError{Operation: "end_version", cause: err}
	}
	return result, nil
}
