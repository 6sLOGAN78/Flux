//nolint:testpackage // These tests verify package-private lifecycle and failure-injection seams.
package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
)

type configProvider struct {
	values map[string]any
	err    error
}

func TestLinksCursorConfigOwnedByAPI(t *testing.T) {
	for _, key := range []string{
		"", "PRIVATE-INVALID-KEY", base64.StdEncoding.EncodeToString([]byte("short")),
		base64.StdEncoding.EncodeToString(make([]byte, 33)),
	} {
		values := configValues()
		values["links"].(map[string]any)["cursor_key"] = key
		_, err := loadConfigForRole(configProvider{values: values}, RoleAPI)
		if err == nil {
			t.Fatal("invalid cursor signing key accepted")
		}
		if strings.Contains(fmt.Sprint(err), key) && key != "" {
			t.Fatal("key value exposed in startup diagnostic")
		}
		for _, role := range []Role{RoleRedirector, RoleWorker, RoleMigrator} {
			if _, roleErr := loadConfigForRole(configProvider{values: values}, role); roleErr != nil {
				t.Fatal("unowned cursor key blocked role")
			}
		}
	}
}

func TestInvitationKeyRingBoundsAndPrivateDiagnostics(t *testing.T) {
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))
	oversized := make(map[string]string)
	for i := range 33 {
		oversized[fmt.Sprintf("key%d", i)] = key
	}
	encoded, err := json.Marshal(oversized)
	if err != nil {
		t.Fatal("fixture encoding failed")
	}
	for _, ring := range []string{
		`null`, `[]`, `{"fixture":null}`, `{"fixture":123}`, `{"fixture":"PRIVATE-KEY-CANARY"}`,
		`{"fixture":"` + base64.StdEncoding.EncodeToString(make([]byte, 31)) + `"}`,
		`{"fixture":"` + base64.StdEncoding.EncodeToString(make([]byte, 33)) + `"}`,
		`{"fixture":"` + key + `\n"}`, string(encoded), strings.Repeat(" ", 16385),
	} {
		values := configValues()
		values["invitations"].(map[string]any)["encryption_keys"] = ring
		_, loadErr := loadConfigForRole(configProvider{values: values}, RoleAPI)
		if loadErr == nil || strings.Contains(fmt.Sprint(loadErr, errors.Unwrap(loadErr)), "PRIVATE-KEY-CANARY") {
			t.Fatal("malformed key ring accepted or disclosed")
		}
	}
}

func TestLinksOperatorConfigOwnedByAPI(t *testing.T) {
	for _, host := range []string{
		"", "https://go.flux.test", "go.flux.test:443", "127.0.0.1", "localhost", "bad_host.example", "-bad.example",
		"links.localhost", "links.local", "links.internal", "metadata.google.internal",
	} {
		values := configValues()
		values["links"] = map[string]any{
			"cursor_key": base64.StdEncoding.EncodeToString(make([]byte, 32)), "managed_host": host,
		}
		if _, err := loadConfigForRole(configProvider{values: values}, RoleAPI); err == nil {
			t.Fatal("invalid managed host accepted")
		}
		for _, role := range []Role{RoleRedirector, RoleWorker, RoleMigrator} {
			if _, roleErr := loadConfigForRole(configProvider{values: values}, role); roleErr != nil {
				t.Fatal("unowned links configuration blocked role")
			}
		}
	}
	values := configValues()
	values["links"] = map[string]any{
		"cursor_key":   base64.StdEncoding.EncodeToString(make([]byte, 32)),
		"managed_host": "GO.FLUX.TEST.", "blocked_hosts": []string{"BLOCKED.EXAMPLE."},
	}
	cfg, err := loadConfigForRole(configProvider{values: values}, RoleAPI)
	if err != nil || cfg.Links.ManagedHost != "go.flux.test" || cfg.Links.BlockedHosts[0] != "blocked.example" {
		t.Fatal("operator host normalization failed")
	}
}

