package job

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/lib/email"
	loggerpkg "github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/hibiken/asynq"
	containerconfig "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

const (
	requestID     = "ac1c5930-0ce6-4851-97c9-565d8192a87b"
	correlationID = "bef0ae77-1fb7-4f52-9777-34cc01939ea8"
	parent        = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	privateMarker = "PRIVATE_QUEUE_METADATA"
)

type logCapture struct {
	mu      sync.Mutex
	records []sdklog.Record
}

func (c *logCapture) Export(_ context.Context, records []sdklog.Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, r := range records {
		c.records = append(c.records, r.Clone())
	}
	return nil
}
func (*logCapture) Shutdown(context.Context) error   { return nil }
func (*logCapture) ForceFlush(context.Context) error { return nil }

func TestCorrelationRetryLegacyRedis(t *testing.T) {
	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{ContainerRequest: testcontainers.ContainerRequest{
		Image: "redis:8.10.2-alpine@sha256:3811787313eba226a2ef38658c6ccb91cd5e110edc89c37767de373120a0e5a0", ExposedPorts: []string{"6379/tcp"},
		HostConfigModifier: func(cfg *containerconfig.HostConfig) {
			cfg.PortBindings = network.PortMap{network.MustParsePort("6379/tcp"): {{HostIP: netip.MustParseAddr("127.0.0.1"), HostPort: "0"}}}
		},
		WaitingFor: wait.ForListeningPort("6379/tcp"),
	}, Started: true})
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)
	host, err := container.Host(context.Background())
	require.NoError(t, err)
	port, err := container.MappedPort(context.Background(), "6379/tcp")
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{Addr: net.JoinHostPort(host, port.Port())})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	for _, mode := range []string{"context", "legacy", "absent"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			spans, logs := tracetest.NewInMemoryExporter(), &logCapture{}
			tel, err := observability.New(ctx, observability.Settings{Enabled: true, SampleRatio: 1, ExportInterval: 10 * time.Millisecond, Exporters: observability.Exporters{Trace: spans, Log: logs}}, "worker")
			require.NoError(t, err)
			var output bytes.Buffer
			log := loggerpkg.NewLogger(config.DefaultObservabilityConfig(), &output, tel.Logger)
			var attempts atomic.Int32
			deliveries := make(chan bool, 4)
			transport := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var payload struct {
					To   []string `json:"to"`
					HTML string   `json:"html"`
				}
				decodeErr := json.NewDecoder(r.Body).Decode(&payload)
				deliveries <- decodeErr == nil && len(payload.To) == 1 && payload.To[0] == "private-recipient@example.com" && strings.Contains(payload.HTML, "PrivateFirstName")
				w.Header().Set("Content-Type", "application/json")
				if attempts.Add(1) == 1 {
					w.WriteHeader(503)
					_, _ = io.WriteString(w, `{"message":"PRIVATE_PROVIDER_ERROR","name":"provider_failure"}`)
					return
				}
				_, _ = io.WriteString(w, `{"id":"local-email"}`)
			}))
			defer transport.Close()
			t.Setenv("RESEND_BASE_URL", transport.URL+"/")
			cfg := &config.Config{Worker: config.RoleConfig{DrainTimeout: time.Second}, Integration: config.IntegrationConfig{ResendAPIKey: "local-key"}}
			consumerConfig := asynq.Config{Concurrency: 1, Queues: map[string]int{"default": 1}, ShutdownTimeout: time.Second, TaskCheckInterval: 10 * time.Millisecond, DelayedTaskCheckInterval: 100 * time.Millisecond, RetryDelayFunc: func(int, error, *asynq.Task) time.Duration { return 3 * time.Second }}
			consumer := newConsumer(&log, client, email.NewClient(cfg, &log), tel, consumerConfig)
			t.Cleanup(func() { require.NoError(t, consumer.Stop()); require.NoError(t, tel.Shutdown(context.Background())) })
			producer := NewProducer(&log, client, tel)
			var task *asynq.Task
			switch mode {
			case "context":
				parentCtx := observability.TraceparentPropagator{}.Extract(ctx, propagation.MapCarrier{"traceparent": parent})
				state, _ := trace.ParseTraceState("private=" + privateMarker)
				sc := trace.SpanContextFromContext(parentCtx).WithTraceState(state)
				member, _ := baggage.NewMember("private", privateMarker)
				bag, _ := baggage.New(member)
				parentCtx = baggage.ContextWithBaggage(trace.ContextWithSpanContext(parentCtx, sc), bag)
				parentCtx = observability.WithCorrelation(parentCtx, requestID, correlationID)
				task, err = NewWelcomeEmailTaskContext(parentCtx, "private-recipient@example.com", "PrivateFirstName")
			case "legacy":
				payload, e := json.Marshal(map[string]any{"to": "private-recipient@example.com", "first_name": "PrivateFirstName", "metadata": map[string]any{"version": 1, "request_id": requestID, "correlation_id": correlationID, "traceparent": parent, "tracestate": "private=" + privateMarker, "baggage": "private=" + privateMarker, "unknown": privateMarker}})
				require.NoError(t, e)
				task = asynq.NewTaskWithHeaders(TaskWelcome, payload, map[string]string{"tracestate": "private=" + privateMarker, "baggage": privateMarker}, asynq.MaxRetry(1))
			case "absent":
				task, err = NewWelcomeEmailTask("private-recipient@example.com", "PrivateFirstName")
			}
			require.NoError(t, err)
			info, err := producer.Client.EnqueueContext(ctx, task)
			require.NoError(t, err)
			inspector := asynq.NewInspectorFromRedisClient(client)
			queued, err := inspector.GetTaskInfo("default", info.ID)
			require.NoError(t, err)
			if mode == "context" {
				require.NotContains(t, string(queued.Payload), privateMarker)
				require.Contains(t, string(queued.Payload), requestID)
			}
			require.NoError(t, consumer.Start())
			select {
			case valid := <-deliveries:
				require.True(t, valid)
			case <-ctx.Done():
				t.Fatal("first delivery deadline")
			}
			var retried *asynq.TaskInfo
			retryID := info.ID
			if mode == "legacy" {
				retryID = "safe:" + info.ID
			}
			require.Eventually(t, func() bool {
				retried, err = inspector.GetTaskInfo("default", retryID)
				return err == nil && retried.State == asynq.TaskStateRetry
			}, 2*time.Second, 10*time.Millisecond)
			if mode == "legacy" {
				_, err = inspector.GetTaskInfo("default", info.ID)
				require.ErrorIs(t, err, asynq.ErrTaskNotFound)
				require.Equal(t, 1, retried.MaxRetry)
			}
			require.NotContains(t, string(retried.Payload), privateMarker)
			require.NotContains(t, retried.LastErr, "PRIVATE_PROVIDER_ERROR")
			require.NotContains(t, retried.Headers["tracestate"], privateMarker)
			require.NotContains(t, retried.Headers["baggage"], privateMarker)
			if mode != "absent" {
				require.Contains(t, string(retried.Payload), parent)
				require.Contains(t, string(retried.Payload), requestID)
				require.Contains(t, string(retried.Payload), correlationID)
			}
			select {
			case valid := <-deliveries:
				require.True(t, valid)
			case <-ctx.Done():
				t.Fatal("retry delivery deadline")
			}
			require.Eventually(t, func() bool { _, e := inspector.GetTaskInfo("default", retryID); return e != nil }, 2*time.Second, 10*time.Millisecond)
			require.NoError(t, consumer.Stop())
			require.NoError(t, tel.Shutdown(ctx))
			captured := spans.GetSpans()
			require.Len(t, captured, 2)
			for _, span := range captured {
				require.Equal(t, "job.process", span.Name)
				require.Equal(t, trace.SpanKindConsumer, span.SpanKind)
				require.Empty(t, span.SpanContext.TraceState().String())
				require.Empty(t, span.Parent.TraceState().String())
				if mode != "absent" {
					require.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", span.SpanContext.TraceID().String())
					require.Equal(t, "00f067aa0ba902b7", span.Parent.SpanID().String())
					require.Len(t, span.Links, 1)
					require.Empty(t, span.Links[0].SpanContext.TraceState().String())
				}
				data, _ := json.Marshal(span)
				for _, marker := range []string{privateMarker, "PRIVATE_PROVIDER_ERROR", "private-recipient@example.com", "PrivateFirstName"} {
					require.NotContains(t, string(data), marker)
				}
			}
			logs.mu.Lock()
			defer logs.mu.Unlock()
			require.NotEmpty(t, logs.records)
			for _, r := range logs.records {
				data := r.Body().AsString()
				r.WalkAttributes(func(kv attribute.KeyValue) bool { data += string(kv.Key) + kv.Value.String(); return true })
				for _, marker := range []string{privateMarker, "PRIVATE_PROVIDER_ERROR", "private-recipient@example.com", "PrivateFirstName"} {
					require.NotContains(t, data, marker)
				}
				if r.TraceID().IsValid() && mode != "absent" {
					require.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", r.TraceID().String())
				}
			}
			for _, marker := range []string{privateMarker, "PRIVATE_PROVIDER_ERROR", "private-recipient@example.com", "PrivateFirstName"} {
				require.NotContains(t, output.String(), marker)
			}
			if mode != "absent" {
				require.Contains(t, output.String(), requestID)
				require.Contains(t, output.String(), correlationID)
				require.Contains(t, output.String(), "4bf92f3577b34da6a3ce929d0e0e4736")
			}
		})
	}
}
