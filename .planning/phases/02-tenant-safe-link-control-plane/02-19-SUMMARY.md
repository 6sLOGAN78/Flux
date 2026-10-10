---
phase: 02-tenant-safe-link-control-plane
plan: "19"
subsystem: invitations
tags: [go, postgres, nextjs, aes-gcm, idempotency, tdd]
requires:
  - phase: 02-18
    provides: Fresh SQL team authority, exclusive workspace locks, durable user provenance and protected audit/replay storage
provides:
  - Atomically queued tenant invitations with encrypted durable delivery intents and digest-only acceptance credentials
  - Freshly authorized bounded invitation listing and native inline Team creation with truthful queued feedback
  - Fail-closed API invitation configuration and meaningful HTTP, crypto, configuration and real browser proof
affects: [02-20, 02-21, 02-26, 02-35]
tech-stack:
  added: []
  patterns: [exclusive workspace then actor locks, AES-256-GCM random nonce with scoped AAD, protected replay authorization before hash, transactional pending expiry]
key-files:
  created:
    - apps/backend/internal/database/migrations/008_invitations.sql
    - apps/backend/internal/config/invitation.go
    - apps/backend/internal/config/invitation_test.go
    - apps/backend/internal/repository/invitation.go
    - apps/backend/internal/repository/delivery.go
    - apps/backend/internal/service/invitation.go
    - apps/backend/internal/service/invitation_crypto.go
    - apps/backend/internal/service/invitation_crypto_test.go
    - apps/frontend/tests/provider-assets.ts
    - apps/frontend/lib/provider-assets.test.ts
  modified:
    - apps/backend/.env.sample
    - apps/backend/internal/app/roles_test.go
    - apps/backend/internal/config/config.go
    - apps/backend/internal/config/config_test.go
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/repository/audit.go
    - apps/backend/internal/router/router.go
    - apps/backend/internal/service/permissions.go
    - apps/backend/internal/service/services.go
    - apps/backend/internal/testing/browser_fixture_test.go
    - apps/backend/internal/transport/health.gen.go
    - apps/backend/static/openapi.json
    - apps/frontend/app/workspaces/[workspaceId]/team/page.tsx
    - apps/frontend/lib/api.ts
    - apps/frontend/tests/control-plane.spec.ts
    - packages/zod/src/identity.ts
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/src/gen.test.ts
    - packages/openapi/openapi.json
    - .planning/phases/02-tenant-safe-link-control-plane/02-19-PLAN.md
    - .planning/phases/02-tenant-safe-link-control-plane/02-SOURCE-AUDIT.md
key-decisions:
  - Queue invitations and encrypted intents in one PostgreSQL transaction; the API never consumes the queue or sends email, and Queued remains truthful until the future worker acknowledges delivery.
  - Require external exact 32-byte AES keys only for the API role, retain safe key identifiers for rotation, and authenticate workspace/invitation/delivery identifiers plus key ID with standard random-nonce GCM.
  - Authorize the current locked actor and protected invitation role snapshot before consulting replay hashes; durable users retain provenance after membership removal.
requirements-completed: []
requirements-addressed: [TEN-03, TEN-05, TEN-07, TEN-08, SAFE-03]
duration: 47min
completed: 2026-10-10
---

# Phase 2 Plan 19: Encrypted invitation queue Summary

**Fresh SQL authority atomically commits scoped invitations, encrypted delivery intents, immutable audits and 24-hour replay records, with a real inline Team form reporting Invitation queued.**

## Performance

- Recorded implementation/verification interval: 2026-10-10T10:11:21Z through 2026-10-10T10:58:11Z (46m50s, rounded 47min). Earlier context loading and initial RED execution were not included in this recorded interval. No paused interval was counted.
- Tasks: 2/2.
- Task paths: 32 unique paths, comprising 30 application/configuration/test/generated paths and 2 planning corrections. SUMMARY and STATE/ROADMAP tracking are additional documentation paths.

## Accomplishments

