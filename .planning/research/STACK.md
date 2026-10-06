# Stack Research

**Domain:** Production-grade multi-tenant link attribution and marketing analytics SaaS
**Researched:** 2026-10-05
**Confidence:** HIGH for the foundation stack; MEDIUM for adoption thresholds for NATS and ClickHouse

## Recommendation in One Sentence

Keep the Go/Echo/pgx/PostgreSQL/Redis/Asynq foundation, upgrade unsupported or security-sensitive pins during the foundation phase, make OpenTelemetry and generated database/API boundaries first-class, then add NATS JetStream for durable event fan-out and ClickHouse for analytical scale only in the phases that need them.

## Recommended Stack

### Core Technologies

| Technology | Version / Selection Policy | Purpose | Why Recommended | Adopt |
|------------|----------------------------|---------|-----------------|-------|
| Go | **1.27.1**; pin the toolchain and patch automatically | API, redirector, workers, migrations | The repository is already Go-first. Go 1.25 is outside Go's current two-release support window as of this research date; 1.27.1 is the current stable patch. One module can produce separately deployable `api`, `redirector`, `worker`, and `migrate` binaries without creating microservices. | Foundation, immediately |
| Echo | **v4.15.4** | HTTP routing and middleware | Preserve the implemented Echo v4 adapters. Upgrade from v4.15.2 because v4.15.3 fixed a static-route encoded-separator bypass; v4.15.4 is the maintained v4 patch. Defer Echo v5 until there is a measured benefit and a dedicated migration plan. | Foundation, immediately |
| PostgreSQL | **18.6**, or the provider's PostgreSQL 18 current minor | Transactional source of truth and initial event store | PostgreSQL 18 is supported through November 2030. It supplies constraints, transactions, JSONB, partitioning, full-text search, and optional row-level security, covering links, tenancy, conversions, idempotency, and early analytics without another datastore. Always apply current minor updates. | Foundation |
| pgx | **v5.11.0** | PostgreSQL driver and pool | Retain direct SQL and pgxpool. The current v5 line supports PostgreSQL 18 protocol/authentication capabilities and avoids ORM opacity on tenant predicates, locking, bulk inserts, and query plans. Upgrade from v5.9.2 while repositories are still empty. | Foundation |
| Redis Open Source / managed Redis | **8.6.x current patch** (8.6.7 at research time), or a provider-supported 7.4+ patch after compatibility tests | Redirect cache, rate-limit coordination, idempotency cache, Asynq storage | Redis is appropriate for disposable, bounded-lifetime state. The 8.6 branch is supported and has multiple security fixes; pin a patch rather than `latest`. Use separate instances or at least separate operational failure domains for redirect cache and job queues before public scale. | Foundation; split failure domains before tracking launch |
| go-redis | **v9.22.0** | Redis client | Current v9 retains the existing API family and receives pool, stream, observability, and security maintenance. Upgrade behind redirect-cache integration and failure tests. | Foundation |
| Asynq | **v0.26.0** | Retryable background jobs | Keep it for email, webhook delivery, metadata fetches, and scheduled/retryable commands. It already exists and provides retries, timeouts, queues, and inspection. It is not the analytics event bus. | Foundation for jobs |
| NATS Server with JetStream | **v2.14.6** stable; never an RC | Durable click/conversion event transport and fan-out | JetStream provides persisted streams, consumers, acknowledgments, replication, replay, and independent consumers for analytics, attribution, enrichment, and webhooks. This is where the product's event topology becomes materially different from a task queue. | Tracking phase, after redirect correctness |
| nats.go | **v1.54.0** | Go JetStream client | Current stable client; minimum Go 1.26 aligns with the recommended Go 1.27 toolchain. | With NATS |
| ClickHouse | **26.8 LTS**, current patch (26.8.16.41 at research time), preferably managed | High-volume immutable event analytics and aggregates | The LTS channel is the stable operational choice when PostgreSQL event queries or ingestion begin competing with OLTP. Use it for click/event facts and analytical projections, never for link configuration or authorization. | Analytics phase only when load evidence triggers it |
| clickhouse-go | **v2.48.0** | Native batched ClickHouse ingestion and queries | Official Go client with native protocol and batch inserts. Keep behind an analytics-store interface so early PostgreSQL storage remains replaceable. | With ClickHouse |
| Next.js | **16.3.8 Active LTS** | Dashboard and marketing/application web UI | The frontend is greenfield and the product spec already selects React/Next.js. 16.3.8 is the current patched Active LTS release; the September 2026 security advisory makes exact patched versions important. Keep business rules and authoritative writes in the Go API. | First dashboard slice |
| React | **19.2.7** with Next.js 16.3.x; advance only after Next certifies the newer line | UI runtime | Matches the mature Next 16 generation and avoids independently moving React Server Components ahead of the framework's tested matrix. | With frontend |
| Node.js | **24 LTS** for CI/build/runtime | Next.js runtime and workspace tooling | Node's own guidance says production applications should use Active or Maintenance LTS. Node 26 is still Current on the research date; use 24 LTS until 26 is formally LTS. | Foundation CI and frontend |
| Bun | **1.3.14** stable baseline; upgrade only through a lockfile-tested change | Workspace package manager and scripts | Preserve the declared Bun workflow but replace the stale 1.2.13 pin. Avoid the 1.4 canary/rewrite transition until a stable release passes the workspace build and Next tests. Remove stray npm lockfiles after confirming Bun is authoritative. | Foundation |
| Tailwind CSS | **4.3.x current patch** | UI styling and design tokens | Current stable major with direct Next/Webpack integration and a small runtime footprint. Build accessible components over tokens; do not make utility strings the design system. | First dashboard slice |

