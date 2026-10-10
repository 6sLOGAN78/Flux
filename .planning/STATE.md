---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: milestone
status: executing
stopped_at: Completed 02-21-PLAN.md
last_updated: "2026-10-10T16:42:58.177Z"
last_activity: 2026-10-10
progress:
  total_phases: 6
  completed_phases: 1
  total_plans: 57
  completed_plans: 43
  percent: 17
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-10-05)

**Core value:** A marketer can create a link, send a visitor through a fast and reliable redirect, record the click asynchronously, attribute a later conversion and revenue to that click, and see the result in analytics.
**Current focus:** Phase 2 — Tenant-Safe Link Control Plane

## Current Position

Phase: 2 (Tenant-Safe Link Control Plane) — EXECUTING
Plan: 22 of 35
Status: Ready to execute
Last activity: 2026-10-10

Progress: [████████░░] 75%

## Performance Metrics

**Velocity:**

- Total plans completed: 44
- Average duration: -
- Total execution time: 0.0 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01 | 22 | - | - |
| 1 | 22 | - | - |

**Recent Trend:**

- Last 5 plans: -
- Trend: No execution data

*Updated after each plan completion*
| Phase 01 P01 | 9min | 2 tasks | 7 files |
| Phase 01 P02 | 3min | 2 tasks | 5 files |
| Phase 01 P03 | 4min | 2 tasks | 9 files |
| Phase 01 P05 | 6min | 2 tasks | 8 files |
| Phase 01 P04 | 9min | 2 tasks | 7 files |
| Phase 01 P06 | 6min | 2 tasks | 4 files |
| Phase 01 P07 | 5min | 2 tasks | 10 files |
| Phase 01 P08 | 6min | 2 tasks | 2 files |
| Phase 01 P09 | 10min | 2 tasks | 12 files |
| Phase 01 P10 | 6min | 2 tasks | 6 files |
| Phase 01 P11 | 7min | 2 tasks | 7 files |
| Phase 01 P12 | 11min | 2 tasks | 10 files |
| Phase 01 P13 | 13min | 2 tasks | 8 files |
| Phase 01 P14 | 7min | 2 tasks | 4 files |
| Phase 01 P15 | 11min | 2 tasks | 9 files |
| Phase 01 P16 | 14min | 2 tasks | 6 files |
| Phase 01 P17 | 9min | 2 tasks | 18 files |
| Phase 01 P18 | 11min | 2 tasks | 5 files |
| Phase 01 P19 | 18min | 2 tasks | 5 files |
| Phase 01 P20 | 46min | 2 tasks | 10 files |
| Phase 01 P21 | 36min | 2 tasks | 4 files |
| Phase 01 P22 | 29min | 2 tasks | 3 files |
| Phase 02 P01 | 25min | 2 tasks | 14 files |
| Phase 02 P02 | 286min | 2 tasks | 16 files |
| Phase 2 P04 | 24min | 2 tasks | 16 files |
| Phase 2 P05 | 12min | 2 tasks | 15 files |
| Phase 2 P06 | 26min | 2 tasks | 20 files |
| Phase 2 P07 | 15min | 2 tasks | 16 files |
| Phase 2 P08 | 15min | 2 tasks | 22 files |
| Phase 2 P09 | 23min | 2 tasks | 11 files |
| Phase 2 P10 | 38min | 2 tasks | 22 files |
| Phase 2 P11 | 25min | 2 tasks | 17 files |
| Phase 2 P12 | 15min | 2 tasks | 13 files |
| Phase 2 P13 | 20min | 2 tasks | 16 files |
| Phase 2 P14 | 34min | 2 tasks | 23 files |
| Phase 2 P15 | 20min | 2 tasks | 14 files |
| Phase 2 P16 | 10min | 2 tasks | 21 files |
| Phase 2 P17 | 30min | 2 tasks | 19 files |
| Phase 2 P18 | 24min | 2 tasks | 19 files |
| Phase 2 P19 | 47min | 2 tasks | 32 files |
| Phase 2 P20 | 20min | 2 tasks | 17 files |
| Phase 2 P21 | 47min | 2 tasks | 29 files |

