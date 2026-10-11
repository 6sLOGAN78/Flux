---
phase: 02-tenant-safe-link-control-plane
plan: "25"
subsystem: invitations
tags: [postgresql, invitations, idempotency, react, delivery, security]
requires:
  - phase: 02-24
    provides: Fresh verified-recipient inspection and atomic explicit acceptance
  - phase: 02-23
    provides: Frozen encrypted delivery requests and bounded provider replay window
provides:
  - Scoped invitation resend with fresh credentials and a new immutable delivery intent
  - Atomic revocation with stale credential and delivery generation fences
  - Native safe-first confirmations and explicit recovery after uncertain mutations
affects: [02-26, 02-30, 02-35]
tech-stack:
  added: []
  patterns: [current-intent history, scoped versioned mutation ledger, explicit uncertain-delivery recovery]
key-files:
  created:
    - apps/backend/internal/database/migrations/011_invitation_generations.sql
  modified:
    - apps/backend/internal/repository/invitation.go
    - apps/backend/internal/repository/delivery.go
    - apps/backend/internal/service/invitation.go
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/router/router.go
    - apps/backend/internal/transport/health.gen.go
    - apps/backend/static/openapi.json
    - apps/frontend/components/invitation-panel.tsx
    - apps/frontend/components/confirm-dialog.tsx
    - apps/frontend/tests/control-plane.spec.ts
    - apps/frontend/lib/team.test.ts
    - packages/zod/src/identity.ts
    - packages/zod/src/identity.test.ts
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/src/gen.test.ts
    - packages/openapi/openapi.json
key-decisions:
  - Invitation body versions are exact PostgreSQL integers 1..2147483647; absent versions return 428, stale versions return 409, and overflow is denied.
  - Resend allocates a new immutable intent identity and retains historical ACK and first-dispatch evidence; old provider identities never carry new content.
  - Fresh actor and original protected-role policy precede replay hash disclosure; uncertain mutations require explicit reload and Failed resend warns about prior delivery.
requirements-completed: []
requirement-slices: [TEN-03, TEN-04, TEN-05, TEN-06, TEN-07, TEN-08, SAFE-02, SAFE-03]
duration: 28min 46s
completed: 2026-10-11
---

# Phase 2 Plan 25: Invitation Rotation and Revocation Summary

**Scoped resend rotates one-use credentials into a new immutable delivery intent, while atomic revocation fences acceptance and delivery races.**

## Performance

- Measured verification and execution window: 28min 46s, from the first retained native RED start at 2026-10-11T05:58:51.357Z through final verification at 2026-10-11T06:27:37Z. Initial preflight/read time was not timed and is excluded; there were no interruptions in this window.
- Tasks: 2/2.
- Source/test/artifact paths: 18, including one new migration and 17 existing files; one additional PLAN scope update. No tracked deletions.

## Accomplishments

- Added actual scoped POST resend and DELETE revoke routes with canonical generated DTOs, strict body validation, fresh SQL authorization, workspace/actor/invitation/intent locking, bounded versions and a scoped 24-hour operation ledger. Protected original-role authorization precedes replay hash comparison.
- Resend generates a fresh 256-bit credential digest and encrypted seven-day envelope, checks pending-email uniqueness, and atomically records its new immutable intent, audit and replay result. Migration 011 preserves historical intents and adds one unique current intent. Its reverse migration refuses destructive history loss.
- Revocation and acceptance invalidate lease generations and clear terminal private envelopes atomically. List, claim, prepare, dispatch reservation and delivery paths fence stale intents. Resend preserves original invitation creation time while acceptance validates the new issued-at lifetime. Existing first-dispatch timing and conservative provider replay limits remain enforced.
- Native Resend/Revoke confirmations use safe-first focus, Escape and focus return. Failed delivery has an explicit uncertainty warning; transport failure disables further mutations and automatic polling until an explicit authoritative reload. Queued remains visible until worker acknowledgement.

## Task Commits

1. Task 1, specify invitation management boundaries: `27e39e1` — meaningful actual PostgreSQL HTTP RED and native missing-action RED.
2. Task 2, implement rotation and revocation: `0207380` — implementation, race/rollback proof, canonical artifacts and recorded necessary scope extensions.

## Verification

