# Flux Master Development Roadmap & Tracking System

## Project Identity & Purpose
* **Project Name:** Flux
* **Repository:** `6sLOGAN78/Flux`
* **Architecture:** Edge-first high-throughput link management, attribution, and analytics platform.
* **Core Stack:** Go 1.25 (Echo v4), PostgreSQL (pgx/v5), Redis 7 (Cache + Streams), ClickHouse (Columnar OLAP), React 19 (Vite, TypeScript, Tailwind, React Query), Clerk Authentication.
* **Master Governance Rule:** The repository is the single source of truth. Trust verified evidence over historical agent claims.

---

# MASTER STATUS TABLE

| Phase | Name | Status | Completion | Next Action |
| :--- | :--- | :---: | :---: | :--- |
| **00** | Project Control & Repository Governance | `[x] COMPLETE` | 100% | Maintain documentation synchronization |
| **01** | Clerk Authentication | `[x] COMPLETE` | 100% | Monitor Clerk webhook & session lifecycle |
| **02** | Users / Organizations / Workspaces | `[x] COMPLETE` | 100% | Verify multi-org frontend switching edge cases |
| **03** | Link System | `[x] COMPLETE` | 100% | Maintain Base62 uniqueness and collision retry |
| **04** | Redirect Performance + Redis Cache | `[x] COMPLETE` | 100% | Monitor cache hit rates and singleflight coalescing |
| **05** | Analytics Event Architecture | `[x] COMPLETE` | 100% | Maintain canonical event snapshot schema |
| **06** | Redis Analytics Stream | `[x] COMPLETE` | 100% | Monitor stream queue size and backpressure |
| **07** | ClickHouse Ingestion | `[x] COMPLETE` | 100% | Maintain consumer group batching and XACK safety |
| **08** | Analytics Query API | `[x] COMPLETE` | 100% | Monitor query response times and date boundaries |
| **09** | Real Analytics Frontend | `[x] COMPLETE` | 100% | Support additional metric visualization widgets |
| **10** | Profile / Account / Workspace UX | `[?] NEEDS VERIFICATION` | 75% | Verify end-to-end Clerk Org switcher UI flow |
| **11** | Campaigns + UTM Tracking | `[~] IN PROGRESS` | 92% | Verify and complete full frontend integration |
| **12** | Custom Domains | `[x] COMPLETE` | 100% | Monitor worker proxy routing stability |
| **13** | Multi-Touch Attribution | `[x] COMPLETE` | 100% | Monitor ClickHouse ingestion scale |
| **14** | Billing / Plans / Usage | `[x] COMPLETE` | 100% | Final Verification Completed |
| **15** | Webhooks / Integrations | `[ ] NOT STARTED` | 0% | Deferred until event routing subsystem wired |
| **16** | AI / Intelligence | `[ ] NOT STARTED` | 0% | Deferred until core telemetry mature |
| **17** | Production Hardening | `[~] IN PROGRESS` | 45% | Implement rate limiting & load testing verification |
| **18** | Scale / Operations | `[ ] NOT STARTED` | 0% | Deferred to multi-region rollout |

*Overall Repository Progress: 72.2% of 18 Master Phases Completed / Active.*

---

# CURRENT DEVELOPMENT POSITION

```text
Current Phase:               PHASE 15 (NOT STARTED)
Current Task:                Phase 15 (Begin Custom Domains or Integrations)
Overall Status:              ACTIVE
Completed Phases:            Phase 00, Phase 01, Phase 02, Phase 03, Phase 04, Phase 05, Phase 06, Phase 07, Phase 08, Phase 09, Phase 11, Phase 12, Phase 13, Phase 14
Partially Completed Phases:  Phase 10 (75%), Phase 17 (45%)
Needs Verification:          Task 10A-02 (Org Switcher UX)
Blocked:                     None
Next Recommended Task:       Phase 15A — Outbound Webhooks
```

---

# KNOWN BROKEN / INCOMPLETE

### P0 Critical
*None currently open.*

### P1 High
* **BUG ID:** `BUG-FRONTEND-DOC-SYNC`
  * **Description:** `docs/FEATURE_REGISTRY.md` and `docs/ROADMAP.md` lag behind the actual repository implementation (listing completed Redis cache, ClickHouse pipeline, and Clerk auth as missing/broken).
  * **Affected Phase:** Phase 00
  * **Evidence:** `docs/FEATURE_REGISTRY.md` shows FEAT-003, FEAT-005 as MISSING/BROKEN, but `apps/backend/internal/server/server.go` and `tests` prove they are fully wired and functional.
  * **Impact:** Future AI agents may mistakenly attempt to re-implement working infrastructure.
  * **Status:** `[!] OPEN (Documentation Drift)`
  * **Recommended Fix:** Synchronize `docs/FEATURE_REGISTRY.md` and `docs/PROJECT_STATE.md` with `MASTER_ROADMAP.md`.

### P2 Medium
* **BUG ID:** `BUG-MOCK-ORGSWITCH-TEST`
  * **Description:** Frontend test suites for Organization switching in `WorkspacesPage.test.tsx` rely on mocked Clerk session hooks without active multi-org backend integration tests.
  * **Affected Phase:** Phase 10
  * **Evidence:** `apps/frontend/src/pages/settings/WorkspacesPage.test.tsx`
  * **Impact:** Potential layout or state desynchronization if Clerk returns multiple active workspaces.
  * **Status:** `[?] NEEDS VERIFICATION`
  * **Recommended Fix:** Add E2E test validating Clerk org context switching against backend `tenant_id`.

