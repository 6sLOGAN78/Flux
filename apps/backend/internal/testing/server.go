package testing

import (
	"flux/apps/backend/internal/config"
	"flux/apps/backend/internal/server"
)

// SetupTestServer initializes a server instance configured for unit and integration testing.
func SetupTestServer() (*server.Server, error) {
	cfg := &config.Config{
		ServerPort:  "0", // Auto-assign free port
		DatabaseURL: "postgres://postgres:postgrespassword@localhost:5432/flux?sslmode=disable",
		RedisURL:    "localhost:6379",
		JWTSecret:   "test-secret-key",
	}

	return server.NewServer(cfg)
}
