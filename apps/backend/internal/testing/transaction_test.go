//nolint:testpackage // These unit tests exercise the private rollback cleanup boundary without a database.
package testing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type rollbackTestTx struct {
	pgx.Tx
	rollback func(context.Context) error
}

func (tx rollbackTestTx) Rollback(ctx context.Context) error { return tx.rollback(ctx) }

type rollbackContextKey struct{}

func TestRollbackCleanupSurvivesCanceledCallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), rollbackContextKey{}, "kept"))
	cancel()
	called := false
	tx := rollbackTestTx{rollback: func(cleanupCtx context.Context) error {
		called = true
		if cleanupCtx.Err() != nil || cleanupCtx.Value(rollbackContextKey{}) != "kept" {
			t.Fatal("rollback lost cleanup context or inherited callback cancellation")
		}
		deadline, ok := cleanupCtx.Deadline()
		if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > transactionRollbackTimeout {
			t.Fatal("rollback cleanup has no bounded independent deadline")
		}
		return nil
	}}
	if err := rollbackTransaction(ctx, tx); err != nil || !called {
		t.Fatalf("rollback did not run successfully: %v", err)
	}
}

func TestRollbackCleanupRetainsFailureCause(t *testing.T) {
	cause := errors.New("rollback failure")
	tx := rollbackTestTx{rollback: func(context.Context) error { return cause }}
	if err := rollbackTransaction(context.Background(), tx); !errors.Is(err, cause) {
		t.Fatalf("rollback failure cause was lost: %v", err)
	}
}

func TestRollbackCleanupAcceptsAlreadyClosedTransaction(t *testing.T) {
	tx := rollbackTestTx{rollback: func(context.Context) error { return pgx.ErrTxClosed }}
	if err := rollbackTransaction(context.Background(), tx); err != nil {
		t.Fatalf("committed or previously rolled-back transaction reported failure: %v", err)
	}
}
