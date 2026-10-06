# Pitfalls Research

**Domain:** Production-grade multi-tenant link attribution and marketing analytics SaaS
**Researched:** 2026-10-05
**Confidence:** HIGH for tenant isolation, retry/idempotency, cache consistency, webhook security, and custom-domain lifecycle; MEDIUM for the final click-durability mechanism and identity behavior because those depend on deployment storage, browser context, and an explicit loss SLO

## Critical Pitfalls

### Pitfall 1: Calling Click Capture “Durable” While Keeping It Only In Memory

**What goes wrong:**
The redirect returns successfully, but the click vanishes if the redirector crashes, receives `SIGKILL`, loses its node, or exhausts an in-memory channel before the broker accepts the event. Waiting synchronously for JetStream or analytical storage avoids that gap but makes redirect latency and availability depend on remote analytics infrastructure. A broker publish acknowledgement protects the message only after the broker has received it; it does not protect the interval between returning the HTTP response and publishing.

**Why it happens:**
“Asynchronous” is mistaken for “durable,” and broker guarantees are applied to a message the broker has not yet accepted. The requirements “analytics never blocks a redirect” and “every click is durable” conceal a product/SRE decision about acceptable loss, local persistence latency, disk ownership, and what happens when buffering reaches capacity.

**How to avoid:**
Define a measurable click-loss SLO before implementation. For a strict target, create the immutable event before the response, append it to a bounded local durable spool, make the response depend only on local spool acceptance, publish in the background, and delete only after broker acknowledgement. Specify `fsync`/group-commit behavior, per-replica spool ownership, restart recovery, disk-full behavior, shutdown draining, readiness thresholds, and whether traffic is shed or clicks may be dropped at capacity. If the deployment cannot provide recoverable local storage, document a bounded-loss mode honestly and measure it with failure injection. Do not call an in-memory queue durable.

**Warning signs:**
- Redirect p95 rises when NATS or the analytics sink slows.
- Click totals fall after redirector restarts, deploys, or broker outages.
- No metric exists for spool depth, oldest event age, dropped events, or publish acknowledgements.
- The shutdown path closes PostgreSQL/Redis before draining event producers; the current scaffold already has incomplete shutdown ordering and only listens for `SIGINT`.
- Readiness remains green as the buffer approaches capacity, or marks the redirector unready solely because the remote broker is temporarily unavailable despite usable spool capacity.

**Phase to address:**
Foundation must fix lifecycle/readiness and process separation; Redirect Data Plane must establish the latency baseline; Durable Tracking must run the required spool-versus-bounded-loss spike and failure-injection benchmark before claiming durable clicks.

---

### Pitfall 2: Tenant Scope Exists in HTTP but Disappears in Storage, Workers, or Analytics

**What goes wrong:**
A handler checks workspace membership, then calls an unscoped repository, cache, export, event consumer, or analytics query that can read or mutate another workspace’s data. Background workers and API keys bypass assumptions embedded in browser middleware. Shared external IDs, idempotency keys, aggregate tables, or cache keys collide across tenants. This is the highest-impact class of failure because a correct UI can hide a cross-tenant data leak.

**Why it happens:**
Teams treat tenancy as authentication context instead of a data invariant. Greenfield product tables and workers are added faster than isolation tests. PostgreSQL RLS is sometimes treated as automatic protection even though table owners and `BYPASSRLS` roles bypass policies, and pooled tenant context can leak if set at session scope.

**How to avoid:**
Put `workspace_id` on every tenant-owned row or require an explicit ownership join. Require workspace scope in repository/query method signatures and authorization in application use cases. Scope uniqueness and idempotency constraints deliberately, propagate tenant identity in versioned events, and validate it again at each consumer. If RLS is adopted, use a non-owner runtime role, consider `FORCE ROW LEVEL SECURITY`, set tenant context transaction-locally, and test owner, worker, migration, API-key, export, and analytics paths. Build a two-workspace isolation matrix covering CRUD, search, pagination, aggregates, caches, jobs, and webhooks.

**Warning signs:**
- Repository methods named `GetByID(id)` or `List(filter)` for tenant-owned records.
- SQL or ClickHouse queries lack a mandatory workspace predicate.
- Global cache keys use only object ID, external event ID, or idempotency key.
- Worker payloads infer a workspace from a mutable object after dequeue.
- Tests use one workspace, or RLS tests run as the table owner.

