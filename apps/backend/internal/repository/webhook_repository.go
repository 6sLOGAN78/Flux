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
	ID             uuid.UUID `json:"id" db:"id"`
	WebhookID      uuid.UUID `json:"webhook_id" db:"webhook_id"`
	EventID        string    `json:"event_id" db:"event_id"`
	Status         string    `json:"status" db:"status"` // "success", "failed", "timeout", etc.
	ResponseStatus *int      `json:"response_status" db:"response_status"` // HTTP status code
	AttemptCount   int       `json:"attempt_count" db:"attempt_count"`
	LastError      *string   `json:"last_error" db:"last_error"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// RecordDelivery minimally inserts a delivery attempt for future UI/Retry capabilities.
func (r *WebhookRepository) RecordDelivery(ctx context.Context, d *WebhookDelivery) error {
	stmt := `
		INSERT INTO webhook_deliveries (
			id, webhook_id, event_id, status, response_status, attempt_count, last_error, created_at, updated_at
		) VALUES (
			@id, @webhook_id, @event_id, @status, @response_status, @attempt_count, @last_error, NOW(), NOW()
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
	}

	_, err := r.pool.Exec(ctx, stmt, args)
	return sqlerr.HandleError(err)
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
