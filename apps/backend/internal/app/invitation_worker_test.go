//nolint:testpackage // Verify worker-owned resource factories and actual dependencies.
package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/lib/email"
	"github.com/6sLOGAN78/flux/internal/lib/invitationcrypto"
	"github.com/6sLOGAN78/flux/internal/lib/job"
	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/6sLOGAN78/flux/internal/server"
	backendTesting "github.com/6sLOGAN78/flux/internal/testing"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func TestInvitationWorkerRoleOwnership(t *testing.T) {
	pg, closePG := backendTesting.SetupTestDB(t)
	defer closePG()
	queue, closeQueue := backendTesting.SetupTestRedis(t)
	defer closeQueue()
	cfg := roleTestConfig()
	cfg.Database, cfg.Server, cfg.Redis = pg.Config.Database, pg.Config.Server, queue.Config
	cfg.Integration.ResendAPIKey = "test-only"
	for _, stage := range []string{"database", "redis", "email", "consumer", "start consumer"} {
		t.Run("partial startup "+stage, func(t *testing.T) {
			verifyInvitationWorkerStartupFailure(t, cfg, stage)
		})
	}
	t.Run("readiness probes actual owned dependencies locally", func(t *testing.T) {
		worker, err := NewWorker(context.Background(), cfg)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if closeErr := worker.Close(context.Background()); closeErr != nil && !errors.Is(closeErr, redis.ErrClosed) {
				t.Error(closeErr)
			}
		})
		assertInvitationWorkerReady(t, worker.RoleRuntime, http.StatusOK, "")
		// Drain the real consumer before deliberately closing its shared resources.
		// Closing Redis underneath Asynq also closes its PubSub channel.
		worker.Server.Job.StopIntake()
		if err = worker.Server.Job.Drain(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err = worker.Server.Redis.Close(); err != nil {
			t.Fatal(err)
		}
		assertInvitationWorkerReady(t, worker.RoleRuntime, http.StatusServiceUnavailable, "redis")
		worker.Server.DB.Pool.Close()
		assertInvitationWorkerReady(t, worker.RoleRuntime, http.StatusServiceUnavailable, "database")
	})
	t.Run("intake stops and active delivery drains before resources close", func(t *testing.T) {
		verifyInvitationWorkerDrain(t, cfg, pg.Pool)
	})
}

//nolint:gocognit // Preserve the real factory graph and reverse cleanup assertions in one failure protocol.
func verifyInvitationWorkerStartupFailure(t *testing.T, cfg *config.Config, stage string) {
	t.Helper()
	f := defaultRoleFactories()
	cause := errors.New("SECRET-MARKER startup")
	var srv *server.Server
	var adapter *email.Client
	var closed []string
	track := func(name string, closeResource func(context.Context) error) func(context.Context) error {
		return func(ctx context.Context) error { closed = append(closed, name); return closeResource(ctx) }
	}
	dbFactory, redisFactory, emailFactory, consumerFactory := f.database, f.redis, f.email, f.consumer
	f.database = func(ctx context.Context, s *server.Server) (*database.Database, func(context.Context) error, error) {
		srv = s
		db, closeResource, err := dbFactory(ctx, s)
		if err != nil {
			return db, closeResource, err
		}
		// Retain the real allocation even when construction reports failure.
		s.DB = db
		if stage == "database" {
			err = cause
		}
		return db, track("database", closeResource), err
	}
	f.redis = func(ctx context.Context, s *server.Server) (*redis.Client, func(context.Context) error, error) {
		client, closeResource, err := redisFactory(ctx, s)
		if err != nil {
			return client, closeResource, err
		}
		s.Redis = client
		if stage == "redis" {
			err = cause
		}
		return client, track("redis", closeResource), err
	}
	f.email = func(s *server.Server) (*email.Client, error) {
		var err error
		adapter, err = emailFactory(s)
		if stage == "email" {
			err = cause
		}
		return adapter, err
	}
	f.consumer = func(s *server.Server, e *email.Client) (*job.JobService, func(context.Context) error, error) {
		j, closeResource, err := consumerFactory(s, e)
		if stage == "consumer" {
			err = cause
		}
		return j, track("consumer", closeResource), err
	}
	f.startConsumer = func(j *job.JobService) error {
		if err := j.Start(); err != nil {
			return err
		}
		return cause
	}
	r, err := newRole(context.Background(), config.RoleWorker, cfg, f)
	if r != nil || !errors.Is(err, cause) || strings.Contains(err.Error(), "SECRET-MARKER") {
		t.Fatal("real resource startup failure lost or disclosed cause")
	}
	if srv == nil || srv.DB.Pool.Stat().TotalConns() != 0 || srv.DB.Pool.Ping(context.Background()) == nil {
		t.Fatal("partial startup leaked its real PostgreSQL pool")
	}
	if srv.Redis != nil && !errors.Is(srv.Redis.Ping(context.Background()).Err(), redis.ErrClosed) {
		t.Fatal("partial startup leaked its real Redis client")
	}
	if adapter != nil && workerReadinessChecks(srv, adapter)[2].Check(context.Background()) == nil {
		t.Fatal("partial startup leaked its email adapter")
	}
	want := []string{"database"}
	if stage != "database" {
		want = append([]string{"redis"}, want...)
	}
	if stage == "consumer" || stage == "start consumer" {
		want = append([]string{"consumer"}, want...)
	}
	if strings.Join(closed, ",") != strings.Join(want, ",") {
		t.Fatalf("cleanup %v, want %v", closed, want)
	}
}

