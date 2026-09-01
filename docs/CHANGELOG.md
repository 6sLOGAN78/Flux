# Changelog

## 2026-09-01 (Part 2)
**Type:** FEATURE
**Task:** 13B (URL Decoration & Tracking)
**Change:** Modified `RedirectHandler` to natively inject `?flux_cid=<uuid>` onto all outbound link destinations using `net/url`.
**Reason:** Establishes cross-domain tracking boundaries required by Multi-Touch attribution logic.


## 2026-09-01
**Type:** FEATURE
**Task:** 13A (Multi-Touch Attribution Data Model)
**Change:** Created `conversions` ClickHouse table schema and `idx_event_id` Data-Skipping Bloom filter index for `analytics_events`.
**Reason:** Establishes the foundational ingestion layer for Phase 13 Multi-Touch Attribution tracking.

## 2026-08-29
**Type:** DOCUMENTATION
**Task:** TASK-000 (Initial Audit)
**Change:** Created project control system (`PROJECT_STATE.md`, `ROADMAP.md`, `FEATURE_REGISTRY.md`, etc.).
**Reason:** To establish a source of truth for the actual state of the repository vs. the claimed state in README.

## 2026-08-29 (Part 2)
**Type:** FEATURE
**Task:** TASK-005 (Implement Authentication)
**Change:** Built native PostgreSQL `users` table, native bcrypt and JWT generation. Created `/auth/signup` and `/auth/login` APIs. Rewrote frontend `AuthContext` to submit credentials securely.
**Reason:** To replace the dangerous frontend mock-auth bypass with a real production-grade authentication flow.

### Phase 8: Clerk Authentication & Workspace Migration
* **Date:** 2026-08-29
* **Feature:** FEAT-004
* **Change:** Stripped out native authentication (bcrypt, local JWT, `/auth` endpoints). Configured Clerk as the single source of truth for Users and Organizations (Workspaces).
* **Change:** Backend now relies on Clerk SDK to verify JWTs, and uses a Just-In-Time (JIT) mapping in `ClerkJWTMiddleware` to map `clerk_user_id` and `clerk_org_id` into native Postgres UUIDs (`users` and `workspaces` tables).

## [0.1.2] - 2026-08-29
### Changed
- **FEAT-005: Finalize Workspace Architecture**: Completely removed `workspace_members` table to rely purely on Clerk as the source of truth for Organizations.
- Added `owner_id` to `workspaces` table via Migration 006 for the Personal Workspace fallback.
- Backend API tests now simulate Clerk Organization contexts cleanly via dynamic header overrides in development.

### Fixed
- **BUG-001A**: Fixed API error mapping where `pgx.ErrNoRows` incorrectly yielded HTTP 500. Added `CustomHTTPErrorHandler` to Echo to serialize domain errors (404, 400, etc.) appropriately.

### Added
- **BUG-001 (Phase 1): Analytics Pipeline Event Contract**:
  - Defined canonical `AnalyticsEvent` schema including `WorkspaceID` for multi-tenant isolation.
  - Added non-blocking `AnalyticsPublisher` abstraction to isolate tracking from application functionality.
  - Implemented async event generation directly in `RedirectHandler` utilizing `utils.HashIP` for PII protection.
  - Handled publisher failure edge cases to guarantee 100% redirect reliability.
  - Created `LogAnalyticsPublisher` as the initial Phase 1 dummy publisher.

### Added
- **BUG-001 (Phase 2): Analytics Pipeline Redis Buffering**:
  - Implemented `RedisAnalyticsPublisher` integrating `go-redis` with Redis Streams (`XADD`).
  - Added bounded channel and dedicated background worker to decouple `RedirectHandler` from Redis I/O latency.
  - Implemented graceful shutdown mechanism natively in the Echo lifecycle to drain in-memory analytics queues during `SIGTERM`/`SIGINT`.
  - Added full testcontainers integration test validating stream logic safely without mocks.

### Added
- **BUG-001 (Phase 3): Analytics Pipeline ClickHouse Ingestion**:
  - Implemented `RedisAnalyticsConsumer` utilizing Redis Streams `XREADGROUP` to ingest background analytics events.
  - Initialized `ClickHouse` database schema with `MergeTree` ordered perfectly for workspace/link analytics queries and a default 90-day TTL.
  - Implemented robust `XACK` semantics ensuring messages are solely acknowledged *after* ClickHouse batch inserts succeed.
  - Added pending message recovery via background `XAUTOCLAIM` worker loop.
  - Extended testing infrastructure with a native ClickHouse TestContainer integration test proving end-to-end event resolution from Publisher to DB.

### Added
- **BUG-001 (Phase 4): Analytics Query API**:
  - Implemented authenticated `/api/v1/analytics/*` endpoints (`/summary`, `/timeseries`, `/top-links`, `/referrers`).
  - Enforced strong workspace isolation via Clerk JWT `tenant_id` context (ignoring client-provided workspace IDs).
  - Designed robust query-time ClickHouse metrics aggregation utilizing `uniqExact(event_id)` and `uniqExact(ip_hash)`.
  - Added strict date boundary controls preventing queries > 1 year and defaulting to rolling 30-days.
  - Implemented ClickHouse Repository layer with direct integration testing using testcontainers.
  - Resolved ClickHouse-go native startup networking race via resilient Ping retry wrapper.
