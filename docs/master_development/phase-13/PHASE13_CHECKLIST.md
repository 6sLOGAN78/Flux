# Phase 13 Checklist

### 13A — Data Model & Migrations
* `[ ]` **13A-01:** Create ClickHouse `conversions` table migration.
* `[ ]` **13A-02:** Create ClickHouse `analytics_events.event_id` bloom filter migration.
* `[ ]` **13A-03:** Update Go `db` package to apply these migrations safely.

### 13B — URL Decoration & Tracking
* `[ ]` **13B-01:** Update `RedirectHandler` to conditionally append `?flux_cid=<event_id>`.
* `[ ]` **13B-02:** Ensure URL hash fragments (`#`) are preserved correctly after query params.

### 13C — Ingestion API
* `[ ]` **13C-01:** Implement `POST /api/v1/events/track` handler (Public endpoint).
* `[ ]` **13C-02:** Validate workspace public Client ID and rate limits.
* `[ ]` **13C-03:** Publish `ConversionEvent` payload to Redis Stream.
* `[ ]` **13C-04:** Implement ClickHouse stream consumer for `conversions`.

### 13D — Attribution API
* `[ ]` **13D-01:** Implement `AttributionRepository.GetConversionsWithTouchpoints()`.
* `[ ]` **13D-02:** Wire `calculator.go` to the fetched data in the Service layer.
* `[ ]` **13D-03:** Expose `GET /api/v1/analytics/attribution` endpoint.

### 13E — Frontend UI
* `[ ]` **13E-01:** Update `@flux/openapi` and `@flux/zod` with new endpoints.
* `[ ]` **13E-02:** Build `AttributionPage.tsx` with Model Selector dropdown.
* `[ ]` **13E-03:** Render data table visualizing attributed revenue and conversions.

### 13F — Final Verification
* `[ ]` **13F-01:** Integration tests for ClickHouse joins and caching.
* `[ ]` **13F-02:** Security audit for cross-tenant CID leakage.
