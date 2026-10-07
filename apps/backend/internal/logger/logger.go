// Package logger emits sanitized structured logs with bounded telemetry fields.
package logger

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
)

const (
	initialLogAttributeCapacity = 16
	logMetadataFieldCapacity    = 6
)

// LoggerService retains the legacy constructor log-sink adapter without ownership.
//
//nolint:revive // Preserve the established exported type name for existing callers.
type LoggerService struct{ LogSink otellog.Logger }

// NewLoggerService constructs the inert compatibility logging adapter.
func NewLoggerService(_ *config.ObservabilityConfig) *LoggerService { return &LoggerService{} }

// Shutdown stops intake and closes owned resources within the caller deadline.
// Shutdown retains the inert compatibility lifecycle hook.
func (*LoggerService) Shutdown() {}

// NewLoggerWithService constructs a logger using an optional telemetry sink.
func NewLoggerWithService(cfg *config.ObservabilityConfig, service *LoggerService) zerolog.Logger {
	var sink otellog.Logger
	if service != nil {
		sink = service.LogSink
	}
	return NewLogger(cfg, os.Stdout, sink)
}

const legacyWarning = "legacy observability settings are deprecated"

// NewLogger sends stdout JSON and OTel records through one sanitizing writer.
// Console remains a compatible setting; output is consistently structured JSON.
func NewLogger(cfg *config.ObservabilityConfig, output io.Writer, sink otellog.Logger) zerolog.Logger {
	if cfg == nil {
		cfg = config.DefaultObservabilityConfig()
	}
	if output == nil {
		output = os.Stdout
	}
	level, err := zerolog.ParseLevel(cfg.GetLogLevel())
	if err != nil {
		level = zerolog.InfoLevel
	}
	writer := &safeWriter{output: output, sink: sink}
	log := zerolog.New(writer).Level(level).Hook(contextHook{}).With().Timestamp().Logger()
	if cfg.NewRelic.LicenseKey != "" ||
		cfg.NewRelic.DebugLogging ||
		!cfg.NewRelic.AppLogForwardingEnabled ||
		!cfg.NewRelic.DistributedTracingEnabled {
		warning := log.Level(zerolog.WarnLevel)
		warning.Warn().Msg(legacyWarning)
	}
	return log
}

// WithContext attaches validated correlation metadata at event emission.
//
//nolint:revive // Preserve the existing logger-first public helper signature for all callers.
func WithContext(log zerolog.Logger, ctx context.Context) zerolog.Logger {
	return log.With().Ctx(ctx).Logger()
}

type contextHook struct{}

func (contextHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	ctx := e.GetCtx()
	if ctx == nil {
		return
	}
	sc := observability.CleanSpanContext(trace.SpanContextFromContext(ctx))
	if sc.IsValid() {
		e.Str("trace_id", sc.TraceID().String()).Str("span_id", sc.SpanID().String())
	}
	ids := observability.CorrelationFromContext(ctx)
	if ids.RequestID != "" {
		e.Str("request_id", ids.RequestID).Str("correlation_id", ids.CorrelationID)
	}
}

type safeWriter struct {
	output io.Writer
	sink   otellog.Logger
	mu     sync.Mutex
}

func (w *safeWriter) Write(p []byte) (int, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(p, &raw); err != nil {
		return 0, err
	}
	if _, exists := raw[zerolog.ErrorFieldName]; exists {
		delete(raw, "error.category")
	}
	attrs := logAttributes(raw)
	safe := make(map[string]any, len(attrs)+logMetadataFieldCapacity)
	for _, a := range attrs {
		safe[string(a.Key)] = a.Value.AsInterface()
	}
	var message, level string
	_ = json.Unmarshal(raw[zerolog.MessageFieldName], &message)
	_ = json.Unmarshal(raw[zerolog.LevelFieldName], &level)
	safe["message"] = observability.SafeOperation(message)
	if message == legacyWarning {
		safe["message"] = legacyWarning
	}
	parsedLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		parsedLevel = zerolog.InfoLevel
		level = "info"
	}
	safe["level"] = level
	now := time.Now().UTC()
	safe["time"] = now.Format(time.RFC3339Nano)
	if _, exists := raw[zerolog.ErrorFieldName]; exists {
		safe["error"] = observability.SafeError(io.ErrUnexpectedEOF)
		safe["error.category"] = "unknown"
	}
	ctx := logTraceContext(raw, safe)
	data, err := json.Marshal(safe)
	if err != nil {
		return 0, err
	}
	data = append(data, '\n')
	w.mu.Lock()
	defer w.mu.Unlock()
	if n, err165 := w.output.Write(data); err165 != nil {
		return 0, err165
	} else if n != len(data) {
		return 0, io.ErrShortWrite
	}
	if w.sink != nil {
		var record otellog.Record
		record.SetTimestamp(now)
		record.SetBody(attribute.StringValue(observability.SafeOperation(message)))
		record.SetSeverity(severity(parsedLevel))
		record.AddAttributes(attrs...)
		w.sink.Emit(ctx, record)
	}
	return len(p), nil
}

