---
phase: 02-tenant-safe-link-control-plane
plan: "18"
subsystem: team
tags: [go, postgres, nextjs, removal, audit, idempotency, tdd]
requires:
  - phase: 02-17
    provides: Serialized workspace/actor/target role mutations and durable audit/replay storage
provides:
  - Scoped member removal with final-owner protection and durable link creator provenance
  - Freshly authorized missing-target replay using protected committed target-role snapshots
  - Native safe-first removal confirmation and immediate self/cross-tab access disposal
affects: [02-19, 02-32, 02-35]
tech-stack:
  added: []
  patterns: [exclusive workspace then actor and target locks, protected snapshot authorization before replay hash, membership-only deletion]
key-files:
  created:
    - apps/frontend/lib/team.ts
    - apps/frontend/lib/team.test.ts
  modified:
    - apps/backend/internal/repository/team.go
    - apps/backend/internal/repository/audit.go
    - apps/backend/internal/service/team.go
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/router/router.go
    - apps/backend/internal/testing/product.go
    - apps/backend/internal/transport/health.gen.go
    - apps/backend/static/openapi.json
    - apps/frontend/app/workspaces/[workspaceId]/team/page.tsx
    - apps/frontend/components/team-list.tsx
    - apps/frontend/components/workspace-switcher.tsx
    - apps/frontend/tests/control-plane.spec.ts
    - packages/zod/src/identity.ts
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/src/gen.test.ts
    - packages/openapi/openapi.json
key-decisions:
  - Missing-target removal replay uses its own committed target-role snapshot only after fresh locked actor authorization, with target policy checked before reading the hash.
  - Delete only memberships and reuse durable user, immutable audit and 24-hour ledger references; existing link provenance needs no migration or link repository change.
  - Send fixed invalidation through the mutating tab's existing owned channel and explicitly dispose self-removal locally; fresh server membership remains authoritative.
requirements-completed: []
requirements-addressed: [TEN-04, TEN-05, TEN-06, TEN-07, TEN-08, SAFE-03]
duration: 24min
completed: 2026-10-10
---

# Phase 2 Plan 18: Durable Member Removal Summary

**Current PostgreSQL authority controls member removal while preserving active/deleted links, creator identity and key reservations, with atomic audit/replay and immediate browser access disposal.**

## Performance

- Tasks: 2/2.
- Recorded execution/verification interval: 2026-10-10 00:10:39 UTC through 00:34:25 UTC, 23 minutes 46 seconds. Initial context loading precedes this interval; documentation and resumed closeout follow it. Interruptions during closeout are excluded.
- Actual implementation/test/generated inventory: 19 paths, two created and 17 modified, exhaustively listed above. Summary and STATE/ROADMAP tracking are additional documentation paths.

## Accomplishments

- Registered DELETE `/api/v1/workspaces/:workspaceId/members/:memberId` accepts a canonical durable user UUID and strict empty JSON body. Existing authentication, exact Origin, JSON/body limit, rate-limit and no-store boundaries remain active. Invalid identifiers, unknown fields, null/trailing bodies and query parameters fail closed.
- Reused the closed role policy: owners can remove any role, admins can remove member/viewer targets, and members/viewers cannot remove. SQL membership controls authority; provider organization claims do not grant permissions.
- One transaction exclusively locks the workspace, freshly authorizes/locks the current actor, then locks the scoped target. Final-owner protection runs inside this transaction. Concurrent owner removal and demotion produce exactly one successful effect and one 409 `OWNER_REQUIRED`, retaining an owner. Self-removal succeeds only when another owner remains.
- Fresh actor authorization precedes ledger access. Live target policy and protected committed snapshot role policy precede request-hash reads. A still-authorized actor can replay its original successful removal after the target row is gone; a new missing target returns 404. Changed target content conflicts, an admin cannot replay removal of a former owner, and removed/demoted actors cannot recover old responses. Ledger scope is workspace/durable actor/`member.remove`/key with at least 24-hour retention.
- Membership deletion, immutable protected audit and replay snapshot commit together. Audit records retain durable actor, fixed action, historical target membership UUID and server timestamp without a foreign key to removable membership. Existing durable user references preserve active and soft-deleted links, creator UUID/display provenance, canonical URLs, counts and deleted-key reservations. No user, link or key is deleted and no schema migration was needed; existing link repository joins already provide safe creator projection.
- Team uses a native named “Remove member?” dialog with “Keep member” initial focus, Escape cancellation and trigger focus return. Final-owner feedback uses “Promote another owner first”. Mutations update only after strict committed-response validation. Failure requires explicit reload and never automatically retries destructive requests.
- Successful third-party removal broadcasts the existing fixed invalidation signal through the mutating tab's owned channel. The removed tab scrubs and rechecks fresh server membership. Successful self-removal explicitly disposes local Team state and returns all affected tabs to the workspace chooser while preserving the signed-in session. Delayed real Links responses cannot restore removed workspace contents.
- Canonical identity schemas/contracts remain the sole authored API authority. Both OpenAPI copies and Go transport were regenerated together; existing enum names and role mutation behavior remain intact.

## Task Commits

