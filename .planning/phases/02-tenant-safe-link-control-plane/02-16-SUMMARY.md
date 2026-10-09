---
phase: 02-tenant-safe-link-control-plane
plan: "16"
subsystem: team
tags: [go, postgres, nextjs, capabilities, tenancy, pagination, tdd]
requires:
  - phase: 02-15
    provides: Scoped Links shell, current membership authority and strict canonical contracts
provides:
  - Fresh PostgreSQL owner/admin authorization for bounded Team inspection
  - Safe identity and role projection with complete UUID seek pagination
  - Responsive Team route with retry, role-loss disposal and stale-response rejection
affects: [02-17, 02-18, 02-32, 02-35]
tech-stack:
  added: []
  patterns: [workspace-first capability transaction, bounded UUID seek pages, role-keyed scoped presentation]
key-files:
  created:
    - apps/backend/internal/repository/team.go
    - apps/backend/internal/service/team.go
    - apps/frontend/app/workspaces/[workspaceId]/team/page.tsx
    - apps/frontend/components/team-list.tsx
  modified:
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/repository/repositories.go
    - apps/backend/internal/router/router.go
    - apps/backend/internal/service/services.go
    - apps/backend/internal/testing/browser_fixture_test.go
    - apps/backend/internal/testing/product.go
    - apps/backend/internal/transport/health.gen.go
    - apps/backend/static/openapi.json
    - apps/frontend/components/app-shell.tsx
    - apps/frontend/components/workspace-switcher.tsx
    - apps/frontend/lib/workspace.test.ts
    - apps/frontend/tests/control-plane.spec.ts
    - packages/openapi/openapi.json
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/src/gen.test.ts
    - packages/zod/src/identity.ts
key-decisions:
  - Read Team under current SQL owner/admin capability in the same workspace-first transaction as its fixed 25-row query; continuation never grants authority.
  - Treat Team 403 as capability loss: dispose rows and refresh current workspace role without broadcasting membership removal; preserve valid Links access.
  - Keep Team role enum names distinct in canonical schema metadata and regenerate both OpenAPI copies and Go transport together.
requirements-completed: []
requirements-addressed: [TEN-04, TEN-05, TEN-07, TEN-08, SAFE-03]
duration: 10min
completed: 2026-10-09
---

# Phase 2 Plan 16: Authorized Team Inspection Summary

**Current PostgreSQL owner/admin capability gates a real responsive Team page with safe identity/role projection, bounded complete pagination and immediate disposal after permission loss.**

## Performance

- Tasks: 2/2.
- Recorded resumed verification interval: 2026-10-09 23:21:14 UTC through approximately 23:30:45 UTC, about 10 minutes. Earlier interrupted execution and initial resumed context preparation are excluded; their duration was not retained. This is a measured resumed interval, not a reconstructed total.
- Actual implementation/test/generated inventory: 21 paths, four created and 17 modified, exhaustively listed above. Summary and STATE/ROADMAP tracking are additional documentation paths.
- Resumed from existing RED commit `3403f8d`; preserved partial implementation rather than restarting it.

## Accomplishments

- Registered GET `/api/v1/workspaces/:workspaceId/members` uses verified durable actor identity and current PostgreSQL Team capability. Workspace-first authorization and the member query share one bounded transaction. Owners/admins read; own-workspace members/viewers receive 403; foreign or removed membership receives safe 404. Provider organization claims never grant Flux authority.
- Parameterized SQL reads current memberships in the requested workspace and joins only their durable users. Projection contains exactly user UUID, workspace UUID, verified email and closed current role; provider identifiers, historical membership fields and other workspace identities are absent.
- Fixed 25-row pages plus one lookahead use stable user UUID ordering. Optional canonical UUID continuation is only a seek position and every request reauthorizes. Unknown, duplicate, malformed and oversized query input fails closed. Null continuation means exhaustion; pagination does not silently truncate larger teams. The existing 64 KiB success budget remains unchanged.
- Owners/admins receive a native Team navigation anchor and real scoped route. Desktop semantic identity/role rows and mobile cards use original native styling, approved 64px rows, token spacing and the below-768px breakpoint. Loading, successful empty, retry and exhaustion states remain distinct. No future invitation or mutation buttons are fabricated.
- Team 403 immediately disposes rows and the old workspace snapshot, then refreshes bootstrap/summary to reflect the current role while retaining valid Links access. Membership removal follows the existing 404 chooser and invalidation path. Session/workspace/request generations and role-keyed Team instances prevent late old responses from repainting restricted rows.
- Canonical identity schemas/contracts own the new strict DTO and route. Both OpenAPI copies and Go transport were regenerated together, with Team-specific enum names preventing collision with the existing generated Viewer constant.

## Task Commits

1. Task 1 — Specify scoped Team inspection and capability disposal: `3403f8d` (`test`, existing RED).
2. Task 2 — Deliver freshly authorized bounded Team inspection: `353b7fa` (`feat`, resumed GREEN).

## Verification

