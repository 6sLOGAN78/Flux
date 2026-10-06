# Architecture Research

**Domain:** Production-grade multi-tenant link attribution and marketing analytics SaaS
**Researched:** 2026-10-05
**Confidence:** HIGH for component boundaries and dependency direction; MEDIUM for the final event transport and analytical-store cutover because they require workload benchmarks and an explicit click-loss SLO

## Executive Recommendation

Evolve the current Go/Echo backend into a **modular system with four composition roots**: `api`, `redirector`, `worker`, and `migrate`. Keep one Go module and one PostgreSQL transactional model. Do not create a network service for each product domain. Independent deployment is warranted where traffic, availability, or lifecycle differs; that is true for control-plane HTTP, redirect traffic, asynchronous processing, and migrations.

The primary architectural seam is between the **control plane** and the **data plane**:

- The control plane owns authenticated workspace administration, link and domain configuration, developer APIs, and canonical transactional records. Its critical path is API → application service → PostgreSQL.
- The data plane owns public host/path resolution, redirect rules, click emission, event ingestion, analytics projections, and attribution processing. Its redirect critical path is Redis cache → PostgreSQL fallback → immediate 3xx. Analytics and webhook availability must never determine whether the redirect succeeds.

Use PostgreSQL as the source of truth for users, workspaces, links, domains, conversions, attribution decisions, API keys, webhook configuration, and outbox records. Use Redis only for derived redirect cache, rate limits, idempotency acceleration, and Asynq. Retain Asynq for finite background jobs such as email, domain verification, metadata fetches, exports, and webhook delivery. Introduce NATS JetStream when the immutable tracking pipeline needs replay and multiple independent consumers; do not force high-volume event-stream semantics through Asynq.

Start analytics in PostgreSQL only if it makes the first vertical slice materially simpler. Hide storage behind an analytics query/write boundary from the first implementation. Move raw click/event projections and pre-aggregates to ClickHouse when measured ingestion, retention, or dashboard-query load begins to contend with transactional work. Canonical attribution decisions should remain explainable in PostgreSQL even after ClickHouse becomes the dashboard store.

## Standard Architecture

### System Overview

```text
                         CONTROL PLANE
┌──────────────┐     ┌─────────────────────────────────────────────┐
│ Web / SDK /  │────▶│ Go API                                      │
│ Public API   │     │ auth → workspace policy → application use  │
└──────────────┘     │ cases → domain modules                     │
                     └──────────────┬───────────────┬──────────────┘
                                    │               │
                                    ▼               ▼
                              PostgreSQL          Redis
                         canonical state +      cache/rate/
                         idempotency + outbox   transient state
                                    │
                                    ▼
                              outbox relay
                                    │
                           cache invalidation /
                              domain events

                           REDIRECT DATA PLANE
┌──────────────┐     ┌─────────────────────────────────────────────┐
│ Visitor /    │────▶│ Go Redirector                               │
│ Preview Bot  │     │ normalize host/path → lookup → rules → 3xx │
└──────────────┘     └──────────┬──────────────┬───────────────────┘
                                │              │
                         Redis redirect    PostgreSQL fallback
                              cache         (read-only path)
                                │              │
                                └──────┬───────┘
                                       │ immutable click envelope
                                       ▼
                        spool-backed event publisher
                                       │
                                       ▼
                              NATS JetStream
                         durable, replayable stream
                         ┌─────────────┼─────────────┐
                         ▼             ▼             ▼
                  analytics       attribution    enrichment /
                   consumer        consumer       fan-out
                         │             │             │
                         ▼             ▼             ▼
                    ClickHouse     PostgreSQL       Asynq
                    projections    decisions      finite jobs
                                                        │
                                                 webhooks/email
```