**Phase to address:**
Identity + Workspace Kernel, before the first product table. Every later phase must extend the isolation matrix as it introduces a new store, consumer, export, or public endpoint.

---

### Pitfall 3: Redirect Cache State Drifts From Canonical Link State

**What goes wrong:**
Deleted, disabled, expired, rerouted, or transferred links keep redirecting from stale cache entries. A negative cache hides a newly created link. Concurrent creators win an application pre-check and produce a domain/key collision. Different code paths normalize hosts or paths differently, creating ambiguous lookups, cache poisoning, or links that work only on cache hits.

**Why it happens:**
Redis is treated as authority, updates and invalidations are separate best-effort writes, and normalization is scattered between API, database, proxy, and redirector. Teams overlook the failure window between committing PostgreSQL and publishing invalidation.

**How to avoid:**
Define one canonical host/path normalization contract and test case, Unicode/IDNA, trailing-dot, port, slash, percent-encoding, and proxy-host behavior. Enforce the canonical `(host, key/path)` uniqueness constraint in PostgreSQL. Cache a versioned immutable redirect snapshot, use a transactional outbox for link/domain changes, and retain bounded positive and shorter negative TTLs as safety nets. Treat Redis errors as misses and protect PostgreSQL fallback with separate pool budgets, deadlines, negative caching, and abuse load shedding. Test stale invalidation, out-of-order updates, cold cache, Redis loss, and concurrent link creation.

**Warning signs:**
- Link updates call `UPDATE`, then `DEL` Redis without an outbox.
- The API and redirector each own a slightly different normalization helper.
- Cache entries contain no schema/link version or use an unbounded TTL.
- Application-level “does key exist?” checks are the only collision defense.
- Cache misses or random-key scans saturate the shared PostgreSQL pool.

**Phase to address:**
Core Link Platform defines canonical identity and database constraints; Redirect Data Plane adds cache-aside behavior, versioned invalidation, outage tests, and latency/load tests.

---

### Pitfall 4: Custom-Domain Ownership, DNS Routing, and TLS Are Collapsed Into One Boolean

**What goes wrong:**
A hostname appears “verified” while its certificate is pending or its DNS still points elsewhere; traffic cuts over to TLS errors; a hostname can be attached to two workspaces; a deleted tenant leaves a dangling mapping that another party can claim; or certificate renewal failure silently breaks every link on the domain.

**Why it happens:**
DNS is eventually consistent, and providers expose separate ownership-verification, certificate-validation, routing, and certificate states. A single `verified` flag cannot represent this lifecycle. Verification tokens are also sometimes reusable, predictable, or accepted after the domain resource has changed owners.

**How to avoid:**
Model ownership, DNS target, certificate/DCV, activation, failure reason, and last check separately. Require globally unique normalized hostnames and workspace-bound high-entropy verification challenges with expiry/rotation. Pre-validate ownership and pre-issue TLS where the provider supports it; activate redirect routing only when the required states are ready. Use provider-managed certificate automation, monitor renewal/expiry and DNS drift, preserve audit history, and define safe detach/reclaim behavior. Re-check ownership when DNS changes materially or a disabled hostname is reactivated.

**Warning signs:**
- One `verified_at` column gates both routing and HTTPS.
- The UI says “active” before both hostname and certificate states are active.
- A domain can be soft-deleted in one workspace and immediately claimed in another without a quarantine/ownership check.
- No alert exists for renewal failure, DNS drift, or domains stuck in pending states.
- Tests assume DNS changes are instantaneous.

**Phase to address:**
Custom Domains, before promising branded production links. Include provider sandbox/contract tests and a zero-downtime cutover test.

---

### Pitfall 5: Visitor Identity Is Presented as Ground Truth

**What goes wrong:**
Unique-visitor counts and conversion matches split one person across browsers/devices or merge different people on shared devices. A redirect on `go.brand.example` cannot set a cookie for an unrelated destination domain, and browser cookie/SameSite rules change which identifiers travel in later requests. Teams compensate with fingerprinting, producing privacy risk and still-fragile identity.

**Why it happens:**
Anonymous cookie IDs, sessions, click IDs, customer IDs, and identity merges are collapsed into one `visitor_id`. Cross-domain and cross-device limitations are not made explicit in the model or UI.

