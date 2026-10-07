//nolint:testpackage // These tests verify package-private lifecycle and failure-injection seams.
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
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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
	data bytes.Buffer
	sync.Mutex
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
	wires []proofWire
	sync.Mutex
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
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		c.Lock()
		c.wires = append(c.wires, proofWire{path: r.URL.Path, body: body})
		c.Unlock()
		if forward != "" {
			req,
				e := http.NewRequestWithContext(r.Context(),
				http.MethodPost,
				forward+r.URL.Path,
				bytes.NewReader(body))
			if e != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Header.Set("Content-Type", "application/x-protobuf")
			response, e := (&http.Client{Timeout: 2 * time.Second}).Do(req)
			if e != nil {
				c.Lock()
				c.failed = true
				c.Unlock()
				w.WriteHeader(http.StatusBadGateway)
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
	if json.Unmarshal(lock["otel-collector"],
		&pin) != nil ||
		pin.Image == "" ||
		!strings.HasPrefix(pin.Digest,
			"sha256:") {
		t.Fatal("collector immutable pin is required")
	}
	data, err = os.ReadFile(filepath.Join(root, "deploy/otel-collector.yaml"))
	if err != nil {
		t.Fatal("collector configuration is required")
	}
	compose, err := os.ReadFile(filepath.Join(root, "compose.yaml"))
	if err != nil ||
		!bytes.Contains(compose,
			[]byte(pin.Image+"@"+pin.Digest)) ||
		!bytes.Contains(compose,
			[]byte(`profiles: ["observability"]`)) ||
		!bytes.Contains(compose,
			[]byte("127.0.0.1:${FLUX_LOCAL_OTLP_HTTP_PORT")) {
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
	text = strings.Replace(text,
		"exporters:\n",
		"exporters:\n  otlphttp/proof:\n    endpoint: "+postServer.URL+
			"\n    compression: none\n    encoding: proto\n"+
			"    retry_on_failure:\n      enabled: false\n",
		1)
	text = strings.ReplaceAll(text, "exporters: [debug]", "exporters: [debug, otlphttp/proof]")
	path := filepath.Join(t.TempDir(), "collector.yaml")
	if os.WriteFile(path, []byte(text), 0644) != nil {
		t.Fatal("write collector test configuration")
	}
	image := pin.Image + "@" + pin.Digest
	// Validate the exact test configuration with the exact immutable binary.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx,
		"docker",
		"run",
		"--rm",
		"--network",
		"host",
		"--mount",
		"type=bind,src="+path+",dst=/etc/otelcol-contrib/config.yaml,readonly",
		image,
		"validate",
		"--config=/etc/otelcol-contrib/config.yaml")
	if output, err198 := cmd.CombinedOutput(); err198 != nil {
		t.Fatalf("exact collector binary rejected configuration: %s", output)
	}
	c,
		err := testcontainers.GenericContainer(context.Background(),
		testcontainers.GenericContainerRequest{ContainerRequest: testcontainers.ContainerRequest{
			Image: image, Cmd: []string{"--config=/etc/otelcol-contrib/config.yaml"},
			Files: []testcontainers.ContainerFile{{HostFilePath: path,
				ContainerFilePath: "/etc/otelcol-contrib/config.yaml",
				FileMode:          0o644}},
			HostConfigModifier: func(c *containerconfig.HostConfig) { c.NetworkMode = "host" },
			WaitingFor:         wait.ForLog("Everything is ready").WithStartupTimeout(20 * time.Second),
		}, Started: true})
	testcontainers.CleanupContainer(t, c)
	if err != nil {
		t.Fatal("start pinned collector")
	}
	return "http://" + address, post, c
}

//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
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
			for _, r := range request.GetResourceSpans() {
				for _, scope := range r.GetScopeSpans() {
					for _, s := range scope.GetSpans() {
						if s.GetTraceState() != "" {
							t.Fatal("OTLP Span.trace_state must be empty")
						}
						for _, link := range s.GetLinks() {
							if link.GetTraceState() != "" {
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
			for _, r := range request.GetResourceLogs() {
				for _, scope := range r.GetScopeLogs() {
					counts[wire.path] += len(scope.GetLogRecords())
				}
			}
		case "/v1/metrics":
			var request metriccollector.ExportMetricsServiceRequest
			if proto.Unmarshal(wire.body, &request) != nil {
				t.Fatal("decode actual OTLP metrics protobuf")
			}
			for _, r := range request.GetResourceMetrics() {
				for _, scope := range r.GetScopeMetrics() {
					counts[wire.path] += len(scope.GetMetrics())
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
	response,
		err := (&http.Client{Timeout: time.Second}).
		Post(s.endpoint,
			"application/json",
			bytes.NewReader(data))
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
	f.logger = func(cfg *config.Config,
		tel *observability.Telemetry) (*zerolog.Logger,
		func(context.Context) error,
		error) {
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
	cleanup := func() {
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
	t.Cleanup(cleanup)
	return cleanup
}

//nolint:gocognit,gocyclo,cyclop // Keep this complete integration protocol and its ordered failure assertions together.
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
	markers := []string{secret,
		vendorState,
		"proof-recipient@example.test",
		"ProofPrivateName",
		"proof-provider-private"}
	state, err := trace.ParseTraceState(vendorState)
	if err != nil {
		t.Fatal("syntactically valid W3C test state required")
	}
	member, err := baggage.NewMember("private", secret)
	if err != nil {
		t.Fatal("syntactically valid test baggage required")
	}
	bag, _ := baggage.New(member)
	parentCtx := observability.TraceparentPropagator{}.Extract(context.Background(),
		mapCarrier{"traceparent": proofParent})
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
	enqueued := make(chan struct{}, 1)
	apiFactory.router = func(_ config.Role, srv *server.Server) (*echo.Echo, error) {
		e := echo.New()
		global := middleware.NewGlobalMiddlewares(srv)
		e.HTTPErrorHandler = global.GlobalErrorHandler
		e.Use(middleware.RequestID(),
			middleware.NewTracingMiddleware(srv,
				srv.Telemetry,
				trace.Link{SpanContext: linkSC,
					Attributes: []attribute.KeyValue{attribute.String("private",
						secret)}}).
				EnhanceTracing(),
			middleware.NewContextEnhancer(srv).
				EnhanceContext(),
			global.RequestLogger(),
			global.Recover())
		e.POST("/proof/:id", func(c echo.Context) error {
			ctx := c.Request().Context()
			if c.Request().
				Header.Get("Tracestate") != "" ||
				c.Request().
					Header.Get("Baggage") != "" ||
				baggage.FromContext(ctx).
					Len() != 0 ||
				trace.SpanContextFromContext(ctx).
					TraceState().
					String() != "" {
				return errors.New("unsafe ingress context")
			}
			httpSpanID = trace.SpanContextFromContext(ctx).SpanID().String()
			task, err372 := job.NewWelcomeEmailTaskContext(ctx, markers[2], markers[3])
			if err372 != nil {
				return err372
			}
			info, err372 := srv.Job.Client.EnqueueContext(ctx, task)
			if err372 != nil {
				return err372
			}
			queuedID = info.ID
			enqueued <- struct{}{}
			return c.NoContent(http.StatusAccepted)
		})
		return e, nil
	}
	api, err := newRole(context.Background(), config.RoleAPI, cfg, apiFactory)
	if err != nil {
		t.Fatal("construct real API graph")
	}
	closeAPI := proofRun(t, api)
	req, _ := http.NewRequest(http.MethodPost, "http://"+api.Address()+"/proof/"+secret+"?token="+secret, nil)
	req = req.WithContext(baggage.ContextWithBaggage(context.Background(), bag))
	req.Header.Set("Traceparent", proofParent)
	req.Header.Set("Tracestate", vendorState)
	req.Header.Set("Baggage", "private="+secret)
	req.Header.Set("X-Request-ID", proofRequestID)
	req.Header.Set("X-Correlation-ID", proofCorrelationID)
	req.Header.Set("Authorization", secret)
	response, err := (&http.Client{Timeout: time.Second}).Do(req)
	if err != nil {
		t.Fatal("send real HTTP ingress")
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusAccepted {
		t.Fatal("real HTTP enqueue failed")
	}
	<-enqueued
	inspector := asynq.NewInspectorFromRedisClient(queue.Client)
	queued, err := inspector.GetTaskInfo("default", queuedID)
	if err != nil {
		t.Fatal("inspect real queued task")
	}
	proofClean(t, queued.Payload, secret, vendorState)
	assertProofMetadata(t, queued, httpSpanID)
	legacyID := "legacy-proof"
	payload,
		_ := json.Marshal(map[string]any{"to": markers[2],
		"first_name": markers[3],
		"metadata": map[string]any{"version": 1,
			"traceparent":    proofParent,
			"request_id":     proofRequestID,
			"correlation_id": proofCorrelationID,
			"tracestate":     vendorState,
			"baggage":        "private=" + secret,
			"unknown":        secret}})
	legacy := asynq.NewTaskWithHeaders(job.TaskWelcome,
		payload,
		map[string]string{"tracestate": vendorState,
			"baggage": "private=" + secret},
		asynq.TaskID(legacyID),
		asynq.Queue("low"),
		asynq.MaxRetry(1),
		asynq.Timeout(7*time.Second),
		asynq.Retention(time.Minute))
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
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, markers[4])
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer emailServer.Close()
	workerFactory := proofFactories(&output)
	workerFactory.consumer = func(srv *server.Server,
		_ *email.Client) (*job.JobService,
		func(context.Context) error,
		error) {
		j := job.NewConsumer(srv.Logger,
			cfg,
			srv.Redis,
			proofEmailSender{endpoint: emailServer.URL},
			srv.Telemetry)
		return j, func(context.Context) error { return j.Stop() }, nil
	}
	worker, err := newRole(context.Background(), config.RoleWorker, cfg, workerFactory)
	if err != nil {
		t.Fatal("construct real worker graph")
	}
	closeWorker := proofRun(t, worker)
	for _, pair := range []struct {
		queue,
		id,
		parent string
	}{{queue: "default",
		id:     queuedID,
		parent: httpSpanID},
		{queue: "low",
			id:     "safe:" + legacyID,
			parent: "00f067aa0ba902b7"}} {
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
	for _, pair := range []struct {
		queue,
		id string
	}{{queue: "default",
		id: queuedID},
		{queue: "low",
			id: "safe:" + legacyID}} {
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
	for _, path := range []string{"../../static/openapi.json",
		"../../templates/emails/welcome.html",
		"../../../../packages/openapi/openapi.json"} {
		data, err519 := os.ReadFile(path)
		if err519 != nil {
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
	if json.Unmarshal(info.Payload,
		&p) != nil ||
		p.Metadata == nil ||
		p.Metadata.Version != 1 ||
		p.Metadata.RequestID != proofRequestID ||
		p.Metadata.CorrelationID != proofCorrelationID ||
		p.Metadata.Traceparent != "00-4bf92f3577b34da6a3ce929d0e0e4736-"+parent+"-01" {
		t.Fatal("queue/retry lost safe identity or sampled flags")
	}
}

//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
func checkProofLineage(t *testing.T, spans []*tracepb.Span, httpSpan string, linksRequired bool) {
	t.Helper()
	jobs := 0
	httpSeen := false
	retries := map[int64]int{}
	for _, s := range spans {
		if s.GetName() != "http.request" && s.GetName() != "job.process" {
			continue
		}
		if hex.EncodeToString(s.GetTraceId()) != "4bf92f3577b34da6a3ce929d0e0e4736" || s.GetFlags()&1 != 1 {
			t.Fatal("trace identity or sampled flag lost")
		}
		ids := map[string]string{}
		for _, a := range s.GetAttributes() {
			ids[a.GetKey()] = a.GetValue().GetStringValue()
			if a.GetKey() == "retry.count" {
				retries[a.GetValue().GetIntValue()]++
			}
		}
		if ids["request_id"] != proofRequestID || ids["correlation_id"] != proofCorrelationID {
			t.Fatal("export lost UUID lineage")
		}
		if s.GetName() == "http.request" {
			httpSeen = true
			if hex.EncodeToString(s.GetParentSpanId()) != "00f067aa0ba902b7" ||
				hex.EncodeToString(s.GetSpanId()) != httpSpan {
				t.Fatal("HTTP trace parent lost")
			}
		}
		if s.GetName() == "job.process" {
			jobs++
			parent := hex.EncodeToString(s.GetParentSpanId())
			if parent != httpSpan && parent != "00f067aa0ba902b7" {
				t.Fatal("worker queue parent lost")
			}
		}
		if linksRequired {
			if len(s.GetLinks()) == 0 {
				t.Fatal("required injected/worker Link was not exported")
			}
			for _, l := range s.GetLinks() {
				if l.GetTraceState() != "" ||
					hex.EncodeToString(l.GetTraceId()) != "4bf92f3577b34da6a3ce929d0e0e4736" ||
					len(l.GetSpanId()) != 8 ||
					bytes.Equal(l.GetSpanId(),
						make([]byte,
							8)) ||
					l.GetFlags()&1 != 1 {
					t.Fatal("Link trace identity/flags/state violated")
				}
			}
		} else if len(s.GetLinks()) != 0 {
			t.Fatal("collector did not clear link collections")
		}
	}
	if !httpSeen || jobs != 4 || retries[0] != 2 || retries[1] != 2 {
		t.Fatal("real HTTP, execution and retry spans were not all captured")
	}
}

//nolint:gocognit,gocyclo,cyclop // Keep this complete integration protocol and its ordered failure assertions together.
func TestCollectorRedactsAllSignals(t *testing.T) {
	endpoint, post, c := proofCollector(t)
	secret := "dirty" + strconv.FormatInt(time.Now().UnixNano(), 10)
	kv := func(k, v string) *commonpb.KeyValue {
		return &commonpb.KeyValue{Key: k,
			Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: v}}}
	}
	resource := &resourcepb.Resource{Attributes: []*commonpb.KeyValue{kv("service.name",
		"flux.api"),
		kv("private",
			secret)}}
	resource.Attributes = append(resource.Attributes,
		kv("process.role",
			secret),
		kv("deployment.environment.name",
			secret))
	scope := &commonpb.InstrumentationScope{Name: secret,
		Version: secret,
		Attributes: []*commonpb.KeyValue{kv("private",
			secret)}}
	traceID, _ := hex.DecodeString("4bf92f3577b34da6a3ce929d0e0e4736")
	spanID, _ := hex.DecodeString("00f067aa0ba902b7")
	messages := map[string]proto.Message{
		"/v1/traces": &tracecollector.ExportTraceServiceRequest{ResourceSpans: []*tracepb.ResourceSpans{{Resource: resource,
			ScopeSpans: []*tracepb.ScopeSpans{{Scope: scope,
				Spans: []*tracepb.Span{{TraceId: traceID,
					SpanId:     spanID,
					Name:       secret,
					TraceState: "vendor=" + secret,
					Attributes: []*commonpb.KeyValue{kv("private",
						secret),
						kv("request_id",
							proofRequestID)},
					Status: &tracepb.Status{Message: secret},
					Events: []*tracepb.Span_Event{{Name: secret}},
					Links: []*tracepb.Span_Link{{TraceId: traceID,
						SpanId:     spanID,
						TraceState: "vendor=" + secret,
						Attributes: []*commonpb.KeyValue{kv("private",
							secret)}}}}}}}}}},
		"/v1/logs": &logcollector.ExportLogsServiceRequest{ResourceLogs: []*logpb.ResourceLogs{{Resource: resource,
			ScopeLogs: []*logpb.ScopeLogs{{Scope: scope,
				LogRecords: []*logpb.LogRecord{{Body: kv("",
					secret).
					GetValue(),
					SeverityText: secret,
					EventName:    secret,
					Attributes: []*commonpb.KeyValue{kv("private",
						secret),
						kv("request_id",
							proofRequestID)}}}}}}}},
		"/v1/metrics": &metriccollector.ExportMetricsServiceRequest{ResourceMetrics: []*metricpb.ResourceMetrics{
			{Resource: resource,
				ScopeMetrics: []*metricpb.ScopeMetrics{{Scope: scope,
					Metrics: []*metricpb.Metric{{Name: "flux.http.requests",
						Description: secret,
						Unit:        secret,
						Data: &metricpb.Metric_Sum{Sum: &metricpb.Sum{
							AggregationTemporality: metricpb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
							DataPoints: []*metricpb.NumberDataPoint{{Attributes: []*commonpb.KeyValue{kv("private",
								secret),
								kv("request_id",
									proofRequestID),
								kv("http.route",
									"/live")},
								Value: &metricpb.NumberDataPoint_AsInt{AsInt: 1}}}}}}}}}}}},
	}
	// Challenge values under allowed keys independently of the SDK sanitizer.
	// A key-only collector allowlist would let every one of these values leak.
	unsafeValues := func() []*commonpb.KeyValue {
		var attrs []*commonpb.KeyValue
		for _, key := range []string{"operation",
			"process.role",
			"http.request.method",
			"http.route",
			"outcome",
			"dependency",
			"error.category",
			"error.stage",
			"job.type",
			"db.system.name",
			"db.operation.name",
			"correlation_id",
			"http.response.status_code",
			"retry.count",
			"start_version",
			"end_version"} {
			attrs = append(attrs, kv(key, secret))
		}
		return attrs
	}
	dirtySpan := messages["/v1/traces"].(*tracecollector.ExportTraceServiceRequest).
		GetResourceSpans()[0].GetScopeSpans()[0].GetSpans()[0]
	dirtySpan.Attributes = append(dirtySpan.Attributes, unsafeValues()...)
	dirtyLog := messages["/v1/logs"].(*logcollector.ExportLogsServiceRequest).
		GetResourceLogs()[0].GetScopeLogs()[0].GetLogRecords()[0]
	dirtyLog.Attributes = append(dirtyLog.Attributes, unsafeValues()...)
	dirtyPoint := messages["/v1/metrics"].(*metriccollector.ExportMetricsServiceRequest).
		GetResourceMetrics()[0].GetScopeMetrics()[0].GetMetrics()[0].GetSum().
		GetDataPoints()[0]
	// OTLP requires attribute keys to be unique. Replace the original route,
	// rather than append a second key whose map lookup would be undefined.
	dirtyPoint.Attributes = dirtyPoint.GetAttributes()[:2]
	dirtyPoint.Attributes = append(dirtyPoint.Attributes, unsafeValues()...)
	dirtyPoint.Exemplars = []*metricpb.Exemplar{{FilteredAttributes: []*commonpb.KeyValue{kv("private",
		secret)},
		TraceId: traceID,
		SpanId:  spanID,
		Value:   &metricpb.Exemplar_AsInt{AsInt: 1}}}
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
		if r.StatusCode != http.StatusOK {
			t.Fatal("collector rejected dirty test transport")
		}
	}
	proofEventually(t, func() bool { return len(post.snapshot()) >= 3 })
	for _, wire := range post.snapshot() {
		if bytes.Contains(wire.body, []byte(secret)) {
			t.Fatalf("dirty probe retained private input in %s", wire.path)
		}
	}
	spans := proofSignals(t, post.snapshot(), secret)
	if len(spans) != 1 || len(spans[0].GetLinks()) != 0 || len(spans[0].GetEvents()) != 0 {
		t.Fatal("collector failed to remove private nested fields")
	}
	if len(spans[0].GetAttributes()) != 1 ||
		spans[0].GetAttributes()[0].GetKey() != "request_id" ||
		spans[0].GetAttributes()[0].GetValue().
			GetStringValue() != proofRequestID {
		t.Fatal("collector did not preserve only valid trace correlation")
	}
	for _, wire := range post.snapshot() {
		switch wire.path {
		case "/v1/logs":
			var request logcollector.ExportLogsServiceRequest
			if proto.Unmarshal(wire.body, &request) != nil {
				t.Fatal("decode dirty probe log output")
			}
			for _, r := range request.GetResourceLogs() {
				for _, scope := range r.GetScopeLogs() {
					for _, record := range scope.GetLogRecords() {
						if len(record.GetAttributes()) != 1 ||
							record.GetAttributes()[0].GetKey() != "request_id" ||
							record.GetAttributes()[0].GetValue().
								GetStringValue() != proofRequestID {
							t.Fatal("collector did not preserve only valid log correlation")
						}
					}
				}
			}
		case "/v1/metrics":
			var request metriccollector.ExportMetricsServiceRequest
			if proto.Unmarshal(wire.body, &request) != nil {
				t.Fatal("decode dirty probe metric output")
			}
			for _, r := range request.GetResourceMetrics() {
				for _, scope := range r.GetScopeMetrics() {
					for _, metric := range scope.GetMetrics() {
						for _, point := range metric.GetSum().GetDataPoints() {
							if len(point.GetAttributes()) != 0 ||
								len(point.GetExemplars()) != 0 ||
								point.GetAsInt() != 1 {
								t.Fatal("collector retained unsafe metric identity or lost its value")
							}
						}
					}
				}
			}
		}
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

//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
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
		response, err722 := client.Get(fmt.Sprintf("http://%s/visitor/%d?token=unique%d", r.Address(), i, i))
		if err722 != nil {
			t.Fatal("varied HTTP request failed")
		}
		_ = response.Body.Close()
	}
	closeRole()
	proofEventually(t, func() bool {
		for _, w := range post.snapshot() {
			if w.path == "/v1/metrics" {
				var req metriccollector.ExportMetricsServiceRequest
				if proto.Unmarshal(w.body, &req) != nil {
					t.Fatal("decode final HTTP metrics")
				}
				for _, r := range req.GetResourceMetrics() {
					for _, s := range r.GetScopeMetrics() {
						for _, m := range s.GetMetrics() {
							if m.GetName() == "flux.http.requests" {
								for _, p := range m.GetSum().GetDataPoints() {
									if p.GetAsInt() == 200 {
										return true
									}
								}
							}
						}
					}
				}
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
		for _, r := range req.GetResourceMetrics() {
			for _, s := range r.GetScopeMetrics() {
				for _, m := range s.GetMetrics() {
					if m.GetName() != "flux.http.requests" {
						continue
					}
					for _, p := range m.GetSum().GetDataPoints() {
						sort.Slice(p.GetAttributes(),
							func(i,
								j int) bool {
								return p.GetAttributes()[i].GetKey() < p.GetAttributes()[j].GetKey()
							})
						data, _ := json.Marshal(p.GetAttributes())
						series[string(data)] = true
						count = max(count, p.GetAsInt())
						for _, a := range p.GetAttributes() {
							if a.GetKey() == "http.route" &&
								a.GetValue().
									GetStringValue() != "unmatched" {
								t.Fatal("metric route cardinality expanded")
							}
							if a.GetKey() == "request_id" ||
								a.GetKey() == "correlation_id" {
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

//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
func TestTelemetryOutageReadinessAndCleanup(t *testing.T) {
	pg, closePG := backendtesting.SetupTestPostgres(t)
	defer closePG()
	queue, closeRedis := backendtesting.SetupTestRedis(t)
	defer closeRedis()
	for _, mode := range []string{"disconnected", "slow"} {
		t.Run(mode, func(t *testing.T) {
			endpoint := "http://" + binaryTestAddress(t)
			if mode == "slow" {
				s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
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
					f.telemetry = func(ctx context.Context,
						cfg *config.Config,
						role config.Role) (*observability.Telemetry,
						func(context.Context) error,
						error) {
						owner, cleanup, err := baseTelemetry(ctx, cfg, role)
						return owner,
							func(ctx context.Context) error { closed["telemetry"] = true; return cleanup(ctx) },
							err
					}
					baseDB := f.database
					f.database = func(ctx context.Context,
						s *server.Server) (*database.Database,
						func(context.Context) error,
						error) {
						db, cleanup, err := baseDB(ctx, s)
						return db,
							func(ctx context.Context) error { closed["database"] = true; return cleanup(ctx) },
							err
					}
					baseRedis := f.redis
					f.redis = func(ctx context.Context,
						s *server.Server) (*redis.Client,
						func(context.Context) error,
						error) {
						db, cleanup, err := baseRedis(ctx, s)
						return db,
							func(ctx context.Context) error { closed["redis"] = true; return cleanup(ctx) },
							err
					}
					baseProducer := f.producer
					f.producer = func(s *server.Server) (*job.JobService,
						func(context.Context) error,
						error) {
						j, cleanup, err := baseProducer(s)
						return j,
							func(ctx context.Context) error { closed["producer"] = true; return cleanup(ctx) },
							err
					}
					baseConsumer := f.consumer
					f.consumer = func(s *server.Server,
						e *email.Client) (*job.JobService,
						func(context.Context) error,
						error) {
						j, cleanup, err := baseConsumer(s, e)
						return j,
							func(ctx context.Context) error { closed["consumer"] = true; return cleanup(ctx) },
							err
					}
					r, err := newRole(context.Background(), role, cfg, f)
					if err != nil {
						t.Fatal("collector outage prevented role construction")
					}
					rec := httptest.NewRecorder()
					r.HTTP.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
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
					for _, name := range []string{"telemetry",
						"database",
						"redis",
						"producer",
						"consumer"} {
						want := name == "telemetry" ||
							(role == config.RoleAPI &&
								(name == "database" ||
									name == "redis" ||
									name == "producer")) ||
							(role == config.RoleWorker &&
								(name == "redis" ||
									name == "consumer"))
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

// The migrator has no readiness listener: its required operation is the actual
// one-shot PostgreSQL migration. Monitoring failures may report an exit failure
// after a successful migration, but cannot block that required operation.
//
//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
func TestTelemetryOutageMigrator(t *testing.T) {
	pg, closePG := backendtesting.SetupTestPostgres(t)
	defer closePG()
	binary := filepath.Join(t.TempDir(), "migrator")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/migrator")
	build.Dir = "../.."
	if output, err := build.CombinedOutput(); err != nil {
		_ = output
		t.Fatal("build migrator outage binary")
	}
	for _, mode := range []string{"disconnected", "slow"} {
		t.Run(mode, func(t *testing.T) {
			endpoint := "http://" + binaryTestAddress(t)
			if mode == "slow" {
				collector := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter,
					r *http.Request) {
					select {
					case <-r.Context().Done():
					case <-time.After(time.Second):
					}
				}))
				defer collector.Close()
				endpoint = collector.URL
			}
			env := append(binaryTestEnv(), binaryDatabaseEnv(pg.Config.Database)...)
			env = append(env,
				"FLUX_OBSERVABILITY.OTLP.ENABLED=true",
				"FLUX_OBSERVABILITY.OTLP.ENDPOINT="+endpoint,
				"FLUX_OBSERVABILITY.OTLP.EXPORT_TIMEOUT=80ms",
				"FLUX_OBSERVABILITY.OTLP.EXPORT_INTERVAL=10ms",
				"FLUX_OBSERVABILITY.HEALTH_CHECKS.TIMEOUT=1s")
			ctx909, cancel1160 := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel1160()
			cmd := exec.CommandContext(ctx909, binary)
			cmd.Dir = t.TempDir()
			cmd.Env = env
			start := time.Now()
			output, err := cmd.CombinedOutput()
			var exit *exec.ExitError
			if err != nil && (!errors.As(err, &exit) || exit.ExitCode() != 1) {
				t.Fatal("migrator outage prevented bounded execution")
			}
			if ctx909.Err() != nil || time.Since(start) > 1500*time.Millisecond {
				t.Fatal("migrator outage exceeded overall exit budget")
			}
			if !bytes.Contains(output,
				[]byte(`"outcome":"success"`)) ||
				!bytes.Contains(output,
					[]byte(`"end_version":1`)) {
				t.Fatal("collector outage prevented required PostgreSQL migration")
			}
			var version int
			if pg.Pool.QueryRow(context.Background(),
				"SELECT version FROM schema_version").
				Scan(&version) != nil ||
				version != 1 {
				t.Fatal("migrator outage lost authoritative PostgreSQL result")
			}
		})
	}
}

type proofFaultTrace struct {
	shutdown *atomic.Int32
	deadline *atomic.Int32
}

func (e proofFaultTrace) ExportSpans(ctx context.Context, _ []sdktrace.ReadOnlySpan) error {
	if _, ok := ctx.Deadline(); ok {
		e.deadline.Add(1)
	}
	<-ctx.Done()
	return errors.New("fault-provider-private")
}
func (e proofFaultTrace) Shutdown(context.Context) error {
	e.shutdown.Add(1)
	return errors.New("fault-provider-private")
}

type proofFaultLog struct {
	shutdown *atomic.Int32
	deadline *atomic.Int32
}

func (e proofFaultLog) Export(ctx context.Context, _ []sdklog.Record) error {
	if _, ok := ctx.Deadline(); ok {
		e.deadline.Add(1)
	}
	<-ctx.Done()
	return errors.New("fault-provider-private")
}
func (e proofFaultLog) Shutdown(context.Context) error {
	e.shutdown.Add(1)
	return errors.New("fault-provider-private")
}
func (proofFaultLog) ForceFlush(context.Context) error { return nil }

type proofFaultMetric struct {
	shutdown *atomic.Int32
	deadline *atomic.Int32
}

func (proofFaultMetric) Temporality(sdkmetric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}
func (proofFaultMetric) Aggregation(k sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return sdkmetric.DefaultAggregationSelector(k)
}
func (e proofFaultMetric) Export(ctx context.Context, _ *metricdata.ResourceMetrics) error {
	if _, ok := ctx.Deadline(); ok {
		e.deadline.Add(1)
	}
	<-ctx.Done()
	return errors.New("fault-provider-private")
}
func (e proofFaultMetric) Shutdown(context.Context) error {
	e.shutdown.Add(1)
	return errors.New("fault-provider-private")
}
func (proofFaultMetric) ForceFlush(context.Context) error { return nil }

func TestObservabilityFaultExportersBoundedCleanup(t *testing.T) {
	var shutdown, deadline atomic.Int32
	owner,
		err := observability.New(context.Background(),
		observability.Settings{Enabled: true,
			SampleRatio:    1,
			ExportTimeout:  80 * time.Millisecond,
			ExportInterval: time.Minute,
			Exporters: observability.Exporters{
				Trace: proofFaultTrace{shutdown: &shutdown,
					deadline: &deadline},
				Metric: proofFaultMetric{shutdown: &shutdown,
					deadline: &deadline},
				Log: proofFaultLog{shutdown: &shutdown,
					deadline: &deadline},
			}}, "api")
	if err != nil {
		t.Fatal("construct fault exporters")
	}
	_, span := owner.Tracer.Start(context.Background(), "http.request")
	span.End()
	metric, err := owner.Meter.Int64Counter("flux.operations")
	if err != nil {
		t.Fatal("create bounded fault metric")
	}
	metric.Add(context.Background(), 1)
	log := loggerpkg.NewLogger(config.DefaultObservabilityConfig(), io.Discard, owner.Logger)
	log.Info().Msg("http.request")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	err = owner.Shutdown(ctx)
	if err == nil || time.Since(start) > 300*time.Millisecond {
		t.Fatal("fault exporter shutdown did not report bounded failure")
	}
	proofClean(t, []byte(err.Error()), "fault-provider-private")
	proofEventually(t, func() bool { return shutdown.Load() == 3 })
	if deadline.Load() != 3 {
		t.Fatal("every signal export must receive a deadline")
	}
}
