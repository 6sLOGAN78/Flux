package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

// Role selects the resources owned by a process.
type Role string

const (
	RoleAPI        Role = "api"
	RoleRedirector Role = "redirector"
	RoleWorker     Role = "worker"
	RoleMigrator   Role = "migrator"
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
	return LoadConfigForRole(RoleAPI)
}

// LoadConfigForRole preserves FLUX environment names and validates owned resources.
func LoadConfigForRole(role Role) (*Config, error) {
	return loadConfigForRole(env.Provider("FLUX_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "FLUX_"))
	}), role)
}

func loadConfig(provider koanf.Provider) (*Config, error) {
	return loadConfigForRole(provider, RoleAPI)
}

func loadConfigForRole(provider koanf.Provider, role Role) (*Config, error) {
	switch role {
	case RoleAPI, RoleRedirector, RoleWorker, RoleMigrator:
	default:
		return nil, &ConfigError{Stage: "role", cause: fmt.Errorf("unsupported process role")}
	}
	k := koanf.New(".")
	if err := k.Load(provider, nil); err != nil {
		return nil, &ConfigError{Stage: "load", cause: err}
	}

	mainConfig := &Config{Observability: DefaultObservabilityConfig()}
	if role == RoleMigrator {
		// A one-shot command needs neither HTTP settings nor pool sizing.
		mainConfig.Database = DatabaseConfig{MaxOpenConns: 1, MaxIdleConns: 1, ConnMaxLifetime: 60, ConnMaxIdleTime: 30}
	}
	if err := k.Unmarshal("", mainConfig); err != nil {
		return nil, &ConfigError{Stage: "unmarshal", cause: err}
	}

	validate := validator.New()
	// Observability owns its validation and optional vendor credentials. Validate
	// only the existing required sections here, keeping stage errors distinct.
	excluded := []string{"Observability"}
	switch role {
	case RoleMigrator:
		excluded = append(excluded, "Server", "Auth", "Redis", "Integration")
	case RoleRedirector:
		excluded = append(excluded, "Database", "Auth", "Redis", "Integration")
	case RoleWorker:
		excluded = append(excluded, "Server", "Database", "Auth")
	}
	if err := validate.StructExcept(mainConfig, excluded...); err != nil {
		return nil, &ConfigError{Stage: "validate", cause: err}
	}
	if role == RoleMigrator {
		if err := validate.Var(mainConfig.Database.Port, "min=1,max=65535"); err != nil {
			return nil, &ConfigError{Stage: "validate", cause: err}
		}
	}

	// Override service name and environment from primary config
	mainConfig.Observability.ServiceName = "flux"
	mainConfig.Observability.Environment = mainConfig.Primary.Env

	// Validate observability config
	if err := mainConfig.Observability.Validate(); err != nil {
		return nil, &ConfigError{Stage: "observability", cause: err}
	}
	if role == RoleMigrator {
		logging := mainConfig.Observability.Logging
		if (logging.Format != "json" && logging.Format != "console") || mainConfig.Observability.HealthChecks.Timeout <= 0 {
			return nil, &ConfigError{Stage: "observability", cause: fmt.Errorf("invalid logging format or operation timeout")}
		}
	}

	return mainConfig, nil
}
