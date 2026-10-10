---
phase: 02-tenant-safe-link-control-plane
plan: "20"
subsystem: invitations
tags: [go, postgres, nextjs, aes-gcm, tdd]
requires:
  - phase: 02-19
    provides: Atomically queued invitations, encrypted delivery intents, fresh SQL authority and native Team form
provides:
  - Consumed InvitationPanel with scoped canonical status and full local expiry timestamps
  - Durable delivery-state projection with terminal invitation-state precedence
  - Key rotation, fail-closed decryption, input bounds, replay and concurrent expiry proof
affects: [02-21, 02-24, 02-25, 02-35]
tech-stack:
  added: []
  patterns: [scoped unique delivery join, strict canonical response projection, retained-key rotation]
key-files:
  created:
    - apps/frontend/components/invitation-panel.tsx
    - packages/zod/src/identity.test.ts
  modified:
    - apps/backend/.env.sample
    - apps/backend/internal/config/config_test.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/repository/invitation.go
    - apps/backend/internal/service/invitation_crypto_test.go
    - apps/backend/internal/transport/health.gen.go
    - apps/backend/static/openapi.json
    - apps/frontend/app/workspaces/[workspaceId]/team/page.tsx
    - apps/frontend/lib/team.ts
    - apps/frontend/lib/team.test.ts
    - apps/frontend/tests/control-plane.spec.ts
    - docs/development.md
    - packages/openapi/openapi.json
    - packages/zod/src/identity.ts
    - .planning/phases/02-tenant-safe-link-control-plane/02-20-PLAN.md
key-decisions:
  - Project delivery state from the unique workspace/invitation intent; expiry, acceptance and revocation take precedence over delivery.
  - Extend the actual canonical status enum with Delivered and Failed and regenerate all three artifacts together; no worker sending is claimed.
  - Retain old encryption keys until live ciphertext drains or is safely re-encrypted; count encryptions across replicas and rotate before the per-key random-nonce limit.
requirements-completed: []
requirements-addressed: [TEN-03, TEN-05, TEN-07, TEN-08, SAFE-03]
duration: 20min
completed: 2026-10-10
---

# Phase 2 Plan 20: Invitation status and encryption boundaries Summary

**The actual Team form consumes an extracted InvitationPanel, displays scoped durable invitation/delivery states and full local expiry timestamps, and has meaningful key-rotation, replay and concurrency proof.**

## Performance

- Recorded implementation/verification interval: 2026-10-10T11:08:12Z through 2026-10-10T11:27:42Z, 19m30s (rounded 20min). Initial context loading is excluded; no interrupted interval was counted.
- Tasks: 2/2.
- Task paths: 17 unique paths: 16 application/configuration/test/generated paths and one PLAN scope correction. SUMMARY and STATE/ROADMAP tracking are additional documentation paths.

## Accomplishments

- Extracted the consumed Team form without replacing its keyed workspace/role disposal, abort/generation checks, or existing dirty draft Stay/Discard confirmation. Fresh owner/admin capability determines whether the form is reachable; owner offers admin/member/viewer, admin offers member/viewer, and member/viewer have no invitation form.
- Strict shared projections reject foreign workspace rows, private fields, guessed Sent state and malformed responses. Team presents Queued, Delivered, Delivery failed, Accepted, Expired and Revoked; exact machine timestamps stay in time elements while visible expiry includes full local date, time and timezone.
- The PostgreSQL list joins intents by both workspace and invitation identifiers. Existing composite uniqueness prevents duplicate projections. Expired, accepted and revoked invitations take precedence; leased work remains Queued. Same invitation/delivery IDs in another tenant cannot affect the result.
- Extended the canonical enum with distinct InviteDelivered/InviteFailed names and regenerated both OpenAPI copies plus Go transport together. Existing routes, enum constants, migration 008 and configuration names remain compatible.
- Added retained-key rotation and retirement rejection, missing/malformed keys, truncated ciphertext with no returned plaintext, random nonce independence, bounded private diagnostics and maximum email length proof. Existing tampering and all workspace/invitation/delivery/key-ID AAD tests remain intact.
- Actual HTTP/PostgreSQL tests inspect encrypted rows and digest-only acceptance storage without exposing token material, verify canonical normalized replay hashes, actor-removal replay denial, at least 24-hour ledger retention, input/log canaries and concurrent replacement of an elapsed pending invitation.
- Documented the actual external FLUX_INVITATIONS.ENCRYPTION_KEYS JSON ring, exact 32-byte canonical Base64 keys, 32-key/16-KiB bounds, private provisioning, aggregate per-key 2^32 encryption limit and retirement only after live ciphertext drains or safe re-encryption. Go 1.26.9's NewGCMWithRandomNonce implementation remains unchanged; no custom nonce scheme was introduced.

## Task Commits

1. **Task 1: Specify invitation status and encryption boundaries** — `6718ebe` (test).
2. **Task 2: Expose scoped invitation status and encryption proof** — `62717ca` (feat).

Both commits exist; no tracked files were deleted. Hooks ran normally.

## Verification