The spool-backed publisher is the target for strict click durability without coupling redirect success to remote broker health. A plain in-memory channel followed by asynchronous publish is simpler but can lose accepted clicks during a process crash. This project must explicitly choose and test a loss SLO before claiming durable click capture. See **Critical unresolved trade-off: redirect latency versus zero-loss event capture** below.

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| API process | Authenticated control-plane and public developer endpoints; workspace authorization; transactional use cases | Existing Echo foundation, with process-specific middleware and explicit domain/application dependencies |
| Redirector process | Public host/path resolution, cache lookup, PostgreSQL fallback, routing rules, cookie/click identity, immediate redirect, click handoff | Separate small Echo or `net/http` composition root with no dashboard/auth/email dependencies |
| Worker process | Run Asynq jobs and JetStream consumers independently of HTTP replicas | One binary initially, configured by queue/consumer role; split only when resource profiles justify it |
| Migrator process | Apply PostgreSQL migrations exactly once as a release/deployment step | Reuse embedded tern migrations; never auto-migrate in every API replica |
| Workspace/identity module | Users, memberships, invitations, roles, workspace resolution and authorization policy | Domain/application package with tenant-scoped repository ports |
| Link/domain module | Canonical link configuration, domain ownership, collision rules, lifecycle, redirect snapshot creation | PostgreSQL transaction plus outbox record; cache representation is derived |
| Redirect lookup module | Resolve `(normalized_host, normalized_path)` to immutable redirect snapshot | Redis cache-aside with PostgreSQL repository fallback and bounded cache fill |
| Event publisher | Accept versioned immutable envelopes from redirector and server-event endpoints | NATS JetStream publisher; strict mode adds local durable spool and replay |
| Event consumers | Validate versions, deduplicate, enrich, write projections, invoke attribution | Durable pull consumers with explicit acknowledgements and idempotent writes |
| Conversion/attribution module | Normalize conversions, correlate identities/clicks, apply first/last-click models, persist explanation | Pure domain rules plus PostgreSQL repositories; emits projection events after commit |
| Analytics write side | Persist append-only fact events and incremental aggregates | PostgreSQL initially if proven adequate; ClickHouse `MergeTree`/aggregate targets later |
| Analytics query service | Validate tenant/filter scope, execute dashboard queries, return stable DTOs | Storage port with PostgreSQL and ClickHouse adapters; UI never queries stores directly |
| Outbox relay | Publish committed control-plane changes without database/broker dual-write gaps | PostgreSQL outbox table claimed with short transactions; idempotent publish and retry |
| Cache invalidator | Evict or replace redirect snapshots after link/domain commits | Outbox subscriber; API may also attempt post-commit eviction for lower staleness |
| Asynq job system | Retryable finite commands with operational state | Email, DNS verification, metadata scraping, exports, webhook deliveries, cleanup |
| Frontend | Dashboard interactions and visualization; no canonical business logic | Separate TypeScript app calling versioned Go API/contracts |

### Process Dependency Matrix

Readiness should be process-specific rather than inherited from the current all-in-one `server.Server`.

| Process | Required to become ready | Degraded but can serve | Must not initialize |
|---------|--------------------------|------------------------|---------------------|
| API | Configuration, PostgreSQL, auth verifier/keys required by active routes | Redis-backed acceleration and noncritical delivery integrations | Asynq consumers, redirect-only cache warmers |
| Redirector | Valid configuration, PostgreSQL fallback path, usable local event spool when strict durability is enabled | Redis cache miss/failure; remote event broker outage while spool has capacity | Email, Clerk workspace administration, analytics queries |
| Worker | Its selected broker/queue plus stores required by selected handlers | Unrelated consumer groups/integrations | HTTP listener unless exposing isolated health/metrics |
| Migrator | PostgreSQL and migration assets | None | HTTP, Redis, workers |

Liveness should indicate that the process can run and should not probe downstream systems. Readiness should reflect only the dependencies required for that process's contract. This removes the current ambiguous `/status` behavior.

## Recommended Project Structure

```text
apps/backend/
├── cmd/
│   ├── api/                  # control-plane composition root
│   ├── redirector/           # redirect data-plane composition root
│   ├── worker/               # Asynq + JetStream consumer composition root
│   └── migrate/              # explicit migration command
├── internal/
│   ├── identity/             # cohesive domain/application/adapter package
│   ├── workspace/
│   ├── link/
│   ├── domain/
│   ├── redirect/
│   ├── tracking/
│   ├── conversion/
│   ├── attribution/
│   ├── analytics/
│   ├── webhook/
│   ├── apikey/
│   ├── campaign/             # add only when its vertical slice begins
│   ├── platform/
│   │   ├── postgres/         # pool, transaction runner, migrations
│   │   ├── redis/            # clients and cache adapters
│   │   ├── asynq/            # finite job transport and registrations
│   │   ├── eventbus/         # JetStream adapter and envelope codec
│   │   ├── analyticsdb/      # ClickHouse adapter when introduced
│   │   ├── observability/
│   │   └── email/
│   ├── transport/
│   │   └── http/             # Echo middleware, envelopes, system routes
│   └── testkit/              # containers, fixtures, tenant assertions
├── migrations/
│   ├── postgres/
│   └── clickhouse/           # only after ClickHouse adoption
├── static/                   # embedded/generated API docs
└── templates/                # preferably embedded at build time

packages/
├── zod/                      # shared TypeScript DTO schemas
├── openapi/                  # generated/verified public API contract
├── emails/
└── tracker/                  # browser SDK when conversion phase begins

apps/frontend/                # dashboard, introduced as its own app
```

This is a direction, not a mandated big-bang move. The existing `internal/config`, `database`, `handler`, `repository`, `service`, and `lib/job` packages can stay until touched. New product slices should establish the domain-oriented structure; composition roots can temporarily adapt old registries. Move shared infrastructure only when a second process needs it or when lifecycle repair touches it.

### Structure Rationale

