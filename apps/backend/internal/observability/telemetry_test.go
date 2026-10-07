//nolint:testpackage // These tests verify package-private lifecycle and failure-injection seams.
package observability

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type traceCapture struct {
	spans []sdktrace.ReadOnlySpan
	mu    sync.Mutex
}

func (e *traceCapture) ExportSpans(_ context.Context, s []sdktrace.ReadOnlySpan) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.spans = append(e.spans, s...)
	return nil
}
func (*traceCapture) Shutdown(context.Context) error { return nil }

type metricCapture struct {
	text  string
	count int
	mu    sync.Mutex
}

func (*metricCapture) Temporality(sdkmetric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}
func (*metricCapture) Aggregation(k sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return sdkmetric.DefaultAggregationSelector(k)
}
func (e *metricCapture) Export(_ context.Context, m *metricdata.ResourceMetrics) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.text += fmt.Sprintf("%+v", m)
	e.count++
	return nil
}
func (*metricCapture) ForceFlush(context.Context) error { return nil }
func (*metricCapture) Shutdown(context.Context) error   { return nil }

type logCapture struct {
	records []sdklog.Record
	mu      sync.Mutex
}

func (e *logCapture) Export(_ context.Context, r []sdklog.Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, v := range r {
		e.records = append(e.records, v.Clone())
	}
	return nil
}
func (*logCapture) ForceFlush(context.Context) error { return nil }
func (*logCapture) Shutdown(context.Context) error   { return nil }

func TestInjectedProvidersSanitizeAllSurfaces(t *testing.T) {
	te, me, le := &traceCapture{}, &metricCapture{}, &logCapture{}
	tel,
		err := New(context.Background(),
		Settings{Enabled: true,
			SampleRatio: 1,
			Exporters: Exporters{Trace: te,
				Metric: me,
				Log:    le}},
		"worker")
	if err != nil {
		t.Fatal(err)
	}
	ts, _ := trace.ParseTraceState("vendor=" + secret)
	parent := trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1},
		SpanID:     trace.SpanID{1},
		TraceState: ts,
		TraceFlags: trace.FlagsSampled})
	ctx := WithCorrelation(trace.ContextWithSpanContext(context.Background(), parent), validID, validID)
	ctx,
		span := tel.Tracer.Start(ctx,
		secret,
		trace.WithAttributes(attribute.String("operation",
			"job.process"),
			attribute.String("request_id",
				validID),
			attribute.String("sql.query",
				secret)),
		trace.WithLinks(trace.Link{SpanContext: parent,
			Attributes: []attribute.KeyValue{attribute.String("token",
				secret)}}))
	span.SetStatus(codes.Error, secret)
	span.AddEvent(secret, trace.WithAttributes(attribute.String("error.stack", secret)))
	span.RecordError(errors.New(secret))
	span.End()
	counter,
		err := tel.Meter.Int64Counter("flux.operations",
		metric.WithDescription(secret),
		metric.WithUnit(secret))
	if err != nil {
		t.Fatal(err)
	}
	counter.Add(ctx,
		1,
		metric.WithAttributes(attribute.String("operation",
			"job.process"),
			attribute.String("request_id",
				validID),
			attribute.String("token",
				secret)))
	histogram, err := tel.Meter.Float64Histogram("flux.duration")
	if err != nil {
		t.Fatal(err)
	}
	histogram.Record(ctx,
		12,
		metric.WithAttributes(attribute.String("operation",
			"job.process"),
			attribute.String("destination",
				secret)))
	bad, _ := tel.Meter.Int64Counter(secret)
	bad.Add(ctx, 1)
	var record otellog.Record
	record.SetBody(attribute.StringValue(secret))
	record.SetEventName(secret)
	record.SetSeverityText(secret)
	record.SetErr(errors.New(secret))
	record.AddAttributes(attribute.String("operation",
		"job.process"),
		attribute.String("request_id",
			validID),
		attribute.String("email",
			secret))
	tel.Logger.Emit(ctx, record)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err154 := tel.Shutdown(shutdownCtx); err154 != nil {
		t.Fatal(err154)
	}
	if len(te.spans) != 1 || len(le.records) != 1 || me.count == 0 {
		t.Fatalf("missing telemetry: %d %d %d", len(te.spans), len(le.records), me.count)
	}
	s := te.spans[0]
	if s.SpanContext().
		TraceState().
		Len() != 0 ||
		s.Parent().
			TraceState().
			Len() != 0 ||
		len(s.Links()) != 1 ||
		s.Links()[0].SpanContext.TraceState().
			Len() != 0 {
		t.Fatal("tracestate leaked")
	}
	got := fmt.Sprint(s.Name(),
		s.Attributes(),
		s.Events(),
		s.Links(),
		s.Status(),
		s.Resource(),
		s.InstrumentationScope(),
		me.text)
	var gotSb122 strings.Builder
	for _, r := range le.records {
		fmt.Fprint(&gotSb122, r.Body(),
			r.EventName(),
			r.SeverityText(),
			r.Resource(),
			r.InstrumentationScope())
		r.WalkAttributes(func(a attribute.KeyValue) bool { got += fmt.Sprint(a); return true })
		if r.TraceID() != s.SpanContext().TraceID() || r.SpanID() != s.SpanContext().SpanID() {
			t.Fatal("log correlation lost")
		}
	}
	got += gotSb122.String()
	if strings.Contains(got, secret) {
		t.Fatal("secret reached exporters:", got)
	}
	if !strings.Contains(got,
		"flux.worker") ||
		!strings.Contains(got,
			validID) ||
		strings.Contains(me.text,
			validID) {
		t.Fatal("resource or cardinality boundary incorrect:", got)
	}
}

