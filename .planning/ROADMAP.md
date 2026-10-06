# Roadmap: Flux

## Overview

Flux v1 progresses from a stable brownfield foundation to one complete, trustworthy managed-domain attribution journey. The sequence first fixes process, lifecycle, contract, test, and observability boundaries; then adds tenant-safe link management, an independent redirect path, durable click tracking, deterministic conversion attribution, and finally the accessible analytics and setup experience that proves the whole loop under retries, outages, cache invalidation, and tenant-isolation checks. Existing code is treated as scaffolding until each phase's observable behavior is verified.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions created after planning

- [ ] **Phase 1: Foundation Stability & System Boundaries** - Make the existing service foundation reproducible, independently operable, observable, and safe to extend.
- [ ] **Phase 2: Tenant-Safe Link Control Plane** - Give authenticated teams isolated workspaces and complete managed-domain link management.
- [ ] **Phase 3: Fast Managed-Domain Redirects** - Serve correct, cache-backed redirects independently of analytics and control-plane failures.
- [ ] **Phase 4: Durable Click Tracking & Privacy** - Capture, enrich, identify, and replay click facts within an explicit privacy and loss contract.
- [ ] **Phase 5: Idempotent Conversion Attribution** - Accept durable conversion events and produce deterministic, explainable first- and last-click decisions.
- [ ] **Phase 6: Trusted Analytics & Complete Setup Journey** - Let a marketer complete and inspect the full link-to-revenue loop in an accessible analytics product.

## Phase Details

### Phase 1: Foundation Stability & System Boundaries
**Goal:** Operators and developers have a boring, reliable foundation with final process and architecture boundaries for every later product slice.
**Mode:** mvp
**Depends on:** Nothing (first phase)
**Requirements:** PLAT-01, PLAT-02, PLAT-03, PLAT-04, PLAT-05, PLAT-06, PLAT-07, PLAT-08, SAFE-01, SAFE-08
**Success Criteria** (what must be TRUE):
  1. An operator can start API, redirector, worker, and migrator roles independently, apply migrations as an explicit release step, and observe that an API replica never starts workers or migrations implicitly.
  2. Each long-running role reports separate sanitized liveness and dependency-backed readiness, handles SIGTERM with bounded draining, and releases every owned resource even after partial startup or cleanup failures.
  3. A developer can reproduce the pinned local/CI environment, and CI fails on formatting, linting, type checking, generated-contract drift, migration failure, tests, builds, dependency findings, or secret findings.
  4. A developer can change the versioned API contract and regenerate consistent Go, TypeScript, OpenAPI, and served documentation artifacts without working-directory-dependent runtime failures.
  5. An operator can correlate requests and asynchronous work through redacted structured logs, correlation IDs, traces, and bounded-cardinality metrics without exposing credentials, event secrets, or personal data.
**Plans:** 6/22 plans executed
**UI hint:** no — this phase has operator endpoints and API documentation, but no user-facing product UI.

### Phase 2: Tenant-Safe Link Control Plane
**Goal:** Authenticated users can manage teams and managed-domain links inside enforceable workspace boundaries.
**Mode:** mvp
**Depends on:** Phase 1
**Requirements:** TEN-01, TEN-02, TEN-03, TEN-04, TEN-05, TEN-06, TEN-07, TEN-08, LINK-01, LINK-02, LINK-03, LINK-05, LINK-06, LINK-07, LINK-08, LINK-09, SAFE-02, SAFE-03, SAFE-04
**Success Criteria** (what must be TRUE):
  1. A user can sign in, create a workspace as its owner, and switch among their workspaces without seeing another workspace's records, identifiers, caches, jobs, or analytical facts.
  2. Owners and admins can invite members, accept unexpired invitations, list members, change permitted roles, and remove members while role enforcement prevents an ownerless workspace.
  3. A permitted member can create a managed-domain link with a generated or available custom key, while unsafe destinations, normalized collisions, duplicate retries, and conflicting concurrent changes are rejected safely.
  4. A member can inspect, search, filter, and paginate links and perform allowed lifecycle transitions; an authorized administrator can suspend abuse with a recorded actor, reason, and time.
  5. Secure sessions, CSRF controls, workspace-scoped parameterized data access, explicit transactions, destination/egress defenses, and two-workspace denial tests protect every delivered control-plane workflow.
**Plans:** TBD
**UI hint:** yes

### Phase 3: Fast Managed-Domain Redirects
**Goal:** Visitors receive correct managed-domain redirects through an independently operable, latency-bounded data plane.
**Mode:** mvp
**Depends on:** Phase 2
**Requirements:** LINK-04, REDIR-01, REDIR-02, REDIR-03, REDIR-04, REDIR-05, REDIR-06, REDIR-07, REDIR-08
**Success Criteria** (what must be TRUE):
  1. A visitor opening an active managed-domain short URL receives the documented redirect, while missing, disabled, archived, deleted, suspended, or unsafe links return documented non-redirect responses.
  2. Redis cache-aside resolution and PostgreSQL fallback use one canonical host/path identity, and version-aware outbox invalidation makes link edits visible within the bounded interval without stale messages restoring an older destination.
  3. Redirects continue to work when analytics storage, consumers, webhooks, or email fail, and trusted-proxy configuration prevents arbitrary forwarding headers from controlling client identity.
  4. Representative load tests demonstrate cached p95 below 30 ms and uncached p95 below 100 ms in the documented environment, with telemetry exposing latency, traffic, cache, fallback, error, and event-handoff outcomes.
