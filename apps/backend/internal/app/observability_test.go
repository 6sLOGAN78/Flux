package app

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/lib/email"
	"github.com/6sLOGAN78/flux/internal/lib/job"
	loggerpkg "github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/middleware"
	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/6sLOGAN78/flux/internal/server"
	backendtesting "github.com/6sLOGAN78/flux/internal/testing"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	containerconfig "github.com/moby/moby/api/types/container"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	logcollector "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metriccollector "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracecollector "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logpb "go.opentelemetry.io/proto/otlp/logs/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

const proofParent = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
const proofRequestID = "ac1c5930-0ce6-4851-97c9-565d8192a87b"
const proofCorrelationID = "bef0ae77-1fb7-4f52-9777-34cc01939ea8"

// All failed sentinel assertions are deliberately value-free, including decoded
// protobuf and container logs. Never use assertion libraries that print inputs.
func proofClean(t *testing.T, data []byte, markers ...string) {
	t.Helper()
	for _, marker := range markers {
		if bytes.Contains(data, []byte(marker)) {
			t.Fatal("private synthetic input reached a prohibited sink")
		}
	}
}

type proofBuffer struct {
	sync.Mutex
	data bytes.Buffer
}

func (b *proofBuffer) Write(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()
	return b.data.Write(p)
}
func (b *proofBuffer) bytes() []byte { b.Lock(); defer b.Unlock(); return bytes.Clone(b.data.Bytes()) }

type proofWire struct {
	path string
	body []byte
}
type proofCapture struct {
	sync.Mutex
	wires  []proofWire
	failed bool
}

func (c *proofCapture) snapshot() []proofWire {
	c.Lock()
	defer c.Unlock()
	return append([]proofWire(nil), c.wires...)
}
func (c *proofCapture) handler(forward string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
		if err != nil {
			w.WriteHeader(400)
			return
		}
		c.Lock()
		c.wires = append(c.wires, proofWire{r.URL.Path, body})
		c.Unlock()
		if forward != "" {
			req, e := http.NewRequestWithContext(r.Context(), "POST", forward+r.URL.Path, bytes.NewReader(body))
			if e != nil {
				w.WriteHeader(500)
				return
			}
			req.Header.Set("Content-Type", "application/x-protobuf")
			response, e := (&http.Client{Timeout: 2 * time.Second}).Do(req)
			if e != nil {
				c.Lock()
				c.failed = true
				c.Unlock()
				w.WriteHeader(502)
				return
			}
			defer response.Body.Close()
			w.Header().Set("Content-Type", "application/x-protobuf")
			w.WriteHeader(response.StatusCode)
			_, _ = io.Copy(w, response.Body)
			return
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
	}
}

func proofEventually(t *testing.T, f func() bool) {
	t.Helper()
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		if f() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("integration condition exceeded its bounded deadline")
}

