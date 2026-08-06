// Package config handles environment configuration loader for Flux backend using koanf/v2.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

// Config holds the application configuration parameters.
type Config struct {
	ServerPort          string `koanf:"server_port"`
	DatabaseURL         string `koanf:"database_url"`
	RedisURL            string `koanf:"redis_url"`
	JWTSecret           string `koanf:"jwt_secret"`
	ClerkSecretKey      string `koanf:"clerk_secret_key"`
	NewRelicLicenseKey  string `koanf:"new_relic_license_key"`
}

// LoadConfig initializes application configuration from environment variables with defaults using koanf/v2.
func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	k := koanf.New(".")

	// Load environment variables with uppercase key transformation
	err := k.Load(env.Provider("", ".", func(s string) string {
		return strings.ToLower(s)
	}), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	port := k.String("server_port")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	dbURL := k.String("database_url")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgrespassword@localhost:5432/flux?sslmode=disable"
	}

	redisURL := k.String("redis_url")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	jwtSecret := k.String("jwt_secret")
	if jwtSecret == "" {
		jwtSecret = "super-secret-default-flux-key-change-in-prod"
	}

	clerkSecretKey := k.String("clerk_secret_key")
	nrLicenseKey := k.String("new_relic_license_key")

	cfg := &Config{
		ServerPort:         port,
		DatabaseURL:        dbURL,
		RedisURL:           redisURL,
		JWTSecret:          jwtSecret,
		ClerkSecretKey:     clerkSecretKey,
		NewRelicLicenseKey: nrLicenseKey,
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks that all required configuration fields are non-empty and valid.
func (c *Config) Validate() error {
	if c.ServerPort == "" {
		return errors.New("server port cannot be empty")
	}
	if c.DatabaseURL == "" {
		return errors.New("database URL cannot be empty")
	}
	if c.RedisURL == "" {
		return errors.New("redis URL cannot be empty")
	}
	return nil
}