- **`cmd/*`:** Each process constructs only the clients it needs and owns their startup/drain order. This removes the present coupling where every HTTP replica starts email workers.
- **Domain packages:** Keep model, rules, use cases, narrow repository interfaces, and adapters close enough to see cohesion. Avoid growing global `service.Services` and `repository.Repositories` registries into cross-domain grab bags.
- **`platform/*`:** Contains vendor-specific adapters. Business packages should not import Echo, pgx, Redis, Asynq, NATS, ClickHouse, Clerk, or Resend.
- **`transport/http`:** Owns binding, authentication adapters, stable API errors, request IDs, and route wiring; handlers call use cases and do not contain attribution or tenancy rules.
- **Separate migrations:** PostgreSQL and ClickHouse schemas have different lifecycle and rollback semantics. Do not make ClickHouse availability a condition for the transactional database migration command.

## Dependency Direction

```text
cmd/api ─────────┐
cmd/redirector ──┼──▶ transport/adapters ──▶ application use cases ──▶ domain
cmd/worker ──────┘             │                       │                 ▲
                               ▼                       ▼                 │
                         vendor clients         repository/event ports ─┘
                               ▲                       │
                               └──── infrastructure implementations ────┘
```

Rules:

1. Domain types and rules import only the standard library and other same-domain value types.
2. Application use cases define transaction, repository, clock, ID, cache, and event ports at the point of use.
3. PostgreSQL/Redis/NATS/ClickHouse adapters implement those ports; they do not expose vendor types above the adapter boundary.
4. HTTP and message handlers translate input/output and errors. They do not decide tenant permissions, attribution credit, or commission amounts.
5. Cross-domain writes are coordinated by an application use case and one visible PostgreSQL transaction when they share the same consistency boundary.
6. Cross-process effects leave the transaction through an outbox or an explicit event-ingestion boundary.

Avoid interface-per-struct ceremony. Define small interfaces only where a use case needs substitution or an infrastructure boundary exists.

## Architectural Patterns

### Pattern 1: Traffic-Plane Separation With Shared Domain Code

**What:** API and redirector are separate deployables, but both import shared link/domain logic from the same Go module. Worker consumers import application/domain packages without calling the API over HTTP.

**When to use:** Immediately. Redirect traffic has a distinct latency, attack, scaling, and availability profile.

**Trade-offs:** Independent deployment and failure isolation require separate configuration, health checks, dashboards, and lifecycle tests. Shared code avoids premature network boundaries but demands disciplined package ownership.

### Pattern 2: Tenant Scope as a Required Input

**What:** Resolve the authenticated actor and workspace at the transport boundary, authorize the requested action in the application layer, and require `WorkspaceID` in every workspace-owned repository method. Put `workspace_id` on each tenant-owned table or make the ownership join explicit. Enforce collision constraints with tenant/domain columns in PostgreSQL.

**When to use:** From the first identity/workspace migration onward.

**Trade-offs:** Repetitive tenant predicates are deliberate safety. PostgreSQL row-level security can add defense in depth, but it should not replace explicit authorization. PostgreSQL documents that table owners and `BYPASSRLS` roles bypass policies; if RLS is adopted, run with a non-owner application role, consider `FORCE ROW LEVEL SECURITY`, set tenant context transaction-locally, and test every table and command path.

**Repository shape:**

```go
type LinkRepository interface {
    GetByID(ctx context.Context, workspaceID, linkID LinkID) (Link, error)
    List(ctx context.Context, workspaceID WorkspaceID, q LinkQuery) (Page[Link], error)
    Save(ctx context.Context, workspaceID WorkspaceID, link Link) error
}
```

Do not expose unscoped `GetByID(id)` methods for tenant-owned records.

### Pattern 3: Transactional Outbox for Control-Plane Side Effects

**What:** The same PostgreSQL transaction writes canonical state and a versioned outbox event. A relay publishes committed events for cache invalidation, webhooks, analytical projections, and integration handlers. Consumers use the envelope ID as an idempotency key.

**When to use:** Link/domain changes, conversion acceptance, attribution decisions, partner commissions, and any transaction that must trigger an external effect.

**Trade-offs:** Adds an outbox table, relay, retention, and lag monitoring. It removes the more serious database/broker dual-write gap. AWS's official pattern guidance explicitly calls out duplicate delivery and therefore still requires idempotent consumers.

```text
BEGIN
  update link version
  insert outbox(id, type, version, workspace_id, aggregate_id, payload)
COMMIT
       ↓
relay claims row → publish → mark published
```

### Pattern 4: Cache-Aside Redirect Snapshots With Versioned Invalidation

**What:** Cache the complete, immutable redirect decision input under `redirect:{normalized_host}:{normalized_path}`. A miss loads from PostgreSQL and fills Redis with a bounded TTL. A committed link/domain change emits an outbox event that evicts or replaces the entry. The redirector treats Redis as optional acceleration, never authority.

**When to use:** Core link and domain phases.

