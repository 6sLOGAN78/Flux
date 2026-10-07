// Package testing provides pinned integration containers and transaction assertions.
package testing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// TxFn represents a function that executes within a transaction
type TxFn func(tx pgx.Tx) error

// WithTransaction runs a callback in a transaction and commits successful callbacks.
// It rolls back when the callback fails or the transaction cannot be committed.
func WithTransaction(ctx context.Context, db *TestDB, fn TxFn) (resultErr error) {
	// Begin transaction
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback happens if commit doesn't occur
	defer func() { resultErr = errors.Join(resultErr, rollbackTransaction(ctx, tx)) }()

	// Run the function within the transaction
	if err26 := fn(tx); err26 != nil {
		return err26
	}

	// Transaction was successful, commit it
	if err31 := tx.Commit(ctx); err31 != nil {
		return fmt.Errorf("failed to commit transaction: %w", err31)
	}

	return nil
}

// WithRollbackTransaction runs a function within a transaction and always rolls it back
// Useful for tests where you want to execute operations but never persist them
func WithRollbackTransaction(ctx context.Context, db *TestDB, fn TxFn) (resultErr error) {
	// Begin transaction
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Always rollback at the end
	defer func() { resultErr = errors.Join(resultErr, rollbackTransaction(ctx, tx)) }()

	// Run the function within the transaction
	return fn(tx)
}

const transactionRollbackTimeout = 5 * time.Second

func rollbackTransaction(ctx context.Context, tx pgx.Tx) error {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), transactionRollbackTimeout)
	defer cancel()
	if err := tx.Rollback(cleanupCtx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return fmt.Errorf("failed to roll back transaction: %w", err)
	}
	return nil
}
