package observability

import (
	"context"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"strings"
)

// CleanSpanContext retains only standard IDs, sampled flag and remote identity.
func CleanSpanContext(sc trace.SpanContext) trace.SpanContext {
	if !sc.IsValid() {
		return trace.SpanContext{}
	}
	return trace.NewSpanContext(trace.SpanContextConfig{TraceID: sc.TraceID(), SpanID: sc.SpanID(), TraceFlags: sc.TraceFlags() & trace.FlagsSampled, Remote: sc.IsRemote()})
}

// TraceparentPropagator never passes incoming baggage or tracestate to extraction.
type TraceparentPropagator struct{}

func (TraceparentPropagator) Fields() []string { return []string{"traceparent"} }
func (TraceparentPropagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	ctx = baggage.ContextWithBaggage(ctx, baggage.Baggage{})
	ctx = trace.ContextWithSpanContext(ctx, trace.SpanContext{})
	ctx = propagation.TraceContext{}.Extract(ctx, propagation.MapCarrier{"traceparent": carrier.Get("traceparent")})
	return trace.ContextWithSpanContext(ctx, CleanSpanContext(trace.SpanContextFromContext(ctx)))
}
func (TraceparentPropagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	// Set empty values first to overwrite stale metadata even on generic carriers.
	carrier.Set("tracestate", "")
	carrier.Set("baggage", "")
	carrier.Set("traceparent", "")
	switch c := carrier.(type) {
	case propagation.MapCarrier:
		for key := range c {
			if strings.EqualFold(key, "traceparent") || strings.EqualFold(key, "tracestate") || strings.EqualFold(key, "baggage") {
				delete(c, key)
			}
		}
	case propagation.HeaderCarrier:
		for key := range c {
			if strings.EqualFold(key, "traceparent") || strings.EqualFold(key, "tracestate") || strings.EqualFold(key, "baggage") {
				delete(c, key)
			}
		}
	}
	propagation.TraceContext{}.Inject(trace.ContextWithSpanContext(ctx, CleanSpanContext(trace.SpanContextFromContext(ctx))), carrier)
}

type correlationKey struct{}

// Correlation holds UUID metadata safe for logs and spans, never metric labels.
type Correlation struct {
	RequestID     string
	CorrelationID string
}

func WithCorrelation(ctx context.Context, requestID, correlationID string) context.Context {
	if !validUUID(requestID) {
		requestID = uuid.NewString()
	}
	if !validUUID(correlationID) {
		correlationID = requestID
	}
	return context.WithValue(ctx, correlationKey{}, Correlation{RequestID: requestID, CorrelationID: correlationID})
}
func CorrelationFromContext(ctx context.Context) Correlation {
	ids, _ := ctx.Value(correlationKey{}).(Correlation)
	return ids
}
