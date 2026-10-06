package observability

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// SafeError never formats an unrecognized provider or driver error.
func SafeError(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, context.DeadlineExceeded):
		return "operation timed out"
	case errors.Is(err, context.Canceled):
		return "operation canceled"
	default:
		return "operation failed"
	}
}

func oneOf(value string, values ...string) bool {
	for _, v := range values {
		if value == v {
			return true
		}
	}
	return false
}
func validUUID(value string) bool {
	id, err := uuid.Parse(value)
	return err == nil && id != uuid.Nil && id.String() == value
}

// SafeOperation maps all uncontrolled text to a constant bounded operation name.
func SafeOperation(value string) string {
	if oneOf(value, "http.request", "database.query", "database.connect", "redis.command", "job.enqueue", "job.process", "email.send", "dependency.check", "telemetry.export", "telemetry.shutdown") {
		return value
	}
	return "operation"
}

func safeAttribute(a attribute.KeyValue, metric bool) bool {
	key := string(a.Key)
	if key == "http.response.status_code" || key == "retry.count" {
		return a.Value.Type() == attribute.INT64 && a.Value.AsInt64() >= 0 && a.Value.AsInt64() <= 999
	}
	if a.Value.Type() != attribute.STRING {
		return false
	}
	v := a.Value.AsString()
	switch key {
	case "request_id", "correlation_id":
		return !metric && validUUID(v)
	case "operation":
		return SafeOperation(v) == v
	case "process.role":
		return oneOf(v, "api", "redirector", "worker", "migrator")
	case "http.request.method":
		return oneOf(v, "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS")
	case "http.route":
		return oneOf(v, "/live", "/ready", "/status", "/openapi", "/openapi.json", "/docs", "unmatched")
	case "outcome":
		return oneOf(v, "success", "error", "canceled", "timeout")
	case "dependency":
		return oneOf(v, "postgres", "redis", "email", "telemetry")
	case "error.category":
		return oneOf(v, "unknown", "timeout", "canceled", "unavailable", "validation")
	case "error.stage":
		return oneOf(v, "connect", "query", "enqueue", "process", "send", "export", "shutdown", "validate")
	case "job.type":
		return v == "email:welcome"
	case "db.system.name":
		return v == "postgresql"
	case "db.operation.name":
		return oneOf(v, "SELECT", "INSERT", "UPDATE", "DELETE", "BEGIN", "COMMIT", "ROLLBACK", "PING")
	default:
		return false
	}
}

// SanitizeAttributes applies a value-aware allowlist. Metric labels exclude IDs.
func SanitizeAttributes(attrs []attribute.KeyValue, metric bool) []attribute.KeyValue {
	out := make([]attribute.KeyValue, 0, 16)
	for _, a := range attrs {
		if safeAttribute(a, metric) {
			out = append(out, a)
			if len(out) == 16 {
				break
			}
		}
	}
	return out
}

type safeFailure struct {
	stage string
	cause error
}

func (e safeFailure) Error() string { return "telemetry " + e.stage + ": " + SafeError(e.cause) }
func (e safeFailure) Unwrap() error { return e.cause }
func safeFailureFor(stage string, err error) error {
	if err == nil {
		return nil
	}
	return safeFailure{stage: stage, cause: err}
}
