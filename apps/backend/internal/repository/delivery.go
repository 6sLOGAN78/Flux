package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	deliveryBatchSize     = 25
	deliveryFinishTimeout = 2 * time.Second
	deliveryMaximumDelay  = 5 * time.Minute
)

// DeliveryIntent is private encrypted storage, never serialized to clients.
type DeliveryIntent struct {
	KeyID        string
	Ciphertext   []byte
	ID           uuid.UUID
	InvitationID uuid.UUID
}

// DeliveryClaim contains only tenant-fenced queue references.
type DeliveryClaim struct {
	WorkspaceID  uuid.UUID `json:"workspaceId"`
	InvitationID uuid.UUID `json:"invitationId"`
	DeliveryID   uuid.UUID `json:"deliveryId"`
	Generation   int64     `json:"generation"`
}

// DeliveryRepository owns the durable invitation dispatch and acknowledgement boundary.
type DeliveryRepository struct{ pool *pgxpool.Pool }

// NewDeliveryRepository injects the worker-owned PostgreSQL pool.
func NewDeliveryRepository(pool *pgxpool.Pool) *DeliveryRepository {
	return &DeliveryRepository{pool: pool}
}

// Prepare durably freezes encrypted request bytes before Redis publication/send.
//
//nolint:lll // Keep scoped SQL fences, bounded validation and exact fixture assertions together.
func (r *DeliveryRepository) Prepare(ctx context.Context, claim DeliveryClaim,
	freeze func(DeliveryIntent, string, string) ([]byte, error),
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	intent := DeliveryIntent{ID: claim.DeliveryID, InvitationID: claim.InvitationID}
	var name, role string
	err = tx.QueryRow(ctx, `SELECT d.key_id,d.ciphertext,w.name,i.role FROM invitation_delivery_intents d
 JOIN invitations i ON i.workspace_id=d.workspace_id AND i.id=d.invitation_id JOIN workspaces w ON w.id=d.workspace_id
 WHERE d.workspace_id=$1 AND d.id=$2 AND d.invitation_id=$3 AND d.lease_generation=$4
 AND d.state='leased' AND d.lease_until>clock_timestamp() AND i.state='pending' AND i.expires_at>clock_timestamp()
 FOR UPDATE OF d,i`, claim.WorkspaceID, claim.DeliveryID, claim.InvitationID, claim.Generation).Scan(&intent.KeyID, &intent.Ciphertext, &name, &role)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	ciphertext, err := freeze(intent, name, role)
	if err != nil {
		_, saveErr := tx.Exec(ctx, `UPDATE invitation_delivery_intents SET attempts=attempts+1,
 state=CASE WHEN attempts+1>=8 THEN 'failed' ELSE 'queued' END,lease_until=NULL,
 available_at=clock_timestamp()+least(power(2,attempts+1),300)*interval '1 second'
 WHERE workspace_id=$1 AND id=$2 AND invitation_id=$3 AND lease_generation=$4 AND state='leased'`, claim.WorkspaceID, claim.DeliveryID, claim.InvitationID, claim.Generation)
		if saveErr != nil {
			return saveErr
		}
		if saveErr = tx.Commit(ctx); saveErr != nil {
			return saveErr
		}
		return errors.New("invitation preparation unavailable")
	}
	defer clear(ciphertext)
	_, err = tx.Exec(ctx, `UPDATE invitation_delivery_intents SET ciphertext=$5 WHERE workspace_id=$1 AND id=$2
 AND invitation_id=$3 AND lease_generation=$4 AND state='leased' AND lease_until>clock_timestamp()`,
		claim.WorkspaceID, claim.DeliveryID, claim.InvitationID, claim.Generation, ciphertext)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Claim leases a bounded pending batch; Redis publication never removes SQL intent.
func (r *DeliveryRepository) Claim(ctx context.Context) ([]DeliveryClaim, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// Terminal invitations cannot be delivered; expire private envelopes as well.
	_, err = tx.Exec(ctx, `WITH terminal AS (SELECT d.workspace_id,d.id FROM invitation_delivery_intents d
 JOIN invitations i ON i.workspace_id=d.workspace_id AND i.id=d.invitation_id
 WHERE d.state IN ('queued','leased') AND (i.state<>'pending' OR i.expires_at<=clock_timestamp())
 ORDER BY d.available_at,d.id LIMIT 25 FOR UPDATE OF d SKIP LOCKED)
 UPDATE invitation_delivery_intents d SET state='cancelled',lease_until=NULL,ciphertext=NULL
 FROM terminal t WHERE d.workspace_id=t.workspace_id AND d.id=t.id`)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `WITH pending AS (
 SELECT d.workspace_id,d.id FROM invitation_delivery_intents d JOIN invitations i
 ON i.workspace_id=d.workspace_id AND i.id=d.invitation_id
 WHERE ((d.state='queued' AND d.available_at<=clock_timestamp()) OR
 (d.state='leased' AND d.lease_until<=clock_timestamp())) AND d.attempts<8
 AND i.state='pending' AND i.expires_at>clock_timestamp()
 ORDER BY d.available_at,d.id LIMIT 25 FOR UPDATE OF d SKIP LOCKED)
 UPDATE invitation_delivery_intents d SET state='leased',lease_until=clock_timestamp()+interval '60 seconds',
 lease_generation=d.lease_generation+1 FROM pending p WHERE d.workspace_id=p.workspace_id AND d.id=p.id
 RETURNING d.workspace_id,d.invitation_id,d.id,d.lease_generation`)
	if err != nil {
		return nil, err
	}
	claims := make([]DeliveryClaim, 0, deliveryBatchSize)
	for rows.Next() {
		var claim DeliveryClaim
		if err = rows.Scan(&claim.WorkspaceID, &claim.InvitationID, &claim.DeliveryID, &claim.Generation); err != nil {
			rows.Close()
			return nil, err
		}
		claims = append(claims, claim)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return claims, nil
}

// Deliver serializes duplicates, rechecks SQL authority and commits only sender acknowledgement.
// The bounded external call is made while holding the intent lock, so duplicate
// Redis references cannot produce concurrent sends for the same lease.
//
//nolint:lll // Keep scoped SQL fences, bounded validation and exact fixture assertions together.
func (r *DeliveryRepository) Deliver(ctx context.Context, claim DeliveryClaim,
	send func(context.Context, DeliveryIntent, string, string) error,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var intent DeliveryIntent
	var workspace, role string
	var attempts int
	intent.ID, intent.InvitationID = claim.DeliveryID, claim.InvitationID
	err = tx.QueryRow(ctx, `SELECT d.key_id,d.ciphertext,w.name,i.role,d.attempts FROM invitation_delivery_intents d
 JOIN invitations i ON i.workspace_id=d.workspace_id AND i.id=d.invitation_id
 JOIN workspaces w ON w.id=d.workspace_id
 WHERE d.workspace_id=$1 AND d.id=$2 AND d.invitation_id=$3 AND d.lease_generation=$4
 AND d.state='leased' AND d.lease_until>clock_timestamp() AND d.attempts<8
 AND i.state='pending' AND i.expires_at>clock_timestamp() FOR UPDATE OF d,i`,
		claim.WorkspaceID, claim.DeliveryID, claim.InvitationID, claim.Generation).Scan(&intent.KeyID, &intent.Ciphertext, &workspace, &role, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	attempts++
	sendErr := send(ctx, intent, workspace, role)
	clear(intent.Ciphertext)
	// Cleanup must remain possible when the sender's deadline has elapsed. This
	// small independent SQL budget commits safe retry state; it never retries send.
	finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), deliveryFinishTimeout)
	defer cancel()
	if sendErr == nil {
		_, err = tx.Exec(finishCtx, `UPDATE invitation_delivery_intents d SET state='delivered',delivered_at=clock_timestamp(),
 ciphertext=NULL,lease_until=NULL,attempts=$5 WHERE d.workspace_id=$1 AND d.id=$2 AND d.invitation_id=$3
 AND d.lease_generation=$4 AND d.state='leased' AND d.lease_until>clock_timestamp()
 AND EXISTS(SELECT 1 FROM invitations i WHERE i.workspace_id=d.workspace_id AND i.id=d.invitation_id
 AND i.state='pending' AND i.expires_at>clock_timestamp())`, claim.WorkspaceID, claim.DeliveryID, claim.InvitationID, claim.Generation, attempts)
	} else {
		delay := min(time.Second*time.Duration(1<<uint(attempts)), deliveryMaximumDelay)
		_, err = tx.Exec(finishCtx, `UPDATE invitation_delivery_intents SET state=CASE WHEN $5>=8 THEN 'failed' ELSE 'queued' END,
 lease_until=NULL,attempts=$5,available_at=clock_timestamp()+$6*interval '1 second'
 WHERE workspace_id=$1 AND id=$2 AND invitation_id=$3 AND lease_generation=$4 AND state='leased'`,
			claim.WorkspaceID, claim.DeliveryID, claim.InvitationID, claim.Generation, attempts, int64(delay/time.Second))
	}
	if err != nil {
		return err
	}
	return tx.Commit(finishCtx)
}

func recordInvitationDelivery(ctx context.Context, tx pgx.Tx, scope Scope, intent DeliveryIntent) error {
	_, err := tx.Exec(ctx, "INSERT INTO invitation_delivery_intents"+
		"(workspace_id,id,invitation_id,key_id,ciphertext) VALUES($1,$2,$3,$4,$5)",
		scope.WorkspaceID, intent.ID, intent.InvitationID, intent.KeyID, intent.Ciphertext)
	return err
}
