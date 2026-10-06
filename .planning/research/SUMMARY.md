# Project Research Summary

**Project:** Flux
**Domain:** Production-grade multi-tenant link attribution and marketing analytics SaaS
**Researched:** 2026-10-05
**Confidence:** HIGH for the product foundation, architectural boundaries, and v1 dependency order; MEDIUM for workload- and provider-dependent choices

## Executive Summary

Flux is a brownfield Go project intended to become a multi-tenant attribution product, not merely a URL shortener. Its credible first release must prove one complete loop: a workspace member creates a managed-domain link, a visitor is redirected quickly, Flux records the click without coupling redirect availability to analytics, a customer backend sends an idempotent lead or purchase, Flux applies an explainable first- or last-click model, and the marketer can inspect clicks, conversions, revenue, freshness, and the exact attribution decision. Experts build this as a small number of independently deployable process roles around one transactional authority, with explicit tenant scope, asynchronous facts, replay-safe consumers, and stable metric contracts.

The strongest recommendation is to preserve and repair the existing Go/Echo/pgx/PostgreSQL/Redis/Asynq foundation, then evolve it into four Go composition roots: `api`, `redirector`, `worker`, and `migrate`. PostgreSQL remains authoritative for tenant-owned configuration, conversions, attribution decisions, idempotency, and outbox records; Redis is a disposable redirect cache and coordination layer; Asynq handles finite retryable jobs; NATS JetStream should carry durable, replayable click and conversion facts once tracking begins. Begin analytics in PostgreSQL behind an explicit storage/query boundary and introduce ClickHouse only when benchmarks show transactional contention, retention cost, or dashboard latency warrants it. Build the Next.js dashboard as a client of the versioned Go API, with no duplicate business rules or direct data-store access.

The main risks are architectural correctness failures rather than missing breadth: cross-tenant leakage, stale redirect cache, clicks lost between response and broker acceptance, at-least-once duplicates, unstable identity matching, mutable attribution, undefined metrics, and privacy data copied into secondary systems. Prevent them from the first schema with workspace-scoped repositories and constraints, a transactional outbox, versioned cache snapshots and events, durable idempotency, deterministic attribution rules, a written metric dictionary, and end-to-end retention/deletion design. The tracking phase must resolve the click-loss SLO and local-spool deployment model through a spike before it can claim durable capture. Custom domains, developer webhooks, and ClickHouse each require focused validation when their phases begin.

## Research Baseline: Observed vs Planned

The repository currently contains useful scaffolding, not a partially finished product. Source-observed capabilities are limited to an Echo service foundation, configuration and logging, middleware, health and API-documentation routes, PostgreSQL/Redis initialization, migrations, Asynq wiring, TypeScript Zod/ts-rest/OpenAPI packages, and React Email templates. The product database schema is empty, `/api/v1` and repositories are placeholders, the frontend is empty, the redirect data plane does not exist, and there are no executable product test suites.

Everything else in this summary—including workspaces, links, redirects, click events, conversion ingestion, attribution, analytics, custom domains, public APIs, and webhooks—is planned capability. Roadmap phases should repair the existing foundation incrementally and validate each new capability as a working vertical slice; they should not treat package names or initialized clients as evidence that a product domain already exists.

## Key Findings

### Recommended Stack

Retain the current Go-first direction and upgrade unsupported or security-sensitive pins during foundation work. Use direct SQL with pgx and sqlc, generated OpenAPI-to-Go boundaries, exact dependency pins, real dependency integration tests, and OpenTelemetry through an OTLP Collector. Avoid a framework rewrite, early microservices, Kafka, Kubernetes, Elasticsearch, and ClickHouse-by-default: none improves the first attribution loop enough to justify its operational cost.

**Core technologies:**

