## Executive Summary
This audit evaluated the current state of Campaigns, UTM tracking, and Attribution across the frontend, backend, database, and analytics pipeline. The audit reveals that while the frontend possesses advanced, highly-styled UI components for Campaign Management, UTM Building, and Multi-Touch Attribution, **the backend and database layers are completely devoid of any campaign or UTM data structures**. The current feature state is entirely mocked.

## Existing Campaign Functionality
| Feature | Status | Frontend | Backend | Database | API | Analytics | Tests | Notes |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Campaign List | MOCK | `CampaignsPage.tsx` | MISSING | MISSING | MISSING | MISSING | Present | Stored in ephemeral React `useState`. |
| UTM Builder | PARTIAL | `UTMBuilderStudio.tsx` | MISSING | MISSING | MISSING | MISSING | None | Generates URLs client-side, saves as standard link. |
| Attribution Dashboard | MOCK | `AttributionPage.tsx` | PARTIAL | MISSING | MISSING | MISSING | `calculator_test.go` | UI relies on static constants (`BASE_CAMPAIGNS`). Backend contains unused `attribution` domain logic. |

## Existing UTM Functionality
Currently, the `UTMBuilderStudio` in the frontend concatenates UTM parameters directly onto the `destinationUrl` (e.g., `https://example.com?utm_source=twitter`). The backend treats this simply as a long string. There is no native parsing, storage, or indexing of UTM parameters on the `links` table, nor in the ClickHouse `analytics_events` table.

## Frontend Audit
- **`/campaigns` (CampaignsPage.tsx)**: Displays a list of campaigns and a visual UTM builder. Creation of a campaign bypasses a real API and instead pushes a mock object into a local `useState` array, which disappears on refresh. It also submits to the standard `createLinkMutation`.
- **`/analytics/attribution` (AttributionPage.tsx)**: Displays an advanced multi-touch attribution timeline and comparison table. It completely lacks React Query hooks, instead using a hardcoded `BASE_CAMPAIGNS` array multiplied by static percentage weights to simulate model differences.

## Backend Audit
- **Handlers/Services/Repos**: There are no Campaign-related handlers, services, or repositories.
- **Domain Logic**: An isolated, robust calculation engine exists in `apps/backend/internal/modules/attribution/calculator.go` (implementing first-touch, last-touch, linear, etc.). It is thoroughly unit-tested but disconnected from the HTTP and database layers.

## Database Audit
- **PostgreSQL**: The `links` table only tracks `destination_url` and `short_code`. There is no `campaigns` table and no `campaign_id` foreign key.
- **ClickHouse**: The `analytics_events` schema does not contain `utm_source`, `utm_medium`, `utm_campaign`, `utm_term`, `utm_content`, or `campaign_id`.

## Analytics Audit
The `AnalyticsEvent` struct generated during a redirect (`GET /:shortCode`) captures `link_id` and `workspace_id`, but silently ignores all UTM parameters since they are not modeled.

## ClickHouse Impact
For accurate historical attribution, UTM tracking and Campaign assignment must be **stored directly in the AnalyticsEvent at the time of the click**. 
If a link is reassigned to a different campaign later, historical clicks should remain attributed to the campaign active at the time of the event. Therefore, the ClickHouse schema requires expansion.

## Workspace / Security Considerations
- Campaigns must belong exclusively to a `workspace_id`.
- Users must only be able to query, mutate, or link to Campaigns within their active Clerk Organization (Tenant).
- The public redirect must remain unauthenticated, but the resulting Analytics Event must securely inherit the `workspace_id` from the cached Postgres data, never from client input.

## Existing Mock Functionality
- `INITIAL_MOCK_LINKS` in `LinksListPage.tsx`.
- `INITIAL_CAMPAIGNS` in `CampaignsPage.tsx`.
- `BASE_CAMPAIGNS` in `AttributionPage.tsx`.
- Static model weights in `AttributionPage.tsx` simulating API responses.

