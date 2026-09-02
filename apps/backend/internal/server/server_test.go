package server_test

import (
	"testing"

	"flux/apps/backend/internal/config"
	"flux/apps/backend/internal/server"

	"github.com/stretchr/testify/assert"
)

func TestServer_Initialization(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: "0"},
		Database: config.DatabaseConfig{Host: "localhost", Port: "5432", User: "pg", Password: "pwd", Name: "db"},
		Redis: config.RedisConfig{Address: "localhost:6379", AnalyticsStream: "test:stream"},
		ClickHouse: config.ClickHouseConfig{Host: "localhost", Port: "9000"},
		Clerk: config.ClerkConfig{SecretKey: "test_key"},
	}

	srv, err := server.NewServer(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, srv)
}
