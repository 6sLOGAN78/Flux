---
phase: 03-fast-managed-domain-redirects
researched: 2026-10-14
confidence: high
mode: mvp
---

# Phase 3 research

## Outcome and scope

As a visitor, I want an active managed short link to send me to its current safe destination quickly, so that control-plane changes are respected without making analytics part of my request. As an authorized member, I can edit destination/title and observe the committed version through that public path.

Sources: ROADMAP Phase 3, REQUIREMENTS LINK-04 and REDIR-01–08, spec.md sections 7 and 42–43, inspected production modules listed below. No Phase 3 CONTEXT exists. Choices below are planning assumptions selected under `--auto`, not user-locked decisions. Phase 2 UI/context decisions remain applicable to existing tenant workflows.

Planning is authorized before Phase 2 closes. Execution remains dependent on Phase 2 acceptance: 34/35 plans complete; 02-35 task 2 has no accepted full-provider proof. Latest full run passed signup verification and actual password login/Go `/me`, then stopped on a recovery heading locator. Six session cases passed separately; separate runs do not satisfy full-suite promotion. Do not send email, restart inbox tooling, or mark this prerequisite passed while planning.

## Verified brownfield seams

| Responsibility | Existing authority | Extension |
|---|---|---|
| Role graph, constructors, readiness, cleanup | apps/backend/internal/app/api.go, redirector.go, lifecycle.go | Redirector acquires PostgreSQL plus optional derived Redis, no Clerk/email/job producer. Existing Phase 1 no-DB assertions must change explicitly. |
| Role configuration | apps/backend/internal/config/config.go and config_test.go | RoleRedirector currently excludes database/Redis/links. Introduce its own normalized redirect policy, preserving dot-nested FLUX names and API cursor requirements. |
| Links and versions | repository/link.go, service/link.go; migrations 005–013 | Global unique managed_host/short_key; BIGINT version; existing scoped locking, mutation ledger and audit conventions. Next migration 014. |
| Safe destinations | safety/url.go | Reuse pure ValidateDestination at resolution and cache decode; no DNS, HTTP previews or destination fetch. |
| Authenticated transport | handler/product.go, router/router.go | Add PATCH detail route with current bearer/CSRF, membership and error behavior; keep public redirect route separate. |
| Canonical contracts | packages/zod/src/identity.ts; packages/openapi/src/contracts/identity.ts | Strict edit schema plus three generated outputs: two OpenAPI JSON copies and transport/health.gen.go. No handwritten competing DTOs. |
| Editing UI | components/link-detail.tsx, lib/links.ts, control-plane.spec.ts | Existing native forms, request scoping, immutable retry intent, lifecycle/suspension conflict behavior. |
| Integration discovery | scripts/check.ts integrationTests | Every new container-backed top-level test is classified explicitly, with positive execution counts and race detector retained. |
| Real dependency fixtures | internal/testing/container.go | SetupTestDB uses migrated pinned PostgreSQL17.11; SetupTestRedis uses pinned Redis8.10.2. Reuse, never mock Redis for scripting correctness. |
| Privacy/metrics | observability/telemetry.go, middleware/tracing.go | Explicit role providers and closed allowlists. Extend collector/redaction checks as well as producing instruments. |

Runtime manifest currently pins Go1.26.9, Bun1.3.14, Node22.23.3. Preserve existing versions; AGENTS stack snapshot predates implementation. No new runtime library is necessary.

## Public routing contract