**How to avoid:**
Use separate opaque click, anonymous visitor, session, and known-customer identifiers with provenance and timestamps. Prefer a short-lived click ID passed to the destination and captured by a first-party SDK/server integration; keep host-only cookies where possible and treat cookies as one evidence source. Make `identify` merges explicit, idempotent, auditable, reversible where required, and scoped to a workspace. Define exactly what “unique visitor” means. Accept an unattributed result when evidence is missing; do not use invasive fingerprinting as a fallback.

**Warning signs:**
- The design assumes one cookie is readable across customer custom domains and unrelated destination sites.
- `visitor_id` is deterministic from IP/user-agent or used indefinitely.
- Unique counts change substantially with bot filtering or cookie settings but the UI provides no definition.
- Identity merge has no provenance, conflict rule, or test for shared devices/cross-workspace IDs.

**Phase to address:**
Durable Tracking defines click/session identifiers and privacy bounds; Conversion Attribution implements first-party handoff, identify semantics, and explicit confidence/unattributed outcomes.

---

### Pitfall 6: Attribution Changes With Retry Timing, Arrival Order, or a Later Code Deploy

**What goes wrong:**
The same conversion is credited twice, switches from one click to another after a replay, uses ingestion time instead of occurrence time, crosses a workspace/currency boundary, or cannot explain why a link received credit. Late clicks and conversions make first/last-click results nondeterministic. Model changes silently rewrite historical revenue.

**Why it happens:**
Canonical conversion acceptance, identity resolution, attribution decisions, and analytical projections are implemented as one mutable worker. Idempotency exists only in Redis or only at the HTTP edge. Model version, window, candidate set, and event-time policy are not persisted.

**How to avoid:**
Persist the canonical conversion in PostgreSQL first, reserving a tenant-scoped idempotency key plus request hash; reject the same key with different parameters. Decide on event-time, allowed lateness, clock-skew, identity precedence, attribution-window boundaries, currency rules, and tie-breakers. Implement first/last-click selection as a deterministic pure function. Persist conversion ID, credited click/link/campaign, model and policy version, window, inputs/reason, and integer minor-unit amount/currency. Define whether reprocessing creates a superseding decision or leaves history immutable. Project only committed decisions to analytics.

**Warning signs:**
- Attribution queries use `ORDER BY created_at` with no stable tie-breaker.
- A consumer updates counters before persisting the canonical decision.
- Replaying the same fixture changes credit or revenue totals.
- The same idempotency key with a changed body is silently accepted.
- Historical reports change after deploy with no reattribution record.

**Phase to address:**
Conversion Events + Attribution, after durable click identity exists and before revenue dashboards/webhooks depend on attribution.

---

### Pitfall 7: At-Least-Once Delivery Inflates Analytics

**What goes wrong:**
Broker redelivery, worker restart, publish retry, manual replay, or an uncertain ClickHouse insert duplicates clicks, conversions, aggregates, and webhooks. Acknowledging before the durable sink commits instead loses data. ClickHouse insert deduplication is mistaken for a permanent unique constraint even though its deduplication window is finite and settings/version behavior matters.

**Why it happens:**
“Exactly once” is inferred from transport features. Event IDs are generated again on retry, deduplication is applied to batches rather than business events, and raw facts and materialized aggregates have different retry behavior.

**How to avoid:**
Assign a stable event ID at the producer and keep it through every retry/replay. Make each consumer idempotent at its durable side-effect boundary and acknowledge only after commit. Maintain an event-ID dedupe strategy whose retention covers the replay horizon; test partial-batch success, consumer crash after write/before ack, and replay into projections. Treat ClickHouse retry deduplication as an optimization with configured windows/tokens, not the sole business invariant. Reconcile broker counts, raw facts, aggregates, and canonical PostgreSQL decisions.

**Warning signs:**
- Consumers acknowledge on receipt or before the database batch commits.
- Retried events receive new IDs.
- Click/revenue totals jump after replay or worker recovery.
- No reconciliation query compares accepted event IDs with raw and aggregate projections.
- ClickHouse deduplication defaults are assumed rather than pinned and tested for the deployed version.

**Phase to address:**
Durable Tracking establishes envelope IDs, ack discipline, dedupe, poison-event handling, and replay tests; Analytics validates raw-to-aggregate reconciliation and storage-version settings.

---

### Pitfall 8: Analytics Semantics Are Undefined Even When Queries Are Fast

