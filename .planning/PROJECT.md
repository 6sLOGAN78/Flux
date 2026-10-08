# Flux

## What This Is

Flux is the temporary internal codename for a production-grade, multi-tenant link attribution and marketing analytics SaaS. It gives marketers, growth teams, developers, agencies, and affiliate teams branded short links, reliable redirects, asynchronous click tracking, conversion attribution, analytics, developer APIs, and later partner and billing capabilities through an original Go-first implementation.

The repository is brownfield, with a verified Go/Echo foundation and TypeScript packages for API contracts and email templates. API, redirector, worker, and migrator roles operate independently, with deterministic generation, embedded assets, sanitized health and observability, bounded lifecycle handling, and executable local/CI checks. Product domains, the product database schema, frontend, and redirect behavior remain assigned to later phases.

## Core Value

A marketer can create a link, send a visitor through a fast and reliable redirect, record the click asynchronously, attribute a later conversion and revenue to that click, and see the result in analytics.

## Requirements

### Validated

- ✓ The existing Go service starts from a layered Echo-based HTTP foundation with configuration, structured logging, middleware, health and API-documentation routes — existing source foundation
- ✓ PostgreSQL, Redis, migrations, and Asynq worker integrations have reusable initialization scaffolding — existing source foundation
- ✓ TypeScript workspaces provide reusable Zod schemas, ts-rest/OpenAPI generation, and React Email templates — existing source foundation
- ✓ Independently configured process roles, explicit migrations, liveness/readiness, bounded shutdown, deterministic contracts/assets, real integration tests, pinned quality gates, safe telemetry, and dependency/secret scans — validated in Phase 1: Foundation Stability & System Boundaries (PLAT-01 through PLAT-08, SAFE-01, SAFE-08)

### Active

- [ ] Enforce identity, workspace membership, roles, invitations, and tenant isolation at service and persistence boundaries.
- [ ] Deliver link creation and management with collision-safe keys, search/filtering, lifecycle operations, and workspace ownership.
- [ ] Separate a latency-sensitive redirect data plane from the control-plane API, backed by PostgreSQL and a Redis cache.
- [ ] Emit durable click events asynchronously so analytics failures never block redirects.
- [ ] Provide custom-domain verification and routing without implementing bespoke certificate infrastructure.
- [ ] Show near-real-time link analytics across time, geography, device, referrer, campaign, and UTM dimensions.
- [ ] Accept browser and server conversion events with idempotency and attribute leads, purchases, revenue, and custom events using explainable first- and last-click models.
- [ ] Expose a versioned developer platform with scoped API keys, OpenAPI documentation, predictable errors, pagination, rate-limit metadata, and signed webhooks.
- [ ] Add campaign productivity features including campaigns, tags, folders, UTM tooling, QR codes, bulk actions, and import/export.
- [ ] Preserve extension points for integrations, partner programs, commissions, billing entitlements, enterprise identity, privacy controls, and production hardening.

### Out of Scope

- Pixel-for-pixel copying, source translation, proprietary assets, branding, naming, or internal architecture from Dub — Flux must be an independent product.
- Kafka, Kubernetes, Elasticsearch, bespoke certificate infrastructure, and dozens of microservices in initial milestones — operational complexity must follow demonstrated need.
- Machine-learning and multi-touch attribution in v1 — first-click and last-click models establish the core loop first.
- Dozens of third-party integrations in v1 — validate one integration boundary and initial Stripe flows before expanding.
- Enterprise SAML, SCIM, advanced RBAC, and disaster-recovery automation in initial milestones — these belong to later production-hardening work.
- Invasive browser fingerprinting or indefinite raw-IP retention — tracking must remain privacy-conscious.

## Context