### Supporting Libraries

| Library / Capability | Version / Selection Policy | Purpose | When to Use |
|----------------------|----------------------------|---------|-------------|
| sqlc | **v1.31.1** | Generate typed pgx query code from reviewed SQL | Introduce with the first product schema. It preserves SQL visibility while preventing repetitive scan/argument code and catches schema/query drift in CI. Keep handwritten repositories only where dynamic queries genuinely require them. |
| Tern | **v2.4.1** (retain) | Embedded PostgreSQL migrations | Keep the current migration runner. Add forward/backward migration tests and one migration-owner process; do not run concurrent app-start migrations. |
| oapi-codegen | **v2.7.2** plus runtime **v1.2+** | Generate Echo boundary types/interfaces from the generated OpenAPI document | Use the existing Zod/ts-rest package to emit one checked OpenAPI artifact, then generate Go boundary types and compile-test them. This turns the current documentation-only contract into an enforced boundary without replacing Echo. |
| OpenTelemetry Go | API/SDK **v1.47.0**; contrib instrumentation **v0.71.0** | Vendor-neutral traces, metrics, context propagation, OTLP export | Introduce during foundation. Instrument Echo/net/http, pgx, Redis, NATS, and workers, plus domain spans for redirect lookup and attribution. Align the whole OTel family in one update to avoid semantic-convention skew. |
| OpenTelemetry Collector | Current stable release, pinned by image digest | Batch, retry, scrub, and route telemetry | Run locally and in production between services and the observability vendor. Export to New Relic initially if desired; phase out parallel proprietary instrumentation to avoid duplicate spans and inconsistent sampling. |
| Google UUID | **v1.6.0** (retain) | UUIDv7 identifiers | Use UUIDv7 for externally visible sortable IDs. Keep database uniqueness constraints authoritative. |
| Testcontainers Go | **v0.42.0** (retain) | PostgreSQL, Redis, NATS, and ClickHouse integration tests | Use real dependencies for repository, tenant-isolation, queue, and migration tests. Update the existing PostgreSQL 15 test image to the production major. |
| TanStack Query | **v5 current stable, exact lockfile pin** | Server-state caching, request cancellation, optimistic UI | Dashboard API reads/writes. Keep the Go API authoritative and define cache keys with workspace identity. |
| TanStack Table | **v8 current stable, exact lockfile pin** | Large link/campaign tables | Use for controlled sorting/filtering/pagination against server endpoints; do not fetch entire workspaces into the browser. |
| Apache ECharts | **v6 current stable, exact lockfile pin** | Dense time-series and categorical analytics charts | Analytics pages that need zoom, tooltips, stacked series, and large datasets. Wrap it behind a small chart component API to keep visual semantics consistent. |
| Zod + ts-rest | **Retain current v3 contract generation during foundation** | Existing TypeScript schemas and OpenAPI generation | Preserve the small working contract pipeline first. Evaluate Zod 4 / generator replacement only as a separate change after generated Go types and contract CI remove drift. |
| Clerk Go SDK / Resend Go SDK | Retain current major; update with automated PRs and integration tests | Identity provider integration and transactional email | These are already wired. Hide them behind application interfaces so tenancy and notification behavior do not become provider-specific. |