- Existing RED evidence preserved: registered real-PG Team owner/admin reads expected 200 but received 404; member/viewer reads expected 403 but received 404. Actual browser RED reached the real Links page and failed on the absent Team link, with no skipped/flaky/infrastructure cases accepted.
- Fresh full root integration gate passed **17 registered top-level tests across six groups**. Registered `TestProductActualHTTP/team` exercises all four roles with forged provider organization claims, foreign workspace 404, removed actor 404, exact safe projection, unsupported/duplicate/invalid queries, body budget and complete pagination over 31 members with no duplicates or foreign rows. Existing HTTP boundary and response/log privacy assertions remain intact.
- Fresh full root unit gate passed **95 discovered race-enabled Go unit tests**, plus every workspace, script and tool suite. Direct frontend units passed **22/22**. The canonical route inventory was extended with the real members endpoint while preserving health/auth/contract assertions.
- Original focused Team browser gate passed **2 completed / 0 skipped / 0 flaky / 0 unexpected / 0 infrastructure errors**. It proves real owner/admin inspection, member/viewer direct-route denial and native navigation removal, continued Links access, responsive 320px cards without horizontal overflow, truthful 503 retry/recovery, membership removal scrubbing, and a held older successful response that cannot repaint Team after a real SQL role change.
- Original full canonical browser gate passed **32 completed**, with unchanged strict zero-skip/flaky/infrastructure validation, preserving all previous 30 cases.
- Root format, lint, typecheck and deterministic generation gates passed. Migration gate passed with existing latest schema version 7 and upgrade coverage; no migration or runtime/gate files changed.
- Final root production build after the last browser run passed all workspace and backend role builds. Existing exported `cleanFrontendBuild` validated lstat/realpath and the production prerender manifest before removing disposable frontend output.
- `CI=true bun run scan` passed: Bun dependencies clean, no vulnerable imported Go packages across all roles/tests, complete worktree secrets clean and full Git history secrets clean. Existing unused-module inventory advisory **GO-2026-5932** for `golang.org/x/crypto@v0.57.0` remains visibly reported; no suppression, exception or gate weakening was added.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Consume narrow existing composition and presentation glue**
- Found during Task 2.
- The declared feature files required existing dependency registries/router and fixture composition to expose the real endpoint, and the workspace switcher to distinguish Team capability loss from membership removal.
- Extra paths: `apps/backend/internal/repository/repositories.go`, `apps/backend/internal/router/router.go`, `apps/backend/internal/service/services.go`, `apps/backend/internal/testing/browser_fixture_test.go`, `apps/frontend/components/workspace-switcher.tsx`, `apps/frontend/lib/workspace.test.ts`, `apps/backend/internal/testing/product.go`, and `packages/openapi/src/gen.test.ts`.
- Existing test-only loopback role mutation remains below application layers; its introduced conditional was converted to a switch to satisfy configured complexity checks. No production route, environment flag or authorization bypass was introduced.
- Previous assertions requiring an unavailable Team control were updated to assert the implemented native anchor and owner/admin-only visibility. The strict canonical route inventory adds only the new real endpoint.
- Verification: full unit/integration/browser, lint and type gates; commit `353b7fa`.

**2. [Rule 1 - Bug] Preserve canonical generation with distinct Team enum names**
- Found during Task 2 before interruption.
- New role enum generation collided with the existing generated Viewer constant.
- Fixed canonical `x-enum-varnames` metadata and regenerated all three owned outputs together; no handwritten transport DTO or new runtime dependency.
- Verification: generation drift, Go compilation, full type/build gates; commit `353b7fa`.

**3. [Rule 1 - Bug] Scope retry assertion to the actual Team content**
- Found during resumed Task 2 browser verification.
- Protected private report proved **1 expected / 1 unexpected / 0 skipped / 0 flaky**: the global alert locator matched both the correct Team retry notice and Next's route announcer.
- Scoped exact retry-copy and recovery-count assertions to the real Team main landmark. No production behavior or assertion was weakened.
- Diagnostic copy/report stayed in a protected OS temporary directory outside the repository; original fixture/config remained unchanged and the final focused/full runs used the original root gate.
- Verification: original focused 2/2 and full 32/32 browser gates; commit `353b7fa`.

## Issues Encountered

- The first fresh full integration run reported the previously deferred `TestTracestateHTTPRedisOTLP` failure. Its canonical runner exposed the test name and discarded the underlying assertion; that original diagnostic was unavailable. An isolated race-enabled run passed in 4.01 seconds, followed by one bounded full integration rerun passing all 17 registered tests. No telemetry source was changed and no repair is claimed; the existing deferred investigation remains relevant.
- Interrupted prior execution required fresh gates. Package scripts, generator authority, runtime pins, strict scanners and browser validator were preserved.

## TDD Gate Compliance

Existing RED `3403f8d` precedes GREEN `353b7fa`; both commits are present. No completed task was repeated or rolled back.

## Next Phase Readiness

Ready for plan 02-17 role mutation and final-owner invariants. Invitations and other future Team mutations remain allocated to later plans. Broad requirement completion and actual live-provider verification/recovery/social/session acceptance remain pending the final 02-35 gate; local signed provider fixtures do not satisfy that evidence.

No blocking stubs or new security surfaces outside this plan's Team trust boundaries were found.

## Self-Check: PASSED

All four created source files and this summary exist; RED and GREEN commits are present. All required gates passed as recorded above.

