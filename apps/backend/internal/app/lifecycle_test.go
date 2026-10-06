package app

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/lib/job"
	"github.com/6sLOGAN78/flux/internal/server"
	backendTesting "github.com/6sLOGAN78/flux/internal/testing"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

func TestPartialStartupRealLifecycleResources(t *testing.T) {
	pg, closePG := backendTesting.SetupTestPostgres(t)
	defer closePG()
	queue, closeQueue := backendTesting.SetupTestRedis(t)
	defer closeQueue()
	for _, role := range []config.Role{config.RoleAPI, config.RoleWorker} {
		t.Run(string(role), func(t *testing.T) {
			cfg := roleTestConfig()
			cfg.Database, cfg.Redis, cfg.Integration = pg.Config.Database, queue.Config, pg.Config.Integration
			cfg.API.ProducerEnabled = true
			f := defaultRoleFactories()
			var db *database.Database
			var client *redis.Client
			openDB, openRedis := f.database, f.redis
			f.database = func(ctx context.Context, srv *server.Server) (*database.Database, func(context.Context) error, error) {
				value, closer, err := openDB(ctx, srv)
				db = value
				return value, closer, err
			}
			f.redis = func(ctx context.Context, srv *server.Server) (*redis.Client, func(context.Context) error, error) {
				value, closer, err := openRedis(ctx, srv)
				client = value
				return value, closer, err
			}
			cause := errors.New("SECRET-MARKER startup")
			var address string
			if role == config.RoleAPI {
				f.router = func(config.Role, *server.Server) (*echo.Echo, error) { return nil, cause }
			} else {
				listen := f.listen
				f.listen = func(value string) (net.Listener, error) {
					listener, err := listen(value)
					if err == nil {
						address = listener.Addr().String()
					}
					return listener, err
				}
				f.startConsumer = func(*job.JobService) error { return cause }
			}
			r, err := newRole(context.Background(), role, cfg, f)
			if r != nil || !errors.Is(err, cause) || strings.Contains(err.Error(), "SECRET-MARKER") {
				t.Fatalf("partial startup: %v", err)
			}
			if db != nil && db.Pool.Stat().TotalConns() != 0 {
				t.Fatal("startup leaked PostgreSQL")
			}
			if client == nil || !errors.Is(client.Ping(context.Background()).Err(), redis.ErrClosed) {
				t.Fatal("startup leaked Redis")
			}
			if address != "" {
				assertRoleListenerReleased(t, address)
			}
		})
	}
}

