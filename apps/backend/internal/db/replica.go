// Package db provides database read-replica query routing, cluster connection pooling, and L1/L2 cache invalidation pub/sub.
package db

import (
	"sync"
	"sync/atomic"
)

// MockPool represents a database connection pool endpoint.
type MockPool struct {
	Name string
}

// DBCluster manages connection routing between primary master database and read replicas.
type DBCluster struct {
	Primary  *MockPool
	Replicas []*MockPool
	counter  uint64
}

// NewCluster initializes a DBCluster with primary master and optional read-replicas.
func NewCluster(primary *MockPool, replicas []*MockPool) *DBCluster {
	return &DBCluster{
		Primary:  primary,
		Replicas: replicas,
	}
}

// GetWritePool always returns the primary master pool for write queries.
func (c *DBCluster) GetWritePool() *MockPool {
	return c.Primary
}

// GetReadPool returns a read replica using round-robin distribution, falling back to primary master if no replicas exist.
func (c *DBCluster) GetReadPool() *MockPool {
	if len(c.Replicas) == 0 {
		return c.Primary
	}

	idx := atomic.AddUint64(&c.counter, 1) - 1
	return c.Replicas[idx%uint64(len(c.Replicas))]
}

// InvalidationHandler is a function type for receiving key invalidation events.
type InvalidationHandler func(key string)

// CacheInvalidator provides multi-node cache invalidation pub/sub synchronization.
type CacheInvalidator struct {
	mu          sync.RWMutex
	subscribers []InvalidationHandler
}

// NewCacheInvalidator creates a new CacheInvalidator instance.
func NewCacheInvalidator() *CacheInvalidator {
	return &CacheInvalidator{
		subscribers: make([]InvalidationHandler, 0),
	}
}

// Subscribe registers a listener function for cache invalidation events.
func (ci *CacheInvalidator) Subscribe(handler InvalidationHandler) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ci.subscribers = append(ci.subscribers, handler)
}

// Invalidate broadcasts a key invalidation event to all subscribed L1/L2 nodes.
func (ci *CacheInvalidator) Invalidate(key string) {
	ci.mu.RLock()
	defer ci.mu.RUnlock()
	for _, sub := range ci.subscribers {
		sub(key)
	}
}