**What goes wrong:**
Dashboard cards disagree because “click,” “unique visitor,” “conversion rate,” bot inclusion, timezone boundaries, and late-event handling differ by endpoint. Enrichment creates a second row and double counts the original. Tags/campaigns are joined using current state, rewriting historical reports. Near-real-time views silently omit late or replayed facts.

**Why it happens:**
Teams optimize storage and charts before defining metric contracts and fact grain. Mutable dimensions and eventual consistency are hidden from users.

**How to avoid:**
Write a metric dictionary before dashboard implementation: fact grain, numerator/denominator, bot/preview policy, uniqueness window, event-time timezone, late-arrival policy, attribution basis, and freshness. Preserve immutable raw facts and versioned enrichment; choose whether provisional rows are visible and ensure enrichment replaces or joins without double counting. Snapshot historically significant dimensions or explicitly label reports as current-state joins. Return `data_as_of`/freshness and use one analytics query service for API and UI. Keep raw facts long enough to rebuild projections.

**Warning signs:**
- API endpoints calculate the same metric independently.
- Totals change when a link title, tag, or campaign changes.
- Enriched-event counts exceed original event IDs.
- There is no “as of” time, lag metric, or late-event test.
- Product copy says “real time” while consumer lag is unbounded.

**Phase to address:**
Analytics Contract + Dashboard, with metric fixtures and replay/rebuild acceptance tests before visual polish.

---

### Pitfall 9: API and Webhook Retries Cause Duplicate Effects or Become an SSRF Channel

**What goes wrong:**
Clients retry an uncertain `POST` and create two links/conversions; webhook recipients receive duplicate money-related events; signatures are calculated over re-serialized JSON and fail verification; old signed requests can be replayed; or an attacker configures a webhook/preview URL that reaches localhost, private networks, cloud metadata, or a public hostname that later DNS-rebinds internally.

**Why it happens:**
Idempotency is treated as an in-memory cache, webhooks are treated as ordinary HTTP calls, and URL validation checks only syntax or the first DNS answer. Retry state, delivery identity, raw signed bytes, timeout policy, and secret rotation are added late.

**How to avoid:**
Store tenant/endpoint-scoped idempotency keys durably with a normalized request hash and stable replayed response. Give every webhook event and delivery stable IDs; sign the exact raw bytes with a versioned HMAC scheme including timestamp, support overlapping secrets during rotation, and document replay tolerance. Persist attempts, response metadata, next retry, and terminal state; use exponential backoff with jitter and manual resend using the same event identity. For outbound requests, allow only HTTP(S), reject credentials and unsafe ports as policy requires, resolve and validate every IPv4/IPv6 destination at connection time, block private/link-local/loopback/metadata ranges, disable or revalidate redirects, and enforce egress controls, byte limits, and deadlines.

**Warning signs:**
- Idempotency lives only in Redis or is not bound to request content/workspace.
- Manual resend creates a new event identity.
- Webhook signatures cover a parsed/re-encoded object rather than raw bytes.
- HTTP clients follow redirects automatically or validate DNS separately from the connection.
- Delivery workers have unrestricted internal network access or no timeout/body limit.

**Phase to address:**
Developer Platform + Webhooks. The shared outbound-fetch security adapter should be introduced earlier if Social Preview metadata fetching ships in Campaign Productivity.

---

### Pitfall 10: Privacy Controls Stop at the Primary Event Table

**What goes wrong:**
Raw IP addresses, full destination/referrer query strings, customer IDs, or stable visitor IDs leak into logs, traces, queues, ClickHouse, dead-letter payloads, exports, and backups. A deletion or retention job removes one copy while projections and replays recreate it. High-cardinality identifiers become metric labels, increasing cost and disclosure risk.

**Why it happens:**
Data minimization and retention are deferred as “enterprise hardening,” while the initial event schema becomes a permanent contract. Derived geography/device data is kept alongside the raw input indefinitely because enrichment and deletion ownership are unclear.

**How to avoid:**
Classify every tracking field and record its purpose, lawful/consent dependency where applicable, precision, and retention before collection. Derive only needed network metadata, then discard or tightly expire raw IP; strip/redact sensitive URL parameters; never place PII in trace context or metric labels. Give raw events, enriched facts, broker retention, dead letters, logs, exports, and backups explicit lifecycles. Design workspace/visitor deletion and consent-aware suppression across canonical facts and projections, with tombstones or replay filters so deleted data is not resurrected. Validate requirements with counsel for launch jurisdictions rather than encoding one jurisdiction’s rules in domain logic.