- Default HTTP302 for GET and HEAD; `Cache-Control: no-store` on every public outcome. HEAD has the same status/Location and no body or click-handoff attempt. Preserve the configured destination's path/query/fragment; do not merge visitor query parameters.
- Match only the configured managed host using NormalizeLinkHost. Accept case/trailing-dot canonical equivalence; strip only a syntactically valid authority port before validation. Ignore Forwarded/X-Forwarded-Host for routing. Deployment proxies must preserve the canonical Host. Compare canonical host against configured managed host before lookup.
- One literal ASCII key segment, lowercase identity matching current create rules. Reject encoded path characters, escaped slashes/backslashes, repeated/trailing slash, dot segments, invalid escapes and reserved keys; never clean a hostile path into another link. An uppercase literal key resolves its lowercase canonical identity. IDNA applies to hosts, not keys.
- Public unavailable states (missing, disabled, archived, deleted, suspended, unsafe, invalid host/key) return identical plain404 with no Location or private details. Unsupported method405 with Allow GET, HEAD. An indeterminate dependency failure returns503, no Location, Retry-After:1. Health remains exact `/live` and `/ready`; API/docs stay absent on redirector.
- Lookup uses one parameterized PostgreSQL query on the global host/key unique index and a minimal routing projection. Never expose creator email, audit reason, membership data, workspace identity or provider information. PostgreSQL primary READ COMMITTED is the authority; no lagging replica reads for route freshness.
- Destination policy is checked even for a syntactically valid cached entry and when operator blocked-host policy changes. Fail closed on malformed/versionless/identity-mismatched/oversized cache content.

## Freshness contract and alternatives

Selected: Redis cache-aside routing projection plus an **authoritative primary PostgreSQL version/status guard on every cache hit**. Cache TTL is a fixed5,000ms with no sliding renewal. API mutation acknowledgment is the PostgreSQL commit; edits become observable on the next successful authoritative resolution, within the documented maximum five seconds for newly admitted requests. Missing/inactive/suspended/unsafe routes never use an old cached target. An already in-flight request linearizes at its authoritative primary read; no design can retract an HTTP response already sent.

Compare: plain DELETE invalidation permits a delayed SQL fill to resurrect an old value; TTL alone would allow disabled links to redirect during the stale interval and fails the literal REDIR-04 requirement. Nonrenewable fill leases bound this interval but do not eliminate it. A persistent version floor is also insufficient after eviction or Redis restoration. The selected primary guard preserves strict state correctness without claiming instantaneous distributed invalidation. It adds a database round trip to cache hits, so performance must be measured honestly. Cached means the destination projection came from Redis, not that PostgreSQL was bypassed.

Concrete algorithm (project inference; validate real dependencies):

1. Use `redirect:v1:{sha256(host + NUL + key)}` as the common cluster hash tag for value and version floor. Bound payload16KiB and concurrent resolutions128. A cache hit is a candidate only, never authority.
2. Atomically read payload/floor with positive PTTL <=5000; require schema1, exact normalized host/key, positive BIGINT decimal-string version, active/unsuspended state and a destination that passes current local safety policy. Reject malformed or impossible TTLs; no negative caching initially.
3. For a valid candidate, query the primary for this host/key's current version, lifecycle and suspension. Only an existing active unsuspended row with the exact candidate version authorizes that destination. If the version differs, fetch a fresh full routing projection (or use one full query initially if measured costs justify simplicity) within the original request budget. An inactive/missing row returns404 regardless of Redis. A SQL error returns503, even with a warm cache. Use fresh READ COMMITTED statements, never a long-lived transaction or lagging replica.
4. On cache miss or cache failure, query full current routing projection from primary, validate state/destination, and return it. Cache fill is optional and deadline-bounded. Reject a fill whose originating context is canceled; never let a rejected old fill become a returned target. A later read always repeats the authoritative guard.
5. Atomic Lua fill compares versions against both current payload and floor, rejects older candidates, and stores with PX5000. Atomic invalidation only advances the positive decimal-string floor and removes strictly older payload; retain floor at least6000ms. Older/equal events cannot restore a destination or renew payload TTL. No Lua tonumber or JavaScript Number for BIGINT comparisons. Redis restart, flush or restored older cache content cannot bypass the primary guard.