**Trade-offs:** Eventual invalidation creates a bounded stale-read window. A short negative-cache TTL protects PostgreSQL from random-key scans but can temporarily hide a just-created link; use different TTLs for positive and negative entries. Prevent thundering herds with request coalescing or a short per-key fill lock only after measurement.

Cache payloads should carry a schema version, link version, destination, status, expiration, redirect code, and compiled routing-rule inputs. They should not contain secrets such as password verifiers unless the redirect path explicitly needs a safe representation.

### Pattern 5: Versioned Event Envelope and Idempotent Consumers

**What:** Every published fact has a stable event ID, type, schema version, occurrence time, producer, workspace ID, aggregate identifiers, request/correlation IDs, and payload. Consumer acknowledgements occur only after their durable write commits.

```json
{
  "id": "evt_...",
  "type": "link.clicked",
  "version": 1,
  "occurred_at": "...",
  "producer": "redirector",
  "workspace_id": "ws_...",
  "request_id": "req_...",
  "traceparent": "...",
  "data": {}
}
```

**When to use:** Tracking pipeline and all later cross-process events.

**Trade-offs:** Schema evolution becomes a maintained contract. NATS JetStream consumers provide at-least-once delivery and redeliver unacknowledged messages, so storage must deduplicate by event ID. JetStream publish deduplication and double acknowledgements reduce duplicates but do not eliminate the need for idempotent business effects.

### Pattern 6: Separate Event Streams From Finite Jobs

**What:** Use JetStream for immutable facts that need fan-out, replay, independent consumer positions, and retention. Use Asynq for commands with retry schedules and operational completion, such as `verify-domain`, `deliver-webhook`, `render-export`, and `send-email`.

**When to use:** Introduce JetStream with click tracking. Keep the existing Asynq investment for appropriate jobs.

**Trade-offs:** Operating two transports costs more than using Redis alone. The semantic clarity is valuable once analytics, attribution, enrichment, and webhook production all consume click/conversion facts. Asynq officially guarantees at least one execution and its uniqueness lock is best effort; handlers must therefore be idempotent too.

### Pattern 7: Canonical Facts and Rebuildable Analytics Projections

**What:** Keep transactional ownership and explainable attribution outcomes in PostgreSQL. Write immutable analytics facts and aggregate projections to the analytical store. The dashboard calls an analytics query service, not ClickHouse directly. Projection code is replay-safe so aggregates can be rebuilt from retained events.

**When to use:** PostgreSQL implementation from day one; ClickHouse adapter when the analytics phase or measured load justifies it.

**Trade-offs:** Eventual consistency is visible in dashboard freshness. ClickHouse incremental materialized views shift aggregation to insert time and are well suited to near-real-time rollups, but their primary/order key is a sort/index choice rather than a uniqueness constraint; event-ID deduplication needs explicit design.

Recommended ClickHouse shape after adoption:

- Append-only raw fact table with event ID, event time, workspace, link/campaign/domain dimensions, visitor/session identifiers, derived geo/device fields, event type, and integer minor-unit revenue.
- `MergeTree` ordering led by common tenant/time filters, validated with actual query patterns; avoid high-cardinality partition keys.
- Time partitioning chosen for retention/drop operations, not as the main query accelerator.
- Incremental materialized views into `SummingMergeTree` or `AggregatingMergeTree` targets for common hourly/daily dashboards.
- Source event retained long enough to replay/backfill projection changes.

### Pattern 8: Pure Attribution Core, Stateful Application Shell

**What:** Model first-click and last-click choice as deterministic functions over eligible touchpoints, conversion time, attribution window, and policy. The application layer loads candidates, invokes the model, persists the credited click/link/campaign plus model/window/reason, then emits an `attribution.decided` event.

**When to use:** Conversion attribution phase.

**Trade-offs:** Reprocessing must define whether historical decisions are immutable, superseded, or recalculated. Store model version and policy inputs so the outcome is explainable.

## Data Flow

### Control-Plane Link Create/Update

```text
Web/API client
  → API authentication
  → workspace membership + capability check
  → transport validation (URL syntax, DTO shape)
  → link application service
  → domain validation (destination policy, key/domain collision rules)
  → PostgreSQL transaction
       ├─ link/domain rows
       ├─ audit record when security-relevant
       └─ outbox: link.changed.v1
  → response from committed canonical state

Outbox relay
  → publish link.changed.v1
  → cache invalidator evicts redirect:{host}:{path}
  → optional webhook-production handler creates Asynq delivery jobs
```

Use a database unique constraint for the normalized host/domain plus normalized short key. Application prechecks improve errors but cannot guarantee collision safety under concurrency.

### Redirect Request

```text
GET https://go.example/abc
  → trusted-proxy/IP extraction + request ID
  → normalize host and path once
  → cheap abuse/rate checks
  → Redis GET redirect:{host}:{path}
       ├─ hit: decode versioned snapshot
       └─ miss/error: tenant-independent PostgreSQL lookup by host+path
                    → bounded cache fill
  → evaluate status/expiration/password/geo/device/fallback rules
  → generate click ID and privacy-conscious visitor/session state
  → enqueue immutable click envelope to local durable publisher
  → issue configured 301/302/307

Publisher loop
  → JetStream publish acknowledgement
  → remove from spool only after broker acceptance
```