**Plans:** TBD

### Phase 4: Durable Click Tracking & Privacy
**Goal:** Every eligible redirect can produce a privacy-conscious click fact that survives and recovers within a measured durability contract without delaying the visitor.
**Mode:** mvp
**Depends on:** Phase 3
**Requirements:** TRK-01, TRK-02, TRK-03, TRK-04, TRK-05, TRK-06, TRK-07, TRK-08, TRK-09, SAFE-05
**Success Criteria** (what must be TRUE):
  1. An eligible redirect creates one versioned, workspace-scoped click fact with a stable event ID and can hand the visitor a documented namespaced click ID without waiting for enrichment or analytics persistence.
  2. Failure-injection tests demonstrate the stated click-loss SLO across crash, broker outage, disk pressure, shutdown, recovery, duplicate delivery, poison events, and replay.
  3. Consumers acknowledge only committed facts, deduplicate by producer event ID, and asynchronously derive privacy-conscious referrer, UTM, geography, device, browser/OS, language, and bot dimensions.
  4. Flux assigns documented anonymous visitor and session identities, permits explicit customer association, and leaves missing or ambiguous evidence unresolved rather than relying on invasive fingerprinting or guesswork.
  5. Operators can observe lag, ingestion, duplicates, rejection, poison events, replay, and freshness, while the documented raw-IP lifecycle prevents indefinite retention before click data ships.
**Plans:** TBD

### Phase 5: Idempotent Conversion Attribution
**Goal:** Customer systems can submit conversions safely and receive deterministic, auditable attribution to eligible clicks.
**Mode:** mvp
**Depends on:** Phase 4
**Requirements:** EVT-01, EVT-02, EVT-03, EVT-04, EVT-05, EVT-06, EVT-07, ATTR-01, ATTR-02, ATTR-03, ATTR-04, ATTR-05, ATTR-06, ATTR-07, SAFE-07
**Success Criteria** (what must be TRUE):
  1. An authenticated workspace integration can submit versioned lead and purchase events with supported identity evidence, metadata, event time, and integer minor-unit revenue, receiving predictable errors and actionable rate-limit metadata for invalid or excessive requests.
  2. A caller can retry an identical tenant-scoped idempotent request and receive its original durable outcome, while reusing the key for a different request returns a deterministic conflict and never duplicates business effects.
  3. First-click and last-click attribution select eligible touchpoints deterministically by event time and stable tie-breakers, then persist the model, policy, window, evidence, credited click/link, reason, and decision time.
  4. Retries, equal timestamps, late facts, identity merges, policy changes, currency, and replay follow tested recomputation/immutability rules; ineligible or ambiguous conversions remain visible with an exact unattributed reason.
  5. Security-relevant membership, identity, link-suspension, role, and integration actions produce workspace-scoped audit records with actor, action, resource, time, and safe metadata.
**Plans:** TBD

### Phase 6: Trusted Analytics & Complete Setup Journey
**Goal:** A marketer can complete the managed-domain click-to-conversion-to-attribution journey and trust the resulting analytics, freshness, and decision evidence.
**Mode:** mvp
**Depends on:** Phase 5
**Requirements:** ANLY-01, ANLY-02, ANLY-03, ANLY-04, ANLY-05, ANLY-06, ANLY-07, ANLY-08, ANLY-09, ANLY-10, SAFE-06
**Success Criteria** (what must be TRUE):
  1. A workspace member can view consistently defined clicks, human clicks, unique visitors, leads, purchases, conversion rate, revenue, and average order value over bounded time ranges, with link filters, top links, time series, and referrer/country/device/UTM breakdowns.
  2. A member can inspect the click, conversion, identity-association and processing states, integer revenue, credited or unattributed decision, and exact rejection or attribution reason, with bot handling applied consistently.
  3. The responsive, keyboard-usable dashboard communicates loading, empty, error, freshness, and delayed-processing states instead of presenting eventual data as complete.
  4. A new workspace member can follow the original setup flow to create and visit a test managed-domain link, submit the sample conversion idempotently, and see the attributed result despite representative retries, delayed consumers, restarts, cache invalidation, and two-workspace isolation checks.
  5. Operators can rebuild and reconcile analytical projections from durable facts without changing canonical links, conversions, or attribution decisions; storage adapters preserve metric semantics and retention/deletion prevents removed data from reappearing through caches, streams, retries, dead letters, exports, logs, or replay.
**Plans:** TBD
**UI hint:** yes

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation Stability & System Boundaries | 6/22 | In Progress|  |
| 2. Tenant-Safe Link Control Plane | 0/TBD | Not started | - |
| 3. Fast Managed-Domain Redirects | 0/TBD | Not started | - |
| 4. Durable Click Tracking & Privacy | 0/TBD | Not started | - |
| 5. Idempotent Conversion Attribution | 0/TBD | Not started | - |
| 6. Trusted Analytics & Complete Setup Journey | 0/TBD | Not started | - |
