---
phase: 02-tenant-safe-link-control-plane
plan: "27"
subsystem: links
tags: [postgres, lifecycle, idempotency, nextjs, audit, tdd]
requires:
  - phase: 02-26
    provides: Session-bound browser privacy and registered authenticated product fixtures
provides:
  - Closed versioned link lifecycle transitions with fresh locked tenant authorization
  - Atomic soft-state effect, immutable scoped link audit and committed idempotency replay
  - Native safe-first confirmation and explicit review after conflict or uncertainty
affects: [02-28, 02-29, 02-31, 02-34, 02-35]
tech-stack:
  added: []
  patterns: [workspace-first locking, exact decimal BIGINT preconditions, scoped audit targets]
key-files:
  created:
    - apps/backend/internal/service/lifecycle.go
    - apps/backend/internal/service/lifecycle_test.go
    - apps/backend/internal/database/migrations/012_link_lifecycle_audit.sql
  modified:
    - apps/backend/internal/repository/link.go
    - apps/backend/internal/repository/audit.go
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/router/router.go
    - apps/frontend/components/link-detail.tsx
    - apps/frontend/components/confirm-dialog.tsx
    - apps/frontend/lib/api.ts
    - apps/frontend/tests/control-plane.spec.ts
    - packages/zod/src/identity.ts
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/src/gen.test.ts
    - packages/openapi/openapi.json
    - apps/backend/static/openapi.json
    - apps/backend/internal/transport/health.gen.go
key-decisions:
  - Link versions cross the API as exact positive BIGINT decimal strings; missing returns 428 and stale or exhausted returns 409.
  - Fresh locked SQL write capability precedes scoped link lookup and replay; effect, increment, audit and hash ledger commit together.
  - Restore returns disabled and ordinary transitions preserve suspension and permanent global host/key reservation.
requirements-completed: []
requirement-slices: [LINK-05, LINK-07, TEN-05, TEN-07, TEN-08, SAFE-03]
duration: 23min 02s
duration-basis: Measured resumed verification window through GREEN self-check; earlier interrupted execution, initial restoration reads and restart downtime are excluded
completed: 2026-10-11
---

# Phase 2 Plan 27: Versioned Link Lifecycle Summary

**Enable, disable, archive, restore and soft-delete now use exact version preconditions, fresh tenant capabilities and atomic audit/replay commits, with native confirmation and conflict recovery.**

## Performance

- Measured window: **23min 02s**, **2026-10-11T15:41:21Z** through **2026-10-11T16:04:23Z**.
- Tasks: **2/2**, preserving the existing RED commit and interrupted GREEN implementation.
- Source/test/artifact paths: **19**, including **3 created files**; PLAN clarification is separate metadata.
- Earlier interrupted work and restart downtime are unmeasured and excluded.

## Accomplishments

- Registered POST transitions accept only the five documented actions, a mandatory idempotency key and an exact positive decimal version. Missing precondition returns 428; stale/exhausted versions, invalid edges and suspended enable return documented 409 constraints. Strict JSON rejects unknown fields, nulls, trailing input and non-string versions.
- Workspace, current actor membership and scoped link locks precede all effects and replay disclosure. Owner/admin/member can write; viewer receives 403 and foreign resources receive safe 404. One version increment, immutable actor/action/time audit, scoped target and canonical payload hash ledger commit atomically.
- Soft-delete never frees a global host/key reservation. Restore of archived or deleted links returns disabled. Ordinary transitions preserve independent suspension; effective state follows Deleted, Suspended, Archived, Disabled, Active precedence.
- Native archive/delete dialogs focus the safe choice. A stale response requires explicit latest-detail review; uncertain saves disable further mutation until reload. Fresh SQL capability loss removes write actions. Deleted detail remains reachable for restore.
- Canonical Zod/REST authority and all three API artifacts include the actual route. Regex guarding precedes BigInt conversion, with malformed, overflow and unknown-input coverage.

## Task Commits

1. **Task 1: Specify versioned scoped link lifecycle transitions** — `03ca72c` (preserved RED).
2. **Task 2: Apply atomic versioned link lifecycle transitions** — `9a5d507` (GREEN).

Exact staging was used; no hook bypass, tracked deletion, push or plan 28 execution occurred. Unrelated `.serena/project.yml` remains unstaged.

## Verification

