package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"flux/apps/backend/internal/model/webhook"
	"flux/apps/backend/pkg/sqlerr"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookRepository struct {
	pool *pgxpool.Pool
}

func NewWebhookRepository(pool *pgxpool.Pool) *WebhookRepository {
	return &WebhookRepository{pool: pool}
}

func (r *WebhookRepository) CreateWebhook(ctx context.Context, wh *webhook.Webhook) (*webhook.Webhook, error) {
	stmt := `
		INSERT INTO webhooks (
			workspace_id, endpoint_url, secret, active, events
		)
		VALUES (
			@workspace_id, @endpoint_url, @secret, @active, @events
		)
		RETURNING id, workspace_id, endpoint_url, secret, active, events, created_at, updated_at
	`

	args := pgx.NamedArgs{
		"workspace_id": wh.WorkspaceID,
		"endpoint_url": wh.EndpointURL,
		"secret":       wh.Secret,
		"active":       wh.Active,
		"events":       wh.Events,
	}

	rows, err := r.pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to insert webhook: %w", err))
	}
	defer rows.Close()

	created, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[webhook.Webhook])
	if err != nil {
		return nil, fmt.Errorf("failed to collect webhook row: %w", err)
	}

	return &created, nil
}

func (r *WebhookRepository) GetWebhook(ctx context.Context, workspaceID, id uuid.UUID) (*webhook.Webhook, error) {
	stmt := `
		SELECT id, workspace_id, endpoint_url, secret, active, events, created_at, updated_at
		FROM webhooks
		WHERE id = @id AND workspace_id = @workspace_id
	`
	rows, err := r.pool.Query(ctx, stmt, pgx.NamedArgs{"id": id, "workspace_id": workspaceID})
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	wh, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[webhook.Webhook])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &wh, nil
}

func (r *WebhookRepository) ListWebhooks(ctx context.Context, workspaceID uuid.UUID) ([]webhook.Webhook, error) {
	stmt := `
		SELECT id, workspace_id, endpoint_url, secret, active, events, created_at, updated_at
		FROM webhooks
		WHERE workspace_id = @workspace_id
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, stmt, pgx.NamedArgs{"workspace_id": workspaceID})
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[webhook.Webhook])
}

func (r *WebhookRepository) UpdateWebhook(ctx context.Context, workspaceID, id uuid.UUID, active *bool, endpointURL *string, events *[]string) (*webhook.Webhook, error) {
	stmt := `
		UPDATE webhooks
		SET 
			active = COALESCE(@active, active),
			endpoint_url = COALESCE(@endpoint_url, endpoint_url),
			events = COALESCE(@events, events),
			updated_at = NOW()
		WHERE id = @id AND workspace_id = @workspace_id
		RETURNING id, workspace_id, endpoint_url, secret, active, events, created_at, updated_at
	`

	args := pgx.NamedArgs{
		"id":           id,
		"workspace_id": workspaceID,
		"active":       active,
		"endpoint_url": endpointURL,
		"events":       events,
	}

	rows, err := r.pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	wh, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[webhook.Webhook])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &wh, nil
}

func (r *WebhookRepository) DeleteWebhook(ctx context.Context, workspaceID, id uuid.UUID) error {
	stmt := `
		DELETE FROM webhooks
		WHERE id = @id AND workspace_id = @workspace_id
	`
	tag, err := r.pool.Exec(ctx, stmt, pgx.NamedArgs{"id": id, "workspace_id": workspaceID})
	if err != nil {
		return sqlerr.HandleError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// WebhookDelivery represents a minimal recording of a delivery outcome.
type WebhookDelivery struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	WebhookID      uuid.UUID  `json:"webhook_id" db:"webhook_id"`
	EventID        string     `json:"event_id" db:"event_id"`
	Status         string     `json:"status" db:"status"` // "success", "retrying", "dead_letter"
	ResponseStatus *int       `json:"response_status" db:"response_status"`
	AttemptCount   int        `json:"attempt_count" db:"attempt_count"`
	LastError      *string    `json:"last_error" db:"last_error"`
	Payload        []byte     `json:"payload" db:"payload"`
	NextAttemptAt  *time.Time `json:"next_attempt_at" db:"next_attempt_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// RecordDelivery inserts a delivery attempt, optionally setting it up for retry.
func (r *WebhookRepository) RecordDelivery(ctx context.Context, d *WebhookDelivery) error {
	stmt := `
		INSERT INTO webhook_deliveries (
			id, webhook_id, event_id, status, response_status, attempt_count, last_error, payload, next_attempt_at, created_at, updated_at
		) VALUES (
			@id, @webhook_id, @event_id, @status, @response_status, @attempt_count, @last_error, @payload, @next_attempt_at, NOW(), NOW()
		)
	`
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}

	args := pgx.NamedArgs{
		"id":              d.ID,
		"webhook_id":      d.WebhookID,
		"event_id":        d.EventID,
		"status":          d.Status,
		"response_status": d.ResponseStatus,
		"attempt_count":   d.AttemptCount,
		"last_error":      d.LastError,
		"payload":         d.Payload,
		"next_attempt_at": d.NextAttemptAt,
	}

	_, err := r.pool.Exec(ctx, stmt, args)
	return sqlerr.HandleError(err)
}

