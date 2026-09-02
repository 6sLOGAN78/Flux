package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"flux/apps/backend/internal/config"
	"flux/apps/backend/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/webhook"
)

type StripeWebhookHandler struct {
	db      *pgxpool.Pool
	repo    *repository.BillingRepository
	cfg     *config.Config
}

func NewStripeWebhookHandler(db *pgxpool.Pool, repo *repository.BillingRepository, cfg *config.Config) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		db:   db,
		repo: repo,
		cfg:  cfg,
	}
}

func (h *StripeWebhookHandler) HandleWebhook(c echo.Context) error {
	const MaxBodyBytes = int64(65536)
	c.Request().Body = http.MaxBytesReader(c.Response().Writer, c.Request().Body, MaxBodyBytes)

	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read webhook payload")
		return c.String(http.StatusServiceUnavailable, "Webhook body read error")
	}

	sigHeader := c.Request().Header.Get("Stripe-Signature")
	if sigHeader == "" {
		log.Error().Msg("Missing Stripe-Signature header")
		return c.String(http.StatusBadRequest, "Missing Stripe-Signature")
	}

	event, err := webhook.ConstructEvent(payload, sigHeader, h.cfg.Stripe.WebhookSecret)
	if err != nil {
		log.Error().Err(err).Msg("Stripe signature verification failed")
		return c.String(http.StatusBadRequest, "Invalid signature")
	}

	// Idempotency check with pgx transaction
	ctx := c.Request().Context()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return c.String(http.StatusInternalServerError, "Database error")
	}
	defer tx.Rollback(ctx)

	// Check if already processed and lock the row to avoid concurrent inserts
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM stripe_events WHERE id = $1 FOR UPDATE SKIP LOCKED)", event.ID).Scan(&exists)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query stripe_events")
		return c.String(http.StatusInternalServerError, "Database error")
	}
	if exists {
		log.Info().Str("event_id", event.ID).Msg("Stripe event already processed")
		return c.String(http.StatusOK, "Already processed")
	}
	
	// If it doesn't exist, try to insert it. If concurrent transaction inserted it and didn't commit yet, we might get an error or block.
	// Actually, just inserting it might be safer, and catching the unique violation.
	_, err = tx.Exec(ctx, "INSERT INTO stripe_events (id, type) VALUES ($1, $2)", event.ID, string(event.Type))
	if err != nil {
		log.Info().Str("event_id", event.ID).Msg("Stripe event duplicate concurrent processing detected")
		return c.String(http.StatusOK, "Concurrent duplicate")
	}

	if event.Type == "customer.subscription.created" ||
		event.Type == "customer.subscription.updated" ||
		event.Type == "customer.subscription.deleted" {
		
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			log.Error().Err(err).Msg("Failed to parse Stripe subscription object")
			return c.String(http.StatusBadRequest, "Invalid subscription payload")
		}

		if sub.Customer == nil || sub.Customer.ID == "" {
			log.Error().Msg("Subscription object missing customer ID")
			return c.String(http.StatusBadRequest, "Missing customer")
		}

		// Resolve workspace via Stripe Customer ID
		var workspaceID string
		err = tx.QueryRow(ctx, "SELECT id FROM workspaces WHERE stripe_customer_id = $1", sub.Customer.ID).Scan(&workspaceID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				log.Warn().Str("customer_id", sub.Customer.ID).Msg("Received Stripe event for unknown customer mapping")
				// Returning 200 avoids Stripe retrying a payload we can never route.
				// Returning 404 would make Stripe retry for days.
				tx.Commit(ctx)
				return c.String(http.StatusOK, "Customer unknown")
			}
			log.Error().Err(err).Msg("Failed to resolve workspace from customer ID")
			return c.String(http.StatusInternalServerError, "Database resolution error")
		}

		planTier := "free"
		if sub.Items != nil && len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
			// Naive tier resolution based on price ID or product metadata.
			// Assuming metadata contains internal tier name mapping for simplicity.
			if tier, ok := sub.Items.Data[0].Price.Metadata["tier"]; ok {
				planTier = tier
			} else if prodTier, ok := sub.Items.Data[0].Price.Product.Metadata["tier"]; ok {
				planTier = prodTier
			} else {
				planTier = "pro" // fallback or lookup
			}
		}

		if event.Type == "customer.subscription.deleted" {
			planTier = "free"
		}

		// Convert Stripe unix timestamps
		start := time.Unix(sub.CurrentPeriodStart, 0).UTC()
		end := time.Unix(sub.CurrentPeriodEnd, 0).UTC()
		
		var canceledAt *time.Time
		if sub.CanceledAt > 0 {
			t := time.Unix(sub.CanceledAt, 0).UTC()
			canceledAt = &t
		}

		// Upsert the subscription mapping
		upsertQuery := `
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
		`
		
		res, err := tx.Exec(ctx, upsertQuery,
			workspaceID,
			sub.ID,
			sub.Customer.ID,
			planTier,
			string(sub.Status),
			start,
			end,
			sub.CancelAtPeriodEnd,
			canceledAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to upsert subscription")
			return c.String(http.StatusInternalServerError, "Database upsert error")
		}

		if res.RowsAffected() == 0 {
			// This means there was a conflict on stripe_subscription_id but the workspace_id didn't match!
			// Security violation! Cross-tenant update attempted!
			log.Error().
				Str("stripe_subscription_id", sub.ID).
				Str("resolved_workspace_id", workspaceID).
				Msg("CRITICAL: Stripe subscription ownership mismatch detected during webhook")
			return c.String(http.StatusConflict, "Cross-tenant mismatch")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to commit webhook transaction")
		return c.String(http.StatusInternalServerError, "Commit failed")
	}

	log.Info().Str("event_id", event.ID).Str("type", string(event.Type)).Msg("Stripe webhook processed successfully")
	return c.String(http.StatusOK, "Success")
}