- Registered real-PG HTTP RED: TestProductActualHTTP/invitations/durable_status_projection expected Delivered and received Queued from a durably delivered intent.
- Native root browser RED completed one invitation-queue case, zero skips/flaky results; actual queue creation succeeded before the missing status label assertion failed. Infrastructure, compile or fixture failures were not accepted as RED.
- Canonical schema RED executed two cases: the new durable-state acceptance case failed on Delivered, and private/unsafe-state rejection passed. Both cases passed after the canonical enum extension.
- Root unit gate passed 102 selected race-enabled Go unit tests and 28 frontend unit tests; schema, OpenAPI, emails, scripts and tool tests also passed. The compiled discovery/strict execution manifests remained intact.
- Full registered integration gate passed 17 tests across six groups. The invitation HTTP suite was rerun after test-helper extraction and passed. Previously deferred TestTracestateHTTPRedisOTLP did not recur; no repair claim is made.
- Targeted native invitation-queue browser gate passed 3/3, zero skips/flaky results. These are actual queue/draft behavior, explicitly labeled canonical status presentation fixture, and fresh SQL role controls. The presentation fixture is not worker/provider delivery evidence.
- The unchanged original root full browser gate passed 38/38 with the existing strict zero-skip/retry/flaky parser. A private native diagnostic copy also passed 38/38; the original gate is the authoritative final result. No CDN/cache change was needed.
- Root typecheck, strict lint, final formatting, generation reproducibility and migration checks passed. Introduced test complexity/shadowing was resolved by narrow helper extraction; the component uses the required import type form, with no new lint suppressions.
- Final production root build passed after the browser gate. Existing exported cleanFrontendBuild validated its production prerender manifest and removed only the disposable frontend build before scanning.
- CI=true full scan passed imported Go package, Bun dependency, worktree secret and full-history secret gates. GO-2026-5932 remains visible only as the existing unused x/crypto module inventory advisory; no vulnerable package is imported across backend roles/tests.
- Native diagnostic reports stayed outside the repository in an owner-only OS temporary directory, with files restricted to mode 0600. Gate source, parser, fixture startup, provider-asset cache, runtime pins and scanner policy have no committed changes.

## Decisions and Rulings

- Use the minimal additional repository path for truthful state projection under Rules 2–3; without it, persisted Delivered/Failed would still be shown as Queued. Cost if wrong: misleading state or tenant leakage; actual-PG state and same-ID tenant tests cover the boundary.
- Correct the plan's stale assumption that all status values already existed. The actual authored enum lacked Delivered/Failed, so add them centrally and regenerate all three outputs. Cost if wrong: contract/build disagreement; schema, generation, typecheck and browser gates verify agreement.
- Preserve actual FLUX_INVITATIONS.ENCRYPTION_KEYS rather than introduce ENCRYPTION_KEYS_JSON, and use the existing apps/backend/.env.sample. Cost if wrong: broken operator provisioning; role-owned configuration and existing environment binding tests remain passing.
- Independently reviewed execution results/artifact reproducibility are established by this executor's gates; worker sending and live-provider acceptance remain later evidence. Cost of treating presentation fixtures as delivery would be false acceptance, so no such claim or requirement completion is made.

## Deviations from Plan

1. **[Rule 2 — Correctness] Scoped delivery projection:** minimal apps/backend/internal/repository/invitation.go join/precedence was necessary beyond the original declared list. Verified through real-PG HTTP RED→GREEN, terminal-state precedence and identical foreign-tenant IDs. Commit: `62717ca`.
2. **[Rule 3 — Blocking] Canonical status and consuming tests:** actual schema lacked Delivered/Failed, blocking both runtime parsing and Next types. Added packages/zod/src/identity.ts, packages/zod/src/identity.test.ts, all three generated artifacts, and apps/frontend/lib/team.test.ts to explicit execution scope; preserved existing enum names/routes. Verified schema RED→GREEN and all canonical/type/browser gates. Commit: `62717ca`.
3. **[Rule 3 — Blocking] Stale sample/config naming:** corrected nonexistent internal/.env.sample and prose ENCRYPTION_KEYS_JSON to the real sample and existing configuration key. PLAN records every additional concrete path. No migration, worker, runtime or gate expansion. Commit: `62717ca`.

## Independent Review and Deferred Minor

An independent read-only gpt-6-astra review found no critical or important issues and confirmed scoped precedence, canonical artifacts, extraction guards and honest fixture evidence. One minor is deferred: the approved failed-delivery copy advises resending while its control is delivered in 02-25. The copy follows UI-SPEC; this plan does not add a fake action.

## Known Stubs

None prevent this plan's goal. Empty initial React collections are wired to actual scoped API reads. No worker/provider send path or skipped/pending runnable browser test was added.

## Next Phase Readiness

Ready for 02-21's actual invitation worker delivery. The API only queues encrypted intents; FLUX_INTEGRATION.RESEND_API_KEY is unchanged and API code performs no invitation provider calls. Acceptance, resend/revoke controls and live-provider factors/session/delivery evidence remain their later plans, including final external acceptance in 02-35. Broad requirements remain incomplete.

## Self-Check: PASSED

Both created application/test files and this SUMMARY exist. Task commits 6718ebe and 62717ca exist. All required acceptance/gate results above were observed; there are no tracked deletions or generated untracked outputs.