- **Go 1.27.1:** API, redirector, worker, and migration binaries — current supported toolchain that preserves the existing implementation model.
- **Echo v4.15.4:** HTTP routing and middleware — retain the working v4 adapters on the patched maintenance release; defer Echo v5.
- **PostgreSQL 18 current minor + pgx v5.11.0 + sqlc v1.31.1:** canonical transactional state — explicit SQL supports tenant predicates, constraints, idempotency, locking, and predictable query plans.
- **Redis 8.6 current patch + go-redis v9.22.0:** redirect cache, rate-limit coordination, and transient acceleration — never authoritative for link configuration or durable business idempotency.
- **Asynq v0.26.0:** finite jobs such as email, DNS verification, exports, metadata fetches, and webhook attempts — keep separate from immutable event-stream semantics.
- **NATS JetStream v2.14.6 + nats.go v1.54.0:** replayable tracking facts and independent consumers — introduce with durable click tracking, after the loss/latency spike.
- **ClickHouse 26.8 LTS + clickhouse-go v2.48.0:** high-volume analytical facts and rollups — conditional adoption after benchmark evidence, never on the redirect path or as canonical state.
- **Next.js 16.3.8 + React 19.2.7 + Node 24 LTS + Tailwind CSS 4.3:** dashboard runtime and UI — the frontend consumes versioned Go APIs and contains no authoritative business logic.
- **OpenTelemetry Go v1.47/contrib v0.71 + Collector:** cross-process traces, metrics, and context propagation — use from foundation onward, with redaction and controlled cardinality.
- **Tern v2.4.1, oapi-codegen v2.7.2, and Testcontainers Go v0.42.0:** deterministic migrations, enforced API boundaries, and integration tests against real infrastructure.

**Critical version requirements:** pin the Go toolchain, database/container images, Next/React pair, and OpenTelemetry family; apply current PostgreSQL/Redis patch releases; align local, CI, and production database majors; use Node 24 LTS rather than Node 26 Current; keep Bun on a tested stable 1.3 patch rather than a canary line. Recheck exact patches at implementation time because those values are intentionally time-sensitive.

### Expected Features

Flux should compete on trustworthy attribution and redirect reliability. The first release should be judged by whether a new workspace can complete and understand the click-to-revenue loop, including failure and unattributed states.

**Must have (v1 table stakes):**

- Authentication, workspaces, invitations, fixed owner/admin/member/viewer roles, and isolation across every store and asynchronous path.
- Managed-domain link creation with generated or custom keys, destination validation, edit/enable/disable/archive/restore, search/filter/list, and cursor pagination.
- An independent cached redirect path with safe missing/disabled behavior and p95 targets of under 30 ms cached and under 100 ms uncached.
- Durable asynchronous click facts with stable event IDs, bot separation, clear unique-visitor semantics, coarse geo/device/referrer/UTM enrichment, and visible processing freshness.
- A documented click-ID handoff plus explicit anonymous-to-known customer association; unknown evidence produces an unattributed result rather than a guessed identity.
- An authenticated server Events API for `lead` and `purchase`, with tenant-scoped durable idempotency, request-hash conflict detection, timestamps, integer minor-unit money, and ISO currency.
- Deterministic first- and last-click attribution using one documented default window, with a persisted model/policy version, credited touchpoint, inputs, and reason.
- Focused analytics for clicks, unique visitors, leads, purchases, conversion rate, revenue, average order value, time series, top links, and bounded referrer/country/device/UTM breakdowns.
- Event and attribution drill-down, explicit delayed/duplicate/rejected/unattributed states, data freshness, setup verification, rate limits, abuse suspension, and privacy-safe collection.

**Should have after the core loop is validated (v1.x competitive breadth):**

- Custom domains with separate ownership, DNS, certificate, activation, and failure states using managed TLS infrastructure.
- Campaigns, tags, folders, UTM templates, QR export, bulk actions, and CSV import/export.
- A minimal browser SDK for first-party identity handoff and trusted non-monetary events, while purchases remain server-authoritative.
- Public Links and Analytics APIs, scoped and rotatable API keys, consistent errors/pagination/rate-limit metadata, and stable OpenAPI contracts.
- Signed webhooks with delivery history and retries, one Stripe adapter into the canonical Events API, saved analytics views, and CSV export.
- Security-relevant audit history and simple link controls such as expiration, password protection, and fallback destinations after redirect rules are stable.

**Defer (v2+):**

