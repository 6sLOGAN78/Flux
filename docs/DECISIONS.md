# Architecture Decisions

## DEC-001
**Date:** 2026-08-29
**Decision:** Focus on wiring existing foundation before expanding features.
**Context:** The repository presents a massive surface area, but core features (Redis caching, ClickHouse analytics, Postgres schemas for tenants/users) are missing or passed as `nil`.
**Chosen Approach:** Do not build new UI or modules. First, connect the existing backend modules (Redis cache, ClickHouse ingestion) to the live request lifecycle.
**Affected Components:** `server.go`, `redirect.go`, `links.go`

## DEC-002
**Date:** 2026-08-29
**Decision:** Implement Native JWT Authentication.
**Context:** The codebase has Clerk support but defaults to a fake local auth if the Clerk key is missing. The `AuthService` already implements JWT and bcrypt.
**Chosen Approach:** Build a native `users` table and `/auth/signup`, `/auth/login` endpoints. Use the existing JWT middleware to protect routes.
**Affected Components:** `server.go`, `router.go`, `v1.go`, `SignInPage.tsx`, `SignUpPage.tsx`, `AuthContext.tsx`.

## DEC-003: Migrate Auth & Workspaces to Clerk
* **Date:** 2026-08-29
* **Status:** Implemented
* **Decision:** Replaced native JWT authentication and RBAC with Clerk. Clerk is now the single source of truth for Users, Organizations, Roles, and Sessions.
* **Reasoning:** Reduces security overhead, offloads identity management, and simplifies multi-tenant organization handling.
* **Implementation:** Backend uses JIT (Just-In-Time) sync via `ClerkJWTMiddleware` to map `clerk_user_id` and `clerk_org_id` to local PostgreSQL UUIDs (`users` and `workspaces` tables) for foreign key relations.

### DEC-004: Dropped `workspace_members` in favor of Clerk Source of Truth
* **Context**: `workspace_members` was originally storing local associations for multi-tenancy.
* **Decision**: We completely removed `workspace_members` from the Postgres schema. Clerk serves as the unequivocal source of truth for organization memberships and roles. For the "Personal Workspaces" fallback (dev-only or B2C users), we added an `owner_id` directly to the `workspaces` table. 
* **Impact**: Simplifies schema, removes dual source of truth for membership, defers all RBAC and multitenant isolation securely to the Clerk JWT payload injected dynamically via middleware.


### DEC-005: Async Analytics Publisher
* **Context**: Analytics for short-link clicks must not block or crash the redirect itself.
* **Decision**: We designed the `AnalyticsPublisher` interface to be injected into the `RedirectHandler`. Events are fired in a background goroutine context. Raw IPs are hashed via `utils.HashIP()` at the handler level to avoid storing PII while maintaining unique tracking capabilities.
* **Impact**: Ensures O(1) high-speed redirects even if the analytics backend (Redis/ClickHouse) is down. Provides a clean contract for future Redis/ClickHouse implementations.

### DEC-006: Redis Stream Publisher Architecture
* **Context**: We need to durably buffer analytics events emitted by the `RedirectHandler` without causing unbounded goroutine growth or memory leaks in a high-traffic environment.
* **Decision**: We implemented `RedisAnalyticsPublisher` with a bounded in-memory Go channel (queue size: 5000) and a single long-lived background worker goroutine calling `XADD`. 
* **Impact**:
  - Drops events safely at `O(1)` time if Redis is permanently hung (preventing backend crash).
  - Handles `SIGTERM`/`SIGINT` via graceful context cancellation, forcefully draining the queue to Redis before process exit up to a 5-second hard limit.
  - Zero latency impact on the critical path of the public short-link redirect.

### DEC-007: Analytics Ingestion & Duplicate Strategy
* **Context**: We need to ingest click events from Redis Streams into ClickHouse. We must choose an ingestion engine, a duplicate event strategy, and a batching strategy.
* **Decision**: We implemented `RedisAnalyticsConsumer` reading via `XREADGROUP` using a consumer group (`analytics-clickhouse`). Events are batched up to 1000 items or 1s intervals. We insert into ClickHouse using the `MergeTree` engine partitioned by month, ordered by `(workspace_id, link_id, timestamp)`. 
* **Impact**:
  - `MergeTree` offers maximum insert throughput over `ReplacingMergeTree`.
  - Duplicate events (e.g., resulting from a crash before `XACK`) are permitted to exist in the underlying table. Query-time deduplication (`uniqExact(event_id)`) will be used to resolve exact analytical counts where required.
  - At-least-once delivery is structurally guaranteed because `XACK` strictly follows the successful ClickHouse batch response.
  - Events have an automatic TTL of 90 days built into the schema.

## DEC-004: Cache-Aside Redis Architecture for Redirects
* **Context**: Public link resolution queries Postgres for every redirect, causing a bottleneck under heavy load.
* **Decision**: Implement a Redis cache using the Cache-Aside pattern. Postgres remains the absolute source of truth.
* **Details**: 
  - **Key Format**: `link:redirect:<slug>`.
  - **Payload**: JSON serialization of `LinkRedirectTarget`.
  - **TTL**: 24 hours.
  - **Invalidation**: `LinkService` deletes the key on update/delete operations.
  - **Singleflight**: Implemented via `golang.org/x/sync/singleflight` to prevent cache stampedes on concurrent misses for the same slug.
  - **Negative Caching**: Explicitly rejected. Missing slugs are caught by `singleflight` for concurrency bursts and by Postgres indexes sequentially. No complex negative invalidation logic is needed.
* **Consequences**: Redis outage gracefully falls back to Postgres. Link resolution performance scales horizontally.
