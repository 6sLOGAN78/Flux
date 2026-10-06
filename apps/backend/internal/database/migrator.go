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

	"github.com/6sLOGAN78/flux/internal/app"
	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/jackc/pgx/v5"
	tern "github.com/jackc/tern/v2/migrate"
	"github.com/rs/zerolog"
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
	Operation string
	cause     error
}

func (e *MigrationError) Error() string { return "database migration " + e.Operation + " failed" }
func (e *MigrationError) Unwrap() error { return e.cause }

// Migrate preserves the existing error-only API for callers awaiting role extraction.
func Migrate(ctx context.Context, logger *zerolog.Logger, cfg *config.Config) error {
	result, err := MigrateWithResult(ctx, cfg)
	if err == nil && logger != nil {
		logger.Info().Int32("start_version", result.StartVersion).Int32("end_version", result.EndVersion).Msg("database migration completed")
	}
	return err
}

// MigrateWithResult applies all embedded migrations and closes its one-shot connection.
func MigrateWithResult(ctx context.Context, cfg *config.Config) (result MigrationResult, err error) {
	if cfg == nil {
		return result, &MigrationError{Operation: "config", cause: errors.New("configuration required")}
	}
	db := cfg.Database
	dsn := &url.URL{Scheme: "postgres", User: url.UserPassword(db.User, db.Password), Host: net.JoinHostPort(db.Host, strconv.Itoa(db.Port)), Path: "/" + db.Name}
	dsn.RawQuery = url.Values{"sslmode": []string{db.SSLMode}}.Encode()
	conn, err := pgx.Connect(ctx, dsn.String())
	if err != nil {
		return result, &MigrationError{Operation: "connect", cause: err}
	}
	var cleanup app.Cleanup
	if err := cleanup.Push("migration_connection", conn.Close); err != nil {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return result, &MigrationError{Operation: "ownership", cause: errors.Join(err, conn.Close(closeCtx))}
	}
	defer func() {
		err = errors.Join(err, closeMigration(&cleanup))
	}()
	m, err := tern.NewMigrator(ctx, conn, "schema_version")
	if err != nil {
		return result, &MigrationError{Operation: "construct", cause: err}
	}
	return applyMigrations(ctx, m)
}

func closeMigration(cleanup *app.Cleanup) error {
	// A canceled migration still needs an independent bounded close attempt.
	closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
func applyMigrations(ctx context.Context, m migrationRunner) (result MigrationResult, err error) {
	subtree, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return result, &MigrationError{Operation: "subtree", cause: err}
	}
	if err := m.LoadMigrations(subtree); err != nil {
		return result, &MigrationError{Operation: "load", cause: err}
	}
	result.StartVersion, err = m.GetCurrentVersion(ctx)
	if err != nil {
		return result, &MigrationError{Operation: "start_version", cause: err}
	}
	if err := m.Migrate(ctx); err != nil {
		return result, &MigrationError{Operation: "apply", cause: err}
	}
	result.EndVersion, err = m.GetCurrentVersion(ctx)
	if err != nil {
		return result, &MigrationError{Operation: "end_version", cause: err}
	}
	return result, nil
}