- Partner/affiliate programs, commissions, payouts, partner portal, and agency-wide cross-workspace consoles.
- SaaS billing, plans, quotas, invoices, and broad entitlement enforcement until packaging and usage drivers are known.
- Dozens of native integrations, advanced geo/device/A/B/deep-link routing, and cross-device/mobile attribution.
- Multi-touch or machine-learning attribution, configurable model proliferation, journey/cohort/LTV analytics, custom report builders, and AI-generated insights.
- Enterprise SAML/SCIM/custom RBAC, data residency, advanced governance, hosted pages/link-in-bio, and bespoke certificate infrastructure.

### Architecture Approach

Use a modular monolith with multiple process composition roots rather than a service per domain. The control plane handles authenticated management and canonical transactions; the redirect data plane resolves `(normalized_host, normalized_path)` through Redis with a PostgreSQL fallback and returns the redirect independently of analytics. Committed control-plane changes write transactional outbox records for cache invalidation and downstream effects. Tracking uses versioned immutable envelopes, stable event IDs, explicit acknowledgements after durable writes, and idempotent consumers. Attribution is a pure deterministic core wrapped by a PostgreSQL-backed application shell. Analytics facts and projections are rebuildable and accessed only through one query service, allowing PostgreSQL-to-ClickHouse migration without changing API semantics.

**Major components:**

1. **API process** — authenticated control plane, developer endpoints, workspace policy, transactional use cases, and canonical writes.
2. **Redirector process** — public host/path normalization, Redis cache-aside lookup, PostgreSQL fallback, routing decision, local event handoff, and immediate redirect.
3. **Worker process** — Asynq jobs and JetStream consumers selected by role; no accidental startup inside API replicas.
4. **Migrator process** — runs PostgreSQL migrations exactly once as a release step, independently of HTTP processes.
5. **Identity/workspace module** — internal users, memberships, invitations, fixed roles, and workspace-scoped authorization.
6. **Link/domain module** — canonical link rules, uniqueness, lifecycle, redirect snapshots, domain state, and outbox records.
7. **Tracking/event pipeline** — versioned envelopes, local durability policy, JetStream streams, dedupe, enrichment, replay, and poison-event handling.
8. **Conversion/attribution module** — durable conversion acceptance, identity correlation, pure first/last-click rules, and persisted explanations.
9. **Analytics write/query boundary** — append-only facts, aggregates, stable metric DTOs, freshness, reconciliation, and a conditional ClickHouse adapter.
10. **Frontend** — accessible, information-dense Next.js dashboard that calls the Go API and reflects eventual consistency explicitly.

**Key patterns:** required tenant scope in every repository/query, PostgreSQL constraints as authority, transaction-local RLS only if safely adopted, transactional outbox, versioned immutable redirect cache snapshots, separate facts from finite jobs, stable event IDs, idempotent consumers, pure attribution functions, and process-specific liveness/readiness.

### Critical Pitfalls

1. **Claiming click durability with an in-memory queue** — define a measurable loss SLO and validate a bounded local spool or explicitly documented bounded-loss design under crashes, broker outage, disk pressure, shutdown, and latency load.
2. **Losing tenant scope below HTTP middleware** — put `workspace_id` into schemas, method signatures, events, cache keys, uniqueness rules, and analytical queries; extend a two-workspace denial matrix in every phase.
3. **Allowing redirect cache to drift from canonical links** — normalize host/path once, enforce uniqueness in PostgreSQL, write outbox events with link changes, use versioned snapshots and bounded positive/negative TTLs, and test lost or reordered invalidations.
4. **Treating at-least-once delivery as exactly-once business behavior** — preserve producer event IDs across retries, deduplicate at every durable sink, acknowledge only after commit, and reconcile broker facts, raw rows, aggregates, conversions, and attribution decisions.
5. **Making identity, attribution, or metrics look more certain than they are** — model click/anonymous/session/customer identifiers separately, use deterministic event-time rules and stable tie-breakers, preserve decision evidence, define each metric centrally, and show freshness and unattributed reasons.
6. **Adding privacy controls only to the primary event table** — classify fields before collection and carry retention, deletion, redaction, and replay-suppression policy through broker data, caches, logs, traces, dead letters, analytical stores, exports, and backups.
7. **Collapsing custom-domain state or exposing unsafe outbound HTTP** — model ownership/DNS/TLS/activation separately; for metadata and webhooks, enforce connection-time SSRF defenses, bounded redirects, raw-body signatures, stable delivery IDs, and restricted egress.

## Strongest Decisions