The redirector must not call analytics, attribution, webhook, email, or frontend services. A Redis outage should reduce cache hit rate and increase PostgreSQL load, not break valid redirects. Protect the fallback with tight pool budgets, query timeouts, negative caching, load shedding for obvious abuse, and an alert on cache hit ratio.

### Click Ingestion and Analytics

```text
JetStream link.clicked.v1
  ├─ analytics consumer
  │    → dedupe event_id
  │    → append raw fact
  │    → incremental aggregates
  │    → ACK after durable insert
  ├─ identity/enrichment consumer
  │    → derive geo/device/referrer fields
  │    → publish enriched fact or update projection
  └─ webhook-event producer
       → filter subscribed endpoints
       → create signed-delivery Asynq jobs
```

Prefer enriching from a stable original event into a new version/type rather than mutating the broker message. Define whether dashboards include provisional un-enriched rows to prevent double counting when enrichment completes.

### Server Conversion Event and Attribution

```text
POST /v1/events + API key + Idempotency-Key
  → authenticate key and workspace scope
  → validate/version payload and money/currency
  → PostgreSQL transaction
       ├─ reserve idempotency key + request hash
       ├─ persist canonical conversion event
       └─ outbox: conversion.accepted.v1
  → stable accepted/replayed response

Attribution consumer
  → dedupe envelope ID
  → resolve click/visitor/customer candidates within window
  → pure first-click/last-click decision
  → PostgreSQL transaction
       ├─ persist model version, credited IDs, reason, amount/currency
       └─ outbox: attribution.decided.v1
  → analytics projection and webhook jobs
```

Reject reuse of the same idempotency key with a different request hash. Do not use Redis alone for durable conversion idempotency.

### State Ownership

| State | Authority | Derived Copies |
|-------|-----------|----------------|
| Workspace membership and roles | PostgreSQL | Request-local authorization context |
| Link/domain configuration | PostgreSQL | Redis redirect snapshots |
| API keys and scopes | PostgreSQL, secret hash only | Short-lived verification cache if justified |
| Click/conversion event stream | Retained JetStream plus durable sink according to retention design | PostgreSQL/ClickHouse facts and aggregates |
| Conversion and attribution decision | PostgreSQL | ClickHouse analytical projection |
| Redirect cache | None; derived | Redis only |
| Webhook configuration and delivery history | PostgreSQL | Asynq pending task state |
| Dashboard aggregates | Rebuildable projection | PostgreSQL initially, ClickHouse later |

## Critical Unresolved Trade-off: Redirect Latency Versus Zero-Loss Event Capture

“Return the redirect without waiting for analytics” and “every accepted click is durable through process or broker failure” cannot both be guaranteed by a plain in-memory asynchronous queue. The event transport alone does not resolve the gap between committing the HTTP response and durably accepting the event.

Recommended production target:

1. Redirect handler creates the event before writing the response.
2. It appends the envelope to a bounded local durable spool owned by the redirector process.
3. The response depends only on local spool acceptance, never remote broker or analytical storage.
4. A background publisher sends batches to JetStream and removes local entries only after a publish acknowledgement.
5. Readiness fails before the spool reaches capacity; existing in-flight requests retain a defined fallback policy.

This requires durable local storage and careful multi-replica ownership/recovery. If the deployment platform cannot provide that, choose and document a bounded-loss policy (for example, in-memory buffering plus a broker publish timeout) rather than describing the result as durable. This decision needs a phase-specific spike and failure-injection benchmark before the tracking milestone is planned.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| Product validation / modest event volume | Separate API and redirector processes; PostgreSQL canonical store; Redis redirect cache; Asynq worker; analytics storage may remain partitioned PostgreSQL behind a port; one JetStream deployment only when durable tracking begins |
| Sustained growth / dashboard contention | Add ClickHouse raw facts and materialized aggregates; scale redirectors independently; use durable pull consumers with bounded batches; add read replicas only for measured PostgreSQL read pressure; tune cache TTL/hit ratio and separate worker roles by queue |
| Very high event and retention volume | Replicated JetStream and ClickHouse; shard/replicate ClickHouse only from measured query/ingest needs; dedicated ingestion workers; tiered retention/object-storage export; regional redirect strategy; load-test failover and replay before expanding topology |

### Scaling Priorities