func proofCollector(t *testing.T) (string, *proofCapture, testcontainers.Container) {
	t.Helper()
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal("resolve repository root")
	}
	var lock map[string]json.RawMessage
	data, err := os.ReadFile(filepath.Join(root, "tools.lock.json"))
	if err != nil || json.Unmarshal(data, &lock) != nil {
		t.Fatal("read tool lock")
	}
	var pin struct {
		Image  string `json:"image"`
		Digest string `json:"digest"`
	}
	if json.Unmarshal(lock["otel-collector"], &pin) != nil || pin.Image == "" || !strings.HasPrefix(pin.Digest, "sha256:") {
		t.Fatal("collector immutable pin is required")
	}
	data, err = os.ReadFile(filepath.Join(root, "deploy/otel-collector.yaml"))
	if err != nil {
		t.Fatal("collector configuration is required")
	}
	compose, err := os.ReadFile(filepath.Join(root, "compose.yaml"))
	if err != nil || !bytes.Contains(compose, []byte(pin.Image+"@"+pin.Digest)) || !bytes.Contains(compose, []byte(`profiles: ["observability"]`)) || !bytes.Contains(compose, []byte("127.0.0.1:${FLUX_LOCAL_OTLP_HTTP_PORT")) {
		t.Fatal("collector compose pin/profile/loopback drift")
	}
	for _, signal := range []string{"traces", "metrics", "logs"} {
		if !bytes.Contains(data, []byte(signal+":")) {
			t.Fatal("collector signal pipeline missing")
		}
	}
	post := &proofCapture{}
	postServer := httptest.NewServer(post.handler(""))
	t.Cleanup(postServer.Close)
	// Linux host networking lets the real binary reach the local protobuf capture
	// without mounting Docker sockets or exposing receiver ports on the LAN.
	address := binaryTestAddress(t)
	grpcAddress := binaryTestAddress(t)
	text := strings.ReplaceAll(string(data), "0.0.0.0:4318", address)
	text = strings.ReplaceAll(text, "0.0.0.0:4317", grpcAddress)
	text = strings.Replace(text, "exporters:\n", "exporters:\n  otlphttp/proof:\n    endpoint: "+postServer.URL+"\n    compression: none\n    encoding: proto\n    retry_on_failure:\n      enabled: false\n", 1)
	text = strings.ReplaceAll(text, "exporters: [debug]", "exporters: [debug, otlphttp/proof]")
	path := filepath.Join(t.TempDir(), "collector.yaml")
	if os.WriteFile(path, []byte(text), 0600) != nil {
		t.Fatal("write collector test configuration")
	}
	image := pin.Image + "@" + pin.Digest
	// Validate the exact test configuration with the exact immutable binary.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "--network", "host", "--mount", "type=bind,src="+path+",dst=/etc/otelcol-contrib/config.yaml,readonly", image, "validate", "--config=/etc/otelcol-contrib/config.yaml")
	if output, err := cmd.CombinedOutput(); err != nil {
		_ = output
		t.Fatal("exact collector binary rejected configuration")
	}
	c, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{ContainerRequest: testcontainers.ContainerRequest{
		Image: image, Cmd: []string{"--config=/etc/otelcol-contrib/config.yaml"},
		Mounts:             testcontainers.Mounts(testcontainers.BindMount(path, "/etc/otelcol-contrib/config.yaml")),
		HostConfigModifier: func(c *containerconfig.HostConfig) { c.NetworkMode = "host" },
		WaitingFor:         wait.ForLog("Everything is ready").WithStartupTimeout(20 * time.Second),
	}, Started: true})
	testcontainers.CleanupContainer(t, c)
	if err != nil {
		t.Fatal("start pinned collector")
	}
	return "http://" + address, post, c
}

func proofSignals(t *testing.T, wires []proofWire, markers ...string) []*tracepb.Span {
	t.Helper()
	var spans []*tracepb.Span
	counts := map[string]int{}
	for _, wire := range wires {
		proofClean(t, wire.body, markers...)
		switch wire.path {
		case "/v1/traces":
			var request tracecollector.ExportTraceServiceRequest
			if proto.Unmarshal(wire.body, &request) != nil {
				t.Fatal("decode actual OTLP trace protobuf")
			}
			for _, r := range request.ResourceSpans {
				for _, scope := range r.ScopeSpans {
					for _, s := range scope.Spans {
						if s.TraceState != "" {
							t.Fatal("OTLP Span.trace_state must be empty")
						}
						for _, link := range s.Links {
							if link.TraceState != "" {
								t.Fatal("OTLP Link.trace_state must be empty")
							}
						}
						spans = append(spans, s)
						counts[wire.path]++
					}
				}
			}
		case "/v1/logs":
			var request logcollector.ExportLogsServiceRequest
			if proto.Unmarshal(wire.body, &request) != nil {
				t.Fatal("decode actual OTLP logs protobuf")
			}
			for _, r := range request.ResourceLogs {
				for _, scope := range r.ScopeLogs {
					counts[wire.path] += len(scope.LogRecords)
				}
			}
		case "/v1/metrics":
			var request metriccollector.ExportMetricsServiceRequest
			if proto.Unmarshal(wire.body, &request) != nil {
				t.Fatal("decode actual OTLP metrics protobuf")
			}
			for _, r := range request.ResourceMetrics {
				for _, scope := range r.ScopeMetrics {
					counts[wire.path] += len(scope.Metrics)
				}
			}
		default:
			t.Fatal("unexpected signal path")
		}
	}
	for _, path := range []string{"/v1/traces", "/v1/logs", "/v1/metrics"} {
		if counts[path] == 0 {
			t.Fatal("missing actual exported signal")
		}
	}
	return spans
}

