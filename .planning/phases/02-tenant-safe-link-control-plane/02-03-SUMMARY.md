---
phase: 02-tenant-safe-link-control-plane
plan: "03"
subsystem: auth
tags: [postgresql, clerk, pgx, identity, uuid, playwright, openapi]
requires:
  - phase: 02-02
    provides: Signed bearer verification, active provider sessions and real-router browser fixture
provides:
  - Immutable internal UUID keyed uniquely by configured issuer and provider subject
  - First-mapping primary-email verification through the explicitly injected Clerk SDK
  - Committed canonical /me identity response backed by PostgreSQL
  - Registered real-PostgreSQL product dispatcher with concurrency, cancellation and rollback proof
affects: [02-04, 02-05, 02-35]
tech-stack:
  added: []
  patterns: [Read-committed conflict insertion and subsequent lookup, Bounded detached rollback, Verified profile snapshots]
key-files:
  created:
    - apps/backend/internal/database/migrations/002_identity.sql
    - apps/backend/internal/repository/user.go
    - apps/backend/internal/service/identity.go
    - apps/backend/internal/handler/product_test.go
  modified:
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/app/api.go
    - apps/backend/internal/router/router.go
    - apps/backend/internal/service/services.go
    - apps/backend/internal/service/auth_test.go
    - apps/backend/internal/app/roles_test.go
    - apps/backend/internal/app/observability_test.go
    - apps/backend/internal/database/migrator_test.go
    - packages/zod/src/identity.ts
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/src/gen.test.ts
    - packages/openapi/openapi.json
    - apps/backend/static/openapi.json
    - apps/backend/internal/transport/health.gen.go
    - apps/frontend/tests/control-plane.spec.ts
    - scripts/check.ts
key-decisions:
  - "Treat verified email as a profile snapshot, never an identity key or membership authority; durable identity is the exact configured issuer/subject pair."
  - "Return only committed identities; use conflict insertion followed by a separate read-committed lookup and detached bounded rollback on cancellation."
  - "Preserve canonical UUID/email formats with authored x-go-type string metadata, avoiding a new generated runtime dependency."
requirements-completed: []
requirements-addressed: [TEN-01, SAFE-02, SAFE-03]
duration: 18min
completed: 2026-10-08
---

# Phase 2 Plan 3: Durable Provider Identity Summary

**Verified Clerk issuer/subject pairs now resolve through PostgreSQL to immutable internal UUIDs, returned by the signed production `/api/v1/me` route only after commit.**

## Performance

- Started: 2026-10-08T18:08:28Z (implementation start after loading execution context)
- Completed: 2026-10-08T18:26:43Z
- Duration: approximately 18 minutes
- Tasks: 2/2
- Product/test/config files created or modified across task commits: 20 (14 declared plus six necessary wiring/test adaptations)

## Accomplishments

- Migration 002 creates global users with UUID primary keys, unique `(issuer, subject)`, verified email and verification/creation timestamps. A database trigger prevents reassignment of the UUID, issuer or subject. Email is deliberately nonunique; two provider identities with the same email remain distinct.
- The identity service requires the configured issuer, retrieves first-mapping profiles through the injected fixed-endpoint Clerk user client, checks matching subject, primary verified email and banned/locked state, and validates email against the canonical response policy. Profile failures are safe bounded denials/unavailability without a partial row. Stored email is a snapshot; invitation authorization must reverify current provider data in its later plan.
- Parameterized PostgreSQL lookup and transactional conflict insertion return the committed row. A separate read after `ON CONFLICT DO NOTHING` observes the concurrent winner. Failed commits and canceled inserts return no identity; rollback uses a detached five-second cleanup deadline.
- Existing role/service/router composition injects the new resolver. The authored Zod and ts-rest modules own the strict authenticated user response; both JSON documents and the generated Go transport were regenerated together. UUID/email retain their wire formats, using schema metadata to emit Go strings without adding a dependency.
- The registered `TestProductActualHTTP` fixture has its own signed SDK clients and real migrated pinned PostgreSQL. The existing browser fixture also resolves against real PostgreSQL. Authentication-only unit cases retain their signed production-router proof with an explicit test-local resolver.

## Task Commits

1. Task 1 behavioral RED: `dae26c2` — `test(02-03): specify durable identity through signed HTTP and browser requests`
2. Task 2 implementation GREEN: `7658e6a` — `feat(02-03): persist verified provider identities as durable internal UUIDs`

## Verification

