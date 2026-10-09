---
phase: 02-tenant-safe-link-control-plane
plan: "10"
subsystem: links
tags: [go, postgres, tenancy, idempotency, url-validation, nextjs]
requires:
  - phase: 02-09
    provides: Scoped workspace navigation and disposal
provides:
  - Atomic authorized generated-key managed-domain link creation
  - Scoped committed detail with durable creator provenance
  - API-only managed-host policy and strict canonical link contracts
affects: [02-11, 02-12, 02-14, 02-18, 02-27, 02-35]
tech-stack:
  added: []
  patterns: [workspace-before-actor transaction locking, named-constraint savepoint retries, generated canonical DTOs, scoped draft disposal]
key-files:
  created:
    - apps/backend/internal/database/migrations/005_links.sql
    - apps/backend/internal/repository/link.go
    - apps/backend/internal/service/link.go
    - apps/frontend/app/workspaces/[workspaceId]/links/new/page.tsx
    - apps/frontend/app/workspaces/[workspaceId]/links/[linkId]/page.tsx
  modified:
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/config/config.go
    - apps/backend/internal/config/config_test.go
    - apps/backend/internal/router/router.go
    - apps/backend/internal/service/services.go
    - apps/backend/internal/testing/browser_fixture_test.go
    - apps/backend/internal/app/roles_test.go
    - apps/backend/internal/database/migrator_test.go
    - apps/frontend/components/workspace-switcher.tsx
    - apps/frontend/tests/control-plane.spec.ts
    - packages/zod/src/identity.ts
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/src/gen.test.ts
    - packages/openapi/openapi.json
    - apps/backend/static/openapi.json
    - apps/backend/internal/transport/health.gen.go
key-decisions:
  - Fresh workspace membership precedes actor serialization and every idempotency replay; effect and response ledger commit together.
  - Retry only the named global managed-host/key unique constraint, at most five attempts; preserve keys across soft deletion.
  - Require operator managed-host configuration only for API startup and expose it through the authorized workspace summary.
requirements-completed: []
requirements-addressed: [LINK-01, LINK-02, LINK-03, LINK-07, LINK-08, TEN-05, TEN-07, TEN-08, SAFE-03, SAFE-04]
duration: 38min
completed: 2026-10-09
---

# Phase 2 Plan 10: Safe Managed-Domain Link Creation Summary

**Signed browser requests create validated generated-key links atomically in PostgreSQL and display scoped detail with creator provenance that survives membership removal.**

## Performance

- **Duration:** Approximately 38 minutes of active execution
- **Completed:** 2026-10-09
- **Tasks:** 2
- **Files changed:** 22 implementation/test/generated paths, plus this summary and tracking metadata

## Accomplishments

- Migration 005 adds workspace-owned links, durable user attribution, globally reserved managed-host keys, closed lifecycle states, positive versions, and consistent suspension fields.
- Creation checks fresh membership before ledger replay, serializes the actor, hashes canonical payloads, and commits the response ledger with the link. Twelve random bytes produce a 20-character lowercase unpadded base32 key. Savepoints retry only the exact global key constraint, with five attempts maximum.
- Core URL policy already rejects non-public HTTP(S) targets, malformed/opaque/relative URLs, credentials, controls, backslashes, bad ports, private and special IPs, metadata/internal hosts, blocked hosts, and managed-host loops. Creation performs no destination fetch, preview, or DNS network request.
- Strict authored Zod/ts-rest contracts generate all three artifacts together. Thin Go adapters expose actual scoped POST and detail GET routes. API startup requires a validated operator managed hostname; other roles do not require link configuration. Comma-separated blocked hosts bind explicitly without changing other environment mapping.
- The actual signed-SDK browser form and detail path use production Go handlers and real PostgreSQL. Draft switching uses the existing accessible confirmation, generation/abort handling, and scope disposal. Detail escapes protected values, copies the short URL, renders full local timestamps and durable creator identity, and accurately states that redirects and analytics are unavailable.

## Task Commits

1. **Task 1: Specify safe generated-key creation and detail** — `5a31beb` (`test`)
2. **Task 2: Deliver atomic safe creation and scoped detail** — `74a5b06` (`feat`)

## Verification