1. **Protect PostgreSQL from redirect misses:** Track hit ratio, fallback QPS, pool saturation, and random-key abuse. Cache correctness and miss behavior matter before sharding.
2. **Batch analytical writes:** Per-event inserts waste throughput in both PostgreSQL and ClickHouse. Consumers should batch within bounded latency and acknowledge only after the batch's durable result.
3. **Pre-aggregate proven queries:** Add ClickHouse incremental materialized views for stable dashboard dimensions. Keep raw facts for new filters and backfills.
4. **Partition lifecycle before distributing services:** Retention, TTLs, and deletion/export workflows usually become painful before service discovery does.
5. **Split workers by resource profile:** Analytics writers, SSRF-sensitive metadata fetchers, and webhook delivery have different network/CPU/failure behavior; configure separate worker modes before creating separate repositories or services.

Do not attach roadmap phases to arbitrary user counts. Trigger changes from measured event rate, retained rows/bytes, p95/p99 latency, broker lag, replay time, database contention, and recovery objectives.

## Pragmatic Evolution From the Current Repository

### Stage 0: Stabilize the Existing Foundation

- Keep Echo; changing routers has no product value and would discard working middleware.
- Fix request-local allocation in generic handlers before any domain route uses them.
- Make configuration return errors rather than terminate from libraries.
- Separate `/live` and process-specific `/ready`; fix Redis dependency semantics.
- Apply pool settings, close every client, handle SIGTERM, drain producers/consumers before stores, and aggregate shutdown errors.
- Embed or explicitly configure docs/email assets so binaries do not depend on working directory.
- Add executable lifecycle, migration, handler concurrency, and integration tests.
- Stop running migrations implicitly from every non-local API process; establish `cmd/migrate`.

### Stage 1: Create Composition Boundaries Without Moving Everything

- Rename or wrap `cmd/flux` as `cmd/api` while preserving existing packages.
- Extract worker startup from `server.New`; create `cmd/worker` using existing Asynq handlers.
- Replace `*server.Server` injection into application code with the narrow dependency set each constructor needs.
- Keep logger/config/database factories shared, but make resource ownership explicit per command.
- Establish liveness/readiness/metrics endpoints per process.

### Stage 2: Build the Tenant Kernel and First Vertical Link Slice

- Add workspace/user/membership/link/domain-key migrations with explicit ownership, foreign keys, indexes, and unique constraints.
- Put authorization in application services and require workspace scope in repositories.
- Implement link create/read/update plus outbox in PostgreSQL.
- Add isolation tests that create two workspaces and attempt every CRUD path across tenants.
- Add a cache invalidation consumer before relying on redirect cache TTL alone.

### Stage 3: Extract the Redirector

- Create `cmd/redirector` with a minimal router/middleware set.
- Implement normalized host/path lookup, Redis cache-aside, PostgreSQL fallback, link state/rule evaluation, and redirect-only observability.
- Load-test cached and uncached paths against the project latency targets.
- Test Redis outage, cold cache, stale invalidation, PostgreSQL timeout, and shutdown during traffic.
- Deliver a functioning redirect before adding click analytics, keeping the latency baseline visible.

### Stage 4: Add the Durable Tracking Pipeline

- Resolve the click-loss SLO and complete the spool/broker failure spike.
- Introduce the versioned envelope and JetStream stream/consumers.
- Preserve Asynq for finite tasks.
- Add event-ID dedupe, poison-event handling, retry/backoff, lag metrics, replay tests, and privacy-conscious enrichment.
- Initially persist events to PostgreSQL if benchmarks meet retention/query needs; otherwise introduce ClickHouse here.

### Stage 5: Add Analytics, Then Conversion Attribution

- Stabilize analytics query DTOs and filters before tuning projections.
- Add ClickHouse raw facts and incremental aggregates when the phase evidence supports it.
- Implement durable server-event idempotency and canonical conversions in PostgreSQL.
- Implement attribution as a pure domain module with persisted model/version/explanation.
- Project attribution and revenue into analytics after the canonical transaction commits.

This order proves the user-visible loop while preserving clear rollback points. Identity/tenancy must precede tenant data; link authority must precede redirect cache; redirect correctness and latency must precede tracking; durable facts must precede trustworthy analytics; click identity must precede attribution.

## Anti-Patterns

### One Resource Container for Every Process

**What people do:** Continue passing `*server.Server` everywhere and initialize PostgreSQL, Redis, jobs, email, and HTTP together.

**Why it is wrong:** API replicas accidentally consume jobs; redirector health becomes coupled to email; lifecycle failures leak resources; tests cannot construct minimal systems.

**Do this instead:** Process-specific composition roots and narrow constructor dependencies.

### Repository/Service Registries as Service Locators

**What people do:** Add every future repository and service to global structs and let handlers reach across domains.

**Why it is wrong:** Dependencies become implicit, tenant scope is easy to omit, and domain boundaries exist only as folder names.

**Do this instead:** Construct cohesive feature modules and pass explicit use-case interfaces to route/message adapters.

### Synchronous Analytics on the Redirect Path

**What people do:** Insert a click row, call ClickHouse, enrich geo/device data, or produce webhooks before returning 3xx.

**Why it is wrong:** Analytical or third-party failure becomes redirect failure and destroys the main availability boundary.

