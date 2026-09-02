package repository

import (
	"context"
	"errors"
	"fmt"

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
