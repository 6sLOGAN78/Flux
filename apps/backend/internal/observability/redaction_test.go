package observability

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"strings"
	"testing"
)

const secret = "SECRET_SENTINEL"
const validID = "3b726791-abb7-4457-8d4c-265b6063b7a5"

func TestAllowlistAndSafeError(t *testing.T) {
	attrs := SanitizeAttributes([]attribute.KeyValue{attribute.String("operation", "http.request"), attribute.String("request_id", validID), attribute.String("http.route", "/ready?token="+secret), attribute.String("error.category", secret), attribute.String("sql.query", secret), attribute.String("authorization", secret), attribute.String("request_id", secret)}, false)
	set := attribute.NewSet(attrs...)
	if len(attrs) != 2 || strings.Contains(set.String(), secret) {
		t.Fatalf("unsafe attributes: %v", attrs)
	}
	if got := SafeError(errors.New(secret)); got != "operation failed" {
		t.Fatalf("unknown error exposed: %s", got)
	}
	if got := SafeError(context.DeadlineExceeded); got != "operation timed out" {
		t.Fatal(got)
	}
	if len(SanitizeAttributes([]attribute.KeyValue{attribute.String("request_id", validID)}, true)) != 0 {
		t.Fatal("metric contains correlation label")
	}
}

func TestTraceparentOnlyPropagation(t *testing.T) {
	p := TraceparentPropagator{}
	ts, _ := trace.ParseTraceState("vendor=" + secret)
	m, _ := baggage.NewMember("token", secret)
	b, _ := baggage.New(m)
	parent := trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{1}, TraceState: ts, TraceFlags: trace.FlagsSampled, Remote: true})
	ctx := baggage.ContextWithBaggage(trace.ContextWithSpanContext(context.Background(), parent), b)
	input := propagation.MapCarrier{"traceparent": "00-12345678901234567890123456789012-1234567890123456-01", "tracestate": "vendor=" + secret, "baggage": "token=" + secret}
	out := p.Extract(ctx, input)
	sc := trace.SpanContextFromContext(out)
	if sc.TraceID().String() != "12345678901234567890123456789012" || !sc.IsSampled() || !sc.IsRemote() || sc.TraceState().Len() != 0 || baggage.FromContext(out).Len() != 0 {
		t.Fatal("invalid safe extraction", sc)
	}
	carrier := propagation.MapCarrier{"tracestate": secret, "baggage": secret, "TraceState": secret, "Baggage": secret, "TraceParent": secret}
	p.Inject(out, carrier)
	if carrier.Get("traceparent") != input.Get("traceparent") || carrier.Get("tracestate") != "" || carrier.Get("baggage") != "" {
		t.Fatal(carrier)
	}
	for _, value := range []string{"invalid", "00-00000000000000000000000000000000-1234567890123456-01", "00-12345678901234567890123456789012-0000000000000000-01"} {
		if trace.SpanContextFromContext(p.Extract(ctx, propagation.MapCarrier{"traceparent": value})).IsValid() {
			t.Fatal("invalid parent accepted", value)
		}
	}
	out = WithCorrelation(out, secret, validID)
	ids := CorrelationFromContext(out)
	if ids.RequestID == secret || ids.CorrelationID != validID {
		t.Fatal(ids)
	}
}
