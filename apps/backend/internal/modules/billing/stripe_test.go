package billing_test

import (
	"testing"
	"time"

	"flux/apps/backend/internal/modules/billing"

	"github.com/google/uuid"
)

func TestGetTierLimits(t *testing.T) {
	freeLimits := billing.GetTierLimits("free")
	if freeLimits.MaxLinks != 1000 {
		t.Errorf("expected Free plan max links 1000, got %d", freeLimits.MaxLinks)
	}

	proLimits := billing.GetTierLimits("pro")
	if proLimits.MaxLinks != 50000 {
		t.Errorf("expected Pro plan max links 50000, got %d", proLimits.MaxLinks)
	}

	bizLimits := billing.GetTierLimits("business")
	if bizLimits.MaxLinks != 500000 {
		t.Errorf("expected Business plan max links 500000, got %d", bizLimits.MaxLinks)
	}
}

func TestCheckUsageLimit(t *testing.T) {
	sub := billing.Subscription{
		PlanTier: "free",
		Status:   "active",
	}

	// 500 links created under 1000 limit -> ALLOWED
	if allowed := billing.CheckUsageLimit(sub, 500, "links"); !allowed {
		t.Error("expected 500 links to be allowed under 1000 limit")
	}

	// 1001 links created over 1000 limit -> DENIED
	if allowed := billing.CheckUsageLimit(sub, 1001, "links"); allowed {
		t.Error("expected 1001 links to be denied over 1000 limit")
	}
}

func TestProcessStripeWebhook(t *testing.T) {
	orgID := uuid.New()
	sub := billing.Subscription{
		OrganizationID:     orgID,
		StripeCustomerID:   "cus_test123",
		PlanTier:           "free",
		Status:             "active",
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
	}

	event := billing.StripeWebhookEvent{
		ID:    "evt_test_001",
		Type:  "customer.subscription.updated",
		Data:  map[string]any{"plan_tier": "pro", "status": "active"},
		SubID: "sub_test999",
	}

	updatedSub, err := billing.ProcessWebhookEvent(sub, event)
	if err != nil {
		t.Fatalf("unexpected error processing webhook event: %v", err)
	}

	if updatedSub.PlanTier != "pro" {
		t.Errorf("expected plan tier to be updated to 'pro', got %q", updatedSub.PlanTier)
	}
	if updatedSub.StripeSubscriptionID != "sub_test999" {
		t.Errorf("expected StripeSubscriptionID to be 'sub_test999', got %q", updatedSub.StripeSubscriptionID)
	}
}