These decisions are supported consistently across the project brief and all research tracks:

1. Preserve the sound Go/Echo scaffold and repair it incrementally; do not rewrite the router or split into many network services.
2. Deploy API, redirector, worker, and migration roles independently from one Go module so lifecycle and scaling match their traffic profiles.
3. Keep PostgreSQL authoritative. Redis is derived/transient; ClickHouse is analytical; neither owns link routing or authorization.
4. Put tenant ownership and authorization into the first product schema and application/repository contracts. RLS is optional defense in depth, not a substitute.
5. Separate redirect success from analytics availability. The redirector never performs enrichment, attribution, webhooks, email, or dashboard queries inline.
6. Use transactional outbox records for canonical changes and stable event IDs plus idempotent consumers for every at-least-once path.
7. Use JetStream for replayable facts and Asynq for finite commands. Do not use Asynq as the analytics event log.
8. Ship a narrow server-authoritative conversion API and deterministic first-/last-click models before browser breadth, integrations, or advanced attribution.
9. Define metric semantics and privacy/retention before freezing event schemas or building polished dashboards.
10. Keep ClickHouse, custom domains, and public webhook breadth behind evidence and phase-specific validation rather than treating long-term architecture as day-one scope.

## V1 Scope Boundary

The v1 milestone ends when the following journey works under retries, delayed consumers, process restarts, cache invalidation, and two-workspace isolation tests:

```text
member creates a managed-domain link
  → visitor receives a fast redirect and click ID
  → click survives the chosen durability envelope and is enriched asynchronously
  → customer backend sends an idempotent lead or purchase
  → Flux links available identity evidence and applies first- or last-click
  → marketer sees clicks, conversion, integer revenue, freshness, and the decision reason
```

The v1 UI should include the link management surface, setup verifier, core analytics, and event/attribution drill-down needed to complete and trust this journey. The managed platform domain is sufficient for v1 acceptance. Custom domains and productivity tooling are the first v1.x expansion because they broaden adoption without redefining the attribution core.

## Implications for Roadmap

Based on the combined research, use the following phase structure. Phases 1–6 form the v1 proof; Phases 7–8 are v1.x; Phase 9 is intentionally deferred expansion.

### Phase 1: Foundation Stabilization and Process Boundaries

**Rationale:** Every later requirement depends on reliable construction, migration ownership, lifecycle, health, tests, and telemetry. The repository has useful initialization code but source-confirmed coupling and correctness gaps.
**Delivers:** Patched/pinned toolchains and dependencies; reproducible CI; executable unit/integration/migration/lifecycle tests; `api`, `worker`, and `migrate` composition roots; correct SIGTERM shutdown and drain order; process-specific `/live` and `/ready`; embedded/configured assets; OpenTelemetry; enforced OpenAPI/Go contract generation.
**Addresses:** The production foundation required by every v1 feature.
**Avoids:** Mistaking initialization scaffolding for runtime readiness, accidental workers in API replicas, concurrent app-start migrations, working-directory asset failures, and unobservable failure paths.

### Phase 2: Tenant Kernel and Managed-Domain Link Control Plane

**Rationale:** Workspace ownership must exist before the first tenant-owned record, and canonical link rules must exist before any redirect cache.
**Delivers:** Internal user mapping, workspaces, invitations, fixed roles, authorization policies, workspace-scoped repositories, first product DDL, link create/edit/enable/disable/archive/restore/search/filter/list, canonical host/path/key normalization, destination policy, database uniqueness, durable idempotency where needed, and outbox records. Add the first focused dashboard slices against the versioned API.
**Addresses:** Authentication, team membership, tenant isolation, core link management, pagination, destination validation, and abuse states.
**Avoids:** Cross-tenant access, application-only collision checks, unscoped IDs, provider identity claims as authorization, and later tenancy rewrites.

### Phase 3: Independent Redirect Data Plane

**Rationale:** Redirect correctness, availability, and latency need a clean baseline before event-pipeline overhead is introduced.
**Delivers:** A dedicated redirector binary; trusted-proxy handling; canonical host/path resolution; Redis cache-aside versioned snapshots; PostgreSQL read fallback; outbox-driven invalidation; positive/negative TTLs; safe missing/disabled/expired responses; separate resource pools; redirect telemetry and load tests against p95 targets.
**Addresses:** Reliable public redirects and immediate managed-domain link usability.
**Avoids:** Redis authority, stale destinations, inconsistent normalization, cache-miss collapse into shared PostgreSQL, and analytics on the critical path.

