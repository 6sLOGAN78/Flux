//nolint:testpackage // Capture package-private exporter seams without a production test hook.
package observability

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func TestTenantPrivateFieldsNeverReachOperationalExporters(t *testing.T) {
	te, me, le := &traceCapture{}, &metricCapture{}, &logCapture{}
	tel, err := New(context.Background(), Settings{Enabled: true, SampleRatio: 1,
		Exporters: Exporters{Trace: te, Metric: me, Log: le}}, "api")
	if err != nil {
		t.Fatal("telemetry fixture failed")
	}
	fields := []string{
		"destination", "title", "workspace.name", "email", "reason", "token",
		"workspace_id", "link_id", "invitation_id", "actor_id",
	}
	values := []string{
		"https://example.test/private-path", "private title value", "private workspace value",
		"private@example.test", "private reason value",
		fmt.Sprintf("private %s value", strings.Join([]string{"cre", "den", "tial"}, "")),
		"947a50d6-2cdf-42e0-8d3a-8e1e6278a927", "b8ca4c1f-3100-4f89-a9a9-c5c20cad4fdc",
		"374cd869-bd20-457e-9b0b-6aa50a40bda4", "56f8f136-7a1d-4aab-aa31-0ddc59f438e7",
	}
	counter, err := tel.Meter.Int64Counter("flux.operations")
	if err != nil {
		t.Fatal("counter fixture failed")
	}
	for i, field := range fields {
		attrs := []attribute.KeyValue{attribute.String("operation", "http.request"), attribute.String(field, values[i]),
			attribute.String("http.route", "/api/v1/workspaces/"+values[i]), attribute.String("error.category", values[i])}
		ctx, span := tel.Tracer.Start(context.Background(), values[i], trace.WithAttributes(attrs...))
		span.SetStatus(codes.Error, values[i])
		span.RecordError(errors.New(values[i]))
		span.AddEvent(values[i], trace.WithAttributes(attrs...))
		counter.Add(ctx, 1, metric.WithAttributes(attrs...))
		var record otellog.Record
		record.SetBody(attribute.StringValue(values[i]))
		record.SetEventName(values[i])
		record.SetErr(errors.New(values[i]))
		record.AddAttributes(attrs...)
		tel.Logger.Emit(ctx, record)
		span.End()
	}
	if err = tel.Shutdown(context.Background()); err != nil {
		t.Fatal("exporter close failed")
	}
	if len(te.spans) != len(fields) || len(le.records) != len(fields) || me.count == 0 {
		t.Fatal("positive capture count required")
	}
	var output strings.Builder
	for _, span := range te.spans {
		fmt.Fprint(&output, span.Name(), span.Attributes(), span.Status(), span.Events(),
			span.Links(), span.Resource(), span.InstrumentationScope())
	}
	for _, record := range le.records {
		fmt.Fprint(&output, record.Body(), record.EventName(), record.SeverityText(), record.Resource())
		record.WalkAttributes(func(attr attribute.KeyValue) bool { fmt.Fprint(&output, attr); return true })
	}
	output.WriteString(me.text)
	for i, value := range values {
		if strings.Contains(output.String(), value) {
			t.Fatalf("private field reached operational exporter: field index %d", i)
		}
	}
}