func assertInvitationWorkerReady(t *testing.T, r *RoleRuntime, status int, component string) {
	t.Helper()
	response := httptest.NewRecorder()
	r.HTTP.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if response.Code != status {
		t.Fatalf("worker readiness %d, want %d", response.Code, status)
	}
	var body struct {
		Checks []struct {
			Name  string `json:"name"`
			State string `json:"state"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if component != "" {
		for _, check := range body.Checks {
			if check.Name == component && check.State == "not_ready" {
				return
			}
		}
		t.Fatalf("readiness did not report unavailable %s", component)
	}
}

type invitationDrainSender struct {
	entered chan struct{}
	release chan struct{}
}

func (s invitationDrainSender) SendInvitation(context.Context, invitationcrypto.Message) error {
	close(s.entered)
	<-s.release
	return nil
}

func verifyInvitationWorkerDrain(t *testing.T, cfg *config.Config, pool *pgxpool.Pool) {
	t.Helper()
	claim, _ := seedInvitationDelivery(t, pool, cfg, "drain@example.test")
	sender := invitationDrainSender{entered: make(chan struct{}), release: make(chan struct{})}
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(sender.release) }) }
	t.Cleanup(release)
	worker, err := NewWorkerWithInvitationSender(context.Background(), cfg, sender)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { release(); _ = worker.Close(context.Background()) })
	select {
	case <-sender.entered:
	case <-time.After(8 * time.Second):
		t.Fatal("actual queued invitation never entered sender")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	started := time.Now()
	err = worker.Close(ctx)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > time.Second {
		t.Fatal("worker shutdown ignored shared deadline")
	}
	assertInvitationWorkerReady(t, worker.RoleRuntime, http.StatusServiceUnavailable, "")
	if worker.Server.DB.Pool.Ping(context.Background()) != nil ||
		worker.Server.Redis.Ping(context.Background()).Err() != nil {
		t.Fatal("shared resource closed underneath active delivery")
	}
	release()
	waitInvitation(t, func() bool { return worker.Server.DB.Pool.Stat().TotalConns() == 0 })
	assertDeliveryState(t, pool, claim, "delivered", true)
	if !errors.Is(worker.Server.Redis.Ping(context.Background()).Err(), redis.ErrClosed) {
		t.Fatal("worker Redis remained open after drain")
	}
	assertRoleListenerReleased(t, worker.Address())
}

//nolint:lll,gocognit,gocyclo,cyclop // Keep actual dependencies and ordered recovery scenarios in one registered suite.
func TestInvitationWorkerActualRedisPostgres(t *testing.T) {
	pg, closePG := backendTesting.SetupTestDB(t)
	defer closePG()
	queue, closeQueue := backendTesting.SetupTestRedis(t)
	defer closeQueue()
	cfg := roleTestConfig()
	cfg.Database, cfg.Server, cfg.Redis = pg.Config.Database, pg.Config.Server, queue.Config
	cfg.Integration = config.IntegrationConfig{ResendAPIKey: "test-only"}
	worker1, err2 := NewWorker(context.Background(), cfg)
	if err2 != nil {
		t.Fatal(err2)
	}
	t.Cleanup(func() {
		if err3 := worker1.Close(context.Background()); err3 != nil {
			t.Error(err3)
		}
	})
	if worker1.Server.DB == nil {
		t.Fatal("durable invitation worker does not own PostgreSQL")
	}
	if err2 = worker1.Close(context.Background()); err2 != nil {
		t.Fatal(err2)
	}
	t.Run("acknowledgement erases envelope and never optimistic delivery", func(t *testing.T) {
		claim, token := seedInvitationDelivery(t, pg.Pool, cfg, "ack@example.test")
		entered, release := make(chan struct{}), make(chan struct{})
		var calls atomic.Int32
		var releaseOnce sync.Once
		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				To   string `json:"to"`
				HTML string `json:"html"`
			}
			if json.NewDecoder(r.Body).Decode(&body) != nil || body.To != "ack@example.test" || !strings.Contains(body.HTML, "/invitations#token="+token) || strings.Contains(body.HTML, "<script>") || !strings.Contains(body.HTML, "&lt;script&gt;") || r.Header.Get("Idempotency-Key") == "" {
				t.Error("provider received invalid rendered invitation")
			}
			if calls.Add(1) == 1 {
				close(entered)
			}
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"id":"acknowledged"}`)
		}))
		defer provider.Close()
		defer releaseOnce.Do(func() { close(release) })
		var capture invitationLogBuffer
		f := defaultRoleFactories()
		f.logger = func(*config.Config, *observability.Telemetry) (*zerolog.Logger, func(context.Context) error, error) {
			log := zerolog.New(&capture)
			return &log, nil, nil
		}
		f.consumer = func(srv *server.Server, adapter *email.Client) (*job.JobService, func(context.Context) error, error) {
			j := job.NewConsumer(srv.Logger, srv.Config, srv.Redis, adapter, srv.Telemetry)
			err4 := j.ConfigureInvitations(srv.DB.Pool, cfg.Invitations, invitationHTTPSender{endpoint: provider.URL, client: provider.Client()})
			return j, func(context.Context) error { return j.Stop() }, err4
		}
		runtime, err5 := newRole(context.Background(), config.RoleWorker, cfg, f)
		if err5 != nil {
			t.Fatal(err5)
		}
		defer func() {
			if err6 := runtime.Close(context.Background()); err6 != nil {
				t.Error(err6)
			}
		}()
		select {
		case <-entered:
		case <-time.After(8 * time.Second):
			t.Fatal("actual SQL/Redis pipeline did not send")
		}
		assertDeliveryState(t, pg.Pool, claim, "leased", false)
		inspector := asynq.NewInspectorFromRedisClient(queue.Client)
		active, err5 := inspector.ListActiveTasks("default")
		if err5 != nil || len(active) != 1 {
			t.Fatal("actual Redis active task unavailable")
		}
		payload := string(active[0].Payload)
		for _, secret := range []string{token, "ack@example.test", cfg.Invitations.EncryptionKeys, "ciphertext", "key_id"} {
			if strings.Contains(payload, secret) {
				t.Fatal("private material in Redis task")
			}
		}
		var reference repository.DeliveryClaim
		if json.Unmarshal(active[0].Payload, &reference) != nil || reference.WorkspaceID != claim.WorkspaceID || reference.DeliveryID != claim.DeliveryID {
			t.Fatal("Redis task lacks scoped durable references")
		}
		// Explicit acknowledgement releases the real HTTP send. Keep teardown safe.
		releaseOnce.Do(func() { close(release) })
		waitInvitation(t, func() bool {
			var state string
			_ = pg.Pool.QueryRow(context.Background(), "SELECT state FROM invitation_delivery_intents WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID).Scan(&state)
			return state == "delivered"
		})
		assertDeliveryState(t, pg.Pool, claim, "delivered", true)
		if calls.Load() != 1 {
			t.Fatal("duplicate provider send")
		}
		for _, secret := range []string{token, "ack@example.test", cfg.Invitations.EncryptionKeys} {
			if strings.Contains(capture.String(), secret) {
				t.Fatal("private invitation material in worker logs")
			}
		}
	})
	t.Run("provider retry payload stays immutable across workspace changes", func(t *testing.T) {
		claim, _ := seedInvitationDelivery(t, pg.Pool, cfg, "immutable@example.test")
		bodies := make(chan string, 2)
		keys := make(chan string, 2)
		var calls atomic.Int32
		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, readErr := io.ReadAll(r.Body)
			if readErr != nil {
				t.Error("provider read failed")
				return
			}
			bodies <- string(body)
			keys <- r.Header.Get("Idempotency-Key")
			if calls.Add(1) == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			_, _ = io.WriteString(w, `{"id":"stable-ack"}`)
		}))
		defer provider.Close()
		worker7, workerErr := NewWorkerWithInvitationSender(context.Background(), cfg, invitationHTTPSender{endpoint: provider.URL, client: provider.Client()})
		if workerErr != nil {
			t.Fatal(workerErr)
		}
		defer func() {
			if closeErr := worker7.Close(context.Background()); closeErr != nil {
				t.Error(closeErr)
			}
		}()
		var first string
		select {
		case first = <-bodies:
		case <-time.After(8 * time.Second):
			t.Fatal("first provider attempt missing")
		}
		waitInvitation(t, func() bool {
			var state string
			_ = pg.Pool.QueryRow(context.Background(), "SELECT state FROM invitation_delivery_intents WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID).Scan(&state)
			return state == "queued"
		})
		_, updateErr := pg.Pool.Exec(context.Background(), "UPDATE workspaces SET name='Changed workspace presentation' WHERE id=$1", claim.WorkspaceID)
		if updateErr != nil {
			t.Fatal(updateErr)
		}
		if closeErr := worker7.Close(context.Background()); closeErr != nil {
			t.Fatal(closeErr)
		}
		cfg.Invitations.Sender = "rotated-sender@example.test"
		cfg.Invitations.PublicOrigin = "https://new-app.flux.test"
		worker7, workerErr = NewWorkerWithInvitationSender(context.Background(), cfg, invitationHTTPSender{endpoint: provider.URL, client: provider.Client()})
		if workerErr != nil {
			t.Fatal(workerErr)
		}
		var second string
		select {
		case second = <-bodies:
		case <-time.After(8 * time.Second):
			t.Fatal("retry provider attempt missing")
		}
		if first != second {
			t.Fatal("same provider idempotency key received changed invitation bytes")
		}
		firstKey, secondKey := <-keys, <-keys
		if firstKey != secondKey {
			t.Fatal("retry changed provider intent identity")
		}
	})
	t.Run("actual HTTP send deadline leaves invitation retryable", func(t *testing.T) {
		claim, _ := seedInvitationDelivery(t, pg.Pool, cfg, "timeout@example.test")
		started := make(chan struct{})
		var once sync.Once
		provider := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			_, _ = io.Copy(io.Discard, r.Body)
			once.Do(func() { close(started) })
			<-r.Context().Done()
		}))
		defer provider.Close()
		timeoutWorker, workerErr := NewWorkerWithInvitationSender(context.Background(), cfg, invitationHTTPSender{endpoint: provider.URL, client: provider.Client()})
		if workerErr != nil {
			t.Fatal(workerErr)
		}
		defer func() {
			if closeErr := timeoutWorker.Close(context.Background()); closeErr != nil {
				t.Error(closeErr)
			}
		}()
		select {
		case <-started:
		case <-time.After(8 * time.Second):
			t.Fatal("real HTTP send did not start")
		}
		assertDeliveryState(t, pg.Pool, claim, "leased", false)
		deadline := time.Now().Add(13 * time.Second)
		for time.Now().Before(deadline) {
			var state string
			_ = pg.Pool.QueryRow(context.Background(), "SELECT state FROM invitation_delivery_intents WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID).Scan(&state)
			if state == "queued" {
				assertDeliveryState(t, pg.Pool, claim, "queued", false)
				return
			}
			time.Sleep(25 * time.Millisecond)
		}
		t.Fatal("per-send deadline did not return safe retry state")
	})
	store := repository.NewDeliveryRepository(pg.Pool)
	t.Run("lost Redis queue outage and restart reconstruct durable intent across key rotation", func(t *testing.T) {
		recoveryPG, closeRecoveryPG := backendTesting.SetupTestDB(t)
		defer closeRecoveryPG()
		recoveryRedis, closeRecoveryRedis := backendTesting.SetupTestRedis(t)
		defer closeRecoveryRedis()
		recoveryCfg := roleTestConfig()
		recoveryCfg.Database, recoveryCfg.Server, recoveryCfg.Redis = recoveryPG.Config.Database, recoveryPG.Config.Server, recoveryRedis.Config
		recoveryCfg.Integration = config.IntegrationConfig{ResendAPIKey: "test-only"}
		recoveryStore := repository.NewDeliveryRepository(recoveryPG.Pool)
		oldIntent, _ := seedInvitationDelivery(t, recoveryPG.Pool, recoveryCfg, "old-key@example.test")
		claims, claimErr := recoveryStore.Claim(context.Background())
		if claimErr != nil {
			t.Fatal(claimErr)
		}
		oldLease := findInvitationClaim(t, claims, oldIntent)
		task, taskErr := job.NewInvitationTask(oldLease)
		if taskErr != nil {
			t.Fatal(taskErr)
		}
		producer := asynq.NewClientFromRedisClient(recoveryRedis.Client)
		if _, taskErr = producer.Enqueue(task); taskErr != nil {
			t.Fatal(taskErr)
		}
		// This is the isolated queue owned by this test, never a shared service.
		if taskErr = recoveryRedis.Client.FlushDB(context.Background()).Err(); taskErr != nil {
			t.Fatal(taskErr)
		}
		stopTimeout := time.Second
		if taskErr = recoveryRedis.Container.Stop(context.Background(), &stopTimeout); taskErr != nil {
			t.Fatal(taskErr)
		}
		outageCtx, cancelOutage := context.WithTimeout(context.Background(), time.Second)
		_, taskErr = producer.EnqueueContext(outageCtx, task)
		cancelOutage()
		if taskErr == nil {
			t.Fatal("publication succeeded during actual Redis outage")
		}
		assertDeliveryState(t, recoveryPG.Pool, oldIntent, "leased", false)
		var attempts int
		if taskErr = recoveryPG.Pool.QueryRow(context.Background(), "SELECT attempts FROM invitation_delivery_intents WHERE workspace_id=$1 AND id=$2", oldIntent.WorkspaceID, oldIntent.DeliveryID).Scan(&attempts); taskErr != nil || attempts != 0 {
			t.Fatal("Redis outage spent a provider attempt")
		}
		if taskErr = recoveryRedis.Container.Start(context.Background()); taskErr != nil {
			t.Fatal(taskErr)
		}
		// Docker may allocate a fresh ephemeral host port on restart.
		port, portErr := recoveryRedis.Container.MappedPort(context.Background(), "6379")
		if portErr != nil {
			t.Fatal(portErr)
		}
		host, hostErr := recoveryRedis.Container.Host(context.Background())
		if hostErr != nil {
			t.Fatal(hostErr)
		}
		recoveryCfg.Redis.Address = net.JoinHostPort(host, port.Port())
		reconnected := redis.NewClient(&redis.Options{Addr: recoveryCfg.Redis.Address})
		defer reconnected.Close()
		waitInvitation(t, func() bool { return reconnected.Ping(context.Background()).Err() == nil })
		if _, taskErr = recoveryPG.Pool.Exec(context.Background(), "UPDATE invitation_delivery_intents SET lease_until=now()-interval '1 second' WHERE workspace_id=$1 AND id=$2", oldIntent.WorkspaceID, oldIntent.DeliveryID); taskErr != nil {
			t.Fatal(taskErr)
		}
		recoveryCfg.Invitations.ActiveKeyID = "rotated"
		recoveryCfg.Invitations.EncryptionKeys = `{"fixture":"` + base64.StdEncoding.EncodeToString(make([]byte, 32)) + `","rotated":"` + base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x7a}, 32)) + `"}`
		newIntent, _ := seedInvitationDelivery(t, recoveryPG.Pool, recoveryCfg, "new-key@example.test")
		for _, expected := range []struct {
			key   string
			claim repository.DeliveryClaim
		}{{"fixture", oldIntent}, {"rotated", newIntent}} {
			var key string
			if taskErr = recoveryPG.Pool.QueryRow(context.Background(), "SELECT key_id FROM invitation_delivery_intents WHERE workspace_id=$1 AND id=$2", expected.claim.WorkspaceID, expected.claim.DeliveryID).Scan(&key); taskErr != nil || key != expected.key {
				t.Fatal("rotation did not preserve old pending and select new write key")
			}
		}
		var calls atomic.Int32
		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			_, _ = io.WriteString(w, `{"id":"recovered"}`)
		}))
		defer provider.Close()
		restarted, workerErr := NewWorkerWithInvitationSender(context.Background(), recoveryCfg, invitationHTTPSender{endpoint: provider.URL, client: provider.Client()})
		if workerErr != nil {
			t.Fatal(workerErr)
		}
		defer func() {
			if closeErr := restarted.Close(context.Background()); closeErr != nil {
				t.Error(closeErr)
			}
		}()
		waitInvitation(t, func() bool {
			var delivered int
			readErr := recoveryPG.Pool.QueryRow(context.Background(), "SELECT count(*) FROM invitation_delivery_intents WHERE state='delivered'").Scan(&delivered)
			return readErr == nil && delivered == 2
		})
		assertDeliveryState(t, recoveryPG.Pool, oldIntent, "delivered", true)
		assertDeliveryState(t, recoveryPG.Pool, newIntent, "delivered", true)
		if calls.Load() != 2 {
			t.Fatal("lost queue recovery duplicated or lost provider sends")
		}
	})
	t.Run("workspace winning revocation fences send", func(t *testing.T) {
		claim, _ := seedInvitationDelivery(t, pg.Pool, cfg, "workspace-winner@example.test")
		claims, claimErr := store.Claim(context.Background())
		if claimErr != nil {
			t.Fatal(claimErr)
		}
		current := findInvitationClaim(t, claims, claim)
		tx, txErr := pg.Pool.Begin(context.Background())
		if txErr != nil {
			t.Fatal(txErr)
		}
		defer tx.Rollback(context.Background())
		if _, txErr = tx.Exec(context.Background(), "SELECT id FROM workspaces WHERE id=$1 FOR UPDATE", claim.WorkspaceID); txErr != nil {
			t.Fatal(txErr)
		}
		called, done := make(chan struct{}, 1), make(chan error, 1)
		go func() {
			done <- store.Deliver(context.Background(), current, func(context.Context, repository.DeliveryIntent, string, string) error {
				called <- struct{}{}
				return nil
			})
		}()
		select {
		case <-called:
			t.Fatal("send bypassed workspace-winning mutation lock")
		case <-time.After(150 * time.Millisecond):
		}
		if _, txErr = tx.Exec(context.Background(), "UPDATE invitations SET state='revoked' WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.InvitationID); txErr != nil {
			t.Fatal(txErr)
		}
		if txErr = tx.Commit(context.Background()); txErr != nil {
			t.Fatal(txErr)
		}
		select {
		case txErr = <-done:
			if txErr != nil {
				t.Fatal(txErr)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("send did not settle after winning revocation")
		}
		if len(called) != 0 {
			t.Fatal("revoked bearer reached sender")
		}
	})
	t.Run("expired uncertain provider window blocks automatic retry", func(t *testing.T) {
		claim, _ := seedInvitationDelivery(t, pg.Pool, cfg, "uncertain@example.test")
		// Legacy attempted records have no durable first-dispatch timestamp. Their
		// creation time is a conservative lower bound, never a fresh retry window.
		if _, updateErr := pg.Pool.Exec(context.Background(), "UPDATE invitation_delivery_intents SET attempts=1,created_at=now()-interval '2 days' WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID); updateErr != nil {
			t.Fatal(updateErr)
		}
		claims, claimErr := store.Claim(context.Background())
		if claimErr != nil {
			t.Fatal(claimErr)
		}
		current := findInvitationClaim(t, claims, claim)
		called := false
		if sendErr := store.Deliver(context.Background(), current, func(context.Context, repository.DeliveryIntent, string, string) error { called = true; return nil }); sendErr != nil {
			t.Fatal(sendErr)
		}
		if called {
			t.Fatal("uncertain outcome retried beyond provider idempotency window")
		}
		assertDeliveryState(t, pg.Pool, claim, "failed", false)
		var reconciliation bool
		if readErr := pg.Pool.QueryRow(context.Background(), "SELECT coalesce((to_jsonb(d)->>'reconciliation_required')::boolean,false) FROM invitation_delivery_intents d WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID).Scan(&reconciliation); readErr != nil || !reconciliation {
			t.Fatal("uncertain outcome lacks operator reconciliation fence")
		}
	})
	t.Run("terminal failed intents erase private envelopes", func(t *testing.T) {
		for _, terminal := range []string{"accepted", "revoked", "expired"} {
			t.Run(terminal, func(t *testing.T) {
				claim, _ := seedInvitationDelivery(t, pg.Pool, cfg, terminal+"-cleanup@example.test")
				if _, updateErr := pg.Pool.Exec(context.Background(), "UPDATE invitation_delivery_intents SET state='failed',attempts=8 WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID); updateErr != nil {
					t.Fatal(updateErr)
				}
				if _, updateErr := pg.Pool.Exec(context.Background(), "UPDATE invitations SET state=$3 WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.InvitationID, terminal); updateErr != nil {
					t.Fatal(updateErr)
				}
				if _, claimErr := store.Claim(context.Background()); claimErr != nil {
					t.Fatal(claimErr)
				}
				assertDeliveryState(t, pg.Pool, claim, "cancelled", true)
			})
		}
	})
	t.Run("send winning workspace lock precedes mutation", func(t *testing.T) {
		claim, _ := seedInvitationDelivery(t, pg.Pool, cfg, "send-winner@example.test")
		claims, claimErr := store.Claim(context.Background())
		if claimErr != nil {
			t.Fatal(claimErr)
		}
		current := findInvitationClaim(t, claims, claim)
		entered, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
		var once sync.Once
		defer once.Do(func() { close(release) })
		go func() {
			done <- store.Deliver(context.Background(), current, func(context.Context, repository.DeliveryIntent, string, string) error {
				close(entered)
				<-release
				return nil
			})
		}()
		select {
		case <-entered:
		case <-time.After(3 * time.Second):
			t.Fatal("send did not start")
		}
		mutationCtx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()
		_, mutationErr := pg.Pool.Exec(mutationCtx, "SELECT id FROM workspaces WHERE id=$1 FOR UPDATE", claim.WorkspaceID)
		if mutationErr == nil {
			t.Fatal("mutation passed send-winning workspace lock")
		}
		once.Do(func() { close(release) })
		if sendErr := <-done; sendErr != nil {
			t.Fatal(sendErr)
		}
		assertDeliveryState(t, pg.Pool, claim, "delivered", true)
	})
	t.Run("crash after provider acknowledgement retains first dispatch fence", func(t *testing.T) {
		claim, _ := seedInvitationDelivery(t, pg.Pool, cfg, "crash-ack@example.test")
		claims, claimErr := store.Claim(context.Background())
		if claimErr != nil {
			t.Fatal(claimErr)
		}
		current := findInvitationClaim(t, claims, claim)
		sendErr := store.Deliver(context.Background(), current, func(context.Context, repository.DeliveryIntent, string, string) error {
			// Kill only the isolated fixture's single sending transaction after the
			// external effect, before its acknowledgement can commit.
			var pid int
			if readErr := pg.Pool.QueryRow(context.Background(), "SELECT pid FROM pg_stat_activity WHERE datname=current_database() AND state='idle in transaction' AND query LIKE '%d.lease_generation=$4%' AND pid<>pg_backend_pid()").Scan(&pid); readErr != nil {
				t.Fatal(readErr)
			}
			if _, killErr := pg.Pool.Exec(context.Background(), "SELECT pg_terminate_backend($1)", pid); killErr != nil {
				t.Fatal(killErr)
			}
			return nil
		})
		if sendErr == nil {
			t.Fatal("terminated acknowledgement transaction reported success")
		}
		assertDeliveryState(t, pg.Pool, claim, "leased", false)
		var started *string
		if readErr := pg.Pool.QueryRow(context.Background(), "SELECT to_jsonb(d)->>'dispatch_started_at' FROM invitation_delivery_intents d WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID).Scan(&started); readErr != nil || started == nil {
			t.Fatal("provider dispatch fence rolled back with acknowledgement")
		}
		if _, updateErr := pg.Pool.Exec(context.Background(), "UPDATE invitation_delivery_intents SET lease_until=now()-interval '1 second' WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID); updateErr != nil {
			t.Fatal(updateErr)
		}
		claims, claimErr = store.Claim(context.Background())
		if claimErr != nil {
			t.Fatal(claimErr)
		}
		current = findInvitationClaim(t, claims, claim)
		if retryErr := store.Deliver(context.Background(), current, func(context.Context, repository.DeliveryIntent, string, string) error { return nil }); retryErr != nil {
			t.Fatal(retryErr)
		}
		var retried *string
		if readErr := pg.Pool.QueryRow(context.Background(), "SELECT to_jsonb(d)->>'dispatch_started_at' FROM invitation_delivery_intents d WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID).Scan(&retried); readErr != nil || retried == nil || *retried != *started {
			t.Fatal("recovery renewed the original dispatch window")
		}
		assertDeliveryState(t, pg.Pool, claim, "delivered", true)
	})
	t.Run("actual Redis malformed future and tampered references cannot send", func(t *testing.T) {
		claim, _ := seedInvitationDelivery(t, pg.Pool, cfg, "tampered@example.test")
		claims, err8 := store.Claim(context.Background())
		if err8 != nil {
			t.Fatal(err8)
		}
		current := findInvitationClaim(t, claims, claim)
		_, err8 = pg.Pool.Exec(context.Background(), "UPDATE invitation_delivery_intents SET ciphertext=set_byte(ciphertext,0,get_byte(ciphertext,0)#1) WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID)
		if err8 != nil {
			t.Fatal(err8)
		}
		producer := asynq.NewClientFromRedisClient(queue.Client)
		future := current
		future.Generation++
		futureTask, err8 := job.NewInvitationTask(future)
		if err8 != nil {
			t.Fatal(err8)
		}
		_, err8 = producer.Enqueue(futureTask, asynq.TaskID("future-reference"), asynq.Retention(time.Minute))
		if err8 != nil {
			t.Fatal(err8)
		}
		foreign, _ := seedInvitationDelivery(t, pg.Pool, cfg, "foreign-tuple@example.test")
		foreignClaims, foreignErr := store.Claim(context.Background())
		if foreignErr != nil {
			t.Fatal(foreignErr)
		}
		foreignLease := findInvitationClaim(t, foreignClaims, foreign)
		if _, foreignErr = pg.Pool.Exec(context.Background(), "UPDATE invitation_delivery_intents SET lease_until=now()-interval '1 second' WHERE workspace_id=$1 AND id=$2", foreign.WorkspaceID, foreign.DeliveryID); foreignErr != nil {
			t.Fatal(foreignErr)
		}
		if _, foreignErr = store.Claim(context.Background()); foreignErr != nil {
			t.Fatal(foreignErr)
		}
		staleTask, staleErr := job.NewInvitationTask(foreignLease)
		if staleErr != nil {
			t.Fatal(staleErr)
		}
		if _, staleErr = producer.Enqueue(staleTask, asynq.TaskID("old-generation-reference"), asynq.Retention(time.Minute)); staleErr != nil {
			t.Fatal(staleErr)
		}
		for _, terminal := range []string{"revoked", "accepted", "expired", "delivered"} {
			terminalIntent, _ := seedInvitationDelivery(t, pg.Pool, cfg, terminal+"-redis@example.test")
			terminalClaims, terminalErr := store.Claim(context.Background())
			if terminalErr != nil {
				t.Fatal(terminalErr)
			}
			terminalLease := findInvitationClaim(t, terminalClaims, terminalIntent)
			if terminal == "delivered" {
				terminalErr = store.Deliver(context.Background(), terminalLease, func(context.Context, repository.DeliveryIntent, string, string) error { return nil })
			} else {
				_, terminalErr = pg.Pool.Exec(context.Background(), "UPDATE invitations SET state=$3 WHERE workspace_id=$1 AND id=$2", terminalIntent.WorkspaceID, terminalIntent.InvitationID, terminal)
			}
			if terminalErr != nil {
				t.Fatal(terminalErr)
			}
			terminalTask, terminalErr := job.NewInvitationTask(terminalLease)
			if terminalErr != nil {
				t.Fatal(terminalErr)
			}
			if _, terminalErr = producer.Enqueue(terminalTask, asynq.TaskID(terminal+"-reference"), asynq.Retention(time.Minute)); terminalErr != nil {
				t.Fatal(terminalErr)
			}
		}
		for index, forged := range []repository.DeliveryClaim{
			{WorkspaceID: foreignLease.WorkspaceID, InvitationID: current.InvitationID, DeliveryID: current.DeliveryID, Generation: current.Generation},
			{WorkspaceID: current.WorkspaceID, InvitationID: foreignLease.InvitationID, DeliveryID: current.DeliveryID, Generation: current.Generation},
			{WorkspaceID: current.WorkspaceID, InvitationID: current.InvitationID, DeliveryID: foreignLease.DeliveryID, Generation: current.Generation},
		} {
			forgedTask, buildErr := job.NewInvitationTask(forged)
			if buildErr != nil {
				t.Fatal(buildErr)
			}
			if _, buildErr = producer.Enqueue(forgedTask, asynq.TaskID("forged-tuple-"+string(rune('0'+index))), asynq.Retention(time.Minute)); buildErr != nil {
				t.Fatal(buildErr)
			}
		}
		malformed := asynq.NewTask(job.TaskInvitation, []byte(`{"email":"private-input-canary","generation":1}`), asynq.MaxRetry(0))
		_, err8 = producer.Enqueue(malformed, asynq.TaskID("malformed-reference"))
		if err8 != nil {
			t.Fatal(err8)
		}
		task, err8 := job.NewInvitationTask(current)
		if err8 != nil {
			t.Fatal(err8)
		}
		_, err8 = producer.Enqueue(task, asynq.TaskID("tampered-reference"))
		if err8 != nil {
			t.Fatal(err8)
		}
		var calls atomic.Int32
		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer provider.Close()
		worker9, err8 := NewWorkerWithInvitationSender(context.Background(), cfg, invitationHTTPSender{endpoint: provider.URL, client: provider.Client()})
		if err8 != nil {
			t.Fatal(err8)
		}
		defer func() {
			if err10 := worker9.Close(context.Background()); err10 != nil {
				t.Error(err10)
			}
		}()
		waitInvitation(t, func() bool {
			var state string
			_ = pg.Pool.QueryRow(context.Background(), "SELECT state FROM invitation_delivery_intents WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID).Scan(&state)
			return state == "queued"
		})
		inspector := asynq.NewInspectorFromRedisClient(queue.Client)
		waitInvitation(t, func() bool {
			info, err11 := inspector.GetTaskInfo("default", "malformed-reference")
			return err11 == nil && info.State == asynq.TaskStateArchived
		})
		for index := range 3 {
			waitInvitation(t, func() bool {
				info, readErr := inspector.GetTaskInfo("default", "forged-tuple-"+string(rune('0'+index)))
				return readErr == nil && info.State == asynq.TaskStateCompleted
			})
		}
		for _, id := range []string{"future-reference", "old-generation-reference", "revoked-reference", "accepted-reference", "expired-reference", "delivered-reference"} {
			waitInvitation(t, func() bool {
				info, readErr := inspector.GetTaskInfo("default", id)
				return readErr == nil && info.State == asynq.TaskStateCompleted
			})
		}
		info, readErr := inspector.GetTaskInfo("default", "malformed-reference")
		if readErr != nil || strings.Contains(info.LastErr, "private-input-canary") || strings.Contains(info.LastErr, "private-provider-response-canary") {
			t.Fatal("private input persisted as retry error")
		}
		if calls.Load() != 0 {
			t.Fatal("unsafe Redis or ciphertext reached provider")
		}
		assertDeliveryState(t, pg.Pool, claim, "queued", false)
	})
	t.Run("concurrent duplicate lease sends once", func(t *testing.T) {
		claim, _ := seedInvitationDelivery(t, pg.Pool, cfg, "concurrent@example.test")
		claims, err12 := store.Claim(context.Background())
		if err12 != nil {
			t.Fatal(err12)
		}
		current := findInvitationClaim(t, claims, claim)
		var calls atomic.Int32
		var group sync.WaitGroup
		for range 2 {
			group.Go(func() {
				err13 := store.Deliver(context.Background(), current, func(context.Context, repository.DeliveryIntent, string, string) error {
					calls.Add(1)
					time.Sleep(50 * time.Millisecond)
					return nil
				})
				if err13 != nil {
					t.Error(err13)
				}
			})
		}
		group.Wait()
		if calls.Load() != 1 {
			t.Fatal("duplicate concurrent sends")
		}
		assertDeliveryState(t, pg.Pool, claim, "delivered", true)
	})
	t.Run("bounded failure retry and terminal failure", func(t *testing.T) {
		claim, _ := seedInvitationDelivery(t, pg.Pool, cfg, "failure@example.test")
		for attempt := 1; attempt <= 8; attempt++ {
			claims, err14 := store.Claim(context.Background())
			if err14 != nil {
				t.Fatal(err14)
			}
			current := findInvitationClaim(t, claims, claim)
			if err14 = store.Deliver(context.Background(), current, func(context.Context, repository.DeliveryIntent, string, string) error {
				return errors.New("private-provider-response-canary")
			}); err14 != nil {
				t.Fatal(err14)
			}
			state := "queued"
			if attempt == 8 {
				state = "failed"
			}
			assertDeliveryState(t, pg.Pool, claim, state, false)
			var got int
			var future bool
			if err14 = pg.Pool.QueryRow(context.Background(), "SELECT attempts,available_at>clock_timestamp() FROM invitation_delivery_intents WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID).Scan(&got, &future); err14 != nil || got != attempt || !future {
				t.Fatal("bounded retry schedule missing")
			}
			_, err14 = pg.Pool.Exec(context.Background(), "UPDATE invitation_delivery_intents SET available_at=now() WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID)
			if err14 != nil {
				t.Fatal(err14)
			}
		}
	})
	t.Run("expiry revocation and cross-tenant fences", func(t *testing.T) {
		for _, mode := range []string{"expired", "revoked", "accepted", "foreign", "future", "stale", "malformed", "foreign-delivery"} {
			t.Run(mode, func(t *testing.T) {
				claim, _ := seedInvitationDelivery(t, pg.Pool, cfg, mode+"@example.test")
				claims, err15 := store.Claim(context.Background())
				if err15 != nil {
					t.Fatal(err15)
				}
				current := findInvitationClaim(t, claims, claim)
				switch mode {
				case "expired":
					_, err15 = pg.Pool.Exec(context.Background(), "UPDATE invitations SET created_at=now()-interval '8 days',expires_at=now()-interval '1 day' WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.InvitationID)
				case "revoked":
					_, err15 = pg.Pool.Exec(context.Background(), "UPDATE invitations SET state='revoked' WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.InvitationID)
				case "accepted":
					_, err15 = pg.Pool.Exec(context.Background(), "UPDATE invitations SET state='accepted' WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.InvitationID)
				case "foreign-delivery":
					current.DeliveryID = uuid.New()
				case "foreign":
					current.WorkspaceID = uuid.New()
				case "future":
					current.Generation++
				case "stale":
					_, err15 = pg.Pool.Exec(context.Background(), "UPDATE invitation_delivery_intents SET lease_until=now()-interval '1 second' WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID)
				case "malformed":
					current.InvitationID = uuid.New()
				}
				if err15 != nil {
					t.Fatal(err15)
				}
				sent := false
				if err15 = store.Deliver(context.Background(), current, func(context.Context, repository.DeliveryIntent, string, string) error { sent = true; return nil }); err15 != nil || sent {
					t.Fatal("unsafe reference reached sender")
				}
			})
		}
	})
	t.Run("durable publication crash and lease recovery", func(t *testing.T) {
		claim, _ := seedInvitationDelivery(t, pg.Pool, cfg, "recovery@example.test")
		first, err16 := store.Claim(context.Background())
		if err16 != nil {
			t.Fatal(err16)
		}
		old := findInvitationClaim(t, first, claim)
		// Simulate death after SQL claim and before publishing anything to Redis.
		_, err16 = pg.Pool.Exec(context.Background(), "UPDATE invitation_delivery_intents SET lease_until=now()-interval '1 second' WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID)
		if err16 != nil {
			t.Fatal(err16)
		}
		next, err16 := store.Claim(context.Background())
		if err16 != nil {
			t.Fatal(err16)
		}
		current := findInvitationClaim(t, next, claim)
		if current.Generation <= old.Generation {
			t.Fatal("recovery did not fence previous lease")
		}
		sent := 0
		send := func(context.Context, repository.DeliveryIntent, string, string) error { sent++; return nil }
		if err16 = store.Deliver(context.Background(), old, send); err16 != nil || sent != 0 {
			t.Fatal("old lease sent")
		}
		if err16 = store.Deliver(context.Background(), current, send); err16 != nil {
			t.Fatal(err16)
		}
		if err16 = store.Deliver(context.Background(), current, send); err16 != nil || sent != 1 {
			t.Fatal("duplicate lease sent")
		}
		assertDeliveryState(t, pg.Pool, claim, "delivered", true)
	})
	t.Run("cancelled send stays retryable and stale acknowledgement cannot deliver", func(t *testing.T) {
		claim, _ := seedInvitationDelivery(t, pg.Pool, cfg, "cancel@example.test")
		claims, err17 := store.Claim(context.Background())
		if err17 != nil {
			t.Fatal(err17)
		}
		current := findInvitationClaim(t, claims, claim)
		ctx, cancel := context.WithCancel(context.Background())
		err17 = store.Deliver(ctx, current, func(ctx context.Context, _ repository.DeliveryIntent, _, _ string) error { cancel(); return ctx.Err() })
		if err17 != nil {
			t.Fatal(err17)
		}
		assertDeliveryState(t, pg.Pool, claim, "queued", false)
		claim, _ = seedInvitationDelivery(t, pg.Pool, cfg, "late@example.test")
		claims, err17 = store.Claim(context.Background())
		if err17 != nil {
			t.Fatal(err17)
		}
		current = findInvitationClaim(t, claims, claim)
		// SQL clock expiry is checked again for acknowledgement, not only before send.
		_, err17 = pg.Pool.Exec(context.Background(), "UPDATE invitation_delivery_intents SET lease_until=now()+interval '100 milliseconds' WHERE workspace_id=$1 AND id=$2", claim.WorkspaceID, claim.DeliveryID)
		if err17 != nil {
			t.Fatal(err17)
		}
		err17 = store.Deliver(context.Background(), current, func(context.Context, repository.DeliveryIntent, string, string) error {
			time.Sleep(150 * time.Millisecond)
			return nil
		})
		if err17 != nil {
			t.Fatal(err17)
		}
		assertDeliveryState(t, pg.Pool, claim, "leased", false)
	})
}

//nolint:lll // Keep scoped SQL fences, bounded validation and exact fixture assertions together.
func seedInvitationDelivery(t *testing.T, pool *pgxpool.Pool, cfg *config.Config, address string) (repository.DeliveryClaim, string) {
	t.Helper()
	ctx := context.Background()
	actor, workspace, invite, delivery := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	_, err18 := pool.Exec(ctx, "INSERT INTO users(id,issuer,subject,verified_email) VALUES($1,'https://fixture.clerk.accounts.dev',$2,'fixture@example.test')", actor, "fixture_"+actor.String())
	if err18 != nil {
		t.Fatal(err18)
	}
	_, err18 = pool.Exec(ctx, "INSERT INTO workspaces(id,name,created_by) VALUES($1,'<script>private-name</script>',$2)", workspace, actor)
	if err18 != nil {
		t.Fatal(err18)
	}
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x42}, 32))
	digest := sha256.Sum256([]byte(token))
	_, err18 = pool.Exec(ctx, "INSERT INTO invitations(workspace_id,id,inviter_id,email,role,token_digest) VALUES($1,$2,$3,$4,'member',$5)", workspace, invite, actor, address, digest[:])
	if err18 != nil {
		t.Fatal(err18)
	}
	keys, err18 := cfg.Invitations.Keys()
	if err18 != nil {
		t.Fatal(err18)
	}
	aead, err18 := invitationcrypto.Cipher(keys[cfg.Invitations.ActiveKeyID])
	if err18 != nil {
		t.Fatal(err18)
	}
	raw, err18 := json.Marshal(invitationcrypto.Envelope{Token: token, Email: address, Sender: cfg.Invitations.Sender, PublicOrigin: cfg.Invitations.PublicOrigin})
	if err18 != nil {
		t.Fatal(err18)
	}
	cipher := aead.Seal(nil, nil, raw, invitationcrypto.AAD(workspace, invite, delivery, cfg.Invitations.ActiveKeyID))
	clear(raw)
	_, err18 = pool.Exec(ctx, "INSERT INTO invitation_delivery_intents(workspace_id,id,invitation_id,key_id,ciphertext) VALUES($1,$2,$3,$4,$5)", workspace, delivery, invite, cfg.Invitations.ActiveKeyID, cipher)
	if err18 != nil {
		t.Fatal(err18)
	}
	return repository.DeliveryClaim{WorkspaceID: workspace, InvitationID: invite, DeliveryID: delivery}, token
}

//nolint:lll // Keep scoped SQL fences, bounded validation and exact fixture assertions together.
func findInvitationClaim(t *testing.T, claims []repository.DeliveryClaim, want repository.DeliveryClaim) repository.DeliveryClaim {
	t.Helper()
	for _, c := range claims {
		if c.WorkspaceID == want.WorkspaceID && c.DeliveryID == want.DeliveryID {
			return c
		}
	}
	t.Fatal("durable intent not claimed")
	return repository.DeliveryClaim{}
}

//nolint:lll // Keep scoped SQL fences, bounded validation and exact fixture assertions together.
func assertDeliveryState(t *testing.T, pool *pgxpool.Pool, c repository.DeliveryClaim, want string, erased bool) {
	t.Helper()
	var state string
	var empty bool
	if err19 := pool.QueryRow(context.Background(), "SELECT state,ciphertext IS NULL FROM invitation_delivery_intents WHERE workspace_id=$1 AND id=$2", c.WorkspaceID, c.DeliveryID).Scan(&state, &empty); err19 != nil || state != want || empty != erased {
		t.Fatalf("delivery state %s erasure %v, want %s/%v", state, empty, want, erased)
	}
}
func waitInvitation(t *testing.T, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		if ready() {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("durable invitation acknowledgement missing")
}

type invitationHTTPSender struct {
	client   *http.Client
	endpoint string
}

//nolint:lll // Keep scoped SQL fences, bounded validation and exact fixture assertions together.
func (s invitationHTTPSender) SendInvitation(ctx context.Context, message invitationcrypto.Message) error {
	payload, err20 := json.Marshal(map[string]string{"to": message.To, "from": message.From, "subject": message.Subject, "html": message.HTML})
	if err20 != nil {
		return err20
	}
	req, err20 := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint+"/emails", bytes.NewReader(payload))
	if err20 != nil {
		return err20
	}
	req.Header.Set("Idempotency-Key", message.IdempotencyKey)
	response, err20 := s.client.Do(req)
	if err20 != nil {
		return errors.New("private-provider-response-canary")
	}
	defer response.Body.Close()
	var ack struct {
		ID string `json:"id"`
	}
	if response.StatusCode != http.StatusOK || json.NewDecoder(response.Body).Decode(&ack) != nil || ack.ID == "" {
		return errors.New("private-provider-response-canary")
	}
	return nil
}

type invitationLogBuffer struct {
	body bytes.Buffer
	mu   sync.Mutex
}

func (b *invitationLogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.body.Write(p)
}
func (b *invitationLogBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.body.String()
}