Timeout defaults: entire public resolution100ms, total Redis read/fill budget10ms, total primary SQL budget80ms, response deadline never refreshed on retry. A paused handler checks its original monotonic deadline before sending Location. Context-aware same-key miss coalescing is bounded128 flights; overload503, no unbounded map or goroutines. Redis retries disabled/bounded within shared budget. Readiness depends on primary PG and intake; Redis/relay backlog/telemetry/email do not gate global availability. Both warm and cold redirects fail closed503 during primary outage. This intentionally sacrifices PG-outage serving to satisfy strict REDIR-04; REDIR-05 only requires independence from analytics, consumers, webhooks and email. Document this distinction explicitly.

## Transactional outbox and independent publication

Migration014 creates `redirect_outbox` with immutable identity/version, unique workspace+link+version, availability/claim token/lease deadline/attempt count and created timestamp. Use an AFTER INSERT OR UPDATE trigger on links; UPDATE only emits when version changes and rejects route-affecting changes that fail to advance version. No destination, actor email, private audit reason or browser token in payload. Rollback also rolls back outbox; idempotent replays do not issue a new mutation. Hard deletion remains unsupported and canonical keys remain reserved.

Run a bounded outbox relay within each **redirector** runtime, independent of existing email workers. This uses the role's existing PG/Redis ownership, avoids making Resend/Clerk a cache-invalidation dependency, and permits multiple redirector replicas. A short transaction claims <=100 rows using FOR UPDATE SKIP LOCKED and a random lease token, then commits before Redis I/O. Lease5s, poll250ms, capped exponential retry to30s. Completion updates/deletes only with the current token; crash after invalidation before ack is harmless replay. Never delete on Redis error. Old poison records retain retryable durable state and emit closed-category errors; no silent drop. Purge only acknowledged rows in bounded batches if retained; no unbounded completed history. Unprocessed rows remain recoverable through outages; alert on age/backlog and document storage growth during extended outage.

Publisher only invalidates, never restores a destination from event contents. It need not read private link fields. Default role credentials may update only outbox lease state in addition to SELECT routing projection; no public mutation route exists. The primary guard preserves correctness while relay is down; the TTL bounds retained derived data. Startup Redis failure degrades to PG-only routing and retries relay asynchronously. Extend cleanup with relay stop/join before closing clients, preserving the single shared shutdown deadline.

## Authorized edit contract

PATCH `/api/v1/workspaces/:workspaceId/links/:linkId`, strict JSON `{destination,title,expectedVersion}` and Idempotency-Key; full replacement of these two fields, title may be empty but <=200 characters under existing title semantics. expectedVersion is the existing positive BIGINT decimal-string format. Preserve host/key/id/workspace/lifecycle/suspension. Owner/admin/member current `CapabilityWrite` capability may edit active or disabled unsuspended links; archived/deleted/suspended produce409. Fresh membership authorization precedes replay. Follow workspace -> actor -> link locking and existing durable mutation ledger; immutable actor/operation/target audit entry, one increment, atomic outbox trigger. A no-op successful edit consumes one canonical version consistently, so retry and cache behavior have no hidden branch. Reused key with differing normalized intent409; stale version409; unknown fields400; unsafe destination400 under existing typed error conventions. Cross-workspace UUIDs remain nondisclosing404 and cause zero writes.

UI edit is an inline labeled form in detail, Save changes/Cancel, latest-version conflict review and immutable retry after unknown result; drafts never cross workspace/session boundaries. Reuse existing style/native focus patterns. The API remains the authority. 03-UI-SPEC records exact states and copy.

## Trust, privacy and observability

Set Echo v4 IPExtractor explicitly. Direct mode trusts only socket peer. Optional XFF mode accepts headers only from operator-configured exact CIDRs, disables built-in private/loopback/link-local trust, walks right-to-left to first untrusted hop, rejects malformed/oversized chains, and ignores X-Real-IP/Forwarded. Trust no all-address CIDR by default. Never derive routing host, scheme or authorization from an arbitrary forwarding header. Admission safety is process-bounded; do not claim a global per-client abuse quota or persist raw addresses in this phase.

