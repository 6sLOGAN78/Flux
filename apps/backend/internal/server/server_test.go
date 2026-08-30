package server_test

import (
	"context"
	"testing"
	"time"

	"flux/apps/backend/internal/server"
	"flux/apps/backend/internal/config"

	"github.com/stretchr/testify/require"
)

func TestServer_GracefulShutdown(t *testing.T) {
	// Initialize a mock configuration
	cfg := &config.Config{
		ServerPort:           "0", // random port
		DatabaseURL:          "postgres://postgres:postgrespassword@localhost:5432/flux?sslmode=disable",
		RedisURL:             "localhost:6379",
		AnalyticsRedisStream: "test:events",
		ClickHouseURL:        "localhost:9000",
		ClerkSecretKey:       "sk_test_placeholder",
	}

	srv, err := server.NewServer(cfg)
	// If it fails to connect to DB because of no mock, we just test the lifecycle methods if possible.
	// Since NewServer actually dials the database, this is an integration test.
	if err != nil {
		t.Skipf("Skipping integration test: requires running databases: %v", err)
	}

	go func() {
		_ = srv.Start()
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Attempt graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = srv.Stop(ctx)
	require.NoError(t, err, "Server should shut down gracefully without error or panic")
}
