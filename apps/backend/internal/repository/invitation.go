package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrInvitationPending exposes only the scoped duplicate constraint.
var ErrInvitationPending = errors.New("invitation already pending")

// ErrInvitationInvalid deliberately unifies absent and unauthorized credentials.
var ErrInvitationInvalid = errors.New("invitation invalid")

// ErrInvitationRecipient contains no invitation details.
var ErrInvitationRecipient = errors.New("invitation recipient mismatch")

// InvitationInspection is disclosed only after fresh verified recipient matching.
type InvitationInspection struct {
	ExpiresAt     time.Time
	WorkspaceName string
	Role          string
	Status        string
	WorkspaceID   uuid.UUID
}

type invitationRedemption struct {
	expiresAt                      time.Time
	accepted                       *uuid.UUID
	email, role, state, memberRole string
	inviter                        uuid.UUID
	hasMembership                  bool
}

const invitationAccepted = "accepted"

// Redeem shares a workspace -> durable actor -> invitation lock boundary.
// Preliminary digest lookup projects no details and ambiguous credentials fail closed.
//
//nolint:nonamedreturns // Deferred rollback preserves bounded cleanup failures.
func (r *InvitationRepository) Redeem(ctx context.Context, actor uuid.UUID, digest []byte,
	verified func(context.Context) (string, error), accept bool,
) (review InvitationInspection, workspace Workspace, err error) {
	if r == nil || r.pool == nil || actor == uuid.Nil || verified == nil {
		return review, workspace, errors.New("invitation store unavailable")
	}
	var scope Scope
	scope.ActorID = actor
	var invite uuid.UUID
	err = r.pool.QueryRow(ctx, "SELECT workspace_id,id FROM invitations WHERE token_digest=$1 "+
		"AND (SELECT count(*) FROM invitations WHERE token_digest=$1)=1", digest).Scan(&scope.WorkspaceID, &invite)
	if errors.Is(err, pgx.ErrNoRows) {
		return review, workspace, ErrInvitationInvalid
	}
	if err != nil {
		return review, workspace, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return review, workspace, err
	}
	defer finishWorkspaceTx(ctx, tx, &err)
	entry, err := readRedemption(ctx, tx, scope, invite, digest, verified)
	if err != nil {
		return review, workspace, err
	}
	workspace.ID, workspace.Role = scope.WorkspaceID, entry.memberRole
	if accept {
		if err = applyRedemption(ctx, tx, scope, invite, entry); err != nil {
			return review, workspace, err
		}
		if !entry.hasMembership {
			workspace.Role = entry.role
		}
	}
	err = tx.QueryRow(ctx, "SELECT name FROM workspaces WHERE id=$1", scope.WorkspaceID).Scan(&workspace.Name)
	if err != nil {
		return review, workspace, err
	}
	review = InvitationInspection{WorkspaceID: workspace.ID, WorkspaceName: workspace.Name,
		Role: entry.role, ExpiresAt: entry.expiresAt, Status: "Pending"}
	if entry.state == invitationAccepted || accept {
		review.Status = "Accepted"
	}
	if err = tx.Commit(ctx); err != nil {
		return InvitationInspection{}, Workspace{}, err
	}
	return review, workspace, nil
}

func readRedemption(ctx context.Context, tx pgx.Tx, scope Scope, invite uuid.UUID, digest []byte,
	verified func(context.Context) (string, error),
) (invitationRedemption, error) {
	var item invitationRedemption
	var locked uuid.UUID
	if err := tx.QueryRow(ctx, "SELECT id FROM workspaces WHERE id=$1 FOR UPDATE",
		scope.WorkspaceID).Scan(&locked); err != nil {
		return item, err
	}
	if err := tx.QueryRow(ctx, "SELECT id FROM users WHERE id=$1 FOR UPDATE",
		scope.ActorID).Scan(&locked); err != nil {
		return item, err
	}
	err := tx.QueryRow(ctx, "SELECT email,role,state,inviter_id,accepted_user_id,expires_at FROM invitations "+
		"WHERE workspace_id=$1 AND id=$2 AND token_digest=$3 FOR UPDATE", scope.WorkspaceID, invite, digest).
		Scan(&item.email, &item.role, &item.state, &item.inviter, &item.accepted, &item.expiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return item, ErrInvitationInvalid
	}
	if err != nil {
		return item, err
	}
	email, err := verified(ctx)
	if err != nil {
		return item, err
	}
	if item.email != email {
		return item, ErrInvitationRecipient
	}
	var valid bool
	// Wall-clock expiry is checked AFTER the bounded fresh profile lookup.
	err = tx.QueryRow(ctx, "SELECT role IN ('admin','member','viewer') AND (state='accepted' OR "+
		"(state='pending' AND expires_at>clock_timestamp() AND expires_at<=created_at+interval '7 days')) "+
		"FROM invitations WHERE workspace_id=$1 AND id=$2", scope.WorkspaceID, invite).Scan(&valid)
	if err != nil {
		return item, err
	}
	if !valid {
		return item, ErrInvitationInvalid
	}
	err = tx.QueryRow(ctx, "SELECT role FROM memberships WHERE workspace_id=$1 AND user_id=$2 FOR UPDATE",
		scope.WorkspaceID, scope.ActorID).Scan(&item.memberRole)
	item.hasMembership = err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return item, err
	}
	if item.state == invitationAccepted &&
		(item.accepted == nil || *item.accepted != scope.ActorID || !item.hasMembership) {
		return item, ErrInvitationInvalid
	}
	return item, nil
}

