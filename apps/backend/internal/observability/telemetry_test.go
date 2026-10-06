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
 sdklog "go.opentelemetry.io/otel/sdk/log"
 sdkmetric "go.opentelemetry.io/otel/sdk/metric"
 "go.opentelemetry.io/otel/sdk/metric/metricdata"
 sdktrace "go.opentelemetry.io/otel/sdk/trace"
 "go.opentelemetry.io/otel/trace"
)

type traceCapture struct {mu sync.Mutex; spans []sdktrace.ReadOnlySpan}
func (e *traceCapture) ExportSpans(_ context.Context,s []sdktrace.ReadOnlySpan)error {e.mu.Lock(); defer e.mu.Unlock(); e.spans=append(e.spans,s...);return nil}
func (*traceCapture) Shutdown(context.Context)error{return nil}
type metricCapture struct {mu sync.Mutex; text string; count int}
func (*metricCapture) Temporality(sdkmetric.InstrumentKind)metricdata.Temporality{return metricdata.CumulativeTemporality}
func (*metricCapture) Aggregation(k sdkmetric.InstrumentKind)sdkmetric.Aggregation{return sdkmetric.DefaultAggregationSelector(k)}
func (e *metricCapture) Export(_ context.Context,m *metricdata.ResourceMetrics)error {e.mu.Lock();defer e.mu.Unlock();e.text+=fmt.Sprintf("%+v",m);e.count++;return nil}
func (*metricCapture) ForceFlush(context.Context)error{return nil}
func (*metricCapture) Shutdown(context.Context)error{return nil}
type logCapture struct{mu sync.Mutex; records []sdklog.Record}
func (e *logCapture) Export(_ context.Context,r []sdklog.Record)error {e.mu.Lock();defer e.mu.Unlock();for _,v:=range r {e.records=append(e.records,v.Clone())};return nil}
func (*logCapture) ForceFlush(context.Context)error{return nil}
func (*logCapture) Shutdown(context.Context)error{return nil}

func TestInjectedProvidersSanitizeAllSurfaces(t *testing.T) {
 te,me,le:=&traceCapture{},&metricCapture{},&logCapture{}
 tel,err:=New(context.Background(),Settings{Enabled:true,SampleRatio:1,Exporters:Exporters{Trace:te,Metric:me,Log:le}},"worker");if err!=nil {t.Fatal(err)}
 ts,_:=trace.ParseTraceState("vendor="+secret)
 parent:=trace.NewSpanContext(trace.SpanContextConfig{TraceID:trace.TraceID{1},SpanID:trace.SpanID{1},TraceState:ts,TraceFlags:trace.FlagsSampled})
 ctx:=WithCorrelation(trace.ContextWithSpanContext(context.Background(),parent),validID,validID)
 ctx,span:=tel.Tracer.Start(ctx,secret,trace.WithAttributes(attribute.String("operation","job.process"),attribute.String("request_id",validID),attribute.String("sql.query",secret)),trace.WithLinks(trace.Link{SpanContext:parent,Attributes:[]attribute.KeyValue{attribute.String("token",secret)}}))
 span.SetStatus(codes.Error,secret);span.AddEvent(secret,trace.WithAttributes(attribute.String("error.stack",secret)));span.RecordError(errors.New(secret));span.End()
 counter,err:=tel.Meter.Int64Counter("flux.operations",metric.WithDescription(secret),metric.WithUnit(secret));if err!=nil {t.Fatal(err)}
 counter.Add(ctx,1,metric.WithAttributes(attribute.String("operation","job.process"),attribute.String("request_id",validID),attribute.String("token",secret)))
 bad,_:=tel.Meter.Int64Counter(secret);bad.Add(ctx,1)
 var record otellog.Record;record.SetBody(attribute.StringValue(secret));record.SetEventName(secret);record.SetSeverityText(secret);record.SetErr(errors.New(secret));record.AddAttributes(attribute.String("operation","job.process"),attribute.String("request_id",validID),attribute.String("email",secret));tel.Logger.Emit(ctx,record)
 shutdownCtx,cancel:=context.WithTimeout(context.Background(),time.Second);defer cancel();if err:=tel.Shutdown(shutdownCtx);err!=nil {t.Fatal(err)}
 if len(te.spans)!=1 || len(le.records)!=1 || me.count==0 {t.Fatalf("missing telemetry: %d %d %d",len(te.spans),len(le.records),me.count)}
 s:=te.spans[0]
 if s.SpanContext().TraceState().Len()!=0 || s.Parent().TraceState().Len()!=0 || len(s.Links())!=1 || s.Links()[0].SpanContext.TraceState().Len()!=0 {t.Fatal("tracestate leaked")}
 got:=fmt.Sprint(s.Name(),s.Attributes(),s.Events(),s.Links(),s.Status(),s.Resource(),s.InstrumentationScope(),me.text)
 for _,r:=range le.records {got+=fmt.Sprint(r.Body(),r.EventName(),r.SeverityText(),r.Resource(),r.InstrumentationScope());r.WalkAttributes(func(a attribute.KeyValue)bool {got+=fmt.Sprint(a);return true});if r.TraceID()!=s.SpanContext().TraceID() || r.SpanID()!=s.SpanContext().SpanID() {t.Fatal("log correlation lost")}}
 if strings.Contains(got,secret) {t.Fatal("secret reached exporters:",got)}
 if !strings.Contains(got,"flux.worker") || !strings.Contains(got,validID) || strings.Contains(me.text,validID) {t.Fatal("resource or cardinality boundary incorrect:",got)}
}

func TestShutdownAttemptsAllAndBoundsCaller(t *testing.T) {
 cause:=errors.New(secret);var attempted atomic.Int32;release:=make(chan struct{})
 tel:=&Telemetry{closers:[]namedCloser{{"trace",func(context.Context)error{attempted.Add(1);return cause}},{"metric",func(context.Context)error{attempted.Add(1);return cause}},{"log",func(ctx context.Context)error{attempted.Add(1);<-ctx.Done();<-release;return ctx.Err()}}}}
 ctx,cancel:=context.WithTimeout(context.Background(),30*time.Millisecond);defer cancel();start:=time.Now();err:=tel.Shutdown(ctx)
 close(release)
 if time.Since(start)>300*time.Millisecond || attempted.Load()!=3 || !errors.Is(err,context.DeadlineExceeded) || strings.Contains(err.Error(),secret) {t.Fatal("unbounded or unsafe shutdown",attempted.Load(),err)}
 ctx2,cancel2:=context.WithTimeout(context.Background(),time.Second);defer cancel2();err=tel.Shutdown(ctx2)
 if !errors.Is(err,cause) || strings.Contains(err.Error(),secret) {t.Fatal("causes not preserved safely",err)}
 if attempted.Load()!=3 {t.Fatal("shutdown repeated")}
}
