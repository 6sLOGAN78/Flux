# Requirements: Flux

**Defined:** 2026-10-05
**Core Value:** A marketer can create a link, send a visitor through a fast and reliable redirect, record the click asynchronously, attribute a later conversion and revenue to that click, and see the result in analytics.

## v1 Requirements

Requirements for the first production-credible release. The managed Flux domain is sufficient for v1; every requirement below must map to exactly one roadmap phase.

### Platform Foundation

- [ ] **PLAT-01**: An operator can start the API, redirector, worker, and migrator as independently configured process roles without an API replica implicitly starting workers or migrations.
- [ ] **PLAT-02**: An operator can apply deterministic, versioned PostgreSQL migrations as a release step and verify forward migration from an empty database in CI.
- [ ] **PLAT-03**: An operator can distinguish process liveness from dependency-backed readiness through separate health endpoints that return sanitized public diagnostics.
- [ ] **PLAT-04**: An operator can terminate API, redirector, and worker processes with SIGTERM and observe bounded draining and cleanup of all owned resources.
- [x] **PLAT-05**: A developer can reproduce the supported toolchain and service dependencies locally and in CI from pinned project configuration.
- [ ] **PLAT-06**: A developer receives a failed CI result when formatting, linting, type checking, code generation, migrations, unit tests, integration tests, or builds drift or fail.
- [ ] **PLAT-07**: An operator can correlate requests and asynchronous work through structured logs, request/correlation IDs, traces, and bounded-cardinality metrics without leaking secrets or raw sensitive values.
- [ ] **PLAT-08**: A developer can update the versioned API contract and regenerate checked Go and TypeScript boundaries, with CI detecting stale generated artifacts.

### Identity and Workspaces

- [ ] **TEN-01**: A user can sign up or sign in through the selected identity provider and Flux maps that identity to an internal user record.
- [ ] **TEN-02**: An authenticated user can create a workspace and becomes its owner atomically.
- [ ] **TEN-03**: A workspace owner or admin can invite a member, and the invitee can accept a valid unexpired invitation.
- [ ] **TEN-04**: A workspace owner or admin can list members, change allowed roles, and remove members without leaving a workspace ownerless.
- [ ] **TEN-05**: Workspace roles owner, admin, member, and viewer enforce documented permissions at the server-side service boundary.
- [ ] **TEN-06**: A user can switch among workspaces they belong to without exposing data, identifiers, cache entries, jobs, or analytics from another workspace.
- [ ] **TEN-07**: Every workspace-owned database record, event, idempotency record, cache key, and analytical fact contains or unambiguously resolves to a workspace identifier.
- [ ] **TEN-08**: Automated two-workspace denial tests prove that direct IDs, listing filters, background work, and analytics queries cannot cross tenant boundaries.

### Link Management

- [ ] **LINK-01**: A workspace member with permission can create a link on the managed Flux domain using a generated collision-resistant short key and a validated HTTP(S) destination.
- [ ] **LINK-02**: A workspace member can request an available custom short key, while PostgreSQL uniqueness prevents normalized domain-and-key collisions under concurrency.
- [ ] **LINK-03**: A workspace member can view a link with its destination, title, lifecycle state, creator, timestamps, and stable public short URL.
- [ ] **LINK-04**: A workspace member can update a link destination and title, and the new canonical version becomes visible to redirects within a bounded invalidation interval.
- [ ] **LINK-05**: A workspace member can enable, disable, archive, restore, and soft-delete a link according to documented lifecycle transitions.
- [ ] **LINK-06**: A workspace member can list links with cursor pagination and filter or search by short key, title, destination, and lifecycle state.
- [ ] **LINK-07**: Link mutations use durable idempotency or conflict handling so retries cannot create duplicate links or silently overwrite concurrent changes.
- [ ] **LINK-08**: Link destination validation rejects unsafe schemes, malformed URLs, embedded credentials, and destinations disallowed by the abuse policy.
- [ ] **LINK-09**: Authorized workspace administrators can suspend an abusive link and record the actor, reason, and time of the action.