func applyRedemption(ctx context.Context, tx pgx.Tx, scope Scope, invite uuid.UUID, item invitationRedemption) error {
	if item.state == invitationAccepted {
		return saveWorkspacePreference(ctx, tx, scope.ActorID, scope.WorkspaceID)
	}
	if !item.hasMembership {
		_, err := tx.Exec(ctx, "INSERT INTO memberships(workspace_id,user_id,role,created_by) VALUES($1,$2,$3,$4)",
			scope.WorkspaceID, scope.ActorID, item.role, item.inviter)
		if err != nil {
			return err
		}
	}
	tag, err := tx.Exec(ctx, "UPDATE invitations SET state='accepted',accepted_user_id=$3 "+
		"WHERE workspace_id=$1 AND id=$2 AND state='pending' AND expires_at>clock_timestamp()",
		scope.WorkspaceID, invite, scope.ActorID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrInvitationInvalid
	}
	if err = cancelInvitationDelivery(ctx, tx, scope.WorkspaceID, invite); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "INSERT INTO audit_events(workspace_id,actor_id,operation,target_invitation_id) "+
		"VALUES($1,$2,'invitation.accept',$3)", scope.WorkspaceID, scope.ActorID, invite)
	if err != nil {
		return err
	}
	return saveWorkspacePreference(ctx, tx, scope.ActorID, scope.WorkspaceID)
}

