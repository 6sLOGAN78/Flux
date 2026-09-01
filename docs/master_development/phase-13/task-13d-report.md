# Task 13D Report - Attribution API / Engine

## 13D Implementation Summary
The core analytical Attribution Engine has been successfully implemented on top of the ClickHouse conversion and analytics event ingestion pipeline. A secure `/api/v1/analytics/attribution` endpoint was introduced that safely joins historical conversion data with chronological click footprints (touchpoints). 

## Attribution Architecture
The `GetConversionsWithTouchpoints` method operates in ClickHouse using native `ARRAY JOIN` functionality against ingested event metadata. The bounded analytical fetch returns precisely sorted touchpoint pipelines scoped to tenant boundaries. Once pulled into the backend Go application, it parses through `calculator.go` computing multi-touch rules seamlessly (e.g. `first_touch`, `last_touch`, `linear`, `position_based`, `time_decay`).

## ClickHouse Query / Touchpoint Resolution
The query explicitly performs a nested scope:
1. Filters conversions by `workspace_id`, parsing bounded date ranges, and sorting by chronological event timestamp (`LIMIT 1 BY conversion_id` enforces deduplication safely).
2. Performs an `ARRAY JOIN click_ids` mapping arrays to individual rows.
3. Operates an `INNER JOIN analytics_events e ON e.event_id = c.cid AND e.workspace_id = c.workspace_id` enforcing identical tenant limits inherently discarding alien/orphaned `click_ids`.

## Tenant Isolation
The ClickHouse query implements strict tenancy. Using explicitly named placeholders (`@workspace_id`), it is impossible to read or attach data spanning multiple organizations. Invalid CID references belonging to other workspaces natively result in an absent row join on the `INNER JOIN` constraint, effectively ignoring spoofed/injected conversion entries gracefully.

## CID Spoofing Protection
CID references injected intentionally by malicious tracking clients pointing to an opponent organization's `flux_cid` footprint simply fail the deterministic ClickHouse `INNER JOIN ... AND e.workspace_id = @workspace_id` constraint.

## Duplicate Conversion Handling
The query utilizes `LIMIT 1 BY conversion_id ORDER BY timestamp DESC` inside the inner sub-query block, executing deterministic single-conversion extraction out of `MergeTree`. Double-counted duplicate ingestion requests are flattened analytically before executing the actual array join.

## Attribution Models
The standard mathematical structures exported by the `calculator.go` logic (`first_touch`, `last_touch`, `linear`, `time_decay`, `position_based`) are entirely accessible. The endpoint requires an explicit query parameter `?model=linear`, bouncing unsupported models rapidly with `400 Bad Request`.

## Revenue Allocation
Revenue attribution exactly mirrors original calculator implementation rules. Linear evenly splits the value across touchpoints without dropping a fraction, and time decay utilizes half-life metrics effectively without creating double-spend revenue allocation bugs.

## API Contract
The `GET /api/v1/analytics/attribution` OpenAPI specification has been fully appended with parameter types (`from`, `to`, `model`) and properly mapping array responses corresponding directly to the `packages/zod` schema definitions.

## Query Limits
The analytical limit logic is natively bounded by standard date-range boundaries (default 30 days) utilizing the `parseDateRange` echo context mapping standard across the Flux application.

## Performance
The ClickHouse nested Array Join explicitly filters conversions locally before attempting the analytical event join limit, dramatically reducing required scan space across millions of rows by shrinking cardinality effectively ahead of the heavier execution pipeline.

## Security Tests
Integrated endpoint tests run against invalid/missing JWT context schemas ensuring strict 401 rejections along with missing tenant context isolation boundaries. Handlers specifically assert unsupported models throw clean HTTP 400 validations.

## Integration Tests
A dedicated ClickHouse Testcontainer test (`TestClickHouse_AttributionQuery_Integration`) inserts overlapping multi-tenant conversions/events and proves tenant data effectively drops unreachable spoofed identifiers across isolated domains.

## Exact Commands Executed
```bash
go test -v -run TestClickHouse_AttributionQuery_Integration ./internal/repository
go test -v -run TestAnalyticsHandler_GetAttribution ./internal/handler/...
go test -v ./...
git add . && git commit -m "feat: implement attribution api and clickhouse query logic (13D)" && git push
```

## Files Changed
- `apps/backend/internal/repository/clickhouse_attribution.go`
- `apps/backend/internal/repository/clickhouse_attribution_test.go`
- `apps/backend/internal/handler/analytics_attribution.go`
- `apps/backend/internal/handler/analytics_attribution_test.go`
- `apps/backend/internal/router/v1/v1.go`
- `apps/backend/internal/repository/interfaces.go`
- `packages/openapi/openapi.json`

## Documentation Updated
- `docs/master_development/phase-13/task-13d-report.md`
- `docs/master_development/MASTER_ROADMAP.md`
- `docs/PROJECT_STATE.md`
- `docs/VERIFICATION.md`
- `docs/CHANGELOG.md`

## Checkpoint
```text
13A [x]
13B [x]
13C [x]
13D [x]
13E [ ]
13F [ ]
```

## Remaining Phase 13 Work
13E — Attribution Frontend
13F — Final Verification

## Next Recommended Task
13E — Attribution Frontend
