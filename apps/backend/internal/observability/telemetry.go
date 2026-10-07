// Package observability owns vendor-neutral, private-by-default telemetry.
package observability

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otellog "go.opentelemetry.io/otel/log"
	lognoop "go.opentelemetry.io/otel/log/noop"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

const (
	defaultExportTimeout    = 2 * time.Second
	defaultExportInterval   = 5 * time.Second
	metricCardinalityLimit  = 128
	logAttributeLimit       = 16
	logAttributeValueLength = 128
)

const scopeName = "github.com/6sLOGAN78/flux"

// Exporters inject capture or transport sinks. Ownership transfers to New.
type Exporters struct {
	Trace  sdktrace.SpanExporter
	Metric sdkmetric.Exporter
	Log    sdklog.Exporter
}

// Settings is independent of application configuration and SDK globals.
// Endpoint is a base HTTP(S) OTLP collector URL; empty disables remote export.
type Settings struct {
	Exporters      Exporters
	Endpoint       string
	Environment    string
	SampleRatio    float64
	QueueSize      int
	BatchSize      int
	ExportTimeout  time.Duration
	ExportInterval time.Duration
	Enabled        bool
}
type namedCloser struct {
	close func(context.Context) error
	name  string
}

// Telemetry owns providers; instrumentation receives fixed-scope API objects.
type Telemetry struct {
	Tracer          trace.Tracer
	Meter           metric.Meter
	Logger          otellog.Logger
	Propagator      propagation.TextMapPropagator
	shutdownErr     error
	shutdownContext context.Context
	done            chan struct{}
	closers         []namedCloser
	shutdownErrors  []error
	shutdownTimeout time.Duration
	once            sync.Once
	shutdownMu      sync.Mutex
}

func normalize(s Settings) (Settings, error) {
	if s.QueueSize == 0 {
		s.QueueSize = 256
	}
	if s.BatchSize == 0 {
		s.BatchSize = 64
	}
	if s.ExportTimeout == 0 {
		s.ExportTimeout = defaultExportTimeout
	}
	if s.ExportInterval == 0 {
		s.ExportInterval = defaultExportInterval
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
		return s, errors.New("invalid telemetry bounds")
	}
	if s.Environment == "" {
		s.Environment = "local"
	}
	if !oneOf(s.Environment, "local", "development", "test", "staging", "production") {
		return s, errors.New("invalid telemetry environment")
	}
	if s.Endpoint != "" {
		u, err := url.Parse(s.Endpoint)
		if err != nil ||
			!oneOf(u.Scheme,
				"http",
				"https") ||
			u.Host == "" ||
			u.User != nil ||
			u.RawQuery != "" ||
			u.Fragment != "" {
			return s, errors.New("invalid telemetry endpoint")
		}
	}
	return s, nil
}

