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
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/lib/email"
	"github.com/6sLOGAN78/flux/internal/lib/job"
	loggerPkg "github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/server"
	backendTesting "github.com/6sLOGAN78/flux/internal/testing"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func roleTestConfig() *config.Config {
	return &config.Config{Primary: config.Primary{Env: "test"}, Observability: config.DefaultObservabilityConfig(),
		API:        config.RoleConfig{ListenAddress: "127.0.0.1:0", DrainTimeout: time.Second, ReadinessTimeout: time.Second},
		Redirector: config.RoleConfig{ListenAddress: "127.0.0.1:0", DrainTimeout: time.Second, ReadinessTimeout: time.Second},
		Worker:     config.RoleConfig{ListenAddress: "127.0.0.1:0", DrainTimeout: time.Second, ReadinessTimeout: time.Second},
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
		logger: func(*config.Config) (*zerolog.Logger, *loggerPkg.LoggerService, func(context.Context) error, error) {
			close, err := stage("logger")
			log := zerolog.Nop()
			return &log, nil, close, err
		},
		database: func(context.Context, *server.Server) (*database.Database, func(context.Context) error, error) {
			close, err := stage("database")
			return &database.Database{}, close, err
		},
		redis: func(context.Context, *server.Server) (*redis.Client, func(context.Context) error, error) {
			close, err := stage("redis")
			return nil, close, err
		},
		producer: func(*server.Server) (*job.JobService, func(context.Context) error, error) {
			close, err := stage("producer")
			return nil, close, err
		},
		email: func(*server.Server) (*email.Client, error) {
			if fail == "email" {
				return nil, cause
			}
			*opened = append(*opened, "email")
			return nil, nil
		},
		consumer: func(*server.Server, *email.Client) (*job.JobService, func(context.Context) error, error) {
			close, err := stage("consumer")
			return nil, close, err
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

func TestRoleResourceGraphs(t *testing.T) {
	for _, test := range []struct {
		role     config.Role
		producer bool
		want     []string
	}{
		{config.RoleAPI, false, []string{"logger", "database", "router", "listener"}},
		{config.RoleAPI, true, []string{"logger", "database", "redis", "producer", "router", "listener"}},
		{config.RoleRedirector, false, []string{"logger", "router", "listener"}},
		{config.RoleWorker, false, []string{"logger", "redis", "email", "consumer", "router", "listener", "start consumer"}},
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
					role.HTTP.ServeHTTP(recorder, httptest.NewRequest("GET", path, nil))
					if recorder.Code != 404 {
						t.Fatalf("%s exposed %s: %d", test.role, path, recorder.Code)
					}
				}
			}
			if err := role.Close(context.Background()); err != nil {
				t.Fatal(err)
			}
			wantClosed := []string{"logger"}
			if test.role == config.RoleAPI {
				wantClosed = []string{"database", "logger"}
				if test.producer {
					wantClosed = []string{"producer", "redis", "database", "logger"}
				}
			}
			if test.role == config.RoleWorker {
				wantClosed = []string{"consumer", "redis", "logger"}
			}
			if !reflect.DeepEqual(closed, wantClosed) {
				t.Fatalf("close order %v, want %v", closed, wantClosed)
			}
			if err := role.Close(context.Background()); err != nil || !reflect.DeepEqual(closed, wantClosed) {
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

func TestRoleConstructionFailureUnwinds(t *testing.T) {
	cause := errors.New("SECRET-MARKER")
	for _, role := range []config.Role{config.RoleAPI, config.RoleRedirector, config.RoleWorker} {
		stages := []string{"logger", "router", "listener"}
		if role == config.RoleAPI {
			stages = []string{"logger", "database", "redis", "producer", "router", "listener"}
		}
		if role == config.RoleWorker {
			stages = []string{"logger", "redis", "email", "consumer", "router", "listener", "start consumer"}
		}
		for _, fail := range stages {
			t.Run(string(role)+"/"+fail, func(t *testing.T) {
				cfg := roleTestConfig()
				cfg.API.ProducerEnabled = true
				var opened, closed []string
				runtime, err := newRole(context.Background(), role, cfg, roleSpies(fail, &opened, &closed, cause))
				if runtime != nil || !errors.Is(err, cause) || strings.Contains(err.Error(), "SECRET-MARKER") {
					t.Fatalf("unsafe or lost failure: %v", err)
				}
				var want []string
				for i := len(opened) - 1; i >= 0; i-- {
					switch opened[i] {
					case "logger", "database", "redis", "producer", "consumer":
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

func TestRoleServerReceivesOnlyExplicitResources(t *testing.T) {
	log := zerolog.Nop()
	srv, err := server.New(roleTestConfig(), &log, nil)
	if err != nil || srv.DB != nil || srv.Redis != nil || srv.Job != nil {
		t.Fatalf("server allocated undeclared dependencies: %v", err)
	}
}

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
		if err := api.Close(ctx); err != nil {
			t.Error(err)
		}
	})
	if api.Server.Redis != nil || api.Server.Job != nil {
		t.Fatal("default API owns queue infrastructure")
	}
	var ledger *string
	if err := api.Server.DB.Pool.QueryRow(ctx, "SELECT to_regclass('schema_version')::text").Scan(&ledger); err != nil || ledger != nil {
		t.Fatalf("API migrated empty PostgreSQL: %v, %v", ledger, err)
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
	if err := worker.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if err := worker.Server.Redis.Ping(ctx).Err(); err == nil {
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
	req, err := http.NewRequestWithContext(requestCtx, "GET", "http://"+redirector.Address()+"/docs", nil)
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
	case err := <-done:
		if err != nil {
			t.Fatal(err)
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
	if err := listener.Close(); err != nil {
		t.Fatal(err)
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
	for _, name := range []string{"consumer", "redis", "logger"} {
		if !strings.Contains(err.Error(), name) {
			t.Fatalf("close skipped %s: %v", name, err)
		}
	}
	assertRoleListenerReleased(t, address)

	opened, closed = nil, nil
	factories = roleSpies("start consumer", &opened, &closed, cause)
	listen := factories.listen
	factories.listen = func(value string) (net.Listener, error) {
		listener, err := listen(value)
		if err == nil {
			address = listener.Addr().String()
		}
		return listener, err
	}
	runtime, err = newRole(context.Background(), config.RoleWorker, roleTestConfig(), factories)
	if runtime != nil || !errors.Is(err, cause) {
		t.Fatalf("consumer startup failure disappeared: %v", err)
	}
	assertRoleListenerReleased(t, address)
}

// TestRoleBinaryStartup launches the real commands from an unrelated directory
// with only their owned configuration. Health and active drain are later gates.
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
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build role binaries: %v: %s", err, output)
	}
	for _, name := range []string{"api", "redirector", "worker", "migrator", "flux"} {
		t.Run(name+"/invalid_config", func(t *testing.T) {
			address := binaryTestAddress(t)
			role := name
			if role == "flux" {
				role = "api"
			}
			env := binaryTestEnv()
			env = append(env, "FLUX_DATABASE.PASSWORD=SECRET-MARKER-password", "FLUX_INTEGRATION.RESEND_API_KEY=SECRET-MARKER-email")
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
			output, err := cmd.CombinedOutput()
			if err == nil || ctx.Err() != nil || !bytes.Contains(output, []byte("configuration")) || bytes.Contains(output, []byte("SECRET-MARKER")) {
				t.Fatalf("invalid configuration exit/redaction: %v: %s", err, output)
			}
			assertRoleListenerReleased(t, address)
		})
	}
	t.Run("redirector", func(t *testing.T) {
		address := binaryTestAddress(t)
		process := startRoleBinary(t, filepath.Join(binaries, "redirector"), address,
			append(binaryTestEnv(), "FLUX_REDIRECTOR.LISTEN_ADDRESS="+address, "FLUX_REDIRECTOR.DRAIN_TIMEOUT=1s"))
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
				if err := pg.Pool.QueryRow(context.Background(), "SELECT to_regclass('public.schema_version')::text").Scan(&ledger); err != nil || ledger != nil {
					t.Fatalf("%s implicitly migrated empty PostgreSQL: ledger %v, %v", name, ledger, err)
				}
				process.stop(t)
				var connections int
				if err := pg.Pool.QueryRow(context.Background(), "SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND pid <> pg_backend_pid()").Scan(&connections); err != nil || connections != 0 {
					t.Fatalf("%s PostgreSQL teardown: connections %d, %v", name, connections, err)
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
				To   []string `json:"to"`
				HTML string   `json:"html"`
			}
			err := json.NewDecoder(r.Body).Decode(&payload)
			valid := err == nil && r.Method == "POST" && r.URL.Path == "/emails" &&
				r.Header.Get("Authorization") == "Bearer local-test-key" &&
				len(payload.To) == 1 && payload.To[0] == "binary-test@example.com" && strings.Contains(payload.HTML, "BinaryTest")
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
			"FLUX_REDIS.ADDRESS="+queue.Config.Address, "FLUX_INTEGRATION.RESEND_API_KEY=local-test-key", "RESEND_BASE_URL="+transport.URL+"/")
		process := startRoleBinary(t, filepath.Join(binaries, "worker"), address, env)
		assertBinaryRoutes(t, address, false)
		producer := asynq.NewClientFromRedisClient(queue.Client)
		task, err := job.NewWelcomeEmailTask("binary-test@example.com", "BinaryTest")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := producer.Enqueue(task); err != nil {
			t.Fatal(err)
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
		servers, err := inspector.Servers()
		if err != nil || len(servers) != 0 {
			t.Fatalf("worker teardown left registered consumers: count %d, %v", len(servers), err)
		}
	})
	t.Run("migrator", func(t *testing.T) {
		pg, closePG := backendTesting.SetupTestPostgres(t)
		defer closePG()
		for _, startVersion := range []int{0, 1} {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			cmd := exec.CommandContext(ctx, filepath.Join(binaries, "migrator"))
			cmd.Dir, cmd.Env = t.TempDir(), append(binaryTestEnv(), binaryDatabaseEnv(pg.Config.Database)...)
			output, err := cmd.CombinedOutput()
			cancel()
			if err != nil || !bytes.Contains(output, []byte(fmt.Sprintf(`"start_version":%d`, startVersion))) || !bytes.Contains(output, []byte(`"end_version":1`)) {
				t.Fatalf("migrator did not exit once with exact versions: %v: %s", err, output)
			}
		}
		var version int
		if err := pg.Pool.QueryRow(context.Background(), "SELECT version FROM schema_version").Scan(&version); err != nil || version != 1 {
			t.Fatalf("migrator schema ledger: version %d, %v", version, err)
		}
	})
	t.Run("task_targets", func(t *testing.T) {
		outputDir := t.TempDir()
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		build := exec.CommandContext(ctx, "task", "--dir", root, "build", "BIN_DIR="+outputDir)
		if output, err := build.CombinedOutput(); err != nil {
			t.Fatalf("task build failed: %v: %s", err, output)
		}
		for _, role := range []string{"api", "redirector", "worker", "migrator"} {
			if info, err := os.Stat(filepath.Join(outputDir, role)); err != nil || !info.Mode().IsRegular() {
				t.Fatalf("task build did not produce %s binary: %v", role, err)
			}
			for _, action := range []string{"run", "build"} {
				cmd := exec.Command("task", "--dir", root, "--dry", action+":"+role)
				output, err := cmd.CombinedOutput()
				if err != nil || !bytes.Contains(output, []byte("./cmd/"+role)) {
					t.Fatalf("task target %s:%s missing role command: %v: %s", action, role, err, output)
				}
			}
		}
	})
}

func binaryTestEnv() []string {
	var env []string
	for _, item := range os.Environ() {
		if !strings.HasPrefix(item, "FLUX_") && !strings.HasPrefix(item, "PG") && !strings.HasPrefix(item, "RESEND_") {
			env = append(env, item)
		}
	}
	return append(env, "FLUX_PRIMARY.ENV=test")
}

func binaryDatabaseEnv(cfg config.DatabaseConfig) []string {
	return []string{"FLUX_DATABASE.HOST=" + cfg.Host, "FLUX_DATABASE.PORT=" + strconv.Itoa(cfg.Port),
		"FLUX_DATABASE.USER=" + cfg.User, "FLUX_DATABASE.PASSWORD=" + cfg.Password, "FLUX_DATABASE.NAME=" + cfg.Name,
		"FLUX_DATABASE.SSL_MODE=" + cfg.SSLMode, "FLUX_DATABASE.MAX_OPEN_CONNS=2", "FLUX_DATABASE.MAX_IDLE_CONNS=1",
		"FLUX_DATABASE.CONN_MAX_LIFETIME=60", "FLUX_DATABASE.CONN_MAX_IDLE_TIME=30"}
}

func binaryTestAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return address
}

type roleBinaryProcess struct {
	cmd     *exec.Cmd
	done    chan error
	output  *bytes.Buffer
	address string
	cancel  context.CancelFunc
	stopped bool
}

func startRoleBinary(t *testing.T, binary, address string, env []string) *roleBinaryProcess {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	process := &roleBinaryProcess{done: make(chan error, 1), output: &bytes.Buffer{}, address: address, cancel: cancel}
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
		case err := <-process.done:
			process.stopped = true
			cancel()
			t.Fatalf("binary exited before binding role address: %v: %s", err, process.output)
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
