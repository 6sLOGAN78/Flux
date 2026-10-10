---
phase: 02-tenant-safe-link-control-plane
plan: "21"
subsystem: invitations
tags: [go, postgres, redis, asynq, aes-gcm, nextjs, tdd]
requires:
  - phase: 02-20
    provides: Encrypted durable intents and canonical scoped delivery-state projection
provides:
  - Independent PostgreSQL/Redis invitation worker with bounded durable retries
  - Encrypted immutable provider requests and atomic acknowledged delivery/ciphertext erasure
  - Actual HTTP/Redis/PostgreSQL and native Team polling delivery evidence
affects: [02-22, 02-23, 02-24, 02-25, 02-35]
tech-stack:
  added: []
  patterns: [SQL-authoritative leases, encrypted frozen provider requests, opaque Redis references, acknowledgement fencing]
key-files:
  created:
    - apps/backend/internal/app/invitation_worker_test.go
    - apps/backend/internal/database/migrations/009_delivery_erasure.sql
    - apps/backend/internal/lib/invitationcrypto/crypto.go
    - apps/backend/internal/lib/job/invitation_handlers.go
    - apps/backend/internal/lib/job/invitation_tasks.go
    - apps/backend/templates/emails/invitation.html
    - packages/emails/src/templates/invitation.tsx
  modified:
    - apps/backend/internal/app/api.go
    - apps/backend/internal/app/lifecycle_test.go
    - apps/backend/internal/app/roles_test.go
    - apps/backend/internal/app/worker.go
    - apps/backend/internal/config/config.go
    - apps/backend/internal/config/config_test.go
    - apps/backend/internal/config/invitation_test.go
    - apps/backend/internal/handler/health_test.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/lib/email/client.go
    - apps/backend/internal/lib/job/job.go
    - apps/backend/internal/repository/delivery.go
    - apps/backend/internal/repository/repositories.go
    - apps/backend/internal/service/invitation_crypto.go
    - apps/backend/internal/service/services.go
    - apps/backend/internal/testing/browser_fixture_test.go
    - apps/frontend/components/invitation-panel.tsx
    - apps/frontend/tests/control-plane.spec.ts
    - scripts/check.ts
    - scripts/generate.ts
    - scripts/generated.test.ts
    - .planning/phases/02-tenant-safe-link-control-plane/02-21-PLAN.md
key-decisions:
  - PostgreSQL owns eight total attempts and lease recovery; Redis publication failures do not consume a send attempt and invitation Asynq retries are disabled.
  - Freeze every provider request field inside the existing encrypted envelope before publication; immutable intent IDs identify content and lease generations only fence processing.
  - Commit Delivered and ciphertext erasure atomically after acknowledgement, preserving retryable state for errors and stale acknowledgements.
  - Keep old encryption keys while retained ciphertext needs them; future resend must replace the intent ID or introduce an explicit immutable content generation.
requirements-completed: []
requirements-addressed: [TEN-03, TEN-07, TEN-08, SAFE-03, SAFE-04]
duration: 47min
completed: 2026-10-10
---

# Phase 2 Plan 21: Independent durable invitation delivery Summary

**An independently owned PostgreSQL/Redis worker delivers encrypted invitations through a bounded provider acknowledgement boundary, freezes retry request bytes, and atomically erases ciphertext when durable state becomes Delivered.**

## Performance

- Recorded implementation/verification interval: 2026-10-10T15:53:41Z through 2026-10-10T16:40:40Z, 46m59s (rounded 47min). Context loading is excluded; no interrupted interval was counted.
- Tasks: 2/2.
- Task scope: 29 unique paths, including 28 application/configuration/test/asset/gate paths and the PLAN scope correction. SUMMARY and STATE/ROADMAP tracking are additional documentation paths.

## Accomplishments

- Worker owns PostgreSQL, Redis, the email adapter and consumer through explicit factories and reverse cleanup. Its readiness requires PostgreSQL/Redis plus locally rendered embedded email assets. API producer independence and redirector independence remain tested; no implicit migrations or Clerk/managed-host/cursor requirements were added to worker configuration.
- The existing job module dispatches every five seconds, transactionally claims at most 25 intents with SKIP LOCKED and a 60-second fenced lease, and publishes only workspace/invitation/delivery IDs and generation. Failed publication leaves durable SQL intent recoverable without spending a send attempt.
- Preparation and handling re-read tenant-scoped pending state, expiry and current lease generation. Complete recipient/sender/subject/HTML/idempotency fields are frozen inside the existing AES-256-GCM envelope before Redis publication. Workspace, template or worker configuration changes cannot alter a retried request using the same provider key.
- PostgreSQL owns eight total attempts with bounded exponential delay; invitation Asynq retries are disabled. Provider sends have ten-second deadlines. Duplicate references serialize on the intent lock, and stale/future/foreign/malformed references cannot send. Success commits Delivered and deletes ciphertext atomically; failures remain Queued or terminal Failed, and stale acknowledgement cannot optimistically deliver.
- The invitation adapter fixes HTTPS Resend origin, disables environment proxies and redirects, uses a bounded owned transport, and accepts success only with a nonempty provider acknowledgement. Existing FLUX_INTEGRATION.RESEND_API_KEY and invitation key names remain unchanged. Test injection is explicit construction, with actual HTTP acknowledgement rather than an environment bypass.
- Invitation TSX and checked HTML are generated together through the fifth declared artifact. Rendering uses the existing closed embedded-template enum, escapes workspace/role values, and puts the token in the configured public-origin /invitations#token= fragment. PostgreSQL/Redis/log scans reject plaintext token/envelope leakage.
- The actual Team panel polls scoped GETs for at most 60 seconds, keeps existing abort/generation/capability disposal, rearms after transient GET failure, and stops for terminal state or disposal. It does not retry POST automatically. The native delivery browser case creates a real invitation, holds HTTP acknowledgement while server/UI remain Queued, observes one GET 503, then proves actual worker acknowledgement, server/UI Delivered and erased ciphertext.