// New creates asynchronous bounded providers without installing SDK globals.
func New(ctx context.Context, settings Settings, role string) (*Telemetry, error) {
	s, err := normalize(settings)
	if err != nil {
		return nil, err
	}
	if !oneOf(role, "api", "redirector", "worker", "migrator") {
		return nil, errors.New("invalid telemetry role")
	}
	t := &Telemetry{Tracer: tracenoop.NewTracerProvider().
		Tracer(scopeName),
		Meter: metricnoop.NewMeterProvider().
			Meter(scopeName),
		Logger: lognoop.NewLoggerProvider().
			Logger(scopeName),
		Propagator:      TraceparentPropagator{},
		shutdownTimeout: s.ExportTimeout}
	if !s.Enabled {
		return t, nil
	}
	exp := s.Exporters
	if s.Endpoint != "" {
		if err = createHTTPExporters(ctx, s, &exp); err != nil {
			return nil, err
		}
	}
	res := resource.NewSchemaless(attribute.String("service.name",
		"flux."+role),
		attribute.String("process.role",
			role),
		attribute.String("deployment.environment.name",
			s.Environment))
	if exp.Trace != nil {
		limits := sdktrace.NewSpanLimits()
		limits.AttributeCountLimit = 16
		limits.AttributeValueLengthLimit = 128
		limits.EventCountLimit = 16
		limits.LinkCountLimit = 16
		limits.AttributePerEventCountLimit = 16
		limits.AttributePerLinkCountLimit = 16
		p := sdktrace.NewTracerProvider(sdktrace.WithResource(res),
			sdktrace.WithRawSpanLimits(limits),
			sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(s.SampleRatio))),
			sdktrace.WithBatcher(traceExporter{SpanExporter: exp.Trace,
				res:     res,
				timeout: s.ExportTimeout},
				sdktrace.WithMaxQueueSize(s.QueueSize),
				sdktrace.WithMaxExportBatchSize(s.BatchSize),
				sdktrace.WithBatchTimeout(s.ExportInterval),
				sdktrace.WithExportTimeout(s.ExportTimeout)))
		t.Tracer = p.Tracer(scopeName)
		t.closers = append(t.closers, namedCloser{name: "trace", close: p.Shutdown})
	}
	if exp.Metric != nil {
		reader := sdkmetric.NewPeriodicReader(metricExporter{Exporter: exp.Metric,
			res:     res,
			timeout: s.ExportTimeout},
			sdkmetric.WithInterval(s.ExportInterval),
			sdkmetric.WithTimeout(s.ExportTimeout),
			sdkmetric.WithMaxExportBatchSize(s.BatchSize))
		p := sdkmetric.NewMeterProvider(sdkmetric.WithResource(res),
			sdkmetric.WithReader(reader),
			sdkmetric.WithCardinalityLimit(metricCardinalityLimit),
			sdkmetric.WithExemplarFilter(exemplar.AlwaysOffFilter),
			sdkmetric.WithView(sdkmetric.NewView(sdkmetric.Instrument{Name: "*"},
				sdkmetric.Stream{AttributeFilter: func(a attribute.KeyValue) bool {
					return safeAttribute(a,
						true)
				}})))
		t.Meter = p.Meter(scopeName)
		t.closers = append(t.closers, namedCloser{name: "metric", close: p.Shutdown})
	}
	if exp.Log != nil {
		batch := sdklog.NewBatchProcessor(logExporter{Exporter: exp.Log,
			timeout: s.ExportTimeout},
			sdklog.WithMaxQueueSize(s.QueueSize),
			sdklog.WithExportMaxBatchSize(s.BatchSize),
			sdklog.WithExportInterval(s.ExportInterval),
			sdklog.WithExportTimeout(s.ExportTimeout))
		p := sdklog.NewLoggerProvider(sdklog.WithResource(res),
			sdklog.WithAttributeCountLimit(logAttributeLimit),
			sdklog.WithAttributeValueLengthLimit(logAttributeValueLength),
			sdklog.WithProcessor(batch))
		t.Logger = p.Logger(scopeName)
		t.closers = append(t.closers, namedCloser{name: "log", close: p.Shutdown})
	}
	return t, nil
}

func createHTTPExporters(ctx context.Context, s Settings, e *Exporters) error {
	base := strings.TrimRight(s.Endpoint, "/")
	owned := []namedCloser{}
	fail := func(err error) error {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), s.ExportTimeout)
		defer cancel()
		t := &Telemetry{closers: owned}
		return errors.Join(safeErrorFor("initialize", err), t.Shutdown(cleanupCtx))
	}
	if e.Trace == nil {
		v,
			err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpointURL(base+"/v1/traces"),
			otlptracehttp.WithTimeout(s.ExportTimeout),
			otlptracehttp.WithRetry(otlptracehttp.RetryConfig{Enabled: false}))
		if err != nil {
			return fail(err)
		}
		e.Trace = v
		owned = append(owned, namedCloser{name: "trace", close: v.Shutdown})
	}
	if e.Metric == nil {
		v,
			err := otlpmetrichttp.New(ctx,
			otlpmetrichttp.WithEndpointURL(base+"/v1/metrics"),
			otlpmetrichttp.WithTimeout(s.ExportTimeout),
			otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{Enabled: false}))
		if err != nil {
			return fail(err)
		}
		e.Metric = v
		owned = append(owned, namedCloser{name: "metric", close: v.Shutdown})
	}
	if e.Log == nil {
		v,
			err := otlploghttp.New(ctx,
			otlploghttp.WithEndpointURL(base+"/v1/logs"),
			otlploghttp.WithTimeout(s.ExportTimeout),
			otlploghttp.WithRetry(otlploghttp.RetryConfig{Enabled: false}))
		if err != nil {
			return fail(err)
		}
		e.Log = v
	}
	return nil
}