type proofEmailSender struct{ endpoint string }

func (s proofEmailSender) SendWelcomeEmail(to, name string) error {
	data, _ := json.Marshal(map[string]string{"to": to, "name": name})
	response, err := (&http.Client{Timeout: time.Second}).Post(s.endpoint, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		body, _ := io.ReadAll(response.Body)
		return errors.New(string(body))
	}
	return nil
}

func proofFactories(output *proofBuffer) roleFactories {
	f := defaultRoleFactories()
	f.logger = func(cfg *config.Config, tel *observability.Telemetry) (*zerolog.Logger, func(context.Context) error, error) {
		log := loggerpkg.NewLogger(cfg.Observability, output, tel.Logger)
		return &log, nil, nil
	}
	return f
}

func proofRun(t *testing.T, r *RoleRuntime) func() {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- r.Run(ctx) }()
	var once sync.Once
	close := func() {
		once.Do(func() {
			cancel()
			select {
			case err := <-done:
				if err != nil {
					t.Error("role bounded shutdown failed")
				}
			case <-time.After(4 * time.Second):
				t.Error("role shutdown exceeded deadline")
			}
		})
	}
	t.Cleanup(close)
	return close
}

func TestTracestateHTTPRedisOTLP(t *testing.T) {
	collector, post, container := proofCollector(t)
	pre := &proofCapture{}
	proxy := httptest.NewServer(pre.handler(collector))
	defer proxy.Close()
	pg, closePG := backendtesting.SetupTestPostgres(t)
	defer closePG()
	queue, closeRedis := backendtesting.SetupTestRedis(t)
	defer closeRedis()
	secret := "private" + strconv.FormatInt(time.Now().UnixNano(), 10)
	vendorState := "vendor=" + secret
	markers := []string{secret, vendorState, "proof-recipient@example.test", "ProofPrivateName", "proof-provider-private"}
	state, err := trace.ParseTraceState(vendorState)
	if err != nil {
		t.Fatal("syntactically valid W3C test state required")
	}
	member, err := baggage.NewMember("private", secret)
	if err != nil {
		t.Fatal("syntactically valid test baggage required")
	}
	bag, _ := baggage.New(member)
	parentCtx := observability.TraceparentPropagator{}.Extract(context.Background(), mapCarrier{"traceparent": proofParent})
	linkSC := trace.SpanContextFromContext(parentCtx).WithTraceState(state)
	var output proofBuffer
	cfg := roleTestConfig()
	cfg.Database = pg.Config.Database
	cfg.Redis = queue.Config
	cfg.API.ProducerEnabled = true
	cfg.API.DrainTimeout = 3 * time.Second
	cfg.Worker.DrainTimeout = 3 * time.Second
	cfg.Integration.ResendAPIKey = "local-test-only"
	cfg.Observability.Environment = "test"
	cfg.Observability.OTLP.Enabled = true
	cfg.Observability.OTLP.Endpoint = proxy.URL
	cfg.Observability.OTLP.ExportInterval = 100 * time.Millisecond
	apiFactory := proofFactories(&output)
	var queuedID string
	var httpSpanID string
	apiFactory.router = func(_ config.Role, srv *server.Server) (*echo.Echo, error) {
		e := echo.New()
		global := middleware.NewGlobalMiddlewares(srv)
		e.HTTPErrorHandler = global.GlobalErrorHandler
		e.Use(middleware.RequestID(), middleware.NewTracingMiddleware(srv, srv.Telemetry, trace.Link{SpanContext: linkSC, Attributes: []attribute.KeyValue{attribute.String("private", secret)}}).EnhanceTracing(), middleware.NewContextEnhancer(srv).EnhanceContext(), global.RequestLogger(), global.Recover())
		e.POST("/proof/:id", func(c echo.Context) error {
			ctx := c.Request().Context()
			if c.Request().Header.Get("tracestate") != "" || c.Request().Header.Get("baggage") != "" || baggage.FromContext(ctx).Len() != 0 || trace.SpanContextFromContext(ctx).TraceState().String() != "" {
				return errors.New("unsafe ingress context")
			}
			httpSpanID = trace.SpanContextFromContext(ctx).SpanID().String()
			task, err := job.NewWelcomeEmailTaskContext(ctx, markers[2], markers[3])
			if err != nil {
				return err
			}
			info, err := srv.Job.Client.EnqueueContext(ctx, task)
			if err != nil {
				return err
			}
			queuedID = info.ID
			return c.NoContent(http.StatusAccepted)
		})
		return e, nil
	}
	api, err := newRole(context.Background(), config.RoleAPI, cfg, apiFactory)
	if err != nil {
		t.Fatal("construct real API graph")
	}
	closeAPI := proofRun(t, api)
	req, _ := http.NewRequest("POST", "http://"+api.Address()+"/proof/"+secret+"?token="+secret, nil)
	req = req.WithContext(baggage.ContextWithBaggage(context.Background(), bag))
	req.Header.Set("traceparent", proofParent)
	req.Header.Set("tracestate", vendorState)
	req.Header.Set("baggage", "private="+secret)
	req.Header.Set("X-Request-ID", proofRequestID)
	req.Header.Set("X-Correlation-ID", proofCorrelationID)
	req.Header.Set("Authorization", secret)
	response, err := (&http.Client{Timeout: time.Second}).Do(req)
	if err != nil {
		t.Fatal("send real HTTP ingress")
	}
	_ = response.Body.Close()
	if response.StatusCode != 202 {
		t.Fatal("real HTTP enqueue failed")
	}
	inspector := asynq.NewInspectorFromRedisClient(queue.Client)
	queued, err := inspector.GetTaskInfo("default", queuedID)
	if err != nil {
		t.Fatal("inspect real queued task")
	}
	proofClean(t, queued.Payload, secret, vendorState)
	assertProofMetadata(t, queued, httpSpanID)
	legacyID := "legacy-proof"
	payload, _ := json.Marshal(map[string]any{"to": markers[2], "first_name": markers[3], "metadata": map[string]any{"version": 1, "traceparent": proofParent, "request_id": proofRequestID, "correlation_id": proofCorrelationID, "tracestate": vendorState, "baggage": "private=" + secret, "unknown": secret}})
	legacy := asynq.NewTaskWithHeaders(job.TaskWelcome, payload, map[string]string{"tracestate": vendorState, "baggage": "private=" + secret}, asynq.TaskID(legacyID), asynq.Queue("low"), asynq.MaxRetry(1), asynq.Timeout(7*time.Second), asynq.Retention(time.Minute))
	if _, err = api.Server.Job.Client.Enqueue(legacy); err != nil {
		t.Fatal("enqueue legacy ingress")
	}
	// Both originals are inspected before a consumer starts. The fake transport
	// fails every first attempt, allowing inspection of persisted retry metadata.
	var attemptsMu sync.Mutex
	attempts := 0
	emailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p map[string]string
		if json.NewDecoder(r.Body).Decode(&p) != nil || p["to"] != markers[2] || p["name"] != markers[3] {
			t.Error("local email transport received invalid private payload")
		}
		attemptsMu.Lock()
		attempts++
		n := attempts
		attemptsMu.Unlock()
		if n <= 2 {
			w.WriteHeader(503)
			_, _ = io.WriteString(w, markers[4])
			return
		}
		w.WriteHeader(200)
	}))
	defer emailServer.Close()
	workerFactory := proofFactories(&output)
	workerFactory.consumer = func(srv *server.Server, _ *email.Client) (*job.JobService, func(context.Context) error, error) {
		j := job.NewConsumer(srv.Logger, cfg, srv.Redis, proofEmailSender{emailServer.URL}, srv.Telemetry)
		return j, func(context.Context) error { return j.Stop() }, nil
	}
	worker, err := newRole(context.Background(), config.RoleWorker, cfg, workerFactory)
	if err != nil {
		t.Fatal("construct real worker graph")
	}
	closeWorker := proofRun(t, worker)
	for _, pair := range []struct{ queue, id, parent string }{{"default", queuedID, httpSpanID}, {"low", "safe:" + legacyID, "00f067aa0ba902b7"}} {
		var retry *asynq.TaskInfo
		proofEventually(t, func() bool {
			retry, err = inspector.GetTaskInfo(pair.queue, pair.id)
			return err == nil && retry.State == asynq.TaskStateRetry
		})
		proofClean(t, retry.Payload, secret, vendorState)
		proofClean(t, []byte(retry.LastErr), markers...)
		if len(retry.Headers) != 0 || retry.Retried != 1 {
			t.Fatal("unsafe or unexecuted retry envelope")
		}
		assertProofMetadata(t, retry, pair.parent)
		if pair.queue == "low" {
			if _, err = inspector.GetTaskInfo("low", legacyID); !errors.Is(err, asynq.ErrTaskNotFound) {
				t.Fatal("unsafe legacy original was not revoked")
			}
			if retry.MaxRetry != 1 || retry.Timeout != 7*time.Second || retry.Retention != time.Minute {
				t.Fatal("legacy replacement lost execution policy")
			}
		}
	}
	// Public inspector acceleration retains actual Asynq retry execution while
	// avoiding the production backoff delay in this integration proof.
	for _, pair := range []struct{ queue, id string }{{"default", queuedID}, {"low", "safe:" + legacyID}} {
		if err = inspector.RunTask(pair.queue, pair.id); err != nil {
			t.Fatal("release real retry")
		}
	}
	proofEventually(t, func() bool { attemptsMu.Lock(); defer attemptsMu.Unlock(); return attempts == 4 })
	proofEventually(t, func() bool {
		_, a := inspector.GetTaskInfo("default", queuedID)
		b, e := inspector.GetTaskInfo("low", "safe:"+legacyID)
		return errors.Is(a, asynq.ErrTaskNotFound) && e == nil && b.State == asynq.TaskStateCompleted
	})
	completed, err := inspector.GetTaskInfo("low", "safe:"+legacyID)
	if err != nil {
		t.Fatal("inspect completed replacement")
	}
	proofClean(t, completed.Payload, secret, vendorState)
	closeWorker()
	closeAPI()
	proofEventually(t, func() bool { return len(post.snapshot()) >= 3 })
	pre.Lock()
	failed := pre.failed
	pre.Unlock()
	if failed {
		t.Fatal("actual collector forwarding failed")
	}
	spans := proofSignals(t, pre.snapshot(), markers...)
	checkProofLineage(t, spans, httpSpanID, true)
	postSpans := proofSignals(t, post.snapshot(), markers...)
	checkProofLineage(t, postSpans, httpSpanID, false)
	proofClean(t, output.bytes(), markers...)
	logs, err := container.Logs(context.Background())
	if err != nil {
		t.Fatal("capture collector debug output")
	}
	debug, err := io.ReadAll(logs)
	_ = logs.Close()
	if err != nil {
		t.Fatal("read collector debug output")
	}
	proofClean(t, debug, markers...)
	if !bytes.Contains(debug, []byte("job.process")) || !bytes.Contains(debug, []byte("http.request")) {
		t.Fatal("debug exporter did not receive redacted work")
	}
	for _, path := range []string{"../../static/openapi.json", "../../templates/emails/welcome.html", "../../../../packages/openapi/openapi.json"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal("read generated sentinel scan artifact")
		}
		proofClean(t, data, markers...)
	}
}

