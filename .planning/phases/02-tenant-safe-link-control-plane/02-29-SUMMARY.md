---
phase: 02-tenant-safe-link-control-plane
plan: "29"
subsystem: api
tags: [postgres, suspension, authorization, idempotency, react, audit, concurrency]
requires:
  - phase: 02-28
    provides: Scoped lifecycle transactions and explicit conflict review with immutable uncertain retries
provides:
  - Owner/admin suspension and explicit clear to Disabled with fresh SQL authorization
  - Atomic immutable suspension audit reason, verified actor and trusted database time
  - Native restriction dialogs with retained drafts, explicit review and exact uncertain retry
affects: [02-30, 02-31, 02-32, 02-34, 02-35]
tech-stack:
  added: []
  patterns: [shared scope lock ordering, immutable restriction attempts, normalized Unicode reason validation]
key-files:
  created:
    - apps/backend/internal/database/migrations/013_link_suspension_audit.sql
  modified:
    - apps/backend/internal/service/link.go
    - apps/backend/internal/repository/link.go
    - apps/backend/internal/repository/audit.go
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/router/router.go
    - apps/backend/internal/transport/health.gen.go
    - apps/backend/static/openapi.json
    - apps/frontend/components/link-detail.tsx
    - apps/frontend/components/confirm-dialog.tsx
    - apps/frontend/tests/control-plane.spec.ts
    - packages/zod/src/identity.ts
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/src/gen.test.ts
    - packages/openapi/openapi.json
key-decisions:
  - Retain historical suspension and clear reasons in the existing immutable audit table through a forward nullable-column migration.
  - Normalize reasons with Unicode White_Space, count Unicode codepoints and reject NUL consistently across server, canonical schema and dialog.
  - Refresh SQL capabilities after a known committed mutation without turning a permission-refresh failure into mutation uncertainty.
requirements-completed: []
requirement-slices: [LINK-09, LINK-05, LINK-07, TEN-05, TEN-07, TEN-08, SAFE-03]
duration: 37min 01s
duration-basis: Earliest reliable captured timestamp through verified GREEN commit; earlier setup and initial RED elapsed are unknown and excluded, as is metadata close-out
completed: 2026-10-11
---

# Phase 2 Plan 29: Audited Link Suspension Summary

**Current SQL owners/admins can suspend links or explicitly clear restrictions to Disabled, with atomic immutable provenance and recoverable native dialogs.**

## Performance

- Measured window: **2026-10-11T20:22:23Z** through **2026-10-11T20:59:24Z**, **37min 01s**. Earlier setup and initial RED elapsed were not reliably captured and are excluded; metadata close-out is excluded.
- Tasks: **2/2**. Source/test paths changed: **16**, including three declared bounded ownership additions. PLAN classification and SUMMARY are separate metadata.

## Accomplishments

- POST and DELETE on the scoped suspension route require verified server identity, current locked SQL owner/admin membership, a positive BIGINT decimal-string version and an idempotency key. Workspace, actor membership and scoped link locks follow the existing order. Deleted links reject restriction effects; foreign existing targets remain hidden. Replay reauthorizes before ledger/hash disclosure.
- Reasons use shared Unicode White_Space trimming and 1–1000 Unicode codepoints. Server decoding rejects duplicate or unknown fields, forged actor/time properties, null values, invalid UTF-8 and trailing JSON. NUL is explicitly rejected by both methods, the canonical schema and the dialog because PostgreSQL text cannot store it. OpenAPI describes normalized constraints without an incorrect raw UTF-16 length limit.
- Suspension stores independent current reason/actor/time provenance; immutable audit events retain historical suspend and clear reasons using forward migration 013. Effect, audit and replay ledger commit together. Clear requires a reason and sets Disabled. Ordinary archive, disable, restore and delete preserve restrictions, and a valid-version direct Enable cannot bypass suspension.
- Native Suspend/Clear dialogs use exact consequence copy, safe initial focus, Escape cancellation and trigger focus return. Stale conflicts refresh authoritative state and SQL capabilities, retain reason drafts and require explicit review before a new key/version intention. A lost committed clear response can be retried only with the original immutable key, reason and version. Successful commits refresh capabilities so revoked management controls disappear; a refresh error never misclassifies a known commit as uncertain.
- Real PostgreSQL tests cover role denial, foreign targets, removed-member replay, strict input, deleted links, lifecycle bypass, audit immutability, same-version concurrency and audit/ledger/deferred-commit rollback. Browser proofs cover escaped reason rendering, privacy canaries, conflict review, exact retry and fresh permissions. No reason or actor canary reaches collected telemetry.

## Task Commits

1. **Task 1: Specify restriction authority, provenance and dialogs** — `b494466` (RED), expanded by `0d277a6` (RED concurrency, foreign-scope and immutable audit proof).
2. **Task 2: Deliver audited suspension and explicit clear** — `a57154e` (GREEN).

