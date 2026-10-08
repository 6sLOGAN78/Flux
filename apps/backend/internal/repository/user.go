package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User is an internal identity with a provider-verified profile snapshot.
type User struct {
	Email string
	ID    uuid.UUID
}

const identityRollbackTimeout = 5 * time.Second

// UserRepository owns parameterized PostgreSQL identity mapping.
type UserRepository struct{ pool *pgxpool.Pool }

// NewUserRepository injects the authoritative transactional store.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository { return &UserRepository{pool: pool} }

// Find resolves only the exact issuer/subject pair; email supplies no authority.
func (r *UserRepository) Find(ctx context.Context, issuer, subject string) (User, error) {
	var result User
	if r == nil || r.pool == nil {
		return result, errors.New("identity store unavailable")
	}
	err := r.pool.QueryRow(ctx,
		"SELECT id, verified_email FROM users WHERE issuer=$1 AND subject=$2", issuer, subject).
		Scan(&result.ID, &result.Email)
	return result, err
}

// Create returns only a committed row. A separate READ COMMITTED select after
// ON CONFLICT waits for a concurrent creator and sees that creator's snapshot.
//
//nolint:nonamedreturns // Deferred rollback joins its failure while keeping canceled-request cleanup bounded.
func (r *UserRepository) Create(ctx context.Context, issuer, subject, email string) (result User, err error) {
	if r == nil || r.pool == nil {
		return result, errors.New("identity store unavailable")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return result, fmt.Errorf("begin identity: %w", err)
	}
	defer func() {
		closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), identityRollbackTimeout)
		defer cancel()
		if rollbackErr := tx.Rollback(closeCtx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			err = errors.Join(err, fmt.Errorf("rollback identity: %w", rollbackErr))
		}
	}()
	_, err = tx.Exec(ctx, "INSERT INTO users (issuer, subject, verified_email) VALUES ($1,$2,$3) "+
		"ON CONFLICT (issuer,subject) DO NOTHING", issuer, subject, email)
	if err != nil {
		return result, fmt.Errorf("insert identity: %w", err)
	}
	err = tx.QueryRow(ctx, "SELECT id, verified_email FROM users WHERE issuer=$1 AND subject=$2", issuer, subject).
		Scan(&result.ID, &result.Email)
	if err != nil {
		return User{}, fmt.Errorf("read identity: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit identity: %w", err)
	}
	return result, nil
}
