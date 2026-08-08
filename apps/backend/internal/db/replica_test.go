package db_test

import (
	"testing"

	"flux/apps/backend/internal/db"
)

func TestDBCluster_RoundRobinReadRouting(t *testing.T) {
	primary := &db.MockPool{Name: "primary-master"}
	replica1 := &db.MockPool{Name: "read-replica-1"}
	replica2 := &db.MockPool{Name: "read-replica-2"}

	cluster := db.NewCluster(primary, []*db.MockPool{replica1, replica2})

	// Writes must always go to primary
	if w := cluster.GetWritePool(); w.Name != "primary-master" {
		t.Errorf("expected write pool 'primary-master', got %q", w.Name)
	}

	// Reads must round-robin between replica1 and replica2
	r1 := cluster.GetReadPool()
	r2 := cluster.GetReadPool()
	r3 := cluster.GetReadPool()

	if r1.Name != "read-replica-1" {
		t.Errorf("expected first read to pick replica-1, got %q", r1.Name)
	}
	if r2.Name != "read-replica-2" {
		t.Errorf("expected second read to pick replica-2, got %q", r2.Name)
	}
	if r3.Name != "read-replica-1" {
		t.Errorf("expected third read to round-robin back to replica-1, got %q", r3.Name)
	}
}

func TestDBCluster_FallbackToPrimaryWhenNoReplicas(t *testing.T) {
	primary := &db.MockPool{Name: "primary-master"}
	cluster := db.NewCluster(primary, nil)

	r := cluster.GetReadPool()
	if r.Name != "primary-master" {
		t.Errorf("expected fallback to primary when no replicas configured, got %q", r.Name)
	}
}

func TestCacheInvalidator_PubSubInvalidation(t *testing.T) {
	invalidator := db.NewCacheInvalidator()
	invalidatedKeys := make([]string, 0)

	invalidator.Subscribe(func(key string) {
		invalidatedKeys = append(invalidatedKeys, key)
	})

	invalidator.Invalidate("link:xyz123")
	invalidator.Invalidate("link:abc999")

	if len(invalidatedKeys) != 2 {
		t.Fatalf("expected 2 invalidated keys, got %d", len(invalidatedKeys))
	}
	if invalidatedKeys[0] != "link:xyz123" || invalidatedKeys[1] != "link:abc999" {
		t.Errorf("unexpected invalidated keys: %v", invalidatedKeys)
	}
}
