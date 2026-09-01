# Verification Log

| Feature | Test Command / Action | Expected | Actual | Status | Date |
|---------|-----------------------|----------|--------|--------|------|
| Links CRUD | Manual API Test | Link created in DB | Link created successfully | VERIFIED | 2026-08-29 |
| Redirect | `curl -i localhost:8080/slug` | HTTP 301/302 | HTTP 301/302 returned | VERIFIED | 2026-08-29 |
| Analytics | `curl localhost:8080/api/v1/analytics/summary` | 200 OK | 500 Error (Provider is nil) | BROKEN | 2026-08-29 |
| Native Auth | `curl POST /api/v1/auth/signup` | HTTP 201 Created | HTTP 201 | VERIFIED | 2026-08-29 |
| Native Auth | `curl POST /api/v1/auth/login` | HTTP 200 OK | HTTP 200 | VERIFIED | 2026-08-29 |
| Native Auth | `curl GET /api/v1/me` with Bearer | HTTP 200 OK | HTTP 200 | VERIFIED | 2026-08-29 |

## V-002: Clerk End-to-End Verification
* **Date:** 2026-08-29
* **Tested By:** Agent
* **Status:** Passed

### Verified:
- **Authentication**: JWT middleware successfully validates tokens against Clerk JWKS.
- **JIT Synchronization**: Users and Workspaces are properly provisioned Just-In-Time to Postgres.
- **Cross-Organization Data Ownership**: Confirmed `LinksHandler` uses the injected `tenant_id` from Context. User A cannot access User B's link directly via ID (404 Not Found), nor update or delete it.
- **Invalid Auth**: Missing token requests correctly return 401. Expired tokens are rejected.
- **Public Redirects**: Short code redirects (e.g. `/:shortCode`) still resolve successfully without Clerk tokens.
- **Database Integrity**: `clerk_user_id` and `clerk_org_id` uniquely constrained. `category_id` was missing from the `links` table in Postgres schema and was patched in Migration 005.

### Not Tested:
- **Clerk Organization Switching**: The configured Clerk instance has Organizations disabled, thus returning `org_id=NULL` (Personal Workspaces) by default. Tested isolation among Personal Workspaces effectively.

### Security Risks / Technical Debt:
- **workspace_members**: Currently populates during JIT sync to map Personal Workspaces, but RBAC is purely deferred to Clerk token claims. Consider migrating personal workspaces to use `owner_id` directly to drop this join table.


## V-003: Database Not-Found HTTP Mapping (BUG-001A)
* **Date:** 2026-08-29
* **Tested By:** Agent
* **Status:** Passed

### Verified:
- Standard `pgx.ErrNoRows` from the repository layer correctly bubble up through `sqlerr.HandleError` as a 404 domain error.
- The `CustomHTTPErrorHandler` intercepts the domain error and forces an HTTP `404 Not Found` JSON response on the wire (previously defaulting to HTTP 500).
- Cross-organization isolation properly yields an actual HTTP 404 to the unprivileged client.
- Malformed UUID/Bad Requests appropriately yield an HTTP 400.

## V-004: Analytics Event Generation & Isolation
* **Date:** 2026-08-29
* **Tested By:** Agent
* **Status:** Passed

### Verified:
- Analytics events are successfully generated upon a valid public short link redirect (`GET /:shortCode`).
- Events accurately contain the parent `WorkspaceID` extracted safely from the Database backend, NOT requested by the client, maintaining cross-tenant data ownership.
- Failing to publish an analytics event explicitly does *not* break or stall the HTTP redirect to the user.

## V-003: Production Readiness & Remediation Audit
* **Date:** 2026-08-30
* **Tested By:** Agent
* **Status:** Passed

### Verified:
- **Authentication**: `X-Test-Clerk-*` headers completely purged from production middleware. Identity spoofing is impossible. Tested via `auth_test.go`.
- **Shutdown**: `server.go` correctly shuts down HTTP server before terminating Redis and ClickHouse consumer threads.
- **Short-Code Generation**: `generateShortCode()` uses `crypto/rand` securely and encodes via `pkg/base62`. DB collision retry loop explicitly tested.
- **Configuration**: Server fails closed (`err != nil`) if `CLERK_SECRET_KEY` is omitted. Tested via `config_test.go`.
- **Frontend Mocks**: Removed fake link fallback in `LinksListPage.tsx`. Error boundaries properly catch backend issues.

