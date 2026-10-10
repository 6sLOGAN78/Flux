package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// recordRoleAudit retains durable actor identity and an immutable target UUID
// snapshot in protected PostgreSQL, atomically with the role effect and ledger.
func recordRoleAudit(ctx context.Context, tx pgx.Tx, scope Scope, membership uuid.UUID) error {
	_, err := tx.Exec(ctx, "INSERT INTO audit_events(workspace_id,actor_id,operation,target_membership_id) "+
		"VALUES($1,$2,'member.role',$3)", scope.WorkspaceID, scope.ActorID, membership)
	return err
}

// recordRemovalAudit retains the historical membership target with a durable
// actor reference; physical removal never cascades these protected records.
func recordRemovalAudit(ctx context.Context, tx pgx.Tx, scope Scope, membership uuid.UUID) error {
	_, err := tx.Exec(ctx, "INSERT INTO audit_events(workspace_id,actor_id,operation,target_membership_id) "+
		"VALUES($1,$2,'member.remove',$3)", scope.WorkspaceID, scope.ActorID, membership)
	return err
}