### Redirect Data Plane

- [ ] **REDIR-01**: A visitor opening an active managed-domain short URL receives the configured destination redirect using one documented default redirect status.
- [ ] **REDIR-02**: The redirector resolves canonical normalized host and path through Redis cache-aside lookup with PostgreSQL fallback, while PostgreSQL remains authoritative.
- [ ] **REDIR-03**: Link changes publish transactional outbox records that drive version-aware cache invalidation without permitting stale messages to restore older destinations.
- [ ] **REDIR-04**: Missing, disabled, archived, deleted, or unsafe links return documented responses and never redirect to a stale cached destination.
- [ ] **REDIR-05**: Redirects remain available when analytics storage, analytics consumers, webhooks, or email infrastructure are unavailable.
- [ ] **REDIR-06**: Representative load tests verify cached redirect processing p95 below 30 ms and uncached processing p95 below 100 ms under a documented environment and concurrency level.
- [ ] **REDIR-07**: Redirect telemetry reports request rate, latency, cache hit ratio, cache-fallback outcomes, error rate, and event-handoff outcomes without unbounded labels.
- [ ] **REDIR-08**: Trusted-proxy and client-address handling is explicitly configured so abuse controls and privacy processing do not trust arbitrary forwarding headers.

### Click Tracking and Identity

- [ ] **TRK-01**: Every eligible redirect creates one versioned click fact with a stable event ID, event time, workspace, link, request correlation, and minimal routing context.
- [ ] **TRK-02**: The redirector hands click facts to a tested durability mechanism without waiting for enrichment or analytics persistence before responding.
- [ ] **TRK-03**: The project defines a measurable click-loss SLO and verifies crash, broker-outage, disk-pressure, shutdown, and recovery behavior for the selected local handoff or bounded-loss design.
- [ ] **TRK-04**: Tracking consumers tolerate at-least-once delivery by preserving producer event IDs, committing dedupe with durable writes, and acknowledging only after commit.
- [ ] **TRK-05**: Click processing derives privacy-conscious referrer domain, UTM fields, coarse geography, device class, browser/OS family, language, and bot classification asynchronously.
- [ ] **TRK-06**: Flux assigns documented anonymous visitor and session identifiers without using invasive fingerprinting as the primary identity mechanism.
- [ ] **TRK-07**: A click ID can be handed to an opted-in destination through a documented, namespaced mechanism that a customer backend can return with a conversion.
- [ ] **TRK-08**: A customer identity can be associated with earlier anonymous evidence explicitly, while missing or ambiguous evidence remains unattributed instead of being guessed.
- [ ] **TRK-09**: Operators can observe queue lag, ingestion rate, duplicates, rejected events, poison events, replay progress, and data freshness.

### Conversion Events

- [ ] **EVT-01**: An authenticated workspace integration can submit versioned `lead` and `purchase` events through a server API.
- [ ] **EVT-02**: A conversion event can include an external event ID, click ID, customer ID, anonymous ID, event time, metadata, and for purchases an integer minor-unit amount with ISO currency.
- [ ] **EVT-03**: A caller can safely retry an event using a tenant-scoped idempotency key and receives the original outcome for an identical request.
- [ ] **EVT-04**: Reusing an idempotency key with a different request hash returns a deterministic conflict without altering the original event.
- [ ] **EVT-05**: Event validation returns a predictable versioned error envelope with code, safe message, field details where appropriate, and request ID.
- [ ] **EVT-06**: Accepted events are durable before success is returned and can be processed asynchronously without duplicate business outcomes.
- [ ] **EVT-07**: Rate limits and abuse controls protect event ingestion and expose actionable rate-limit metadata to callers.

### Attribution

