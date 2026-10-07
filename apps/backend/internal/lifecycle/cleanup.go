// Package lifecycle runs ordered resource cleanup within one caller-owned deadline.
package lifecycle

import (
	"context"
	"errors"
	"sync"
)

// ErrCleanupClosed rejects ownership registration after shutdown begins.
var ErrCleanupClosed = errors.New("cleanup already closed")

// ResourceError identifies a failed closer while preserving its private cause.
type ResourceError struct {
	cause error
	Name  string
}

func (e *ResourceError) Error() string { return "cleanup failed: " + e.Name }
func (e *ResourceError) Unwrap() error { return e.cause }

type cleanupEntry struct {
	close func(context.Context) error
	name  string
}

// Cleanup owns registered resources and closes them in reverse order once.
// A Cleanup must not be copied after first use.
type Cleanup struct {
	err     error
	entries []cleanupEntry
	once    sync.Once
	mu      sync.Mutex
	closed  bool
}

// Push registers ownership immediately. Names must be stable resource labels,
// without credentials, addresses, or other externally supplied diagnostics.
func (c *Cleanup) Push(name string, cleanup func(context.Context) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrCleanupClosed
	}
	c.entries = append(c.entries, cleanupEntry{name: name, close: cleanup})
	return nil
}

// Close attempts every closer, joins named failures, and returns the same result
// on repeated calls. Callbacks receive the caller's shutdown context.
func (c *Cleanup) Close(ctx context.Context) error {
	c.once.Do(func() {
		c.mu.Lock()
		c.closed = true
		entries := c.entries
		c.entries = nil
		c.mu.Unlock()
		var failures []error
		for i := len(entries) - 1; i >= 0; i-- {
			if err := entries[i].close(ctx); err != nil {
				failures = append(failures, &ResourceError{Name: entries[i].name, cause: err})
			}
		}
		c.err = errors.Join(failures...)
	})
	return c.err
}