- RED reached the production HTTP router: expected creation 201, actual 405 for the missing route. The independently captured targeted browser RED executed the named creation/detail case and timed out at the absent Destination URL control, with zero infrastructure errors, skips, or flaky cases. Raw reports remained outside the repository.
- Final `bun run test:unit` passed all **88 discovered race-enabled Go top-level unit tests**, all four workspace suites, and script/tool regression checks.
- Full registered integration passed **17 top-level tests across six groups** using the existing runner and private bounded capture. The registered real-PG link corpus proves concurrent replay, payload conflict, viewer and removed-member denial, foreign detail isolation, rollback without a ledger, two collisions followed by success, exactly five attempts on exhaustion, and exactly one attempt for an unrelated unique violation. Creator projection remains safe and durable after membership removal.
- Targeted browser passed **3 completed cases**; full local browser passed **22 completed cases**, both with zero skips or flaky cases. Cases cover committed escaped detail, retained/discarded draft switching, and missing-resource versus confirmed-membership-loss recovery.
- Root format, lint, typecheck, canonical generation check, and final production build passed. The production build ran after browser testing; recognized build output was cleaned before the scan.
- `CI=true bun run scan` passed imported-package Go vulnerability checks for all roles/tests, Bun audit, worktree secret scan, and full-history secret scan. Existing inventory-only **GO-2026-5932** remains visible; its vulnerable package is not imported. No pins, dependencies, allowlists, or scanner bounds changed.
- `git diff --check` passed. No tracked files were deleted.

## Decisions Made

- Authorization is evaluated before idempotency replay, and creator attribution references durable users rather than removable memberships.
- Global generated keys remain reserved after soft deletion; unrelated database errors are never treated as retryable collisions.
- The form obtains immutable operator managed-host configuration from the existing authorized workspace summary, avoiding an additional configuration endpoint.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Add the eight narrowly required wiring and affected baseline paths**
- **Found during:** Task 2
- **Issue:** The fourteen declared paths alone could not register/inject the actual service, configure real fixtures, propagate draft/access loss, or preserve existing affected configuration/migration/contract checks.
- **Fix:** Updated `router/router.go`, `service/services.go`, `testing/browser_fixture_test.go`, `components/workspace-switcher.tsx`, `config/config_test.go`, `app/roles_test.go`, `database/migrator_test.go`, and `packages/openapi/src/gen.test.ts`. These are the eight additional paths expressly authorized by the orchestrator; no other plan feature was implemented.
- **Commit:** `74a5b06`

**2. [Rule 1 - Bug] Correct new feature integration and strict runtime details**
- **Found during:** Task 2 verification
- **Issue/Fix:** Corrected creator projection to the existing `verified_email` column; bound comma-separated blocked-host policy through the installed environment provider; provided explicit React ref initialization; decomposed lint findings and aligned/tagged repository snapshots without security or complexity suppressions.
- **Commit:** `74a5b06`

**3. [Rule 3 - Blocking] Keep business fixtures independent of production rate-limit exhaustion**
- **Found during:** Task 2 real-PG verification
- **Issue:** The expanded burst of signed business requests exhausted the production limiter, obscuring assertions.
- **Fix:** Test-only requests use distinct valid client IPs. The production limiter is unchanged. Updated the affected workspace envelope assertion to retain exact workspace comparison and explicitly check the configured hostname.
- **Commit:** `74a5b06`

## Issues Encountered

No remaining production blocker or authentication gate. The previous plan's telemetry transient remains in its existing deferred record; it did not recur in this plan's final integration run and is not claimed fixed.

## Known Stubs and Scope Limits

- The existing library first-link CTA remains disabled until plan 14; the bounded new-form/detail routes are functional directly. Existing Team navigation remains reserved for plan 16. These approved staged navigation limits do not replace the real creation/detail implementation.
- Redirect serving, analytics, custom keys/domains, editing, lifecycle mutations, and billing remain later work. The displayed short URL carries an explicit management-only availability notice.
- Broad requirements and live-provider final acceptance remain pending plan 35; no whole requirement is marked complete by this slice.

## Next Phase Readiness

Plan 11 can extract and adversarially refine the already enforced URL policy. The canonical contracts, actual routes, real-PG collision/rollback corpus, and browser draft/detail path are available for subsequent slices.

## Self-Check: PASSED

All five created implementation files and this summary exist. Both task commits resolve in repository history. Final gates passed before the GREEN commit.