- [ ] **ATTR-01**: A workspace can use a documented first-click or last-click attribution model with one documented default lookback window.
- [ ] **ATTR-02**: Attribution selects eligible touchpoints deterministically by event time with stable tie-breakers, independent of consumer arrival order.
- [ ] **ATTR-03**: A persisted attribution decision records the conversion, credited click, link, model, policy version, attribution window, evaluated evidence, decision time, and reason.
- [ ] **ATTR-04**: Reprocessing the same facts and policy produces the same decision without duplicating attributed conversions or revenue.
- [ ] **ATTR-05**: Late clicks, late conversions, identity merges, retries, and policy changes follow documented recomputation and immutability rules.
- [ ] **ATTR-06**: A conversion with missing, expired, ambiguous, bot-only, or otherwise ineligible evidence is retained with a visible unattributed reason.
- [ ] **ATTR-07**: Unit and integration tests cover attribution-window boundaries, equal timestamps, event-time ordering, identity association, duplicates, late data, currency, and tenant isolation.

### Analytics and Product UI

- [ ] **ANLY-01**: A workspace member can view clicks, human clicks, unique visitors, leads, purchases, conversion rate, revenue, and average order value using centrally documented metric definitions.
- [ ] **ANLY-02**: A workspace member can view time-series results and top links for a bounded date range and filter results by link.
- [ ] **ANLY-03**: A workspace member can break down results by referrer domain, country, device class, and UTM source, medium, and campaign.
- [ ] **ANLY-04**: Analytics consistently exclude or separate bot traffic according to the metric dictionary.
- [ ] **ANLY-05**: A workspace member can inspect a click, conversion, identity-association state, processing state, revenue, attribution decision, and unattributed or rejection reason.
- [ ] **ANLY-06**: The dashboard shows data freshness and delayed-processing states rather than presenting eventual data as complete.
- [ ] **ANLY-07**: A new workspace member can follow an original setup flow to create a test link, visit it, send a sample idempotent conversion, and see the attributed result.
- [ ] **ANLY-08**: Analytical projections can be rebuilt and reconciled from durable facts without changing canonical links, conversions, or attribution decisions.
- [ ] **ANLY-09**: Analytics query and storage adapters keep API metric semantics stable if measured load later requires migration from PostgreSQL to ClickHouse.
- [ ] **ANLY-10**: The web application is responsive, keyboard-usable, accessible, and includes clear loading, empty, error, and eventual-consistency states for v1 workflows.

### Security and Privacy

- [ ] **SAFE-01**: Authentication credentials, session tokens, API credentials, event payload secrets, and personal data are redacted from logs, traces, metrics, errors, and generated documentation.
- [ ] **SAFE-02**: Session handling uses secure cookie and CSRF protections appropriate to the chosen authentication flow, with expiry and rotation behavior tested.
- [ ] **SAFE-03**: All database access uses parameterized SQL, explicit transaction boundaries, and workspace-scoped queries or constraints.
- [ ] **SAFE-04**: Any outbound URL processing resolves and connects with SSRF defenses against loopback, private, link-local, metadata, and disallowed redirect destinations.
- [ ] **SAFE-05**: Raw IP addresses are not retained indefinitely; the collection, derivation, retention, and deletion policy is documented before click data ships.
- [ ] **SAFE-06**: Retention or deletion rules cover primary rows, analytics projections, caches, event streams, retries, dead letters, exports, logs, and replay suppression so deleted data is not silently resurrected.
- [ ] **SAFE-07**: Security-relevant membership, role, link suspension, identity, and integration actions produce tenant-scoped audit records with actor, action, resource, time, and safe metadata.
- [ ] **SAFE-08**: Dependency and secret scanning run in CI and produce actionable failures without publishing discovered secret values.

## v2 Requirements

Deferred until the v1 attribution loop is reliable under retries, outages, cache invalidation, and tenant-isolation tests. These requirements are tracked but are not part of the initial roadmap.

### Custom Domains and Link Productivity

- **DOM-01**: Workspace administrators can connect custom domains through separate ownership, DNS, certificate, activation, failure, and disablement states using managed TLS infrastructure.
- **DOM-02**: Verified custom domains route links through the same canonical redirect and tracking pipeline as the managed domain.
- **PROD-01**: Members can organize links with campaigns, tags, and folders without conflating organizational folders with marketing campaigns.
- **PROD-02**: Members can build UTM parameters, generate PNG/SVG QR codes, and perform authorized bulk link operations.
- **PROD-03**: Members can import and export links with row-level validation and resumable error reporting.
- **PROD-04**: Members can configure expiration, password protection, and fallback destinations after redirect rule precedence is defined.

