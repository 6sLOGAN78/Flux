package config_test

import (
	"os"
	"testing"

	"flux/apps/backend/internal/config"
)

func TestLoadConfig_Defaults(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("REDIS_URL")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error loading default config, got: %v", err)
	}

	if cfg.ServerPort != "8080" {
		t.Errorf("expected default ServerPort '8080', got '%s'", cfg.ServerPort)
	}

	expectedDBURL := "postgres://postgres:postgrespassword@localhost:5432/flux?sslmode=disable"
	if cfg.DatabaseURL != expectedDBURL {
		t.Errorf("expected default DatabaseURL '%s', got '%s'", expectedDBURL, cfg.DatabaseURL)
	}
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://user:pass@db:5432/testdb?sslmode=require")
	t.Setenv("REDIS_URL", "redis:6379")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error loading env config, got: %v", err)
	}

	if cfg.ServerPort != "9090" {
		t.Errorf("expected ServerPort '9090', got '%s'", cfg.ServerPort)
	}

	expectedDBURL := "postgres://user:pass@db:5432/testdb?sslmode=require"
	if cfg.DatabaseURL != expectedDBURL {
		t.Errorf("expected DatabaseURL '%s', got '%s'", expectedDBURL, cfg.DatabaseURL)
	}
}

func TestConfig_Validate(t *testing.T) {
	cfg := &config.Config{
		ServerPort:  "8080",
		DatabaseURL: "postgres://postgres:postgrespassword@localhost:5432/flux?sslmode=disable",
		RedisURL:    "localhost:6379",
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}

	invalidCfg := &config.Config{
		ServerPort:  "",
		DatabaseURL: "",
		RedisURL:    "",
	}

	if err := invalidCfg.Validate(); err == nil {
		t.Error("expected error for empty config, got nil")
	}
}