- Preserved pre-interruption evidence: complete formatting, strict lint and typechecking; canonical unit gate with **108 selected race-enabled Go tests**, **30 frontend**, **6 email**, **14 OpenAPI** and schema/script/tool suites. Source stayed unchanged during resumed verification except the already-recorded integration dispatcher ordering correction.
- Preserved focused evidence: **40 pure lifecycle/suspension combinations**, actual PostgreSQL lifecycle HTTP and state-machine race checks, and **2/2 targeted native lifecycle browser cases**.
- Final original canonical integration passed **20 registered tests across 6 groups**, including lifecycle HTTP, transactional rollback, all lifecycle edges, owner/admin/member/viewer authorization, foreign existing resources, exact BIGINT behavior, one-version concurrency, authorized replay, immutable audit and reserved-key retention.
- Private native browser diagnostic passed **50/50**, with **0 skipped, unexpected, flaky or global errors**. Final original canonical full browser also passed **50/50**, including both lifecycle cases and all existing account/invitation/library paths. Original assertions, deadlines and zero retries stayed intact.
- Production build and migration verification passed. `generate:check` passed all **five** outputs: both OpenAPI copies, generated Go transport and both email templates.
- Existing `cleanFrontendBuild` validated owned nonsymlink production output and manifest v4 before disposal. **CI=true scan passed** for all-role/test imported Go vulnerability checks, Bun dependencies, full worktree secrets and full Git history secrets. Unused GO-2026-5932 remained visible as inventory evidence without an exception.
- `git diff --check` passed; GREEN contains no tracked deletions or generated leftovers.

## Decisions Made

Exact decimal versions avoid JavaScript integer precision loss. Missing version remains representable in the canonical transport so the server can return the documented 428 instead of a generic binding error.

Fresh actor authorization precedes both scoped lookup and stored replay. Suspensions remain independent and restoring never enables routing. Link audit target uses a composite workspace/link foreign key in migration 012.

## Deviations from Plan

1. **[Rules 2/3 - Required correctness and wiring]** Added the concrete route registration, scoped immutable audit target migration, existing ConfirmDialog extension, API constraint projection and canonical generator tests. These extras are recorded in PLAN; they introduce no new architecture, suspension mutation or future feature. Commit: `9a5d507`.
2. **[Rule 1 - Bug]** Guarded schema refinement with the decimal regex before BigInt conversion; malformed input must be rejected without throwing. Regression cases include malformed strings and maximum overflow. Files: packages/zod/src/identity.ts and packages/openapi/src/gen.test.ts. Commit: `9a5d507`.
3. **[Rule 1 - Test isolation]** The earlier broad integration run reached the TestProductActualHTTP deadline. The newly inserted lifecycle dispatcher ran before baseline identity checks and seeded extra users/profiles: stable-profile proof expected **1**, observed **3**; concurrent identity proof also expects **2 global users**. A fatal assertion while the provider mutex was held skipped its Unlock, causing cascading provider-request hangs. Moved lifecycle after all existing product cases without changing identity counts, mutex assertions, deadlines or gates. Focused lifecycle proof and final full integration **20/20** passed. File: apps/backend/internal/handler/product_test.go. Commit: `9a5d507`.

## Issues Encountered

The first resumed original integration run failed unchanged TestTracestateHTTPRedisOTLP behind the sanitized wrapper. Its underlying assertion was not retained. A private diagnostic initially applied restrictive umask to the whole command, making the nonroot collector configuration unreadable; those diagnostic failures are excluded. Corrected capture precreated a mode-0600 report inside a mode-0700 OS temporary directory while retaining normal fixture permissions. The exact native 13-test app command then passed, followed by all 20 tests in the original canonical integration gate. No telemetry code, assertions or limits changed; the original failure was not reproduced and no cause is claimed.

The first original full browser gate also failed behind the sanitized wrapper. A private runner copy changed only repository-root resolution and the first command-close capture of stdout, preserving fixture setup, deadlines and native validators. It passed 50/50, followed by 50/50 in the original canonical gate. The initial failure was not reproduced; no browser behavior, fixture, retry, timeout or gate changes were made during this resumed run. Failed runs are not counted as completed verification.

## Known Stubs

None in the delivered lifecycle path. Existing detail copy about future redirects/analytics is outside this plan; no redirect or analytics behavior is claimed.

## Remaining Acceptance and Next Plan Readiness

Plan 27 is ready for subsequent suspension and lifecycle refinements. Plan 28 was not started. Broad requirements remain pending until final acceptance in **02-35**, so requirements-completed is intentionally empty. Local signed provider fixtures do not establish actual provider verification/recovery/social sign-in/session-cookie acceptance; that external checkpoint remains outstanding.

## Self-Check: PASSED

All 19 owned source/test/artifact paths exist. Both RED `03ca72c` and GREEN `9a5d507` are present in Git history. GREEN contains no tracked deletions. No undisclosed security surface or goal-blocking stub was found in the owned changes.
