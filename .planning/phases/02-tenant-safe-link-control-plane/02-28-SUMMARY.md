---
phase: 02-tenant-safe-link-control-plane
plan: "28"
subsystem: ui
tags: [react, postgres, lifecycle, idempotency, concurrency, accessibility]
requires:
  - phase: 02-27
    provides: Versioned scoped lifecycle transitions with atomic effect, audit and replay ledger
provides:
  - Authoritative conflict refresh followed by explicit review and a new mutation intention
  - Explicit uncertain-result retry preserving identical key, payload and displayed decimal version
  - Native dialog keyboard proof and real PostgreSQL rollback, replay and Deleted restore assertions
affects: [02-29, 02-31, 02-34, 02-35]
tech-stack:
  added: []
  patterns: [immutable scoped mutation attempts, explicit conflict review, generation-guarded focus]
key-files:
  created: []
  modified:
    - apps/backend/internal/handler/product_test.go
    - apps/frontend/components/link-detail.tsx
    - apps/frontend/lib/links.ts
    - apps/frontend/lib/links.test.ts
    - apps/frontend/tests/control-plane.spec.ts
key-decisions:
  - Keep an uncertain lifecycle attempt immutable until explicit identical retry or deliberate replacement; conflict refresh never resubmits its action.
  - Reuse the existing scoped lifecycle transactions, Deleted filter and native dialogs; prove their recovery invariants through real PostgreSQL and browser paths.
requirements-completed: []
requirement-slices: [LINK-05, LINK-07, TEN-05, TEN-07, TEN-08, SAFE-03]
duration: 15min 33s
duration-basis: Measured resumed execution through verified GREEN commit; earlier interrupted reads and metadata close-out overhead are excluded
completed: 2026-10-11
---

# Phase 2 Plan 28: Lifecycle Conflict Review Summary

**Lifecycle conflicts now refresh authoritative details before explicit review, while uncertain committed mutations can be retried with their original immutable key, payload and version.**

## Performance

- Measured execution: **2026-10-11T20:00:25Z** through **2026-10-11T20:15:58Z**, **15min 33s**.
- Tasks: **2/2**. Source/test paths modified: **5**. PLAN requirement classification and SUMMARY are separate metadata.
- No plan 28 source changes existed before the interrupted session. No restart downtime is included.

## Accomplishments

- A stale version triggers a scoped detail GET and fresh workspace membership resolution. Updated authoritative version/state are displayed while mutation controls remain disabled. The stale-state announcement remains until explicit review; choosing another action then creates a new key using the reviewed displayed version. Archive is exercised as the stale destructive action, with no automatic repeat.
- A network-uncertain result retains one immutable attempt in disposable component memory. Only the explicit retry button repeats its exact key and serialized payload, including the original decimal version. A deliberate new intention gets a new key; workspace/request disposal clears the attempt. No success is claimed until a committed response validates.
- Existing native Archive/Delete dialogs retain approved consequence text, safe initial focus, Escape cancellation and return focus to the trigger. After uncertain mutation, focus moves to the retry button only after the current React commit. Missing-link refresh rechecks the workspace before declaring workspace access lost.
- Real PostgreSQL assertions prove ledger-insert and deferred-commit failures roll back lifecycle/version, audit and ledger together; identical retries then commit once and replay the same response. Same-version concurrency produces one 200 and one 409 with one version increment, audit and ledger. Viewer/removed-member replay is denied before hash disclosure, with no additional effects.
- Deleted links are absent from the default list and retrieved by the actual Deleted filter. Native restore returns Disabled, retains creator and short URL, and uses a new key. The existing valid-version suspended-enable direct-route denial is paired with a canonical browser presentation assertion for the disabled control and explanatory copy.

## Task Commits

1. **Task 1: Specify lifecycle conflict review and uncertain retry** — `9ebc0c8` (RED).
2. **Task 2: Deliver explicit review and immutable uncertain retry** — `7d0312e` (GREEN).

