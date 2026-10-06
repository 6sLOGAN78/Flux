package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

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
			if got != ctx { t.Error("cleanup received a fresh deadline") }
			events = append(events, name)
			return cause
		})
	}
	l.stop = func() { if l.Ready() { t.Error("intake stopped before unready") }; events = append(events, "stop") }
	l.drain = func(got context.Context) error { if got != ctx { t.Error("drain received a fresh deadline") }; events = append(events, "drain"); return cause }
	err := l.Shutdown(ctx)
	if !errors.Is(err, cause) || strings.Contains(err.Error(), "SECRET-MARKER") { t.Fatalf("unsafe/lost errors: %v", err) }
	if !reflect.DeepEqual(events, []string{"stop", "drain", "dependent", "shared"}) { t.Fatalf("order: %v", events) }
	if !strings.Contains(err.Error(), "drain") || !strings.Contains(err.Error(), "dependent") || !strings.Contains(err.Error(), "shared") { t.Fatalf("missing stage: %v", err) }
	if l.Ready() { t.Fatal("shutdown stayed ready") }
	if again := l.Shutdown(context.Background()); again != err { t.Fatal("shutdown result changed") }
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
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > 250*time.Millisecond { t.Fatalf("unbounded shutdown: %v", err) }
	<-blocked
	select { case <-later: t.Fatal("shared resource closed under active dependent"); default: }
	close(release)
	select { case <-later: case <-time.After(time.Second): t.Fatal("later closer skipped") }
}

func TestLifecycleReadinessGate(t *testing.T) {
	role, err := NewRedirector(context.Background(), roleTestConfig())
	if err != nil { t.Fatal(err) }
	defer role.Close(context.Background())
	if err := role.Close(context.Background()); err != nil { t.Fatal(err) }
	response := httptest.NewRecorder()
	role.HTTP.ServeHTTP(response, httptest.NewRequest("GET", "/ready", nil))
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), `"not_ready"`) { t.Fatalf("shutdown readiness %d: %s", response.Code, response.Body) }
}