### Development and Operations Tools

| Tool | Purpose | Recommendation |
|------|---------|----------------|
| Docker Compose | One-command local dependencies | Start with PostgreSQL 18, Redis 8.6, and an OTel Collector. Add NATS and ClickHouse only when their phases begin. Pin image versions/digests; include health checks and persistent volumes. |
| golangci-lint | Go static analysis | Pin the executable in CI rather than relying on an ambient version. Keep the repository config, and run it with Go 1.27. |
| govulncheck | Go dependency and reachability scanning | Run on every pull request and scheduled builds. This is especially relevant to HTTP, database, and auth dependencies. |
| Renovate or Dependabot | Controlled dependency updates | Group OpenTelemetry modules, Next/React, and testcontainer images; do not auto-merge major upgrades or data-store image changes. |
| k6 | Redirect/API load tests | Add before the redirect milestone is accepted. Test hot cache, cold cache, missing link, Redis outage, event-broker slowdown, and long-tail latency. |
| GitHub Actions or equivalent CI | Reproducible quality gate | Pin actions; run Go tests/race tests, migrations, generated-file drift, TypeScript lint/typecheck/tests, container integration tests, vulnerability scans, and builds. |

## Data-Store and Queue Roles

| Concern | Use | Do Not Use |
|---------|-----|------------|
| Link/workspace/domain configuration | PostgreSQL | Redis or ClickHouse as authority |
| Redirect lookup | Redis cache, PostgreSQL fallback | Synchronous analytics query |
| Retryable command/job | Asynq | NATS event stream for one-off email/job semantics |
| Domain/event fan-out and replay | NATS JetStream | Asynq as a multi-consumer event log |
| Early event facts and small analytics | Partitioned PostgreSQL tables and summary tables | A second database before workload evidence |
| High-volume analytics | ClickHouse facts/projections fed by consumers | Cross-database transactional writes from request handlers |
| Search | PostgreSQL full-text/trigram indexes | Elasticsearch before PostgreSQL search fails measured requirements |

## Adoption Sequence

1. **Foundation stabilization**
   - Upgrade Go 1.27.1, Echo v4.15.4, pgx v5.11.0, and go-redis v9.22.0 behind smoke and integration tests.
   - Standardize local/test PostgreSQL on 18 and Redis on a supported pinned patch.
   - Keep Asynq v0.26.0, but move worker startup to its own binary and remove package-global dependencies.
   - Add sqlc, generated OpenAPI-to-Go boundary checks, OpenTelemetry/OTLP, deterministic migrations, and CI.

2. **Identity, workspaces, and links**
   - Use PostgreSQL `workspace_id` columns, composite constraints/foreign keys, and transaction-scoped tenant context.
   - Add row-level security as defense in depth only after tests prove the application role cannot bypass it; table owners and `BYPASSRLS` roles can bypass policies, so application predicates and isolation tests remain mandatory.
   - Keep Redis non-authoritative and invalidate redirect cache after committed link/domain changes.

3. **Redirect data plane**
   - Add a separate Go `redirector` binary using the same domain packages, pgx, and go-redis.
   - Put strict timeouts on Redis/PostgreSQL and event publishing. A failed analytics publish must never prevent the redirect response.
   - Give redirect cache its own memory/eviction policy and, before high traffic, its own Redis failure domain from Asynq.

4. **Tracking pipeline**
   - Introduce NATS JetStream v2.14.6 and nats.go v1.54.0 when click events get multiple consumers or replay requirements.
   - Use durable consumers, explicit acknowledgments, bounded retries/dead-letter handling, event IDs, and idempotent consumers. Treat delivery as at-least-once and make deduplication explicit.
   - Continue using Asynq for email/webhook attempts and other command-style jobs.

