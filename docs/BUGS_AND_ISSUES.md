# Bugs and Issues

## BUG-001
**Severity:** CRITICAL
**Status:** OPEN
**Title:** Analytics endpoints panic/fail due to nil provider
**Location:** `apps/backend/internal/handler/analytics.go`, `apps/backend/internal/server/server.go`
**Affected Feature:** FEAT-003
**Description:** `GetSummary` and `GetLinkMetrics` return 500 errors because `AnalyticsProvider` is initialized as `nil` in `server.go`.
**Root Cause:** `analyticsHandler := handler.NewAnalyticsHandler(nil)`
**Impact:** Frontend analytics pages cannot load real data.

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
**Status:** OPEN
**Title:** Redirect clicks are not tracked
**Location:** `apps/backend/internal/handler/redirect.go`
**Affected Feature:** FEAT-004
**Description:** Resolving a short link does not emit any events or write to ClickHouse.
**Root Cause:** Implementation was never written.
**Impact:** The analytics platform has no data to display.

## BUG-004
**Severity:** MEDIUM
**Status:** OPEN
**Title:** Shortcodes do not use Base62 encoder
**Location:** `apps/backend/internal/service/links.go`
**Affected Feature:** FEAT-014
**Description:** `generateShortCode()` uses a random math generator instead of the custom `pkg/base62` implementation.
**Root Cause:** Likely a temporary mock implementation that was never replaced.

## BUG-005
**Severity:** CRITICAL
**Status:** VERIFIED
**Title:** Frontend pages rely on mock state / Demo auth bypass
**Location:** `apps/frontend/src/pages/*`, `AuthContext.tsx`
**Affected Feature:** FEAT-015, FEAT-006, FEAT-007, FEAT-008, FEAT-010
**Description:** Previously, auth bypassed passwords and used fake mock state. Now, `AuthContext` uses real JWT `/auth/login` and `/auth/signup` against PostgreSQL. Many other pages (Campaigns, Billing) still use local mocked arrays.
**Root Cause:** Lack of backend implementation initially.
**Impact:** Auth is fixed. Other pages still need backend endpoints.

## BUG-006
**Severity:** CRITICAL
**Status:** OPEN
**Title:** Database schema missing critical tables
**Location:** `apps/backend/internal/database/migrations/*`
**Affected Feature:** ALL
**Description:** The Postgres schema only contains `links`, `link_categories`, and `link_attachments`. It is missing users, tenants, campaigns, domains, webhooks, invoices, etc.
**Root Cause:** Migrations were never written.

### BUG-004
* **Issue:** `category_id` missing from `links` table
* **Status:** FIXED
* **Details:** The second migration created the `link_categories` table but forgot to add the `category_id` column to `links`. Caused a SQLSTATE 42703 error in CreateLink. Fixed in Migration 005.


### BUG-001A
* **Issue:** `pgx.ErrNoRows` returning HTTP 500 on the wire despite being logged as a 404.
* **Status:** FIXED
* **Details:** Echo's default HTTP error handler did not understand our custom `errs.HTTPError` type returned by `sqlerr.HandleError`, causing it to swallow the 404 status and return a generic 500. Fixed by registering a custom `errs.CustomHTTPErrorHandler` on the Echo server instance that maps domain errors to standard HTTP JSON schemas correctly.
