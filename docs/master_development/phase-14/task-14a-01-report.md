# Task 14A-01 Report: PostgreSQL Billing Foundation

## Objective
Establish the foundational PostgreSQL data models for subscription management and Stripe customer mapping natively bound to the existing tenant isolation architecture without utilizing third-party frameworks.

## Existing Architecture Inspected
- Checked `workspaces` base architecture spanning `004` and `006` migrations resolving `clerk_org_id` schemas natively.
- Confirmed strict usage of `UUID` tenant mappings instead of generic user bounds.
- Examined existing Stripe model structs inside `apps/backend/internal/modules/billing/stripe.go`.
- Reviewed database package constraints utilizing testcontainers inside `apps/backend/internal/testing`.

## Database Design
1. Added `stripe_customer_id VARCHAR(255) UNIQUE` directly onto the `workspaces` table. Mapping customers directly onto the core tenant entity eliminates ambiguous junction tables and protects strictly against 1:N cross-tenant overlap.
2. Created `subscriptions` table mapping `workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE`.
3. Created explicit `cancel_at_period_end`, `plan_tier`, `current_period_start/end`, and `stripe_subscription_id` indices.
4. Enforced strict HTTP status enumerations: `'active', 'trialing', 'past_due', 'canceled', 'incomplete', 'incomplete_expired', 'unpaid', 'paused'` matching Stripe native architectures precisely.

## Migration Created
`apps/backend/internal/database/migrations/010_create_billing_schema.sql` (Utilizing Tern structure natively matching the codebase).

## Repository/Domain Changes
- Refactored `Subscription` struct inside `apps/backend/internal/modules/billing/stripe.go` changing `OrganizationID` into `WorkspaceID` bridging natively against the Postgres struct.
- Engineered `BillingRepository` utilizing `UpsertSubscription` and `GetSubscriptionByWorkspace`.
- **Security Logic**: `ON CONFLICT (stripe_subscription_id) DO UPDATE` explicitly specifies `WHERE subscriptions.workspace_id = EXCLUDED.workspace_id` ensuring rogue webhook updates containing cross-workspace payloads throw natively dropping malicious SQL updates.

## Tests
- Added `TestBillingRepository_Integration` inside `apps/backend/internal/repository/billing_repository_test.go`.
- Spawned `testcontainers/ryuk` and `postgres:16-alpine`.
- Successfully validated mapping constraints, fetch routines, Upsert structures, and explicit uniqueness rejections when simulating duplicate Stripe IDs across bounded workspaces.

## Exact Commands Executed
```bash
go test ./...
cd apps/backend && go test -v ./internal/repository/...
```

## Exact Results
All backend tests passed dynamically returning `ok flux/apps/backend/internal/repository 49.757s`.

## Known Limitations
- Does not listen for live Stripe Webhooks yet (Wait for 14A-02).
- Does not enforce Plan Limits globally against `TrackingHandler` natively yet (Wait for 14A-03).
- Does not handle checkout API routing natively (Wait for 14A-04).

## Next Task
14A-02 — Stripe webhook listener
