package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TeamMember projects durable identity without provider or historical fields.
type TeamMember struct {
	Email, Role     string
	ID, WorkspaceID uuid.UUID
}

const teamLookahead = 26

// TeamRepository reads only workspace-bound current memberships.
type TeamRepository struct{ pool *pgxpool.Pool }

// NewTeamRepository injects authoritative PostgreSQL.
func NewTeamRepository(pool *pgxpool.Pool) *TeamRepository { return &TeamRepository{pool: pool} }

// List holds workspace authorization and the scoped query in one transaction.
// The fixed limit includes one lookahead; UUID continuation grants no authority.
//
//nolint:nonamedreturns // Deferred bounded cleanup retains rollback failures.
func (r *TeamRepository) List(ctx context.Context, scope Scope, after uuid.UUID,
	authorize func(context.Context, pgx.Tx, Scope) error,
) (result []TeamMember, err error) {
	if r == nil || r.pool == nil || authorize == nil {
		return nil, errors.New("team store unavailable")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer finishWorkspaceTx(ctx, tx, &err)
	if err = authorize(ctx, tx, scope); err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, "SELECT m.user_id,m.workspace_id,u.verified_email,m.role "+
		"FROM memberships m JOIN users u ON u.id=m.user_id "+
		"WHERE m.workspace_id=$1 AND m.user_id>$2 ORDER BY m.user_id LIMIT 26", scope.WorkspaceID, after)
	if err != nil {
		return nil, err
	}
	result = make([]TeamMember, 0, teamLookahead)
	for rows.Next() {
		var item TeamMember
		if err = rows.Scan(&item.ID, &item.WorkspaceID, &item.Email, &item.Role); err != nil {
			rows.Close()
			return nil, err
		}
		result = append(result, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