func logTraceContext(raw map[string]json.RawMessage, safe map[string]any) context.Context {
	ctx := context.Background()
	var tid, sid string
	_ = json.Unmarshal(raw["trace_id"], &tid)
	_ = json.Unmarshal(raw["span_id"], &sid)
	traceID, te := trace.TraceIDFromHex(tid)
	spanID, se := trace.SpanIDFromHex(sid)
	if te == nil &&
		se == nil &&
		traceID.IsValid() &&
		spanID.IsValid() &&
		traceID.String() == tid &&
		spanID.String() == sid {
		safe["trace_id"] = tid
		safe["span_id"] = sid
		ctx = trace.ContextWithSpanContext(ctx,
			trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID,
				SpanID: spanID}))
	}
	return ctx
}

func severity(level zerolog.Level) otellog.Severity {
	switch level {
	case zerolog.TraceLevel:
		return otellog.SeverityTrace
	case zerolog.DebugLevel:
		return otellog.SeverityDebug
	case zerolog.WarnLevel:
		return otellog.SeverityWarn
	case zerolog.ErrorLevel:
		return otellog.SeverityError
	case zerolog.FatalLevel, zerolog.PanicLevel:
		return otellog.SeverityFatal
	case zerolog.InfoLevel, zerolog.NoLevel, zerolog.Disabled:
		return otellog.SeverityInfo
	default:
		return otellog.SeverityInfo
	}
}

// NewPgxLogger rejects SQL, arguments and opaque driver errors at the sink.
func NewPgxLogger(level zerolog.Level) zerolog.Logger {
	return NewLogger(config.DefaultObservabilityConfig(), os.Stdout, nil).Level(level)
}

// GetPgxTraceLogLevel maps the configured logger level to PostgreSQL tracing verbosity.
func GetPgxTraceLogLevel(level zerolog.Level) int {
	switch level {
	case zerolog.DebugLevel:
		return pgxDebugTraceLevel
	case zerolog.InfoLevel:
		return pgxInfoTraceLevel
	case zerolog.WarnLevel:
		return pgxWarnTraceLevel
	case zerolog.ErrorLevel:
		return pgxErrorTraceLevel
	case zerolog.FatalLevel, zerolog.PanicLevel, zerolog.NoLevel, zerolog.Disabled, zerolog.TraceLevel:
		return 0
	default:
		return 0
	}
}

const (
	pgxDebugTraceLevel = 6
	pgxInfoTraceLevel  = 4
	pgxWarnTraceLevel  = 3
	pgxErrorTraceLevel = 2
)

func logAttributes(raw map[string]json.RawMessage) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, initialLogAttributeCapacity)
	keys := make([]string, 0, len(raw))
	for key := range raw {
		keys = append(keys, key)
	}
	// Correlation precedes bounded operational attributes, independently of map order.
	sort.Strings(keys)
	keys = append([]string{"request_id", "correlation_id"}, keys...)
	seen := make(map[string]bool, len(keys))
	for _, key := range keys {
		if seen[key] {
			continue
		}
		seen[key] = true
		value := raw[key]
		var s string
		if json.Unmarshal(value, &s) == nil {
			attrs = append(attrs, attribute.String(key, s))
			continue
		}
		var n int64
		if json.Unmarshal(value, &n) == nil {
			attrs = append(attrs, attribute.Int64(key, n))
		}
	}
	if _, exists := raw[zerolog.ErrorFieldName]; exists {
		// The original error has already been marshaled by zerolog. Never inspect it.
		attrs = append([]attribute.KeyValue{attribute.String("error.category", "unknown")}, attrs...)
	}
	attrs = observability.SanitizeAttributes(attrs, false)
	return attrs
}
