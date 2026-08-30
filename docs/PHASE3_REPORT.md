# CAMPAIGNS + UTM — PHASE 3 COMPLETE

## Click-Time Attribution Architecture
The platform now successfully implements a strict immutable snapshot architecture. At the instant of a link click, the current Campaign ID and UTM metadata are resolved, bundled into the `LinkRedirectTarget`, injected into the `AnalyticsEvent`, serialized into Redis, and eventually flushed into ClickHouse.

## UTM Resolution Rule
The `PostgresRedirectRepository` enforces the requested resolution algorithm at query time via `LEFT JOIN`:
```text
resolved_utm_campaign = link.utm_campaign (if present) OR campaign.utm_campaign (if present)
```
This means Link overrides natively take precedence over Campaign defaults.

## Campaign Snapshot
When a `short_code` resolves via the `RedirectService`, `CampaignID` and all 5 UTM parameters (Source, Medium, Campaign, Term, Content) are eagerly loaded into the struct.

## Redis Cache Hit / Miss Behavior & Parity
**Cache Miss:** `GetBySlug` queries Postgres, performs the `LEFT JOIN` and resolution, returns a `LinkRedirectTarget`, which the service then caches as JSON in Redis.
**Cache Hit:** Redis retrieves the exact JSON representation. No DB queries are made. 
This parity ensures `AnalyticsEvent` creation is identical and deterministic whether the target came from RAM or Disk. Verified explicitly via `TestRedirectParity_CacheHitMiss_UTMResolution`.

## Cache Invalidation
1. **Link Update:** Already implemented (Phase 2), modifying a link drops its cache immediately.
2. **Campaign Update / Delete:** Modified `CampaignService` to accept `LinkRepository` and `RedirectCache`. When a campaign's defaults are modified or the campaign is deleted, it queries all associated link `short_codes` and surgically evicts them from the Redis Cache. The next redirect fetches fresh data (or `NULL` campaign).

## AnalyticsEvent Changes
`AnalyticsEvent` explicitly captures `CampaignID`, `UTMSource`, `UTMMedium`, `UTMCampaign`, `UTMTerm`, and `UTMContent`. The `RedirectHandler` initializes them from the `LinkRedirectTarget`.

## Redis Stream & ClickHouse Verification
- `RedisAnalyticsPublisher` natively serializes the new `omitempty` string pointers into JSON.
- `RedisAnalyticsConsumer` consumes the JSON and uses `batch.Append` targeting the 6 new ClickHouse columns safely with `(*string)(nil)` nullability support. Historical integrity is 100% guaranteed as no Postgres joins happen post-click.

## Multi-Tenant Security
No changes were needed to the public redirect handler. Validation of tenant ownership occurs purely at Link CRUD time (Phase 2). Because the attribution state is resolved strictly from trusted backend components (`campaign_id` is an FK owned by a tenant), the `AnalyticsEvent` intrinsically inherits safe workspace ownership.

## Failure Semantics & Concurrency Verification
Analytics publishing remains entirely non-blocking. A failed `ClickHouse` consumer or `Redis` queue full-buffer explicitly drops events or retries *without* interrupting the `301 Redirect`.
`go test -race ./...` passed across the entire repository.

## Tests Added & Passed
- `TestRedirectParity_CacheHitMiss_UTMResolution`
All `go test ./...` and `go test -race ./...` suites pass. No analytics regressions were detected.

## Files Changed
- `apps/backend/internal/repository/redirect.go` (Resolved queries)
- `apps/backend/internal/repository/redirect_parity_test.go` (Parity integration test)
- `apps/backend/internal/handler/redirect.go` (Snapshot creation)
- `apps/backend/internal/service/campaigns.go` (Bulk cache eviction)
- `apps/backend/internal/repository/link_campaign.go` (Shortcode bulk lookup)
- `apps/backend/internal/server/server.go` (DI wiring)

## Production Readiness
PHASE 3 COMPLETE. 

The backend tracking and analytics insertion pipeline is fully operational. You are clear to proceed with Phase 4 (Analytics API) or Phase 5 (Frontend).
