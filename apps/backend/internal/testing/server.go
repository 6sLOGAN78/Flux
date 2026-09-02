package testing

import (
	"flux/apps/backend/internal/config"
	"flux/apps/backend/internal/server"
)

// SetupTestServer initializes a server instance configured for unit and integration testing.
func SetupTestServer() (*server.Server, error) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "0",
		},
		Database: config.DatabaseConfig{
			Host: "localhost",
			Port: "5432",
			User: "postgres",
			Password: "postgrespassword",
			Name: "flux",
			SSLMode: "disable",
		},
		Redis: config.RedisConfig{
			Address: "localhost:6379",
		},
		Auth: config.InternalAuthConfig{
			JWTSecret: "test-secret-key",
		},
	}

	return server.NewServer(cfg)
}
