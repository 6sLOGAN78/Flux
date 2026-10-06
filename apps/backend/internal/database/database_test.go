package database

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/netip"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	loggerpkg "github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/observability"
	containerconfig "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

type failingPool struct {
	err    error
	closed int
}

func TestDatabaseTelemetryParameterizedPostgres(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: testcontainers.ContainerRequest{
		Image: "postgres:17.11-alpine@sha256:b0f9560a2de083e2cc7382e75f808c7381a32852a7ec49117deedb300e552b24", ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{"POSTGRES_USER": "testuser", "POSTGRES_PASSWORD": "PRIVATE_DATABASE_PASSWORD", "POSTGRES_DB": "telemetry_test"},
		HostConfigModifier: func(cfg *containerconfig.HostConfig) {
			cfg.PortBindings = network.PortMap{network.MustParsePort("5432/tcp"): {{HostIP: netip.MustParseAddr("127.0.0.1"), HostPort: "0"}}}
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}, Started: true})
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	portNumber, err := strconv.Atoi(port.Port())
	require.NoError(t, err)
	exporter := tracetest.NewInMemoryExporter()
	tel, err := observability.New(ctx, observability.Settings{Enabled: true, SampleRatio: 1, ExportInterval: time.Millisecond, Exporters: observability.Exporters{Trace: exporter}}, "api")
	require.NoError(t, err)
	var logs bytes.Buffer
	log := loggerpkg.NewLogger(config.DefaultObservabilityConfig(), &logs, nil)
	cfg := &config.Config{Primary: config.Primary{Env: "local"}, Database: config.DatabaseConfig{Host: host, Port: portNumber, User: "testuser", Password: "PRIVATE_DATABASE_PASSWORD", Name: "telemetry_test", SSLMode: "disable"}}
	db, err := New(cfg, &log, nil, tel)
	require.NoError(t, err)
	defer db.Close()
	ctx = observability.WithCorrelation(ctx, "ac1c5930-0ce6-4851-97c9-565d8192a87b", "bef0ae77-1fb7-4f52-9777-34cc01939ea8")
	var got string
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT $1::text /* PRIVATE_SQL_TEXT */", "PRIVATE_SQL_ARGUMENT").Scan(&got))
	require.Equal(t, "PRIVATE_SQL_ARGUMENT", got)
	_, err = db.Pool.Exec(ctx, "SELECT $1::integer /* PRIVATE_ERROR_SQL */", "PRIVATE_DRIVER_ERROR")
	require.Error(t, err)
	require.Contains(t, err.Error(), "PRIVATE_DRIVER_ERROR")
	require.NoError(t, tel.Shutdown(ctx))
	captured := exporter.GetSpans()
	require.GreaterOrEqual(t, len(captured), 3)
	var outcomes []string
	for _, span := range captured {
		require.Equal(t, "database.query", span.Name)
		require.Equal(t, trace.SpanKindClient, span.SpanKind)
		for _, attr := range span.Attributes {
			require.Contains(t, []string{"dependency", "operation", "outcome", "error.category", "db.system.name", "request_id", "correlation_id"}, string(attr.Key))
			if string(attr.Key) == "outcome" {
				outcomes = append(outcomes, attr.Value.AsString())
			}
		}
		data, _ := json.Marshal(span)
		for _, marker := range []string{"PRIVATE_SQL_TEXT", "PRIVATE_SQL_ARGUMENT", "PRIVATE_DRIVER_ERROR", "PRIVATE_ERROR_SQL", "PRIVATE_DATABASE_PASSWORD"} {
			require.NotContains(t, string(data), marker)
		}
	}
	require.Contains(t, outcomes, "success")
	require.Contains(t, outcomes, "error")
	for _, marker := range []string{"PRIVATE_SQL_TEXT", "PRIVATE_SQL_ARGUMENT", "PRIVATE_DRIVER_ERROR", "PRIVATE_ERROR_SQL", "PRIVATE_DATABASE_PASSWORD"} {
		require.False(t, strings.Contains(logs.String(), marker))
	}
}

func (p *failingPool) Ping(context.Context) error { return p.err }
func (p *failingPool) Close()                     { p.closed++ }

func TestDatabasePingFailureClosesPool(t *testing.T) {
	cause := errors.New("ping failed")
	pool := &failingPool{err: cause}
	err := pingPool(context.Background(), pool)
	if !errors.Is(err, cause) {
		t.Fatalf("ping cause lost: %v", err)
	}
	if pool.closed != 1 {
		t.Fatalf("failed pool closed %d times", pool.closed)
	}
}

func TestDatabaseSuccessfulPingKeepsPool(t *testing.T) {
	pool := &failingPool{}
	if err := pingPool(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
	if pool.closed != 0 {
		t.Fatal("successful construction closed its pool")
	}
}