## Task Commits

1. **Task 1: Specify independent durable invitation delivery** — `8aa47dc` (test).
2. **Task 2: Deliver encrypted invitations through the independent worker** — `84034f6` (feat).

Both commits exist. Normal commit commands were used without hook bypass; no tracked files were deleted.

## Verification

- Meaningful real-PG/Redis RED failed because the old worker did not own PostgreSQL. The original browser attempt failed; an owner-only private copy of the native runner then classified the actual missing behavior with selectively preserved/restored owned production bytes: real POST/list showed Queued, but Delivered timed out. Compile, fixture and infrastructure failures were not accepted as feature RED.
- Independent review identified two correctness issues. Actual HTTP RED demonstrated changed request bytes after workspace changes under one provider key; GREEN additionally restarts the worker with changed sender/origin and proves unchanged retry bytes. Native browser RED demonstrated polling halted after GET 503; targeted native GREEN passed 1/1 with eventual actual Delivered, no skips/flaky/infrastructure errors.
- Expanded registered TestInvitationWorkerActualRedisPostgres covers held acknowledgement, ciphertext erasure, Redis/log privacy scans, actual ten-second HTTP timeout, immutable restart payloads, malformed/future/tampered tasks, duplicate leases, eight-attempt terminal failure, expiry/revocation/tenant/generation fences, publication crash recovery, cancellation and stale acknowledgement.
- Final original root unit gate passed all 102 selected race-enabled Go unit tests, 28 frontend unit tests and all package/script/tool suites. Affected Go unit precheck separately passed app 19, config 15, handler 8, email 3 and jobs 2. Existing service encryption tests passed in the full gate.
- Full registered integration gate passed 18 tests across six groups, increased from 17 by explicit invitation worker registration. App/worker lifecycle and PostgreSQL ownership tests, actual health failures and real product HTTP projection all passed. TestTracestateHTTPRedisOTLP did not recur; no repair claim is made.
- Initial full integration found a fixture CASE expression inferred text for bytea; explicit bytea casts fixed it, the isolated actual-PG product test passed, and the complete integration rerun passed. Unit gates caught remaining legacy worker ownership/config expectations and an accidentally altered standalone Redis cancellation expectation; these were corrected, affected prechecks passed, and the final full unit gate passed.
- Final root typecheck, strict lint, formatting, generation reproducibility and migration checks passed. Production role decoding was narrowly factored to meet complexity checks; only established test-suite complexity exceptions are used. Existing generated byte/failure tests remain strict with the invitation manifest entry added.
- The unchanged original root full browser gate passed 39/39 under strict zero-skip/retry/flaky/infrastructure validation. No private full diagnostic or full browser retry was needed.
- Final production root build passed after browser completion. Existing exported cleanFrontendBuild validated lstat/realpath and the production prerender manifest before removing only disposable frontend build output.
- CI=true full scan passed imported Go package, Bun dependency, worktree secret and full-history secret checks. GO-2026-5932 remains visible as the existing unused x/crypto inventory advisory; no vulnerable package is imported across roles/tests. Runtime pins, provider asset cache, browser parser and scanner policy remain unchanged.
- Native diagnostic files stayed outside the repository in an owner-only OS temporary directory; reports are mode 0600. Private reports were used only to classify actual failures, with no provider credentials or raw provider responses published.

## Decisions and Recovery Limits

The immutable delivery-intent UUID identifies message content. Lease generation is a concurrency fence, not a content version. Future resend must create a new intent identity or introduce an explicit immutable content generation so new tokens/messages do not collide with provider deduplication. Retain keys needed by queued, leased or retained failed ciphertext until those envelopes drain or are safely re-encrypted.

Provider acknowledgement followed by a SQL crash remains uncertain. Recovery retries the same frozen request and provider key; deduplication beyond the provider's retention window cannot be guaranteed. Terminal Failed retains encrypted material for later controlled recovery/cleanup. Migration 009 permits terminal erasure without weakening queued/leased envelope constraints; rollback refuses to recreate erased ciphertext and must not be used after erasure without an explicit recovery strategy.

