package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/middleware"
	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

const secret = "secret-person-marker"
const parent = "00-11111111111111111111111111111111-2222222222222222-01"

type fixture struct {
	e      *echo.Echo
	spans  *tracetest.InMemoryExporter
	reader *sdkmetric.ManualReader
	logs   *bytes.Buffer
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	spans := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(spans))
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()); _ = mp.Shutdown(context.Background()) })
	logs := new(bytes.Buffer)
	log := logger.NewLogger(config.DefaultObservabilityConfig(), logs, nil)
	s := &server.Server{Role: config.RoleAPI, Logger: &log}
	tel := &observability.Telemetry{Tracer: tp.Tracer("flux.http"), Meter: mp.Meter("flux.http"), Propagator: observability.TraceparentPropagator{}}
	state, err := trace.ParseTraceState("private=" + secret)
	if err != nil {
		t.Fatal(err)
	}
	link := trace.Link{SpanContext: trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceState: state, TraceFlags: trace.FlagsSampled}), Attributes: []attribute.KeyValue{attribute.String("private", secret)}}
	tm := middleware.NewTracingMiddleware(s, tel, link)
	global := middleware.NewGlobalMiddlewares(s)
	e := echo.New()
	e.HTTPErrorHandler = global.GlobalErrorHandler
	e.Use(middleware.RequestID(), tm.EnhanceTracing(), middleware.NewContextEnhancer(s).EnhanceContext(), global.RequestLogger(), global.Recover())
	e.GET("/status", func(c echo.Context) error {
		if baggage.FromContext(c.Request().Context()).Len() != 0 {
			t.Error("baggage reached handler")
		}
		if c.Request().Header.Get("tracestate") != "" || c.Request().Header.Get("baggage") != "" {
			t.Error("private propagation headers reached handler")
		}
		c.Set(middleware.UserIDKey, secret)
		return c.NoContent(http.StatusOK)
	})
	e.GET("/ready", func(c echo.Context) error { return errors.New(secret) })
	e.GET("/live", func(c echo.Context) error { panic(secret) })
	e.GET("/openapi", func(c echo.Context) error { return echo.NewHTTPError(400, secret) })
	e.POST("/status", func(c echo.Context) error {
		return errs.NewBadRequestError("Validation failed", true, nil, []errs.FieldError{{Field: "name", Error: "is required"}}, nil)
	})
	return fixture{e, spans, reader, logs}
}

func request(f fixture, method, target, requestID, correlationID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	req.Header.Set("traceparent", parent)
	req.Header.Set("tracestate", "private="+secret)
	req.Header.Set("baggage", "private="+secret)
	req.Header.Set("Authorization", "Bearer "+secret)
	req.Header.Set("Cookie", "session="+secret)
	req.Header.Set("User-Agent", secret)
	req.Header.Set("X-Forwarded-For", "192.0.2.71")
	req.Header.Set(middleware.RequestIDHeader, requestID)
	req.Header.Set("X-Correlation-ID", correlationID)
	member, _ := baggage.NewMember("private", secret)
	bag, _ := baggage.New(member)
	req = req.WithContext(baggage.ContextWithBaggage(req.Context(), bag))
	rec := httptest.NewRecorder()
	f.e.ServeHTTP(rec, req)
	return rec
}

func TestHTTPInjectedTelemetryAndCorrelation(t *testing.T) {
	f := newFixture(t)
	requestID, correlationID := uuid.NewString(), uuid.NewString()
	rec := request(f, "GET", "/status?token="+secret, requestID, correlationID)
	if rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	if rec.Header().Get(middleware.RequestIDHeader) != requestID || rec.Header().Get("X-Correlation-ID") != correlationID {
		t.Fatal("response correlation mismatch")
	}
	spans := f.spans.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("got %d spans", len(spans))
	}
	span := spans[0]
	if span.SpanKind != trace.SpanKindServer || span.Name != "http.request" {
		t.Fatal("missing server span")
	}
	if span.Parent.SpanID().String() != "2222222222222222" || span.SpanContext.TraceID().String() != "11111111111111111111111111111111" {
		t.Fatal("parent lost")
	}
	if span.Parent.TraceState().Len() != 0 || span.SpanContext.TraceState().Len() != 0 {
		t.Fatal("span tracestate retained")
	}
	if len(span.Links) != 1 || span.Links[0].SpanContext.TraceState().Len() != 0 || len(span.Links[0].Attributes) != 0 {
		t.Fatal("unsafe Link retained")
	}
	attrs := attrMap(span.Attributes)
	if attrs["request_id"] != requestID || attrs["correlation_id"] != correlationID || attrs["http.route"] != "/status" {
		t.Fatal(attrs)
	}
	var record map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(f.logs.Bytes()), &record); err != nil {
		t.Fatal(err)
	}
	if record["request_id"] != requestID || record["correlation_id"] != correlationID || record["trace_id"] != span.SpanContext.TraceID().String() || record["span_id"] != span.SpanContext.SpanID().String() {
		t.Fatal(record)
	}
	assertPrivateAbsent(t, string(f.logs.Bytes())+fmt.Sprint(spans))
}