5. **Analytics scale-up**
   - Begin with PostgreSQL partitions and rollups if they meet load tests.
   - Add ClickHouse 26.8 LTS when event ingestion/query load measurably harms transactional latency, retention makes PostgreSQL materially expensive, or product queries miss their latency SLO under expected concurrency.
   - Dual-write only through replayable event consumers, never from the redirect handler. Compare aggregate results during migration before switching reads.

6. **Frontend slices**
   - Create the dashboard on Next.js 16.3.8/React 19.2.7/Node 24 LTS, using the versioned Go API.
   - Server Components may compose reads for rendering, but Next route handlers/server actions must not become a second business-logic API or directly access product tables.

## Installation / Pinning Reference

These commands document target pins; apply them as reviewed changes with tests rather than as a bulk upgrade.

```bash
# Foundation Go dependencies
go get github.com/labstack/echo/v4@v4.15.4
go get github.com/jackc/pgx/v5@v5.11.0
go get github.com/redis/go-redis/v9@v9.22.0
go get github.com/hibiken/asynq@v0.26.0

# Generated boundaries
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.2

# OpenTelemetry foundation
go get go.opentelemetry.io/otel@v1.47.0
go get go.opentelemetry.io/otel/sdk@v1.47.0
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace@v1.47.0
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@v0.71.0

# Tracking phase
go get github.com/nats-io/nats.go@v1.54.0

# ClickHouse phase
go get github.com/ClickHouse/clickhouse-go/v2@v2.48.0

# Frontend phase; keep exact versions in bun.lock
bun add next@16.3.8 react@19.2.7 react-dom@19.2.7
bun add tailwindcss@^4.3 @tanstack/react-query@^5 @tanstack/react-table@^8 echarts@^6
```

## Alternatives Considered

| Recommended | Alternative | When the Alternative Is Better |
|-------------|-------------|--------------------------------|
| Echo v4 | Echo v5 | A later planned migration can absorb handler/middleware breaking changes and benchmarks show a product benefit. It is unnecessary for foundation stabilization. |
| pgx + sqlc | GORM or another ORM | A domain dominated by simple CRUD and a team that accepts less control over SQL. Flux needs explicit tenant predicates, conflict handling, bulk event operations, and predictable query plans. |
| PostgreSQL-first analytics | ClickHouse from day one | Launch traffic, retention, or contractual query SLOs are already known to exceed a single PostgreSQL deployment, and the team can operate the extra store immediately. |
| NATS JetStream | Redis Streams | A constrained early deployment needs the fewest services and uses a dedicated Redis instance. Migrate before multi-consumer replay, independent scaling, and broker isolation become important. |
| NATS JetStream | Kafka-compatible platform | Sustained scale, ecosystem connectors, long retention, or an existing platform team already standardizes on Kafka/Redpanda. None is established for this milestone. |
| Asynq for jobs | JetStream work queues | All asynchronous work is already event-oriented and operationally standardized on NATS. Until then, Asynq's job semantics are a better fit for retries and schedules. |
| Next.js 16 | Vite SPA | The dashboard becomes entirely client-side, SEO/SSR/streaming offer no value, and simpler static hosting matters more than the chosen product architecture. |
| Clerk | Self-hosted auth | Regulatory, data-residency, enterprise federation, or provider economics justify owning credential/session complexity. Keep provider boundaries so this remains possible. |
| Managed ClickHouse | Self-hosted ClickHouse | The team has proven database operations capacity, workload economics favor it, and backup/upgrade/on-call ownership is explicit. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Go 1.25 in production | It is outside the two-current-major support window on 2026-10-05. | Go 1.27.1, pinned in CI and build images. |
| Echo v4.15.2 | It predates an encoded-separator route-bypass fix. | Echo v4.15.4. |
| Echo v5 migration during foundation | It expands a stabilization phase into a framework migration with no product requirement. | Patched Echo v4 with narrow adapters. |
| Redis as authoritative link storage | Eviction, flushes, topology changes, or TTL mistakes would break link correctness. | PostgreSQL authority plus cache-aside Redis. |
| Asynq as the click event log | A task queue does not provide the clean multi-consumer replay model needed for analytics, attribution, enrichment, and webhooks. | NATS JetStream for domain events; Asynq for jobs. |
| ClickHouse on the redirect request path | Analytics outages would affect redirect availability and latency. | Publish an event with a strict bound; consume asynchronously. |
| Kafka in initial milestones | Operational cost and partition/rebalance concerns exceed the demonstrated need. | NATS JetStream when event streaming begins. |
| Kubernetes in initial milestones | It adds control-plane and deployment complexity before scaling characteristics are known. | Containers on a managed application platform; add orchestration only for a demonstrated need. |
| Elasticsearch for initial search | PostgreSQL full-text/trigram search satisfies the initial link/campaign scope and avoids another stateful cluster. | PostgreSQL search; add a search service only after measured failure. |
| GORM introduced alongside pgx | Two persistence styles obscure transaction and tenant boundaries and increase review burden. | pgx + sqlc + small repository interfaces. |
| New Relic-only instrumentation in domain code | It couples telemetry to one vendor and conflicts with the spec's OpenTelemetry requirement. | OTel API/SDK + Collector + OTLP export, retaining New Relic as a backend if desired. |
| Business logic in Next.js route handlers or server actions | It creates a second backend and inconsistent authorization/validation paths. | Versioned Go API; Next.js is a client/composition layer. |
| `latest`, canary, or RC datastore/container tags | Builds and runtime behavior become unreproducible; data upgrades can occur unintentionally. | Exact patch tags and preferably immutable image digests. |