type mapCarrier map[string]string

func (c mapCarrier) Get(k string) string { return c[k] }
func (c mapCarrier) Set(k, v string)     { c[k] = v }
func (c mapCarrier) Keys() []string {
	var keys []string
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

func assertProofMetadata(t *testing.T, info *asynq.TaskInfo, parent string) {
	t.Helper()
	var p job.WelcomeEmailPayload
	if json.Unmarshal(info.Payload, &p) != nil || p.Metadata == nil || p.Metadata.Version != 1 || p.Metadata.RequestID != proofRequestID || p.Metadata.CorrelationID != proofCorrelationID || p.Metadata.Traceparent != "00-4bf92f3577b34da6a3ce929d0e0e4736-"+parent+"-01" {
		t.Fatal("queue/retry lost safe identity or sampled flags")
	}
}

func checkProofLineage(t *testing.T, spans []*tracepb.Span, httpSpan string, linksRequired bool) {
	t.Helper()
	jobs := 0
	httpSeen := false
	retries := map[int64]int{}
	for _, s := range spans {
		if s.Name != "http.request" && s.Name != "job.process" {
			continue
		}
		if hex.EncodeToString(s.TraceId) != "4bf92f3577b34da6a3ce929d0e0e4736" || s.Flags&1 != 1 {
			t.Fatal("trace identity or sampled flag lost")
		}
		ids := map[string]string{}
		for _, a := range s.Attributes {
			ids[a.Key] = a.Value.GetStringValue()
			if a.Key == "retry.count" {
				retries[a.Value.GetIntValue()]++
			}
		}
		if ids["request_id"] != proofRequestID || ids["correlation_id"] != proofCorrelationID {
			t.Fatal("export lost UUID lineage")
		}
		if s.Name == "http.request" {
			httpSeen = true
			if hex.EncodeToString(s.ParentSpanId) != "00f067aa0ba902b7" || hex.EncodeToString(s.SpanId) != httpSpan {
				t.Fatal("HTTP trace parent lost")
			}
		}
		if s.Name == "job.process" {
			jobs++
			parent := hex.EncodeToString(s.ParentSpanId)
			if parent != httpSpan && parent != "00f067aa0ba902b7" {
				t.Fatal("worker queue parent lost")
			}
		}
		if linksRequired {
			if len(s.Links) == 0 {
				t.Fatal("required injected/worker Link was not exported")
			}
			for _, l := range s.Links {
				if l.TraceState != "" || hex.EncodeToString(l.TraceId) != "4bf92f3577b34da6a3ce929d0e0e4736" || l.Flags&1 != 1 {
					t.Fatal("Link trace identity/flags/state violated")
				}
			}
		} else if len(s.Links) != 0 {
			t.Fatal("collector did not clear link collections")
		}
	}
	if !httpSeen || jobs != 4 || retries[0] != 2 || retries[1] != 2 {
		t.Fatal("real HTTP, execution and retry spans were not all captured")
	}
}

func TestCollectorRedactsAllSignals(t *testing.T) {
	endpoint, post, c := proofCollector(t)
	secret := "dirty" + strconv.FormatInt(time.Now().UnixNano(), 10)
	kv := func(k, v string) *commonpb.KeyValue {
		return &commonpb.KeyValue{Key: k, Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: v}}}
	}
	resource := &resourcepb.Resource{Attributes: []*commonpb.KeyValue{kv("service.name", "flux.api"), kv("private", secret)}}
	scope := &commonpb.InstrumentationScope{Name: secret, Version: secret, Attributes: []*commonpb.KeyValue{kv("private", secret)}}
	traceID, _ := hex.DecodeString("4bf92f3577b34da6a3ce929d0e0e4736")
	spanID, _ := hex.DecodeString("00f067aa0ba902b7")
	messages := map[string]proto.Message{
		"/v1/traces":  &tracecollector.ExportTraceServiceRequest{ResourceSpans: []*tracepb.ResourceSpans{{Resource: resource, ScopeSpans: []*tracepb.ScopeSpans{{Scope: scope, Spans: []*tracepb.Span{{TraceId: traceID, SpanId: spanID, Name: secret, TraceState: "vendor=" + secret, Attributes: []*commonpb.KeyValue{kv("private", secret), kv("request_id", proofRequestID)}, Status: &tracepb.Status{Message: secret}, Events: []*tracepb.Span_Event{{Name: secret}}, Links: []*tracepb.Span_Link{{TraceId: traceID, SpanId: spanID, TraceState: "vendor=" + secret, Attributes: []*commonpb.KeyValue{kv("private", secret)}}}}}}}}}},
		"/v1/logs":    &logcollector.ExportLogsServiceRequest{ResourceLogs: []*logpb.ResourceLogs{{Resource: resource, ScopeLogs: []*logpb.ScopeLogs{{Scope: scope, LogRecords: []*logpb.LogRecord{{Body: kv("", secret).Value, SeverityText: secret, EventName: secret, Attributes: []*commonpb.KeyValue{kv("private", secret), kv("request_id", proofRequestID)}}}}}}}},
		"/v1/metrics": &metriccollector.ExportMetricsServiceRequest{ResourceMetrics: []*metricpb.ResourceMetrics{{Resource: resource, ScopeMetrics: []*metricpb.ScopeMetrics{{Scope: scope, Metrics: []*metricpb.Metric{{Name: "flux.http.requests", Description: secret, Unit: secret, Data: &metricpb.Metric_Sum{Sum: &metricpb.Sum{AggregationTemporality: metricpb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE, DataPoints: []*metricpb.NumberDataPoint{{Attributes: []*commonpb.KeyValue{kv("private", secret), kv("request_id", proofRequestID), kv("http.route", "/live")}, Value: &metricpb.NumberDataPoint_AsInt{AsInt: 1}}}}}}}}}}}},
	}
	for path, message := range messages {
		data, err := proto.Marshal(message)
		if err != nil {
			t.Fatal("encode dirty probe")
		}
		r, err := http.Post(endpoint+path, "application/x-protobuf", bytes.NewReader(data))
		if err != nil {
			t.Fatal("send dirty collector probe")
		}
		_ = r.Body.Close()
		if r.StatusCode != 200 {
			t.Fatal("collector rejected dirty test transport")
		}
	}
	proofEventually(t, func() bool { return len(post.snapshot()) >= 3 })
	spans := proofSignals(t, post.snapshot(), secret)
	if len(spans) != 1 || len(spans[0].Links) != 0 || len(spans[0].Events) != 0 {
		t.Fatal("collector failed to remove private nested fields")
	}
	logs, err := c.Logs(context.Background())
	if err != nil {
		t.Fatal("capture dirty probe debug output")
	}
	data, err := io.ReadAll(logs)
	_ = logs.Close()
	if err != nil {
		t.Fatal("read dirty probe output")
	}
	proofClean(t, data, secret)
}