- `spec.md` is the authoritative product-intent document. It describes a long-term product with 20 domains and explicitly requires vertical delivery rather than simultaneous implementation.
- `.planning/codebase/` records the initial source-inspected baseline and is due for a refresh after the foundation changes. Current evidence is in the Phase 1 verification, review, security audit, and development runbook. The backend exposes health and API documentation; product API routes and repositories remain extension points, and bootstrap migrations contain no product DDL.
- `apps/frontend/` is empty. Existing React usage is limited to email templates, so the dashboard remains greenfield.
- Phase 1 separated HTTP and email-worker ownership, removed implicit API migrations, embedded runtime assets, established executable unit/integration suites, and repaired health, signal handling, validation, generated-contract, credential encoding, and quality-gate boundaries. Telemetry is injected through vendor-neutral OpenTelemetry and scrubbed before stdout and OTLP export.
- The spec proposes Go binaries for API, redirector, worker, and migrations; PostgreSQL as the source of truth; Redis for redirect cache and transient coordination; NATS JetStream or Redis Streams for events; and ClickHouse when event volume justifies it.
- Control-plane traffic (`web → API → PostgreSQL`) and data-plane traffic (`visitor → redirector → Redis/PostgreSQL → redirect`, with asynchronous events) must remain operationally independent.
- The user-facing product should be original, minimalist, responsive, accessible, information-dense, keyboard-friendly, and professional.

## Current State

Phase 1 is complete: 22/22 plans, 53/53 verified must-haves, ten requirements and 19 locked decisions. The follow-up code review is clean, all 39 planned threats have verified dispositions, and the full local gate and hosted GitHub Actions run passed. Verification and audit artifacts are committed with the implementation.

Phase 2 is the next planned slice and has not started. Execution stopped after Phase 1 as requested by `--no-transition`. The marketer's link-to-revenue core loop remains future work.

## Constraints

- **Architecture**: Go-first modular architecture with independently deployable binaries where useful — scale API, redirects, and workers separately without premature microservices.
- **Existing code**: Preserve sound boilerplate conventions and adapt incrementally — do not rewrite functioning code solely to match the target folder tree.
- **Redirect performance**: Cached redirect lookup targets p95 under 30 ms server processing; uncached targets p95 under 100 ms — redirect availability cannot depend synchronously on analytics.
- **Reliability**: ClickHouse, analytics workers, webhooks, and email may fail without preventing redirects — the redirect response is the critical path.
- **Tenancy**: Every workspace-owned entity contains or resolves to a workspace identifier, with server-side authorization and isolation tests from the first product schema.
- **Data**: PostgreSQL is authoritative for transactional entities; Redis is never authoritative for link configuration; analytics may be eventually consistent.
- **Security**: Strict destination URL validation, SSRF defenses, secure sessions, CSRF protection where applicable, API-key hashing, webhook signing, rate limits, and parameterized SQL are mandatory.
- **Privacy**: Minimize collected data, avoid unnecessary raw network identifiers, support configurable retention later, and do not use invasive fingerprinting as primary identity.
- **Money**: Store monetary values in integer minor units with explicit currency — never use floating-point commission or revenue arithmetic.
- **Developer experience**: Versioned REST API, OpenAPI, idempotency, pagination, predictable errors, request IDs, rate-limit metadata, and SDK-friendly contracts are product requirements.
- **Quality**: Major domain logic requires automated tests; external input must be validated; no placeholder production paths or silently swallowed errors.
- **Brand**: Flux is temporary; no Dub-derived branding or protected assets may be used.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Use `Flux` as the temporary internal codename | The repository already uses the name and the specification requires an independent temporary codename | — Pending |
| Preserve the Go/Echo boilerplate where sound | Existing initialization, middleware, database, worker, and contract patterns provide a useful base | Validated in Phase 1 |
| Separate control plane and redirect data plane | Redirect latency and availability must not depend on dashboard or analytics availability | Independent role foundation verified in Phase 1; product redirects remain Phase 3 |
| Start modular with a few deployable binaries | Independent scaling matters, while early microservice sprawl would slow delivery | Four independently operable roles verified in Phase 1 |
| Keep PostgreSQL authoritative and Redis as cache/transient infrastructure | Strong consistency is required for ownership and link configuration | — Pending |
| Introduce ClickHouse before high-scale analytics, not by default on day one | Early PostgreSQL event storage can reduce complexity if the abstraction preserves migration | — Pending |
| Use first-click and last-click attribution initially | They are explainable, testable, and sufficient to validate the core attribution loop | — Pending |
| Build vertical MVP slices | Each phase should produce observable user value and validate architecture through working flows | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `$gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `$gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-10-08 after Phase 1 verification and completion*