func TestLinksEnvironmentPolicy(t *testing.T) {
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if strings.HasPrefix(key, "FLUX_") {
			t.Setenv(key, "")
			if err := os.Unsetenv(key); err != nil {
				t.Fatal(err)
			}
		}
	}
	for section, value := range configValues() {
		for key, field := range value.(map[string]any) {
			if section != "links" {
				t.Setenv("FLUX_"+strings.ToUpper(section+"."+key), fmt.Sprint(field))
			}
		}
	}
	t.Setenv("FLUX_LINKS.CURSOR_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	t.Setenv("FLUX_LINKS.MANAGED_HOST", "BÜCHER.EXAMPLE.")
	t.Setenv("FLUX_LINKS.BLOCKED_HOSTS", " BLOCKED.EXAMPLE., Bücher.EXAMPLE ")
	cfg, err := LoadConfigForRole(RoleAPI)
	if err != nil || cfg.Links.ManagedHost != "xn--bcher-kva.example" ||
		strings.Join(cfg.Links.BlockedHosts, ",") != "blocked.example,xn--bcher-kva.example" {
		t.Fatal("environment policy failed normalization")
	}
	t.Setenv("FLUX_LINKS.BLOCKED_HOSTS", "blocked.example,,other.example")
	if _, err = LoadConfigForRole(RoleAPI); err == nil {
		t.Fatal("empty blocked-host list entry accepted")
	}
	t.Setenv("FLUX_LINKS.MANAGED_HOST", "")
	if _, err = LoadConfigForRole(RoleAPI); err == nil {
		t.Fatal("missing owned host accepted")
	}
	if _, err = LoadConfigForRole(RoleRedirector); err != nil {
		t.Fatal("unowned policy blocked redirector")
	}
}

func TestObservabilityCompatibleKeysAndSettings(t *testing.T) {
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if strings.HasPrefix(key, "FLUX_") {
			t.Setenv(key, "")
			if err := os.Unsetenv(key); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Setenv("FLUX_PRIMARY.ENV", "test")
	t.Setenv("FLUX_OBSERVABILITY.NEW_RELIC.LICENSE_KEY", "SECRET-MARKER")
	t.Setenv("FLUX_OBSERVABILITY.NEW_RELIC.DEBUG_LOGGING", "true")
	t.Setenv("FLUX_OBSERVABILITY.OTLP.ENABLED", "true")
	t.Setenv("FLUX_OBSERVABILITY.OTLP.ENDPOINT", "http://localhost:4318")
	t.Setenv("FLUX_OBSERVABILITY.OTLP.EXPORT_TIMEOUT", "750ms")
	t.Setenv("FLUX_OBSERVABILITY.OTLP.SAMPLE_RATIO", "0.25")
	cfg, err := LoadConfigForRole(RoleRedirector)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Observability.NewRelic.LicenseKey != "SECRET-MARKER" || !cfg.Observability.NewRelic.DebugLogging {
		t.Fatal("legacy keys no longer bind")
	}
	s := cfg.Observability.TelemetrySettings()
	if !s.Enabled ||
		s.Endpoint != "http://localhost:4318" ||
		s.ExportTimeout != 750*time.Millisecond ||
		s.SampleRatio != .25 ||
		s.Environment != "test" {
		t.Fatalf("mapping changed: %+v", s)
	}
}

func TestObservabilityTypedSafeValidation(t *testing.T) {
	for _, mutate := range []func(*ObservabilityConfig){
		func(c *ObservabilityConfig) { c.OTLP.Endpoint = "http://user:SECRET-MARKER@collector" },
		func(c *ObservabilityConfig) { c.OTLP.Endpoint = "http://collector/?token=SECRET-MARKER" },
		func(c *ObservabilityConfig) { c.OTLP.Endpoint = "SECRET-MARKER" },
		func(c *ObservabilityConfig) { c.OTLP.Endpoint = "http://collector:99999" },
		func(c *ObservabilityConfig) { c.OTLP.ExportTimeout = -time.Second },
		func(c *ObservabilityConfig) { c.OTLP.ExportInterval = time.Hour },
		func(c *ObservabilityConfig) { c.OTLP.BatchSize = c.OTLP.QueueSize + 1 },
		func(c *ObservabilityConfig) { c.Logging.Level = "SECRET-MARKER" },
	} {
		c := DefaultObservabilityConfig()
		c.OTLP.Enabled = true
		mutate(c)
		err := c.Validate()
		var typed *ConfigError
		if !errors.As(err, &typed) || typed.Stage != "observability" || errors.Unwrap(typed) == nil {
			t.Fatalf("missing typed wrapped error: %v", err)
		}
		if strings.Contains(fmt.Sprint(err, errors.Unwrap(typed)), "SECRET-MARKER") {
			t.Fatal("validation exposed input")
		}
	}
	for _, enabled := range []bool{false, true} {
		c := DefaultObservabilityConfig()
		c.OTLP.Enabled = enabled
		if err := c.Validate(); err != nil {
			t.Fatalf("optional endpoint rejected: %v", err)
		}
	}
	c := DefaultObservabilityConfig()
	c.OTLP.Endpoint = "SECRET-MARKER"
	if err := c.Validate(); err != nil || c.TelemetrySettings().Endpoint != "" {
		t.Fatalf("disabled export became a dependency: %v", err)
	}
}

func (p configProvider) Read() (map[string]any, error) { return p.values, p.err }
func (p configProvider) ReadBytes() ([]byte, error)    { return nil, p.err }

func configValues() map[string]any {
	return map[string]any{
		"primary": map[string]any{"env": "test"},
		"invitations": map[string]any{"active_key_id": "fixture",
			"encryption_keys": `{"fixture":"` + base64.StdEncoding.EncodeToString(make([]byte, 32)) + `"}`,
			"sender":          "invites@example.test", "public_origin": "https://app.flux.test"},
		"links": map[string]any{
			"cursor_key":   base64.StdEncoding.EncodeToString(make([]byte, 32)),
			"managed_host": "go.flux.test", "blocked_hosts": []string{"blocked.example"},
		},
		"server": map[string]any{"port": "8080",
			"read_timeout":         5,
			"write_timeout":        5,
			"idle_timeout":         5,
			"cors_allowed_origins": []string{"https://example.test"}},
		"database": map[string]any{"host": "db.test",
			"port":               5432,
			"user":               "flux",
			"password":           "SECRET-MARKER",
			"name":               "flux",
			"ssl_mode":           "disable",
			"max_open_conns":     5,
			"max_idle_conns":     1,
			"conn_max_lifetime":  60,
			"conn_max_idle_time": 30},
		"auth":        map[string]any{"secret_key": "SECRET-MARKER"},
		"redis":       map[string]any{"address": "localhost:6379"},
		"integration": map[string]any{"resend_api_key": "SECRET-MARKER"},
	}
}

//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
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
				delete(p.values, "database")
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
			value = environmentFixtureValue(field, value)
			t.Setenv("FLUX_"+strings.ToUpper(section+"."+field), fmt.Sprint(value))
		}
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("existing keys did not bind: %v (%v)", err, errors.Unwrap(err))
	}
	if cfg.Database.Host != "db.test" ||
		cfg.Server.ReadTimeout != 5 ||
		cfg.Integration.ResendAPIKey != "SECRET-MARKER" {
		t.Fatalf("existing keys changed: %v", cfg.Database.Host)
	}
	if cfg.Observability.Environment != "test" || cfg.Observability.ServiceName != "flux" {
		t.Fatalf("primary metadata not preserved")
	}
	if cfg.Links.ManagedHost != "go.flux.test" || len(cfg.Links.BlockedHosts) != 2 {
		t.Fatal("links environment policy did not bind")
	}
}