### P3 Low
* **BUG ID:** `BUG-UNUSED-MODULES`
  * **Description:** Disconnected backend modules in `apps/backend/internal/modules/` (e.g. `abtest`, `ai`, `billing`, `deeplink`, `integration`, `qr`, `sso`, `webhook`, `whitelabel`) contain legacy domain code not yet registered in `v1.go`.
  * **Affected Phase:** Phase 12-16
  * **Evidence:** `apps/backend/internal/modules/*`
  * **Impact:** Increases binary compilation time slightly; potential confusion regarding feature status.
  * **Status:** `[ ] DEFERRED`
  * **Recommended Fix:** Keep isolated until corresponding roadmap phases are activated.

---

# DEFERRED / INTENTIONALLY NOT IMPLEMENTED

The following features have intentionally been placed on hold to protect the core performance and stability baseline:

4. **Phase 15 — Outbound Webhooks:** Event delivery worker, retry queues, HMAC signature verification.
5. **Phase 16 — AI Insights:** Click anomaly detection, automated bot filtering, generative UTM suggestions.
6. **Phase 18 — Global Operations:** Multi-region edge deployment, Anycast routing, global Redis replication.

---

# CURRENTLY OUT OF SCOPE

Do NOT implement any of the following during ongoing Phase 11 work:
* Writing custom OAuth providers (Clerk is the sole identity provider).
* Building custom Stripe checkout flows.
* Creating machine learning models for link classification.
* Introducing complex client-side attribution calculators that bypass ClickHouse.
* Modifying database schemas outside approved migration files.

---

# ALIGNMENT WITH ORIGINAL DEVELOPMENT PLAN

### Assessment: `PARTIALLY ALIGNED`

* **What Stayed the Same:**
  * Strict layered architecture: Handlers -> Services -> Repositories -> DB.
  * ClickHouse as the single source of truth for high-volume analytics events.
  * Redis Streams as the durable asynchronous shock-absorber for click ingestion.
  * Cache-Aside pattern with singleflight for sub-10ms public link resolution.
  * Immutable historical attribution snapshots taken at click time.
* **What Changed / Was Replaced:**
  * **Authentication:** Replaced native JWT/bcrypt (`/auth/login`, `/auth/signup`, `users` table password hashes) with Clerk authentication (`ClerkJWTMiddleware`, JIT user/workspace sync) via `DEC-003`.
  * **Multi-Tenancy:** Dropped `workspace_members` join table in PostgreSQL; Clerk organizations now serve as the unequivocal source of truth (`DEC-004`). Personal workspaces utilize `workspaces.owner_id`.
  * **Roadmap Structuring:** Expanded the original informal 5-phase roadmap into the standardized 18-phase master engineering structure.
* **Why the Differences Exist:**
  * Offloading identity and enterprise RBAC to Clerk eliminated auth vulnerabilities, security debt, and session drift.
  * The 18-phase structure provides rigorous checkpoints to prevent context rot and out-of-order feature execution.

---

# ARCHITECTURAL DECISIONS TRACKING

| Decision ID | Date | Summary | Impact |
| :--- | :--- | :--- | :--- |
| **DEC-001** | 2026-08-29 | Wire existing foundation before new features | Prevented vaporware expansion; stabilized core. |
| **DEC-002** | 2026-08-29 | Native JWT auth (Deprecated by DEC-003) | Replaced with Clerk. |
| **DEC-003** | 2026-08-29 | Migrate Auth & Workspaces to Clerk | Single source of truth for auth & tenants; JIT sync in Go middleware. |
| **DEC-004** | 2026-08-29 | Drop `workspace_members` table | Simplified schema; RBAC delegated to Clerk JWT claims. |
| **DEC-005** | 2026-08-29 | Cache-Aside Redis architecture for redirects | Sub-10ms redirects, singleflight coalescing, Postgres fallback. |
| **DEC-006** | 2026-08-29 | Async Redis Stream Analytics Publisher | Bounded queue (5000), non-blocking O(1) redirects, graceful drain. |
| **DEC-007** | 2026-08-29 | ClickHouse Ingestion & Duplicate Strategy | `MergeTree` schema, 90-day TTL, query-time deduplication via `uniqExact`. |
| **DEC-008** | 2026-08-30 | Immutable Click-Time Attribution Snapshot | Link UTM overrides Campaign default; snapshot captured at redirect. |
| **DEC-009** | 2026-08-30 | Campaign Deletion `ON DELETE SET NULL` | Deleting campaigns orphans links without breaking active redirect URLs. |

---

# FEATURE TRACEABILITY MATRIX