// The fixture uses the actual role composition and real PostgreSQL. The only
// injected behavior is a deterministic HTTP handler so the active-work state
// is observable before signaling, independent of network scheduling.
func TestSIGTERMActiveHTTPFixture(t *testing.T) {
	role := config.Role(os.Getenv("FLUX_LIFECYCLE_FIXTURE"))
	if role == "" {
		return
	}
	cfg, err := config.LoadConfigForRole(role)
	if err != nil {
		t.Fatal(err)
	}
	f := defaultRoleFactories()
	f.router = func(role config.Role, srv *server.Server) (*echo.Echo, error) {
		e, err := defaultRoleRouter(role, srv)
		e.GET("/active", func(c echo.Context) error {
			c.Response().WriteHeader(200)
			_, _ = io.WriteString(c.Response(), "started\n")
			c.Response().Flush()
			time.Sleep(200 * time.Millisecond)
			_, err := io.WriteString(c.Response(), "completed\n")
			return err
		})
		return e, err
	}
	ctx, stop := NotifyContext(context.Background())
	defer stop()
	r, err := newRole(ctx, role, cfg, f)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Run(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestSIGTERMActiveHTTPSubprocess(t *testing.T) {
	pg, closePG := backendTesting.SetupTestPostgres(t)
	defer closePG()
	for _, role := range []string{"api", "redirector"} {
		t.Run(role, func(t *testing.T) {
			address := binaryTestAddress(t)
			env := append(binaryTestEnv(), binaryDatabaseEnv(pg.Config.Database)...)
			env = append(env, "FLUX_LIFECYCLE_FIXTURE="+role, "GORACE=atexit_sleep_ms=0", "FLUX_"+strings.ToUpper(role)+".LISTEN_ADDRESS="+address, "FLUX_"+strings.ToUpper(role)+".DRAIN_TIMEOUT=1s")
			// Launch our race-instrumented executable with only the fixture test.
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			p := &roleBinaryProcess{done: make(chan error, 1), output: &bytes.Buffer{}, address: address, cancel: cancel}
			p.cmd = exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSIGTERMActiveHTTPFixture$")
			p.cmd.Env = env
			p.cmd.Stdout, p.cmd.Stderr = p.output, p.output
			if err := p.cmd.Start(); err != nil {
				cancel()
				t.Fatal(err)
			}
			go func() { p.done <- p.cmd.Wait() }()
			t.Cleanup(func() { p.stop(t) })
			limit := time.Now().Add(3 * time.Second)
			for lifecycleHTTPStatus(address, "/ready") != 200 {
				if time.Now().After(limit) {
					t.Fatal("fixture did not become ready")
				}
				time.Sleep(10 * time.Millisecond)
			}
			client := &http.Client{Timeout: 2 * time.Second}
			response, err := client.Get("http://" + address + "/active")
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			reader := bufio.NewReader(response.Body)
			if line, err := reader.ReadString('\n'); err != nil || line != "started\n" {
				t.Fatalf("handler not active: %q %v", line, err)
			}
			started := time.Now()
			if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
				t.Fatal(err)
			}
			if line, err := reader.ReadString('\n'); err != nil || line != "completed\n" {
				t.Fatalf("active handler lost on SIGTERM: %q %v", line, err)
			}
			waitLifecycleExit(t, p, started, 0, 1200*time.Millisecond)
		})
	}
}

func TestLifecycleActiveHTTPHandler(t *testing.T) {
	for _, expire := range []bool{false, true} {
		t.Run(map[bool]string{false: "drain", true: "deadline"}[expire], func(t *testing.T) {
			entered, release, canceled := make(chan struct{}), make(chan struct{}), make(chan struct{})
			factories := defaultRoleFactories()
			factories.router = func(role config.Role, srv *server.Server) (*echo.Echo, error) {
				e, err := defaultRoleRouter(role, srv)
				e.GET("/active", func(c echo.Context) error {
					close(entered)
					select {
					case <-release:
						return c.String(200, "completed")
					case <-c.Request().Context().Done():
						close(canceled)
						return c.Request().Context().Err()
					}
				})
				return e, err
			}
			cfg := roleTestConfig()
			if expire {
				cfg.Redirector.DrainTimeout = 100 * time.Millisecond
			}
			r, err := newRole(context.Background(), config.RoleRedirector, cfg, factories)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done, request := make(chan error, 1), make(chan int, 1)
			go func() { done <- r.Run(ctx) }()
			go func() { request <- lifecycleHTTPStatus(r.Address(), "/active") }()
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("HTTP handler not entered")
			}
			started := time.Now()
			cancel()
			for r.lifecycle.Ready() {
				time.Sleep(time.Millisecond)
			}
			if !expire {
				select {
				case err := <-done:
					t.Fatalf("did not drain handler: %v", err)
				default:
				}
				close(release)
				if status := <-request; status != 200 {
					t.Fatalf("active HTTP lost: %d", status)
				}
			}
			select {
			case err := <-done:
				if expire != errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("shutdown status %v", err)
				}
			case <-time.After(time.Second):
				t.Fatal("HTTP deadline exceeded")
			}
			if expire {
				select {
				case <-canceled:
				case <-time.After(time.Second):
					t.Fatal("forced HTTP close did not cancel handler")
				}
				<-request
				if time.Since(started) > 500*time.Millisecond {
					t.Fatal("fresh HTTP cleanup budget")
				}
			}
			assertRoleListenerReleased(t, r.Address())
		})
	}
}