### Developer Platform and Integrations

- **DEV-01**: Workspace administrators can create, scope, expire, rotate, and revoke hashed API keys with prefix identification and last-used metadata.
- **DEV-02**: Developers can automate links and analytics through stable public APIs with OpenAPI, cursor pagination, predictable errors, idempotency, and rate-limit metadata.
- **DEV-03**: Developers can subscribe to signed, filtered webhooks with stable delivery IDs, bounded retries, delivery history, manual resend, and automatic disablement policy.
- **DEV-04**: A lightweight browser SDK preserves first-party click and anonymous identity and sends trusted non-monetary events asynchronously.
- **DEV-05**: Go and TypeScript SDKs provide thin generated or contract-backed clients for supported public endpoints.
- **INT-01**: A Stripe adapter verifies provider events and translates purchase facts into the canonical idempotent Events API model.

### Partner, Billing, and Enterprise Expansion

- **PART-01**: Workspaces can operate partner programs with applications, referral links, attribution, integer commissions, lifecycle controls, and a tenant-isolated partner portal.
- **BILL-01**: Flux can enforce subscriptions, usage, limits, and retention through centralized entitlements rather than scattered plan-name checks.
- **ENT-01**: Enterprise workspaces can add SAML/OIDC SSO, SCIM, advanced RBAC, configurable retention, data residency, and expanded audit/export controls.
- **ADV-01**: Link owners can configure advanced geo, device, OS, language, A/B, deep-link, and crawler routing after a separate precedence and cache-safety design.
- **ATTR-08**: Workspaces can choose additional attribution models only after clean-data volume and customer decision use cases justify them.

## Out of Scope

Explicit exclusions for the initial roadmap.

| Feature | Reason |
|---------|--------|
| Copying Dub code, branding, protected assets, or pixel-exact design | Flux must be an original product; public references inform expectations only |
| Kafka | JetStream is sufficient for the planned replayable event stream without Kafka's initial operational cost |
| Kubernetes | Deployment requirements do not yet justify cluster complexity |
| Elasticsearch | PostgreSQL search covers initial link discovery needs |
| ClickHouse as a day-one dependency | PostgreSQL can validate metrics and the product loop; benchmark evidence must justify migration |
| Dozens of microservices | API, redirector, worker, and migrator process roles provide needed isolation within a modular codebase |
| Bespoke certificate authority or TLS automation | Custom domains will use managed provider or proxy certificate automation |
| Machine-learning or opaque attribution | Explainable deterministic first-/last-click models are required before advanced models |
| Invasive fingerprinting | Conflicts with the privacy-conscious identity strategy |
| Synchronous analytics, enrichment, webhooks, or email on redirect | Violates redirect availability and latency requirements |
| Floating-point money | Revenue and commissions require integer minor units with explicit currency |
| Broad integration catalog, affiliate platform, billing, and enterprise governance in v1 | These depend on validated attribution, identity, money, and contract semantics |

## Definition of Done

The v1 milestone is complete when an isolated workspace can execute this journey under representative retries, delayed consumers, process restarts, and cache invalidation:

1. A member creates a managed-domain link.
2. A visitor receives an immediate redirect and a documented click ID.
3. The click becomes durable within the stated loss SLO and is enriched asynchronously.
4. A customer backend submits an idempotent lead or purchase.
5. Flux associates available identity evidence and applies first- or last-click attribution.
6. The marketer sees the click, conversion, integer revenue, freshness, and exact attribution reason.

Completion also requires passing unit, integration, migration, API, two-workspace isolation, failure-injection, performance, and critical browser-flow checks; updated operator and integration documentation; and no open critical security finding in the delivered scope.

## Traceability

Every v1 requirement maps to exactly one roadmap phase. v2 requirements are intentionally excluded.