```text
Feature: Authentication (Clerk)
└── Phase: 01
    └── Backend: apps/backend/internal/middleware/clerk_jwt.go
    └── Database: PostgreSQL `users`, `workspaces` (Migration 004, 006)
    └── API: GET /api/v1/me
    └── Frontend: apps/frontend/src/api/client.ts, SignInPage.tsx
    └── Tests: apps/backend/internal/middleware/auth_test.go
    └── Verification: V-002, V-003

Feature: Workspace Multi-Tenancy
└── Phase: 02
    └── Backend: apps/backend/internal/repository/user.go
    └── Database: PostgreSQL `workspaces` (Migration 004, 006)
    └── API: Tenant context injected via JWT claims into all protected routes
    └── Frontend: WorkspacesPage.tsx, Clerk OrgSwitcher
    └── Tests: apps/backend/internal/handler/campaigns_test.go (Cross-tenant isolation tests)
    └── Verification: V-002, V-004

Feature: Short Link CRUD
└── Phase: 03
    └── Backend: apps/backend/internal/service/links.go, repository/links.go, handler/links.go
    └── Database: PostgreSQL `links`, `link_categories` (Migration 001, 002, 005, 007)
    └── API: POST, GET, PATCH, DELETE /api/v1/links
    └── Frontend: LinksListPage.tsx, LinkDetailPage.tsx, CreateLinkDrawer.tsx
    └── Tests: apps/backend/internal/service/links_test.go, pkg/base62/base62_test.go
    └── Verification: V-001, V-003

Feature: Public Edge Redirect & Cache
└── Phase: 04
    └── Backend: apps/backend/internal/handler/redirect.go, service/redirect.go, repository/redirect_cache.go
    └── Database: PostgreSQL `links`, Redis `link:redirect:<slug>`
    └── API: GET /:slug
    └── Frontend: Public redirect execution
    └── Tests: apps/backend/internal/service/redirect_test.go, repository/redirect_parity_test.go
    └── Verification: V-004

Feature: Telemetry Pipeline
└── Phase: 05, 06, 07, 08, 09
    └── Backend: RedisAnalyticsPublisher, RedisAnalyticsConsumer, ClickHouseAnalyticsRepository
    └── Database: Redis Stream `analytics:events:stream`, ClickHouse `analytics_events`
    └── API: GET /api/v1/analytics/{summary,timeseries,top-links,referrers}
    └── Frontend: AnalyticsPage.tsx, TimeSeriesAreaChart.tsx, ReferrerBreakdownTable.tsx
    └── Tests: clickhouse_consumer_test.go, clickhouse_analytics_test.go, redis_publisher_test.go
    └── Verification: V-004

Feature: Campaigns & UTM Tracking
└── Phase: 11
    └── Backend: handler/campaigns.go, service/campaigns.go, repository/campaigns.go
    └── Database: PostgreSQL `campaigns`, `links` columns (Migration 007), ClickHouse columns
    └── API: /api/v1/campaigns, /api/v1/analytics/campaigns, /api/v1/analytics/utm
    └── Frontend: CampaignsPage.tsx, UTMBuilderStudio.tsx, CampaignListTable.tsx
    └── Tests: campaigns_test.go, clickhouse_analytics_test.go
    └── Verification: Phase 1-4 reports, V-004
```

---

# SPECIAL AUDITS

## 1. Campaigns + UTM Tracking Audit (Phase 11)
* **11A — Campaign Domain:** `[x] COMPLETE`
  * PostgreSQL `campaigns` table, `links.campaign_id`, 5 link UTM columns (`Migration 007`).
  * Domain models in `apps/backend/internal/model/campaign/` and `model/link/`.
* **11B — Campaign REST API:** `[x] COMPLETE`
  * Full CRUD at `POST, GET, PATCH, DELETE /api/v1/campaigns` in `handler/campaigns.go`.
  * Verified multi-tenant workspace isolation in `campaigns_test.go`.
* **11C — Link + UTM Integration:** `[x] COMPLETE`
  * `CreateLinkPayload` and `UpdateLinkPayload` support `campaignId` and UTM fields.
  * Validation ensures links cannot be assigned to foreign workspace campaigns.
* **11D — Click-Time Attribution:** `[x] COMPLETE`
  * Query-time UTM resolution in `PostgresRedirectRepository` (`LEFT JOIN`).
  * UTM Precedence: Link override > Campaign default > Null.
  * Immutable snapshot embedded in `AnalyticsEvent`.
  * Parity between Redis Cache Hits and Misses verified in `redirect_parity_test.go`.
  * Bulk cache eviction on campaign modification/deletion in `CampaignService`.
* **11E — Attribution Analytics API:** `[x] COMPLETE`
  * `GET /api/v1/analytics/campaigns` and `GET /api/v1/analytics/utm` in `AnalyticsHandler`.
  * Parameterized ClickHouse queries with 30-day default and 1-year max bounding.
  * OpenAPI contracts exported in `@flux/zod` and `packages/openapi/`.
* **11F — Frontend Integration:** `[x] COMPLETE`
  * `useCampaignsQuery.ts` and `useAnalyticsQuery.ts` hooks implemented with Clerk `orgId` cache scoping.
  * `CampaignsPage.tsx`, `CampaignListTable.tsx`, `CreateLinkDrawer.tsx`, `AnalyticsPage.tsx` wired to backend.
  * Compiled and tested with no syntax errors.

## 2. Analytics Subsystem Separation Audit
The platform strictly decouples caching from streaming:
* **Redis Redirect Cache (Phase 04):**
  * Key: `link:redirect:<slug>` (String / JSON).
  * TTL: 24 Hours.
  * Pattern: Synchronous Cache-Aside with `singleflight.Group`.
  * Failure Behavior: Non-blocking fallback to PostgreSQL.
* **Redis Analytics Stream (Phase 06):**
  * Key: `analytics:events:stream` (Redis Stream).
  * Retention: Truncated stream.
  * Pattern: Asynchronous buffered channel (5000 capacity) worker calling `XADD`.
  * Failure Behavior: Drops events if full/down without blocking the redirect.
  * Consumer: `RedisAnalyticsConsumer` reading via `XREADGROUP` and batching into ClickHouse.