func TestLifecycleOrderAndFailures(t *testing.T) {
	var events []string
	cause := errors.New("SECRET-MARKER")
	var cleanup Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	l := Lifecycle{cleanup: &cleanup}
	for _, name := range []string{"shared", "dependent"} {
		name := name
		_ = cleanup.Push(name, func(got context.Context) error {
			if got != ctx {
				t.Error("cleanup received a fresh deadline")
			}
			events = append(events, name)
			return cause
		})
	}
	l.stop = func() {
		if l.Ready() {
			t.Error("intake stopped before unready")
		}
		events = append(events, "stop")
	}
	l.drain = func(got context.Context) error {
		if got != ctx {
			t.Error("drain received a fresh deadline")
		}
		events = append(events, "drain")
		return cause
	}
	err := l.Shutdown(ctx)
	if !errors.Is(err, cause) || strings.Contains(err.Error(), "SECRET-MARKER") {
		t.Fatalf("unsafe/lost errors: %v", err)
	}
	if !reflect.DeepEqual(events, []string{"stop", "drain", "dependent", "shared"}) {
		t.Fatalf("order: %v", events)
	}
	if !strings.Contains(err.Error(), "drain") || !strings.Contains(err.Error(), "dependent") || !strings.Contains(err.Error(), "shared") {
		t.Fatalf("missing stage: %v", err)
	}
	if l.Ready() {
		t.Fatal("shutdown stayed ready")
	}
	if again := l.Shutdown(context.Background()); again != err {
		t.Fatal("shutdown result changed")
	}
}

