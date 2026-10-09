package repository

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Scope requires both durable actor identity and workspace identity.
type Scope struct{ WorkspaceID, ActorID uuid.UUID }

// Workspace is a freshly authorized workspace and membership snapshot.
type Workspace struct {
	Name, Role string
	ID         uuid.UUID
}

// WorkspaceRepository owns tenant-scoped transactional access.
type WorkspaceRepository struct{ pool *pgxpool.Pool }

// NewWorkspaceRepository injects PostgreSQL as the authoritative store.
func NewWorkspaceRepository(pool *pgxpool.Pool) *WorkspaceRepository {
	return &WorkspaceRepository{pool: pool}
}

// ErrWorkspaceConflict signals an existing retry key with different content.
var ErrWorkspaceConflict = errors.New("workspace request conflict")

func finishWorkspaceTx(ctx context.Context, tx pgx.Tx, failure *error) {
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), identityRollbackTimeout)
	defer cancel()
	if err := tx.Rollback(cleanup); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		*failure = errors.Join(*failure, fmt.Errorf("rollback workspace: %w", err))
	}
}

// Create commits workspace, owner and replay ledger together. Before a workspace
// exists, lock durable actor identity to serialize bootstrap keys without races.
//
//nolint:nonamedreturns // Deferred bounded cleanup preserves rollback failures.
func (r *WorkspaceRepository) Create(ctx context.Context, actor uuid.UUID,
	name, key string, hash []byte,
) (result Workspace, err error) {
	if r == nil || r.pool == nil || actor == uuid.Nil {
		return result, errors.New("workspace store unavailable")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer finishWorkspaceTx(ctx, tx, &err)
	var locked uuid.UUID
	if err = tx.QueryRow(ctx, "SELECT id FROM users WHERE id=$1 FOR UPDATE", actor).Scan(&locked); err != nil {
		return result, err
	}
	var storedHash []byte
	err = tx.QueryRow(ctx, "SELECT workspace_id, request_hash FROM workspace_bootstrap_requests "+
		"WHERE actor_id=$1 AND request_key=$2", actor, key).Scan(&result.ID, &storedHash)
	switch {
	case err == nil:
		if !bytes.Equal(hash, storedHash) {
			return Workspace{}, ErrWorkspaceConflict
		}
		// Release bootstrap actor lock before taking an existing workspace lock.
		// Replay authorization follows the normal workspace-first ordering.
		if err = tx.Commit(ctx); err != nil {
			return Workspace{}, err
		}
		return r.Summary(ctx, Scope{WorkspaceID: result.ID, ActorID: actor})
	case errors.Is(err, pgx.ErrNoRows):
		err = tx.QueryRow(ctx, "INSERT INTO workspaces(name,created_by) VALUES($1,$2) RETURNING id,name",
			name, actor).Scan(&result.ID, &result.Name)
		if err != nil {
			return Workspace{}, err
		}
		_, err = tx.Exec(ctx, "INSERT INTO memberships(workspace_id,user_id,role,created_by) "+
			"VALUES($1,$2,'owner',$2)", result.ID, actor)
		if err != nil {
			return Workspace{}, err
		}
		_, err = tx.Exec(ctx, "INSERT INTO workspace_bootstrap_requests "+
			"(actor_id,request_key,request_hash,workspace_id) VALUES($1,$2,$3,$4)", actor, key, hash, result.ID)
		if err != nil {
			return Workspace{}, err
		}
		result.Role = "owner"
	default:
		return Workspace{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Workspace{}, err
	}
	return result, nil
}

// List exposes only memberships currently granted to this durable actor.
func (r *WorkspaceRepository) List(ctx context.Context, actor uuid.UUID) ([]Workspace, error) {
	if r == nil || r.pool == nil || actor == uuid.Nil {
		return nil, errors.New("workspace store unavailable")
	}
	rows, err := r.pool.Query(ctx, "SELECT w.id,w.name,m.role FROM workspaces w "+
		"JOIN memberships m ON m.workspace_id=w.id WHERE m.user_id=$1 ORDER BY w.created_at,w.id", actor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Workspace, 0)
	for rows.Next() {
		var item Workspace
		if err = rows.Scan(&item.ID, &item.Name, &item.Role); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// LockScope locks workspace first, then freshly reads actor membership. All
// membership mutations must use exclusive=true; ordinary writes use shared locks
// and keep authorization plus their effect in this same transaction.
func (r *WorkspaceRepository) LockScope(ctx context.Context, tx pgx.Tx,
	scope Scope, exclusive bool,
) (Workspace, error) {
	var result Workspace
	if scope.WorkspaceID == uuid.Nil || scope.ActorID == uuid.Nil {
		return result, pgx.ErrNoRows
	}
	query := "SELECT id,name FROM workspaces WHERE id=$1 FOR SHARE"
	if exclusive {
		query = "SELECT id,name FROM workspaces WHERE id=$1 FOR UPDATE"
	}
	if err := tx.QueryRow(ctx, query, scope.WorkspaceID).Scan(&result.ID, &result.Name); err != nil {
		return Workspace{}, err
	}
	err := tx.QueryRow(ctx, "SELECT role FROM memberships WHERE workspace_id=$1 AND user_id=$2",
		scope.WorkspaceID, scope.ActorID).Scan(&result.Role)
	return result, err
}

// Summary authorizes using a new PostgreSQL membership read on every request.
//
//nolint:nonamedreturns // Deferred bounded cleanup preserves rollback failures.
func (r *WorkspaceRepository) Summary(ctx context.Context, scope Scope) (result Workspace, err error) {
	if r == nil || r.pool == nil {
		return result, errors.New("workspace store unavailable")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer finishWorkspaceTx(ctx, tx, &err)
	result, err = r.LockScope(ctx, tx, scope, false)
	if err != nil {
		return Workspace{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Workspace{}, err
	}
	return result, nil
}
