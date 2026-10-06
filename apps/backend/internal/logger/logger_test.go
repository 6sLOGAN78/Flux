package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

const sentinel = "SECRET-MARKER person@example.test postgres://user:password@db/private Bearer token"

type logCapture struct {
	mu      sync.Mutex
	records []sdklog.Record
}

func (e *logCapture) Export(_ context.Context, records []sdklog.Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, record := range records {
		e.records = append(e.records, record.Clone())
	}
	return nil
}
func (*logCapture) ForceFlush(context.Context) error { return nil }
func (*logCapture) Shutdown(context.Context) error   { return nil }

func TestJSONAndBridgeShareSafeContext(t *testing.T) {
	exporter := &logCapture{}
	tel, err := observability.New(context.Background(), observability.Settings{Enabled: true, Exporters: observability.Exporters{Log: exporter}}, "api")
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	cfg := config.DefaultObservabilityConfig()
	log := NewLogger(cfg, &output, tel.Logger)
	ts, _ := trace.ParseTraceState("vendor=SECRET-MARKER")
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceState: ts, TraceFlags: trace.FlagsSampled})
	requestID := "4c0f261c-7013-4c39-9b1e-8c7fd0203ec2"
	correlationID := "c1c6e4de-3c0e-49ec-8091-6b267dce5602"
	ctx := observability.WithCorrelation(trace.ContextWithSpanContext(context.Background(), sc), requestID, correlationID)
	log = WithContext(log, ctx)
	log.Error().Stack().Err(errors.New(sentinel)).Str("token", sentinel).Str("sql", sentinel).
		Str("request_id", sentinel).Str("trace_id", sentinel).Str("operation", "http.request").
		Str("http.route", "/ready").Int("http.response.status_code", 503).Msg(sentinel)
	if err := tel.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	var stdout map[string]any
	if err := json.Unmarshal(output.Bytes(), &stdout); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "SECRET-MARKER") || strings.Contains(output.String(), "person@example.test") || strings.Contains(output.String(), "password") {
		t.Fatal("stdout leaked private data")
	}
	for key, want := range map[string]string{"trace_id": sc.TraceID().String(), "span_id": sc.SpanID().String(), "request_id": requestID, "correlation_id": correlationID} {
		if stdout[key] != want {
			t.Fatalf("missing safe %s: %v", key, stdout)
		}
	}
	if stdout["error"] != "operation failed" || stdout["operation"] != "http.request" {
		t.Fatalf("operational diagnostic missing: %v", stdout)
	}
	if len(exporter.records) != 1 {
		t.Fatalf("got %d bridge records", len(exporter.records))
	}
	r := exporter.records[0]
	attrs := map[string]any{}
	r.WalkAttributes(func(a attribute.KeyValue) bool { attrs[string(a.Key)] = a.Value.AsInterface(); return true })
	if r.TraceID() != sc.TraceID() || r.SpanID() != sc.SpanID() || attrs["request_id"] != requestID || attrs["correlation_id"] != correlationID || attrs["operation"] != "http.request" {
		t.Fatalf("bridge lost shared context: %v", attrs)
	}
	if strings.Contains(fmt.Sprint(attrs, r.Body(), r.SeverityText()), "SECRET-MARKER") {
		t.Fatal("bridge leaked private data")
	}
}

func TestLoggerRejectsUnvalidatedFieldsAndWarnsSafely(t *testing.T) {
	var output bytes.Buffer
	cfg := config.DefaultObservabilityConfig()
	cfg.NewRelic.LicenseKey = sentinel
	cfg.NewRelic.DebugLogging = true
	service := NewLoggerService(cfg)
	if service.GetApplication() != nil {
		t.Fatal("legacy vendor was initialized")
	}
	log := NewLogger(cfg, &output, nil)
	log.Info().Str("request_id", sentinel).Str("correlation_id", sentinel).Str("trace_id", strings.Repeat("0", 32)).Str("span_id", strings.Repeat("0", 16)).Str("http.route", sentinel).Str("error.stage", sentinel).Str("stack", sentinel).Msg(sentinel)
	if strings.Contains(output.String(), "SECRET-MARKER") || strings.Contains(output.String(), "person@example.test") {
		t.Fatal("private values leaked")
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 || !strings.Contains(lines[0], "deprecated") {
		t.Fatalf("expected generic legacy warning: %s", output.String())
	}
	var event map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &event); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"request_id", "correlation_id", "trace_id", "span_id", "http.route", "error.stage", "stack"} {
		if _, ok := event[key]; ok {
			t.Fatalf("unvalidated field retained: %s", key)
		}
	}
}

func TestPGXCompatibilityLoggerSanitizesOutput(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = original; _ = w.Close(); _ = r.Close() })
	log := NewPgxLogger(zerolog.DebugLevel)
	log.Debug().Str("sql", sentinel).Interface("args", []string{sentinel}).Msg(sentinel)
	log.Error().Err(errors.New(sentinel)).Str("sql", sentinel).Msg(sentinel)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(output), "SECRET-MARKER") {
		t.Fatal("pgx fields escaped the sink")
	}
	if lines := bytes.Count(output, []byte("\n")); lines != 2 {
		t.Fatalf("missing pgx records: %d", lines)
	}
}

func TestConcurrentContextLoggersKeepRecordsSeparate(t *testing.T) {
	var output bytes.Buffer
	base := NewLogger(config.DefaultObservabilityConfig(), &output, nil)
	var workers sync.WaitGroup
	for i := 0; i < 20; i++ {
		workers.Go(func() {
			log := WithContext(base, observability.WithCorrelation(context.Background(), sentinel, sentinel))
			log.Info().Str("operation", "job.process").Msg("job.process")
		})
	}
	workers.Wait()
	lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
	if len(lines) != 20 {
		t.Fatalf("missing records: %d", len(lines))
	}
	ids := map[string]bool{}
	for _, line := range lines {
		var record map[string]any
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatal(err)
		}
		id, ok := record["request_id"].(string)
		if !ok || id == "" || ids[id] || record["correlation_id"] != id {
			t.Fatalf("crossed context: %v", record)
		}
		ids[id] = true
	}
}
