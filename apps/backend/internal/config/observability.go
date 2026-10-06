package config

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/6sLOGAN78/flux/internal/observability"
)

type ObservabilityConfig struct {
	ServiceName  string             `koanf:"service_name" validate:"required"`
	Environment  string             `koanf:"environment" validate:"required"`
	Logging      LoggingConfig      `koanf:"logging" validate:"required"`
	NewRelic     NewRelicConfig     `koanf:"new_relic" validate:"required"`
	HealthChecks HealthChecksConfig `koanf:"health_checks" validate:"required"`
	OTLP         OTLPConfig         `koanf:"otlp"`
}

// OTLPConfig binds compatible FLUX_OBSERVABILITY.OTLP.* names. Export is optional.
type OTLPConfig struct {
	Enabled        bool          `koanf:"enabled"`
	Endpoint       string        `koanf:"endpoint"`
	SampleRatio    float64       `koanf:"sample_ratio"`
	QueueSize      int           `koanf:"queue_size"`
	BatchSize      int           `koanf:"batch_size"`
	ExportTimeout  time.Duration `koanf:"export_timeout"`
	ExportInterval time.Duration `koanf:"export_interval"`
}

type LoggingConfig struct {
	Level              string        `koanf:"level" validate:"required"`
	Format             string        `koanf:"format" validate:"required"`
	SlowQueryThreshold time.Duration `koanf:"slow_query_threshold"`
}

type NewRelicConfig struct {
	LicenseKey                string `koanf:"license_key" validate:"required"`
	AppLogForwardingEnabled   bool   `koanf:"app_log_forwarding_enabled"`
	DistributedTracingEnabled bool   `koanf:"distributed_tracing_enabled"`
	DebugLogging              bool   `koanf:"debug_logging"`
}

type HealthChecksConfig struct {
	Enabled  bool          `koanf:"enabled"`
	Interval time.Duration `koanf:"interval" validate:"min=1s"`
	Timeout  time.Duration `koanf:"timeout" validate:"min=1s"`
	Checks   []string      `koanf:"checks"`
}

func DefaultObservabilityConfig() *ObservabilityConfig {
	return &ObservabilityConfig{
		ServiceName: "flux",
		Environment: "development",
		OTLP:        OTLPConfig{SampleRatio: 1, QueueSize: 256, BatchSize: 64, ExportTimeout: 2 * time.Second, ExportInterval: 5 * time.Second},
		Logging: LoggingConfig{
			Level:              "info",
			Format:             "json",
			SlowQueryThreshold: 100 * time.Millisecond,
		},
		NewRelic: NewRelicConfig{
			LicenseKey:                "",
			AppLogForwardingEnabled:   true,
			DistributedTracingEnabled: true,
			DebugLogging:              false, // Disabled by default to avoid mixed log formats
		},
		HealthChecks: HealthChecksConfig{
			Enabled:  true,
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
			Checks:   []string{"database", "redis"},
		},
	}
}

func (c *ObservabilityConfig) Validate() error {
	fail := func(message string) error { return &ConfigError{Stage: "observability", cause: errors.New(message)} }
	if c.ServiceName == "" {
		return fail("service_name is required")
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLevels[c.Logging.Level] {
		return fail("invalid logging level")
	}
	if c.Logging.Format != "json" && c.Logging.Format != "console" {
		return fail("invalid logging format")
	}

	// Validate slow query threshold
	if c.Logging.SlowQueryThreshold < 0 {
		return fail("logging slow_query_threshold must be non-negative")
	}
	s := c.TelemetrySettings()
	switch s.Environment {
	case "local", "development", "test", "staging", "production":
	default:
		return fail("invalid telemetry environment")
	}
	if s.QueueSize < 1 || s.QueueSize > 4096 || s.BatchSize < 1 || s.BatchSize > s.QueueSize || s.SampleRatio < 0 || s.SampleRatio > 1 || s.SampleRatio != s.SampleRatio || s.ExportTimeout < time.Millisecond || s.ExportTimeout > 30*time.Second || s.ExportInterval < time.Millisecond || s.ExportInterval > time.Minute {
		return fail("invalid telemetry bounds")
	}
	if s.Endpoint != "" {
		u, err := url.Parse(s.Endpoint)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" {
			return fail("invalid telemetry endpoint")
		}
		if port := u.Port(); port != "" {
			n, err := strconv.Atoi(port)
			if err != nil || n < 1 || n > 65535 {
				return fail("invalid telemetry endpoint port")
			}
		}
		if u.Host[len(u.Host)-1] == ':' {
			return fail("invalid telemetry endpoint host")
		}
	}

	return nil
}

// TelemetrySettings maps settings without introducing a collector readiness gate.
// Disabled endpoints are ignored, including stale legacy monitoring values.
func (c *ObservabilityConfig) TelemetrySettings() observability.Settings {
	o := c.OTLP
	endpoint := ""
	if o.Enabled {
		endpoint = o.Endpoint
	}
	return observability.Settings{Enabled: o.Enabled, Endpoint: endpoint, Environment: c.Environment, SampleRatio: o.SampleRatio, QueueSize: o.QueueSize, BatchSize: o.BatchSize, ExportTimeout: o.ExportTimeout, ExportInterval: o.ExportInterval}
}

func (c *ObservabilityConfig) GetLogLevel() string {
	switch c.Environment {
	case "production":
		if c.Logging.Level == "" {
			return "info"
		}
	case "development":
		if c.Logging.Level == "" {
			return "debug"
		}
	}
	return c.Logging.Level
}

func (c *ObservabilityConfig) IsProduction() bool {
	return c.Environment == "production"
}
