//nolint:testpackage // These tests verify package-private lifecycle and failure-injection seams.
package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/lib/email"
	"github.com/6sLOGAN78/flux/internal/lib/job"
	loggerPkg "github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/6sLOGAN78/flux/internal/server"
	backendTesting "github.com/6sLOGAN78/flux/internal/testing"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func roleTestConfig() *config.Config {
	return &config.Config{Primary: config.Primary{Env: "test"}, Observability: config.DefaultObservabilityConfig(),
		API: config.RoleConfig{ListenAddress: "127.0.0.1:0",
			DrainTimeout:     time.Second,
			ReadinessTimeout: time.Second},
		Redirector: config.RoleConfig{ListenAddress: "127.0.0.1:0",
			DrainTimeout:     time.Second,
			ReadinessTimeout: time.Second},
		Worker: config.RoleConfig{ListenAddress: "127.0.0.1:0",
			DrainTimeout:     time.Second,
			ReadinessTimeout: time.Second},
	}
}

type roleTraceCapture struct{ *tracetest.InMemoryExporter }

func (e *roleTraceCapture) Shutdown(context.Context) error { return nil }

type roleLogCapture struct {
	records []sdklog.Record
	mu      sync.Mutex
}

func (e *roleLogCapture) Export(_ context.Context, records []sdklog.Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, r := range records {
		e.records = append(e.records, r.Clone())
	}
	return nil
}
func (*roleLogCapture) Shutdown(context.Context) error   { return nil }
func (*roleLogCapture) ForceFlush(context.Context) error { return nil }

//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
func TestRoleTelemetryExportsAfterWorkAndDependencyClose(t *testing.T) {
	for _, role := range []config.Role{config.RoleAPI, config.RoleRedirector, config.RoleWorker} {
		t.Run(string(role), func(t *testing.T) {
			var opened, closed []string
			f := roleSpies("", &opened, &closed, nil)
			var output bytes.Buffer
			traces := &roleTraceCapture{tracetest.NewInMemoryExporter()}
			logs := &roleLogCapture{}
			var owner *observability.Telemetry
			f.telemetry = func(ctx context.Context,
				cfg *config.Config,
				gotRole config.Role) (*observability.Telemetry,
				func(context.Context) error,
				error) {
				s := cfg.Observability.TelemetrySettings()
				s.Enabled = true
				s.ExportInterval = time.Minute
				s.Exporters = observability.Exporters{Trace: traces, Log: logs}
				var err error
				owner, err = observability.New(ctx, s, string(gotRole))
				return owner, func(ctx context.Context) error {
					if role != config.RoleRedirector && len(closed) == 0 {
						t.Error("provider closed before dependencies")
					}
					return owner.Shutdown(ctx)
				}, err
			}
			f.logger = func(cfg *config.Config,
				tel *observability.Telemetry) (*zerolog.Logger,
				func(context.Context) error,
				error) {
				log := loggerPkg.NewLogger(cfg.Observability, &output, tel.Logger)
				return &log, nil, nil
			}
			baseDB := f.database
			f.database = func(ctx context.Context,
				srv *server.Server) (*database.Database,
				func(context.Context) error,
				error) {
				if srv.Telemetry != owner {
					t.Error("database did not receive role owner")
				}
				db, closer, err := baseDB(ctx, srv)
				return db, func(ctx context.Context) error {
					_, span := srv.Telemetry.Tracer.Start(ctx, "database.query")
					span.End()
					return closer(ctx)
				}, err
			}
			r, err := newRole(context.Background(), role, roleTestConfig(), f)
			if err != nil {
				t.Fatal(err)
			}
			if r.Server.Telemetry != owner {
				t.Fatal("server owner differs")
			}
			req := httptest.NewRequest(http.MethodGet, "/live?token=SECRET-MARKER", nil)
			req.Header.Set("Authorization", "SECRET-MARKER")
			r.HTTP.ServeHTTP(httptest.NewRecorder(), req)
			if err116 := r.Close(context.Background()); err116 != nil {
				t.Fatal(err116)
			}
			spans := traces.GetSpans()
			wantSpans := 1
			if role == config.RoleAPI {
				wantSpans++
			}
			if len(spans) != wantSpans {
				t.Fatalf("flush exported %d spans, want %d", len(spans), wantSpans)
			}
			for _, span := range spans {
				if strings.Contains(fmt.Sprint(span), "SECRET-MARKER") {
					t.Fatal("span leaked request")
				}
			}
			logs.mu.Lock()
			defer logs.mu.Unlock()
			if len(logs.records) != 1 || logs.records[0].TraceID() != spans[0].SpanContext.TraceID() {
				t.Fatal("role logger was not injected/correlated")
			}
			if strings.Contains(output.String(), "SECRET-MARKER") {
				t.Fatal("stdout leaked request")
			}
		})
	}
}

