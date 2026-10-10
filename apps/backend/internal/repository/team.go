package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrOwnerRequired prevents removing the workspace's final owner capability.
var ErrOwnerRequired = errors.New("workspace owner required")

// MemberRemoval exposes only committed removal identity, never a private role snapshot.
type MemberRemoval struct {
	UserID      uuid.UUID `json:"removedUserId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	SelfRemoved bool      `json:"selfRemoved"`
}

type removalSnapshot struct {
	Role string `json:"role"`
	MemberRemoval
}

// Remove serializes workspace -> current actor -> scoped target. Actor policy
// precedes ledger access. A missing target can replay only its committed snapshot;
// snapshot target policy is checked before reading/comparing its private hash.
//
//nolint:nonamedreturns // Bounded deferred rollback preserves cleanup failures.
func (r *TeamRepository) Remove(ctx context.Context, scope Scope, target uuid.UUID,
	key string, hash []byte, authorize TeamAuthorize, allow func(string, string) bool,
) (result MemberRemoval, err error) {
	if r == nil || r.pool == nil || authorize == nil || allow == nil {
		return result, errors.New("team store unavailable")
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
	var membership uuid.UUID
	var currentRole string
	targetErr := tx.QueryRow(ctx, "SELECT id,role FROM memberships WHERE workspace_id=$1 AND user_id=$2 FOR UPDATE",
		scope.WorkspaceID, target).Scan(&membership, &currentRole)
	if targetErr != nil && !errors.Is(targetErr, pgx.ErrNoRows) {
		return result, targetErr
	}
	if targetErr == nil && !allow(actorRole, currentRole) {
		return result, ErrTeamForbidden
	}
	if _, err = tx.Exec(ctx, "DELETE FROM mutation_requests WHERE workspace_id=$1 AND actor_id=$2 "+
		"AND operation='member.remove' AND retain_until < now()", scope.WorkspaceID, scope.ActorID); err != nil {
		return result, err
	}
	var snapshot []byte
	err = tx.QueryRow(ctx, "SELECT result FROM mutation_requests WHERE workspace_id=$1 AND actor_id=$2 "+
		"AND operation='member.remove' AND request_key=$3", scope.WorkspaceID, scope.ActorID, key).Scan(&snapshot)
	switch {
	case err == nil:
		result, err = replayMemberRemoval(ctx, tx, scope, actorRole, key, hash, snapshot, allow)
		if err != nil {
			return result, err
		}
	case errors.Is(err, pgx.ErrNoRows):
		if targetErr != nil {
			return result, targetErr
		}
		result, err = applyMemberRemoval(ctx, tx, scope, target, membership, currentRole, key, hash)
		if err != nil {
			return result, err
		}
	default:
		return result, err
	}
	if err = tx.Commit(ctx); err != nil {
		return MemberRemoval{}, err
	}
	return result, nil
}

// Called only after current actor and any live target have been authorized.
func replayMemberRemoval(ctx context.Context, tx pgx.Tx, scope Scope, actorRole, key string,
	hash, snapshot []byte, allow func(string, string) bool,
) (MemberRemoval, error) {
	var stored removalSnapshot
	if err := json.Unmarshal(snapshot, &stored); err != nil {
		return MemberRemoval{}, err
	}
	if stored.WorkspaceID != scope.WorkspaceID || stored.UserID == uuid.Nil {
		return MemberRemoval{}, errors.New("invalid removal snapshot")
	}
	if !allow(actorRole, stored.Role) {
		return MemberRemoval{}, ErrTeamForbidden
	}
	var storedHash []byte
	err := tx.QueryRow(ctx, "SELECT request_hash FROM mutation_requests WHERE workspace_id=$1 AND actor_id=$2 "+
		"AND operation='member.remove' AND request_key=$3", scope.WorkspaceID, scope.ActorID, key).Scan(&storedHash)
	if err != nil {
		return MemberRemoval{}, err
	}
	if err = matchRequest(hash, storedHash); err != nil {
		return MemberRemoval{}, err
	}
	return stored.MemberRemoval, nil
}

func applyMemberRemoval(ctx context.Context, tx pgx.Tx, scope Scope, target, membership uuid.UUID,
	role, key string, hash []byte,
) (MemberRemoval, error) {
	result := MemberRemoval{UserID: target, WorkspaceID: scope.WorkspaceID, SelfRemoved: target == scope.ActorID}
	if err := protectFinalOwner(ctx, tx, scope, role, ""); err != nil {
		return MemberRemoval{}, err
	}
	// Delete only the scoped membership. Links, key reservations, users and audit
	// actors resolve to durable users and remain untouched.
	if _, err := tx.Exec(ctx, "DELETE FROM memberships WHERE workspace_id=$1 AND user_id=$2",
		scope.WorkspaceID, target); err != nil {
		return MemberRemoval{}, err
	}
	if err := recordRemovalAudit(ctx, tx, scope, membership); err != nil {
		return MemberRemoval{}, err
	}
	snapshot, err := json.Marshal(removalSnapshot{MemberRemoval: result, Role: role})
	if err != nil {
		return MemberRemoval{}, err
	}
	_, err = tx.Exec(ctx, "INSERT INTO mutation_requests "+
		"(workspace_id,actor_id,operation,request_key,request_hash,result) "+
		"VALUES($1,$2,'member.remove',$3,$4,$5)", scope.WorkspaceID, scope.ActorID, key, hash, snapshot)
	return result, err
}

// TeamAuthorize locks the workspace and actor, then returns current authority.
type TeamAuthorize func(context.Context, pgx.Tx, Scope) (string, error)

// ChangeRole serializes workspace -> actor membership -> target membership.
// Authorization (including current target policy) precedes ledger/hash reads.
//
//nolint:nonamedreturns // Bounded deferred rollback preserves cleanup failures.
func (r *TeamRepository) ChangeRole(ctx context.Context, scope Scope, target uuid.UUID,
	role, key string, hash []byte, authorize TeamAuthorize,
	allow func(string, string, string) bool,
) (result TeamMember, err error) {
	if r == nil || r.pool == nil || authorize == nil || allow == nil {
		return result, errors.New("team store unavailable")
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
	var membership uuid.UUID
	err = tx.QueryRow(ctx, "SELECT m.id,m.user_id,m.workspace_id,u.verified_email,m.role "+
		"FROM memberships m JOIN users u ON u.id=m.user_id WHERE m.workspace_id=$1 AND m.user_id=$2 FOR UPDATE OF m",
		scope.WorkspaceID, target).Scan(&membership, &result.ID, &result.WorkspaceID, &result.Email, &result.Role)
	if err != nil {
		return TeamMember{}, err
	}
	result.ActorRole = actorRole
	if !allow(actorRole, result.Role, role) {
		return TeamMember{}, ErrTeamForbidden
	}
	if _, err = tx.Exec(ctx, "DELETE FROM mutation_requests WHERE workspace_id=$1 AND actor_id=$2 "+
		"AND operation='member.role' AND retain_until < now()", scope.WorkspaceID, scope.ActorID); err != nil {
		return TeamMember{}, err
	}
	var storedHash, snapshot []byte
	err = tx.QueryRow(ctx, "SELECT request_hash,result FROM mutation_requests WHERE workspace_id=$1 "+
		"AND actor_id=$2 AND operation='member.role' AND request_key=$3", scope.WorkspaceID, scope.ActorID, key).
		Scan(&storedHash, &snapshot)
	switch {
	case err == nil:
		if err = matchRequest(hash, storedHash); err != nil {
			return TeamMember{}, err
		}
		if err = json.Unmarshal(snapshot, &result); err != nil {
			return TeamMember{}, err
		}
		// The member snapshot is original; capability projection remains current.
		result.ActorRole = actorRole
	case errors.Is(err, pgx.ErrNoRows):
		result, err = applyTeamRole(ctx, tx, scope, result, membership, role, key, hash)
		if err != nil {
			return TeamMember{}, err
		}
	default:
		return TeamMember{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return TeamMember{}, err
	}
	return result, nil
}

// ErrTeamForbidden denotes a denied current target/grant matrix.
var ErrTeamForbidden = errors.New("team role change forbidden")

func applyTeamRole(ctx context.Context, tx pgx.Tx, scope Scope, result TeamMember,
	membership uuid.UUID, role, key string, hash []byte,
) (TeamMember, error) {
	if err := protectFinalOwner(ctx, tx, scope, result.Role, role); err != nil {
		return TeamMember{}, err
	}
	_, err := tx.Exec(ctx, "UPDATE memberships SET role=$3 WHERE workspace_id=$1 AND user_id=$2",
		scope.WorkspaceID, result.ID, role)
	if err != nil {
		return TeamMember{}, err
	}
	result.Role = role
	if result.ID == scope.ActorID {
		result.ActorRole = role
	}
	if err = recordRoleAudit(ctx, tx, scope, membership); err != nil {
		return TeamMember{}, err
	}
	snapshot, err := json.Marshal(result)
	if err != nil {
		return TeamMember{}, err
	}
	_, err = tx.Exec(ctx, "INSERT INTO mutation_requests "+
		"(workspace_id,actor_id,operation,request_key,request_hash,result) "+
		"VALUES($1,$2,'member.role',$3,$4,$5)", scope.WorkspaceID, scope.ActorID, key, hash, snapshot)
	return result, err
}

func protectFinalOwner(ctx context.Context, tx pgx.Tx, scope Scope, current, role string) error {
	if current != teamOwnerRole || role == teamOwnerRole {
		return nil
	}
	var owners int
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM memberships WHERE workspace_id=$1 AND role='owner'",
		scope.WorkspaceID).Scan(&owners); err != nil {
		return err
	}
	if owners <= 1 {
		return ErrOwnerRequired
	}
	return nil
}

// TeamMember projects durable identity without provider or historical fields.
type TeamMember struct {
	ActorRole   string    `json:"actorRole"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
}

const teamOwnerRole = "owner"

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
