package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/app"
	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/handler"
	"github.com/6sLOGAN78/flux/internal/router"
	"github.com/6sLOGAN78/flux/internal/server"
	backendTesting "github.com/6sLOGAN78/flux/internal/testing"
	"github.com/6sLOGAN78/flux/internal/transport"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func healthRouter(log *zerolog.Logger, timeout time.Duration, checks ...handler.ReadinessCheck) *echo.Echo {
	r := echo.New()
	router.RegisterHealthRoutes(r, handler.NewReadinessHandler(log, timeout, checks))
	return r
}

func healthRequest(r http.Handler, ctx context.Context, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil).WithContext(ctx))
	return rec
}

func TestLiveDoesNotProbeDependencies(t *testing.T) {
	var calls atomic.Int32
	r := healthRouter(nil, time.Second, handler.ReadinessCheck{Name: "database", Check: func(context.Context) error {
		calls.Add(1)
		return errors.New("offline")
	}})
	assertHealthContract(t, healthRequest(r, context.Background(), "/live"), "/live", 200, `{"status":"alive"}`)
	if calls.Load() != 0 {
		t.Fatal("liveness probed dependencies")
	}
}

func TestReadySanitizedFailuresAndSingleLog(t *testing.T) {
	var logs bytes.Buffer
	log := zerolog.New(&logs)
	secret := "postgres://user:secret@private-host provider-payload STACK-MARKER"
	r := healthRouter(&log, time.Second,
		handler.ReadinessCheck{Name: "database", Check: func(context.Context) error { return errors.New(secret) }},
		handler.ReadinessCheck{Name: "redis", Check: func(context.Context) error { return nil }},
	)
	rec := healthRequest(r, context.Background(), "/ready")
	assertHealthContract(t, rec, "/ready", 503, `{"status":"not_ready","checks":[{"name":"database","state":"not_ready"},{"name":"redis","state":"ready"}]}`)
	for _, marker := range strings.Fields(secret) {
		if strings.Contains(rec.Body.String(), marker) || strings.Contains(logs.String(), marker) {
			t.Fatal("health diagnostic leaked private failure")
		}
	}
	if strings.Count(logs.String(), "readiness check failed") != 1 || !strings.Contains(logs.String(), `"failure_kind":"dependency_error"`) {
		t.Fatalf("expected one classified failure log: %s", logs.String())
	}
}

func TestReadyConcurrentOneRequestDeadline(t *testing.T) {
	const budget = 150 * time.Millisecond
	type observation struct {
		deadline time.Time
		value    any
	}
	observed := make(chan observation, 2)
	started := make(chan struct{}, 2)
	check := func(ctx context.Context) error {
		deadline, _ := ctx.Deadline()
		observed <- observation{deadline, ctx.Value("health-test")}
		started <- struct{}{}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			return errors.New("missing cancellation")
		}
	}
	r := healthRouter(nil, budget, handler.ReadinessCheck{Name: "database", Check: check}, handler.ReadinessCheck{Name: "redis", Check: check})
	ctx := context.WithValue(context.Background(), "health-test", "request-value")
	begin := time.Now()
	rec := healthRequest(r, ctx, "/ready")
	if elapsed := time.Since(begin); elapsed > 260*time.Millisecond {
		t.Fatalf("multiplied deadline: %s", elapsed)
	}
	assertHealthContract(t, rec, "/ready", 503, `{"status":"not_ready","checks":[{"name":"database","state":"not_ready"},{"name":"redis","state":"not_ready"}]}`)
	a, b := <-observed, <-observed
	if !a.deadline.Equal(b.deadline) || a.value != "request-value" || b.value != "request-value" || len(started) != 2 {
		t.Fatal("checks did not share request-derived context")
	}
}