### Phase 4: Durable Tracking, Identity Envelope, and Privacy Baseline

**Rationale:** Trustworthy analytics requires a measured durability contract, stable identifiers, replay behavior, and data-minimization rules before production traffic is collected.
**Delivers:** The click-loss SLO decision; validated local-spool or bounded-loss implementation; JetStream and versioned event envelopes; stable event IDs; explicit ack/retry/dedupe/poison handling; click/session/anonymous identifiers; bot policy; asynchronous geo/device/referrer/UTM enrichment; field inventory and retention; lag/freshness/spool metrics; replay and failure-injection tests.
**Addresses:** Asynchronous click capture, unique-visitor semantics, minimal enrichment, processing transparency, and privacy-conscious identity.
**Avoids:** Lost in-memory clicks, broker-coupled redirects, duplicate facts, fingerprinting, raw-IP sprawl, and deletion resurrection through replay.

### Phase 5: Conversion Events and Explainable Attribution

**Rationale:** Click identity and durable facts must exist before conversions can be correlated and credited deterministically.
**Delivers:** Authenticated server Events API for leads and purchases; tenant-scoped idempotency keys with request hashes; integer amount/currency validation; click-ID handoff guidance; explicit anonymous-to-customer association; event-time/late-arrival/tie-break rules; pure first-/last-click models; persisted model/window/version/candidates/reason; unattributed and conflict states; setup verification from sample click through attributed conversion.
**Addresses:** The product's core differentiator: server-authoritative, auditable click-to-revenue attribution.
**Avoids:** Duplicate purchases, Redis-only idempotency, mutable historical credit, arrival-order-dependent results, invented identity certainty, and opaque unattributed conversions.

### Phase 6: Analytics Contract, Dashboard, and V1 Hardening

**Rationale:** The dashboard becomes trustworthy only after canonical facts and decisions exist. Metric definitions should precede storage tuning and visual polish.
**Delivers:** Metric dictionary; one analytics query service; PostgreSQL facts/partitions/rollups initially where benchmarks permit; clicks/uniques/leads/purchases/conversion rate/revenue/AOV; time series and bounded breakdowns; date/link filters; top links; event and attribution drill-down; `data_as_of` and lag; raw-to-aggregate reconciliation; projection rebuild tests; end-to-end v1 acceptance, abuse, isolation, privacy, and degraded-dependency tests.
**Addresses:** Focused analytics, operator diagnostics, empty/delayed/error states, and a complete user-facing activation journey.
**Avoids:** Double-counted enrichment, current-state history rewrites, undefined metrics, hidden eventual consistency, premature ClickHouse, and charts that disagree.

### Phase 7: Custom Domains and Campaign Productivity

**Rationale:** These are expected category capabilities but should build on proven link, redirect, invalidation, and analytics semantics.
**Delivers:** Provider-managed custom hostnames/TLS with separate ownership/DNS/certificate/activation states; safe detach/reclaim and renewal monitoring; campaigns, tags, folders, UTM templates, QR export, bulk actions, and CSV import/export; basic audit history.
**Addresses:** Branded links and the workflows real teams need at higher link counts.
**Avoids:** A single misleading `verified` flag, duplicate hostname claims, unsafe cutovers, historical grouping drift, and bulk actions without stable entities.

### Phase 8: Developer Platform, Webhooks, and First Integration

**Rationale:** Public compatibility commitments should follow settled internal resource and event contracts.
**Delivers:** Scoped/rotatable API keys, public Links and Analytics APIs, documented pagination/errors/rate-limit metadata, stable OpenAPI, a minimal browser helper/SDK, signed webhooks with persistent delivery attempts and manual replay, SSRF-safe egress, and one Stripe conversion adapter into the canonical Events API.
**Addresses:** Automation, customer system integration, first-party identity handoff, and event export.
**Avoids:** Duplicate client effects, unstable API promises, reversible API-key storage, signature/replay flaws, DNS-rebinding SSRF, and integration-specific attribution models.