## [Unreleased]
### Security & Reliability Fixes (BUG-001 Remediation)
- **Security**: Removed `X-Test-Clerk-*` spoofing headers from production `ClerkJWTMiddleware` to guarantee multi-tenant isolation.
- **Reliability**: Corrected graceful shutdown order in `server.go` to stop the Echo HTTP server *before* terminating the Redis analytics publisher, preventing panics and data loss.
- **Reliability**: Deprecated per-call time-seeded PRNG in `links.go` for shortcode generation; now uses `crypto/rand` bounds, `pkg/base62`, and a 5-attempt unique-constraint retry loop.
- **Security**: Hardened configuration validation to fail-closed if `CLERK_SECRET_KEY` is missing.
- **Frontend**: Removed optimistic mock-success fallback in `LinksListPage.tsx` so API failures correctly surface to the user.
### Performance Enhancements (FEAT-005)
- **Caching**: Implemented a Redis-backed Cache-Aside strategy for public redirect resolution (`GET /:shortCode`).
  - Added singleflight request coalescing to prevent cache stampedes on concurrent misses.
  - Failures to reach Redis gracefully fall back to querying PostgreSQL without disrupting the end-user.
  - Linked `LinkService` to explicitly invalidate cached entries on link modification or deletion.

### Phase 3 - Click-time Attribution & Analytics Pipeline
- Added `LinkRedirectTarget` UTM resolution at query time (`LEFT JOIN` on campaigns).
- Implemented Cache Hit/Miss Parity for attribution data.
- Eager invalidation of Redis redirect caches when Campaigns are modified or deleted.
- Non-blocking injection of `CampaignID` and 5 standard UTMs into `AnalyticsEvent` payloads.
- Mapped Analytics Consumer directly to ClickHouse schema for immutable historical tracking.

### Phase 4 - Attribution Analytics API
- Implemented `GET /api/v1/analytics/campaigns` returning click/visitor aggregation per campaign.
- Implemented `GET /api/v1/analytics/utm` returning click/visitor aggregation per selected UTM dimension.
- Maintained immutable historical tracking in ClickHouse without mutating old data.
- Fixed `AnalyticsHandler` type assertion for `tenant_id` extracting `uuid.UUID` safely.
- Added exhaustive integration tests for workspace isolation and historical attribution spanning cache/DB.
- Added strict `@flux/zod` contract exports for campaign & UTM analytics performance.

### Phase 11F - Campaign & UTM Frontend Integration
- Wired up `CampaignsPage.tsx`, `CampaignListTable.tsx`, and `UTMBuilderStudio.tsx` to the real backend APIs.
- Implemented `useCampaignsQuery.ts` and `useAnalyticsQuery.ts` hooks with proper React Query caching and Clerk `orgId` isolation.
- Integrated `CreateLinkDrawer.tsx` to handle `campaignId` and UTM overrides.
- Updated `AnalyticsPage.tsx` to include `UTMPerformanceTable.tsx` for real-time campaign attribution.
- Removed legacy `MOCK_` and `INITIAL_` fallback data.
- Fixed OpenAPI `@ts-rest/core` contract syntax errors.

### Phase 12B - Custom Domain Data Model
- Implemented `custom_domains` PostgreSQL schema (Migration 008) tracking domain state and DNS validation tokens.
- Bound domains strongly to `tenant_id` ensuring isolated ownership per workspace.
- Added strict normalizations (`CHECK (hostname = LOWER(hostname))`, `UNIQUE`, no trailing dots).
- Connected `links` table to `custom_domains` via `custom_domain_id` (ON DELETE SET NULL).
- Implemented repository layer and extensive DB-level integration tests via TestContainers.

### Phase 12E - Custom Domain Routing & Cache
- Updated `RedirectHandler` to extract and normalize `Host` header (stripping ports/trailing dots and lowercasing).
- Re-architected Redis cache keys from `link:{slug}` to `redirect:{hostname}:{slug}` to isolate routing contexts.
- Overhauled `GetByHostAndSlug` SQL query to assert strong tenant boundaries natively, preventing cross-tenant access via domain spoofing.
- Augmented cache invalidation for Links and Campaigns to dynamically lookup and purge `redirect:{hostname}:{slug}` keys based on the link's custom domain attachment.
- Preserved `CustomDomainID` and `Hostname` metadata into `AnalyticsEvent` payload for Phase 12G integration.

### Phase 13D: Attribution API / Engine
- **Implemented** `AttributionProvider` joining `conversions` and `analytics_events` on ClickHouse.
- **Implemented** array join algorithms natively mapping `click_ids` safely per tenant without performance regression.
- **Added** `GET /api/v1/analytics/attribution` endpoint to securely execute analytical computations dynamically leveraging `calculator.go`.
- **Integrated** OpenAPI schemas bounding the multi-touch models endpoint logically to Zod payload boundaries.