// ClaimDueRetries atomically selects and locks a batch of due retries, advancing their status to 'processing'.
func (r *WebhookRepository) ClaimDueRetries(ctx context.Context, limit int) ([]WebhookDelivery, error) {
	stmt := `
		UPDATE webhook_deliveries
		SET status = 'processing', updated_at = NOW()
		WHERE id IN (
			SELECT id
			FROM webhook_deliveries
			WHERE status = 'retrying' AND next_attempt_at <= NOW()
			ORDER BY next_attempt_at ASC
			LIMIT @limit
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, webhook_id, event_id, status, response_status, attempt_count, last_error, payload, next_attempt_at, created_at, updated_at
	`
	rows, err := r.pool.Query(ctx, stmt, pgx.NamedArgs{"limit": limit})
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[WebhookDelivery])
}

// UpdateDeliveryState updates an existing delivery (e.g. after a retry attempt).

// RecoverStuckDeliveries finds deliveries stuck in 'processing' state for longer than timeout
// and resets them to 'retrying' so they can be picked up again.
func (r *WebhookRepository) RecoverStuckDeliveries(ctx context.Context, timeout time.Duration) (int64, error) {
	stmt := `
		UPDATE webhook_deliveries
		SET status = 'retrying', updated_at = NOW()
		WHERE status = 'processing' AND updated_at < NOW() - $1::interval
	`
	intervalStr := fmt.Sprintf("%d seconds", int(timeout.Seconds()))
	tag, err := r.pool.Exec(ctx, stmt, intervalStr)
	if err != nil {
		return 0, sqlerr.HandleError(err)
	}
	return tag.RowsAffected(), nil
}

func (r *WebhookRepository) UpdateDeliveryState(ctx context.Context, id uuid.UUID, status string, responseStatus *int, attemptCount int, lastError *string, nextAttemptAt *time.Time) error {
	stmt := `
		UPDATE webhook_deliveries
		SET status = @status, response_status = @response_status, attempt_count = @attempt_count, last_error = @last_error, next_attempt_at = @next_attempt_at, updated_at = NOW()
		WHERE id = @id
	`
	_, err := r.pool.Exec(ctx, stmt, pgx.NamedArgs{"id": id, "status": status, "response_status": responseStatus, "attempt_count": attemptCount, "last_error": lastError, "next_attempt_at": nextAttemptAt})
	return sqlerr.HandleError(err)
}

// ListDeliveries returns the most recent deliveries for a specific webhook.
func (r *WebhookRepository) ListDeliveries(ctx context.Context, webhookID uuid.UUID, limit int) ([]WebhookDelivery, error) {
	stmt := `
		SELECT id, webhook_id, event_id, status, response_status, attempt_count, last_error, payload, next_attempt_at, created_at, updated_at
		FROM webhook_deliveries
		WHERE webhook_id = @webhook_id
		ORDER BY created_at DESC
		LIMIT @limit
	`
	rows, err := r.pool.Query(ctx, stmt, pgx.NamedArgs{"webhook_id": webhookID, "limit": limit})
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[WebhookDelivery])
}

// GetWebhookByID internally fetches a webhook ignoring workspace checks (used by retry worker).

func (r *WebhookRepository) GetWebhookByID(ctx context.Context, id uuid.UUID) (*webhook.Webhook, error) {
	stmt := `
		SELECT id, workspace_id, endpoint_url, secret, active, events, created_at, updated_at
		FROM webhooks
		WHERE id = @id
	`
	rows, err := r.pool.Query(ctx, stmt, pgx.NamedArgs{"id": id})
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	wh, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[webhook.Webhook])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &wh, nil
}

// GetActiveWebhooksForWorkspace returns webhooks that are active and could potentially receive an event.
func (r *WebhookRepository) GetActiveWebhooksForWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]webhook.Webhook, error) {
	stmt := `
		SELECT id, workspace_id, endpoint_url, secret, active, events, created_at, updated_at
		FROM webhooks
		WHERE workspace_id = @workspace_id AND active = true
	`
	rows, err := r.pool.Query(ctx, stmt, pgx.NamedArgs{"workspace_id": workspaceID})
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[webhook.Webhook])
}
