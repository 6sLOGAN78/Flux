package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// recordLinkAudit records the exact lifecycle verb with durable scoped provenance.
func recordLinkAudit(ctx context.Context, tx pgx.Tx, scope Scope, link uuid.UUID, action string) error {
	_, err := tx.Exec(ctx, "INSERT INTO audit_events(workspace_id,actor_id,operation,target_link_id) "+
		"VALUES($1,$2,$3,$4)", scope.WorkspaceID, scope.ActorID, "link."+action, link)
	return err
}

func recordSuspensionAudit(ctx context.Context, tx pgx.Tx, scope Scope, link uuid.UUID, action, reason string) error {
	_, err := tx.Exec(ctx, "INSERT INTO audit_events(workspace_id,actor_id,operation,target_link_id,reason,created_at) "+
		"VALUES($1,$2,$3,$4,$5,clock_timestamp())", scope.WorkspaceID, scope.ActorID, "link."+action, link, reason)
	return err
}

// recordInvitationAudit binds durable actor and scoped invitation atomically.
func recordInvitationAudit(ctx context.Context, tx pgx.Tx, scope Scope, invitation uuid.UUID) error {
	_, err := tx.Exec(ctx, "INSERT INTO audit_events(workspace_id,actor_id,operation,target_invitation_id) "+
		"VALUES($1,$2,'invitation.create',$3)", scope.WorkspaceID, scope.ActorID, invitation)
	return err
}

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