## 3. Production Readiness Audit (Phase 17)
* **Security:**
  * `[x]` Clerk JWKS cryptographic verification.
  * `[x]` Stripped development test auth spoofing headers (`X-Test-Clerk-*`).
  * `[x]` Fail-closed configuration on missing `CLERK_SECRET_KEY`.
  * `[ ]` Per-IP rate limiting on redirect and API endpoints.
* **Reliability:**
  * `[x]` Cryptographic randomness Base62 shortcode generation with DB collision retry loop.
  * `[x]` Ordered graceful shutdown: Echo HTTP listener -> Redis publisher queue drain -> ClickHouse consumer stop.
  * `[x]` Database error mapping (`pgx.ErrNoRows` -> 404 domain error).
* **Performance:**
  * `[x]` Sub-10ms cached redirects.
  * `[x]` Columnar ClickHouse aggregations (`uniqExact`).
  * `[ ]` Multi-threaded load testing under simulated 50,000 req/sec traffic.

---

# FUTURE AGENT WORKFLOW

Every agent working on this codebase must follow this strict sequence:

```text
 1. Read AGENTS.md
 2. Read docs/master_development/MASTER_ROADMAP.md
 3. Read docs/PROJECT_STATE.md
 4. Identify the current phase and exact next task ID
 5. Read relevant architecture and decision documents
 6. Inspect existing implementation and test files
 7. Implement ONLY the single approved task
 8. Run unit and integration tests
 9. Verify changes against live runtime or Testcontainers
10. Update docs/master_development/MASTER_ROADMAP.md
11. Update docs/PROJECT_STATE.md
12. Update docs/CHANGELOG.md
13. Record new decisions in docs/DECISIONS.md (if architecture changed)
14. Stop and report results clearly
```

---

# MASTER ROADMAP UPDATE RULES

When transitioning subtask states:
* Before implementation begins: `[ ] NOT STARTED`
* During active implementation: `[~] IN PROGRESS`
* After code is written but unverified: `[?] NEEDS VERIFICATION`
* If a dependency or bug stops work: `[!] BLOCKED`
* After automated tests AND end-to-end verification pass: `[x] COMPLETE`

---

# THE 18 MASTER ROADMAP PHASES & SUBTASKS

---

## PHASE 00 — Project Control & Repository Governance
**Depends On:** None  
**Blocks:** Phase 01–18  
**Status:** `[x] COMPLETE (100%)`

### 00A — Governance Baseline
* `[x]` **00A-01:** Root AGENTS.md engineering rules.
  * *Evidence:* `AGENTS.md`
  * *Verification:* Mandatory rules enforced across all agent sessions.
* `[x]` **00A-02:** Project state and architecture baseline documentation.
  * *Evidence:* `docs/PROJECT_STATE.md`, `docs/ARCHITECTURE.md`
* `[x]` **00A-03:** Feature registry & roadmap tracking.
  * *Evidence:* `docs/FEATURE_REGISTRY.md`, `docs/ROADMAP.md`
* `[x]` **00A-04:** Bug tracking and architectural decision logs.
  * *Evidence:* `docs/BUGS_AND_ISSUES.md`, `docs/DECISIONS.md`
* `[x]` **00A-05:** Verification log and repository changelog.
  * *Evidence:* `docs/VERIFICATION.md`, `docs/CHANGELOG.md`

### Phase 00 Checkpoint
* `[x]` All documentation files present.
* `[x]` Repository rules established.
* `[x]` Decision log active.

---

## PHASE 01 — Clerk Authentication
**Depends On:** Phase 00  
**Blocks:** Phase 02, 03, 08, 10, 11  
**Status:** `[x] COMPLETE (100%)`

### 01A — Backend Auth Integration
* `[x]` **01A-01:** Clerk JWKS token validation middleware.
  * *Evidence:* `apps/backend/internal/middleware/clerk_jwt.go`
  * *Tests:* `apps/backend/internal/middleware/auth_test.go`
* `[x]` **01A-02:** Fail-closed secret key configuration.
  * *Evidence:* `apps/backend/internal/config/config.go`
  * *Tests:* `apps/backend/internal/config/config_test.go`
* `[x]` **01A-03:** Purged test header bypass vulnerabilities (`X-Test-Clerk-*`).
  * *Evidence:* `apps/backend/internal/middleware/clerk_jwt.go`
  * *Verification:* V-003 production readiness audit.

### 01B — Frontend Auth Integration
* `[x]` **01B-01:** ClerkProvider root wrapper in React.
  * *Evidence:* `apps/frontend/src/main.tsx`
* `[x]` **01B-02:** API Client Bearer token injection.
  * *Evidence:* `apps/frontend/src/api/client.ts`
* `[x]` **01B-03:** Auth pages (`SignInPage`, `SignUpPage`, `SSOPage`).
  * *Evidence:* `apps/frontend/src/pages/auth/`

### Phase 01 Checkpoint
* `[x]` Clerk SDK integrated on backend and frontend.
* `[x]` Unauthenticated requests to protected APIs return 401.
* `[x]` Security tests pass.

---

## PHASE 02 — Users / Organizations / Workspaces
**Depends On:** Phase 01  
**Blocks:** Phase 03, 08, 10, 11  
**Status:** `[x] COMPLETE (100%)`

