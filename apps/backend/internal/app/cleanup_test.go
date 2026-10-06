package app

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestCleanupReverseAllAndIdempotent(t *testing.T) {
	var cleanup Cleanup
	var order []string
	first := errors.New("SECRET-MARKER first")
	last := errors.New("SECRET-MARKER last")
	for _, entry := range []struct {
		name string
		err  error
	}{{"database", first}, {"redis", nil}, {"worker", last}} {
		cleanup.Push(entry.name, func(ctx context.Context) error {
			if ctx.Value("marker") != "kept" {
				t.Error("context not passed to closer")
			}
			order = append(order, entry.name)
			return entry.err
		})
	}
	err := cleanup.Close(context.WithValue(context.Background(), "marker", "kept"))
	if !reflect.DeepEqual(order, []string{"worker", "redis", "database"}) {
		t.Fatalf("wrong cleanup order: %v", order)
	}
	if !errors.Is(err, first) || !errors.Is(err, last) {
		t.Fatalf("cleanup lost errors: %v", err)
	}
	if !strings.Contains(err.Error(), "database") || !strings.Contains(err.Error(), "worker") || strings.Contains(err.Error(), "SECRET-MARKER") {
		t.Fatalf("unsafe or missing diagnostic: %v", err)
	}
	if again := cleanup.Close(context.Background()); again != err || len(order) != 3 {
		t.Fatal("repeat close changed result or repeated effects")
	}
}

func TestCleanupConcurrentClose(t *testing.T) {
	var cleanup Cleanup
	var calls atomic.Int32
	cleanup.Push("resource", func(context.Context) error { calls.Add(1); return nil })
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			if err := cleanup.Close(context.Background()); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("closer called %d times", calls.Load())
	}
}