func environmentFixtureValue(field string, value any) any {
	if field == "cors_allowed_origins" {
		return "https://example.test"
	}
	if field == "blocked_hosts" {
		return "blocked.example,another.example"
	}
	return value
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
	for _, role := range []Role{RoleWorker, Role("unknown")} {
		if _, err187 := loadConfigForRole(configProvider{values: values}, role); err187 == nil {
			t.Fatalf("incomplete/unknown role %q accepted", role)
		}
	}
	for _, field := range []string{"host", "port", "user", "name", "ssl_mode"} {
		t.Run(field, func(t *testing.T) {
			fields := values["database"].(map[string]any)
			original := fields[field]
			delete(fields, field)
			defer func() { fields[field] = original }()
			_, err197 := loadConfigForRole(configProvider{values: values}, RoleMigrator)
			var typed *ConfigError
			var validationErrors validator.ValidationErrors
			if !errors.As(err197,
				&typed) ||
				typed.Stage != "validate" ||
				!errors.As(err197,
					&validationErrors) {
				t.Fatalf("missing database field did not retain typed cause: %v", err197)
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

func TestRoleConfigOnlyOwnedDependencies(t *testing.T) {
	for _, role := range []Role{RoleAPI, RoleRedirector, RoleWorker} {
		t.Run(string(role), func(t *testing.T) {
			values := configValues()
			delete(values, "auth")
			if role == RoleWorker {
				delete(values, "links")
			}
			if role != RoleWorker {
				delete(values, "integration")
				delete(values, "redis")
			}
			if role == RoleRedirector {
				delete(values, "database")
			}
			cfg, err := loadConfigForRole(configProvider{values: values}, role)
			if err != nil {
				t.Fatalf("unused secrets blocked %s: %v (%v)", role, err, errors.Unwrap(err))
			}
			settings := cfg.ForRole(role)
			if settings.ListenAddress == "" ||
				settings.DrainTimeout <= 0 ||
				settings.ReadinessTimeout <= 0 ||
				cfg.Server.ReadTimeout != 5 {
				t.Fatalf("role settings missing or legacy fallback changed: %+v", settings)
			}
		})
	}
}

func TestRoleConfigOverridesAndValidation(t *testing.T) {
	for _, role := range []Role{RoleAPI, RoleRedirector, RoleWorker} {
		values := configValues()
		values[string(role)] = map[string]any{"listen_address": "127.0.0.1:9099",
			"drain_timeout":     "3s",
			"readiness_timeout": "500ms"}
		cfg, err := loadConfigForRole(configProvider{values: values}, role)
		if err != nil {
			t.Fatal(err)
		}
		if settings := cfg.ForRole(role); settings.ListenAddress != "127.0.0.1:9099" ||
			settings.DrainTimeout.String() != "3s" ||
			settings.ReadinessTimeout.String() != "500ms" {
			t.Fatalf("role overrides lost: %+v", settings)
		}
		for _, key := range []string{"listen_address", "drain_timeout", "readiness_timeout"} {
			fields := values[string(role)].(map[string]any)
			old := fields[key]
			if key == "listen_address" {
				fields[key] = "SECRET-MARKER"
			} else {
				fields[key] = "-1s"
			}
			_, err272 := loadConfigForRole(configProvider{values: values}, role)
			if err272 == nil || strings.Contains(err272.Error(), "SECRET-MARKER") {
				t.Fatalf("invalid %s accepted or leaked: %v", key, err272)
			}
			fields[key] = old
		}
	}
	values := configValues()
	delete(values, "redis")
	delete(values, "integration")
	values["api"] = map[string]any{"producer_enabled": true}
	if _, err := loadConfigForRole(configProvider{values: values}, RoleAPI); err == nil {
		t.Fatal("declared producer accepted without Redis")
	}
	delete(values, "api")
	if _, err := loadConfigForRole(configProvider{values: values}, RoleWorker); err == nil {
		t.Fatal("worker accepted without Redis/email")
	}
}

func TestRoleEnvironmentSettings(t *testing.T) {
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if strings.HasPrefix(key, "FLUX_") {
			t.Setenv(key, "")
			if err := os.Unsetenv(key); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Setenv("FLUX_PRIMARY.ENV", "test")
	t.Setenv("FLUX_SERVER.PORT", "8181")
	t.Setenv("FLUX_SERVER.READ_TIMEOUT", "7")
	t.Setenv("FLUX_REDIRECTOR.LISTEN_ADDRESS", "127.0.0.1:9191")
	t.Setenv("FLUX_REDIRECTOR.DRAIN_TIMEOUT", "2s")
	t.Setenv("FLUX_REDIRECTOR.READINESS_TIMEOUT", "250ms")
	cfg, err := LoadConfigForRole(RoleRedirector)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Redirector.ListenAddress != "127.0.0.1:9191" ||
		cfg.Redirector.DrainTimeout.String() != "2s" ||
		cfg.Redirector.ReadinessTimeout.String() != "250ms" ||
		cfg.Server.ReadTimeout != 7 {
		t.Fatalf("environment overrides changed: %+v", cfg.Redirector)
	}
	values := map[string]any{"primary": map[string]any{"env": "test"},
		"invitations": map[string]any{"active_key_id": "fixture",
			"encryption_keys": `{"fixture":"` + base64.StdEncoding.EncodeToString(make([]byte, 32)) + `"}`,
			"sender":          "invites@example.test", "public_origin": "https://app.flux.test"},
		"redis":       map[string]any{"address": "localhost:6379"},
		"integration": map[string]any{"resend_api_key": "test-key"}}
	values["database"] = configValues()["database"]
	worker, err := loadConfigForRole(configProvider{values: values}, RoleWorker)
	if err != nil || worker.Worker.ListenAddress != "127.0.0.1:8082" {
		t.Fatalf("worker management default changed: %v", err)
	}
}

func TestRoleUnusedSettingsCannotBlockStartup(t *testing.T) {
	values := configValues()
	values["database"].(map[string]any)["port"] = "SECRET-MARKER"
	values["api"] = map[string]any{"drain_timeout": "SECRET-MARKER"}
	values["worker"] = map[string]any{"readiness_timeout": "SECRET-MARKER"}
	if _, err := loadConfigForRole(configProvider{values: values}, RoleRedirector); err != nil {
		t.Fatalf("unused configuration blocked redirector: %v", err)
	}
}
