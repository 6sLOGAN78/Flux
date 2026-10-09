package repository

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Link is an authoritative scoped link and its durable creator projection.
type Link struct {
	SuspendedAt      *time.Time `json:"suspendedAt"`
	SuspendedBy      *uuid.UUID `json:"suspendedBy"`
	SuspensionReason *string    `json:"suspensionReason"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	Host             string     `json:"host"`
	Key              string     `json:"key"`
	Destination      string     `json:"destination"`
	Title            string     `json:"title"`
	CreatorEmail     string     `json:"creatorEmail"`
	Lifecycle        string     `json:"lifecycle"`
	ID               uuid.UUID  `json:"id"`
	WorkspaceID      uuid.UUID  `json:"workspaceId"`
	CreatorID        uuid.UUID  `json:"creatorId"`
	Version          int64      `json:"version"`
}

// LinkRepository owns transactional links; authorization is injected explicitly.
type LinkRepository struct {
	pool   *pgxpool.Pool
	random io.Reader
}

// NewLinkRepository injects the PostgreSQL authority.
func NewLinkRepository(pool *pgxpool.Pool) *LinkRepository {
	return NewLinkRepositoryWithRandom(pool, rand.Reader)
}

// NewLinkRepositoryWithRandom injects an entropy source for deterministic failure
// verification. Production callers use NewLinkRepository and crypto/rand.Reader.
func NewLinkRepositoryWithRandom(pool *pgxpool.Pool, random io.Reader) *LinkRepository {
	return &LinkRepository{pool: pool, random: random}
}

// LinkAuthorize runs inside the effect transaction before any replay or link read.
type LinkAuthorize func(context.Context, pgx.Tx, Scope) error

// ErrLinkKeyUnavailable hides global key ownership from callers.
var ErrLinkKeyUnavailable = errors.New("link key unavailable")

// Create commits the effect and exact replay snapshot together, after fresh
// workspace authorization. Actor serialization follows the shared workspace lock.
//
//nolint:nonamedreturns // Deferred bounded cleanup preserves rollback failures.
func (r *LinkRepository) Create(ctx context.Context, scope Scope, host, destination, title, key, customKey string,
	hash []byte, authorize LinkAuthorize,
) (result Link, err error) {
	if r == nil || r.pool == nil {
		return result, errors.New("link store unavailable")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer finishWorkspaceTx(ctx, tx, &err)
	if err = authorize(ctx, tx, scope); err != nil {
		return result, err
	}
	var actor uuid.UUID
	if err = tx.QueryRow(ctx, "SELECT id FROM users WHERE id=$1 FOR UPDATE", scope.ActorID).Scan(&actor); err != nil {
		return result, err
	}
	if _, err = tx.Exec(ctx, "DELETE FROM mutation_requests WHERE workspace_id=$1 AND actor_id=$2 "+
		"AND operation='link.create' AND retain_until < now()", scope.WorkspaceID, scope.ActorID); err != nil {
		return result, err
	}
	var storedHash, snapshot []byte
	err = tx.QueryRow(ctx, "SELECT request_hash,result FROM mutation_requests WHERE workspace_id=$1 "+
		"AND actor_id=$2 AND operation='link.create' AND request_key=$3", scope.WorkspaceID, scope.ActorID, key).
		Scan(&storedHash, &snapshot)
	switch {
	case err == nil:
		if err = matchRequest(hash, storedHash); err != nil {
			return result, err
		}
		if err = json.Unmarshal(snapshot, &result); err != nil {
			return Link{}, err
		}
	case errors.Is(err, pgx.ErrNoRows):
		result, err = r.insertLink(ctx, tx, scope, host, destination, title, customKey)
		if err != nil {
			return Link{}, err
		}
		snapshot, err = json.Marshal(result)
		if err != nil {
			return Link{}, err
		}
		_, err = tx.Exec(ctx, "INSERT INTO mutation_requests "+
			"(workspace_id,actor_id,operation,request_key,request_hash,result) "+
			"VALUES($1,$2,'link.create',$3,$4,$5)", scope.WorkspaceID, scope.ActorID, key, hash, snapshot)
		if err != nil {
			return Link{}, err
		}
	default:
		return Link{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Link{}, err
	}
	return result, nil
}

func (r *LinkRepository) insertLink(ctx context.Context, tx pgx.Tx, scope Scope,
	host, destination, title, customKey string,
) (Link, error) {
	if customKey != "" {
		return insertCustomLink(ctx, tx, scope, host, destination, title, customKey)
	}
	return insertGeneratedLink(ctx, tx, scope, host, destination, title, r.random)
}

func insertCustomLink(ctx context.Context, tx pgx.Tx, scope Scope, host, destination, title, key string) (Link, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, "INSERT INTO links "+
		"(workspace_id,creator_user_id,managed_host,short_key,destination,title) "+
		"VALUES($1,$2,$3,$4,$5,$6) RETURNING id", scope.WorkspaceID, scope.ActorID, host, key, destination, title).Scan(&id)
	var constraint *pgconn.PgError
	if errors.As(err, &constraint) && constraint.Code == "23505" &&
		constraint.ConstraintName == "links_managed_host_short_key_unique" {
		return Link{}, ErrLinkKeyUnavailable
	}
	if err != nil {
		return Link{}, err
	}
	return readLink(ctx, tx, scope, id)
}

func insertGeneratedLink(ctx context.Context, tx pgx.Tx, scope Scope, host, destination, title string,
	random io.Reader,
) (Link, error) {
	if random == nil {
		return Link{}, errors.New("entropy source unavailable")
	}
	for range 5 {
		var entropy [12]byte
		if _, err := io.ReadFull(random, entropy[:]); err != nil {
			return Link{}, err
		}
		key := strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(entropy[:]))
		// A failed INSERT aborts PostgreSQL transactions. Savepoints allow only
		// the named host/key conflict to retry without losing authority locks.
		attempt, err := tx.Begin(ctx)
		if err != nil {
			return Link{}, err
		}
		var id uuid.UUID
		err = attempt.QueryRow(ctx, "INSERT INTO links "+
			"(workspace_id,creator_user_id,managed_host,short_key,destination,title) "+
			"VALUES($1,$2,$3,$4,$5,$6) RETURNING id", scope.WorkspaceID, scope.ActorID, host, key, destination, title).
			Scan(&id)
		if err == nil {
			if err = attempt.Commit(ctx); err != nil {
				return Link{}, err
			}
			return readLink(ctx, tx, scope, id)
		}
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), identityRollbackTimeout)
		rollbackErr := attempt.Rollback(cleanup)
		cancel()
		if rollbackErr != nil {
			return Link{}, errors.Join(err, rollbackErr)
		}
		var constraint *pgconn.PgError
		if !errors.As(err, &constraint) || constraint.Code != "23505" ||
			constraint.ConstraintName != "links_managed_host_short_key_unique" {
			return Link{}, err
		}
	}
	return Link{}, errors.New("generated link key attempts exhausted")
}

func readLink(ctx context.Context, tx pgx.Tx, scope Scope, id uuid.UUID) (Link, error) {
	var result Link
	err := tx.QueryRow(ctx, "SELECT l.id,l.workspace_id,l.creator_user_id,l.managed_host,l.short_key,"+
		"l.destination,l.title,u.verified_email,l.lifecycle,l.created_at,l.updated_at,l.version,"+
		"l.suspended_at,l.suspended_by,l.suspension_reason FROM links l JOIN users u ON u.id=l.creator_user_id "+
		"WHERE l.workspace_id=$1 AND l.id=$2 FOR SHARE OF l", scope.WorkspaceID, id).
		Scan(&result.ID, &result.WorkspaceID, &result.CreatorID, &result.Host, &result.Key, &result.Destination,
			&result.Title, &result.CreatorEmail, &result.Lifecycle, &result.CreatedAt, &result.UpdatedAt,
			&result.Version, &result.SuspendedAt, &result.SuspendedBy, &result.SuspensionReason)
	return result, err
}

// Detail always authorizes membership before reading the tenant-qualified row.
//
//nolint:nonamedreturns // Deferred bounded cleanup preserves rollback failures.
func (r *LinkRepository) Detail(ctx context.Context, scope Scope, id uuid.UUID,
	authorize LinkAuthorize,
) (result Link, err error) {
	if r == nil || r.pool == nil {
		return result, errors.New("link store unavailable")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer finishWorkspaceTx(ctx, tx, &err)
	if err = authorize(ctx, tx, scope); err != nil {
		return result, err
	}
	result, err = readLink(ctx, tx, scope, id)
	if err != nil {
		return Link{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Link{}, err
	}
	return result, nil
}

// LinkPosition is a typed keyset position; it never supplies tenant authority.
type LinkPosition struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

// List reads scoped literal-search/lifecycle matches in newest-first order after
// fresh membership authorization under the shared workspace transaction lock.
//
//nolint:nonamedreturns // Deferred bounded cleanup preserves rollback failures.
func (r *LinkRepository) List(ctx context.Context, scope Scope, limit int, position *LinkPosition,
	search, state string, authorize LinkAuthorize,
) (result []Link, err error) {
	if r == nil || r.pool == nil {
		return nil, errors.New("link store unavailable")
	}
	if limit < 1 || limit > 101 {
		return nil, errors.New("invalid link limit")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer finishWorkspaceTx(ctx, tx, &err)
	if err = authorize(ctx, tx, scope); err != nil {
		return nil, err
	}
	var createdAt *time.Time
	var id *uuid.UUID
	if position != nil {
		createdAt, id = &position.CreatedAt, &position.ID
	}
	rows, err := tx.Query(ctx, "SELECT l.id,l.workspace_id,l.creator_user_id,l.managed_host,l.short_key,"+
		"l.destination,l.title,u.verified_email,l.lifecycle,l.created_at,l.updated_at,l.version,"+
		"l.suspended_at,l.suspended_by,l.suspension_reason FROM links l JOIN users u ON u.id=l.creator_user_id "+
		"WHERE l.workspace_id=$1 AND (($6='nondeleted' AND l.lifecycle <> 'deleted') OR l.lifecycle=$6) "+
		"AND ($5='' OR l.short_key ILIKE $5 ESCAPE '!' OR l.title ILIKE $5 ESCAPE '!' OR l.destination ILIKE $5 ESCAPE '!') "+
		"AND ($3::timestamptz IS NULL OR (l.created_at,l.id) < ($3::timestamptz,$4::uuid)) "+
		"ORDER BY l.created_at DESC,l.id DESC LIMIT $2",
		scope.WorkspaceID, limit, createdAt, id, linkSearchPattern(search), state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result = make([]Link, 0, limit)
	for rows.Next() {
		var item Link
		if err = rows.Scan(&item.ID, &item.WorkspaceID, &item.CreatorID, &item.Host, &item.Key,
			&item.Destination, &item.Title, &item.CreatorEmail, &item.Lifecycle, &item.CreatedAt,
			&item.UpdatedAt, &item.Version, &item.SuspendedAt, &item.SuspendedBy, &item.SuspensionReason); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

// linkSearchPattern treats SQL pattern metacharacters as literal user text.
// A dedicated escape character also keeps backslashes literal.
func linkSearchPattern(search string) string {
	if search == "" {
		return ""
	}
	return "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(search) + "%"
}