func TestShutdownAttemptsAllAndBoundsCaller(t *testing.T) {
	cause := errors.New(secret)
	var attempted atomic.Int32
	release := make(chan struct{})
	tel := &Telemetry{closers: []namedCloser{{name: "trace",
		close: func(context.Context) error { attempted.Add(1); return cause }},
		{name: "metric",
			close: func(context.Context) error { attempted.Add(1); return cause }},
		{name: "log",
			close: func(ctx context.Context) error { attempted.Add(1); <-ctx.Done(); <-release; return ctx.Err() }}}}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := tel.Shutdown(ctx)
	close(release)
	if time.Since(start) > 300*time.Millisecond ||
		attempted.Load() != 3 ||
		!errors.Is(err,
			context.DeadlineExceeded) ||
		!errors.Is(err,
			cause) ||
		strings.Contains(err.Error(),
			secret) {
		t.Fatal("unbounded or unsafe shutdown", attempted.Load(), err)
	}
	ctx, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	err = tel.Shutdown(ctx)
	if !errors.Is(err, cause) || strings.Contains(err.Error(), secret) {
		t.Fatal("causes not preserved safely", err)
	}
	if attempted.Load() != 3 {
		t.Fatal("shutdown repeated")
	}
}

type failedTrace struct {
	err error
	traceCapture
	closed atomic.Int32
}

func (e *failedTrace) Shutdown(context.Context) error { e.closed.Add(1); return e.err }

type failedMetric struct {
	err error
	metricCapture
	closed atomic.Int32
}

func (e *failedMetric) Shutdown(context.Context) error { e.closed.Add(1); return e.err }

type failedLog struct {
	err error
	logCapture
	closed atomic.Int32
}

func (e *failedLog) Shutdown(context.Context) error { e.closed.Add(1); return e.err }

func TestActualProviderShutdownAggregatesSafeFailures(t *testing.T) {
	a, b, c := errors.New(secret+"trace"), errors.New(secret+"metric"), errors.New(secret+"log")
	te, me, le := &failedTrace{err: a}, &failedMetric{err: b}, &failedLog{err: c}
	tel,
		err := New(context.Background(),
		Settings{Enabled: true,
			Exporters: Exporters{Trace: te,
				Metric: me,
				Log:    le}},
		"api")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = tel.Shutdown(ctx)
	if !errors.Is(err,
		a) ||
		!errors.Is(err,
			b) ||
		!errors.Is(err,
			c) ||
		strings.Contains(err.Error(),
			secret) ||
		te.closed.Load() != 1 ||
		me.closed.Load() != 1 ||
		le.closed.Load() != 1 {
		t.Fatal("provider shutdown lost errors or ownership", err)
	}
}

func TestOptionalProvidersAndSettingsBounds(t *testing.T) {
	for _, s := range []Settings{{QueueSize: -1},
		{QueueSize: 4097},
		{BatchSize: 5000},
		{SampleRatio: 2},
		{ExportTimeout: time.Minute},
		{ExportInterval: time.Hour},
		{Environment: secret},
		{Endpoint: "https://user:password@localhost"},
		{Endpoint: "http://localhost?token=" + secret}} {
		if _, err := New(context.Background(), s, "api"); err == nil || strings.Contains(err.Error(), secret) {
			t.Fatal("unsafe settings accepted", err)
		}
	}
	if _,
		err := New(context.Background(),
		Settings{},
		secret); err == nil ||
		strings.Contains(err.Error(),
			secret) {
		t.Fatal("unsafe role")
	}
	tel, err := New(context.Background(), Settings{}, "redirector")
	if err != nil {
		t.Fatal(err)
	}
	_, span := tel.Tracer.Start(context.Background(), "http.request")
	span.End()
	counter, _ := tel.Meter.Int64Counter("flux.operations")
	counter.Add(context.Background(), 1)
	if err327 := tel.Shutdown(context.Background()); err327 != nil {
		t.Fatal(err327)
	}
}