- Focused real PostgreSQL race and management cases passed: old-token acceptance denial; fresh intent identity; original replay and hash conflict; lost authorization and protected-role replay; absent/null/fractional/stale/foreign versions; 24-hour ledger retention; accepted/revoked terminal denial; pending-email uniqueness; overflow rejection; transactional rollback; concurrent resends; accept/resend/revoke races; winning-revoke stale callbacks; and bounded Send linearization with retained ACK history.
- Canonical full unit gate passed: 106 selected Go race tests, 28 frontend tests and 6 email tests, plus contract/tooling suites.
- Canonical integration gate passed: 20 tests across 6 registered groups. Target native invitation management passed 1/1; canonical full native browser gate passed 43/43.
- Format, lint, typecheck, production build, migration checks and all five generated-artifact reproducibility checks passed. Production build cleanup used the existing owned-output validation before scanning.
- `CI=true` full scan passed: no vulnerable imported Go packages across backend roles/tests, clean Bun dependencies, and clean secret scans of both worktree and Git history. Unused GO-2026-5932 remains visible as inventory evidence without an exception.
- Required independent review was read-only and reported no actionable findings.

## Decisions Made

Invitation generation uses bounded PostgreSQL integer body versions. The RESEARCH bigint/decimal-string guidance is scoped to Link lifecycle; this invitation representation remains exact throughout its admitted range. Missing preconditions return 428, stale or terminal mutations return 409, and the maximum version cannot increment or wrap.

Resend never reuses an old frozen provider request identity for new credential content. A partial unique current-intent index allows retained historical delivery and audit references. Revoked invitations remain terminal. Explicit Failed resend acknowledges that an earlier message may have arrived and invalidates its credential before queuing a new message.

## Deviations from Plan

1. **[Rule 2 — Missing critical functionality]** Added narrow router registration and forward migration 011 to make the planned routes reachable and preserve immutable intent history. Both are recorded in PLAN ownership; no earlier migration was rewritten.
2. **[Rule 3 — Blocking fixtures]** Updated the exact OpenAPI path inventory and existing schema/frontend fixtures for the mandatory canonical invitation version. These three necessary test paths are recorded in PLAN ownership. All three generated API artifacts were regenerated from the authored Zod/contract authority.
3. **[Rule 1 — Bug]** The executable transport-failure recovery case exposed enabled mutation buttons after uncertain failure. Corrected the controls and mutation guard, then passed the canonical target, type/lint/format and full browser gates. No hidden resend or polling retry substitutes for explicit recovery.

Total: three bounded deviations, all included in `0207380`. Five necessary source/test paths extend the original 13-path scope, for 18 source/test/artifact paths plus the PLAN update. No unrelated feature work.

## Issues Encountered

Native runner failures suppress subprocess diagnostics. A private mode-0700 OS temporary copy retained mode-0600 reports while preserving native fixture startup, report validation and teardown. An initial diagnostic instrumentation placement was corrected in that private copy only. Meaningful RED was the absent route/action, not fixture authentication or diagnostics setup.

One early canonical target run and the original first full browser run failed without retained assertions; subsequent private runs passed, so their causes remain unproven. No speculative product change was attributed to them. A separate retained target report proved the enabled-button recovery defect described above. The final canonical target and full 43-test gate passed; the private full report also recorded 43 expected, zero skipped, unexpected or flaky tests and zero report errors. Native acceptance did not use retries, skips, custom fixtures or runner bypasses.

## Known Stubs

None. Changed production paths are wired to real SQL authorization, persistence and delivery behavior. Stub-pattern and threat-surface scans found no unfinished behavior or additional surface outside this plan's invitation endpoints/schema boundary.

## Next Phase Readiness

Plan 26 can build on versioned invitations and immutable intent history. Broad phase requirements remain pending: `requirements-completed` intentionally stays empty. Live Clerk verification/recovery/Google/GitHub and protected Resend configuration with an approved recipient remain the mandatory plan-35 external acceptance checkpoint. Local fixture proof is not external acceptance; no real email was sent.

## Self-Check: PASSED

Verified the created migration exists, both task commits exist, the initial-to-task-2 diff contains exactly 18 source/test/artifact paths and one PLAN update, and no tracked files were deleted. All required gates passed before completion tracking. The unrelated `.serena/project.yml` change remains unstaged and untouched.