**Warning signs:**
- “We will add retention later” while raw IP and full URLs are already immutable event fields.
- Deletion tests cover PostgreSQL but not Redis, broker, analytical store, logs, exports, or replay.
- Request/event/customer IDs appear as Prometheus-style labels.
- Redacted UI/API responses coexist with unredacted structured logs.

**Phase to address:**
Tracking Schema + Enrichment, before collecting production traffic; Analytics and Production Hardening must complete retention/deletion propagation and jurisdiction-specific review before broader launch.

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Run API, redirector duties, and workers in one process | Fewer binaries | Deploys, health checks, resource pools, and failures become coupled; redirect independence is false | Local development only, never the production topology |
| In-memory click channel | Minimal tracking code | Accepted clicks disappear on crash/deploy; no outage buffer | Only in an explicitly labeled bounded-loss prototype with measured loss |
| Database write followed by best-effort publish/cache delete | Easy happy path | Missing events and permanently stale redirect cache after crashes | Never for canonical changes with required downstream effects; use an outbox |
| Redis-only idempotency/deduplication | Fast and simple | Key eviction or outage duplicates durable business effects | Acceleration only, backed by a durable invariant |
| One generic `metadata` JSON object for all event fields | Schema flexibility | Unindexable contracts, PII creep, silent type drift | For bounded optional metadata with size/depth/key controls |
| Current-state joins for all historical dimensions | Fewer fact columns | Historical campaign/tag/link reports rewrite themselves | Only when the UI explicitly promises current-state grouping |
| One `verified` flag for custom domains | Simple UI/schema | Cannot distinguish ownership, DNS, certificate, and routing failure | Never beyond a throwaway mock |
| Store raw IP/full URL “for future analysis” | Maximum future options | Privacy, breach, deletion, and storage burden become architectural | Never without a documented purpose and short retention |
| Adopt ClickHouse before metric contracts and event volume evidence | Impressive ingest benchmark | Two databases, migrations, replay, and correctness problems arrive early | When analytics-phase benchmarks show PostgreSQL contention/retention need |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| NATS JetStream | Treat broker acceptance or dedupe as end-to-end exactly-once | Stable producer IDs, publish acknowledgement, explicit consumer ack after commit, sink idempotency, replay tests |
| ClickHouse | Assume `ORDER BY` is uniqueness or retry dedupe is permanent | Explicit event-ID strategy, pinned settings, configured dedupe window/token, raw/aggregate reconciliation |
| Redis | Serve unversioned redirect state as authority | PostgreSQL authority, cache-aside snapshots, outbox invalidation, bounded positive/negative TTLs |
| DNS/TLS provider | Treat hostname verification as certificate/routing readiness | Track ownership, DNS, TLS/DCV, and activation independently; pre-validate and monitor renewal |
| GeoIP/UA data | Enrich on redirect path or assume classifications never change | Enrich asynchronously, record database/parser version, keep raw-retention bounds, support unknown values |
| Identity provider | Treat provider organization claims as complete workspace authorization | Map verified identity to internal user; keep memberships/roles authoritative in PostgreSQL |
| Stripe/other conversions | Map provider payloads directly into attribution tables | Verify provider signatures, dedupe provider event ID, translate through an integration boundary into canonical conversion events |
| Webhook destinations | Use default HTTP client behavior | Isolated egress, SSRF-safe resolver/dialer, redirect revalidation, strict time/body limits, signed stable deliveries |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Remote broker/storage on redirect critical path | Redirect p95/p99 tracks broker or database incidents | Local handoff with explicit durability policy; never enrich or deliver webhooks inline | First broker slowdown or regional network incident |
| Redis outage sends every request to shared PostgreSQL pool | Pool saturation takes down API and redirector together | Separate budgets, short timeouts, negative caching, load shedding, cache-hit alerts | Cold cache, crawler/random-key burst, or Redis failover |
| Per-event analytical inserts | High CPU/part count and low ingest throughput | Bounded-latency batching with ack after batch commit | Well before “big data”; visible as event rate rises and parts/transactions multiply |
| Unbounded analytics dimensions/cardinality | Slow scans, memory blowups, expensive telemetry | Metric/query contracts, tenant/time-led sort keys, dimension limits, no customer IDs as metric labels | At sustained event retention, not a specific user count |
| Query-time exact uniqueness across raw history | Dashboard timeouts for “unique visitors” | Defined uniqueness window and tested aggregate/sketch strategy where appropriate | When retained raw rows outgrow interactive scan budget |
| Synchronous bot/geo/preview work | Redirect latency and third-party failure coupling | Minimal redirect event; asynchronous versioned enrichment | First external-service slowdown |
| Unbounded webhook retries | Queue starvation and endpoint amplification | Per-endpoint concurrency, jittered backoff, max attempts, terminal state, disable policy | One persistently failing or malicious endpoint |