## Version Compatibility and Upgrade Notes

| Package / Service | Compatible With | Notes |
|-------------------|-----------------|-------|
| Go 1.27.1 | pgx v5.11, nats.go v1.54, OTel Go v1.47 | nats.go v1.54 requires Go 1.26+, and current OTel explicitly tests Go 1.26/1.27. |
| Echo v4.15.4 | Existing Echo v4 handlers/middleware | Review CSRF/CORS/static routing behavior; v4.15 introduced Fetch Metadata-aware CSRF behavior. |
| PostgreSQL 18.6 | pgx v5.11.0 | Align Testcontainers and local Compose with production major. Always run current minor; PostgreSQL says minor updates are safer than remaining on an older minor. |
| Redis 8.6.x | go-redis v9.22, Asynq v0.26 | Run Asynq scheduling, retry, Lua/script, ACL/TLS, failover, and stream tests before production. If a provider is behind, supported Redis 7.4 is acceptable temporarily. |
| OTel SDK v1.47 | contrib instrumentation v0.71 | Upgrade the family together and pin semantic conventions. Avoid duplicate New Relic and OTel middleware during transition. |
| NATS Server v2.14.6 | nats.go v1.54 | v1.54 also understands future v2.15 info fields; do not deploy the v2.15 release candidate. Use a three-node JetStream cluster for production durability when introduced. |
| ClickHouse 26.8 LTS | clickhouse-go v2.48 | Pin the current LTS patch and test schema/query compatibility before each monthly/LTS update. Prefer native batch writes from consumers. |
| Next.js 16.3.8 | Node 24 LTS; React 19.2.7 | 16.3.8 is the patched Active LTS release. Treat framework security patches as urgent and keep Next/React upgrades grouped. |
| TypeScript workspace | One exact compiler line | The repository currently mixes TypeScript 5.8 and 6.0 ranges. Normalize only after building every package; avoid multiple compiler interpretations of shared declarations. |

## Confidence Assessment

| Area | Confidence | Basis |
|------|------------|-------|
| Preserve Go/Echo/pgx/PostgreSQL/Redis/Asynq | HIGH | Source-inspected existing code and explicit project constraints. |
| Immediate Go/Echo/PostgreSQL patch guidance | HIGH | Official release/support/security documentation. |
| OpenTelemetry as the primary telemetry API | HIGH | Explicit product requirement and stable official Go trace/metric support. |
| sqlc and OpenAPI generation boundary | HIGH | Fits existing direct-SQL and TypeScript-contract scaffolds without replacing frameworks. |
| NATS JetStream at tracking phase | HIGH | Explicit spec preference and multi-consumer/replay requirements; exact capacity must be load-tested. |
| PostgreSQL-to-ClickHouse adoption point | MEDIUM | Correct architectural direction, but the threshold depends on event volume, retention, query shapes, hardware/provider, and SLOs. |
| Frontend library details beyond Next/React/Tailwind | MEDIUM | Appropriate current majors; exact choices should be validated during the first dashboard slice against design and bundle requirements. |