type blockedTrace struct {
	entered chan struct{}
	release chan struct{}
	traceCapture
	once sync.Once
}

func (e *blockedTrace) ExportSpans(ctx context.Context, s []sdktrace.ReadOnlySpan) error {
	e.once.Do(func() { close(e.entered) })
	select {
	case <-e.release:
		return e.traceCapture.ExportSpans(ctx, s)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestBatchQueueIsBoundedAndDoesNotBlockInstrumentation(t *testing.T) {
	e := &blockedTrace{entered: make(chan struct{}), release: make(chan struct{})}
	tel,
		err := New(context.Background(),
		Settings{Enabled: true,
			SampleRatio:    1,
			QueueSize:      4,
			BatchSize:      2,
			ExportInterval: time.Millisecond,
			Exporters:      Exporters{Trace: e}},
		"api")
	if err != nil {
		t.Fatal(err)
	}
	_, span := tel.Tracer.Start(context.Background(), "http.request")
	span.End()
	select {
	case <-e.entered:
	case <-time.After(time.Second):
		t.Fatal("export never started")
	}
	start := time.Now()
	for range 1000 {
		_, span = tel.Tracer.Start(context.Background(), "http.request")
		span.End()
	}
	if time.Since(start) > time.Second {
		t.Fatal("instrumentation waited for exporter")
	}
	close(e.release)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err381 := tel.Shutdown(ctx); err381 != nil {
		t.Fatal(err381)
	}
	if len(e.spans) > 8 {
		t.Fatal("queue was not bounded", len(e.spans))
	}
}

func TestRootSamplingIsExplicit(t *testing.T) {
	e := &traceCapture{}
	tel,
		err := New(context.Background(),
		Settings{Enabled: true,
			SampleRatio: 0,
			Exporters:   Exporters{Trace: e}},
		"migrator")
	if err != nil {
		t.Fatal(err)
	}
	_, span := tel.Tracer.Start(context.Background(), "database.query")
	span.End()
	if err402 := tel.Shutdown(context.Background()); err402 != nil {
		t.Fatal(err402)
	}
	if len(e.spans) != 0 {
		t.Fatal("zero sampling ignored")
	}
}

func TestShutdownHasInternalBudgetWithoutCallerDeadline(t *testing.T) {
	release := make(chan struct{})
	tel := &Telemetry{shutdownTimeout: 10 * time.Millisecond,
		closers: []namedCloser{{name: "trace",
			close: func(context.Context) error { <-release; return nil }}}}
	start := time.Now()
	err := tel.Shutdown(context.Background())
	close(release)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > time.Second {
		t.Fatal("internal shutdown budget ignored", err)
	}
	<-tel.done
}

func TestExporterRemovesMetricExemplarsAndResources(t *testing.T) {
	capture := &metricCapture{}
	e := metricExporter{Exporter: capture, timeout: time.Second}
	attrs := attribute.NewSet(attribute.String("token", secret), attribute.String("operation", "database.query"))
	sample := metricdata.Exemplar[int64]{Value: 1,
		FilteredAttributes: []attribute.KeyValue{attribute.String("token",
			secret)}}
	data := metricdata.ResourceMetrics{Resource: resource.NewSchemaless(attribute.String("secret",
		secret)),
		ScopeMetrics: []metricdata.ScopeMetrics{{Scope: instrumentation.Scope{Name: secret,
			Version:    secret,
			Attributes: attrs},
			Metrics: []metricdata.Metrics{{Name: "flux.operations",
				Description: secret,
				Unit:        secret,
				Data: metricdata.Sum[int64]{DataPoints: []metricdata.DataPoint[int64]{{Value: 1,
					Attributes: attrs,
					Exemplars:  []metricdata.Exemplar[int64]{sample}}}}}}}}}
	if err := e.Export(context.Background(), &data); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(capture.text, secret) || !strings.Contains(capture.text, "database.query") {
		t.Fatal("metric surface leaked", capture.text)
	}
	if data.ScopeMetrics[0].Metrics[0].Description != secret {
		t.Fatal("sanitizer mutated input")
	}
}