| Requirement | Phase | Status |
|-------------|-------|--------|
| PLAT-01 | Phase 1 | Pending |
| PLAT-02 | Phase 1 | Pending |
| PLAT-03 | Phase 1 | Pending |
| PLAT-04 | Phase 1 | Pending |
| PLAT-05 | Phase 1 | Complete |
| PLAT-06 | Phase 1 | Pending |
| PLAT-07 | Phase 1 | Pending |
| PLAT-08 | Phase 1 | Pending |
| TEN-01 | Phase 2 | Pending |
| TEN-02 | Phase 2 | Pending |
| TEN-03 | Phase 2 | Pending |
| TEN-04 | Phase 2 | Pending |
| TEN-05 | Phase 2 | Pending |
| TEN-06 | Phase 2 | Pending |
| TEN-07 | Phase 2 | Pending |
| TEN-08 | Phase 2 | Pending |
| LINK-01 | Phase 2 | Pending |
| LINK-02 | Phase 2 | Pending |
| LINK-03 | Phase 2 | Pending |
| LINK-04 | Phase 3 | Pending |
| LINK-05 | Phase 2 | Pending |
| LINK-06 | Phase 2 | Pending |
| LINK-07 | Phase 2 | Pending |
| LINK-08 | Phase 2 | Pending |
| LINK-09 | Phase 2 | Pending |
| REDIR-01 | Phase 3 | Pending |
| REDIR-02 | Phase 3 | Pending |
| REDIR-03 | Phase 3 | Pending |
| REDIR-04 | Phase 3 | Pending |
| REDIR-05 | Phase 3 | Pending |
| REDIR-06 | Phase 3 | Pending |
| REDIR-07 | Phase 3 | Pending |
| REDIR-08 | Phase 3 | Pending |
| TRK-01 | Phase 4 | Pending |
| TRK-02 | Phase 4 | Pending |
| TRK-03 | Phase 4 | Pending |
| TRK-04 | Phase 4 | Pending |
| TRK-05 | Phase 4 | Pending |
| TRK-06 | Phase 4 | Pending |
| TRK-07 | Phase 4 | Pending |
| TRK-08 | Phase 4 | Pending |
| TRK-09 | Phase 4 | Pending |
| EVT-01 | Phase 5 | Pending |
| EVT-02 | Phase 5 | Pending |
| EVT-03 | Phase 5 | Pending |
| EVT-04 | Phase 5 | Pending |
| EVT-05 | Phase 5 | Pending |
| EVT-06 | Phase 5 | Pending |
| EVT-07 | Phase 5 | Pending |
| ATTR-01 | Phase 5 | Pending |
| ATTR-02 | Phase 5 | Pending |
| ATTR-03 | Phase 5 | Pending |
| ATTR-04 | Phase 5 | Pending |
| ATTR-05 | Phase 5 | Pending |
| ATTR-06 | Phase 5 | Pending |
| ATTR-07 | Phase 5 | Pending |
| ANLY-01 | Phase 6 | Pending |
| ANLY-02 | Phase 6 | Pending |
| ANLY-03 | Phase 6 | Pending |
| ANLY-04 | Phase 6 | Pending |
| ANLY-05 | Phase 6 | Pending |
| ANLY-06 | Phase 6 | Pending |
| ANLY-07 | Phase 6 | Pending |
| ANLY-08 | Phase 6 | Pending |
| ANLY-09 | Phase 6 | Pending |
| ANLY-10 | Phase 6 | Pending |
| SAFE-01 | Phase 1 | Pending |
| SAFE-02 | Phase 2 | Pending |
| SAFE-03 | Phase 2 | Pending |
| SAFE-04 | Phase 2 | Pending |
| SAFE-05 | Phase 4 | Pending |
| SAFE-06 | Phase 6 | Pending |
| SAFE-07 | Phase 5 | Pending |
| SAFE-08 | Phase 1 | Pending |

**Coverage:**
- v1 requirements: 74 total
- Mapped to phases: 74
- Unmapped: 0 ✓

---
*Requirements defined: 2026-10-05*
*Last updated: 2026-10-05 after roadmap creation*