Exact staging and normal commit hooks were used. No tracked deletions, runtime/gate changes, package installs, pushes or plan 29 execution occurred. Unrelated `.serena/project.yml` remains unstaged.

## Verification

- RED: original canonical targeted browser runner failed. Its unchanged native private-capture runner executed **2 cases**, both failing on the intended missing behavior: displayed version remained **1 instead of 2**, and the identical retry button was absent after a real committed delete response was lost. There were **0 skipped/flaky cases**. The private diagnostic was removed after recording these assertions. The new helper's preimplementation export failure was not used as business RED evidence.
- Focused race-enabled actual PostgreSQL lifecycle command passed twice, including **27 nested lifecycle cases**. Focused link helper tests passed **5/5**. Original targeted browser subsequently passed **2/2**.
- Final original unit gate passed **108 selected race-enabled Go tests**, **31 frontend**, **6 email**, **14 OpenAPI**, and schema/script/tool suites.
- Final original integration gate passed **20 registered tests across 6 groups**, including unchanged worker ownership tests and all expanded real PostgreSQL lifecycle assertions.
- First original full browser gate passed **51/51**, with **0 skipped, unexpected, flaky or global errors**. It includes the final stronger stale Archive/new-version request assertions, lost-response replay, keyboard focus, Deleted restore and suspension presentation. Suspension presentation uses a canonical response projection; its actual server denial is proved separately with PostgreSQL. Local fixtures are not live-provider acceptance evidence.
- Final formatting, strict lint, typechecking, production build and migration verification passed. `generate:check` passed all **five artifacts**: both OpenAPI copies, generated Go transport and both email templates. No canonical artifact changes were needed.
- Existing `cleanFrontendBuild` validated the owned nonsymlink production output and manifest before removal. **CI=true scan passed locally** for all-role/test imported Go vulnerabilities, Bun dependencies, full worktree secrets and full Git history secrets. Unused **GO-2026-5932** remained visible as inventory evidence; no scanner exception was added.
- `git diff --check` passed. GREEN has no tracked deletions, and only the pre-existing unrelated Serena modification remained afterward.

## Decisions Made

Uncertainty retains the original immutable request rather than adopting a refetched version. Conflict definitively ends that attempt, refreshes authoritative state and requires explicit review before creating another intention. No draft-editing or suspension-reason flow is introduced in this slice; existing detail values and confirmation context stay intact.

Backend production lifecycle, reservation, authorization, transaction and dialog mechanisms from plan 27 already satisfied the required invariants. This slice changes only the browser recovery behavior and strengthens executable proof within the declared paths.

## Deviations from Plan

No ownership expansion or architectural deviation. Five of the eight declared source/test paths needed changes; the existing service, mutation repository and ConfirmDialog were reused unchanged. Broad `requirements` were classified as `requirement-slices` in PLAN so incomplete phase-wide requirements are not prematurely closed before final plan 35.

## Issues Encountered

The initial lint gate reported one new test-local shadowed `code` variable. Renamed it to `listCode`; the final original lint gate passed. No test assertions, gate limits, runtime pins or scanner policies were weakened.

No authentication gate occurred. All plan 28 final canonical gates passed; no hosted CI result is claimed by this local execution.

## Known Stubs

No recovery or lifecycle stubs were introduced. The pre-existing detail availability copy at `apps/frontend/components/link-detail.tsx:226` states that redirects and analytics are unavailable; those capabilities are intentionally assigned to later phases and do not block this plan's control-plane goal.

## User Setup Required

None for this slice. Live provider verification, recovery/OAuth and real recipient delivery acceptance remain pending final plan 02-35; no provider configuration, credentials or live email were accessed.

## Next Phase Readiness

Plan 29 can build suspension operations on the preserved versioned lifecycle boundary. Broad phase requirements remain open until final acceptance. No plan 29 work or transition was performed here.

## Self-Check: PASSED

All five modified source/test files exist. Both task commits `9ebc0c8` and `7d0312e` exist. Final original verification gates passed, no tracked files were deleted, and no generated or diagnostic files remain untracked.
