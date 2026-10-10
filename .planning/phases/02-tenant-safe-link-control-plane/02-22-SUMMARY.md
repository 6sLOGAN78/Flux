---
phase: 02-tenant-safe-link-control-plane
plan: "22"
subsystem: invitations
tags: [go, postgres, redis, asynq, aes-gcm, recovery, tdd]
requires:
  - phase: 02-21
    provides: Independent durable worker and encrypted immutable provider requests
provides:
  - Workspace-first send serialization and durable first-dispatch recovery fence
  - Conservative expired-uncertainty blocking with encrypted reconciliation evidence
  - Actual Redis outage, forged tuple, terminal replay and key-rotation proof
affects: [02-23, 02-24, 02-25, 02-30, 02-35]
tech-stack:
  added: []
  patterns: [workspace-first send locking, durable provider-window fence, terminal ciphertext cleanup]
key-files:
  created:
    - apps/backend/internal/database/migrations/010_delivery_recovery.sql
    - packages/emails/src/templates/invitation.test.tsx
  modified:
    - .planning/phases/02-tenant-safe-link-control-plane/02-22-PLAN.md
    - apps/backend/internal/app/invitation_worker_test.go
    - apps/backend/internal/lib/email/client_test.go
    - apps/backend/internal/repository/delivery.go
    - apps/backend/internal/testing/browser_fixture_test.go
    - apps/frontend/tests/control-plane.spec.ts
key-decisions:
  - Persist the first possible provider exposure independently of acknowledgement commit; never renew its 23-hour retry window.
  - Hold workspace, invitation and intent locks in that order through the bounded external send; erase retained failed ciphertext when the invitation becomes terminal.
requirements-completed: []
requirement-slices: [TEN-03, TEN-07, TEN-08, SAFE-03, SAFE-04]
duration: 20min 31s
duration-basis: Measured resumed verification only; original execution, interruption and initial context restoration excluded
completed: 2026-10-10
---

# Phase 2 Plan 22: Invitation Delivery Recovery Summary

**Workspace-first sends, crash-durable provider-window fencing, expired-uncertainty blocking and actual Redis/key-rotation recovery preserve tenant-safe invitation delivery.**

## Performance

- Measured resumed verification: **20min 31s**, from 2026-10-10T21:36:25Z to 2026-10-10T21:56:56Z.
- This window starts at the first resumed private Go-run file creation. Earlier execution, interruption and initial context restoration are excluded; total original execution duration is unknown.
- Tasks completed: **2/2**.
- Concrete task paths: **8** across RED and GREEN, including the PLAN scope adjustment. Seven paths changed in GREEN.
- Continued existing RED commit and partial production implementation without discarding either.

## Accomplishments