## Deviations from Plan

1. **[Rule 3 — Blocking] Actual browser delivery proof:** plan prose incorrectly assumed a delivery browser case already existed. Added the real worker/test-only HTTP acknowledgement fixture, native control-plane case and bounded consumed panel refresh. Added backend testing/browser_fixture_test.go, frontend control-plane.spec.ts and invitation-panel.tsx to explicit scope. Verified meaningful missing-delivery RED, targeted GREEN and full 39-case native browser gate. Commit: `84034f6`.
2. **[Rule 3 — Blocking] Resource expectations and injection cycles:** updated worker-owned configuration, readiness, lifecycle and resource graph fixtures now rather than defer to 02-23. Shared authenticated envelope helper plus compatibility wrappers avoid the job/service/server cycle; the repository registry accepts its existing owned pool, with the sole service caller remaining nil-safe. Additional app/config/handler/shared-crypto/service/registry paths are listed in PLAN and frontmatter. Verified actual worker ownership RED, affected unit prechecks and full unit/integration gates. Commit: `84034f6`.
3. **[Rule 2 — Correctness] Atomic cipher erasure:** added forward migration 009 instead of retroactively editing 008; terminal rows can erase ciphertext while queued/leased rows still require a valid envelope. Updated actual product projection fixtures to preserve ciphertext when restoring Queued. Verified acknowledged SQL erasure, stale acknowledgement, migration and actual HTTP gates. Commit: `84034f6`.
4. **[Rule 2 — Correctness] Stable request identity and transient refresh:** independent review found retry rendering could change bytes under the same provider key and GET failure could halt polling. Freeze all semantic provider fields before publication in the existing encrypted format; rearm only current/live scoped reads without resetting the overall polling budget. Verified both actual RED failures and worker restart/browser GREEN. Commit: `84034f6`.
5. **[Rule 3 — Blocking] Generator authority manifest:** added scripts/generated.test.ts to scope so the existing exact manifest requires all five generated outputs and keeps byte/failure checks intact. No unrelated generated files or gate policies changed. Commit: `84034f6`.

## Independent Review and Remaining Acceptance

An independent read-only review found no high-severity tenant, token, SSRF, lease or acknowledgement defect; both reported P2 correctness findings above were fixed and meaningfully verified. Its suggested live provider/inbox acceptance remains 02-35 evidence. Existing welcome-email transport behavior was outside invitation scope and was preserved. The unrelated user/tool .serena/project.yml modification was left untouched and unstaged.

## Known Stubs

None prevent independent durable invitation delivery. Initial empty UI collections read actual scoped API data. Local fake provider acknowledgement is explicitly test evidence, not live Resend/inbox acceptance. Invitation acceptance, resend/revoke controls and later product domains were not implemented; broad requirements remain incomplete.

## Self-Check: PASSED

All seven created application/test/asset files exist; task commits 8aa47dc and 84034f6 exist. The 29 task paths are explicitly recorded, all required positive gates executed without skipped/flaky/infrastructure acceptance, and no tracked deletions or generated untracked outputs remain. Task work is clean; unrelated .serena/project.yml remains modified and unstaged.

## Hosted CI Repair Addendum

Hosted run 38068839933 failed at @flux/emails:test after earlier generation, formatting, lint, build and type gates passed. The root runner intentionally withheld subprocess stdout, so the hosted assertion is unknown. Local official email tests passed both in the existing checkout and a clean archived checkout with frozen Bun dependencies; these passes alone did not explain the hosted failure.

The unchanged official command in that clean checkout under CPU contention reproduced a concrete failure: the existing locked CLI byte test reached Bun's default 5,000ms case limit, its export child was killed, and the command failed after 6.13s. That test performs two exports, each already bounded to 60 seconds, but lacked a matching outer test budget. This is proven local timeout behavior consistent with hosted timing, not confirmation of the hidden hosted assertion.

Commit `47d4177` adds an explicit finite 130-second timeout only to packages/emails/src/templates/welcome.test.tsx's existing two-export case. All byte comparisons, substitution-token assertions and 60-second child deadlines are unchanged. Under the same clean-checkout CPU contention, all four email tests passed; the byte case took 6.203s, exceeding the old limit and demonstrating the repair. Email formatting, strict lint and typecheck passed, root generate:check completed successfully with no byte mismatches, and CI=true full scan passed with the same visible unused GO-2026-5932 inventory advisory. No tests were waived or skipped, and no gate, scanner, runtime, provider or workflow policy changed.

This test-only timing repair does not repeat the previously completed 39-case browser or backend integration gates. The original 47-minute execution metric is unchanged; interruption time is excluded from repair evidence. Phase position remains 22/35 and no plan 22 work began. Root must push this repair and observe a new full hosted CI run before proceeding; hosted success is not yet claimed. The unrelated .serena/project.yml change remains untouched and unstaged.

Repair self-check passed: the scoped test file and this addendum exist, commit 47d4177 exists, and no tracked files were deleted.