func TestHTTPInvalidIDsAreReplaced(t *testing.T) {
	for _, invalid := range []string{"", secret, strings.Repeat("a", 8192), uuid.Nil.String(), "11111111-1111-1111-1111-11111111111A"} {
		t.Run(fmt.Sprint(len(invalid))+invalid[:min(len(invalid), 8)], func(t *testing.T) {
			f := newFixture(t)
			rec := request(f, "GET", "/status", invalid, invalid)
			for _, header := range []string{middleware.RequestIDHeader, "X-Correlation-ID"} {
				value := rec.Header().Get(header)
				id, err := uuid.Parse(value)
				if err != nil || id == uuid.Nil || id.String() != value || len(value) != 36 {
					t.Fatalf("invalid response ID %q", value)
				}
				if invalid != "" && value == invalid {
					t.Fatal("unsafe ID echoed")
				}
			}
			attrs := attrMap(f.spans.GetSpans()[0].Attributes)
			if attrs["request_id"] != rec.Header().Get(middleware.RequestIDHeader) || attrs["correlation_id"] != rec.Header().Get("X-Correlation-ID") {
				t.Fatal("span ID mismatch")
			}
		})
	}
}

func TestHTTPRejectsDuplicateIDsAndUntrustedContext(t *testing.T) {
	f := newFixture(t)
	first, second := uuid.NewString(), uuid.NewString()
	req := httptest.NewRequest("GET", "/status", nil)
	req.Header.Add(middleware.RequestIDHeader, first)
	req.Header.Add(middleware.RequestIDHeader, second)
	req.Header.Add("X-Correlation-ID", first)
	req.Header.Add("X-Correlation-ID", second)
	req.Header.Set("traceparent", "00-00000000000000000000000000000000-2222222222222222-01")
	state, _ := trace.ParseTraceState("private=" + secret)
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceState: state})
	req = req.WithContext(trace.ContextWithSpanContext(req.Context(), sc))
	rec := httptest.NewRecorder()
	f.e.ServeHTTP(rec, req)
	if rec.Code != 200 || rec.Header().Get(middleware.RequestIDHeader) == first || rec.Header().Get("X-Correlation-ID") == first {
		t.Fatal("duplicate IDs were trusted")
	}
	span := f.spans.GetSpans()[0]
	if span.Parent.IsValid() || span.SpanContext.TraceID() == sc.TraceID() || span.SpanContext.TraceState().Len() != 0 {
		t.Fatal("invalid ingress inherited untrusted context")
	}
	assertPrivateAbsent(t, f.logs.String()+fmt.Sprint(f.spans.GetSpans()))
}

func TestHTTPBoundedLabelsAndSafeFailures(t *testing.T) {
	f := newFixture(t)
	for i := 0; i < 40; i++ {
		request(f, "GET", fmt.Sprintf("/status?token=%s-%d", secret, i), "", "")
		request(f, "GET", fmt.Sprintf("/%s-%d", secret, i), "", "")
	}
	for _, route := range []string{"/ready", "/live"} {
		rec := request(f, "GET", route, "", "")
		if rec.Code != 500 || strings.Contains(rec.Body.String(), secret) {
			t.Fatalf("unsafe failure %d %s", rec.Code, rec.Body)
		}
	}
	validation := request(f, "POST", "/status", "", "")
	if validation.Code != 400 || !strings.Contains(validation.Body.String(), "Validation failed") || !strings.Contains(validation.Body.String(), "is required") {
		t.Fatal("public validation regressed", validation.Body)
	}
	var data metricdata.ResourceMetrics
	if err := f.reader.Collect(context.Background(), &data); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, scope := range data.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name != "flux.http.requests" {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok {
				t.Fatal("counter not captured")
			}
			if len(sum.DataPoints) != 5 {
				t.Fatalf("label sets %d", len(sum.DataPoints))
			}
			for _, p := range sum.DataPoints {
				count += int(p.Value)
				if p.Attributes.Len() != 4 {
					t.Fatal("unbounded labels", p.Attributes)
				}
				attrs := attrMap(p.Attributes.ToSlice())
				if _, exists := attrs["request_id"]; exists {
					t.Fatal("ID metric label")
				}
				route := attrs["http.route"]
				if route != "/status" && route != "unmatched" && route != "/ready" && route != "/live" {
					t.Fatal("raw route", route)
				}
			}
		}
	}
	if count != 83 {
		t.Fatalf("requests %d", count)
	}
	assertPrivateAbsent(t, f.logs.String()+fmt.Sprint(data)+fmt.Sprint(f.spans.GetSpans()))
}

func TestHTTPUnknownEchoErrorIsPrivate(t *testing.T) {
	f := newFixture(t)
	rec := request(f, "GET", "/openapi", "", "")
	if rec.Code != 400 || strings.Contains(rec.Body.String(), secret) {
		t.Fatal("unknown Echo error escaped", rec.Body)
	}
	assertPrivateAbsent(t, f.logs.String()+fmt.Sprint(f.spans.GetSpans()))
}

func attrMap(attrs []attribute.KeyValue) map[string]any {
	result := map[string]any{}
	for _, a := range attrs {
		result[string(a.Key)] = a.Value.AsInterface()
	}
	return result
}
func assertPrivateAbsent(t *testing.T, output string) {
	t.Helper()
	for _, forbidden := range []string{secret, "192.0.2.71", "user_agent", "user_id", "Authorization", "Cookie", "tracestate", "baggage", "?token="} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("private marker retained: %s", forbidden)
		}
	}
}