### 02A — Multi-Tenant Schema & JIT Sync
* `[x]` **02A-01:** PostgreSQL `users` and `workspaces` tables.
  * *Evidence:* `003_create_users_table.sql`, `004_clerk_auth_workspaces.sql`, `006_finalize_workspace_schema.sql`
* `[x]` **02A-02:** Just-In-Time (JIT) provisioning in JWT middleware.
  * *Evidence:* `apps/backend/internal/middleware/clerk_jwt.go`, `repository/user.go`
* `[x]` **02A-03:** Multi-tenant workspace isolation in repository queries.
  * *Evidence:* `apps/backend/internal/repository/links.go`, `repository/campaigns.go`
  * *Tests:* `apps/backend/internal/handler/campaigns_test.go`

### Phase 02 Checkpoint
* `[x]` Cross-workspace data leakage blocked at repository layer.
* `[x]` JIT user/workspace creation verified.

---

## PHASE 03 — Link System
**Depends On:** Phase 02  
**Blocks:** Phase 04, 05, 08, 11  
**Status:** `[x] COMPLETE (100%)`

### 03A — Link Domain & Storage
* `[x]` **03A-01:** PostgreSQL `links` and `link_categories` tables.
  * *Evidence:* `001_create_links_table.sql`, `002_create_link_categories_and_attachments.sql`, `005_add_category_id_to_links.sql`
* `[x]` **03A-02:** Cryptographically secure Base62 short code generator with retry loop.
  * *Evidence:* `apps/backend/internal/service/links.go`, `pkg/base62/base62.go`
  * *Tests:* `apps/backend/pkg/base62/base62_test.go`, `apps/backend/internal/service/links_test.go`
* `[x]` **03A-03:** Link CRUD API handlers.
  * *Evidence:* `apps/backend/internal/handler/links.go`, `service/links.go`, `repository/links.go`
* `[x]` **03A-04:** Frontend Link management UI & React Query hooks.
  * *Evidence:* `apps/frontend/src/pages/links/LinksListPage.tsx`, `components/links/CreateLinkDrawer.tsx`, `hooks/useLinksQuery.ts`

### Phase 03 Checkpoint
* `[x]` Link CRUD operations pass unit and integration tests.
* `[x]` Shortcode collisions handle retries safely.
* `[x]` Tenant isolation strictly enforced.

---

## PHASE 04 — Redirect Performance + Redis Cache
**Depends On:** Phase 03  
**Blocks:** Phase 05, 11D  
**Status:** `[x] COMPLETE (100%)`

### 04A — Cache-Aside Redirect Engine
* `[x]` **04A-01:** Public edge redirect endpoint `GET /:slug`.
  * *Evidence:* `apps/backend/internal/handler/redirect.go`
* `[x]` **04A-02:** Redis Cache-Aside layer with 24h TTL.
  * *Evidence:* `apps/backend/internal/repository/redirect_cache.go`, `service/redirect.go`
* `[x]` **04A-03:** Singleflight request coalescing on cache misses.
  * *Evidence:* `apps/backend/internal/service/redirect.go` (`golang.org/x/sync/singleflight`)
* `[x]` **04A-04:** Cache invalidation on link update/delete.
  * *Evidence:* `apps/backend/internal/service/links.go`
* `[x]` **04A-05:** Graceful DB fallback when Redis is unavailable.
  * *Evidence:* `apps/backend/internal/service/redirect.go`
  * *Tests:* `apps/backend/internal/service/redirect_test.go`

### Phase 04 Checkpoint
* `[x]` Cache hits bypass PostgreSQL completely.
* `[x]` Redis outages do not crash or stall public redirects.
* `[x]` Invalidation verified on link mutations.

---

## PHASE 05 — Analytics Event Architecture
**Depends On:** Phase 04  
**Blocks:** Phase 06, 07, 08  
**Status:** `[x] COMPLETE (100%)`

### 05A — Event Contract & Anonymization
* `[x]` **05A-01:** Canonical `AnalyticsEvent` domain schema.
  * *Evidence:* `apps/backend/internal/model/analytics/analytics.go`
* `[x]` **05A-02:** IP address cryptographic hashing for privacy (`utils.HashIP`).
  * *Evidence:* `apps/backend/internal/handler/redirect.go`
* `[x]` **05A-03:** Non-blocking `AnalyticsPublisher` interface.
  * *Evidence:* `apps/backend/internal/model/analytics/analytics.go`
  * *Tests:* `apps/backend/internal/model/domain_test.go`

### Phase 05 Checkpoint
* `[x]` Event schema captures tenant, link, device, geo, and UTM metadata.
* `[x]` Publishing is completely decoupled from the redirect response.

---

## PHASE 06 — Redis Analytics Stream
**Depends On:** Phase 05  
**Blocks:** Phase 07  
**Status:** `[x] COMPLETE (100%)`

### 06A — Bounded Stream Publisher
* `[x]` **06A-01:** `RedisAnalyticsPublisher` with bounded channel (5000 events).
  * *Evidence:* `apps/backend/internal/service/redis_publisher.go`
* `[x]` **06A-02:** Non-blocking drop on queue overflow.
  * *Evidence:* `apps/backend/internal/service/redis_publisher.go`
* `[x]` **06A-03:** Graceful queue draining on server shutdown.
  * *Evidence:* `apps/backend/internal/server/server.go`
  * *Tests:* `apps/backend/internal/service/redis_publisher_test.go`

