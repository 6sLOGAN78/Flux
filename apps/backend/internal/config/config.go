package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

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
	API           RoleConfig           `koanf:"api"`
	Redirector    RoleConfig           `koanf:"redirector"`
	Worker        RoleConfig           `koanf:"worker"`
}

// RoleConfig controls an independently owned listener and operation deadlines.
// Durations use Go duration strings in FLUX_<ROLE>.* environment variables.
type RoleConfig struct {
	ListenAddress    string        `koanf:"listen_address"`
	DrainTimeout     time.Duration `koanf:"drain_timeout"`
	ReadinessTimeout time.Duration `koanf:"readiness_timeout"`
	ProducerEnabled  bool          `koanf:"producer_enabled"`
}

// ForRole returns the selected role's normalized settings.
func (c *Config) ForRole(role Role) RoleConfig {
	switch role {
	case RoleAPI:
		return c.API
	case RoleRedirector:
		return c.Redirector
	case RoleWorker:
		return c.Worker
	default:
		return RoleConfig{}
	}
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
	// Do not decode numeric/duration settings owned by another process. An
	// invalid database port or worker deadline cannot block a redirector.
	for _, other := range []Role{RoleAPI, RoleRedirector, RoleWorker} {
		if other != role {
			k.Delete(string(other))
		}
	}
	if role == RoleRedirector || role == RoleWorker {
		k.Delete("database")
	}
	if role == RoleMigrator {
		k.Delete("server")
	}

	mainConfig := &Config{Observability: DefaultObservabilityConfig()}
	if role == RoleMigrator {
		// A one-shot command needs neither HTTP settings nor pool sizing.
		mainConfig.Database = DatabaseConfig{MaxOpenConns: 1, MaxIdleConns: 1, ConnMaxLifetime: 60, ConnMaxIdleTime: 30}
	}
	if err := k.Unmarshal("", mainConfig); err != nil {
		return nil, &ConfigError{Stage: "unmarshal", cause: err}
	}
	if role != RoleMigrator {
		if err := normalizeRole(mainConfig, role, k); err != nil {
			return nil, &ConfigError{Stage: "validate", cause: err}
		}
	}

	validate := validator.New()
	// Observability owns its validation and optional vendor credentials. Validate
	// only the existing required sections here, keeping stage errors distinct.
	excluded := []string{"Observability", "API", "Redirector", "Worker", "Auth", "Server"}
	switch role {
	case RoleAPI:
		excluded = append(excluded, "Integration")
		if !mainConfig.API.ProducerEnabled {
			excluded = append(excluded, "Redis")
		}
	case RoleMigrator:
		excluded = append(excluded, "Server", "Redis", "Integration")
	case RoleRedirector:
		excluded = append(excluded, "Server", "Database", "Redis", "Integration")
	case RoleWorker:
		excluded = append(excluded, "Server", "Database")
	}
	if err := validate.StructExcept(mainConfig, excluded...); err != nil {
		return nil, &ConfigError{Stage: "validate", cause: err}
	}
	if role == RoleMigrator || role == RoleAPI {
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

func normalizeRole(cfg *Config, role Role, k *koanf.Koanf) error {
	settings := cfg.ForRole(role)
	prefix := string(role) + "."
	if !k.Exists(prefix + "listen_address") {
		port := cfg.Server.Port
		if port == "" {
			port = map[Role]string{RoleAPI: "8080", RoleRedirector: "8081", RoleWorker: "8082"}[role]
		}
		host := ""
		if role == RoleWorker {
			host = "127.0.0.1"
		}
		settings.ListenAddress = net.JoinHostPort(host, port)
	}
	if !k.Exists(prefix + "drain_timeout") {
		settings.DrainTimeout = 30 * time.Second
	}
	if !k.Exists(prefix + "readiness_timeout") {
		settings.ReadinessTimeout = cfg.Observability.HealthChecks.Timeout
	}
	_, port, err := net.SplitHostPort(settings.ListenAddress)
	if err != nil {
		return fmt.Errorf("invalid role listen address")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 0 || portNumber > 65535 || settings.DrainTimeout <= 0 || settings.ReadinessTimeout <= 0 {
		return fmt.Errorf("invalid role port or deadline")
	}
	if !k.Exists("server.read_timeout") {
		cfg.Server.ReadTimeout = 5
	}
	if !k.Exists("server.write_timeout") {
		cfg.Server.WriteTimeout = 10
	}
	if !k.Exists("server.idle_timeout") {
		cfg.Server.IdleTimeout = 60
	}
	if cfg.Server.ReadTimeout <= 0 || cfg.Server.WriteTimeout <= 0 || cfg.Server.IdleTimeout <= 0 {
		return fmt.Errorf("invalid HTTP timeout")
	}
	// Keep legacy transport fields populated for middleware and existing callers.
	cfg.Server.Port = port
	switch role {
	case RoleAPI:
		cfg.API = settings
	case RoleRedirector:
		cfg.Redirector = settings
	case RoleWorker:
		cfg.Worker = settings
	}
	return nil
}
