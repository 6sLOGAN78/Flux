---
phase: 02-tenant-safe-link-control-plane
plan: "17"
subsystem: team
tags: [go, postgres, nextjs, roles, audit, idempotency, tdd]
requires:
  - phase: 02-16
    provides: Freshly authorized bounded Team inspection and capability-loss disposal
provides:
  - Scoped current-role mutation with serialized final-owner protection
  - Atomic role effect, protected audit and 24-hour actor/workspace/operation replay ledger
  - Native safe-first role confirmation and committed self-demotion capability refresh
affects: [02-18, 02-32, 02-35]
tech-stack:
  added: []
  patterns: [exclusive workspace then actor and target membership locks, current authorization before replay, atomic protected audit]
key-files:
  created:
    - apps/backend/internal/repository/audit.go
    - apps/backend/internal/service/team_test.go
  modified:
    - apps/backend/internal/repository/team.go
    - apps/backend/internal/service/team.go
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/router/router.go
    - apps/backend/internal/testing/product.go
    - apps/backend/internal/transport/health.gen.go
    - apps/backend/static/openapi.json
    - apps/frontend/app/workspaces/[workspaceId]/team/page.tsx
    - apps/frontend/components/team-list.tsx
    - apps/frontend/components/confirm-dialog.tsx
    - apps/frontend/lib/api.ts
    - apps/frontend/tests/control-plane.spec.ts
    - packages/zod/src/identity.ts
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/src/gen.test.ts
    - packages/openapi/openapi.json
key-decisions:
  - Reuse the existing closed AllowsRoleChange policy and protect owner count in the same exclusive workspace transaction as the role effect.
  - Reauthorize current actor and target before reading replay hashes; original member replay snapshots never replace current actor capability projection.
  - Reuse existing immutable audit_events and mutation_requests tables with durable user references; no migration or membership-target foreign key is needed.
  - Require explicit native confirmation and reload after mutation failure; committed self-demotion disposes Team through the existing capability-loss boundary while preserving Links.
requirements-completed: []
requirements-addressed: [TEN-04, TEN-05, TEN-06, TEN-07, TEN-08, SAFE-03]
duration: 30min
completed: 2026-10-10
---

# Phase 2 Plan 17: Owner-Preserving Role Changes Summary

**Current PostgreSQL authority controls scoped role changes with serialized final-owner protection, atomic immutable audit and replay, and an accessible explicit confirmation flow.**

## Performance

- Tasks: 2/2.
- Recorded execution/verification interval: 2026-10-09 23:35:58 UTC through 2026-10-10 00:06:06 UTC, 30 minutes 8 seconds. Initial context loading precedes this interval; final documentation, tracking and history verification follow it. No interrupted interval or invented dispatch time is included.
- Actual implementation/test/generated inventory: 19 paths, two created and 17 modified, exhaustively listed above. Summary and STATE/ROADMAP tracking are additional documentation paths.

## Accomplishments

- Registered PATCH `/api/v1/workspaces/:workspaceId/members/:memberId` accepts canonical durable user UUID and closed role input. Existing production bearer/session, exact Origin, JSON, body limit, rate-limit and no-store boundaries remain active. Unknown fields, null/invalid roles, trailing JSON, query parameters and invalid identifiers fail closed.
- The service reuses the existing `AllowsRoleChange` matrix. Owners manage all four roles and promote existing members to owner. Admins may target and grant member/viewer only; they cannot manage admin/owner targets or their own admin membership. Members/viewers and forged provider organization owners cannot mutate roles.
- One transaction locks workspace exclusively, then current actor membership, then scoped target membership. The actor is rechecked after its row lock. Owner count is read inside that transaction before demotion, so concurrent current owners cannot both leave the workspace ownerless. The final-owner failure is 409 `OWNER_REQUIRED`, with exact copy “Promote another owner first.” Ownership transfer explicitly promotes an existing member before demoting an owner.
- Current actor and target authorization precede all replay/hash reads. SHA-256 binds canonical target UUID and role. Workspace/durable actor/`member.role`/key scope retains committed snapshots at least 24 hours; identical authorized retries replay, changed content conflicts, and removed or demoted actors cannot recover old responses. Replay projects current actor capability alongside the original member snapshot.
- Role effect, ledger and protected audit commit together. Existing audit records contain workspace, durable user actor, fixed operation, historical membership target UUID snapshot and server timestamp. Existing immutability triggers reject audit update/deletion; actor/creator provenance never references removable membership rows. No schema migration was necessary.
- Team presents closed role options from the current workspace role, owner-preservation guidance, and a native named “Change role?” dialog. “Keep current role” receives initial focus; Escape cancels and restores trigger focus. Management demotion uses the approved exact explanation. UI updates only after a validated committed response. Failures require explicit reload rather than automatic mutation retry. Confirmed self-demotion clears Team and refreshes authoritative workspace role through plan 16's boundary, retaining valid Links access.
- Canonical identity schemas and route contract remain the sole authored API authority. Both OpenAPI copies and Go transport were regenerated together with distinct role enum names. Team GET keeps its safe four-field identities, fixed 25-row UUID seek pages, authoritative exhaustion and existing 64 KiB response budget.

## Task Commits

1. Task 1 — Specify role matrix, owner preservation and native confirmation: `197d6b2` (`test`).
2. Task 2 — Deliver scoped role mutation, atomic owner/audit/replay protection and UI: `2ffdf51` (`feat`).

## Verification