## Sources

### Runtime and foundation

- [Go downloads](https://go.dev/dl/) and [Go release policy/history](https://go.dev/doc/devel/release) — current Go 1.27.1 and two-newer-major support policy (HIGH).
- [Echo releases](https://github.com/labstack/echo/releases) and [Echo changelog](https://github.com/labstack/echo/blob/master/CHANGELOG.md) — current v4 patch, security fix, and v4.15 middleware changes (HIGH).
- [PostgreSQL versioning policy](https://www.postgresql.org/support/versioning/) — PostgreSQL 18.6 current minor, five-year support, and minor-update guidance (HIGH).
- [PostgreSQL row security documentation](https://www.postgresql.org/docs/current/ddl-rowsecurity.html) — RLS behavior and bypass caveats (HIGH).
- [pgx changelog](https://github.com/jackc/pgx/blob/master/CHANGELOG.md) — v5.11.0, Go support, PostgreSQL 18/protocol/security changes (HIGH).
- [Redis Open Source release notes](https://redis.io/docs/latest/operate/oss_and_stack/stack-with-enterprise/release-notes/redisce/) and [Redis security support policy](https://github.com/redis/redis/security) — supported Redis branches and security-patch guidance (HIGH).
- [go-redis releases](https://github.com/redis/go-redis/releases) — current v9 client line (HIGH).
- [Asynq releases](https://github.com/hibiken/asynq/releases) — v0.26.0 and Go version requirements (HIGH).

### Events, analytics, and observability

- [NATS JetStream documentation](https://docs.nats.io/learn/) — persisted streams, consumers, acknowledgments, replication, and operational guidance (HIGH).
- [NATS Server releases](https://github.com/nats-io/nats-server/releases) and [nats.go releases](https://github.com/nats-io/nats.go/releases) — current stable server/client selection and Go compatibility (HIGH).
- [ClickHouse official packages](https://packages.clickhouse.com/) and [support policy](https://clickhouse.com/legal/support-services-policy) — current stable/LTS releases and supported-release policy (HIGH).
- [ClickHouse Go integration](https://clickhouse.com/integrations/go) and [clickhouse-go releases](https://github.com/ClickHouse/clickhouse-go/releases) — official driver and batch/native protocol support (HIGH).
- [OpenTelemetry Go documentation](https://opentelemetry.io/docs/languages/go/) and [instrumentation guidance](https://opentelemetry.io/docs/languages/go/instrumentation/) — signal stability and SDK/provider requirements (HIGH).
- [OpenTelemetry Go releases](https://github.com/open-telemetry/opentelemetry-go/releases) and [contrib releases](https://github.com/open-telemetry/opentelemetry-go-contrib/releases) — current aligned module families and Go support (HIGH).

### API, database tooling, and frontend

- [sqlc releases](https://github.com/sqlc-dev/sqlc/releases) — v1.31.1 (HIGH).
- [oapi-codegen releases](https://github.com/oapi-codegen/oapi-codegen/releases) — v2.7.2 and runtime v1.2+ requirement (HIGH).
- [Node.js release policy](https://nodejs.org/en/about/previous-releases) — production LTS guidance and Node 24/26 status (HIGH).
- [Next.js release/security feed](https://nextjs.org/blog) and [Next.js releases](https://github.com/vercel/next.js/releases) — 16.3.8 Active LTS patched release (HIGH).
- [React versions](https://react.dev/versions) — stable React patch history (HIGH).
- [Tailwind CSS v4.3](https://tailwindcss.com/blog/tailwindcss-v4-3) — current stable feature release and Next/Webpack integration (HIGH).
- [Bun 1.3 release](https://bun.com/blog/bun-v1.3) and [Bun 1.4 rewrite status](https://bun.com/blog/bun-in-rust) — stable baseline versus canary transition (MEDIUM; exact latest patch should be rechecked at upgrade time).

---
*Stack research for: Flux link attribution and marketing analytics platform*
*Researched: 2026-10-05*
