//nolint:testpackage // These tests verify package-private lifecycle and failure-injection seams.
package database

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/lifecycle"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	tern "github.com/jackc/tern/v2/migrate"
	containerconfig "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/rs/zerolog"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type failingMigration struct {
	cause error
	stage string
	reads int
}

func (m *failingMigration) LoadMigrations(fs.FS) error {
	if m.stage == "load" {
		return m.cause
	}
	return nil
}
func (m *failingMigration) GetCurrentVersion(context.Context) (int32, error) {
	m.reads++
	if m.stage == "start_version" || (m.stage == "end_version" && m.reads == 2) {
		return 0, m.cause
	}
	return 0, nil
}
func (m *failingMigration) Migrate(context.Context) error {
	if m.stage == "apply" {
		return m.cause
	}
	return nil
}

func TestMigrationFailuresPreserveSafeCauses(t *testing.T) {
	for _, stage := range []string{"load", "start_version", "apply", "end_version"} {
		t.Run(stage, func(t *testing.T) {
			cause := errors.New("SECRET-MARKER postgres://private:password@host")
			_, err := applyMigrations(context.Background(), &failingMigration{stage: stage, cause: cause})
			var typed *MigrationError
			if !errors.Is(err, cause) || !errors.As(err, &typed) || typed.Operation != stage {
				t.Fatalf("stage/cause not preserved: %v", err)
			}
			if strings.Contains(err.Error(),
				"SECRET-MARKER") ||
				strings.Contains(err.Error(),
					"postgres://") {
				t.Fatal("migration diagnostic leaked provider data")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_,
		err := MigrateWithResult(ctx,
		&config.Config{Database: config.DatabaseConfig{Host: "127.0.0.1",
			Port:     5432,
			User:     "SECRET-MARKER",
			Password: "SECRET-MARKER",
			Name:     "test",
			SSLMode:  "disable"}})
	if !errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "SECRET-MARKER") {
		t.Fatalf("connect cause/redaction failed: %v", err)
	}
	var cleanup lifecycle.Cleanup
	cause := errors.New("SECRET-MARKER cleanup failure")
	if err78 := cleanup.Push("migration_connection", func(closeCtx context.Context) error {
		if closeCtx.Err() != nil {
			t.Fatal("cleanup inherited canceled operation context")
		}
		return cause
	}); err78 != nil {
		t.Fatal(err78)
	}
	err = closeMigration(&cleanup)
	var typed *MigrationError
	if !errors.Is(err,
		cause) ||
		!errors.As(err,
			&typed) ||
		typed.Operation != "close" ||
		strings.Contains(err.Error(),
			"SECRET-MARKER") {
		t.Fatalf("cleanup stage/cause/redaction failed: %v", err)
	}
}

//nolint:gocognit,gocyclo,cyclop // Keep this complete integration protocol and its ordered failure assertions together.
func TestMigrationEmptyDatabaseAndBinary(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	const image = "postgres:17.11-alpine@sha256:b0f9560a2de083e2cc7382e75f808c7381a32852a7ec49117deedb300e552b24"
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: image, ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{"POSTGRES_USER": "testuser",
				"POSTGRES_PASSWORD": "SECRET-MARKER !@:/?+",
				"POSTGRES_DB":       "migration_test"},
			HostConfigModifier: func(cfg *containerconfig.HostConfig) {
				cfg.PortBindings = network.PortMap{network.MustParsePort("5432/tcp"): {{HostIP: netip.MustParseAddr("127.0.0.1"),
					HostPort: "0"}}}
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		}, Started: true,
	})
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("start pinned PostgreSQL: %v", err)
	}
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatal(err)
	}
	portNumber, err := strconv.Atoi(port.Port())
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Database: config.DatabaseConfig{Host: host,
		Port:     portNumber,
		User:     "testuser",
		Password: "SECRET-MARKER !@:/?+",
		Name:     "migration_test",
		SSLMode:  "disable"}}
	connectionConfig, err := pgx.ParseConfig("sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	connectionConfig.Host, connectionConfig.Port = host, uint16(portNumber)
	connectionConfig.User,
		connectionConfig.Password,
		connectionConfig.Database = cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name
	conn, err := pgx.ConnectConfig(ctx, connectionConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	var before *string
	if err136 := conn.QueryRow(ctx,
		"SELECT to_regclass('public.schema_version')::text").
		Scan(&before); err136 != nil ||
		before != nil {
		t.Fatalf("database is not empty: %v", err136)
	}
	files, err := fs.Glob(migrations, "migrations/*.sql")
	if err != nil {
		t.Fatal(err)
	}
	want := int32(len(files))
	for i := range 2 {
		result, err145 := MigrateWithResult(ctx, cfg)
		if err145 != nil {
			t.Fatal(err145)
		}
		start := int32(0)
		if i == 1 {
			start = want
		}
		if result.StartVersion != start || result.EndVersion != want {
			t.Fatalf("run %d: got %+v, want %d -> %d", i, result, start, want)
		}
		var version int32
		if err157 := conn.QueryRow(ctx,
			"SELECT version FROM schema_version").
			Scan(&version); err157 != nil ||
			version != want {
			t.Fatalf("exact schema version: %d, %v", version, err157)
		}
	}
	var logs bytes.Buffer
	logger := zerolog.New(&logs)
	if err163 := Migrate(ctx, &logger, cfg); err163 != nil {
		t.Fatal(err163)
	}
	if !strings.Contains(logs.String(),
		"start_version") ||
		!strings.Contains(logs.String(),
			"end_version") ||
		strings.Contains(logs.String(),
			"SECRET-MARKER") {
		t.Fatalf("unsafe/missing version logs: %s", &logs)
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "migrator")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/migrator")
	build.Dir = root
	if output, err176 := build.CombinedOutput(); err176 != nil {
		t.Fatalf("build binary: %v: %s", err176, output)
	}
	minimalEnv := []string{}
	for _, item := range os.Environ() {
		if !strings.HasPrefix(item, "FLUX_") && !strings.HasPrefix(item, "PG") {
			minimalEnv = append(minimalEnv, item)
		}
	}
	minimalEnv = append(minimalEnv,
		"FLUX_PRIMARY.ENV=test",
		"FLUX_DATABASE.HOST="+host,
		"FLUX_DATABASE.PORT="+port.Port(),
		"FLUX_DATABASE.USER=testuser",
		"FLUX_DATABASE.PASSWORD="+cfg.Database.Password,
		"FLUX_DATABASE.NAME=migration_test",
		"FLUX_DATABASE.SSL_MODE=disable")
	// Exercise the executable's first run from an empty migration ledger too.
	// Undo the actual schema through the same embedded migration corpus before
	// resetting its ledger; product migrations cannot safely be replayed over DDL.
	migrator, err := tern.NewMigrator(ctx, conn, "schema_version")
	if err != nil {
		t.Fatal(err)
	}
	subtree, err := fs.Sub(migrations, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	if loadErr := migrator.LoadMigrations(subtree); loadErr != nil {
		t.Fatal(loadErr)
	}
	if prefixErr := migrator.MigrateTo(ctx, 1); prefixErr != nil {
		t.Fatal(prefixErr)
	}
	upgraded, upgradeErr := MigrateWithResult(ctx, cfg)
	if upgradeErr != nil || upgraded.StartVersion != 1 || upgraded.EndVersion != want {
		t.Fatal("existing bootstrap prefix did not upgrade to the exact latest product schema")
	}
	var usersTable *string
	if queryErr := conn.QueryRow(ctx,
		"SELECT to_regclass('public.users')::text").Scan(&usersTable); queryErr != nil ||
		usersTable == nil {
		t.Fatal("bootstrap upgrade did not create the actual identity table")
	}
	noOp, noOpErr := MigrateWithResult(ctx, cfg)
	if noOpErr != nil || noOp.StartVersion != want || noOp.EndVersion != want {
		t.Fatal("upgraded product schema was not an exact no-op")
	}
	if downErr := migrator.MigrateTo(ctx, 0); downErr != nil {
		t.Fatal(downErr)
	}
	if _, err187 := conn.Exec(ctx, "DROP TABLE schema_version"); err187 != nil {
		t.Fatal(err187)
	}
	for i := range 2 {
		cmd := exec.CommandContext(ctx, binary)
		cmd.Dir, cmd.Env = t.TempDir(), minimalEnv
		output, err193 := cmd.CombinedOutput()
		start := int32(0)
		if i == 1 {
			start = want
		}
		if err193 != nil ||
			!bytes.Contains(output,
				[]byte(fmt.Sprintf("\"start_version\":%d",
					start))) ||
			!bytes.Contains(output,
				[]byte(fmt.Sprintf("\"end_version\":%d",
					want))) ||
			bytes.Contains(output,
				[]byte("SECRET-MARKER")) {
			t.Fatalf("one-shot binary: %v: %s", err193, output)
		}
	}
	check := exec.CommandContext(ctx, "task", "--dir", root, "migrate:check")
	check.Env, check.Stdin = minimalEnv, strings.NewReader("")
	if output, err204 := check.CombinedOutput(); err204 != nil {
		t.Fatalf("CI non-interactive task: %v: %s", err204, output)
	}
	failed := exec.CommandContext(ctx, binary)
	failed.Dir, failed.Env = t.TempDir(), append(append([]string(nil), minimalEnv...), "FLUX_DATABASE.PORT=1")
	if output,
		err209 := failed.CombinedOutput(); err209 == nil ||
		bytes.Contains(output,
			[]byte("SECRET-MARKER")) ||
		bytes.Contains(output,
			[]byte(net.JoinHostPort(host,
				"1"))) {
		t.Fatalf("failed binary status/redaction: %v: %s", err209, output)
	}
	if downErr := migrator.MigrateTo(ctx, 0); downErr != nil {
		t.Fatal(downErr)
	}
	if _,
		err212 := conn.Exec(ctx,
		"UPDATE schema_version SET version = 0; "+
			"CREATE FUNCTION reject_version_update() RETURNS trigger LANGUAGE plpgsql "+
			"AS $$ BEGIN RAISE EXCEPTION 'SECRET-MARKER'; END $$; "+
			"CREATE TRIGGER reject_version_update BEFORE UPDATE ON schema_version FOR EACH ROW "+
			"EXECUTE FUNCTION reject_version_update()"); err212 != nil {
		t.Fatal(err212)
	}
	_, err = MigrateWithResult(ctx, cfg)
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || strings.Contains(err.Error(), "SECRET-MARKER") {
		t.Fatalf("real migration failure cause/redaction: %v", err)
	}
	var failedVersion int32
	if queryErr := conn.QueryRow(ctx,
		"SELECT version FROM schema_version").Scan(&failedVersion); queryErr != nil ||
		failedVersion != 0 {
		t.Fatal("failed migration advanced its ledger")
	}
	if queryErr := conn.QueryRow(ctx,
		"SELECT to_regclass('public.users')::text").Scan(&usersTable); queryErr != nil ||
		usersTable != nil {
		t.Fatal("failed migration published a partial identity schema")
	}
	var connections int
	if err221 := conn.QueryRow(ctx,
		"SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() "+
			"AND usename = current_user AND pid <> pg_backend_pid()").
		Scan(&connections); err221 != nil ||
		connections != 0 {
		t.Fatalf("one-shot connection leaked: count %d, %v", connections, err221)
	}
}