- Owner grants admin/member/viewer; admin grants member/viewer; owner invitation grants are rejected. Exclusive workspace then actor locks serialize authority, removal and uniqueness decisions. Removed inviters cannot replay or inspect foreign hashes.
- Trimmed, case-folded email preserves dots and plus aliases, has a 254-byte maximum, and is unique among pending workspace invitations. Elapsed pending records expire transactionally before the partial uniqueness check; the index contains no volatile clock expression. New invitations expire in seven days.
- A crypto/rand 256-bit token is retained only inside the encrypted envelope; acceptance storage contains its SHA-256 digest. AES-256-GCM uses the standard [NewGCMWithRandomNonce API](https://pkg.go.dev/crypto/cipher#NewGCMWithRandomNonce), with workspace, invitation, delivery and key identifiers authenticated as AAD. The precise G407 annotation explains the standard API's required nil nonce without weakening other checks.
- Durable inviter/accepted-user references and scoped composite invitation/delivery/audit foreign keys survive membership deletion. Lease-generation fields prepare future worker fencing without implementing a consumer.
- GET returns only safe role/email/status/expiry identifiers, bounded to 25 entries with explicit UUID pagination. Strict POST parsing retains the existing 64 KiB browser mutation boundary. Tokens, digests, ciphertext, key IDs and private envelopes do not enter API responses or logs.
- The native inline Team form defaults to member, applies current server capabilities, reports exact queued/expiry/duplicate feedback, and retains the existing abort/generation and safe-first dirty-workspace-switch behavior.

## Task Commits

1. **Task 1: Specify invitation queue and actual Team form** — `e75e3c9` (test).
2. **Task 2: Deliver encrypted invitation queue and native Team form** — `8ddb0b6` (feat).

Both commits exist; no tracked files were deleted.

## Verification

- Registered actual HTTP RED used real migrated PostgreSQL: expected 201, actual 404 for the absent invitation endpoint. Native browser RED executed and failed because the named invitation form/default role were absent; missing infrastructure or compile failures were not accepted as RED.
- Registered `TestProductActualHTTP` GREEN proves atomic encrypted intent/digest storage, all actor/grant policies, duplicate uniqueness under concurrent requests, transactional expiry replacement, protected snapshot replay, removed-inviter durable provenance, foreign scope denial, private response/log exclusion, 25-entry pagination, lock-wait cancellation, ledger insertion rollback and deferred commit-time rollback.
- Crypto/configuration units prove authenticated decryption, fresh token/nonce generation, all AAD identities and ciphertext tampering, exact external key length, safe configuration errors and API-only requirements.
- Final root unit gate passed: 100 Go race unit tests and 26 frontend unit tests, including two public provider-asset cache controls; other workspace, script and tool suites passed. The full registered integration gate passed 17 tests across six groups. The earlier deferred TestTracestate HTTP/Redis/OTLP failure did not recur here; this plan makes no repair claim about it.
- Final root typecheck, lint, formatting, generated-artifact verification and migration verification passed. All three canonical generated artifacts were regenerated together, preserving existing transport enum names/constants.
- Final canonical `bun run test:e2e` passed **36** real browser cases through the original native fixture startup and strict report validator: positive execution, zero skipped, zero flaky/retried and zero infrastructure errors. Invitation happy path, actual duplicate 409 feedback, focused Stay choice, draft preservation, Discard switch and clean second-workspace form are exercised alongside all 35 existing regressions.
- After the browser gate, the production root build passed. Exported `cleanFrontendBuild(root)` performed its existing strict lstat/realpath/production-manifest ownership checks and removed only disposable frontend output. `CI=true bun run scan` then passed the complete worktree and full-history secret policies, Bun dependencies and imported Go packages across all roles/tests. Existing unused inventory advisory GO-2026-5932 remains visible; no vulnerable package is imported.
- Independent read-only review found one input-policy mismatch, corrected with real HTTP RED/GREEN below. No other blocking findings remained. Stub and added-trust-surface review found no production placeholder preventing this plan's goal.

## Deviations from Plan

### Auto-fixed issues

**1. [Rule 3 - Blocking] Preserve the actual migration sequence and reachable wiring**
- Existing 007_link_listing.sql already owns version 007. Added forward 008_invitations.sql and corrected plan artifact/action references, preserving every existing migration.
- Production reachability and operational correctness required the concrete config module/tests/sample, router/service registries, audit helper, shared role constants, binary/browser fixture configuration, canonical route inventory test, closed frontend duplicate error mapping and crypto units beyond the stale 14-path declaration. All paths appear in frontmatter; no new dependency or unrelated callback refactor was introduced. SOURCE-AUDIT records this authorized expansion.

**2. [Rule 1 - Bug] Align Go input validation with the canonical email policy**
- Independent review found that net/mail alone accepted `a@b`, while canonical installed Zod 3 email validation rejected it. A real HTTP regression failed expected 400/actual 201 before repair, then passed after exact ASCII syntactic validation was aligned. Trim/case-fold and dot/plus preservation remain unchanged.

**3. [Rule 3 - Blocking] Remove demonstrated repeated public CDN fixture fetches**
- Initial native full-browser diagnostic returned 34 expected/2 unexpected, zero skipped/flaky, with genuine pinned Clerk CDN asset timeouts. A later canonical run failed; its original runner discarded failed stdout. A private OS-temporary copy retained failed stdout inside native runCommand without replacing fixture startup/imports/report validation. That diagnostic again returned 34 expected/2 unexpected: one provider asset request ended during test completion, and one newly added dirty-switch assertion used invented button names.
- Corrected the new assertion to existing native **Stay/Discard** labels, retaining focus and draft/switch proof. This was a test selector error; production dirty-state wiring was already connected.
- Added a test-process cache for successful exact pinned Clerk JS 6.38.1/UI 1.39.1 public JavaScript bytes. It strips request headers to public Accept, disallows redirects, validates status/content type, bounds individual/total bytes and entries, deduplicates in-flight fetches, clears failed misses, isolates buffers and disposes network responses. Provider API/session responses and credential headers are never cached. Existing 15-second timeout and per-page cancellation/drain remain unchanged. Repeated-byte identity and fetch/non-200/private/oversize failure controls passed before the unchanged canonical 36-case GREEN run.
- Diagnostic reports remained private outside the repository (directory 0700, files 0600); no raw token/key/provider payload was printed or committed. No production workaround, timeout increase, skip/retry rule or scanner exception was added.

## Decisions and Scope

New API configuration requires ACTIVE_KEY_ID, ENCRYPTION_KEYS (JSON safe key ID to canonical base64 exact 32-byte key), SENDER and HTTPS PUBLIC_ORIGIN with safe-label errors and no plaintext fallback. The actual .env.sample documents external configuration/rotation without committing production secrets; worker ownership evolves in plan 21.

Intentional sequencing: plan 19 queues durable encrypted intent only. Acceptance/resend/revoke belong to subsequent plans, worker/provider delivery to plan 21, and configured live provider factor/session/delivery acceptance to plan 35. No real external email or production provider send occurred. No API queue consumer, domains, redirects, analytics, billing or destination fetch was added. Broad requirements remain pending; `requirements-completed` is deliberately empty.

## Self-Check: PASSED

All ten created task files and canonical artifacts exist. Commits e75e3c9 and 8ddb0b6 resolve. No tracked deletions, blocking production stubs or unresolved review findings remain. Final canonical browser and complete security gates passed with the counts and limits above.