- Meaningful HTTP RED used registered real migrated PostgreSQL `TestProductActualHTTP/roles`: absent PATCH returned 404 rather than expected 200/403, and final-owner demotion returned 404 rather than 409. No compile or fixture error was accepted as RED.
- Meaningful browser RED reached the actual Team page and failed because “Promote another owner first” was absent: 0 passed, 1 unexpected, 0 skipped/flaky/infrastructure errors. An earlier diagnostic mistakenly supplied a custom runner that suppressed native fixture startup; its beforeEach 404 was rejected. A private OS-temporary copy then retained failed `runCommand` stdout internally while preserving native fixture startup and strict report validation. The original gate stayed unchanged.
- Full root unit gate passed **97 discovered race-enabled Go unit tests** and all workspace/script/tool suites. Added unit proof rejects unknown/provider roles and malformed role/key/target input before store access. Direct frontend units passed **22/22**.
- Full root integration gate passed **17 registered top-level tests across six groups**. The registered roles dispatcher includes **64 actual HTTP role matrix subtests**, foreign target 404, invalid input, last-owner 409, identical replay/changed-hash conflict, independent actor and workspace key scopes, audit exactly once, required protected audit fields, durable actor provenance after membership removal, and immutable audit rejection.
- Real PostgreSQL synchronization holds the workspace lock until both current-owner demotions are waiting; release yields exactly one 200 and one 409, with one owner remaining. Another pending operation waits behind a transaction that commits actor demotion, then receives 403 before changed-hash replay. Deferred lock cleanup and error-returning concurrent HTTP helpers keep assertions on the test goroutine.
- A deferred constraint trigger fails at commit and proves complete rollback: target role unchanged, no ledger row and no additional audit. Its private failure marker never reaches public logs.
- Focused native browser GREEN passed **1/1** new role flow. Original canonical `--grep roles` gate passed **2/2** selected cases. Original full browser gate passed **33 completed / 0 skipped / 0 flaky / 0 unexpected / 0 infrastructure errors**, preserving the previous 32 cases.
- Browser proof covers actual final-owner denial, current role retention, safe focus, Escape/focus return with no mutation, exactly one explicit 503 submission without automatic retry, successful promotion, approved self-demotion confirmation, immediate Team disposal and continued Links access. Existing responsive Team, all-role direct denials, stale-response disposal and workspace switching cases remain passing.
- Root format, lint, typecheck, generation byte check, migration check and final production root build passed. After browser execution, the final production build passed and the existing exported strict `cleanFrontendBuild` validated lstat/realpath/production prerender metadata before removing disposable output.
- `CI=true bun run scan` passed the complete worktree and full Git history, Bun dependencies and all-role/test imported Go package exposure. Existing unused-module `GO-2026-5932` inventory advisory remains visible; no vulnerable package is imported. Final metadata history scan follows the tracking commit.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Connect the endpoint, current role projection and executable fixture/contract inventory**
- The declared files omitted necessary existing composition and verification seams.
- Narrow additional paths: `apps/backend/internal/router/router.go` registers PATCH; `apps/frontend/app/workspaces/[workspaceId]/team/page.tsx` passes current actor role; `apps/frontend/lib/api.ts` admits only canonical status-matched OWNER_REQUIRED feedback; `apps/backend/internal/testing/product.go` seeds one real colleague through the existing test-only loopback protocol; `packages/openapi/src/gen.test.ts` verifies the canonical implemented route/status/security inventory.
- No gate/runtime/dependency changes, production fixture route or broad refactor. Committed in `197d6b2` and `2ffdf51`; verified by the unchanged root gates and actual browser/HTTP tests.

**2. [Rule 1 - Test fixture] Correct race keys and guarantee failure cleanup**
- Initial race keys contained 15 characters, so validation rejected them before workspace locks. The wait assertion then left its fixture lock open and delayed pool cleanup.
- Used valid 16-character keys, deferred rollback for both fixture locks, and transport helpers that return errors to the test goroutine. Actual synchronized concurrency then passed. Committed in `2ffdf51`.

**3. [Rule 1 - Test expectations] Account for the implemented PATCH route and native role options**
- Existing Team GET-on-member assertion now correctly expects unsupported-method 405 instead of absent-route 404. No member GET projection was added.
- Existing inspection test continues asserting no invitation/removal controls; role controls are now intentionally implemented. Role assertions target the displayed table/card role rather than matching native option text. Security denial, stale-response and responsive assertions remain intact. Committed in `2ffdf51`.

## Issues Encountered

- New-code lint findings were resolved with a smaller transactional effect helper, explicit persisted snapshot JSON tags, the existing shared role policy, nonshadowing locals, formatted SQL and error-returning concurrent HTTP helpers. No linter or gate was weakened.
- The initial full integration failure was the intentional 404-to-405 method expectation change, diagnosed from protected private actual-HTTP output. The complete canonical integration gate then passed. No telemetry recurrence or telemetry repair is claimed.

## User Setup Required

None for this slice. Actual configured provider factor, recovery, OAuth, session/cookie and delivery acceptance remain pending final plan 02-35; local injected provider transport is not live-provider evidence.

## Next Phase Readiness

- Ready for plan 02-18 member removal using the same workspace/actor/target ordering and durable audit provenance.
- Invitations, removals, editing, domains, redirects, analytics, billing and destination fetching were not introduced.
- Broad phase requirements remain incomplete until their remaining slices and final acceptance are verified; `requirements-completed` intentionally remains empty.

## Self-Check: PASSED

- Both created files exist, and all 19 actual code/test/generated paths are committed.
- RED `197d6b2` and GREEN `2ffdf51` exist in Git in the required order.
- No tracked file deletions, unwired production stubs or new unregistered integration test names were introduced. The new endpoint/audit surface is covered by this plan's threat model.