### Phase 06 Checkpoint
* `[x]` Ingestion shock-absorber buffers bursts up to channel capacity.
* `[x]` Shutdown drains pending events up to 5-second timeout.

---

## PHASE 07 — ClickHouse Ingestion
**Depends On:** Phase 06  
**Blocks:** Phase 08  
**Status:** `[x] COMPLETE (100%)`

### 07A — Consumer & OLAP Storage
* `[x]` **07A-01:** ClickHouse `analytics_events` schema (`MergeTree`, 90-day TTL).
  * *Evidence:* `apps/backend/internal/db/clickhouse.go`
* `[x]` **07A-02:** `RedisAnalyticsConsumer` using `XREADGROUP` consumer groups.
  * *Evidence:* `apps/backend/internal/service/clickhouse_consumer.go`
* `[x]` **07A-03:** Batch inserts (1000 items or 1s intervals) with post-insert `XACK`.
  * *Evidence:* `apps/backend/internal/service/clickhouse_consumer.go`
* `[x]` **07A-04:** Background `XAUTOCLAIM` loop for pending message recovery.
  * *Evidence:* `apps/backend/internal/service/clickhouse_consumer.go`
  * *Tests:* `apps/backend/internal/service/clickhouse_consumer_test.go`

### Phase 07 Checkpoint
* `[x]` At-least-once delivery guaranteed.
* `[x]` Unacknowledged crash recovery verified via testcontainers.

---

## PHASE 08 — Analytics Query API
**Depends On:** Phase 07  
**Blocks:** Phase 09, 11E  
**Status:** `[x] COMPLETE (100%)`

### 08A — Query Service & API Endpoints
* `[x]` **08A-01:** ClickHouse Repository aggregations (`uniqExact` counts).
  * *Evidence:* `apps/backend/internal/repository/clickhouse_analytics.go`
* `[x]` **08A-02:** Authenticated endpoints (`/summary`, `/timeseries`, `/top-links`, `/referrers`).
  * *Evidence:* `apps/backend/internal/handler/analytics.go`, `router/v1/v1.go`
* `[x]` **08A-03:** Enforced query bounds (30-day default, 1-year max limit).
  * *Evidence:* `apps/backend/internal/handler/analytics.go`
  * *Tests:* `apps/backend/internal/repository/clickhouse_analytics_test.go`

### Phase 08 Checkpoint
* `[x]` Multi-tenant workspace isolation guaranteed in ClickHouse queries.
* `[x]` Integration tests pass using testcontainers.

---

## PHASE 09 — Real Analytics Frontend
**Depends On:** Phase 08  
**Blocks:** Phase 10, 11F  
**Status:** `[x] COMPLETE (100%)`

### 09A — Dashboard Integration
* `[x]` **09A-01:** React Query analytics hooks (`useAnalyticsSummary`, `useAnalyticsTimeseries`, etc.).
  * *Evidence:* `apps/frontend/src/hooks/useAnalyticsQuery.ts`
* `[x]` **09A-02:** Live `AnalyticsPage.tsx` with date-range filters (`1h`, `24h`, `7d`, `30d`, `90d`).
  * *Evidence:* `apps/frontend/src/pages/analytics/AnalyticsPage.tsx`
* `[x]` **09A-03:** Live chart components (`TimeSeriesAreaChart`, `ReferrerBreakdownTable`, `TopLinksTable`).
  * *Evidence:* `apps/frontend/src/components/analytics/`

### Phase 09 Checkpoint
* `[x]` Mocks removed from core analytics explorer.
* `[x]` Loading skeletons and error boundaries verified.

---

## PHASE 10 — Profile / Account / Workspace UX
**Depends On:** Phase 02, 09  
**Blocks:** Phase 11F, 14  
**Status:** `[?] NEEDS VERIFICATION (75%)`

### 10A — Organization Management
* `[x]` **10A-01:** Clerk user profile and avatar rendering.
  * *Evidence:* `apps/frontend/src/pages/auth/`, Header navigation.
* `[~]` **10A-02:** Multi-Organization switcher and workspace management page.
  * *Evidence:* `apps/frontend/src/pages/settings/WorkspacesPage.tsx`
  * *Verification Required:* Verify that switching Clerk organization invalidates all React Query cache keys across Links, Campaigns, and Analytics.

### Phase 10 Checkpoint
* `[ ]` Seamless organization switching verified end-to-end without page reload data leaks.

---

## PHASE 11 — Campaigns + UTM Tracking
**Depends On:** Phase 03, 04, 07, 08  
**Blocks:** Phase 13  
**Status:** `[~] IN PROGRESS (92%)`

### 11A — Campaign Domain & Schema
* `[x]` **11A-01:** PostgreSQL `campaigns` table & `links` columns (Migration 007).
  * *Evidence:* `007_create_campaigns_and_utm.sql`, `model/campaign/campaign.go`, `model/link/link.go`
* `[x]` **11A-02:** ClickHouse schema expansion for campaign & UTM columns.
  * *Evidence:* `apps/backend/internal/db/clickhouse.go`
  * *Tests:* `apps/backend/internal/model/domain_test.go`

### 11B — Campaign REST API
* `[x]` **11B-01:** Full Campaign CRUD (`/api/v1/campaigns`).
  * *Evidence:* `apps/backend/internal/handler/campaigns.go`, `service/campaigns.go`, `repository/campaigns.go`
* `[x]` **11B-02:** Cross-workspace association denial.
  * *Evidence:* `apps/backend/internal/service/links.go`
  * *Tests:* `apps/backend/internal/handler/campaigns_test.go`

