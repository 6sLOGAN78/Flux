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
	PrimaryEnv         string `koanf:"primary_env"`
	ServerPort         string `koanf:"server_port"`
	DatabaseURL        string `koanf:"database_url"`
	RedisURL           string `koanf:"redis_url"`
	AnalyticsRedisStream string `koanf:"analytics_redis_stream"`
	ClickHouseURL      string `koanf:"clickhouse_url"`
	JWTSecret          string `koanf:"jwt_secret"`
	ClerkSecretKey     string `koanf:"clerk_secret_key"`
	ResendAPIKey       string `koanf:"resend_api_key"`
	NewRelicLicenseKey string `koanf:"new_relic_license_key"`
	AWSRegion          string `koanf:"aws_region"`
	AWSAccessKeyID     string `koanf:"aws_access_key_id"`
	AWSSecretAccessKey string `koanf:"aws_secret_access_key"`
	AWSUploadBucket    string `koanf:"aws_upload_bucket"`
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

	primaryEnv := k.String("flux_primary.env")
	if primaryEnv == "" {
		primaryEnv = os.Getenv("FLUX_PRIMARY_ENV")
	}
	if primaryEnv == "" {
		primaryEnv = "local"
	}

	// 1. Server Port
	port := k.String("flux_server.port")
	if port == "" {
		port = k.String("server_port")
	}
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	// 2. Database URL
	dbURL := k.String("database_url")
	if dbURL == "" {
		dbHost := k.String("flux_database.host")
		dbPort := k.String("flux_database.port")
		dbUser := k.String("flux_database.user")
		dbPass := k.String("flux_database.password")
		dbName := k.String("flux_database.name")
		dbSSL := k.String("flux_database.ssl_mode")

		if dbHost != "" && dbUser != "" && dbName != "" {
			if dbPort == "" {
				dbPort = "5432"
			}
			if dbSSL == "" {
				dbSSL = "disable"
			}
			dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPass, dbHost, dbPort, dbName, dbSSL)
		}
	}
	if dbURL == "" {
		dbURL = "postgres://postgres:postgrespassword@localhost:5432/flux?sslmode=disable"
	}

	// 3. Redis Address / URL
	redisURL := k.String("flux_redis.address")
	if redisURL == "" {
		redisURL = k.String("redis_url")
	}
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	analyticsRedisStream := k.String("analytics_redis_stream")
	if analyticsRedisStream == "" {
		analyticsRedisStream = "analytics:events"
	}
	
	clickHouseURL := k.String("clickhouse_url")
	if clickHouseURL == "" {
		clickHouseURL = "localhost:9000"
	}

	// 4. Auth & Clerk Secret Key
	clerkSecretKey := k.String("flux_auth.secret_key")
	if clerkSecretKey == "" {
		clerkSecretKey = k.String("clerk_secret_key")
	}
	if clerkSecretKey == "" {
		clerkSecretKey = os.Getenv("CLERK_SECRET_KEY")
	}

	jwtSecret := k.String("flux_auth.secret_key")
	if jwtSecret == "" {
		jwtSecret = k.String("jwt_secret")
	}
	if jwtSecret == "" {
		jwtSecret = "super-secret-default-flux-key-change-in-prod"
	}

	// 5. Integration Keys
	resendKey := k.String("flux_integration.resend_api_key")
	if resendKey == "" {
		resendKey = os.Getenv("RESEND_API_KEY")
	}

	// 6. New Relic APM License Key
	nrLicenseKey := k.String("flux_observability.new_relic.license_key")
	if nrLicenseKey == "" {
		nrLicenseKey = k.String("new_relic_license_key")
	}

	// 7. AWS S3 Settings
	awsRegion := k.String("flux_aws.region")
	awsAccessKey := k.String("flux_aws.access_key_id")
	awsSecretKey := k.String("flux_aws.secret_access_key")
	awsBucket := k.String("flux_aws.upload_bucket")

	cfg := &Config{
		PrimaryEnv:         primaryEnv,
		ServerPort:         port,
		DatabaseURL:        dbURL,
		RedisURL:           redisURL,
		AnalyticsRedisStream: analyticsRedisStream,
		ClickHouseURL:      clickHouseURL,
		JWTSecret:          jwtSecret,
		ClerkSecretKey:     clerkSecretKey,
		ResendAPIKey:       resendKey,
		NewRelicLicenseKey: nrLicenseKey,
		AWSRegion:          awsRegion,
		AWSAccessKeyID:     awsAccessKey,
		AWSSecretAccessKey: awsSecretKey,
		AWSUploadBucket:    awsBucket,
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