func TestCardinalityVariedHTTPInputs(t *testing.T) {
	collector, post, _ := proofCollector(t)
	var output proofBuffer
	cfg := roleTestConfig()
	cfg.Observability.OTLP.Enabled = true
	cfg.Observability.OTLP.Endpoint = collector
	cfg.Observability.OTLP.ExportInterval = 100 * time.Millisecond
	r, err := newRole(context.Background(), config.RoleRedirector, cfg, proofFactories(&output))
	if err != nil {
		t.Fatal("construct cardinality role")
	}
	closeRole := proofRun(t, r)
	client := &http.Client{Timeout: time.Second}
	for i := range 200 {
		response, err := client.Get(fmt.Sprintf("http://%s/visitor/%d?token=unique%d", r.Address(), i, i))
		if err != nil {
			t.Fatal("varied HTTP request failed")
		}
		_ = response.Body.Close()
	}
	closeRole()
	proofEventually(t, func() bool {
		for _, w := range post.snapshot() {
			if w.path == "/v1/metrics" {
				return true
			}
		}
		return false
	})
	series := map[string]bool{}
	count := int64(0)
	for _, w := range post.snapshot() {
		if w.path != "/v1/metrics" {
			continue
		}
		var req metriccollector.ExportMetricsServiceRequest
		if proto.Unmarshal(w.body, &req) != nil {
			t.Fatal("decode cardinality metrics")
		}
		for _, r := range req.ResourceMetrics {
			for _, s := range r.ScopeMetrics {
				for _, m := range s.Metrics {
					if m.Name != "flux.http.requests" {
						continue
					}
					for _, p := range m.GetSum().DataPoints {
						data, _ := json.Marshal(p.Attributes)
						series[string(data)] = true
						count = max(count, p.GetAsInt())
						for _, a := range p.Attributes {
							if a.Key == "http.route" && a.Value.GetStringValue() != "unmatched" {
								t.Fatal("metric route cardinality expanded")
							}
							if a.Key == "request_id" || a.Key == "correlation_id" {
								t.Fatal("UUID entered metric label")
							}
						}
					}
				}
			}
		}
	}
	if len(series) != 1 || count != 200 {
		t.Fatal("varied paths did not retain one fixed metric series")
	}
}