Scale changes should be triggered by measured event rate, retained bytes, cache hit ratio, fallback QPS, broker lag, replay time, and dashboard p95 rather than arbitrary customer counts.

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Trusting `Host`/client IP from any proxy | Cross-tenant routing, spoofed rate-limit identity, poisoned logs | Explicit trusted-proxy topology; reject unknown hosts; normalize once |
| Weak destination URL policy | Phishing, credential-bearing URLs, unsafe schemes, brand/reputation damage | Parse canonically; allow HTTP(S); reject credentials/ambiguous hosts; abuse states, quotas, review and suspension workflow |
| Server-side fetching of destinations/webhooks without connection-time IP checks | SSRF, metadata/internal-service access, DNS rebinding | Validate all resolved IPv4/IPv6 addresses, bind to validated address, revalidate redirects, restrict egress |
| Global or guessable domain verification token | Hostname hijack across tenants | High-entropy workspace/domain-bound expiring challenge and globally unique normalized hostname |
| API keys stored reversibly or logged | Workspace takeover | Prefix for lookup, strong secret hash, scopes/environment/expiry, rotation, redaction and last-used audit |
| Webhook HMAC without timestamp/delivery identity/rotation | Replay and brittle secret changes | Versioned signature over raw body plus timestamp; constant-time compare; replay window; overlapping secrets |
| RLS used with owner/BYPASSRLS runtime role | Silent cross-tenant access | Non-owner app role, `FORCE RLS` where chosen, transaction-local context, explicit service authorization and isolation tests |
| Destination/referrer query strings in logs/events | Token and personal-data disclosure | Redaction/allowlisted parameters before logging and event emission |
| Preview bots counted and allowed to trigger conversion-like effects | Inflated metrics and unintended one-time-link behavior | Classify preview/crawler traffic; define eligibility; make preview handling side-effect safe |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Analytics shown without freshness or metric definitions | Users interpret eventual, deduping data as final | Show `data_as_of`, processing lag/status, and accessible metric definitions |
| Domain screen says only pending/verified | Users cannot diagnose DNS versus certificate failures | Separate ownership, DNS routing, TLS, and activation states with exact next action |
| “Unique visitors” implies people | Marketers over-trust probabilistic browser identity | Define the identifier/window and state cross-device/browser limits |
| Link edit appears immediately successful while old cache still routes | Users send traffic to stale destinations | Confirm canonical commit, expose propagation state only if material, keep invalidation SLO short and monitored |
| Unattributed conversions are hidden | Dashboard looks cleaner but loses trust and debugging value | Show attributed/unattributed totals and machine-readable reason codes |
| Webhook retry UI lacks stable delivery IDs and payload versions | Developers cannot reconcile duplicates or support incidents | Persistent delivery log, attempt history, signature version, event ID, manual resend |
| Abuse suspension returns generic failure | Legitimate owners cannot remediate; visitors see unsafe detail | Safe public fallback plus authenticated reason, review status, and appeal workflow |

## “Looks Done But Isn’t” Checklist

