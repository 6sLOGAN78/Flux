package app

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
)

// NotifyContext supplies the same supervisor signals to every long-running role.
func NotifyContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
}

// Lifecycle stops intake, drains work and releases ownership under one deadline.
// Its zero value is ready. Shutdown must never close a shared dependency while
// an uncooperative dependent is still draining: the serial pipeline continues
// after the bounded caller returns, and the command decides the process exit.
type Lifecycle struct {
	unready atomic.Bool
	once    sync.Once
	err     error
	cleanup *Cleanup
	stop    func()
	drain   func(context.Context) error
}

func (l *Lifecycle) Ready() bool { return !l.unready.Load() }

type shutdownError struct {
	stage string
	cause error
}

func (e *shutdownError) Error() string { return "shutdown failed: " + e.stage }
func (e *shutdownError) Unwrap() error { return e.cause }

func (l *Lifecycle) Shutdown(ctx context.Context) error {
	l.once.Do(func() {
		l.unready.Store(true)
		result := make(chan error, 1)
		go func() {
			if l.stop != nil {
				l.stop()
			}
			var failures []error
			if l.drain != nil {
				if err := l.drain(ctx); err != nil {
					failures = append(failures, &shutdownError{stage: "drain", cause: err})
				}
			}
			if l.cleanup != nil {
				failures = append(failures, l.cleanup.Close(ctx))
			}
			result <- errors.Join(failures...)
		}()
		select {
		case l.err = <-result:
		case <-ctx.Done():
			l.err = &shutdownError{stage: "deadline", cause: ctx.Err()}
		}
	})
	return l.err
}
