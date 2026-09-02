# Task 14A-03 Report: Feature Tier Enforcement

## Objective
Implement server-side enforcement of billing-plan limits (link quotas, analytics retention) utilizing the authoritative Stripe subscription state stored in PostgreSQL.

## Work Completed
1. **Plan Tiers Centralized**:
   - Upgraded `PlanTierLimits` in `apps/backend/internal/modules/billing/stripe.go` to include `AnalyticsRetentionDays` (Free: 7, Pro: 30, Business: 365).

2. **Concurrent Link Quota Enforcement (Service/DB Boundary)**:
   - Modified `LinkRepository.CreateLink` to dynamically construct an `INSERT INTO ... SELECT ... WHERE (SELECT COUNT(...) FROM links) < max_links`.
   - Prevented race conditions that naïve `SELECT count(*) THEN INSERT` logic would have allowed.
   - Refactored `LinkService.CreateLink` to query the authenticated user's active billing subscription, falling back to the `free` tier limits, and passing the quota directly into the database repository layer.
   - Fixed edge cases related to `pgx.ErrNoRows` consumption during empty query results and gracefully mapped them to a newly added `ErrQuotaExceeded` / HTTP `402 Payment Required`.

3. **Analytics Date Boundaries**:
   - Added `AnalyticsHandler.enforceRetention` to query the current workspace's subscription limit and silently cap the `from` date query parameters directly in memory before they ever reach the ClickHouse infrastructure.
   - Applied this constraint across all analytics timeseries, dimension, summary, and attribution endpoints, securing historical data based on tier entitlement without permanently deleting any records.

4. **Integration Testing**:
   - Created `apps/backend/internal/service/links_test.go` and implemented `TestLinkService_Quota` using `testcontainers-go`.
   - Proved strict multi-tenant isolation, automated default to free-tier quotas (1,000 links limit), and dynamic quota expansions (upgrading to Pro seamlessly permits subsequent creations).

## Status
Task 14A-03 is 100% Complete. The service now respects billing plan entitlements using robust backend/database mechanisms.