- [ ] **Tenant isolation:** CRUD works in one workspace — verify two-workspace denial across repositories, caches, workers, search, analytics, exports, and API keys.
- [ ] **Redirect independence:** Broker outage still redirects — also verify PostgreSQL/Redis degradation, spool capacity, disk full, restart recovery, `SIGTERM`, and in-flight drain.
- [ ] **Durable clicks:** Events arrive normally — verify crash after response/before remote publish and consumer crash after sink commit/before ack.
- [ ] **Cache correctness:** Updates invalidate in the happy path — verify lost/out-of-order invalidation, negative-cache creation race, cold cache, and schema-version mismatch.
- [ ] **Custom domains:** DNS verifies — verify certificate readiness/renewal, DNS drift, duplicate claim, detach/reclaim, and zero-downtime cutover.
- [ ] **Visitor identity:** Cookie is set — verify unrelated destination domains, browser privacy settings, shared devices, identity merge conflicts, and unattributed fallback.
- [ ] **Conversions:** Endpoint returns `202` — verify durable idempotency, changed-payload conflict, late/out-of-order events, stable replay, money/currency validation.
- [ ] **Analytics:** Charts render — verify one metric dictionary, bot policy, time zones, late data, freshness, raw/aggregate reconciliation, and projection rebuild.
- [ ] **Webhooks:** Signed delivery succeeds — verify raw-body signature vectors, rotation overlap, replay rejection, SSRF/DNS rebinding, bounded retries, and manual resend identity.
- [ ] **Privacy:** Raw IP is absent from the main row — verify logs, traces, broker, dead letters, analytical projections, exports, backups, retention and deletion replay.

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Lost clicks from in-memory handoff | HIGH; facts cannot be reconstructed reliably | Quantify loss from request logs if privacy policy permits, mark dashboard gap, fix handoff/SLO, run failure tests; do not fabricate events |
| Cross-tenant exposure | CRITICAL/HIGH | Contain affected endpoints/jobs, preserve audit evidence, rotate exposed credentials, identify all stores/exports, correct scope and isolation tests, follow incident/legal process |
| Stale redirect cache | MEDIUM | Purge/version affected keys, replay outbox, compare cache version to canonical link version, add reconciliation/lag alert |
| Duplicate events/aggregates | MEDIUM-HIGH | Freeze projection updates, identify stable event IDs, rebuild projections from canonical retained facts, add sink-level idempotency/reconciliation |
| Wrong historical attribution | HIGH | Preserve original decisions, deploy versioned policy, produce superseding decisions with audit reason, rebuild downstream projections and disclose restatement |
| Custom-domain takeover or TLS failure | HIGH | Disable routing, preserve ownership history, revoke/detach provider hostname/certificate, re-verify owner, restore with monitored cutover |
| Webhook duplicate/replay | MEDIUM-HIGH | Identify delivery/event IDs, stop retries, notify affected endpoint owners if effects matter, add idempotency/replay window, resend only stable original events |
| Excess retained tracking data | HIGH | Stop collection, inventory all copies, apply deletion/retention jobs and replay filters, validate backups/export policy, document the gap and prevention |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Coupled lifecycle defeats redirect independence | Foundation | Independent API/redirector/worker construction; process-specific readiness; `SIGTERM` drain and failure-path tests |
| Tenant scope leakage | Identity + Workspace Kernel | Two-workspace isolation matrix across SQL, cache, workers, APIs, analytics and exports; non-owner RLS test if adopted |
| Collision and normalization mismatch | Core Link Platform | Concurrent-create test plus canonical host/path corpus and database uniqueness constraint |
| Stale/authoritative Redis redirect state | Redirect Data Plane | Redis outage/cold-cache/invalidation-loss tests; cache-version and fallback-pool metrics; latency SLO load test |
| Domain ownership/TLS state collapse | Custom Domains | Provider lifecycle contract tests, duplicate-claim denial, DNS drift/renewal alerts, cutover test |
| Redirect independence versus click durability | Durable Tracking | Required platform-storage spike; kill-after-response test; broker-outage/spool-full recovery; measured p95 overhead and loss SLO |
| At-least-once duplicates | Durable Tracking | Crash-after-write-before-ack, partial-batch, poison-event and full-replay tests with stable counts |
| Undefined/double-counted analytics | Analytics Contract + Dashboard | Metric dictionary fixtures; event-time/late/bot/timezone tests; raw-to-aggregate reconciliation and rebuild |
| Fragile visitor identity and nondeterministic attribution | Conversion Events + Attribution | Cross-domain handoff tests; idempotency conflict; late/out-of-order replay; persisted model/reason and stable decision fixture |
| API/webhook retry and SSRF flaws | Developer Platform + Webhooks | Durable request-hash idempotency; signature vectors/rotation/replay; DNS-rebinding/redirect/private-IP egress tests |
| Data minimization and deletion gaps | Tracking Schema, then Production Hardening | Field inventory/retention schedule; cross-store deletion and replay-resurrection test; launch-jurisdiction review |

## Research Flags for Roadmap

