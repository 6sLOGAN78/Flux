package config

import (
	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"strings"
)

type Config struct {
	Primary       Primary              `koanf:"primary" validate:"required"`
	Server        ServerConfig         `koanf:"server" validate:"required"`
	Database      DatabaseConfig       `koanf:"database" validate:"required"`
	Auth          AuthConfig           `koanf:"auth" validate:"required"`
	Redis         RedisConfig          `koanf:"redis" validate:"required"`
	Integration   IntegrationConfig    `koanf:"integration" validate:"required"`
	Observability *ObservabilityConfig `koanf:"observability"`
}

type Primary struct {
	Env string `koanf:"env" validate:"required"`
}

type ServerConfig struct {
	Port               string   `koanf:"port" validate:"required"`
	ReadTimeout        int      `koanf:"read_timeout" validate:"required"`
	WriteTimeout       int      `koanf:"write_timeout" validate:"required"`
	IdleTimeout        int      `koanf:"idle_timeout" validate:"required"`
	CORSAllowedOrigins []string `koanf:"cors_allowed_origins" validate:"required"`
}

type DatabaseConfig struct {
	Host            string `koanf:"host" validate:"required"`
	Port            int    `koanf:"port" validate:"required"`
	User            string `koanf:"user" validate:"required"`
	Password        string `koanf:"password"`
	Name            string `koanf:"name" validate:"required"`
	SSLMode         string `koanf:"ssl_mode" validate:"required"`
	MaxOpenConns    int    `koanf:"max_open_conns" validate:"required"`
	MaxIdleConns    int    `koanf:"max_idle_conns" validate:"required"`
	ConnMaxLifetime int    `koanf:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime int    `koanf:"conn_max_idle_time" validate:"required"`
}
type RedisConfig struct {
	Address string `koanf:"address" validate:"required"`
}

type IntegrationConfig struct {
	ResendAPIKey string `koanf:"resend_api_key" validate:"required"`
}

type AuthConfig struct {
	SecretKey string `koanf:"secret_key" validate:"required"`
}

// ConfigError identifies a failed stage without exposing provider diagnostics.
// The original error remains available through errors.Is and errors.As.
type ConfigError struct {
	Stage string
	cause error
}

func (e *ConfigError) Error() string { return "configuration " + e.Stage + " failed" }
func (e *ConfigError) Unwrap() error { return e.cause }

func LoadConfig() (*Config, error) {
	return loadConfig(env.Provider("FLUX_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "FLUX_"))
	}))
}

func loadConfig(provider koanf.Provider) (*Config, error) {
	k := koanf.New(".")
	if err := k.Load(provider, nil); err != nil {
		return nil, &ConfigError{Stage: "load", cause: err}
	}

	mainConfig := &Config{Observability: DefaultObservabilityConfig()}
	if err := k.Unmarshal("", mainConfig); err != nil {
		return nil, &ConfigError{Stage: "unmarshal", cause: err}
	}

	validate := validator.New()
	// Observability owns its validation and optional vendor credentials. Validate
	// only the existing required sections here, keeping stage errors distinct.
	if err := validate.StructExcept(mainConfig, "Observability"); err != nil {
		return nil, &ConfigError{Stage: "validate", cause: err}
	}

	// Override service name and environment from primary config
	mainConfig.Observability.ServiceName = "flux"
	mainConfig.Observability.Environment = mainConfig.Primary.Env

	// Validate observability config
	if err := mainConfig.Observability.Validate(); err != nil {
		return nil, &ConfigError{Stage: "observability", cause: err}
	}

	return mainConfig, nil
}
