package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

type configProvider struct {
	values map[string]any
	err    error
}

func (p configProvider) Read() (map[string]any, error) { return p.values, p.err }
func (p configProvider) ReadBytes() ([]byte, error)    { return nil, p.err }

func configValues() map[string]any {
	return map[string]any{
		"primary":     map[string]any{"env": "test"},
		"server":      map[string]any{"port": "8080", "read_timeout": 5, "write_timeout": 5, "idle_timeout": 5, "cors_allowed_origins": []string{"https://example.test"}},
		"database":    map[string]any{"host": "db.test", "port": 5432, "user": "flux", "password": "SECRET-MARKER", "name": "flux", "ssl_mode": "disable", "max_open_conns": 5, "max_idle_conns": 1, "conn_max_lifetime": 60, "conn_max_idle_time": 30},
		"auth":        map[string]any{"secret_key": "SECRET-MARKER"},
		"redis":       map[string]any{"address": "localhost:6379"},
		"integration": map[string]any{"resend_api_key": "SECRET-MARKER"},
	}
}

func TestConfigStages(t *testing.T) {
	loadCause := errors.New("SECRET-MARKER load error")
	for _, stage := range []string{"load", "unmarshal", "validate", "observability"} {
		t.Run(stage, func(t *testing.T) {
			p := configProvider{values: configValues()}
			switch stage {
			case "load":
				p.err = loadCause
			case "unmarshal":
				p.values["database"].(map[string]any)["port"] = "SECRET-MARKER"
			case "validate":
				delete(p.values, "auth")
			case "observability":
				p.values["observability"] = map[string]any{"logging": map[string]any{"level": "SECRET-MARKER"}}
			}
			cfg, err := loadConfig(p)
			if cfg != nil || err == nil {
				t.Fatalf("expected rejected config, got %v, %v", cfg, err)
			}
			var typed *ConfigError
			if !errors.As(err, &typed) || typed.Stage != stage {
				t.Fatalf("want stage %s, got %v", stage, err)
			}
			if errors.Unwrap(typed) == nil {
				t.Fatalf("original cause was lost")
			}
			if stage == "load" && !errors.Is(err, loadCause) {
				t.Fatalf("load cause was lost")
			}
			if stage == "validate" {
				var validationErrors validator.ValidationErrors
				if !errors.As(err, &validationErrors) {
					t.Fatalf("validator cause was lost")
				}
			}
			if strings.Contains(fmt.Sprint(err), "SECRET-MARKER") {
				t.Fatalf("public diagnostic leaked a secret")
			}
		})
	}
}

func TestConfigExistingEnvironmentKeys(t *testing.T) {
	// Isolate the provider from developer environment without parallel env mutation.
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if strings.HasPrefix(key, "FLUX_") {
			t.Setenv(key, "")
			if err := os.Unsetenv(key); err != nil {
				t.Fatalf("unset isolated environment: %v", err)
			}
		}
	}
	for section, fields := range configValues() {
		for field, value := range fields.(map[string]any) {
			if field == "cors_allowed_origins" {
				value = "https://example.test"
			}
			t.Setenv("FLUX_"+strings.ToUpper(section+"."+field), fmt.Sprint(value))
		}
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("existing keys did not bind: %v (%v)", err, errors.Unwrap(err))
	}
	if cfg.Database.Host != "db.test" || cfg.Server.ReadTimeout != 5 || cfg.Integration.ResendAPIKey != "SECRET-MARKER" {
		t.Fatalf("existing keys changed: %v", cfg.Database.Host)
	}
	if cfg.Observability.Environment != "test" || cfg.Observability.ServiceName != "flux" {
		t.Fatalf("primary metadata not preserved")
	}
}

func TestConfigMigratorRole(t *testing.T) {
	values := configValues()
	delete(values, "auth")
	delete(values, "redis")
	delete(values, "integration")
	delete(values, "server")
	cfg, err := loadConfigForRole(configProvider{values: values}, RoleMigrator)
	if err != nil || cfg == nil {
		t.Fatalf("PostgreSQL-only configuration rejected: %v", err)
	}
	if cfg.Server.Port != "" || cfg.Auth.SecretKey != "" || cfg.Redis.Address != "" {
		t.Fatal("unused services were configured")
	}
	for _, role := range []Role{RoleAPI, RoleRedirector, RoleWorker, Role("unknown")} {
		if _, err := loadConfigForRole(configProvider{values: values}, role); err == nil {
			t.Fatalf("incomplete/unknown role %q accepted", role)
		}
	}
	for _, field := range []string{"host", "port", "user", "name", "ssl_mode"} {
		t.Run(field, func(t *testing.T) {
			fields := values["database"].(map[string]any)
			original := fields[field]
			delete(fields, field)
			defer func() { fields[field] = original }()
			_, err := loadConfigForRole(configProvider{values: values}, RoleMigrator)
			var typed *ConfigError
			var validationErrors validator.ValidationErrors
			if !errors.As(err, &typed) || typed.Stage != "validate" || !errors.As(err, &validationErrors) {
				t.Fatalf("missing database field did not retain typed cause: %v", err)
			}
		})
	}
}

func TestConfigMigratorLoggingAndTimeoutValidation(t *testing.T) {
	for _, values := range []map[string]any{
		{"observability": map[string]any{"logging": map[string]any{"level": "SECRET-MARKER"}}},
		{"observability": map[string]any{"logging": map[string]any{"format": "SECRET-MARKER"}}},
		{"observability": map[string]any{"health_checks": map[string]any{"timeout": "-1s"}}},
		{"database": map[string]any{"port": -1}},
	} {
		base := configValues()
		for section, fields := range values {
			if section == "database" {
				base[section].(map[string]any)["port"] = fields.(map[string]any)["port"]
			} else {
				base[section] = fields
			}
		}
		_, err := loadConfigForRole(configProvider{values: base}, RoleMigrator)
		if err == nil || strings.Contains(err.Error(), "SECRET-MARKER") {
			t.Fatalf("invalid logging/timeout accepted or leaked: %v", err)
		}
	}
}