- **Durable Tracking requires a spike before planning:** choose the acceptable loss SLO, deployment disk model, local spool technology/ownership, group-commit policy, and capacity fallback. The architecture cannot promise both remote independence and zero loss without resolving this boundary.
- **RLS requires a focused design only if adopted:** verify runtime roles, `FORCE ROW LEVEL SECURITY`, pooled transaction-local tenant context, migration ownership, worker access, backups, and referential-integrity side channels.
- **Analytics needs a benchmark at phase entry:** decide PostgreSQL versus ClickHouse from event volume, retention, query shapes, replay time, and transactional contention; pin ClickHouse version/settings if selected.
- **Identity/enrichment needs a privacy review before schema freeze:** document cookie/click-ID flow, raw IP lifecycle, bot policy, data purposes, retention, consent signals, and deletion behavior.
- **Custom domains need provider-specific validation:** exact ownership, certificate issuance/renewal, rate limits, migration, and hostname-reclaim behavior depend on the chosen managed provider.

## Sources

Primary and official sources consulted:

- [PostgreSQL: Row Security Policies](https://www.postgresql.org/docs/current/ddl-rowsecurity.html) — RLS default behavior, owner and `BYPASSRLS` exceptions, policy semantics. **Confidence: HIGH.**
- [PostgreSQL: CREATE POLICY](https://www.postgresql.org/docs/current/sql-createpolicy.html) — `USING`/`WITH CHECK`, constraint and referential-integrity caveats. **Confidence: HIGH.**
- [AWS Prescriptive Guidance: Transactional Outbox](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html) — database/message dual-write failure, ordering, and idempotent consumer requirement. **Confidence: HIGH.**
- [NATS JetStream documentation](https://docs.nats.io/learn/jetstream/) and [pull-consumer acknowledgement guidance](https://docs.nats.io/learn/jetstream/pull-consumers) — persistence, explicit acknowledgement, batching, consumer behavior, and recovery boundaries. **Confidence: HIGH for accepted messages; MEDIUM for the project’s pre-publish durability design.**
- [ClickHouse: Deduplicating inserts on retries](https://clickhouse.com/docs/concepts/features/operations/insert/deduplicating-inserts-on-retries) — uncertain insert status, finite deduplication windows, tokens, and materialized-view behavior. **Confidence: HIGH; exact settings must be pinned to the deployed version.**
- [Cloudflare for SaaS: Hostname validation](https://developers.cloudflare.com/cloudflare-for-platforms/cloudflare-for-saas/domain-support/hostname-validation/) and [zero-downtime migration](https://developers.cloudflare.com/cloudflare-for-platforms/cloudflare-for-saas/domain-support/hostname-validation/zero-downtime-migration/) — separate ownership/certificate states, pre-validation, and activation requirements. **Confidence: HIGH for Cloudflare; MEDIUM for portability to a provider not yet selected.**
- [MDN: Set-Cookie](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Set-Cookie) and [Using HTTP cookies](https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/Cookies) — domain/public-suffix constraints, host-only cookies, SameSite behavior. **Confidence: HIGH.**
- [Stripe: Idempotent requests](https://docs.stripe.com/api/idempotent_requests) — replaying the first result and rejecting key reuse with changed parameters. **Confidence: HIGH as a mature API pattern.**
- [GitHub: Validating webhook deliveries](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries) and [webhook best practices](https://docs.github.com/en/webhooks/using-webhooks/best-practices-for-using-webhooks) — HMAC over payload bytes, stable delivery identity, asynchronous processing, replay protection, and redelivery. **Confidence: HIGH as a public webhook pattern.**
- [OWASP: SSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html) — connection-time IP validation, IPv4/IPv6 private-range blocking, DNS rebinding, redirect controls, and egress defense. **Confidence: HIGH.**
- [European Commission: Principles of the GDPR](https://commission.europa.eu/law/law-topic/data-protection/information-business-and-organisations/principles-gdpr_en) — purpose limitation, data minimization, storage limitation, and privacy by design/default. **Confidence: HIGH for EU scope; launch obligations still require jurisdiction-specific review.**
- [W3C Trace Context](https://www.w3.org/TR/trace-context/) — prohibition on PII/sensitive information in trace fields and correlation/abuse considerations. **Confidence: HIGH.**
- Project evidence: `.planning/PROJECT.md`, `spec.md`, `.planning/codebase/CONCERNS.md`, and `.planning/research/ARCHITECTURE.md`. **Confidence: HIGH for current state and intended constraints.**

---
*Pitfalls research for: Flux link attribution and marketing analytics platform*
*Researched: 2026-10-05*