**Do this instead:** Create a minimal event and hand it to the independent durable publisher; enrich and fan out asynchronously.

### Database Update Followed by Best-Effort Publish

**What people do:** Commit a link or conversion, then publish/invalidate in a separate call with no durable record.

**Why it is wrong:** A crash between operations creates permanently stale cache or missing downstream facts.

**Do this instead:** Transactional outbox plus idempotent relay/consumers.

### Treating Broker “Exactly Once” as Business Exactly Once

**What people do:** Assume publish deduplication or task uniqueness makes side effects safe.

**Why it is wrong:** Redelivery, lock expiry, partial batch success, consumer restarts, and downstream commits can still duplicate effects.

**Do this instead:** Idempotency keys and unique constraints at every side-effect sink; acknowledge only after commit.

### Redis as Link Authority

**What people do:** Accept link updates into Redis first or serve indefinitely from unversioned cache entries.

**Why it is wrong:** Cache loss or stale values alter customer routing without a canonical audit trail.

**Do this instead:** PostgreSQL authority, versioned cache snapshots, outbox invalidation, and bounded TTL.

### Tenant Isolation Only in HTTP Middleware

**What people do:** Check workspace membership once, then call unscoped repository methods.

**Why it is wrong:** Background workers, API keys, new handlers, and programming mistakes can bypass the middleware assumption.

**Do this instead:** Authorization in application use cases, required workspace IDs in repositories, database constraints, isolation tests, and carefully configured RLS as defense in depth.

### Premature ClickHouse/Kafka/Microservice Topology

**What people do:** Deploy every long-term component before a working link/redirect flow or measured load.

**Why it is wrong:** Operational surface expands faster than product evidence, while transactional correctness remains untested.

**Do this instead:** Preserve ports and event schemas, introduce JetStream with durable tracking, and introduce ClickHouse from measured analytical need or the analytics milestone's verified benchmark.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| Identity provider | Adapter maps verified identity to internal user; internal workspace membership remains authoritative | Avoid provider organization claims as the sole tenant policy |
| DNS/provider TLS | Asynq verification jobs and infrastructure-managed certificates | Store state transitions in PostgreSQL; never build certificate authority logic into domains |
| GeoIP/UA enrichment | Worker-local databases/services behind pure enrichment ports | Keep redirect path minimal; record source/version for reproducibility |
| Email provider | Asynq command with idempotency key and delivery result | Embed templates or use explicit asset path |
| Webhook destinations | Signed Asynq delivery task with timeout/backoff and persisted attempt | SSRF defenses, HMAC versioning, manual replay, endpoint disable policy |
| Stripe and later integrations | Integration adapter consumes/emits canonical domain events | Stripe payloads must not define attribution domain models |
| OpenTelemetry backend | OTLP via process instrumentation/collector | Propagate W3C trace context in event envelopes; never put secrets/PII in baggage |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| Frontend ↔ API | Versioned HTTPS JSON/OpenAPI | Frontend has no direct database access or canonical business rules |
| API ↔ domain/application | In-process calls | Strong consistency boundary; no HTTP between modules |
| API transaction ↔ cache/events | PostgreSQL outbox | Optional fast post-commit eviction does not replace durable relay |
| Redirector ↔ Redis/PostgreSQL | Cache-aside direct adapters | Separate pool budgets and query timeouts from API |
| Redirector ↔ event pipeline | Local spool then JetStream | Failure policy and capacity must be explicit |
| JetStream ↔ consumers | Versioned events, durable pull consumers, explicit ack | Independent consumer names for analytics, attribution, enrichment |
| Consumer ↔ Asynq | Consumer creates finite delivery/task command | Do not copy raw stream semantics into job queues |
| Analytics service ↔ store | Query/write ports | Allows PostgreSQL-to-ClickHouse migration without API contract change |

## Observability and Operations Boundary

Every process should have a distinct OpenTelemetry service name and propagate request ID, trace context, event ID, workspace ID (where safe), and message delivery count. OpenTelemetry's official guidance supports context injection/extraction through messaging carriers, allowing HTTP → event → consumer traces to remain causally linked.

Minimum service indicators:

- **Redirector:** request rate, cached/uncached/error latency, redirect status, cache hit ratio, PostgreSQL fallback QPS/errors, spool depth/age/capacity, broker publish acknowledgement latency, event drops.
- **API:** route latency/errors, auth/authorization denials, pool saturation, transaction retries, idempotency conflicts, outbox oldest age.
- **Workers:** consumer lag, delivery attempts, ack latency, poison messages, batch size/latency, sink errors, Asynq retry/archive counts.
- **Analytics:** ingestion delay, raw/aggregate row counts, query latency by shape, materialized-view lag, replay duration.

High-cardinality customer identifiers should not become metric labels. Put request/event IDs in traces and structured logs. Redact destination query secrets and avoid raw IP retention in logs/events.

