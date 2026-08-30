# Bugs and Issues

## BUG-001
**Severity:** CRITICAL
**Status:** FIXED
**Title:** Analytics endpoints panic/fail due to nil provider
**Description:** `GetSummary` and `GetLinkMetrics` returned 500 errors because `AnalyticsProvider` was initialized as `nil`.
**Root Cause:** `analyticsHandler := handler.NewAnalyticsHandler(nil)`
**Fix:** Bounded async Redis publisher, ClickHouse consumer, and Analytics Query API fully implemented and wired. Verified via production readiness audit.

## BUG-002
**Severity:** HIGH
**Status:** OPEN
**Title:** Redirects do not use Redis cache
**Location:** `apps/backend/internal/server/server.go`
**Affected Feature:** FEAT-005
**Description:** The cache mechanism is built (`RedisRedirectCache`) but is passed as `nil` when initializing the service.
**Root Cause:** `redirectSvc := service.NewRedirectService(redirectRepo, nil)`
**Impact:** All redirects hit PostgreSQL, missing the performance targets.

## BUG-003
**Severity:** HIGH
**Status:** FIXED
**Title:** Redirect clicks are not tracked
**Description:** Resolving a short link did not emit any events or write to ClickHouse.
**Fix:** Redirect handler now fires AnalyticsEvent to bounded Redis Stream, ingested into ClickHouse. Verified via production readiness audit.

## BUG-004
**Severity:** MEDIUM
**Status:** FIXED
**Title:** Shortcodes do not use Base62 encoder / PRNG collision risk
**Description:** `generateShortCode()` used a per-call time-seeded random math generator and ignored Base62.
**Fix:** Modified `links.go` to use `crypto/rand` bound to `base62.Encode`, wrapped in a database unique-constraint retry loop (max 5 retries).

## BUG-005
**Severity:** CRITICAL
**Status:** PARTIALLY FIXED
**Title:** Frontend pages rely on mock state / Demo auth bypass
**Description:** Links, Auth, and Analytics have been fully migrated to real APIs (Clerk Auth, React Query to real backend). Campaigns and Billing still use local mocked arrays.
**Fix:** Mock optimistic success removed from `LinksListPage.tsx` onError blocks. Auth backdoor (X-Test-Clerk headers) completely stripped from production middleware.

## BUG-006
**Severity:** CRITICAL
**Status:** FIXED
**Title:** Database schema missing critical tables
**Description:** The Postgres schema only contained links/categories.
**Fix:** Clerk-integrated users and workspaces schema implemented securely.

### BUG-004 (Schema issue)
* **Status:** FIXED
* **Details:** `category_id` missing from `links` table. Fixed in Migration 005.

### BUG-001A
* **Status:** FIXED
* **Details:** `pgx.ErrNoRows` returning HTTP 500 on the wire despite being logged as a 404. Fixed by registering a custom `errs.CustomHTTPErrorHandler`.

### BUG-007 (Audit finding)
* **Severity:** HIGH
* **Status:** FIXED
* **Title:** Inverted graceful shutdown causes analytics data loss
* **Description:** HTTP server was shut down after AnalyticsPublisher.
* **Fix:** Reordered lifecycle in `server.go` to stop Echo before draining ingestion queues.

### BUG-008 (Audit finding)
* **Severity:** HIGH
* **Status:** FIXED
* **Title:** Missing CLERK_SECRET_KEY validation
* **Description:** App could boot without auth secrets.
* **Fix:** `config.Validate()` now fails closed securely.