- Actual-HTTP RED executed against the production router and migrated PostgreSQL: HTTP 200 lacked the required internal UUID. The failure was `invalid UUID length: 0`, with a working fixture.
- Browser RED executed one identity case and failed on the missing user identity; the private JSON report recorded zero infrastructure errors. Signed tokens/provider diagnostics were kept outside the repository in a private OS temporary directory.
- Final race-enabled `TestProductActualHTTP`: **15 identity subtests passed**, covering committed/stable UUID, **eight concurrent first mappings**, distinct same-email subjects and their actual returned UUIDs, profile outage/deadline/mismatch/unverified/missing-primary/invalid-email/banned/noncanonical-email denials, deferred commit rollback, configured issuer separation, immutable keys, authenticated insert cancellation and successful/error log privacy. Cancellation waits for the actual PostgreSQL insert lock before canceling the request and verifies no persisted row or idle transaction remains.
- `bun run test:integration`: **passed all 16 registered top-level integration groups**. The final additional identity cases were subsequently verified with the targeted race-enabled dispatcher.
- `bun run test:e2e -- --project=local --grep identity`: **one actual browser case passed**, including the final strict TypeScript result guard. The complete local browser suite passed **10 cases**, preserving all prior native sign-in and signed bearer cases.
- `bun run test:unit`: **passed**, including all workspace/script/tool self-tests and **80 discovered race-enabled Go unit tests**. The OpenAPI suite now has **13 tests**, including strict identity schema and canonical UUID/email assertions. An initial frontend build caught unchecked array access in the new test; adding an explicit result guard fixed it.
- Updated real PostgreSQL migrator regression passed: existing bootstrap version **1 upgrades to 2**, actual Tern down/up permits blank-ledger executable checks, and the private version-update failure injection remains meaningful. Updated role binary startup and telemetry-outage migrator tests passed with race detection and exact latest-version assertions.
- Owned Go packages lint: **zero issues**. Changed authored TypeScript formatting and `git diff --check` passed. `bun run generate:check` passed with the three contract outputs synchronized.
- Verified frontend build cleanup ran immediately before final `bun run scan:secrets`: **worktree clean; full history clean**. No scanner rules or exceptions changed.

## Decisions Made

- Provider namespace and subject identify a user; profile email never merges accounts or grants tenancy.
- Existing mappings reuse their durable snapshot while every HTTP request continues to verify active provider session state.
- Explicit resolver injection keeps provider-only unit tests bounded while browser/domain evidence exercises real PostgreSQL through production routing.

## Deviations from Plan

### Auto-fixed Blocking Issues

**1. [Rule 3 - Blocking] Complete existing production dependency wiring**
- The declared handler/service files could not become reachable without updating `app/api.go`, `router/router.go` and `service/services.go`.
- The API uses the existing service constructor, which injects the authoritative user repository and identity resolver. No role resources, runtime bypass or unrelated service layer was introduced.
- Verified by actual router/domain tests, existing role binary tests and the unit gate; committed in `7658e6a`.

**2. [Rule 3 - Blocking] Preserve affected bearer and contract assertions**
- Updated `service/auth_test.go` with a verified profile response for the real-PG browser fixture and an explicit unit-only resolver for the provider-boundary corpus. Existing signature, revocation, authorization, deadline and concurrency cases remain intact.
- Updated `packages/openapi/src/gen.test.ts` to assert the new strict schema and exact canonical formats without dropping health/privacy tests.
- Verified by unit/browser/contracts gates; committed in `7658e6a`.

**3. [Rule 3 - Blocking] Preserve baseline blank-ledger migration proof after product DDL**
- The existing `database/migrator_test.go` replayed migrations by deleting only the ledger, which would collide with real product tables and mask its intended failure injection.
- It now uses the actual embedded Tern down migrations before resetting the ledger and before installing the version-update failure trigger, and adds real version-1 upgrade proof.
- Verified by the migrator/binary integration regression; committed in `7658e6a`.

Six additional concrete files were necessary; total ownership remained 20 implementation/test/config files. No other plan was executed.

## Issues Encountered

New-code lint findings and strict browser-test array indexing were corrected before GREEN. UUID/email generation initially selected an absent runtime type dependency; canonical field metadata preserves format validation while generating standard Go strings. No package was installed or substituted.

## Known Stubs

None in the implemented production identity path. No new network endpoint or trust boundary outside the plan's identity threat model was introduced.

## User Setup Required

No new setup for this bounded plan. Live Clerk configuration and provider-factor acceptance remain pending the final plan 02-35.

## Next Plan Readiness

Plan 02-04 can use durable internal identity for explicit workspace creation. Workspace/membership authority remains future work; provider organization claims grant no access. Broader TEN-01/SAFE-02/SAFE-03 requirements are intentionally not marked complete before their remaining phase evidence and live provider acceptance.

## Self-Check: PASSED

- Created migration, repository, identity service and product dispatcher exist.
- RED `dae26c2` and GREEN `7658e6a` exist in repository history in that order.
- All required behavioral and generated artifacts exist; no unexpected file deletions occurred.