func TestTelemetryOutageReadinessAndCleanup(t *testing.T) {
	pg, closePG := backendtesting.SetupTestPostgres(t)
	defer closePG()
	queue, closeRedis := backendtesting.SetupTestRedis(t)
	defer closeRedis()
	for _, mode := range []string{"disconnected", "slow"} {
		t.Run(mode, func(t *testing.T) {
			endpoint := "http://" + binaryTestAddress(t)
			if mode == "slow" {
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					select {
					case <-r.Context().Done():
					case <-time.After(time.Second):
					}
				}))
				defer s.Close()
				endpoint = s.URL
			}
			for _, role := range []config.Role{config.RoleAPI, config.RoleRedirector, config.RoleWorker} {
				t.Run(string(role), func(t *testing.T) {
					var output proofBuffer
					cfg := roleTestConfig()
					cfg.Database = pg.Config.Database
					cfg.Redis = queue.Config
					cfg.API.ProducerEnabled = true
					cfg.Integration.ResendAPIKey = "local-test-only"
					cfg.Observability.OTLP.Enabled = true
					cfg.Observability.OTLP.Endpoint = endpoint
					cfg.Observability.OTLP.ExportTimeout = 80 * time.Millisecond
					cfg.Observability.OTLP.ExportInterval = 10 * time.Millisecond
					f := proofFactories(&output)
					closed := map[string]bool{}
					baseTelemetry := f.telemetry
					f.telemetry = func(ctx context.Context, cfg *config.Config, role config.Role) (*observability.Telemetry, func(context.Context) error, error) {
						owner, close, err := baseTelemetry(ctx, cfg, role)
						return owner, func(ctx context.Context) error { closed["telemetry"] = true; return close(ctx) }, err
					}
					baseDB := f.database
					f.database = func(ctx context.Context, s *server.Server) (*database.Database, func(context.Context) error, error) {
						db, close, err := baseDB(ctx, s)
						return db, func(ctx context.Context) error { closed["database"] = true; return close(ctx) }, err
					}
					baseRedis := f.redis
					f.redis = func(ctx context.Context, s *server.Server) (*redis.Client, func(context.Context) error, error) {
						db, close, err := baseRedis(ctx, s)
						return db, func(ctx context.Context) error { closed["redis"] = true; return close(ctx) }, err
					}
					baseProducer := f.producer
					f.producer = func(s *server.Server) (*job.JobService, func(context.Context) error, error) {
						j, close, err := baseProducer(s)
						return j, func(ctx context.Context) error { closed["producer"] = true; return close(ctx) }, err
					}
					baseConsumer := f.consumer
					f.consumer = func(s *server.Server, e *email.Client) (*job.JobService, func(context.Context) error, error) {
						j, close, err := baseConsumer(s, e)
						return j, func(ctx context.Context) error { closed["consumer"] = true; return close(ctx) }, err
					}
					r, err := newRole(context.Background(), role, cfg, f)
					if err != nil {
						t.Fatal("collector outage prevented role construction")
					}
					rec := httptest.NewRecorder()
					r.HTTP.ServeHTTP(rec, httptest.NewRequest("GET", "/ready", nil))
					if rec.Code != 200 {
						t.Fatal("collector outage changed required-dependency readiness")
					}
					ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
					defer cancel()
					start := time.Now()
					_ = r.Close(ctx)
					if time.Since(start) > 750*time.Millisecond {
						t.Fatal("collector outage exceeded shared shutdown deadline")
					}
					for _, name := range []string{"telemetry", "database", "redis", "producer", "consumer"} {
						want := name == "telemetry" || (role == config.RoleAPI && (name == "database" || name == "redis" || name == "producer")) || (role == config.RoleWorker && (name == "redis" || name == "consumer"))
						if want && !closed[name] {
							t.Fatal("collector outage skipped owned resource closer")
						}
					}
					assertRoleListenerReleased(t, r.Address())
				})
			}
		})
	}
}