### 11C — Link + UTM Integration
* `[x]` **11C-01:** Link creation/update payloads accept `campaignId` and 5 UTM fields.
  * *Evidence:* `apps/backend/internal/model/link/dto.go`, `repository/links.go`
* `[x]` **11C-02:** `ON DELETE SET NULL` link campaign orphan policy.
  * *Evidence:* `007_create_campaigns_and_utm.sql`, `campaigns_test.go`

### 11D — Click-Time Attribution Pipeline
* `[x]` **11D-01:** Query-time UTM resolution in `PostgresRedirectRepository` (`LEFT JOIN`).
  * *Evidence:* `apps/backend/internal/repository/redirect.go`
* `[x]` **11D-02:** Redis Cache Hit/Miss parity for attribution snapshot.
  * *Evidence:* `apps/backend/internal/handler/redirect.go`
  * *Tests:* `apps/backend/internal/repository/redirect_parity_test.go`
* `[x]` **11D-03:** Bulk cache invalidation in `CampaignService` on campaign edits/deletion.
  * *Evidence:* `apps/backend/internal/service/campaigns.go`, `repository/link_campaign.go`

### 11E — Attribution Analytics API
* `[x]` **11E-01:** Campaign performance API `GET /api/v1/analytics/campaigns`.
  * *Evidence:* `apps/backend/internal/handler/analytics.go`, `repository/clickhouse_analytics.go`
* `[x]` **11E-02:** UTM performance API `GET /api/v1/analytics/utm?dimension=...`.
  * *Evidence:* `apps/backend/internal/handler/analytics.go`, `repository/clickhouse_analytics.go`
  * *Tests:* `apps/backend/internal/repository/clickhouse_analytics_test.go`

### 11F — Frontend Integration
* `[x]` **11F-01:** React Query hooks (`useCampaignsQuery.ts`, `useAnalyticsCampaigns`, `useAnalyticsUTM`).
  * *Evidence:* `apps/frontend/src/hooks/useCampaignsQuery.ts`, `hooks/useAnalyticsQuery.ts`
* `[x]` **11F-02:** Live `CampaignsPage.tsx`, `CampaignListTable.tsx`, `UTMPerformanceTable.tsx`.
  * *Evidence:* `apps/frontend/src/pages/growth/CampaignsPage.tsx`, `components/analytics/UTMPerformanceTable.tsx`
* `[x]` **11F-03:** Visual verification and complete frontend compilation pass.
  * *Evidence:* `apps/frontend/src/pages/growth/CampaignsPage.test.tsx`

### Phase 11 Checkpoint
* `[x]` Backend domain, API, and ClickHouse pipelines 100% complete and tested.
* `[x]` Redis cache parity and invalidation tested.
* `[ ]` Frontend integration test suite verified.

---

## PHASE 12 — Custom Domains & Edge TLS
**Depends On:** Phase 01, 03, 04  
**Blocks:** Phase 18  
**Status:** `[~] IN PROGRESS (10%)`

### 12A — Edge TLS Infrastructure
* `[x]` **12A-01:** Caddy proxy for on-demand Let's Encrypt certificates.

### 12B — Custom Domain Data Model
* `[x]` **12B-01:** PostgreSQL `custom_domains` schema and `links` relation.
* `[x]` **12B-02:** Workspace ownership and hostname normalization/uniqueness.
* `[x]` **12B-03:** Data model unit tests and parity verification.

### 12C — Core Domain API
* `[x]` **12C-01:** Domain CRUD endpoints and Clerk tenant isolation.
### 12D — DNS Verification Worker
* `[x]` **12D-01:** Background polling worker for TXT/CNAME validation.

### 12E — Routing Engine & Redis Updates
* `[x]` **12E-01:** Host-aware resolution and Redis composite cache keys.

### 12F — Internal TLS Authorization API
* `[x]` **12F-01:** Build `GET /api/internal/tls/ask` for Caddy dynamic SSL.

### 12G — Analytics Expansion
* `[x]` **12G-01:** Expand `AnalyticsEvent` struct with `custom_domain_id` and `hostname`.
* `[x]` **12G-02:** Alter ClickHouse schema to include nullable domain columns.
* `[x]` **12G-03:** Support caching domain properties in Redis for parity.
* `[x]` **12G-04:** Extend `AnalyticsRepository` API with domain metrics support.

### 12H — Frontend Domains UI
* `[x]` **12H-01:** Domain management page and list.
* `[x]` **12H-02:** Add domain flow and modal.
* `[x]` **12H-03:** DNS verification instructions UX.
* `[x]` **12H-04:** Delete domain flow.
* `[x]` **12H-05:** React Query and Clerk org isolation.
* `[x]` **12H-06:** Build, test, and typecheck validation.

### 12I — E2E Security & Routing Verification
* `[x]` **12I-01:** Production verification and penetration testing limits.
* `[x]` **12I-02:** Strict workspace segregation testing.
* `[x]` **12I-03:** Host header injection tests.
* `[x]` **12I-04:** TLS Authorization API validation.
* `[x]` **12I-05:** End-to-end caching parity checks.
* `[x]` **12I-06:** Final Phase 12 Sign-Off.

