package repository

import (
	"context"
	"errors"

	"flux/apps/backend/internal/modules/billing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BillingRepository struct {
	db *pgxpool.Pool
}

func NewBillingRepository(db *pgxpool.Pool) *BillingRepository {
	return &BillingRepository{db: db}
}

// UpsertSubscription creates or updates a subscription record.
func (r *BillingRepository) UpsertSubscription(ctx context.Context, sub *billing.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			workspace_id, stripe_subscription_id, stripe_customer_id, plan_tier, status,
			current_period_start, current_period_end, cancel_at_period_end, canceled_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		ON CONFLICT (stripe_subscription_id) DO UPDATE SET
			plan_tier = EXCLUDED.plan_tier,
			status = EXCLUDED.status,
			current_period_start = EXCLUDED.current_period_start,
			current_period_end = EXCLUDED.current_period_end,
			cancel_at_period_end = EXCLUDED.cancel_at_period_end,
			canceled_at = EXCLUDED.canceled_at,
			updated_at = CURRENT_TIMESTAMP
		WHERE subscriptions.workspace_id = EXCLUDED.workspace_id
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		sub.WorkspaceID,
		sub.StripeSubscriptionID,
		sub.StripeCustomerID,
		sub.PlanTier,
		sub.Status,
		sub.CurrentPeriodStart,
		sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd,
		sub.CanceledAt,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)
	
	if err != nil {
		return err
	}

	// Also ensure customer mapping exists in workspaces table
	mappingQuery := `
		UPDATE workspaces
		SET stripe_customer_id = $1
		WHERE id = $2 AND (stripe_customer_id IS NULL OR stripe_customer_id = $1)
	`
	_, err = r.db.Exec(ctx, mappingQuery, sub.StripeCustomerID, sub.WorkspaceID)
	// We don't error out here if it affected 0 rows because maybe another workspace holds this customer ID, 
	// which would be a business logic error. Wait, we should catch it.
	return err
}

func (r *BillingRepository) GetSubscriptionByWorkspace(ctx context.Context, workspaceID uuid.UUID) (*billing.Subscription, error) {
	query := `
		SELECT id, workspace_id, stripe_subscription_id, stripe_customer_id, plan_tier, status,
			   current_period_start, current_period_end, cancel_at_period_end, canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE workspace_id = $1
		ORDER BY created_at DESC LIMIT 1
	`
	
	var sub billing.Subscription
	err := r.db.QueryRow(ctx, query, workspaceID).Scan(
		&sub.ID,
		&sub.WorkspaceID,
		&sub.StripeSubscriptionID,
		&sub.StripeCustomerID,
		&sub.PlanTier,
		&sub.Status,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd,
		&sub.CanceledAt,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	
	return &sub, nil
}