// Invitation is the safe, bounded public projection.
type Invitation struct {
	ExpiresAt   time.Time `json:"expiresAt"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
}

// InvitationRepository owns authoritative invite and durable delivery writes.
type InvitationRepository struct{ pool *pgxpool.Pool }

// NewInvitationRepository injects PostgreSQL without an email consumer.
func NewInvitationRepository(pool *pgxpool.Pool) *InvitationRepository {
	return &InvitationRepository{pool: pool}
}

// Create locks workspace then actor through authorize, before any replay hash.
//
//nolint:nonamedreturns // Deferred bounded rollback preserves cleanup failure.
func (r *InvitationRepository) Create(ctx context.Context, scope Scope, email, role, key string, hash []byte,
	authorize TeamAuthorize, prepare func(uuid.UUID, uuid.UUID) (DeliveryIntent, []byte, error),
	allow func(string, string) bool,
) (result Invitation, err error) {
	if r == nil || r.pool == nil || authorize == nil || prepare == nil || allow == nil {
		return result, errors.New("invitation store unavailable")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer finishWorkspaceTx(ctx, tx, &err)
	actorRole, err := authorize(ctx, tx, scope)
	if err != nil {
		return result, err
	}
	if !allow(actorRole, role) {
		return result, ErrTeamForbidden
	}
	if _, err = tx.Exec(ctx, "DELETE FROM mutation_requests WHERE workspace_id=$1 AND actor_id=$2 "+
		"AND operation='invitation.create' AND retain_until<now()", scope.WorkspaceID, scope.ActorID); err != nil {
		return result, err
	}
	var snapshot []byte
	err = tx.QueryRow(ctx, "SELECT result FROM mutation_requests WHERE workspace_id=$1 AND actor_id=$2 "+
		"AND operation='invitation.create' AND request_key=$3", scope.WorkspaceID, scope.ActorID, key).Scan(&snapshot)
	switch {
	case err == nil:
		if err = json.Unmarshal(snapshot, &result); err != nil {
			return result, err
		}
		if result.WorkspaceID != scope.WorkspaceID || !allow(actorRole, result.Role) {
			return Invitation{}, ErrTeamForbidden
		}
		if err = matchInvitationReplay(ctx, tx, scope, key, hash); err != nil {
			return Invitation{}, err
		}
	case errors.Is(err, pgx.ErrNoRows):
		result, err = applyInvitation(ctx, tx, scope, email, role, key, hash, prepare)
		if err != nil {
			return Invitation{}, err
		}
	default:
		return result, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Invitation{}, err
	}
	return result, nil
}

func matchInvitationReplay(ctx context.Context, tx pgx.Tx, scope Scope, key string, hash []byte) error {
	var savedHash []byte
	err := tx.QueryRow(ctx, "SELECT request_hash FROM mutation_requests WHERE workspace_id=$1 AND actor_id=$2 "+
		"AND operation='invitation.create' AND request_key=$3", scope.WorkspaceID, scope.ActorID, key).Scan(&savedHash)
	if err != nil {
		return err
	}
	return matchRequest(hash, savedHash)
}

func applyInvitation(ctx context.Context, tx pgx.Tx, scope Scope, email, role, key string, hash []byte,
	prepare func(uuid.UUID, uuid.UUID) (DeliveryIntent, []byte, error),
) (Invitation, error) {
	// Time-dependent expiration is transactional, never an invalid volatile index predicate.
	if _, err := tx.Exec(ctx, "UPDATE invitations SET state='expired' WHERE workspace_id=$1 AND email=$2 "+
		"AND state='pending' AND expires_at<=now()", scope.WorkspaceID, email); err != nil {
		return Invitation{}, err
	}
	result := Invitation{ID: uuid.New(), WorkspaceID: scope.WorkspaceID, Email: email, Role: role, Status: "Queued"}
	intent, digest, err := prepare(result.ID, uuid.New())
	if err != nil {
		return Invitation{}, err
	}
	err = tx.QueryRow(ctx, "INSERT INTO invitations(workspace_id,id,inviter_id,email,role,token_digest) "+
		"VALUES($1,$2,$3,$4,$5,$6) RETURNING expires_at", scope.WorkspaceID, result.ID,
		scope.ActorID, email, role, digest).Scan(&result.ExpiresAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "invitations_pending_email" {
			return Invitation{}, ErrInvitationPending
		}
		return Invitation{}, err
	}
	if err = recordInvitationDelivery(ctx, tx, scope, intent); err != nil {
		return Invitation{}, err
	}
	if err = recordInvitationAudit(ctx, tx, scope, result.ID); err != nil {
		return Invitation{}, err
	}
	snapshot, err := json.Marshal(result)
	if err != nil {
		return Invitation{}, err
	}
	_, err = tx.Exec(ctx, "INSERT INTO mutation_requests"+
		"(workspace_id,actor_id,operation,request_key,request_hash,result) "+
		"VALUES($1,$2,'invitation.create',$3,$4,$5)", scope.WorkspaceID, scope.ActorID, key, hash, snapshot)
	return result, err
}

// List bounds rows and rechecks current SQL capability in its transaction.
//
//nolint:nonamedreturns // Deferred bounded rollback preserves cleanup failure.
func (r *InvitationRepository) List(ctx context.Context, scope Scope, after uuid.UUID,
	authorize TeamAuthorize,
) (items []Invitation, err error) {
	if r == nil || r.pool == nil || authorize == nil {
		return nil, errors.New("invitation store unavailable")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer finishWorkspaceTx(ctx, tx, &err)
	if _, err = authorize(ctx, tx, scope); err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, "SELECT i.id,i.workspace_id,i.email,i.role,CASE "+
		"WHEN i.state='pending' AND i.expires_at<=now() THEN 'Expired' "+
		"WHEN i.state='expired' THEN 'Expired' WHEN i.state='accepted' THEN 'Accepted' "+
		"WHEN i.state='revoked' THEN 'Revoked' WHEN d.state='delivered' THEN 'Delivered' "+
		"WHEN d.state='failed' THEN 'Failed' ELSE 'Queued' END,i.expires_at "+
		"FROM invitations i LEFT JOIN invitation_delivery_intents d "+
		"ON d.workspace_id=i.workspace_id AND d.invitation_id=i.id "+
		"WHERE i.workspace_id=$1 AND i.id>$2 ORDER BY i.id LIMIT 26", scope.WorkspaceID, after)
	if err != nil {
		return nil, err
	}
	const lookahead = 26
	items = make([]Invitation, 0, lookahead)
	for rows.Next() {
		var item Invitation
		if err = rows.Scan(&item.ID, &item.WorkspaceID, &item.Email, &item.Role, &item.Status, &item.ExpiresAt); err != nil {
			rows.Close()
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return items, nil
}
