package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

type ServerConfig struct {
	Port               string `koanf:"port" validate:"required"`
	FrontendURL        string `koanf:"frontend_url"`
	PlatformDomain     string `koanf:"platform_domain"`
	ReadTimeout        string `koanf:"read_timeout"`
	WriteTimeout       string `koanf:"write_timeout"`
	IdleTimeout        string `koanf:"idle_timeout"`
	CorsAllowedOrigins string `koanf:"cors_allowed_origins"`
}

type DatabaseConfig struct {
	Host            string `koanf:"host" validate:"required"`
	Port            string `koanf:"port" validate:"required"`
	User            string `koanf:"user" validate:"required"`
	Password        string `koanf:"password" validate:"required"`
	Name            string `koanf:"name" validate:"required"`
	SSLMode         string `koanf:"ssl_mode"`
	MaxOpenConns    string `koanf:"max_open_conns"`
	MaxIdleConns    string `koanf:"max_idle_conns"`
	ConnMaxLifetime string `koanf:"conn_max_lifetime"`
	ConnMaxIdleTime string `koanf:"conn_max_idle_time"`
}

type RedisConfig struct {
	Address         string `koanf:"address" validate:"required"`
	AnalyticsStream string `koanf:"analytics_stream"`
}

type ClickHouseConfig struct {
	Host string `koanf:"host"`
	Port string `koanf:"port"`
}

type ClerkConfig struct {
	SecretKey      string `koanf:"secret_key" validate:"required"`
	PublishableKey string `koanf:"publishable_key"`
}

type InternalAuthConfig struct {
	JWTSecret      string `koanf:"jwt_secret"`
	InternalAPIKey string `koanf:"internal_api_key"`
}

type StripeConfig struct {
	SecretKey     string `koanf:"secret_key"`
	WebhookSecret string `koanf:"webhook_secret"`
}

type IntegrationConfig struct {
	ResendAPIKey string `koanf:"resend_api_key"`
}

type AWSConfig struct {
	Region          string `koanf:"region"`
	AccessKeyID     string `koanf:"access_key_id"`
	SecretAccessKey string `koanf:"secret_access_key"`
	UploadBucket    string `koanf:"upload_bucket"`
	EndpointURL     string `koanf:"endpoint_url"`
}

type WebhookConfig struct {
	WorkerConcurrency int    `koanf:"worker_concurrency"`
	DeliveryTimeout   string `koanf:"delivery_timeout"`
	MaxRetries        int    `koanf:"max_retries"`
	RetryInitialDelay string `koanf:"retry_initial_delay"`
	RetryMaxDelay     string `koanf:"retry_max_delay"`
	RetryPollInterval string `koanf:"retry_poll_interval"`
	RetryConcurrency  int    `koanf:"retry_concurrency"`
}

type PrimaryConfig struct {
	Env string `koanf:"env" validate:"required"`
}

type Config struct {
	Primary       PrimaryConfig       `koanf:"primary" validate:"required"`
	Server        ServerConfig        `koanf:"server" validate:"required"`
	Database      DatabaseConfig      `koanf:"database" validate:"required"`
	Redis         RedisConfig         `koanf:"redis" validate:"required"`
	ClickHouse    ClickHouseConfig    `koanf:"clickhouse"`
	Clerk         ClerkConfig         `koanf:"auth" validate:"required"`
	Stripe        StripeConfig        `koanf:"stripe"`
	Auth          InternalAuthConfig  `koanf:"auth_internal"`
	Integration   IntegrationConfig   `koanf:"integration"`
	Observability ObservabilityConfig `koanf:"observability"`
	AWS           AWSConfig           `koanf:"aws"`
	Webhook       WebhookConfig       `koanf:"webhook"`
}

// LoadConfig initializes application configuration from environment variables.
func LoadConfig() (*Config, error) {
	k := koanf.New(".")

	err := k.Load(
		env.ProviderWithValue(
			"FLUX_",
			".",
			func(key, value string) (string, any) {
				return strings.ToLower(
					strings.TrimPrefix(key, "FLUX_"),
				), value
			},
		),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	cfg := &Config{}
	if err := k.Unmarshal("", cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Apply fail-closed security logic for production
	if strings.ToLower(cfg.Primary.Env) == "production" || strings.ToLower(cfg.Primary.Env) == "prod" {
		if cfg.Stripe.WebhookSecret == "" || cfg.Stripe.WebhookSecret == "whsec_test_secret" {
			return nil, fmt.Errorf("stripe webhook secret cannot be empty or test in production")
		}
		if cfg.Stripe.SecretKey == "" || cfg.Stripe.SecretKey == "sk_test_secret" {
			return nil, fmt.Errorf("stripe secret key cannot be empty or test in production")
		}
	}

	// Apply defaults where missing
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Server.FrontendURL == "" {
		cfg.Server.FrontendURL = "http://localhost:3000"
	}
	if cfg.Server.PlatformDomain == "" {
		cfg.Server.PlatformDomain = "flux.ly"
	}
	if cfg.Redis.AnalyticsStream == "" {
		cfg.Redis.AnalyticsStream = "analytics:events"
	}
	if cfg.ClickHouse.Host == "" {
		cfg.ClickHouse.Host = "localhost"
	}
	if cfg.ClickHouse.Port == "" {
		cfg.ClickHouse.Port = "9000"
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.Auth.JWTSecret == "" {
		cfg.Auth.JWTSecret = "super-secret-default-flux-key-change-in-prod"
	}
	if cfg.Stripe.WebhookSecret == "" {
		cfg.Stripe.WebhookSecret = "whsec_test_secret"
	}
	if cfg.Stripe.SecretKey == "" {
		cfg.Stripe.SecretKey = "sk_test_secret"
	}
	if cfg.Webhook.WorkerConcurrency <= 0 {
		cfg.Webhook.WorkerConcurrency = 10
	}
	if cfg.Webhook.DeliveryTimeout == "" {
		cfg.Webhook.DeliveryTimeout = "10s"
	}
	if cfg.Webhook.MaxRetries <= 0 {
		cfg.Webhook.MaxRetries = 5 // max total attempts = initial + 5 retries = 6
	}
	if cfg.Webhook.RetryInitialDelay == "" {
		cfg.Webhook.RetryInitialDelay = "5s"
	}
	if cfg.Webhook.RetryMaxDelay == "" {
		cfg.Webhook.RetryMaxDelay = "1h"
	}
	if cfg.Webhook.RetryPollInterval == "" {
		cfg.Webhook.RetryPollInterval = "5s"
	}
	if cfg.Webhook.RetryConcurrency <= 0 {
		cfg.Webhook.RetryConcurrency = 5
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// GetDatabaseURL computes the DSN
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Name, c.Database.SSLMode)
}

// GetRedisURL computes the Redis connection string
func (c *Config) GetRedisURL() string {
	return c.Redis.Address
}

// GetClickHouseURL computes the ClickHouse connection string
func (c *Config) GetClickHouseURL() string {
	return fmt.Sprintf("%s:%s", c.ClickHouse.Host, c.ClickHouse.Port)
}

// GetClerkSecretKey returns the Clerk secret
func (c *Config) GetClerkSecretKey() string {
	return c.Clerk.SecretKey
}