## Build-Order Implications for the Roadmap

1. **Foundation/lifecycle first:** Later process separation depends on clean construction, shutdown, readiness, migrations, tests, and telemetry.
2. **Tenant kernel before product tables:** Retrofitting workspace ownership after links/events exist is a high-risk rewrite.
3. **Canonical link model before cache:** Redis representation must derive from tested PostgreSQL rules and constraints.
4. **Redirector before tracking:** Establish correctness and latency baseline without the event pipeline, then measure pipeline overhead separately.
5. **Tracking durability before analytics UI:** A polished dashboard over lossy or duplicate data erodes trust.
6. **Analytics query boundary before ClickHouse:** Stable query semantics make storage migration tractable.
7. **Conversions before attribution projections:** Persist accepted events/idempotency first, then make credit decisions explainable and replayable.
8. **Developer webhooks after domain events stabilize:** Webhook contracts become public compatibility commitments.

### Phase Research Flags

| Phase Topic | Research Needed | Reason |
|-------------|-----------------|--------|
| Redirect event durability | **Required spike** | Must reconcile p95 latency, broker outage, local persistence, deployment disk model, and acceptable click loss |
| PostgreSQL RLS | **Required design/test if adopted** | Runtime role ownership, `FORCE RLS`, pooled connection context, migrations, and worker access can silently bypass policy |
| ClickHouse schema/cutover | **Benchmark in analytics phase** | Ordering, partitioning, batching, retention, and aggregates depend on real query/event distributions |
| Geo/device identity | **Privacy/security review** | Retention, consent, raw IP lifecycle, cookie behavior, and bot classification affect schema and compliance |
| Redirect cache | Standard pattern plus load test | TTL, negative-cache behavior, invalidation lag, and fallback pool need workload validation |
| Asynq finite jobs | Standard pattern | Existing dependency is suitable once handlers are idempotent and the worker is separate |

## Confidence Assessment

| Area | Confidence | Basis |
|------|------------|-------|
| Control-plane/data-plane separation | HIGH | Product SLOs, current coupling evidence, and standard independent-scaling/lifecycle reasoning agree |
| Four Go composition roots in one module | HIGH | Fits current Go scaffold and avoids premature distributed boundaries |
| PostgreSQL authority + Redis derived cache | HIGH | Explicit project constraint and transactional correctness requirements |
| Outbox and idempotent consumers | HIGH | Official outbox guidance and at-least-once broker/job semantics |
| Asynq jobs vs JetStream facts | HIGH | Official capability/guarantee differences align with required replay/fan-out |
| ClickHouse target architecture | MEDIUM-HIGH | Official materialized-view guidance; exact schema needs workload evidence |
| Strict redirect click durability mechanism | MEDIUM | Architectural gap is clear; final choice depends on platform disk and loss/latency SLO |
| PostgreSQL RLS use | MEDIUM | Valuable defense in depth, but safe pooled-session operation requires a project-specific design |

## Sources

Primary and official sources consulted:

- [PostgreSQL 18: Row Security Policies](https://www.postgresql.org/docs/current/ddl-rowsecurity.html) — policy behavior, default deny, owner and `BYPASSRLS` exceptions.
- [AWS Prescriptive Guidance: Transactional Outbox Pattern](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html) — database/message dual-write failure and idempotent-consumer requirement.
- [NATS JetStream Deep Dive](https://docs.nats.io/learn/jetstream/) — stored/replayable messages, publish acknowledgements, consumer acknowledgements/double acknowledgements, durable consumption, and scaling concepts.
- [NATS JetStream Consumers](https://docs.nats.io/nats-concepts/jetstream/consumers) — at-least-once delivery, explicit acknowledgement, and redelivery behavior.
- [Asynq official repository](https://github.com/hibiken/asynq) — at-least-one execution, retries, crash recovery, priorities, and operational tooling.
- [Asynq: Unique Tasks](https://github.com/hibiken/asynq/wiki/Unique-Tasks) — TTL-based best-effort uniqueness and duplicate-task behavior.
- [ClickHouse: Incremental Materialized Views](https://clickhouse.com/docs/concepts/features/materialized-views/incremental-materialized-view) — insert-time transformations and aggregate targets.
- [ClickHouse: Choosing a Primary Key](https://clickhouse.com/docs/concepts/best-practices/choosing-a-primary-key) — ordering-key/query-pattern relationship for MergeTree tables.
- [OpenTelemetry: Context Propagation](https://opentelemetry.io/docs/concepts/context-propagation/) — W3C trace-context propagation across process boundaries.
- Repository evidence: `.planning/PROJECT.md`, `spec.md`, `.planning/codebase/ARCHITECTURE.md`, `.planning/codebase/STRUCTURE.md`, `.planning/codebase/CONCERNS.md`, and inspected Go composition/resource files.

---
*Architecture research for: Flux link attribution and marketing analytics platform*
*Researched: 2026-10-05*
