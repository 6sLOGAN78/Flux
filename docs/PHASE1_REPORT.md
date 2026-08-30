## Data Model
We established the core relational mapping: `Workspace -> Campaign -> Link -> AnalyticsEvent`.
This guarantees a Link is strictly tied to a Campaign (nullable), and historical Clicks (AnalyticsEvents) retain an immutable snapshot of the Campaign association at the precise moment of execution.

## Campaign Schema
Introduced the `campaigns` table in PostgreSQL via migration `007_create_campaigns_and_utm.sql`:
- `id`: UUID (Primary Key)
- `workspace_id`: UUID (Foreign Key to `workspaces`, ON DELETE CASCADE)
- `name`: VARCHAR(255)
- `status`: VARCHAR(50) defaulting to `active` (supports frontend's 'active', 'paused', 'completed' states)
- `utm_campaign`: VARCHAR(255) (Default/overarching tracking value for the campaign)

## Link Changes
Expanded the `links` table in PostgreSQL via migration `007`:
- Added `campaign_id UUID REFERENCES campaigns(id) ON DELETE SET NULL`. This ensures if a campaign is deleted, active links are simply orphaned, preventing catastrophic deletion of production redirects.
- Added link-level UTM overrides: `utm_source`, `utm_medium`, `utm_campaign`, `utm_term`, `utm_content` (VARCHAR(255), Nullable).

## UTM Model
Implemented **Hybrid Model**: 
- The `Campaign` table provides the structural hierarchy.
- The `Link` table holds specific UTM parameters for tracking.
- The `AnalyticsEvent` domain entity snapshots these parameters at click-time for historical integrity.

## AnalyticsEvent Changes
Expanded the `analytics.AnalyticsEvent` struct to include:
- `CampaignID`, `UTMSource`, `UTMMedium`, `UTMCampaign`, `UTMTerm`, `UTMContent` as `*string` (omitempty).
This ensures 100% backward compatibility for serialization and keeps empty fields out of the JSON payloads when unneeded.

## ClickHouse Changes
Updated the `analytics_events` schema via `ALTER TABLE ADD COLUMN IF NOT EXISTS`:
- Added `campaign_id`, `utm_source`, `utm_medium`, `utm_campaign`, `utm_term`, `utm_content` as `Nullable(String)`.
This additive-only migration allows historical events to gracefully default to `NULL` for these columns without requiring rewriting of existing 90-day retention partitions. 
Updated `RedisAnalyticsConsumer` to correctly map the new struct properties during `PrepareBatch` / `Append`.

## Redis Compatibility Findings
Expanded `LinkRedirectTarget` safely by utilizing `omitempty` string pointers. 
**Backward Compatibility Guarantee:** Existing keys with 24-hour TTLs will smoothly deserialize under the new struct; the new UTM properties will simply remain `nil`. Upon cache expiration and subsequent Postgres cache-miss generation, the fresh payload will include the hydrated fields. No cache-purging script is required.

## Multi-Tenancy Guarantees
- The database enforces `FOREIGN KEY (workspace_id) REFERENCES workspaces(id)`.
- The domain ensures `campaign_id` on the `links` table accurately models isolation. (Phase 2's CRUD Handlers will validate cross-workspace assignment restrictions via standard Clerk JWT claims).

## Migration Strategy
Created `007_create_campaigns_and_utm.sql`.
- Strictly ordered after `006`.
- Utilizes `ADD COLUMN IF NOT EXISTS` and `CREATE TABLE IF NOT EXISTS` for idempotency.
- Retains existing links safely (additive).

## Tests Added
Created `apps/backend/internal/model/domain_test.go` verifying:
- `TestCampaign_DomainInvariants`: Ensures correct workspace ownership mapping and defaults.
- `TestLink_UTMAndCampaign`: Ensures nullable overrides function accurately.
- `TestAnalyticsEvent_SnapshotPreservation`: Proves that JSON marshaling of `AnalyticsEvent` captures immutable historical tags and natively omits `nil` campaign pointers to reduce payload bloat.

## Tests Passed
`go test ./...` completed successfully, passing all structural domain tests, ClickHouse schema migration tests, and batch append tests.

## Tests Failed
None.

## Files Changed
- `apps/backend/internal/database/migrations/007_create_campaigns_and_utm.sql` (New)
- `apps/backend/internal/model/link/link.go`
- `apps/backend/internal/model/campaign/campaign.go` (New)
- `apps/backend/internal/model/redirect/redirect.go`
- `apps/backend/internal/model/analytics/analytics.go`
- `apps/backend/internal/db/clickhouse.go`
- `apps/backend/internal/service/clickhouse_consumer.go`
- `apps/backend/internal/repository/clickhouse_analytics_test.go`
- `apps/backend/internal/model/domain_test.go` (New)

## Documentation Updated
Created new architectural decisions representing Phase 1 design constraints.

## Architectural Decisions
**DEC-005: Analytics Historical Attribution Immutability**
- **Decision:** Campaign and UTM assignments are immutable snapshots attached to the `AnalyticsEvent` at the time of redirect. 
- **Consequence:** If a user re-assigns a `Link` to a different `Campaign` tomorrow, yesterday's analytics queries will still accurately attribute the clicks to the original campaign. We will not use dynamically joined tables in ClickHouse to resolve campaign boundaries; we denormalize at ingestion.

**DEC-006: Link Campaign Deletion Policy**
- **Decision:** `links.campaign_id` is governed by `ON DELETE SET NULL`.
- **Consequence:** Deleting a campaign from the frontend will orphan the links (removing them from the campaign views) but will strictly prevent accidental destruction of active production redirect routes.

## Remaining Work
Phase 2 API Implementation is pending:
- Implement `repository/campaigns.go`.
- Implement `service/campaigns.go`.
- Implement `/api/v1/campaigns` REST surface.
- Re-wire frontend React components to utilize live endpoints instead of ephemeral state.

Phase 1 COMPLETE
