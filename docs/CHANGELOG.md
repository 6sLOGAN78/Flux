# Changelog

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