### Phase 12 Checkpoint
* `[x]` Domain schema
* `[x]` Workspace ownership
* `[x]` Domain uniqueness
* `[x]` Link relationship
* `[x]` Migration
* `[x]` Tests
* `[x]` DNS verification
* `[x]` Verification worker
* `[x]` Routing
* `[x]` Redis compatibility
* `[x]` Cache invalidation
* `[ ]` TLS
* `[ ]` Security
* `[x]` API
* `[ ]` Frontend
* `[ ]` Integration tests
* `[ ]` Failure tests
* `[ ]` Documentation
* `[ ]` Production verification

---

## PHASE 13 — Multi-Touch Attribution

### 13A — Data Model & Migrations
* `[x]` **13A-01:** Create ClickHouse `conversions` table migration.
* `[x]` **13A-02:** Create ClickHouse `analytics_events.event_id` bloom filter migration.

### 13B — URL Decoration & Tracking
* `[x]` **13B-01:** Update `RedirectHandler` to conditionally append `?flux_cid=<event_id>`.

### 13C — Conversion Ingestion API
* `[x]` **13C-01:** Implement `POST /api/v1/events/track` handler (Public endpoint).
* `[x]` **13C-02:** Implement ClickHouse stream consumer for `conversions`.

### 13D — Attribution API / Engine
* `[x]` **13D-01:** Implement `GetConversionsWithTouchpoints` in ClickHouse.
* `[x]` **13D-02:** Wire `calculator.go` to real data via `GET /api/v1/analytics/attribution`.

### 13E — Frontend Attribution UI
* `[x]` **13E-01:** Connect `AttributionPage.tsx` to real attribution queries.

### 13F — Final Verification
* `[ ]` **13F-01:** Perform end-to-end attribution checks and audit tenant isolation.

### Phase 13 Checkpoint
* `[ ]` Multi-touch attribution calculations match test models against real clickstream.

---

## PHASE 14 — Billing / Plans / Usage
**Depends On:** Phase 02, 08
**Blocks:** Phase 18
**Status:** `[x] COMPLETE (100%)`

### 14A — Stripe Integration
* `[x]` **14A-01:** PostgreSQL `subscriptions` schema and customer mapping.
* `[x]` **14A-02:** Stripe webhook listener for subscription lifecycle.
* `[x]` **14A-03:** Feature tier enforcement middleware (link limits, analytics retention).
* `[x]` **14A-04:** Frontend `BillingPage.tsx` connection to Stripe Customer Portal.

### Phase 14 Checkpoint
* `[x]` Subscription state changes enforce limits accurately.

---

## PHASE 15 — Webhooks / Integrations
**Depends On:** Phase 05, 08  
**Blocks:** None  
**Status:** `[ ] NOT STARTED (0%) - DEFERRED`

### 15A — Outbound Event Delivery
* `[ ]` **15A-01:** PostgreSQL `webhooks` table and secret token generation.
* `[ ]` **15A-02:** Async webhook delivery worker with HMAC-SHA256 signing.
* `[ ]` **15A-03:** Retry queue with exponential backoff and dead-letter queue.
* `[ ]` **15A-04:** Frontend `WebhooksPage.tsx` and delivery logs.

### Phase 15 Checkpoint
* `[ ]` Outbound webhooks deliver reliably with verified signatures.

---

## PHASE 16 — AI / Intelligence
**Depends On:** Phase 07, 08  
**Blocks:** None  
**Status:** `[ ] NOT STARTED (0%) - DEFERRED`

### 16A — Machine Learning & Insights
* `[ ]` **16A-01:** Click anomaly detection (traffic spike / bot flood identification).
* `[ ]` **16A-02:** Automated UTM campaign suggestion engine.
* `[ ]` **16A-03:** Connect `AIInsightsPage.tsx` to live backend telemetry.

### Phase 16 Checkpoint
* `[ ]` Anomaly alerts fire with low false-positive rates.

---

## PHASE 17 — Production Hardening
**Depends On:** Phase 01–11  
**Blocks:** Phase 18  
**Status:** `[~] IN PROGRESS (45%)`

### 17A — Security & Reliability Audit
* `[x]` **17A-01:** Cryptographic randomness Base62 generator (`crypto/rand`).
* `[x]` **17A-02:** Ordered graceful shutdown of HTTP, Redis, and ClickHouse workers.
* `[x]` **17A-03:** Fail-closed secret configuration.
* `[ ]` **17A-04:** Per-IP / Per-Token token-bucket rate limiting middleware.
* `[ ]` **17A-05:** Automated load testing (50k req/sec) and race-condition audit.
* `[ ]` **17A-06:** Secret rotation and automated disaster recovery runbooks.

### Phase 17 Checkpoint
* `[ ]` 0 security audit findings.
* `[ ]` Load tests pass without memory leaks or goroutine accumulation.

---

## PHASE 18 — Scale / Operations / Continuous Development
**Depends On:** Phase 17  
**Blocks:** None  
**Status:** `[ ] NOT STARTED (0%) - DEFERRED`

### 18A — Global Edge Infrastructure
* `[ ]` **18A-01:** Multi-region edge deployment configuration.
* `[ ]` **18A-02:** Global Anycast DNS routing.
* `[ ]` **18A-03:** Cross-region Redis read-replica synchronization.
* `[ ]` **18A-04:** Centralized OpenTelemetry distributed tracing and metrics dashboard.

### Phase 18 Checkpoint
* `[ ]` Global p99 redirect latency < 10ms globally.

---
*Roadmap reconciled on 2026-08-30 against live repository code and test artifacts.*