// Shutdown attempts every independent provider even if another blocks or fails.
// SDK shutdown is idempotent; repeated callers observe the same final result.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	t.once.Do(func() {
		timeout := t.shutdownTimeout
		if timeout == 0 {
			timeout = defaultExportTimeout
		}
		shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
		t.shutdownContext = shutdownCtx
		t.done = make(chan struct{})
		results := make(chan error, len(t.closers))
		for _, closer := range t.closers {
			go func(c namedCloser) {
				results <- safeErrorFor(c.name+" shutdown",
					c.close(shutdownCtx))
			}(closer)
		}
		go func() {
			defer cancel()
			var errs []error
			for range t.closers {
				if err := <-results; err != nil {
					errs = append(errs, err)
					t.shutdownMu.Lock()
					t.shutdownErrors = append(t.shutdownErrors, err)
					t.shutdownMu.Unlock()
				}
			}
			t.shutdownErr = errors.Join(errs...)
			close(t.done)
		}()
	})
	select {
	case <-t.done:
		return t.shutdownErr
	case <-ctx.Done():
		t.shutdownMu.Lock()
		errs := append([]error(nil), t.shutdownErrors...)
		t.shutdownMu.Unlock()
		return errors.Join(append(errs, safeErrorFor("shutdown", ctx.Err()))...)
	case <-t.shutdownContext.Done():
		select {
		case <-t.done:
			return t.shutdownErr
		default:
		}
		t.shutdownMu.Lock()
		errs := append([]error(nil), t.shutdownErrors...)
		t.shutdownMu.Unlock()
		return errors.Join(append(errs, safeErrorFor("shutdown", t.shutdownContext.Err()))...)
	}
}

type safeSpan struct {
	sdktrace.ReadOnlySpan
	res *resource.Resource
}

func (s safeSpan) Name() string { return SafeOperation(s.ReadOnlySpan.Name()) }
func (s safeSpan) SpanContext() trace.SpanContext {
	return CleanSpanContext(s.ReadOnlySpan.SpanContext())
}
func (s safeSpan) Parent() trace.SpanContext { return CleanSpanContext(s.ReadOnlySpan.Parent()) }
func (s safeSpan) Attributes() []attribute.KeyValue {
	return SanitizeAttributes(s.ReadOnlySpan.Attributes(), false)
}
func (s safeSpan) Resource() *resource.Resource { return s.res }
func (s safeSpan) InstrumentationScope() instrumentation.Scope {
	return instrumentation.Scope{Name: scopeName}
}
func (s safeSpan) InstrumentationLibrary() instrumentation.Scope { return s.InstrumentationScope() }
func (s safeSpan) Status() sdktrace.Status {
	status := s.ReadOnlySpan.Status()
	status.Description = ""
	return status
}
func (s safeSpan) Links() []sdktrace.Link {
	links := append([]sdktrace.Link(nil), s.ReadOnlySpan.Links()...)
	for i := range links {
		links[i].SpanContext = CleanSpanContext(links[i].SpanContext)
		links[i].Attributes = SanitizeAttributes(links[i].Attributes, false)
	}
	return links
}
func (s safeSpan) Events() []sdktrace.Event {
	events := append([]sdktrace.Event(nil), s.ReadOnlySpan.Events()...)
	for i := range events {
		events[i].Name = SafeOperation(events[i].Name)
		events[i].Attributes = SanitizeAttributes(events[i].Attributes, false)
	}
	return events
}

type traceExporter struct {
	sdktrace.SpanExporter
	res     *resource.Resource
	timeout time.Duration
}

func (e traceExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	safe := make([]sdktrace.ReadOnlySpan, len(spans))
	for i, s := range spans {
		safe[i] = safeSpan{s, e.res}
	}
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()
	return safeErrorFor("export", e.SpanExporter.ExportSpans(ctx, safe))
}
func (e traceExporter) Shutdown(ctx context.Context) error {
	return safeErrorFor("shutdown", e.SpanExporter.Shutdown(ctx))
}

type logExporter struct {
	sdklog.Exporter
	timeout time.Duration
}