func TestReadyRequestCancellation(t *testing.T) {
	started := make(chan struct{})
	ended := make(chan struct{})
	r := healthRouter(nil, time.Second, handler.ReadinessCheck{Name: "redis", Check: func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		close(ended)
		return ctx.Err()
	}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { <-started; cancel() }()
	begin := time.Now()
	assertHealthContract(t, healthRequest(r, ctx, "/ready"), "/ready", 503, `{"status":"not_ready","checks":[{"name":"redis","state":"not_ready"}]}`)
	if time.Since(begin) > 200*time.Millisecond {
		t.Fatal("request cancellation ignored")
	}
	select {
	case <-ended:
	case <-time.After(time.Second):
		t.Fatal("check did not observe cancellation")
	}
}

func TestReadyBypassesCustomerLimiter(t *testing.T) {
	log := zerolog.Nop()
	cfg := healthRoleConfig()
	srv := &server.Server{Config: cfg, Logger: &log}
	h := &handler.Handlers{Health: handler.NewReadinessHandler(&log, time.Second, nil), OpenAPI: handler.NewOpenAPIHandler(srv)}
	r := router.NewRouter(srv, h, nil)
	limited := false
	for i := 0; i < 60; i++ {
		if healthRequest(r, context.Background(), "/docs").Code == 429 {
			limited = true
		}
	}
	if !limited {
		t.Fatal("test did not exhaust customer limiter")
	}
	for i := 0; i < 60; i++ {
		assertHealthContract(t, healthRequest(r, context.Background(), "/live"), "/live", 200, `{"status":"alive"}`)
		assertHealthContract(t, healthRequest(r, context.Background(), "/ready"), "/ready", 200, `{"status":"ready","checks":[]}`)
	}
	if healthRequest(r, context.Background(), "/status").Code == 200 {
		t.Fatal("legacy status still public")
	}
}

func healthRoleConfig() *config.Config {
	settings := config.RoleConfig{ListenAddress: "127.0.0.1:0", DrainTimeout: time.Second, ReadinessTimeout: 200 * time.Millisecond}
	return &config.Config{Primary: config.Primary{Env: "test"}, Observability: config.DefaultObservabilityConfig(), API: settings, Redirector: settings, Worker: settings}
}

func TestRoleHealthActualHTTP(t *testing.T) {
	db, _ := backendTesting.SetupTestPostgres(t)
	queue, _ := backendTesting.SetupTestRedis(t)
	var providerCalls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { providerCalls.Add(1); w.WriteHeader(500) }))
	defer provider.Close()
	t.Setenv("RESEND_BASE_URL", provider.URL)
	for _, role := range []string{"api", "api-producer", "redirector", "worker"} {
		t.Run(role, func(t *testing.T) {
			cfg := healthRoleConfig()
			cfg.Database = db.Config.Database
			cfg.Redis = queue.Config
			cfg.Integration.ResendAPIKey = "provider-secret"
			var runtime *app.RoleRuntime
			switch role {
			case "api", "api-producer":
				cfg.API.ProducerEnabled = role == "api-producer"
				instance, err := app.NewAPI(context.Background(), cfg)
				if err != nil {
					t.Fatal(err)
				}
				runtime = instance.RoleRuntime
			case "redirector":
				instance, err := app.NewRedirector(context.Background(), cfg)
				if err != nil {
					t.Fatal(err)
				}
				runtime = instance.RoleRuntime
			case "worker":
				instance, err := app.NewWorker(context.Background(), cfg)
				if err != nil {
					t.Fatal(err)
				}
				runtime = instance.RoleRuntime
			}
			ctx, cancel := context.WithCancel(context.Background())
			finished := make(chan error, 1)
			go func() { finished <- runtime.Run(ctx) }()
			t.Cleanup(func() {
				cancel()
				select {
				case err := <-finished:
					if err != nil {
						t.Error(err)
					}
				case <-time.After(5 * time.Second):
					t.Error("role did not stop")
				}
			})
			request := func(path string) *httptest.ResponseRecorder {
				client := http.Client{Timeout: time.Second}
				res, err := client.Get("http://" + runtime.Address() + path)
				if err != nil {
					t.Fatal(err)
				}
				defer res.Body.Close()
				rec := httptest.NewRecorder()
				rec.Code = res.StatusCode
				rec.HeaderMap = res.Header
				if _, err := rec.Body.ReadFrom(res.Body); err != nil {
					t.Fatal(err)
				}
				return rec
			}
			assertHealthContract(t, request("/live"), "/live", 200, `{"status":"alive"}`)
			var fixture string
			switch role {
			case "api":
				fixture = `{"status":"ready","checks":[{"name":"database","state":"ready"}]}`
			case "api-producer":
				fixture = `{"status":"ready","checks":[{"name":"database","state":"ready"},{"name":"redis","state":"ready"}]}`
			case "redirector":
				fixture = `{"status":"ready","checks":[]}`
			case "worker":
				fixture = `{"status":"ready","checks":[{"name":"redis","state":"ready"},{"name":"email","state":"ready"}]}`
			}
			assertHealthContract(t, request("/ready"), "/ready", 200, fixture)
			if role == "api-producer" {
				activeQueue := runtime.Server.Redis
				failedQueue := redis.NewClient(&redis.Options{Addr: queue.Config.Address})
				if err := failedQueue.Close(); err != nil {
					t.Fatal(err)
				}
				runtime.Server.Redis = failedQueue
				t.Cleanup(func() { runtime.Server.Redis = activeQueue })
				fixture = `{"status":"not_ready","checks":[{"name":"database","state":"ready"},{"name":"redis","state":"not_ready"}]}`
				assertHealthContract(t, request("/ready"), "/ready", 503, fixture)
			}
			if role == "worker" {
				cfg.Integration.ResendAPIKey = ""
				assertHealthContract(t, request("/ready"), "/ready", 503, `{"status":"not_ready","checks":[{"name":"redis","state":"ready"},{"name":"email","state":"not_ready"}]}`)
				cfg.Integration.ResendAPIKey = "provider-secret"
			}
			if role == "api" || role == "api-producer" {
				runtime.Server.DB.Pool.Close()
				fixture = strings.Replace(fixture, `"status":"ready"`, `"status":"not_ready"`, 1)
				fixture = strings.Replace(fixture, `"name":"database","state":"ready"`, `"name":"database","state":"not_ready"`, 1)
			}
			if role == "worker" {
				// Inject probe failure without closing the active consumer's shared
				// broker before its ordered shutdown (an upstream Asynq invariant).
				activeQueue := runtime.Server.Redis
				failedQueue := redis.NewClient(&redis.Options{Addr: queue.Config.Address})
				if err := failedQueue.Close(); err != nil {
					t.Fatal(err)
				}
				runtime.Server.Redis = failedQueue
				t.Cleanup(func() { runtime.Server.Redis = activeQueue })
				fixture = `{"status":"not_ready","checks":[{"name":"redis","state":"not_ready"},{"name":"email","state":"ready"}]}`
			}
			if role != "redirector" {
				assertHealthContract(t, request("/ready"), "/ready", 503, fixture)
				assertHealthContract(t, request("/live"), "/live", 200, `{"status":"alive"}`)
			}
			if role == "worker" || role == "redirector" {
				for _, path := range []string{"/docs", "/api/v1/links", "/static/openapi.json", "/status"} {
					if request(path).Code != 404 {
						t.Fatalf("management role exposed %s", path)
					}
				}
			}
		})
	}
	if providerCalls.Load() != 0 {
		t.Fatal("readiness contacted email provider")
	}
}

