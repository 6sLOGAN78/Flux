package app

import (
	"context"
	"errors"
	"sync"
)

// ErrCleanupClosed rejects ownership registration after shutdown begins.
var ErrCleanupClosed = errors.New("cleanup already closed")

// ResourceError identifies a failed closer while preserving its private cause.
type ResourceError struct {
	Name  string
	cause error
}

func (e *ResourceError) Error() string { return "cleanup failed: " + e.Name }
func (e *ResourceError) Unwrap() error { return e.cause }

type cleanupEntry struct {
	name  string
	close func(context.Context) error
}

// Cleanup owns registered resources and closes them in reverse order once.
// A Cleanup must not be copied after first use.
type Cleanup struct {
	mu      sync.Mutex
	once    sync.Once
	entries []cleanupEntry
	closed  bool
	err     error
}

// Push registers ownership immediately. Names must be stable resource labels,
// without credentials, addresses, or other externally supplied diagnostics.
func (c *Cleanup) Push(name string, close func(context.Context) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrCleanupClosed
	}
	c.entries = append(c.entries, cleanupEntry{name: name, close: close})
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