func (e logExporter) Export(ctx context.Context, records []sdklog.Record) error {
	safe := make([]sdklog.Record, len(records))
	for i := range records {
		r := records[i].Clone()
		var attrs []attribute.KeyValue
		r.WalkAttributes(func(a attribute.KeyValue) bool { attrs = append(attrs, a); return true })
		r.SetAttributes(SanitizeAttributes(attrs, false)...)
		r.SetBody(attribute.StringValue(SafeOperation(r.Body().AsString())))
		r.SetEventName(SafeOperation(r.EventName()))
		r.SetSeverityText("")
		safe[i] = r
	}
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()
	return safeErrorFor("export", e.Exporter.Export(ctx, safe))
}
func (e logExporter) Shutdown(ctx context.Context) error {
	return safeErrorFor("shutdown", e.Exporter.Shutdown(ctx))
}
func (e logExporter) ForceFlush(ctx context.Context) error {
	return safeErrorFor("flush", e.Exporter.ForceFlush(ctx))
}

type metricExporter struct {
	sdkmetric.Exporter
	res     *resource.Resource
	timeout time.Duration
}

func (e metricExporter) Export(ctx context.Context, data *metricdata.ResourceMetrics) error {
	safe := metricdata.ResourceMetrics{Resource: e.res}
	for _, scope := range data.ScopeMetrics {
		out := metricdata.ScopeMetrics{Scope: instrumentation.Scope{Name: scopeName}}
		for _, m := range scope.Metrics {
			if !oneOf(m.Name,
				"flux.operations",
				"flux.duration",
				"flux.http.requests",
				"flux.http.duration",
				"flux.db.operations",
				"flux.db.duration",
				"flux.jobs",
				"flux.job.duration") {
				continue
			}
			m.Description = ""
			m.Unit = ""
			m.Data = sanitizeMetric(m.Data)
			if m.Data != nil {
				out.Metrics = append(out.Metrics, m)
			}
		}
		if len(out.Metrics) > 0 {
			safe.ScopeMetrics = append(safe.ScopeMetrics, out)
		}
	}
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()
	return safeErrorFor("export", e.Exporter.Export(ctx, &safe))
}
func (e metricExporter) Shutdown(ctx context.Context) error {
	return safeErrorFor("shutdown", e.Exporter.Shutdown(ctx))
}
func (e metricExporter) ForceFlush(ctx context.Context) error {
	return safeErrorFor("flush", e.Exporter.ForceFlush(ctx))
}

func sanitizePoints[N int64 | float64](points []metricdata.DataPoint[N]) []metricdata.DataPoint[N] {
	out := append([]metricdata.DataPoint[N](nil), points...)
	for i := range out {
		out[i].Attributes = attribute.NewSet(SanitizeAttributes(out[i].Attributes.ToSlice(), true)...)
		out[i].Exemplars = nil
	}
	return out
}
func sanitizeHistogram[N int64 | float64](
	points []metricdata.HistogramDataPoint[N],
) []metricdata.HistogramDataPoint[N] {
	out := append([]metricdata.HistogramDataPoint[N](nil), points...)
	for i := range out {
		out[i].Attributes = attribute.NewSet(SanitizeAttributes(out[i].Attributes.ToSlice(), true)...)
		out[i].Exemplars = nil
	}
	return out
}
func sanitizeExponential[N int64 | float64](
	points []metricdata.ExponentialHistogramDataPoint[N],
) []metricdata.ExponentialHistogramDataPoint[N] {
	out := append([]metricdata.ExponentialHistogramDataPoint[N](nil), points...)
	for i := range out {
		out[i].Attributes = attribute.NewSet(SanitizeAttributes(out[i].Attributes.ToSlice(), true)...)
		out[i].Exemplars = nil
	}
	return out
}
func sanitizeMetric(data metricdata.Aggregation) metricdata.Aggregation {
	switch d := data.(type) {
	case metricdata.Sum[int64]:
		d.DataPoints = sanitizePoints(d.DataPoints)
		return d
	case metricdata.Sum[float64]:
		d.DataPoints = sanitizePoints(d.DataPoints)
		return d
	case metricdata.Gauge[int64]:
		d.DataPoints = sanitizePoints(d.DataPoints)
		return d
	case metricdata.Gauge[float64]:
		d.DataPoints = sanitizePoints(d.DataPoints)
		return d
	case metricdata.Histogram[int64]:
		d.DataPoints = sanitizeHistogram(d.DataPoints)
		return d
	case metricdata.Histogram[float64]:
		d.DataPoints = sanitizeHistogram(d.DataPoints)
		return d
	case metricdata.ExponentialHistogram[int64]:
		d.DataPoints = sanitizeExponential(d.DataPoints)
		return d
	case metricdata.ExponentialHistogram[float64]:
		d.DataPoints = sanitizeExponential(d.DataPoints)
		return d
	default:
		return nil
	}
}
