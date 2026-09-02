package handler

import (
	"fmt"
	"net/http"

	"flux/apps/backend/internal/errs"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/config"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v78"
	"flux/apps/backend/internal/modules/billing"

	portalsession "github.com/stripe/stripe-go/v78/billingportal/session"
)

type BillingHandler struct {
	billingRepo *repository.BillingRepository
	cfg         *config.Config
}

func NewBillingHandler(billingRepo *repository.BillingRepository, cfg *config.Config) *BillingHandler {
	return &BillingHandler{
		billingRepo: billingRepo,
		cfg:         cfg,
	}
}

func (h *BillingHandler) getTenantID(c echo.Context) (uuid.UUID, error) {
	tenantID, ok := c.Get("tenant_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "workspace context not found")
	}
	return tenantID, nil
}

func (h *BillingHandler) GetSubscription(c echo.Context) error {
	tenantID, err := h.getTenantID(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	sub, err := h.billingRepo.GetSubscriptionByWorkspace(ctx, tenantID)
	
	planTier := "free"
	status := "active"
	id := "free-sub"
	var currentPeriodEnd string
	
	if err == nil {
		if sub.Status == "active" || sub.Status == "trialing" || sub.Status == "past_due" {
			planTier = sub.PlanTier
		}
		status = sub.Status
		id = sub.StripeSubscriptionID
		currentPeriodEnd = sub.CurrentPeriodEnd.Format("2006-01-02T15:04:05Z")
	} else if err != repository.ErrNotFound {
		return err
	}

	limits := billing.GetTierLimits(planTier)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":                 id,
		"orgId":              tenantID.String(),
		"plan":               planTier,
		"status":             status,
		"currentPeriodEnd":   currentPeriodEnd,
		"maxLinks":           limits.MaxLinks,
		"analyticsRetention": limits.AnalyticsRetentionDays,
	})
}

func (h *BillingHandler) CreateCustomerPortal(c echo.Context) error {
	tenantID, err := h.getTenantID(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	sub, err := h.billingRepo.GetSubscriptionByWorkspace(ctx, tenantID)
	if err != nil {
		return errs.NewBadRequestError("No active Stripe customer found for this workspace.", true, nil, nil, err)
	}

	if sub.StripeCustomerID == "" {
		return errs.NewBadRequestError("No Stripe customer mapped.", true, nil, nil, nil)
	}

	stripe.Key = h.cfg.Stripe.SecretKey
	if stripe.Key == "" {
		return errs.NewInternalServerError()
	}

	returnURL := h.cfg.Server.FrontendURL + "/billing"

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(sub.StripeCustomerID),
		ReturnURL: stripe.String(returnURL),
	}

	session, err := portalsession.New(params)
	if err != nil {
		return fmt.Errorf("failed to create Stripe Customer Portal session: %w", err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"url": session.URL,
	})
}