### Phase 9: Commercial, Partner, and Enterprise Expansion

**Rationale:** These capabilities form distinct products and depend on validated attribution, integer money, authorization, audit, usage, and integration seams.
**Delivers:** Only after product evidence: billing/entitlements, partner programs and commissions, broader integrations, enterprise identity/governance, advanced routing, additional attribution models, and extended analytics.
**Addresses:** Segment expansion and monetization after core value is proven.
**Avoids:** Allowing billing, payouts, enterprise policy, or ML complexity to obscure flaws in the core attribution loop.

### Phase Ordering Rationale

- Lifecycle and composition boundaries come first because redirect independence is not real while HTTP replicas own workers or migrations.
- Tenant scope precedes product records; canonical link rules precede cache snapshots; redirect behavior precedes click tracking.
- Stable click identity and durable event semantics precede conversion matching; accepted conversions precede attribution; canonical attribution decisions precede revenue dashboards.
- Metric contracts and a storage port precede ClickHouse so a data-store migration does not redefine product behavior.
- Public webhooks and integration contracts follow stabilized domain events because they become long-lived compatibility promises.
- Custom domains are pulled immediately after the v1 core loop: they matter commercially, but they do not validate attribution and carry a provider-specific operational lifecycle.

### Research Flags

Phases likely needing deeper research or a spike during planning:

- **Phase 2 — RLS only if adopted:** test non-owner roles, `FORCE ROW LEVEL SECURITY`, transaction-local tenant context in pooled connections, migration ownership, workers, backups, and referential-integrity side channels.
- **Phase 4 — redirect event durability (required spike):** choose the click-loss SLO, deployment disk model, spool technology and ownership, fsync/group-commit policy, capacity behavior, recovery, and measured latency cost.
- **Phase 4 — identity/enrichment privacy review:** decide click-ID propagation, cookie boundaries, consent signals, raw-IP lifecycle, bot policy, retention, and deletion across every copy before schema freeze.
- **Phase 6 — analytical-store benchmark:** select PostgreSQL or ClickHouse from expected event volume, retention, query shapes, batch throughput, replay time, dashboard p95, and OLTP contention; design ClickHouse ordering/partitioning only if selected.
- **Phase 7 — custom-domain provider validation:** verify the chosen provider's ownership, certificate issuance/renewal, rate limits, hostname migration, DNS drift, and reclaim behavior.
- **Phase 8 — webhook/egress security validation:** test HMAC vectors and rotation, retry/replay behavior, private IPv4/IPv6 ranges, DNS rebinding, redirects, body/time limits, and network egress policy.

Phases with established patterns that can usually skip a separate research phase:

- **Phase 1:** Go process composition, migrations, lifecycle, CI, OpenTelemetry, and generated-boundary patterns are well documented; implementation still needs repository-specific tests.
- **Phase 2:** PostgreSQL-backed workspaces, scoped repositories, fixed roles, links, and outbox patterns are standard once the RLS choice is isolated.
- **Phase 3:** Cache-aside redirect lookup and versioned invalidation are standard; use load/failure tests rather than broad research.
- **Phase 5:** Durable API idempotency and pure first-/last-click selection are established patterns after product policy questions are fixed.
- **Phase 9:** Do not research as one large phase; split each validated commercial or enterprise initiative into its own future project/phase when evidence exists.

## Deferred Choices and Adoption Triggers

