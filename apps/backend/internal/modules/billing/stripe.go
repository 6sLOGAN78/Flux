// Package billing provides Stripe subscription lifecycle management, webhook event processing, and usage tier quota enforcement.
package billing

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Subscription represents a workspace's SaaS subscription state.
type Subscription struct {
	ID                   uuid.UUID `json:"id,omitempty" db:"id"`
	WorkspaceID          uuid.UUID `json:"workspace_id" db:"workspace_id"`
	StripeCustomerID     string    `json:"stripe_customer_id" db:"stripe_customer_id"`
	StripeSubscriptionID string    `json:"stripe_subscription_id,omitempty" db:"stripe_subscription_id"`
	PlanTier             string    `json:"plan_tier" db:"plan_tier"` // "free", "pro", "business"
	Status               string    `json:"status" db:"status"`       // "active", "past_due", "canceled", "trialing"
	CurrentPeriodStart   time.Time `json:"current_period_start,omitempty" db:"current_period_start"`
	CurrentPeriodEnd     time.Time `json:"current_period_end,omitempty" db:"current_period_end"`
	CancelAtPeriodEnd    bool       `json:"cancel_at_period_end" db:"cancel_at_period_end"`
	CanceledAt           *time.Time `json:"canceled_at,omitempty" db:"canceled_at"`
	CreatedAt            time.Time  `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at,omitempty" db:"updated_at"`
}

// PlanTierLimits defines resource usage quotas for a subscription plan tier.
type PlanTierLimits struct {
	MaxLinks               int64 `json:"max_links"`
	MaxClicks              int64 `json:"max_clicks"`
	MaxWorkspaces          int   `json:"max_workspaces"`
	AnalyticsRetentionDays int   `json:"analytics_retention_days"`
}

// StripeWebhookEvent represents a parsed Stripe webhook payload event.
type StripeWebhookEvent struct {
	ID    string         `json:"id"`
	Type  string         `json:"type"`
	SubID string         `json:"sub_id,omitempty"`
	Data  map[string]any `json:"data"`
}

// GetTierLimits returns the quota limits for a given subscription plan tier.
func GetTierLimits(planTier string) PlanTierLimits {
	switch strings.ToLower(strings.TrimSpace(planTier)) {
	case "pro":
		return PlanTierLimits{MaxLinks: 50000, MaxClicks: 500000, MaxWorkspaces: 5, AnalyticsRetentionDays: 30}
	case "business":
		return PlanTierLimits{MaxLinks: 500000, MaxClicks: 5000000, MaxWorkspaces: 100, AnalyticsRetentionDays: 365}
	default: // "free"
		return PlanTierLimits{MaxLinks: 1000, MaxClicks: 10000, MaxWorkspaces: 1, AnalyticsRetentionDays: 7}
	}
}

// CheckUsageLimit checks if current resource usage exceeds the tier limit.
func CheckUsageLimit(sub Subscription, currentUsage int64, resourceType string) bool {
	limits := GetTierLimits(sub.PlanTier)

	switch strings.ToLower(strings.TrimSpace(resourceType)) {
	case "links":
		return currentUsage <= limits.MaxLinks
	case "clicks":
		return currentUsage <= limits.MaxClicks
	case "workspaces":
		return currentUsage <= int64(limits.MaxWorkspaces)
	default:
		return true
	}
}

// ProcessWebhookEvent updates the Subscription state based on incoming Stripe webhook events.
func ProcessWebhookEvent(sub Subscription, event StripeWebhookEvent) (Subscription, error) {
	if event.ID == "" {
		return sub, fmt.Errorf("webhook event ID cannot be empty")
	}

	if event.SubID != "" {
		sub.StripeSubscriptionID = event.SubID
	}

	if planTier, ok := event.Data["plan_tier"].(string); ok && planTier != "" {
		sub.PlanTier = planTier
	}

	if status, ok := event.Data["status"].(string); ok && status != "" {
		sub.Status = status
	}

	switch event.Type {
	case "customer.subscription.deleted":
		sub.Status = "canceled"
		sub.PlanTier = "free"
	case "invoice.payment_failed":
		sub.Status = "past_due"
	}

	return sub, nil
}
