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
				t.Fatal("original cause was lost")
			}
			if stage == "load" && !errors.Is(err, loadCause) {
				t.Fatal("load cause was lost")
			}
			if stage == "validate" {
				var validationErrors validator.ValidationErrors
				if !errors.As(err, &validationErrors) {
					t.Fatal("validator cause was lost")
				}
			}
			if strings.Contains(fmt.Sprint(err), "SECRET-MARKER") {
				t.Fatal("public diagnostic leaked a secret")
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
		t.Fatal("primary metadata not preserved")
	}
}