| Choice | Current recommendation | Revisit when |
|--------|------------------------|--------------|
| Click durability | Spike a bounded local durable spool; if platform storage cannot support it, publish an explicit bounded-loss SLO | Deployment platform, acceptable-loss target, latency budget, and disk recovery model are known |
| PostgreSQL RLS | Use explicit application authorization and tenant-scoped SQL first; add RLS only as tested defense in depth | Runtime roles and pooled transaction-local context can be proven safe |
| ClickHouse | Keep analytics behind a port and start with PostgreSQL if benchmarks pass | Ingest/query load harms OLTP, retention cost grows materially, replay becomes impractical, or dashboard SLOs fail |
| Custom-domain provider | Use managed TLS/hostname infrastructure; do not build certificate machinery | Phase 7 provider evaluation against lifecycle, migration, limits, and cost |
| Both attribution models in first launch | Preserve first- and last-click in v1 because the product intent names both | Design-partner evidence shows one model is enough for an earlier technical preview |
| Browser SDK | Keep v1 conversion ingestion server-authoritative and provide a narrow click-ID handoff | Customers cannot complete identity handoff reliably without an SDK |
| Redis topology | Use supported managed Redis, with bounded cache state | Tracking traffic or incidents show cache and Asynq need separate failure domains; split before public scale |
| Frontend supporting libraries | TanStack Query/Table and ECharts are sensible defaults | First dashboard slice validates bundle, accessibility, interaction, and visualization requirements |
| Advanced infrastructure | Avoid Kafka, Kubernetes, Elasticsearch, and many services | Measured scale or an existing platform capability makes their operational cost worthwhile |

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Foundation choices match the inspected repository and official release/security documentation. Exact future patch numbers must be rechecked. NATS adoption timing is strong; ClickHouse timing remains workload-dependent. |
| Features | MEDIUM-HIGH | Category expectations and the full click-to-revenue loop are supported by official product documentation and project intent. Prioritization lacks direct target-user interviews, pricing evidence, and product telemetry. |
| Architecture | HIGH | Control/data-plane separation, four composition roots, PostgreSQL authority, outbox, cache-aside, tenant-scoped repositories, and idempotent consumers all align with product SLOs and current code constraints. |
| Pitfalls | HIGH | Tenant isolation, dual writes, retry/idempotency, cache consistency, webhook security, custom-domain lifecycle, and privacy propagation are well documented. The final click-durability mechanism and identity behavior are project-specific. |

**Overall confidence:** HIGH for roadmap order and v1 boundaries; MEDIUM for workload-, browser-, jurisdiction-, and provider-specific implementation choices.

### Gaps to Address

- **No target-user evidence:** validate activation time, metric comprehension, attribution-model value, and the next adoption barrier with the first 5–10 integrated workspaces before expanding v1.x.
- **Click-loss policy is undefined:** establish an explicit loss SLO and deployment storage model in the Phase 4 spike; do not label in-memory buffering durable.
- **Click-ID propagation is unsettled:** choose query parameter, first-party storage, server handoff, expiry, and privacy behavior using representative customer applications and browsers.
- **Launch traffic and retention are unknown:** benchmark PostgreSQL-first analytics with expected event rates and query shapes before committing to ClickHouse.
- **Custom-domain provider is unselected:** compare managed providers during Phase 7 against ownership proof, certificate states, zero-downtime cutover, renewal, reclaim, quotas, and cost.
- **Metric policies need product decisions:** define bot inclusion, uniqueness window, timezone, lateness, provisional enrichment, historical dimension snapshots, and data freshness before dashboard implementation.
- **Jurisdictions are unspecified:** perform a launch-jurisdiction privacy review covering purpose, consent, raw IP, cookies, retention, deletion, exports, backups, and replay suppression.
- **Operational targets beyond redirect latency are incomplete:** set broker lag, spool capacity/age, analytics freshness, replay duration, cache invalidation, recovery, and deletion SLOs as their phases begin.
- **Current scaffold requires direct verification during execution:** research identified likely health, signal, validation, OpenAPI, lifecycle, global-state, and asset-path issues; Phase 1 should convert each into a failing test or source-backed fix rather than assume all scaffolding is reusable.

## Sources

### Project and Repository Evidence (HIGH confidence)

- [Flux project definition](../PROJECT.md) — authoritative scope, constraints, observed scaffold, active requirements, and non-goals.
- [Stack research](STACK.md) — current version guidance and technology selection.
- [Feature research](FEATURES.md) — category baseline, v1 journey, dependency graph, and validation gaps.
- [Architecture research](ARCHITECTURE.md) — component boundaries, data flow, build order, and unresolved durability trade-off.
- [Pitfalls research](PITFALLS.md) — domain failure modes, prevention, verification, and recovery.
- `spec.md` and `.planning/codebase/` — authoritative long-term intent and source-inspected current-state evidence, as cited by the research files.

### Primary Technical Sources (HIGH confidence)