Metrics: fixed instruments for requests, seconds histogram, cache lookup result, PG fallback result, errors, invalidation outcome/backlog age and handoff outcome. Closed attribute values only; no host/path/key/destination/workspace/user/IP/header labels. Route template is fixed, not incoming path. Extend application, stdout and collector allowlists together. Production handoff outcome is honestly `not_configured` for successful GET: Phase4 owns durable click facts and its loss contract. Do not invent accepted/delivered events, analytics writes, raw IP collection or an unbounded goroutine/channel. A later producer must replace this explicit boundary under Phase4 tests. HEAD/unavailable requests do not report accepted handoff.

## Validation architecture

Use real migrated PostgreSQL and pinned Redis integration groups in `internal/app` plus existing `TestProductActualHTTP` subtests. Each added top-level integration test is registered in scripts/check.ts, not silently discovered as a unit. Race-enabled root discovery remains authoritative. Units exercise canonicalization, decimal versions, strict schema/input and proxy extraction without containers. Local Playwright uses signed provider fixtures and real Go API; it must not send email or count as Phase2 live acceptance.

Acceptance matrix: active GET/HEAD; canonical identity hostility; every lifecycle/suspension state warm/cold; unsafe stored destinations and malformed cache; cross-tenant edit; stale/replayed/conflicting/concurrent edits; transaction/audit/outbox rollback; stale fill vs newer invalidation and mandatory primary guard; reverse/duplicate messages including >2^53 versions; no hit/message TTL extension; Redis flush/restart; SQL timeout; Redis-only outage; PG-only outage warm/cold; both unavailable; relay crash/reclaim/multiple instances; dead email/analytics/telemetry; shutdown with in-flight relay; proxy spoofing; fixed-label privacy; full browser create/edit/redirect.

Performance proof must start the actual redirector transport with real PG/Redis, disable HTTP auto-follow, and exercise hot-cache and distinct-key uncached workloads. Record runtime/image pins, OS/CPU/RAM, process resources, dataset, concurrency32, load generator location, warmup5s, steady30s, cached200rps and uncached50rps initial profiles, actual sample counts >=1000 per profile, failures and cache classifications. Use actual server-duration histogram with explicit buckets including29ms/99ms and conservative percentile upper bounds; require cached p95<30ms and uncached<100ms. Include response serialization scope in the measured handler duration and explain excluded network/client costs. Failed/too-small/skipped samples never pass. Tune only with published before/after evidence and unchanged thresholds. Report cached and uncached separately, not an aggregate. Local results are scoped to the documented environment and do not establish an unmeasured production SLO.

## Official sources (retrieved 2026-10-14)

- [PostgreSQL17 SELECT locking](https://www.postgresql.org/docs/17/sql-select.html): SKIP LOCKED is suitable for contended queue-like reads, not a replacement for ordinary consistent reads.
- [PostgreSQL17 trigger semantics](https://www.postgresql.org/docs/17/trigger-definition.html): triggers execute in the triggering transaction; rollback is shared.
- [Redis scripting](https://redis.io/docs/latest/develop/programmability/eval-intro/): atomic execution, volatile script cache; keep scripts bounded and support NOSCRIPT reload.
- [Redis PTTL](https://redis.io/docs/latest/commands/pttl/): millisecond remaining lifetime and missing/no-expiry distinctions.
- [Redis expiry](https://redis.io/docs/latest/develop/using-commands/keyspace/): expiry timestamps and clock dependence.
- [Echo IP extraction](https://echo.labstack.com/next/guide/ip-address/): explicitly documents v4's unsafe legacy default and configured extractor/trust options. Verify signatures against the repository's pinned v4 source, not v5 examples.
- [Go net/http Redirect](https://pkg.go.dev/net/http#Redirect): default body/header behavior; set no-store and HEAD handling explicitly.

The algorithm, response matrix, budgets and decomposition above are project design inferences based on these primitives, not guarantees furnished by the sources. Resolved research risk: a TTL-only lease would weaken REDIR-04. The selected primary guard avoids that reduction. Remaining measured risk is cached-path latency with the mandatory PG check; do not remove that guard to pass a benchmark.
