package middleware

import (
	"net/http"
	"time"

	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

const (
	maxSpanLinks         = 16
	httpStatusClassWidth = 100
)

// TracingMiddleware records safe request traces and metrics.
type TracingMiddleware struct {
	tracer   trace.Tracer
	requests metric.Int64Counter
	duration metric.Float64Histogram
	role     string
	links    []trace.Link
}

// NewTracingMiddleware accepts injected API objects. The legacy registry may
// pass its inert adapter until plan 01-17; no vendor provider is initialized.
func NewTracingMiddleware(s *server.Server, injected any, links ...trace.Link) *TracingMiddleware {
	tracer := tracenoop.NewTracerProvider().Tracer("flux.http")
	meter := metricnoop.NewMeterProvider().Meter("flux.http")
	if tel, ok := injected.(*observability.Telemetry); ok && tel != nil {
		if tel.Tracer != nil {
			tracer = tel.Tracer
		}
		if tel.Meter != nil {
			meter = tel.Meter
		}
	}
	requests, err := meter.Int64Counter("flux.http.requests")
	if err != nil {
		s.Logger.Warn().
			Str("operation",
				"http.request").
			Str("error.category",
				"unavailable").
			Msg("http.request")
		requests, _ = metricnoop.NewMeterProvider().Meter("flux.http").Int64Counter("flux.http.requests")
	}
	duration, err := meter.Float64Histogram("flux.http.duration", metric.WithUnit("s"))
	if err != nil {
		s.Logger.Warn().
			Str("operation",
				"http.request").
			Str("error.category",
				"unavailable").
			Msg("http.request")
		duration, _ = metricnoop.NewMeterProvider().Meter("flux.http").Float64Histogram("flux.http.duration")
	}
	role := string(s.Role)
	switch role {
	case "api", "redirector", "worker", "migrator":
	default:
		role = "api"
	}
	cleanLinks := make([]trace.Link, 0, min(len(links), maxSpanLinks))
	for _, link := range links {
		if len(cleanLinks) == maxSpanLinks {
			break
		}
		sc := observability.CleanSpanContext(link.SpanContext)
		if sc.IsValid() {
			cleanLinks = append(cleanLinks,
				trace.Link{SpanContext: sc,
					Attributes: observability.SanitizeAttributes(link.Attributes,
						false)})
		}
	}
	return &TracingMiddleware{tracer: tracer, requests: requests, duration: duration, role: role, links: cleanLinks}
}

// EnhanceTracing only extracts traceparent and records closed operational fields.
func (tm *TracingMiddleware) EnhanceTracing() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			for key := range req.Header {
				if equalPropagationHeader(key) {
					delete(req.Header, key)
				}
			}
			ctx := observability.TraceparentPropagator{}.Extract(req.Context(),
				propagation.HeaderCarrier(req.Header))
			ids := observability.CorrelationFromContext(ctx)
			ctx,
				span := tm.tracer.Start(ctx,
				"http.request",
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithLinks(tm.links...),
				trace.WithAttributes(attribute.String("request_id",
					ids.RequestID),
					attribute.String("correlation_id",
						ids.CorrelationID)))
			defer span.End()
			c.SetRequest(req.WithContext(ctx))
			err := next(c)
			if err != nil {
				c.Error(err)
			}
			status := c.Response().Status
			attrs := []attribute.KeyValue{attribute.String("http.request.method",
				SafeMethod(req.Method)),
				attribute.String("http.route",
					SafeRoute(c)),
				attribute.String("process.role",
					tm.role),
				attribute.Int("http.response.status_code",
					status)}
			span.SetAttributes(attrs...)
			if status >= http.StatusInternalServerError {
				span.SetStatus(codes.Error, "")
				span.SetAttributes(attribute.String("error.category", "unknown"))
			}
			// Status classes prevent individually chosen response codes expanding labels.
			attrs[3] = attribute.Int("http.response.status_code",
				status/httpStatusClassWidth*httpStatusClassWidth)
			tm.requests.Add(ctx, 1, metric.WithAttributes(attrs...))
			tm.duration.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(attrs...))
			return nil
		}
	}
}

// SafeRoute uses the matched Echo template, then the shared closed route policy.
func SafeRoute(c echo.Context) string {
	route := c.Path()
	attrs := observability.SanitizeAttributes([]attribute.KeyValue{attribute.String("http.route", route)}, true)
	if len(attrs) == 0 {
		return "unmatched"
	}
	return route
}

// SafeMethod maps arbitrary methods to a bounded HTTP method vocabulary.
func SafeMethod(method string) string {
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return method
	default:
		return "GET"
	}
}