Exact path staging and normal commits were used. No tracked deletion, package install, gate change, runtime-pin change, push or plan 30 execution occurred. Unrelated `.serena/project.yml` remains unstaged.

## Verification

- Initial real PostgreSQL RED reached the actual suspension route and received 404 instead of 200. The canonical targeted browser RED lacked the Suspend action. The unchanged native runner's private diagnostic capture confirmed executed cases with no skipped/flaky acceptance; it has been removed.
- Review-driven RED proofs caught stale manager controls after a committed mutation and SQL role loss, Unicode NEL trimming divergence, contradictory raw OpenAPI length, and duplicate-property acceptance. Focused tests subsequently passed. A late concrete NUL regression produced expected 400 versus actual 503 before the consistent validator correction; focused actual PostgreSQL GREEN passed afterward.
- Focused race-enabled real PostgreSQL lifecycle/suspension proof passed. The two original targeted suspension browser cases passed before broad verification. The final full browser suite covers both without an additional repeat.
- Final original unit gate passed **108 selected race-enabled Go tests**, **31 frontend**, **6 email**, **15 OpenAPI**, and schema/script/tool suites.
- Final original integration gate passed **20 registered tests across 6 groups**, including the expanded registered actual HTTP PostgreSQL dispatcher and unchanged worker ownership cases.
- The first original canonical full browser run passed **53/53**. After the NUL validation change, the final original full browser run again passed **53/53**, with no skipped, unexpected, flaky or global errors. Local fixture proofs do not establish live-provider acceptance.
- Final original formatting, strict lint, typechecking, production build and migration verification passed. `generate:check` verified all **five artifacts**: both OpenAPI copies, generated Go transport and both email templates. Canonical suspension generation changed all three API artifacts together.
- Existing `cleanFrontendBuild` validated the owned nonsymlink production output/manifest before removal. **CI=true scan passed locally** for imported Go vulnerabilities across roles/tests, Bun dependencies, full worktree secrets and full Git history secrets. Unused **GO-2026-5932** remains visible inventory evidence; no scanner exception was added.
- `git diff --check` passed. GREEN has no tracked deletions; only the pre-existing Serena modification remains. No hosted CI result is claimed here.

## Decisions Made

The existing current-link provenance columns could not retain reasons after clear. A forward nullable reason column in the already immutable audit table preserves historical provenance without a new table or rewriting migrations.

Reason normalization uses Unicode characters consistently across Go and TypeScript. Explicit rejection handles values PostgreSQL cannot represent. A known mutation commit stays known while fresh SQL capability resolution controls subsequent actions. Uncertainty retains the exact original request; conflict ends that attempt and requires explicit review.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Declare reachable route wiring and canonical regression ownership**
- Declared `router/router.go` and `packages/openapi/src/gen.test.ts` in PLAN before production edits so both methods are reachable and exact route/schema inventory remains enforced.

**2. [Rule 2 - Missing critical functionality] Preserve historical immutable audit reasons**
- Declared and added forward `013_link_suspension_audit.sql` before production edits. Existing 005 current provenance fields remain unchanged. No new table or architecture was introduced.

**3. [Rule 1/2 - Correctness and security] Resolve reviewed boundary defects**
- Added fresh post-commit capability resolution, exact Unicode normalization, accurate OpenAPI constraints, strict duplicate-field decoding and explicit NUL rejection. Meaningful failing assertions preceded corrections, followed by focused proof and final original gates. All implementation fixes are in `a57154e`.

Broad requirement IDs were classified as requirement slices in PLAN; phase-wide completion remains pending final plan 35. No architectural deviation or authentication gate occurred.

## Issues Encountered

Initial strict lint identified new complexity, alignment, shadowing, repeated constants and test-goroutine assertion issues. Bounded helper extraction and naming/formatting corrections resolved them. No assertion, timeout, test count, scanner policy or lint rule was weakened. A stale package build was rebuilt before treating canonical-schema assertions as business RED evidence.

The concrete late NUL storage failure justified rerunning the affected focused proof and final full canonical gate chain after the first full pass. No further unguided scope was added.

## Known Stubs

No suspension, authorization, provenance or recovery stubs were introduced. Existing detail copy about redirects and analytics reflects later phases and does not block this control-plane slice.

## User Setup Required

None for this slice. Live provider recovery/OAuth and real recipient delivery acceptance remain pending plan 02-35. No provider configuration, secrets or live email were accessed.

## Next Phase Readiness

Plan 29 is complete and the suspension boundary is available for later control-plane verification. Phase-wide requirements remain open. No plan 30 work or phase transition was performed.

## Self-Check: PASSED

All 16 source/test files exist, and RED commits `b494466`, `0d277a6` and GREEN `a57154e` were verified as Git commit objects before this summary was written. Final gate results and the untouched unrelated modification were checked.
