# Task 14A-02 Report: Stripe Webhook Listener

## Objective
Implement a production-safe Stripe webhook endpoint capable of synchronizing subscription lifecycle state into PostgreSQL in an idempotent and secure manner.

## Implementation Details

### Webhook Endpoint
- Created `POST /api/v1/webhooks/stripe` utilizing Echo v4.
- Routed cleanly via `RegisterWebhookRoutes` in `router/webhook.go` and initialized in `server.go`.

### Signature Verification
- Fully implemented using official Stripe SDK `v78` `webhook.ConstructEvent`.
- `STRIPE_WEBHOOK_SECRET` configuration added to `apps/backend/internal/config/config.go`.
- Configuration strictly fails closed in `production` environments if the secret is missing or default.

### Supported Events
- `customer.subscription.created`
- `customer.subscription.updated`
- `customer.subscription.deleted` (Forces tier to `free` regardless of incoming data to guarantee revocation).

### Tenant Resolution (Security Feature)
- Strictly avoids trusting any arbitrary `workspace_id` from the JSON payload.
- Retrieves `sub.Customer.ID` exclusively from the Stripe-Signed payload.
- Queries `workspaces` table dynamically to resolve `stripe_customer_id` into a trusted `workspace_id` inside the same database transaction.
- If the customer is unknown, returns `200 OK` (so Stripe stops retrying the unroutable event) but cleanly logs a warning and aborts the sync.

### Idempotency Mechanism
- Implemented PostgreSQL-backed idempotency table `stripe_events` via migration `011_create_stripe_events.sql`.
- Checks `SELECT EXISTS(SELECT 1 FROM stripe_events WHERE id = $1 FOR UPDATE SKIP LOCKED)` inside a transaction.
- Duplicate executions correctly yield `200 OK` ("Already processed") preventing database state corruption during retry storms.

### Transaction Behavior
- Encapsulated the idempotency check, tenant resolution, event logging, and subscription Upsert inside a single PostgreSQL transaction (`tx.Begin()`).
- Commits *only* if everything succeeds. If the process crashes mid-way, Stripe retries safely because the transaction rolls back, abandoning the idempotency lock.

### Database Upsert & Security Constraints
- Utilizes the `ON CONFLICT DO UPDATE` query written in 14A-01.
- Validates that `EXCLUDED.workspace_id = subscriptions.workspace_id`. If `RowsAffected == 0`, a severe cross-tenant mismatch is identified, returning `409 Conflict`.

## Security & Integration Tests
Created comprehensive Testcontainers-driven validation in `stripe_webhook_test.go`:
1. **Missing Signature:** Rejected with `400`.
2. **Invalid Signature:** Rejected with `400`.
3. **Valid Sync:** Validates DB Upsert succeeds and `subscriptions` count updates.
4. **Idempotency:** Validates identical Stripe event IDs are caught and bypassed.
5. **Unknown Customer:** Validates event is dropped and yields `200 OK`.
6. **Tenant Mismatch:** Creates two workspaces, bounds subscription to workspace 1, simulates update mapped to workspace 2 -> DB catches boundary violation and fails successfully!

## Exact Commands Executed
```bash
go get github.com/stripe/stripe-go/v78
go test -v ./internal/handler/ -run TestStripeWebhookHandler
go test -race ./...
```

## Known Limitations
- Feature tier limits (e.g. max links/clicks) are not yet enforced in the middleware (14A-03).
- Billing UI/Customer Portal generation is missing (14A-04).

## Next Task
14A-03 — Feature tier enforcement middleware