- [Go releases](https://go.dev/doc/devel/release), [Echo releases](https://github.com/labstack/echo/releases), [PostgreSQL version policy](https://www.postgresql.org/support/versioning/), [pgx changelog](https://github.com/jackc/pgx/blob/master/CHANGELOG.md), [Redis release notes](https://redis.io/docs/latest/operate/oss_and_stack/stack-with-enterprise/release-notes/redisce/), and [Asynq releases](https://github.com/hibiken/asynq/releases) — foundation support and patch selection.
- [PostgreSQL row security](https://www.postgresql.org/docs/current/ddl-rowsecurity.html) and [CREATE POLICY](https://www.postgresql.org/docs/current/sql-createpolicy.html) — RLS behavior, owner/bypass exceptions, and policy caveats.
- [AWS transactional outbox guidance](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html) — dual-write failure and idempotent consumers.
- [NATS JetStream documentation](https://docs.nats.io/learn/jetstream/) and [consumer documentation](https://docs.nats.io/nats-concepts/jetstream/consumers) — persistence, replay, acknowledgements, and at-least-once behavior.
- [ClickHouse materialized views](https://clickhouse.com/docs/concepts/features/materialized-views/incremental-materialized-view), [primary-key guidance](https://clickhouse.com/docs/concepts/best-practices/choosing-a-primary-key), and [insert deduplication](https://clickhouse.com/docs/concepts/features/operations/insert/deduplicating-inserts-on-retries) — projections, ordering, and retry limits.
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/), [context propagation](https://opentelemetry.io/docs/concepts/context-propagation/), and [W3C Trace Context](https://www.w3.org/TR/trace-context/) — instrumentation, propagation, and sensitive-data constraints.
- [Cloudflare hostname validation](https://developers.cloudflare.com/cloudflare-for-platforms/cloudflare-for-saas/domain-support/hostname-validation/) and [zero-downtime migration](https://developers.cloudflare.com/cloudflare-for-platforms/cloudflare-for-saas/domain-support/hostname-validation/zero-downtime-migration/) — managed hostname/TLS lifecycle evidence; provider portability remains to be validated.
- [MDN cookie guidance](https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/Cookies), [Stripe idempotent requests](https://docs.stripe.com/api/idempotent_requests), [GitHub webhook best practices](https://docs.github.com/en/webhooks/using-webhooks/best-practices-for-using-webhooks), and [OWASP SSRF prevention](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html) — identity limits, API retry semantics, delivery design, and safe egress.
- [European Commission GDPR principles](https://commission.europa.eu/law/law-topic/data-protection/information-business-and-organisations/principles-gdpr_en) — minimization and storage limitation for EU scope; launch obligations require jurisdiction-specific review.

### Primary Product/Category Sources (HIGH confidence)

- [Bitly product overview](https://support.bitly.com/hc/en-us/articles/230895688-What-is-Bitly) and [analytics metrics](https://support.bitly.com/hc/en-us/articles/20370474672141-What-metrics-are-available-in-Bitly) — link-management and analytics expectations.
- [Rebrandly create-link flow](https://developers.rebrandly.com/docs/get-started), [workspaces](https://developers.rebrandly.com/docs/workspaces), [tags](https://developers.rebrandly.com/docs/tags), and [advanced routing](https://developers.rebrandly.com/docs/advanced-link-options) — workspace-scoped links and extension features.
- [Branch attribution documentation](https://help.branch.io/docs/attribution-page-new) — attribution windows and downstream events.
- [Dub Conversions](https://dub.co/blog/introducing-dub-conversions) — public category evidence for click-to-lead-to-sale and revenue attribution; not used as an implementation blueprint.
- [PostHog identity documentation](https://posthog.com/docs/data/persons) — anonymous and identified event linkage patterns.

### Secondary or Conditional Evidence (MEDIUM confidence)

- Exact adoption thresholds for ClickHouse, Redis separation, and more distributed topology — require Flux-specific benchmarks and operational objectives.
- Final custom-domain lifecycle portability beyond the evaluated provider documentation — depends on the selected managed provider.
- Bun patch selection and frontend supporting-library choices — current recommendations should be revalidated during implementation.
- V1 versus v1.x ordering of custom domains and shipping both attribution models — strong roadmap judgment, pending design-partner validation.

---
*Research completed: 2026-10-05*
*Ready for roadmap: yes*