## Product Requirements Gap
To transform this into a functional product, the following are required:
1. PostgreSQL schema for `campaigns` and `campaign_id` on `links`.
2. ClickHouse schema expansion for `analytics_events`.
3. Backend CRUD API (`/api/v1/campaigns`).
4. Expansion of `AnalyticsEvent` pipeline in the Redirect flow.
5. Analytics Query API expansion to execute the attribution algorithms.
6. Wiring frontend React Query hooks to replace the hardcoded state.

## Recommended Data Model
**Workspace -> Campaign -> Links**
A new `campaigns` table should be introduced in PostgreSQL. A Campaign is a first-class entity (name, status, utm_campaign) that belongs to a Workspace. Links should receive an optional `campaign_id` foreign key.

## Recommended UTM Model
**Option A & Option B Hybrid:**
- `Campaign` entity stores the overarching metadata (Name, default UTM tags).
- `Link` entity stores the specific UTM parameters (`utm_source`, `utm_medium`, `utm_campaign`, etc.) for granular tracking.
- `AnalyticsEvent` stores the resolved values dynamically at click-time for immutable historical accuracy.

## Recommended API
```text
POST   /api/v1/campaigns
GET    /api/v1/campaigns
GET    /api/v1/campaigns/:id
PATCH  /api/v1/campaigns/:id
DELETE /api/v1/campaigns/:id

GET    /api/v1/analytics/attribution (Accepts ?model=linear&timeframe=30d)
```

## Recommended Frontend Integration
1. Swap `useState(INITIAL_CAMPAIGNS)` with `useQuery('/api/v1/campaigns')`.
2. Wire `UTMBuilderStudio` to trigger `POST /api/v1/campaigns` and then `POST /api/v1/links`.
3. Swap `BASE_CAMPAIGNS` in `AttributionPage` with a `useQuery` fetching from the new `/api/v1/analytics/attribution` endpoint.

## Recommended Analytics Architecture
1. **RedirectHandler**: Extract `campaign_id` and UTM tags from the cached `LinkRedirectTarget` and embed them into the `AnalyticsEvent`.
2. **ClickHouseConsumer**: Insert the enriched event into an expanded `analytics_events` table.
3. **Analytics API**: Query ClickHouse to feed the `calculator.go` attribution models dynamically.

## Implementation Order
1. **Data Model**: PostgreSQL migrations (`campaigns` table, `links.campaign_id`, `links.utm_*`).
2. **Backend Domain**: Models, Repositories, Services for Campaigns.
3. **API Contract**: Campaign CRUD handlers.
4. **Analytics Modification**: Expand `AnalyticsEvent`, update Redis Publisher, update ClickHouse Schema.
5. **Redirect Enrichment**: Update the cache schema and RedirectHandler to capture UTMs.
6. **Analytics API**: Implement `/attribution` query routes hooking into `calculator.go`.
7. **Frontend Integration**: Replace mocks with React Query hooks.

## Risks
- Modifying the ClickHouse schema in-place requires a carefully structured migration for `analytics_events` to avoid breaking the existing Phase 4 dashboards.
- Expanding the Redis Cache schema (`LinkRedirectTarget`) may temporarily break active 24-hour TTL entries unless the unmarshaler is backward-compatible.

## Open Decisions
- Should a Link be strictly bound to a single Campaign forever, or can a Link be moved between Campaigns? (Recommendation: Movable, but ClickHouse events remain immutable).
- Do we need an explicit `Conversions` API to feed the attribution models, or are we tracking conversions purely via webhooks/pixels? (The `calculator.go` implies a `Conversion` entity exists).

## Estimated Scope
Moderate to Large. Affects Database, Caching, Event Streaming, OLAP Storage, and REST API layers simultaneously.

## Files Likely To Change
- `apps/backend/internal/database/migrations/*`
- `apps/backend/internal/model/redirect/redirect.go`
- `apps/backend/internal/model/analytics/event.go`
- `apps/backend/internal/db/clickhouse.go`
- `apps/backend/internal/handler/campaigns.go` (New)
- `apps/backend/internal/service/campaigns.go` (New)
- `apps/backend/internal/repository/campaigns.go` (New)
- `apps/backend/internal/router/v1/v1.go`
- `apps/frontend/src/pages/growth/CampaignsPage.tsx`
- `apps/frontend/src/pages/analytics/AttributionPage.tsx`