## V-004: Redis Redirect Cache
* **Date:** 2026-08-30
* **Tested By:** Agent
* **Status:** Passed

### Verified:
- **Cache Hits**: Redirect responses pull natively from Redis without touching PostgreSQL.
- **Cache Misses**: Concurrent bursts are coalesced securely using `golang.org/x/sync/singleflight`, resulting in a single database lookup for the target slug.
- **Invalidation**: `UpdateLink` and `DeleteLink` mutations properly issue a cache invalidation request via `RedisRedirectCache.Delete`.
- **Failure Isolation**: Simulated cache timeouts and unmarshaling faults gracefully fallback to Postgres database calls without harming the redirect HTTP lifecycle.

## V-005: Phase 11F Frontend Integration Verification
* **Date:** 2026-08-30
* **Tested By:** Agent
* **Status:** Passed

### Verified:
- **Campaign CRUD**: Frontend `useCampaignsQuery` queries correctly map to `/api/v1/campaigns` using Clerk orgId scoping.
- **Link Mutators**: `CreateLinkDrawer` correctly handles `campaignId`, `utmSource`, `utmMedium`, `utmCampaign` values.
- **Analytics Visibility**: `AnalyticsPage` correctly implements `UTMPerformanceTable` fetching data from `/api/v1/analytics/utm` and `/api/v1/analytics/campaigns`.
- **Compilation**: Full `npm run build` succeeds with zero TypeScript errors across `@flux/zod`, `@flux/openapi`, and `@flux/frontend`.
- **Isolation**: React Query keys correctly scope to the tenant. Backend mocked data successfully eradicated.

### V-006: Custom Domain Data Model (Phase 12B)
- **Status**: PASSED
- **Method**: Automated DB integration tests via TestContainers (`TestDomainRepository`).
- **Assertions**:
  - `custom_domains` successfully persists domains mapped to specific workspaces.
  - Cross-tenant requests correctly return `ErrNotFound`.
  - Duplicate domains (globally) trigger SQL constraint errors.
  - Case-normalization and trailing-dot protections successfully block invalid domains.
  - Deleting a domain safely sets the associated `links.custom_domain_id` to NULL without cascade deleting the links.

### V-007: Custom Domain Routing & Security (Phase 12E)
- **Status**: PASSED
- **Method**: Automated DB integration tests via TestContainers (`TestRedirectSecurity_CustomDomains` and `TestRedirectParity_CacheHitMiss_UTMResolution`).
- **Assertions**:
  - `domain-a.com/slug` successfully routes to Workspace A's link.
  - Cross-tenant requests (`domain-a.com` trying to access Workspace B's slug) correctly return `ErrNotFound`.
  - Unverified, pending, and disabled domains successfully fail to route.
  - Links without a custom domain remain accessible via the platform default domain.
  - Parity is strictly preserved between Redis cache hits and PostgreSQL cache misses.

### V-008: Attribution Conversions Schema & Integrity (Phase 13A)
- **Status**: PASSED
- **Method**: Automated DB integration tests via TestContainers (`TestPhase13_ClickHouseSchemaMigration`).
- **Assertions**:
  - `conversions` table successfully established with correct workspace isolation.
  - Arrays (`click_ids`) persist correctly to ClickHouse.
  - Schema creation is idempotent and safely ignores pre-existing tables.
  - The `idx_event_id` Data-Skipping Bloom Filter index applies correctly to `analytics_events`.

### V-009: Click Tracking URL Decoration (Phase 13B)
- **Status**: PASSED
- **Method**: Go unit and integration tests (`TestRedirectHandler_URLDecoration`, `TestRedirectHandler_URLDecoration_CacheParity`).
- **Assertions**:
  - `flux_cid` successfully appends to redirect target URLs accurately.
  - Native parameters, encoded sequences, and hash fragments correctly bypass overwrite or stripping.
  - Parity holds definitively between Postgres cache misses and Redis cache hits on target mappings.
  - Redundant or malicious `flux_cid` parameters are strictly overridden natively.

### Phase 13D: Attribution API / Engine
- Verify multi-touch linear conversion calculates revenue precisely evenly on click pipelines across `GET /api/v1/analytics/attribution`.