func assertHealthContract(t *testing.T, rec *httptest.ResponseRecorder, path string, status int, fixture string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("%s status=%d body=%s, want %d", path, rec.Code, rec.Body.String(), status)
	}
	if !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatal("non-JSON response")
	}
	var actual, want any
	if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(fixture), &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("%s body=%s want=%s", path, rec.Body.String(), fixture)
	}
	var generated any = &transport.HealthReadyResponse{}
	name := "HealthReadyResponse"
	if path == "/live" {
		generated = &transport.HealthLiveResponse{}
		name = "HealthLiveResponse"
	}
	if err := json.Unmarshal(rec.Body.Bytes(), generated); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(generated)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip any
	if err := json.Unmarshal(data, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(roundTrip, actual) {
		t.Fatal("payload differs from generated transport")
	}
	data, err = os.ReadFile("../../../../packages/openapi/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	operation := doc["paths"].(map[string]any)[path].(map[string]any)["get"].(map[string]any)
	response := operation["responses"].(map[string]any)[strconv.Itoa(status)].(map[string]any)
	ref := response["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)["$ref"]
	if ref != "#/components/schemas/transport."+name {
		t.Fatalf("unexpected canonical response ref %v", ref)
	}
	schema := doc["components"].(map[string]any)["schemas"].(map[string]any)["transport."+name].(map[string]any)
	if err := validateHealthSchema(schema, actual); err != nil {
		t.Fatal(err)
	}
}

func validateHealthSchema(schema map[string]any, value any) error {
	switch schema["type"] {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string")
		}
		if values, ok := schema["enum"].([]any); ok {
			for _, allowed := range values {
				if allowed == value {
					return nil
				}
			}
			return fmt.Errorf("outside canonical enum")
		}
	case "object":
		object, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("expected object")
		}
		properties := schema["properties"].(map[string]any)
		for _, name := range schema["required"].([]any) {
			if _, ok := object[name.(string)]; !ok {
				return fmt.Errorf("missing %s", name)
			}
		}
		for name, field := range object {
			property, ok := properties[name]
			if !ok {
				return fmt.Errorf("undocumented %s", name)
			}
			if err := validateHealthSchema(property.(map[string]any), field); err != nil {
				return err
			}
		}
	case "array":
		items, ok := value.([]any)
		if !ok {
			return fmt.Errorf("expected array")
		}
		for _, item := range items {
			if err := validateHealthSchema(schema["items"].(map[string]any), item); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported canonical schema type")
	}
	return nil
}