func roleSpies(fail string, opened, closed *[]string, cause error) roleFactories {
	stage := func(name string) (func(context.Context) error, error) {
		if name == fail {
			return nil, cause
		}
		*opened = append(*opened, name)
		return func(context.Context) error { *closed = append(*closed, name); return cause }, nil
	}
	return roleFactories{
		telemetry: func(ctx context.Context,
			cfg *config.Config,
			role config.Role) (*observability.Telemetry,
			func(context.Context) error,
			error) {
			cleanup, err := stage("telemetry")
			if err != nil {
				return nil, cleanup, err
			}
			owner, err := observability.New(ctx, cfg.Observability.TelemetrySettings(), string(role))
			return owner,
				func(ctx context.Context) error {
					return errors.Join(cleanup(ctx),
						owner.Shutdown(ctx))
				},
				err
		},
		logger: func(*config.Config,
			*observability.Telemetry) (*zerolog.Logger,
			func(context.Context) error,
			error) {
			cleanup, err := stage("logger")
			log := zerolog.Nop()
			return &log, cleanup, err
		},
		database: func(context.Context,
			*server.Server) (*database.Database,
			func(context.Context) error,
			error) {
			cleanup, err := stage("database")
			return &database.Database{}, cleanup, err
		},
		redis: func(context.Context, *server.Server) (*redis.Client, func(context.Context) error, error) {
			cleanup, err := stage("redis")
			return nil, cleanup, err
		},
		producer: func(*server.Server) (*job.JobService, func(context.Context) error, error) {
			cleanup, err := stage("producer")
			return nil, cleanup, err
		},
		email: func(*server.Server) (*email.Client, error) {
			if fail == "email" {
				return nil, cause
			}
			*opened = append(*opened, "email")
			return nil, nil
		},
		consumer: func(*server.Server, *email.Client) (*job.JobService, func(context.Context) error, error) {
			cleanup, err := stage("consumer")
			return nil, cleanup, err
		},
		startConsumer: func(*job.JobService) error {
			if fail == "start consumer" {
				return cause
			}
			*opened = append(*opened, "start consumer")
			return nil
		},
		router: func(role config.Role, srv *server.Server) (*echo.Echo, error) {
			if fail == "router" {
				return nil, cause
			}
			*opened = append(*opened, "router")
			return defaultRoleRouter(role, srv)
		},
		listen: func(address string) (net.Listener, error) {
			if fail == "listener" {
				return nil, cause
			}
			*opened = append(*opened, "listener")
			return net.Listen("tcp", address)
		},
	}
}

