# Phase 15A-01 — Outbound Webhook Database Foundation

## Overview
This task establishes the robust PostgreSQL foundation and strictly isolated domain repository necessary to support outbound webhook event propagation. We successfully migrated `012_create_webhooks_table.sql` and instantiated the Go application schema bindings.

## Schema Design & Tenant Relationship
The `webhooks` table enforces absolute tenant isolation by referencing the `workspaces.id` primary key. Every operation in the `WebhookRepository` structurally mandates passing an authenticated `workspace_id`.
The schema includes:
- `id`: UUID 
- `workspace_id`: UUID (Foreign Key)
- `endpoint_url`: TEXT 
- `secret`: TEXT
- `active`: BOOLEAN
- `events`: TEXT[] (Array of canonical event types, e.g., `["link.redirect", "conversion"]`)
- `created_at` / `updated_at`: TIMESTAMPTZ

## URL Validation
`GenerateRandomHex` and `IsValidURL` abstractions inside `apps/backend/internal/lib/utils/utils.go` are utilized. Only strictly valid `http` or `https` formatted URLs are syntactically permitted at the configuration boundary.

## Secret Generation & Storage
- **Generation**: Implemented `utils.GenerateWebhookSecret()` to deterministically produce 32 bytes of secure entropy via `crypto/rand`, prefixed with `whsec_`. It is cryptographically distinct and unpredictable.
- **Storage**: Since outbound webhooks absolutely require the raw secret to produce payload HMAC-SHA256 signatures for our customers, standard one-way bcrypt hashing is mathematically impossible. Because the repository has no existing KMS infrastructure abstraction (like AWS KMS or Hashicorp Vault) for column-level encryption-at-rest, secrets are stored in plaintext natively in PostgreSQL as a deliberate architectural limitation that matches the prevailing infrastructure baseline.

## Event Subscription Representation
Utilized a PostgreSQL native `TEXT[]` array column (`events`). This allows multiple canonical events (such as `link.redirect` or `conversion` from Phase 05) to be dynamically registered to a single webhook destination payload without duplicating relational mappings.

## Indexes & Constraints
- `idx_webhooks_workspace_id`: Accelerates CRUD fetches for user management.
- `idx_webhooks_worker`: Composite index on `(workspace_id, active)` designed specifically to rapidly assist Phase 15A-02 worker queue evaluation when filtering dispatch eligibility.

## Security & PostgreSQL Integration Tests
Explicit tenant isolation boundaries were verified in `WebhookRepository_Integration`.
- Random Secret Generation verified to produce `whsec_` entropy correctly.
- Cross-tenant injection explicitly attempted (`Workspace B` attempting to fetch, mutate, or destroy `Workspace A`'s webhook ID). All correctly blocked resulting in an opaque `ErrNotFound`.
- Invalid webhook URLs and foreign-key orphaned rows successfully caught via strict DB constraints.

## Next Task
**Phase 15A-02:** Build the asynchronous webhook delivery worker (likely leveraging the active Redis Streams queue pipeline) coupled with an HMAC-SHA256 signature algorithm using this exact PostgreSQL foundation to dispatch events.