1. Task 1 — Specify actual HTTP removal/provenance and browser confirmation: `e02c28e` (`test`).
2. Task 2 — Deliver authorized removal, durable provenance, atomic replay/audit and browser disposal: `ed5909e` (`feat`).

## Verification

- Meaningful registered real PostgreSQL HTTP RED covered all 16 actor/target role combinations: absent DELETE returned 405 instead of expected 200/403; foreign target returned 405 instead of 404 and final-owner removal returned 405 instead of 409. No fixture or compilation failure was accepted as RED.
- Meaningful actual browser RED reached Team and failed on the absent “Remove member” button: one existing selected case passed, one unexpected failure, zero skips/flaky/infrastructure errors. An earlier overlapping diagnostic/canonical browser run caused a port collision and was rejected as infrastructure evidence. Serial native-fixture diagnostics then captured the actual absent behavior. A private mode-700 OS-temporary copy retained failed native `runCommand` stdout into mode-600 JSON; original gate, fixture startup and strict validator remained unchanged.
- Root unit gate passed **97 race-enabled Go unit tests** and all workspace/script/tool suites. Direct frontend units passed **24/24**, including closed removal policy and strict cross-resource/private-snapshot response rejection.
- Root integration gate passed **17 registered top-level tests across six groups**. Actual removal HTTP proof covers all 16 role combinations, scoped foreign/missing targets, strict input, last-owner denial, self-removal, authorized missing-target replay, changed-hash conflict, independent actor/workspace scopes and former-actor denial.
- Actual PostgreSQL proof retains both active and soft-deleted creator links with original UUIDs, URLs, counts and deleted-key reservation; the removed creator loses old read/write/replay access while another authorized member retains safe creator projection. Audit/ledger exactly-once and retention checks, immutable audit rejection and deferred commit-failure rollback pass.
- Synchronized real PostgreSQL mixed owner removal/demotion retains one owner. A waiting replay observes committed actor-membership removal and is denied before old replay recovery. Snapshot policy denies an admin's replay of former-owner removal even with an identical hash.
- Focused native browser GREEN and original canonical removal gate passed **3/3** selected cases. Original full browser gate passed **35 completed / 0 skipped / 0 flaky / 0 unexpected / 0 infrastructure errors**, preserving previous 33 cases.
- Browser proof covers safe focus/Escape/return, actual last-owner denial, explicit 503 submission without automatic retry, successful self-removal, same-browser separately signed colleague removal, fresh server access denial, retained creator link and disposal of delayed actual Links responses.
- Root format, lint, typecheck, generation byte check, migration check and final production root build passed. After full browser execution, the final root build passed and existing exported strict `cleanFrontendBuild` validated lstat/realpath/production prerender metadata before removing owned disposable output.
- `CI=true bun run scan` passed complete worktree/full Git history, Bun dependencies and all-role/test imported Go exposure. Existing unused-module inventory advisory `GO-2026-5932` remains visible; no vulnerable package is imported. Final metadata history verification follows tracking closeout.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Wire existing endpoint, audit, channel and executable fixture/contract seams**
- Six narrow additional paths were needed: `router/router.go` registers DELETE; `repository/audit.go` adds the fixed removal audit helper; the Team page passes the existing channel callback; `workspace-switcher.tsx` exposes invalidation through its owned channel; `testing/product.go` provides a separately signed test-only colleague session; `openapi/src/gen.test.ts` verifies the canonical route/security/schema inventory.
- All additional paths are listed above. No gate/runtime/dependency changes, production fixture endpoint, new tables or broad refactor. Committed in `ed5909e` and verified through unchanged root gates.

**2. [Rule 1 - Existing expectation] Preserve implemented controls and unsupported member GET semantics**
- Team inspection now checks that invitations remain absent while removal controls are intentionally present. The plan-17 member GET expectation remains 405. No new member GET route or invitation feature was added.

## Issues Encountered

- New-code lint findings were corrected by extracting bounded replay/effect helpers, ordering private snapshot fields, sharing the existing actor authorization/error mapping, using a typed fixture session builder and formatting SQL/test locals. No linter was weakened.
- The installed contract generator omits `requestBody.required` metadata for existing operations. The new contract test asserts canonical strict empty-object JSON schema and retains route/status/security/idempotency checks rather than asserting unsupported generator metadata.
- No recurrence or repair of the previously deferred telemetry failure is claimed.

## User Setup Required

None for this slice. Actual configured provider factor, recovery, OAuth, session/cookie and delivery acceptance remain pending plan 02-35; local injected signed sessions do not satisfy live-provider acceptance.

## Next Phase Readiness

- Ready for plan 02-19. It was not started.
- Broad phase requirements remain incomplete until remaining slices and final acceptance pass; `requirements-completed` intentionally remains empty.
- No unwired production stubs remain. The new DELETE and audit surface is covered by this plan's threat register.

## Self-Check: PASSED

- Both created files exist; all 19 actual code/test/generated paths are committed in `ed5909e`.
- RED `e02c28e` and GREEN `ed5909e` exist in Git in the required order.
- No tracked files were deleted, no new integration dispatcher was left unregistered, and the worktree was clean before summary creation.