## Accumulated Context

### Roadmap Evolution

- Phase 1 edited: Normalized the same foundation goal to canonical MVP user-story syntax; success criteria, requirements and scope unchanged

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- [Roadmap]: Treat existing code as scaffolding until observable product behavior is verified.
- [Phase 1]: Establish independent API, redirector, worker, and migrator roles before product slices.
- [v1]: PostgreSQL remains authoritative; Redis is derived cache/transient state; analytics stays behind a stable adapter.
- [v1]: The milestone ends with the managed-domain link → click → conversion → attribution → analytics journey.
- [Phase 01]: Retain verified Node22.23.3 LTS and Bun1.3.14; select supported Go1.26.8 with identical digest-pinned PostgreSQL17.11 and Redis8.10.2 for local/test infrastructure.
- [Phase 01]: Separate raw SetupTestPostgres from migrating SetupTestDB; preserve migration failures for plan01-04.
- [Phase 01]: Generic Echo handlers require per-invocation request factories; binder failures return Invalid request and validation failures return Validation failed with recognized field errors.
- [Phase 01]: Configuration and cleanup errors expose stable stage/resource labels while retaining private causes through Unwrap; Asynq void Shutdown uses a narrow error-returning test seam. — Preserve safe diagnostics, original failures, existing FLUX key compatibility, and deterministic shutdown failure tests without starting workers.
- [Phase 01]: Recover named health component schemas from the same authored ts-rest router inside the owned generator. — Existing document construction replaces components; preserve operation metadata and security schemes without expanding file ownership.
- [Phase 01]: Preserve Migrate compatibility through MigrateWithResult; use PostgreSQL-only migrator role and explicit harmless bootstrap SQL. — Avoid adjacent caller changes while enabling exact deterministic migration results and safe one-shot resource ownership.
- [Phase 01]: Pin official oapi-codegen v2.8.0 and generate package-local health aliases through the inline typedef template. — Preserve canonical TS/OpenAPI schema authority, verified module sums, exact byte reproducibility, and owned-file boundaries.
- [Phase 01]: Embed package-owned docs and email FS; pin Scalar 1.73.0 standalone with verified sha384 SRI and hash-only script CSP; render emails through closed enum ParseFS with missing-key errors. — Preserve working-directory independence, canonical contract bytes, safe escaping, and provider-free tests without changing existing constructors or installing dependencies.
- [Phase 01]: Regenerate authored packages and all four artifact categories in isolation; verify pinned generator sums and compare bytes without Git. — Avoid stale build authority, preserve checked files during checks and generator failures, and keep tool output and temporary files inside safe boundaries.
- [Phase 01]: Use independent app role graphs with opt-in API producers, consumer-only worker processing and role-owned shared Redis; retain app cleanup aliases over lifecycle. — Preserve existing package adapters and cleanup APIs, prevent composition import cycles, and guarantee partial-startup resource release without starting undeclared consumers.
- [Phase 01]: Use thin explicit role mains, an API-only Flux shim, and SDK-supported local email transport override for subprocess verification. — Preserve tested role resource ownership, keep migrations explicit, avoid new production seams, and verify real command behavior with minimal configuration.
- [Phase 01]: Inject role-owned readiness checks with one request deadline; worker email health stays local and health diagnostics use safe classifications. — Preserve required-only resource boundaries, avoid provider availability gating, prevent credential-bearing error disclosure, and match canonical generated health contracts.
- [Phase 01]: Use one readiness-first serial lifecycle deadline and budget-aware Asynq idle polling; commands share NotifyContext and return bounded failure status. — Preserve active-work drain, observable worker shutdown and reverse dependency ownership without unsafe overlapping cleanup or fresh stage budgets.
- [Phase 01]: Pin official OTel v1.47.0 API/SDK with v0.23.0 HTTP logs; use fixed scopes, closed export allowlists and bounded independent provider shutdown. — Preserve safe trace correlation, module provenance, optional monitoring and existing role compatibility before later integration.
- [Phase 01]: Preserve inert vendor consumer adapters through plan 01-17; bind optional dot-nested OTLP settings and sanitize stdout plus injected OTel logs through one closed allowlist. — Preserve external configuration and adjacent compilation without vendor initialization, global providers or premature integration ownership.
- [Phase 01]: Inject HTTP tracer/meter APIs with validated UUID correlation, traceparent-only ingress and closed route/method/role/status-class metrics; retain the minimal Server.Telemetry seam for later role composition. — Preserve response/log/span identity and safe public validation while stripping private propagation and request/provider fields before local telemetry capture.
- [Phase 01]: Migrate unsafe legacy jobs via canonical replacement enqueue followed by RevokeTask; inject narrow email delivery and optional telemetry while retaining existing drain behavior. — Asynq shares payload/header objects with concurrent cancellation persistence; public APIs retain recoverability and remaining retry budget without a data race or Redis-internal schema dependency.
- [Phase 01]: Register role telemetry first and flush last; share one migrator PostgreSQL/provider exit context; remove vendor types and preserve safe operational diagnostics. — Keep tested serial drain and source-compatible constructors without global providers, successive exit deadlines or vendor dependencies.
- [Phase 01]: Preserve safe nonzero application Links but drop complete collector Link/event collections with pinned OTTL 0.162.0; validate dirty signals and final 200 cumulative metrics independently.
- [Phase 01]: Pin official Biome2.5.15, golangci-lint2.14.0, staticcheck2026.2.1, Gitleaks8.30.1 and source-built govulncheckv1.8.0 with Go1.26.8 and verified source/executable hashes. — Upstream govulncheck has no prebuilts; isolated readonly source builds reproduce exact measured hashes without application dependencies; preserve Go checks through v2 migration.
- [Phase 01]: Use explicit uncached quality stages and compare the compiled Go test listing with all selected unit/integration tests; preserve direct Taskfile migrations to prevent recursive test dispatch. — Guarantee backend and workspace completeness, successful race-enabled test execution and bounded private diagnostics without weakening existing gates.
- [Phase 01]: Scan all backend roles and tests at package exposure level; retain unused module inventory advisories visibly; constrain public scanner metadata to recognized identifiers and manifest/file-backed labels. — Fail every vulnerable imported package and execution/report error without advisory exceptions or disclosure of private scanner content.
- [Phase 01]: Use shared runtime manifest inputs and assert installed Go/Node/Bun match package, module and scanner manifests; retain verified official action SHAs.
- [Phase 01]: Disable caches and artifact uploads initially; retain only bounded sanitized CI logs and preserve mandatory full-history scanning.
- [Phase 01]: Make explicitly local scanner fixtures select ci:false while preserving production CI full-history enforcement and its missing/shallow-history regressions.
- [Phase 02]: Local browser evidence uses real Clerk SDK/UI with test-only transport interception; live verification, recovery and OAuth acceptance remain pending final plan 02-35.
- [Phase 02]: Reuse already-locked @types/node 22.19.19 for Next preflight, preserve runtime and React graphs, and keep generated Next outputs outside tracked source.
- [Phase 02]: Use explicitly keyed fixed-endpoint SDK clients and one two-second JWKS/session deadline for bearer-only Go authentication; live provider acceptance remains pending 02-35.
- [Phase 2]: Protect the installed versioned product router with exact browser mutation safeguards and verified Actor context; registered test-only real-PG probes verify enforcement without production endpoints.
- [Phase 2]: Validate account responses through canonical @flux/zod workspace exports and bind rendered identity to its active session; live provider cookie/session acceptance remains pending 02-35.
- [Phase 2]: Reuse signed provider helpers below application layers with external test-only product composition and private OS temporary browser output. — Preserve existing protocol, generated bytes and production auth boundaries while guaranteeing completed browser evidence and bounded teardown.
- [Phase 02]: Patch pinned Go to 1.26.9 and x/net to v0.60.0; reproduce govulncheck v1.8.0 binaries with unchanged source verification and deterministic recipe. — Clear newly published imported-package advisories without scanner exceptions, unrelated upgrades or phase advancement; full local check passed.
- [Phase 2]: Workspace bootstrap serializes durable actor identity; replay drops that lock before fresh workspace-first membership authorization. — Keep atomic canonical retries and tenant isolation without lock-order inversion or provider organization authority.
- [Phase 2]: Standalone browser execution builds its canonical schema dependency; the first-link CTA remains disabled until 02-14 implements creation. — Prevent stale dist from substituting for authored contracts and avoid inventing an absent production route.
- [Phase 2]: Authorize current workspace membership before bootstrap hash comparison; actor-scoped cleanup removes only expired ledger records. — Preserve safe denial, database-enforced minimum 24-hour retention and workspace-first lock ordering.
- [Phase 2]: Consume the explicit repository registry and show authorized Team availability without inventing an endpoint. — Native Links shell is reachable; actual Team operations remain plan 02-16, first-link creation 02-14 and live provider acceptance 02-35.
- [Phase 2]: Derive restored selection from the same current membership snapshot as the chooser; a stored preference never grants authority or exposes removed workspace identifiers.
- [Phase 2]: Select under the existing workspace-first shared lock and recheck membership before saving the preference; creation commits its initial preference atomically.
- [Phase 2]: Keep provider-only unit bootstrap substitution strictly test-local; all product and browser fixtures use real migrated PostgreSQL and production middleware.
- [Phase 2]: Broadcast only a fixed invalidation signal; server bootstrap and committed selection remain workspace authority.
- [Phase 2]: Dispose the scoped subtree before switching, signout or access loss; retain valid same-workspace content during transient refresh failure.
- [Phase 2]: Consume native unsaved confirmation in the existing onboarding name flow; future link drafts must use the same scoped disposal and confirmation boundaries.
- [Phase 2]: Fresh membership precedes link idempotency replay; link effect and response ledger commit atomically. — Membership removal must deny replay without erasing durable creator provenance.
- [Phase 2]: Generated managed-host keys stay reserved after soft deletion; retry only the named global key constraint at most five times. — Bounded savepoints preserve atomicity and never hide unrelated database failures.
- [Phase 2]: Preserve network-free creation and conservatively deny special IPv4/IPv6 literals; syntactic host acceptance never proves public DNS answers.
- [Phase 2]: Keep crypto/rand.Reader in the original repository constructor; explicit io.Reader injection verifies entropy failure and bounded collision rollback without runtime toggles.
- [Phase 2]: Consume extracted form/detail components through existing scoped routes while preserving canonical DTOs, dirty confirmation and request disposal.
- [Phase 2]: Validate ASCII before lowercase custom-key hashing and preserve global host/key uniqueness including deleted links.
- [Phase 2]: Omit empty customKey from canonical create hashes to preserve existing generated-key replay compatibility.
- [Phase 2]: Project only recognized matching-status canonical constraint codes through a bounded browser error reader; never render server messages.
- [Phase 2]: List bounded nondeleted links under fresh membership with deterministic timestamp/UUID order; reject unsupported cursor/filter input until complete refinements.
- [Phase 2]: Preserve custom-key migration 006; introduce library index through forward migration 007.
- [Phase 2]: Bound UUID-scoped collection GET success at 8 MiB for worst-case escaped destinations; retain 64 KiB other success and 8 KiB errors.
- [Phase 2]: Bind versioned HMAC cursor positions to workspace and exact effective search/lifecycle fingerprint; fresh caller SQL scope remains authority.
- [Phase 2]: Require a private standard-Base64 32-byte cursor signing key only for API startup, without fallback or unsafe diagnostics.
- [Phase 2]: Remember only successful scoped pages locally and announce exhaustion only from an authoritative null continuation.
- [Phase 2]: Connect the approved library creation CTA to the existing real create route and verify its committed browser flow.
- [Phase 2]: Validate raw UTF-8 search at 200 Unicode characters before trimming Unicode whitespace; bind exact normalized search and closed lifecycle into signed cursors.
- [Phase 2]: Use fixed parameterized ILIKE with explicit literal pattern escaping; native committed search/filter changes abort older generations and reset scoped page history while preserving drafts on transient failures.
- [Phase 2]: Read Team under current SQL owner/admin capability in the same workspace-first transaction as its fixed 25-row query; UUID continuation never grants authority.
- [Phase 2]: Treat Team 403 as capability loss: dispose rows and refresh current workspace role without broadcasting membership removal; preserve valid Links access.
- [Phase 2]: Keep Team role enum names distinct in canonical schema metadata and regenerate both OpenAPI copies and Go transport together.
- [Phase 2]: Role mutation locks workspace then current actor and target; reuse the closed role matrix and count owners in the same atomic effect/audit/ledger transaction.
- [Phase 2]: Authorize actor and current target before replay hash reads; replay keeps the original member snapshot and current actor capability projection.
- [Phase 2]: Committed self-demotion clears Team through the existing capability-loss refresh while retaining Links; mutation failure requires explicit reload and native confirmation.
- [Phase 2]: Existing immutable audit_events and scoped mutation_requests provide durable user provenance and 24-hour replay; role changes require no schema migration or removable membership foreign key.
- [Phase 2]: Missing-target removal replay uses its own committed target-role snapshot after fresh locked actor authorization, with target policy before replay hash.
- [Phase 2]: Membership-only deletion reuses durable users, immutable audit and 24-hour ledger; links and reserved keys need no migration.
- [Phase 2]: Committed removal sends fixed invalidation through the owned channel; explicit local self-disposal preserves signed-in session and fresh server membership authority.
- [Phase 2]: Queue invitations and encrypted intents atomically in PostgreSQL; API does not consume the queue or send email and Queued remains truthful until future worker acknowledgment.
- [Phase 2]: API requires externally configured exact 32-byte AES keys with safe IDs; standard random-nonce GCM authenticates workspace, invitation, delivery and key identifiers, while durable users preserve invitation provenance.
- [Phase 2]: Fresh locked actor and protected invitation role snapshot authorization precede replay hash access; native Team invitations preserve the existing safe-first dirty switch and request generation boundaries.
- [Phase 2]: Project invitation delivery from the unique scoped intent with expiry, acceptance and revocation precedence; browser status fixtures prove presentation, not worker sending.
- [Phase 2]: Extend actual canonical invitation enum with Delivered and Failed and regenerate both OpenAPI copies and Go transport together; preserve existing configuration names and external key provisioning.
- [Phase 2]: Retain old encryption keys until live ciphertext drains or is safely re-encrypted; count messages across replicas and rotate before the per-key random-nonce limit.
- [Phase 2]: PostgreSQL owns eight total invitation attempts and lease recovery; failed Redis publication does not consume a send attempt and invitation Asynq retries are disabled.
- [Phase 2]: Freeze complete provider requests inside the existing encrypted envelope before publication; immutable intent IDs identify content and lease generations only fence processing.
- [Phase 2]: Commit Delivered and ciphertext erasure atomically after acknowledgement, preserving retryable state for errors and stale acknowledgements.
- [Phase 2]: Keep old encryption keys while retained ciphertext needs them; future resend must replace the intent ID or introduce an explicit immutable content generation.

### Pending Todos

None yet.

### Blockers/Concerns

- [Phase 4]: Click-loss SLO and durable handoff/storage model must be decided and failure-tested during planning.
- [Phase 6]: Validate PostgreSQL-first analytics against expected load before adopting ClickHouse.

## Deferred Items

Items acknowledged and carried forward from the v1 scope boundary:

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| v2 | Custom domains, productivity features, developer platform, integrations, partner/billing, and enterprise expansion | Deferred | Initial roadmap |

## Session Continuity

Last session: 2026-10-10T16:42:58.161Z
Stopped at: Completed 02-21-PLAN.md
Resume file: None
