package config

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/6sLOGAN78/flux/internal/observability"
)

const (
	defaultTelemetryQueueSize      = 256
	defaultTelemetryBatchSize      = 64
	defaultTelemetryExportTimeout  = 2 * time.Second
	defaultTelemetryExportInterval = 5 * time.Second
	defaultSlowQueryThreshold      = 100 * time.Millisecond
	defaultHealthInterval          = 30 * time.Second
	defaultHealthTimeout           = 5 * time.Second
)

// ObservabilityConfig contains logging, readiness, and optional export settings.
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
	Endpoint       string        `koanf:"endpoint"`
	SampleRatio    float64       `koanf:"sample_ratio"`
	QueueSize      int           `koanf:"queue_size"`
	BatchSize      int           `koanf:"batch_size"`
	ExportTimeout  time.Duration `koanf:"export_timeout"`
	ExportInterval time.Duration `koanf:"export_interval"`
	Enabled        bool          `koanf:"enabled"`
}

// LoggingConfig controls local structured log output.
type LoggingConfig struct {
	Level              string        `koanf:"level" validate:"required"`
	Format             string        `koanf:"format" validate:"required"`
	SlowQueryThreshold time.Duration `koanf:"slow_query_threshold"`
}

// NewRelicConfig retains legacy environment bindings without initializing vendor SDKs.
type NewRelicConfig struct {
	LicenseKey                string `koanf:"license_key" validate:"required"`
	AppLogForwardingEnabled   bool   `koanf:"app_log_forwarding_enabled"`
	DistributedTracingEnabled bool   `koanf:"distributed_tracing_enabled"`
	DebugLogging              bool   `koanf:"debug_logging"`
}

// HealthChecksConfig bounds dependency probes and operation timeouts.
type HealthChecksConfig struct {
	Checks   []string      `koanf:"checks"`
	Interval time.Duration `koanf:"interval" validate:"min=1s"`
	Timeout  time.Duration `koanf:"timeout" validate:"min=1s"`
	Enabled  bool          `koanf:"enabled"`
}

// DefaultObservabilityConfig returns conservative logging and telemetry defaults.
func DefaultObservabilityConfig() *ObservabilityConfig {
	return &ObservabilityConfig{
		ServiceName: "flux",
		Environment: environmentDevelopment,
		OTLP: OTLPConfig{SampleRatio: 1,
			QueueSize:      defaultTelemetryQueueSize,
			BatchSize:      defaultTelemetryBatchSize,
			ExportTimeout:  defaultTelemetryExportTimeout,
			ExportInterval: defaultTelemetryExportInterval},
		Logging: LoggingConfig{
			Level:              logLevelInfo,
			Format:             logFormatJSON,
			SlowQueryThreshold: defaultSlowQueryThreshold,
		},
		NewRelic: NewRelicConfig{
			LicenseKey:                "",
			AppLogForwardingEnabled:   true,
			DistributedTracingEnabled: true,
			DebugLogging:              false, // Disabled by default to avoid mixed log formats
		},
		HealthChecks: HealthChecksConfig{
			Enabled:  true,
			Interval: defaultHealthInterval,
			Timeout:  defaultHealthTimeout,
			Checks:   []string{"database", "redis"},
		},
	}
}

// Validate checks logging and telemetry settings without exposing secret values.
func (c *ObservabilityConfig) Validate() error {
	fail := func(message string) error {
		return &ConfigError{Stage: stageObservability,
			cause: errors.New(message)}
	}
	if c.ServiceName == "" {
		return fail("service_name is required")
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true, logLevelInfo: true, "warn": true, "error": true,
	}
	if !validLevels[c.Logging.Level] {
		return fail("invalid logging level")
	}
	if c.Logging.Format != logFormatJSON && c.Logging.Format != "console" {
		return fail("invalid logging format")
	}

	// Validate slow query threshold
	if c.Logging.SlowQueryThreshold < 0 {
		return fail("logging slow_query_threshold must be non-negative")
	}
	s := c.TelemetrySettings()
	switch s.Environment {
	case "local", environmentDevelopment, "test", "staging", environmentProduction:
	default:
		return fail("invalid telemetry environment")
	}
	if s.QueueSize < 1 ||
		s.QueueSize > 4096 ||
		s.BatchSize < 1 ||
		s.BatchSize > s.QueueSize ||
		s.SampleRatio < 0 ||
		s.SampleRatio > 1 ||
		s.SampleRatio != s.SampleRatio ||
		s.ExportTimeout < time.Millisecond ||
		s.ExportTimeout > 30*time.Second ||
		s.ExportInterval < time.Millisecond ||
		s.ExportInterval > time.Minute {
		return fail("invalid telemetry bounds")
	}
	if err := validateTelemetryEndpoint(s.Endpoint); err != nil {
		return fail(err.Error())
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
	return observability.Settings{Enabled: o.Enabled,
		Endpoint:       endpoint,
		Environment:    c.Environment,
		SampleRatio:    o.SampleRatio,
		QueueSize:      o.QueueSize,
		BatchSize:      o.BatchSize,
		ExportTimeout:  o.ExportTimeout,
		ExportInterval: o.ExportInterval}
}

// GetLogLevel returns the configured level with environment-aware defaults.
func (c *ObservabilityConfig) GetLogLevel() string {
	switch c.Environment {
	case environmentProduction:
		if c.Logging.Level == "" {
			return logLevelInfo
		}
	case environmentDevelopment:
		if c.Logging.Level == "" {
			return "debug"
		}
	}
	return c.Logging.Level
}

// IsProduction reports whether the configured environment is production.
func (c *ObservabilityConfig) IsProduction() bool {
	return c.Environment == environmentProduction
}

func validateTelemetryEndpoint(endpoint string) error {
	if endpoint == "" {
		return nil
	}
	u, err := url.Parse(endpoint)
	if err != nil ||
		(u.Scheme != "http" &&
			u.Scheme != "https") ||
		u.Hostname() == "" ||
		u.User != nil ||
		u.RawQuery != "" ||
		u.ForceQuery ||
		u.Fragment != "" ||
		u.Opaque != "" {
		return errors.New("invalid telemetry endpoint")
	}
	if port := u.Port(); port != "" {
		n, err113 := strconv.Atoi(port)
		if err113 != nil || n < 1 || n > 65535 {
			return errors.New("invalid telemetry endpoint port")
		}
	}
	if u.Host[len(u.Host)-1] == ':' {
		return errors.New("invalid telemetry endpoint host")
	}
	return nil
}
