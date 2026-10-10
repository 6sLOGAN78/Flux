//nolint:testpackage // Exercise the package-private barrier against actual pinned collector output.
package app

import (
	"bytes"
	"context"
	"encoding/hex"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	logcollector "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metriccollector "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracecollector "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logpb "go.opentelemetry.io/proto/otlp/logs/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

func TestCollectorExportCompletenessBarrier(t *testing.T) {
	endpoint, post, _ := proofCollector(t)
	entered, release := make(chan struct{}), make(chan struct{})
	var enterOnce, releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	defer unblock()
	post.Lock()
	post.beforeCapture = func(wire proofWire) {
		if !collectorContainsRetry(wire) {
			return
		}
		enterOnce.Do(func() { close(entered) })
		<-release
	}
	post.Unlock()
	spans, httpSpan := collectorBarrierSpans(t)
	collectorProofPost(t, endpoint, "/v1/traces", collectorBarrierTraceRequest(spans[:3]))
	collectorProofPost(t, endpoint, "/v1/logs", &logcollector.ExportLogsServiceRequest{
		ResourceLogs: []*logpb.ResourceLogs{{ScopeLogs: []*logpb.ScopeLogs{{
			LogRecords: []*logpb.LogRecord{{Body: &commonpb.AnyValue{
				Value: &commonpb.AnyValue_StringValue{StringValue: "job.process"},
			}}},
		}}}},
	})
	collectorProofPost(t, endpoint, "/v1/metrics", &metriccollector.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricpb.ResourceMetrics{{ScopeMetrics: []*metricpb.ScopeMetrics{{
			Metrics: []*metricpb.Metric{{Name: "http.server.requests", Data: &metricpb.Metric_Sum{
				Sum: &metricpb.Sum{AggregationTemporality: metricpb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
					IsMonotonic: true, DataPoints: []*metricpb.NumberDataPoint{{
						Value: &metricpb.NumberDataPoint_AsInt{AsInt: 1},
					}}},
			}}},
		}}}},
	})
	// Actual initial traces, logs and metrics are already exported before the
	// collector's final retry batch reaches the deliberately held HTTP sink.
	proofEventually(t, func() bool { return len(post.snapshot()) >= 3 })
	collectorProofPost(t, endpoint, "/v1/traces", collectorBarrierTraceRequest(spans[3:]))
	select {
	case <-entered:
	case <-time.After(8 * time.Second):
		t.Fatal("controlled collector retry batch never reached output")
	}
	if proofExportComplete(post.snapshot()) {
		t.Fatal("collector barrier accepted unrelated exports before final retry spans")
	}
	unblock()
	proofEventually(t, func() bool { return proofExportComplete(post.snapshot()) })
	checkProofLineage(t, proofSignals(t, post.snapshot()), httpSpan, false)
}

func collectorContainsRetry(wire proofWire) bool {
	if wire.path != "/v1/traces" {
		return false
	}
	var request tracecollector.ExportTraceServiceRequest
	if proto.Unmarshal(wire.body, &request) != nil {
		return false
	}
	for _, resource := range request.GetResourceSpans() {
		for _, scope := range resource.GetScopeSpans() {
			for _, span := range scope.GetSpans() {
				for _, attr := range span.GetAttributes() {
					if span.GetName() == "job.process" && attr.GetKey() == "retry.count" && attr.GetValue().GetIntValue() == 1 {
						return true
					}
				}
			}
		}
	}
	return false
}

func collectorBarrierTraceRequest(spans []*tracepb.Span) *tracecollector.ExportTraceServiceRequest {
	return &tracecollector.ExportTraceServiceRequest{ResourceSpans: []*tracepb.ResourceSpans{{
		ScopeSpans: []*tracepb.ScopeSpans{{Spans: spans}},
	}}}
}

func collectorBarrierSpans(t *testing.T) ([]*tracepb.Span, string) {
	t.Helper()
	traceID, err := hex.DecodeString("4bf92f3577b34da6a3ce929d0e0e4736")
	if err != nil {
		t.Fatal("invalid public trace fixture")
	}
	parentID, err := hex.DecodeString("00f067aa0ba902b7")
	if err != nil {
		t.Fatal("invalid public parent fixture")
	}
	httpSpanID := bytes.Repeat([]byte{0x24}, 8)
	attrs := func() []*commonpb.KeyValue {
		return []*commonpb.KeyValue{
			{Key: "request_id", Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: proofRequestID}}},
			{Key: "correlation_id", Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: proofCorrelationID}}},
		}
	}
	spans := []*tracepb.Span{{Name: "http.request", TraceId: traceID, SpanId: httpSpanID,
		ParentSpanId: parentID, Flags: 1, Attributes: attrs()}}
	for i := range 4 {
		jobAttrs := append(attrs(), &commonpb.KeyValue{Key: "retry.count",
			Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: int64(i / 2)}}})
		spans = append(spans, &tracepb.Span{Name: "job.process", TraceId: traceID,
			SpanId: bytes.Repeat([]byte{byte(i + 1)}, 8), ParentSpanId: httpSpanID, Flags: 1, Attributes: jobAttrs})
	}
	return spans, hex.EncodeToString(httpSpanID)
}

func collectorProofPost(t *testing.T, endpoint, path string, message proto.Message) {
	t.Helper()
	body, err := proto.Marshal(message)
	if err != nil {
		t.Fatal("encode official collector fixture protobuf")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+path, bytes.NewReader(body))
	if err != nil {
		t.Fatal("construct loopback collector fixture request")
	}
	req.Header.Set("Content-Type", "application/x-protobuf")
	response, err := (&http.Client{Timeout: 2 * time.Second}).Do(req)
	if err != nil {
		t.Fatal("send loopback collector fixture request")
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)
	if response.StatusCode != http.StatusOK {
		t.Fatal("collector rejected official fixture protobuf")
	}
}