//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
func TestRoleResourceGraphs(t *testing.T) {
	for _, test := range []struct {
		role     config.Role
		want     []string
		producer bool
	}{
		{role: config.RoleAPI,
			producer: false,
			want: []string{"telemetry",
				"logger",
				"database",
				"router",
				"listener"}},
		{role: config.RoleAPI,
			producer: true,
			want: []string{"telemetry",
				"logger",
				"database",
				"redis",
				"producer",
				"router",
				"listener"}},
		{role: config.RoleRedirector,
			producer: false,
			want: []string{"telemetry",
				"logger",
				"router",
				"listener"}},
		{role: config.RoleWorker,
			producer: false,
			want: []string{"telemetry",
				"logger",
				"redis",
				"email",
				"consumer",
				"router",
				"listener",
				"start consumer"}},
	} {
		t.Run(string(test.role)+string(rune('0'+boolInt(test.producer))), func(t *testing.T) {
			cfg := roleTestConfig()
			cfg.API.ProducerEnabled = test.producer
			var opened, closed []string
			role, err := newRole(context.Background(), test.role, cfg, roleSpies("", &opened, &closed, nil))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(opened, test.want) {
				t.Fatalf("role graph %v, want %v", opened, test.want)
			}
			if role.Server.DB != nil && test.role != config.RoleAPI {
				t.Fatal("non-API allocated PostgreSQL")
			}
			if test.role != config.RoleAPI {
				for _, path := range []string{"/docs", "/static/openapi.json", "/api/v1/links"} {
					recorder := httptest.NewRecorder()
					role.HTTP.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
					if recorder.Code != 404 {
						t.Fatalf("%s exposed %s: %d", test.role, path, recorder.Code)
					}
				}
			}
			if err247 := role.Close(context.Background()); err247 != nil {
				t.Fatal(err247)
			}
			wantClosed := []string{"logger", "telemetry"}
			if test.role == config.RoleAPI {
				wantClosed = []string{"database", "logger", "telemetry"}
				if test.producer {
					wantClosed = []string{"producer", "redis", "database", "logger", "telemetry"}
				}
			}
			if test.role == config.RoleWorker {
				wantClosed = []string{"consumer", "redis", "logger", "telemetry"}
			}
			if !reflect.DeepEqual(closed, wantClosed) {
				t.Fatalf("close order %v, want %v", closed, wantClosed)
			}
			if err263 := role.Close(context.Background()); err263 != nil ||
				!reflect.DeepEqual(closed,
					wantClosed) {
				t.Fatal("repeated close changed effects")
			}
		})
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
func TestRoleConstructionFailureUnwinds(t *testing.T) {
	cause := errors.New("SECRET-MARKER")
	for _, role := range []config.Role{config.RoleAPI, config.RoleRedirector, config.RoleWorker} {
		stages := []string{"telemetry", "logger", "router", "listener"}
		if role == config.RoleAPI {
			stages = []string{"telemetry", "logger", "database", "redis", "producer", "router", "listener"}
		}
		if role == config.RoleWorker {
			stages = []string{"telemetry",
				"logger",
				"redis",
				"email",
				"consumer",
				"router",
				"listener",
				"start consumer"}
		}
		for _, fail := range stages {
			t.Run(string(role)+"/"+fail, func(t *testing.T) {
				cfg := roleTestConfig()
				cfg.API.ProducerEnabled = true
				var opened, closed []string
				runtime,
					err := newRole(context.Background(),
					role,
					cfg,
					roleSpies(fail,
						&opened,
						&closed,
						cause))
				if runtime != nil ||
					!errors.Is(err,
						cause) ||
					strings.Contains(err.Error(),
						"SECRET-MARKER") {
					t.Fatalf("unsafe or lost failure: %v", err)
				}
				var want []string
				for i := len(opened) - 1; i >= 0; i-- {
					switch opened[i] {
					case "telemetry", "logger", "database", "redis", "producer", "consumer":
						want = append(want, opened[i])
					}
				}
				if !reflect.DeepEqual(closed, want) {
					t.Fatalf("failed %s: closed %v, want %v", fail, closed, want)
				}
			})
		}
	}
}

func TestRolePartialProviderFailureStillReleasesOwner(t *testing.T) {
	for _, role := range []config.Role{config.RoleAPI, config.RoleRedirector, config.RoleWorker} {
		var opened, closed []string
		f := roleSpies("", &opened, &closed, nil)
		cause := errors.New("SECRET-MARKER provider")
		f.telemetry = func(context.Context,
			*config.Config,
			config.Role) (*observability.Telemetry,
			func(context.Context) error,
			error) {
			return nil,
				func(context.Context) error {
					closed = append(closed,
						"provider")
					return cause
				},
				cause
		}
		r, err := newRole(context.Background(), role, roleTestConfig(), f)
		if r != nil ||
			!errors.Is(err,
				cause) ||
			strings.Contains(err.Error(),
				"SECRET-MARKER") ||
			!reflect.DeepEqual(closed,
				[]string{"provider"}) {
			t.Fatalf("%s partial provider owner leaked: %v %v", role, err, closed)
		}
	}
}

func TestAPIMigratorSpecialCredentials(t *testing.T) {
	ctx := context.Background()
	pg, closePG := backendTesting.SetupTestPostgres(t)
	defer closePG()
	// Disposable literal fixtures cover every URL component without diagnostics.
	if _, err := pg.Pool.Exec(ctx,
		`CREATE ROLE "synthetic user@/?:+" LOGIN SUPERUSER PASSWORD 'synthetic password @:/?+%#'`); err != nil {
		t.Fatal("create special-character role")
	}
	if _, err := pg.Pool.Exec(ctx, `CREATE DATABASE "synthetic db@/?:+" OWNER "synthetic user@/?:+"`); err != nil {
		t.Fatal("create special-character database")
	}
	cfg := roleTestConfig()
	cfg.Server = pg.Config.Server
	cfg.Database = pg.Config.Database
	cfg.Database.User = "synthetic user@/?:+"
	cfg.Database.Password = "synthetic password @:/?+%#"
	cfg.Database.Name = "synthetic db@/?:+"
	if _, err := database.MigrateWithResult(ctx, cfg); err != nil {
		t.Fatal("migrator rejected accepted credentials")
	}
	api, err := NewAPI(ctx, cfg)
	if err != nil {
		t.Fatal("API rejected migrator credentials")
	}
	defer api.Close(ctx)
	var user, name string
	if err = api.Server.DB.Pool.QueryRow(ctx, "SELECT current_user, current_database()").Scan(&user, &name); err != nil {
		t.Fatal("query API connection identity")
	}
	if user != cfg.Database.User || name != cfg.Database.Name {
		t.Fatal("API changed connection identity")
	}
}

func TestRoleServerReceivesOnlyExplicitResources(t *testing.T) {
	log := zerolog.Nop()
	srv, err := server.New(roleTestConfig(), &log, nil)
	if err != nil || srv.DB != nil || srv.Redis != nil || srv.Job != nil {
		t.Fatalf("server allocated undeclared dependencies: %v", err)
	}
}

//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
func TestRoleRealDependenciesRemainIndependent(t *testing.T) {
	ctx := context.Background()
	pg, closePG := backendTesting.SetupTestPostgres(t)
	defer closePG()
	queue, closeQueue := backendTesting.SetupTestRedis(t)
	defer closeQueue()
	cfg := roleTestConfig()
	cfg.Database = pg.Config.Database
	cfg.Server = pg.Config.Server
	api, err := NewAPI(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err348 := api.Close(ctx); err348 != nil {
			t.Error(err348)
		}
	})
	if api.Server.Redis != nil || api.Server.Job != nil {
		t.Fatal("default API owns queue infrastructure")
	}
	var ledger *string
	if err356 := api.Server.DB.Pool.QueryRow(ctx,
		"SELECT to_regclass('schema_version')::text").
		Scan(&ledger); err356 != nil ||
		ledger != nil {
		t.Fatalf("API migrated empty PostgreSQL: %v, %v", ledger, err356)
	}

	cfg = roleTestConfig()
	cfg.Server = pg.Config.Server
	cfg.Redis = queue.Config
	cfg.Integration.ResendAPIKey = "test-key"
	worker, err := NewWorker(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if worker.Server.DB != nil || worker.Server.Job.Client != nil {
		t.Fatal("worker owns PostgreSQL or producer")
	}
	if err371 := worker.Close(ctx); err371 != nil {
		t.Fatal(err371)
	}
	if err374 := worker.Server.Redis.Ping(ctx).Err(); err374 == nil {
		t.Fatal("worker leaked shared Redis client")
	}
	assertRoleListenerReleased(t, worker.Address())

	cfg = roleTestConfig()
	cfg.Server = pg.Config.Server
	redirector, err := NewRedirector(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if redirector.Server.DB != nil || redirector.Server.Redis != nil || redirector.Server.Job != nil {
		t.Fatal("redirector owns data-plane resources")
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- redirector.Run(cancelCtx) }()
	requestCtx, cancelRequest := context.WithTimeout(ctx, time.Second)
	defer cancelRequest()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, "http://"+redirector.Address()+"/docs", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		cancel()
		t.Fatal("redirector exposed docs")
	}
	cancel()
	select {
	case err409 := <-done:
		if err409 != nil {
			t.Fatal(err409)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("role cancellation did not finish")
	}
	assertRoleListenerReleased(t, redirector.Address())
}

func assertRoleListenerReleased(t *testing.T, address string) {
	t.Helper()
	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatalf("listener leaked after cleanup: %v", err)
	}
	if err425 := listener.Close(); err425 != nil {
		t.Fatal(err425)
	}
}

func TestRoleCloseAggregatesAndStartupReleasesListener(t *testing.T) {
	cause := errors.New("SECRET-MARKER cleanup")
	var opened, closed []string
	factories := roleSpies("", &opened, &closed, cause)
	runtime, err := newRole(context.Background(), config.RoleWorker, roleTestConfig(), factories)
	if err != nil {
		t.Fatal(err)
	}
	address := runtime.Address()
	err = runtime.Close(context.Background())
	if !errors.Is(err, cause) || strings.Contains(err.Error(), "SECRET-MARKER") {
		t.Fatalf("close failure lost or leaked: %v", err)
	}
	for _, name := range []string{"consumer", "redis", "logger", "telemetry"} {
		if !strings.Contains(err.Error(), name) {
			t.Fatalf("close skipped %s: %v", name, err)
		}
	}
	assertRoleListenerReleased(t, address)

	opened, closed = nil, nil
	factories = roleSpies("start consumer", &opened, &closed, cause)
	listen := factories.listen
	factories.listen = func(value string) (net.Listener, error) {
		listener, err454 := listen(value)
		if err454 == nil {
			address = listener.Addr().String()
		}
		return listener, err454
	}
	runtime, err = newRole(context.Background(), config.RoleWorker, roleTestConfig(), factories)
	if runtime != nil || !errors.Is(err, cause) {
		t.Fatalf("consumer startup failure disappeared: %v", err)
	}
	assertRoleListenerReleased(t, address)
}

// TestRoleBinaryStartup launches the real commands from an unrelated directory
// with only their owned configuration. Health and active drain are later gates.
//
//nolint:gocognit,gocyclo,cyclop // Keep this complete integration protocol and its ordered failure assertions together.
func TestRoleBinaryStartup(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binaries := t.TempDir()
	buildCtx, cancelBuild := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancelBuild()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", binaries+string(os.PathSeparator),
		"./cmd/api", "./cmd/redirector", "./cmd/worker", "./cmd/migrator", "./cmd/flux")
	build.Dir = root
	if output, err480 := build.CombinedOutput(); err480 != nil {
		t.Fatalf("build role binaries: %v: %s", err480, output)
	}
	for _, name := range []string{"api", "redirector", "worker", "migrator", "flux"} {
		t.Run(name+"/invalid_config", func(t *testing.T) {
			address := binaryTestAddress(t)
			role := name
			if role == "flux" {
				role = "api"
			}
			env := binaryTestEnv()
			env = append(env,
				"FLUX_DATABASE.PASSWORD=SECRET-MARKER-password",
				"FLUX_INTEGRATION.RESEND_API_KEY=SECRET-MARKER-email")
			if role == "migrator" {
				env = append(env, "FLUX_DATABASE.PORT=SECRET-MARKER-invalid")
			} else {
				env = append(env, "FLUX_"+strings.ToUpper(role)+".LISTEN_ADDRESS="+address,
					"FLUX_"+strings.ToUpper(role)+".DRAIN_TIMEOUT=SECRET-MARKER-invalid")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, filepath.Join(binaries, name))
			cmd.Dir, cmd.Env = t.TempDir(), env
			output, err502 := cmd.CombinedOutput()
			if err502 == nil ||
				ctx.Err() != nil ||
				!bytes.Contains(output,
					[]byte("configuration")) ||
				bytes.Contains(output,
					[]byte("SECRET-MARKER")) {
				t.Fatalf("invalid configuration exit/redaction: %v: %s", err502, output)
			}
			assertRoleListenerReleased(t, address)
		})
	}
	t.Run("redirector", func(t *testing.T) {
		address := binaryTestAddress(t)
		process := startRoleBinary(t, filepath.Join(binaries, "redirector"), address,
			append(binaryTestEnv(),
				"FLUX_REDIRECTOR.LISTEN_ADDRESS="+address,
				"FLUX_REDIRECTOR.DRAIN_TIMEOUT=1s"))
		assertBinaryRoutes(t, address, false)
		process.stop(t)
	})
	t.Run("api_and_compatibility", func(t *testing.T) {
		pg, closePG := backendTesting.SetupTestPostgres(t)
		defer closePG()
		for _, name := range []string{"api", "flux"} {
			t.Run(name, func(t *testing.T) {
				address := binaryTestAddress(t)
				env := append(binaryTestEnv(), binaryDatabaseEnv(pg.Config.Database)...)
				env = append(env, "FLUX_API.LISTEN_ADDRESS="+address, "FLUX_API.DRAIN_TIMEOUT=1s")
				process := startRoleBinary(t, filepath.Join(binaries, name), address, env)
				assertBinaryRoutes(t, address, true)
				var ledger *string
				if err527 := pg.Pool.QueryRow(context.Background(),
					"SELECT to_regclass('public.schema_version')::text").
					Scan(&ledger); err527 != nil ||
					ledger != nil {
					t.Fatalf("%s implicitly migrated empty PostgreSQL: ledger %v, %v",
						name,
						ledger,
						err527)
				}
				process.stop(t)
				var connections int
				if err532 := pg.Pool.QueryRow(context.Background(),
					"SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND pid <> pg_backend_pid()").
					Scan(&connections); err532 != nil ||
					connections != 0 {
					t.Fatalf("%s PostgreSQL teardown: connections %d, %v",
						name,
						connections,
						err532)
				}
			})
		}
	})
	t.Run("worker", func(t *testing.T) {
		queue, closeQueue := backendTesting.SetupTestRedis(t)
		defer closeQueue()
		delivered := make(chan bool, 1)
		transport := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var payload struct {
				HTML string   `json:"html"`
				To   []string `json:"to"`
			}
			err547 := json.NewDecoder(r.Body).Decode(&payload)
			valid := err547 == nil && r.Method == http.MethodPost && r.URL.Path == "/emails" &&
				r.Header.Get("Authorization") == "Bearer local-test-key" &&
				len(payload.To) == 1 &&
				payload.To[0] == "binary-test@example.com" &&
				strings.Contains(payload.HTML,
					"BinaryTest")
			select {
			case delivered <- valid:
			default:
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"id":"local-email"}`)
		}))
		defer transport.Close()
		address := binaryTestAddress(t)
		env := append(binaryTestEnv(), "FLUX_WORKER.LISTEN_ADDRESS="+address, "FLUX_WORKER.DRAIN_TIMEOUT=1s",
			"FLUX_REDIS.ADDRESS="+queue.Config.Address,
			"FLUX_INTEGRATION.RESEND_API_KEY=local-test-key",
			"RESEND_BASE_URL="+transport.URL+"/")
		process := startRoleBinary(t, filepath.Join(binaries, "worker"), address, env)
		assertBinaryRoutes(t, address, false)
		producer := asynq.NewClientFromRedisClient(queue.Client)
		task, err565 := job.NewWelcomeEmailTask("binary-test@example.com", "BinaryTest")
		if err565 != nil {
			t.Fatal(err565)
		}
		if _, err569 := producer.Enqueue(task); err569 != nil {
			t.Fatal(err569)
		}
		select {
		case valid := <-delivered:
			if !valid {
				t.Fatal("worker delivered an invalid local email request")
			}
		case <-time.After(6 * time.Second):
			t.Fatal("worker did not process the task through the local email transport")
		}
		process.stop(t)
		inspector := asynq.NewInspectorFromRedisClient(queue.Client)
		servers, err565 := inspector.Servers()
		if err565 != nil || len(servers) != 0 {
			t.Fatalf("worker teardown left registered consumers: count %d, %v", len(servers), err565)
		}
	})
	t.Run("migrator", func(t *testing.T) {
		pg, closePG := backendTesting.SetupTestPostgres(t)
		defer closePG()
		latest := latestMigrationVersion(t)
		for _, startVersion := range []int{0, latest} {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			cmd := exec.CommandContext(ctx, filepath.Join(binaries, "migrator"))
			cmd.Dir,
				cmd.Env = t.TempDir(),
				append(binaryTestEnv(),
					binaryDatabaseEnv(pg.Config.Database)...)
			output, err594 := cmd.CombinedOutput()
			cancel()
			if err594 != nil ||
				!bytes.Contains(output,
					[]byte(fmt.Sprintf(`"start_version":%d`,
						startVersion))) ||
				!bytes.Contains(output,
					[]byte(fmt.Sprintf(`"end_version":%d`, latest))) {
				t.Fatalf("migrator did not exit once with exact versions: %v: %s", err594, output)
			}
		}
		var version int
		if err601 := pg.Pool.QueryRow(context.Background(),
			"SELECT version FROM schema_version").
			Scan(&version); err601 != nil ||
			version != latest {
			t.Fatalf("migrator schema ledger: version %d, %v", version, err601)
		}
	})
	t.Run("task_targets", func(t *testing.T) {
		outputDir := t.TempDir()
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		build609 := exec.CommandContext(ctx, "task", "--dir", root, "build", "BIN_DIR="+outputDir)
		if output, err610 := build609.CombinedOutput(); err610 != nil {
			t.Fatalf("task build failed: %v: %s", err610, output)
		}
		for _, role := range []string{"api", "redirector", "worker", "migrator"} {
			if info,
				err614 := os.Stat(filepath.Join(outputDir,
				role)); err614 != nil ||
				!info.Mode().
					IsRegular() {
				t.Fatalf("task build did not produce %s binary: %v", role, err614)
			}
			for _, action := range []string{"run", "build"} {
				cmd := exec.Command("task", "--dir", root, "--dry", action+":"+role)
				output, err619 := cmd.CombinedOutput()
				if err619 != nil || !bytes.Contains(output, []byte("./cmd/"+role)) {
					t.Fatalf("task target %s:%s missing role command: %v: %s",
						action,
						role,
						err619,
						output)
				}
			}
		}
	})
}

func latestMigrationVersion(t *testing.T) int {
	t.Helper()
	files, err := filepath.Glob("../database/migrations/[0-9]*.sql")
	if err != nil || len(files) == 0 {
		t.Fatal("numbered migration corpus is missing")
	}
	return len(files)
}

func binaryTestEnv() []string {
	var env []string
	for _, item := range os.Environ() {
		if !strings.HasPrefix(item,
			"FLUX_") &&
			!strings.HasPrefix(item,
				"PG") &&
			!strings.HasPrefix(item,
				"RESEND_") {
			env = append(env, item)
		}
	}
	return append(env, "FLUX_PRIMARY.ENV=test")
}

func binaryDatabaseEnv(cfg config.DatabaseConfig) []string {
	return []string{"FLUX_DATABASE.HOST=" + cfg.Host, "FLUX_DATABASE.PORT=" + strconv.Itoa(cfg.Port),
		"FLUX_DATABASE.USER=" + cfg.User,
		"FLUX_DATABASE.PASSWORD=" + cfg.Password,
		"FLUX_DATABASE.NAME=" + cfg.Name,
		"FLUX_DATABASE.SSL_MODE=" + cfg.SSLMode,
		"FLUX_DATABASE.MAX_OPEN_CONNS=2",
		"FLUX_DATABASE.MAX_IDLE_CONNS=1",
		"FLUX_DATABASE.CONN_MAX_LIFETIME=60", "FLUX_DATABASE.CONN_MAX_IDLE_TIME=30"}
}

func binaryTestAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err652 := listener.Close(); err652 != nil {
		t.Fatal(err652)
	}
	return address
}

type roleBinaryProcess struct {
	cmd     *exec.Cmd
	done    chan error
	output  *bytes.Buffer
	cancel  context.CancelFunc
	address string
	stopped bool
}

func startRoleBinary(t *testing.T, binary, address string, env []string) *roleBinaryProcess {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	process := &roleBinaryProcess{done: make(chan error,
		1),
		output:  &bytes.Buffer{},
		address: address,
		cancel:  cancel}
	process.cmd = exec.CommandContext(ctx, binary)
	process.cmd.Dir, process.cmd.Env = t.TempDir(), env
	process.cmd.Stdout, process.cmd.Stderr = process.output, process.output
	if err := process.cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	go func() { process.done <- process.cmd.Wait() }()
	t.Cleanup(func() { process.stop(t) })
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		connection, err := net.DialTimeout("tcp", address, 50*time.Millisecond)
		if err == nil {
			_ = connection.Close()
			return process
		}
		select {
		case err688 := <-process.done:
			process.stopped = true
			cancel()
			t.Fatalf("binary exited before binding role address: %v: %s", err688, process.output)
		default:
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("binary did not bind its configured role address")
	return nil
}

func (p *roleBinaryProcess) stop(t *testing.T) {
	t.Helper()
	if p.stopped {
		return
	}
	p.stopped = true
	defer p.cancel()
	if err := p.cmd.Process.Signal(os.Interrupt); err != nil {
		t.Errorf("signal binary teardown: %v", err)
	}
	select {
	case err := <-p.done:
		if err != nil {
			t.Errorf("binary teardown exit: %v: %s", err, p.output)
		}
	case <-time.After(4 * time.Second):
		p.cancel()
		err := <-p.done
		t.Errorf("binary teardown deadline; forced termination: %v: %s", err, p.output)
	}
	assertRoleListenerReleased(t, p.address)
}

func assertBinaryRoutes(t *testing.T, address string, api bool) {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	for _, path := range []string{"/docs", "/static/openapi.json", "/api/v1/links"} {
		response, err := client.Get("http://" + address + path)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
		want := http.StatusNotFound
		if api && path != "/api/v1/links" {
			want = http.StatusOK
		}
		if response.StatusCode != want {
			t.Fatalf("role route %s returned %d, want %d", path, response.StatusCode, want)
		}
	}
}
