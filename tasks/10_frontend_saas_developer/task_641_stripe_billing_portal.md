---
id: TASK-641
title: Subscription Plans, Quota Metering & Stripe Billing Portal
layer: Level 5 (Executable Task Unit)
status: Completed
owner: Frontend Engineer
depends_on:
  - tasks/10_frontend_saas_developer/task_640_workspaces_rbac_settings.md
references:
  - api/openapi_v3_saas.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - billing
  - stripe
  - subscription
  - quota
---

# TASK-641: Subscription Plans, Quota Metering & Stripe Billing Portal

## 1. Goal
Implement the Billing & Subscriptions page (`/settings/billing`) displaying the active tier (`Free`, `Pro`, `Enterprise`), monthly click quota progress bar, seat usage meter, and Stripe Customer Portal integration button.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/settings/BillingPage.tsx`
- `apps/frontend/src/components/billing/CurrentPlanCard.tsx`
- `apps/frontend/src/components/billing/UsageQuotaProgressBar.tsx`
- `apps/frontend/src/components/billing/InvoicesList.tsx`

## 3. Dependencies & Prerequisites
- `@flux/openapi` (`getSubscription`, `ZSubscription`)

## 4. Referenced Architecture & Product Specs
- [docs/saas/billing_stripe.md](file:///home/logan78/Desktop/flux/docs/saas/billing_stripe.md)
- [tasks/03_saas/task_302_stripe_billing.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_302_stripe_billing.md)

## 5. Acceptance Criteria
- [x] Usage meter displays percentage of monthly click allowance consumed (warning color over 80%).
- [x] Manage Billing button opens Stripe Customer Portal in a new tab for payment updates.
- [x] Upgrade plan modal initiates Stripe Checkout flow with transparent pricing tiers.

## 6. Target Deliverables
- `apps/frontend/src/pages/settings/BillingPage.tsx`
- `apps/frontend/src/components/billing/CurrentPlanCard.tsx`

## 7. Definition of Done (DoD)
- [x] Step 1: Write quota progress percentage calculation unit tests.
- [x] Step 2: Confirm test failure.
- [x] Step 3: Implement Billing page.
- [x] Step 4: Confirm tests pass.
- [x] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/settings/BillingPage.test.tsx`