// Build the actual commands, rather than a test-only lifecycle executable.
func TestSIGTERMRoleProcesses(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binaries := t.TempDir()
	for _, role := range []string{"api", "redirector", "worker"} {
		cmd := exec.Command("go", "build", "-race", "-o", filepath.Join(binaries, role), "./cmd/"+role)
		cmd.Dir = root
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build %s: %v: %s", role, err, output)
		}
	}
	t.Run("http", func(t *testing.T) {
		pg, closePG := backendTesting.SetupTestPostgres(t)
		defer closePG()
		for _, role := range []string{"api", "redirector"} {
			t.Run(role, func(t *testing.T) {
				address := binaryTestAddress(t)
				env := append(binaryTestEnv(), binaryDatabaseEnv(pg.Config.Database)...)
				env = append(env, "FLUX_"+strings.ToUpper(role)+".LISTEN_ADDRESS="+address, "FLUX_"+strings.ToUpper(role)+".DRAIN_TIMEOUT=2s")
				p := startRoleBinary(t, filepath.Join(binaries, role), address, env)
				assertLifecycleHTTP(t, address, "/ready", 200)
				started := time.Now()
				if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
					t.Fatal(err)
				}
				waitLifecycleExit(t, p, started, 0, 3*time.Second)
				if role == "api" {
					var count int
					if err := pg.Pool.QueryRow(context.Background(), "SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND pid <> pg_backend_pid()").Scan(&count); err != nil || count != 0 {
						t.Fatalf("pool leaked: %d %v", count, err)
					}
				}
			})
		}
	})
	for _, deadline := range []bool{false, true} {
		name := "worker_drain"
		if deadline {
			name = "worker_deadline"
		}
		t.Run(name, func(t *testing.T) {
			queue, closeQueue := backendTesting.SetupTestRedis(t)
			defer closeQueue()
			active, release := make(chan struct{}, 2), make(chan struct{})
			transport := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				active <- struct{}{}
				<-release
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"id":"drained"}`)
			}))
			defer transport.Close()
			defer close(release)
			address := binaryTestAddress(t)
			budget := "3s"
			if deadline {
				budget = "300ms"
			}
			p := startRoleBinary(t, filepath.Join(binaries, "worker"), address, append(binaryTestEnv(), "GORACE=atexit_sleep_ms=0",
				"FLUX_WORKER.LISTEN_ADDRESS="+address, "FLUX_WORKER.DRAIN_TIMEOUT="+budget,
				"FLUX_REDIS.ADDRESS="+queue.Config.Address, "FLUX_INTEGRATION.RESEND_API_KEY=local-test-key", "RESEND_BASE_URL="+transport.URL+"/"))
			assertLifecycleHTTP(t, address, "/ready", 200)
			producer := asynq.NewClientFromRedisClient(queue.Client)
			task, err := job.NewWelcomeEmailTask("drain@example.com", "Active")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := producer.Enqueue(task); err != nil {
				t.Fatal(err)
			}
			select {
			case <-active:
			case <-time.After(5 * time.Second):
				t.Fatal("real worker job did not start")
			}
			started := time.Now()
			if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
				t.Fatal(err)
			}
			// A queue worker remains observable while its accepted work drains.
			limit := time.Now().Add(200 * time.Millisecond)
			for lifecycleHTTPStatus(address, "/ready") != 503 {
				if time.Now().After(limit) {
					t.Fatal("worker did not report unready during active job")
				}
				time.Sleep(5 * time.Millisecond)
			}
			assertLifecycleHTTP(t, address, "/live", 200)
			if deadline {
				waitLifecycleExit(t, p, started, 1, 1500*time.Millisecond)
				if !bytes.Contains(p.output.Bytes(), []byte("deadline")) {
					t.Fatalf("missing safe deadline status: %s", p.output)
				}
				return
			}
			second, _ := job.NewWelcomeEmailTask("queued@example.com", "Queued")
			if _, err := producer.Enqueue(second); err != nil {
				t.Fatal(err)
			}
			time.Sleep(100 * time.Millisecond)
			select {
			case <-active:
				t.Fatal("worker claimed new work during drain")
			default:
			}
			// Permit only the in-flight job to complete; the second remains queued.
			release <- struct{}{}
			waitLifecycleExit(t, p, started, 0, 4*time.Second)
			inspector := asynq.NewInspectorFromRedisClient(queue.Client)
			info, err := inspector.GetQueueInfo("default")
			if err != nil || info.Pending != 1 || info.Processed != 1 {
				t.Fatalf("active/queued work lost: %+v %v", info, err)
			}
			servers, err := inspector.Servers()
			if err != nil || len(servers) != 0 {
				t.Fatalf("consumer leaked: %v %v", servers, err)
			}
		})
	}
}

func lifecycleHTTPStatus(address, path string) int {
	client := &http.Client{Timeout: 200 * time.Millisecond}
	response, err := client.Get("http://" + address + path)
	if err != nil {
		return 0
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)
	return response.StatusCode
}

func assertLifecycleHTTP(t *testing.T, address, path string, want int) {
	t.Helper()
	if got := lifecycleHTTPStatus(address, path); got != want {
		t.Fatalf("%s status %d, want %d", path, got, want)
	}
}

func waitLifecycleExit(t *testing.T, p *roleBinaryProcess, started time.Time, code int, bound time.Duration) {
	t.Helper()
	p.stopped = true
	defer p.cancel()
	select {
	case err := <-p.done:
		got := p.cmd.ProcessState.ExitCode()
		if got != code || time.Since(started) > bound {
			t.Fatalf("exit %d, wanted %d within %s: %v: %s", got, code, bound, err, p.output)
		}
	case <-time.After(bound):
		p.cancel()
		<-p.done
		t.Fatalf("SIGTERM exceeded %s: %s", bound, p.output)
	}
	if bytes.Contains(p.output.Bytes(), []byte("DATA RACE")) {
		t.Fatalf("binary race: %s", p.output)
	}
	assertRoleListenerReleased(t, p.address)
}

func TestLifecycleDeadlinePreservesSerialOwnership(t *testing.T) {
	var cleanup Cleanup
	blocked, release, later := make(chan struct{}), make(chan struct{}), make(chan struct{})
	_ = cleanup.Push("shared", func(context.Context) error { close(later); return nil })
	_ = cleanup.Push("dependent", func(context.Context) error { close(blocked); <-release; return nil })
	l := Lifecycle{cleanup: &cleanup}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	started := time.Now()
	err := l.Shutdown(ctx)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > 250*time.Millisecond {
		t.Fatalf("unbounded shutdown: %v", err)
	}
	<-blocked
	select {
	case <-later:
		t.Fatal("shared resource closed under active dependent")
	default:
	}
	close(release)
	select {
	case <-later:
	case <-time.After(time.Second):
		t.Fatal("later closer skipped")
	}
}

func TestLifecycleReadinessGate(t *testing.T) {
	role, err := NewRedirector(context.Background(), roleTestConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer role.Close(context.Background())
	if err := role.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	role.HTTP.ServeHTTP(response, httptest.NewRequest("GET", "/ready", nil))
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), `"not_ready"`) {
		t.Fatalf("shutdown readiness %d: %s", response.Code, response.Body)
	}
}