- Preparation and sending acquire exclusive workspace, invitation and intent locks in the shared mutation order. A winning revocation prevents the old credential from reaching the sender; a send that wins prevents the competing workspace mutation until the bounded send finishes. Actual PostgreSQL transactions prove both orders.
- Forward migration 010 adds private first-dispatch and reconciliation metadata to the existing intent table; migrations 008/009 remain unchanged. Reservation commits independently before the effect transaction and survives a terminated acknowledgement transaction. Retries and lease generations never renew the first-dispatch timestamp.
- Automatic retries stop after a conservative 23-hour first-exposure window. The row becomes Failed with reconciliation_required=true and retains encrypted evidence; no renewed provider identity or uncontrolled send occurs. Legacy attempted/leased records use creation time as a conservative lower bound. Resend documents a 24-hour key lifetime and rejects changed request payloads under the same key. [Official Resend idempotency documentation](https://resend.com/changelog/idempotency-keys)
- Terminal accepted, revoked or expired invitations also clean retained failed ciphertext and clear reconciliation flags. Existing acknowledged delivery and atomic ciphertext erasure remain intact.
- Actual Redis tests execute known foreign workspace/invitation/delivery combinations, future and old generations, malformed/tampered tasks, accepted/revoked/expired intents and delivered duplicates. Completed/archive inspection proves execution and finite safe errors; no forged reference reaches the HTTP provider.
- An isolated actual Redis queue is deleted, its test-owned container is stopped, and publication demonstrably fails without spending a provider attempt. Restart and SQL lease recovery deliver the old-key pending envelope and a new-key write exactly once, with ciphertext erased. Real HTTP retry/restart checks preserve frozen bytes and provider identity across changed workspace presentation, sender and public origin.
- Native Team recovery observes an actual uncertain HTTP attempt, simulates a two-day outage, displays the approved Delivery failed state, records reconciliation blocking and proves explicit reload does not trigger another send. No acceptance, revoke or resend production HTTP endpoint was introduced.

## Task Commits

1. **Task 1: Specify reject forged worker envelopes and recover delivery across failures** — `0e7bff3` (test; preserved).
2. **Task 2: Deliver reject forged worker envelopes and recover delivery across failures** — `a68a4d1` (feat).

Normal commits were used without hook bypass. Neither task commit deletes tracked files.

## Verification

- Preserved original meaningful RED evidence: PostgreSQL winning-lock bypass, expired uncertain retry and terminal failed cleanup failures; native uncertain delivery remained Queued after an actual HTTP attempt and simulated outage. Infrastructure failures were not counted as feature RED. RED was not repeated during continuation.
- Expanded registered worker suite passed **15 direct scenarios**, with **27 total test/subtest pass events**, in 25.75 seconds. The outage/rotation case additionally passed its targeted rerun after correcting restart address discovery.
- Full root unit gate passed **103 selected race-enabled Go unit tests**, **28 frontend unit tests**, all canonical package and script suites and tool self-tests. Official email tests passed **6/6**, including deterministic invitation substitution and HTML/href escaping plus the preserved finite welcome CLI export test.
- Full root integration gate passed **18 registered tests across 6 groups**. Worker/role/resource, actual PostgreSQL/Redis and production HTTP checks remained intact. The existing tracestate test passed; no unrelated repair is claimed.
- Original targeted native delivery-recovery gate passed **1/1**, with strict positive-count and zero-skip/flaky/infrastructure validation.
- Final original full native browser gate passed **40/40**. Before that pass, one original full run failed with withheld output; the private native diagnostic subsequently passed **40/40** in 152.665 seconds with zero skipped/unexpected/flaky cases and zero infrastructure errors. No cause is inferred for the earlier unclassified full failure and no gate settings changed. The original confirmation run passed with builds serialized afterward.
- Final root strict lint, formatting and typechecking passed. Native Next unit/build/type/browser stages were serialized.
- Migration verification and five-artifact generator reproducibility passed. Both OpenAPI copies, generated Go transport and both email outputs retain exact byte authority.
- Final root production build passed. Existing cleanFrontendBuild validated ownership, nonsymlink containment and the production prerender manifest before disposing only its generated output.
- **CI=true full scan passed:** no vulnerable imported Go packages across roles/tests; Bun dependencies clean; worktree and full-history secrets clean. Existing unused GO-2026-5932 x/crypto inventory advisory remains visible, with no scanner exception.
- Private native diagnostics and Go reports stayed outside the repository; directory mode 0700 and report mode 0600. No credentials or live provider responses were published.

## Deviations from Plan

1. **[Rule 2 — Missing critical recovery metadata]** The original slice lacked a crash-durable first-exposure bound and expired-uncertainty reconciliation marker. Added forward migration 010 and independent reservation in delivery.go rather than rewriting earlier migrations. Actual PostgreSQL acknowledgement-transaction termination and expired retry tests verify the fence. Commits: `0e7bff3`, `a68a4d1`.
2. **[Rule 3 — Blocking browser evidence]** Added the absent recovery browser case and bounded external-package fixture endpoint to explicit PLAN scope. Extracted test-only evidence/time simulation helpers to preserve strict complexity checks. The fixture retains RED compatibility through column discovery and does not introduce production auth bypass. Original targeted/full gates passed. Commits: `0e7bff3`, `a68a4d1`.
3. **[Rule 3 — Fixture and assertion corrections]** Docker restart changed its ephemeral host port, so the isolated outage test re-reads the mapped address before reconnecting. Native diagnostics proved the approved failed-state copy is Delivery failed; corrected the new test's Failed wording. Formatting, field layout and domain-role test props were corrected without relaxing any lint rule. Commit: `a68a4d1`.

All adjustments stay within the eight recorded paths and the authorized recovery slice.

## Issues Encountered

The first isolated outage run failed at reconnecting to the old Redis mapped port; its targeted correction passed. Native targeted diagnostics identified an assertion wording mismatch after recovery had already succeeded. One original full browser failure remains unclassified; the private full diagnostic and final original full run passed without source or gate changes. No test was skipped, waived or accepted as flaky.

## Known Stubs

None prevent this recovery slice. Local HTTP acknowledgement and provider transport interception are test evidence; actual provider/inbox/session acceptance remains plan 02-35. Acceptance, resend and revoke product controls remain future plans 02-24/02-25. Their SQL mutation/send ordering is proven here without claiming those HTTP endpoints exist.

## Recovery Limits and Next Plan

Keep old encryption keys while any retained pending or failed ciphertext needs them. A future resend must use a new immutable intent identity/content generation; lease generation only fences processing. Expired uncertain rows are blocked for authorized operator reconciliation and must not be automatically requeued under a renewed key/window. Migration rollback refuses to discard an established dispatch fence.

Ready for **02-23**. Broad requirements remain pending; this plan addresses their worker/recovery slices and does not mark TEN-03/TEN-07/TEN-08/SAFE-03/SAFE-04 complete. Live provider evidence remains 02-35. The unrelated .serena/project.yml change was left untouched and unstaged. No push or plan 23 execution occurred.

## Self-Check: PASSED

Both created task files and this SUMMARY exist. RED 0e7bff3 and GREEN a68a4d1 exist; their eight-path union has no tracked deletions. All required positive gates passed, no generated untracked files remain, and unrelated Serena changes are unstaged.

