---
phase: 02-tenant-safe-link-control-plane
plan: "14"
subsystem: links
tags: [go, postgres, nextjs, hmac, pagination, tenancy, tdd]
requires:
  - phase: 02-13
    provides: Freshly authorized deterministic link collection and real library
provides:
  - Versioned HMAC-signed workspace/query-bound timestamp-UUID cursor pagination
  - Successful local page history, exact recovery copy and authoritative exhaustion
  - API-only private cursor key validation and operator provisioning instructions
affects: [02-15, 02-27, 02-32, 02-35]
tech-stack:
  added: []
  patterns: [HMAC-SHA256 continuation, limit-plus-one seek, scoped local cursor history]
key-files:
  created:
    - apps/backend/internal/service/cursor.go
    - apps/backend/internal/service/cursor_test.go
  modified:
    - apps/backend/.env.sample
    - apps/backend/internal/app/roles_test.go
    - apps/backend/internal/config/config.go
    - apps/backend/internal/config/config_test.go
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/repository/link.go
    - apps/backend/internal/service/link.go
    - apps/backend/internal/testing/browser_fixture_test.go
    - apps/backend/internal/transport/health.gen.go
    - apps/backend/static/openapi.json
    - apps/frontend/components/link-list.tsx
    - apps/frontend/lib/api.test.ts
    - apps/frontend/lib/api.ts
    - apps/frontend/lib/links.test.ts
    - apps/frontend/lib/links.ts
    - apps/frontend/tests/control-plane.spec.ts
    - docs/development.md
    - packages/openapi/openapi.json
    - packages/openapi/src/contracts/identity.ts
    - packages/zod/src/identity.ts
key-decisions:
  - Bind versioned cursor positions to workspace and exact effective search/lifecycle fingerprint; fresh caller SQL scope remains authority.
  - Require a private standard-Base64 32-byte signing key only for API startup, with no fallback and safe diagnostics.
  - Remember only successful scoped pages locally; expose exhaustion only from authoritative null continuation.
  - Connect the approved library creation CTA to the existing real create route and verify the committed browser flow.
requirements-completed: []
requirements-addressed: [LINK-03, LINK-06, TEN-06, TEN-07, TEN-08, SAFE-03]
duration: 34min
completed: 2026-10-09
---

# Phase 2 Plan 14: Signed Tenant-Bound Pagination Summary

**HMAC-signed workspace/query-bound timestamp-UUID seeks deliver real Next/Previous pages, exact cursor recovery and authoritative exhaustion under fresh PostgreSQL membership checks.**

## Performance

- Tasks: 2/2.
- Recorded execution interval: 2026-10-09T17:04:57Z–2026-10-09T17:39:15Z, approximately 34 minutes; initial context loading precedes this interval.
- Actual implementation/test/documentation inventory: **23 paths**, exhaustively listed above; two newly created and 21 modified. Summary and state/roadmap tracking are additional documentation paths.

## Accomplishments

- Standard-library HMAC-SHA256 signs version-1 opaque bounded cursors containing typed timestamp and UUID, workspace and SHA-256 fingerprint of the exact effective search/lifecycle query. Strict canonical URL-safe Base64, signature length, constant-time `hmac.Equal`, version, nonzero identifiers, timestamp and closed JSON payload checks reject malformed/tampered tokens. Raw cursor input is limited to 2048 bytes; whole-query overflow also returns the approved cursor recovery error without parsing an unbounded query.
- `FLUX_LINKS.CURSOR_KEY` must be canonical standard Base64 decoding to exactly 32 bytes. API startup fails with safe labels for absent/invalid keys; redirector, worker and migrator remain independent of this setting. Service construction also fails closed when used directly without a validated key. No production secret or fallback was added. Documentation specifies private provisioning, API replica agreement and rotation invalidating old cursors.
- Every actual list query still opens the existing explicit transaction and freshly requires read capability using caller `Scope`, workspace locking and parameterized workspace/nondeleted predicates. A verified cursor contributes only `(created_at,id)` seek values. Descending tuple seeks fetch `limit+1`; only an actual extra row produces a continuation. No guessed total or snapshot isolation across requests is claimed. Migration 006 custom keys and 007 listing remain intact.
- Native Next/Previous controls consume only successfully visited local positions within the exact workspace/query scope. Scope changes reset history, and the existing keyed workspace subtree, abort/generation ownership, transient row retention, access-loss disposal and viewer permissions remain intact. Current effective query is nondeleted with empty search; search/filter controls remain plan 15. Cursor binding already rejects signed tokens for another effective filter. The UI announces **“No more links.”** only after a successful null continuation and shows **“This page is no longer available. Return to the first page.”** with **“Return to first page.”** recovery.
- The library now exposes the approved creation anchors to allowed users, including **“Create your first link”** in the first-page empty state. Both anchors use the already implemented real create route. A dedicated actual browser test creates and observes a committed link through this CTA; viewers have no creation anchor. A later empty continuation does not claim that the entire workspace has no links.
- Canonical authored schemas/contracts describe nonempty bounded continuations, supported queries, signed pagination and rejection behavior. Both OpenAPI copies and the generated Go boundary were regenerated together. The bounded browser error reader projects only a matching-status canonical `CURSOR_INVALID` code and renders locally authored copy. Existing collection GET 8 MiB, other success 64 KiB and error 8 KiB ceilings are unchanged, with all prior boundary tests retained.

## Task Commits

1. Task 1 — Specify actual pagination, key validation and browser navigation: `cb11215` (`test`).
2. Task 2 — Deliver signed tenant-bound pagination and recovery: `4dfda1c` (`feat`).

## Verification

- Meaningful actual HTTP RED: registered `TestProductActualHTTP/link-library` failed because 104 nondeleted equal-timestamp rows produced a null continuation. Configuration RED accepted an invalid signing key. Actual browser RED rendered the real 26-link collection but lacked “Previous page”: **0 expected / 1 unexpected / 0 skipped / 0 flaky / 0 infrastructure errors**. Initial rapid seeding received 429 and was rejected as RED evidence; pacing real POSTs by 70 ms respects the unchanged production limiter.
- Actual migrated-PG production-router GREEN proves all 104 original equal-timestamp rows appear exactly once through exhaustion, including a new head inserted between requests, plus default/max limits, deleted exclusion, viewer reads, foreign workspace reuse, correctly signed filter-fingerprint mismatch, malformed and tampered signatures, 2049-byte and 9000-byte oversized input, exact `400 CURSOR_INVALID` recovery and fresh membership denial even with a valid cursor. Codec units additionally exercise typed/unknown/version/trailing payload rejection, timestamp precision, key changes and search/lifecycle binding.
- During boundary review, actual HTTP RED demonstrated that a 9000-byte cursor incorrectly received `BAD_REQUEST`; the bounded early rejection was corrected and the same actual HTTP case passed.
- Full `bun run test:unit`: **95 discovered race-enabled Go unit tests**, all workspace/script suites and tool checks passed. Direct complete frontend unit execution: **21 passed / 0 failed** across three files, including all 12 preceding API boundary tests and new scoped history/canonical cursor projection proof.
- Full `bun run test:integration`: **17 registered top-level tests across six groups passed**. The final full integration rerun after the oversized-query correction also passed all 17. The previously deferred telemetry transient did not recur; no unrelated repair is claimed.
- Focused root browser gate: **2 completed**, zero skipped/flaky/infrastructure failures, covering actual pagination/tamper recovery and committed empty-state creation. Final canonical full `bun run test:e2e -- --project=local`: **29 completed**, with unchanged fail-closed report validation rejecting skipped/flaky/infrastructure results.
- Root format/lint/typecheck and generation drift checks passed. Final format/lint passed after boundary changes; the final browser-source formatter comparison and frontend typecheck also passed after the assertion adjustment. Migration verification passed the preserved latest version **7** and upgrade-prefix checks.
- Required final root production build after the last browser run passed all role/workspace builds. The unchanged exported `cleanFrontendBuild` validated lstat/realpath and the production prerender manifest before removing disposable frontend output.
- `CI=true bun run scan` passed Bun dependency, all-role/test imported Go vulnerability, complete worktree secret and full-history secret scans. Existing unused inventory advisory `GO-2026-5932` remains visible; no vulnerable package is imported. No runtime pins, ownership registries, scanner exceptions, gates or cleanup constraints changed.

## Deviations from Plan

1. **[Rule 3 - Blocking] Correct stale sample path.** Used existing `apps/backend/.env.sample` instead of nonexistent `apps/backend/internal/.env.sample`; no duplicate was created.
2. **[Rule 3 - Blocking] Supply the new owned setting in verification fixtures.** Narrow changes to `config/config_test.go`, `app/roles_test.go` and `testing/browser_fixture_test.go` preserve real startup/browser tests with explicit test-only keys. `docs/development.md` documents the new private operator setting. These necessary paths were authorized by the orchestrator.
3. **[Rule 3 - Blocking] Keep canonical artifacts current.** Updated `packages/zod/src/identity.ts`, `packages/openapi/src/contracts/identity.ts` and all three generated outputs. The stale contract otherwise claimed cursors were unsupported. No competing handwritten transport was introduced.
4. **[Rule 3 - Blocking] Recognize canonical browser recovery.** Added `CURSOR_INVALID` to the closed matching-status projection in `lib/api.ts` and a narrow `api.test.ts` regression. The new 400 could otherwise not reach the specified recovery UI. Response budgets and preceding tests are preserved.
5. **[Rule 2 - Missing Critical Functionality] Connect existing creation flow.** The owned library lacked the reachable approved first-link CTA after creation had already shipped. Connected its existing route, applied approved empty-state copy, retained viewer denial and added an explicit committed browser happy path. This is a narrow integration refinement; no whole onboarding/provider requirement is marked complete.

All five adjustments are within the orchestrator's explicit authorization; the GREEN commit contains them. No feature from plan 15 was started.

## Issues Encountered

- The new API error unit fixture initially omitted required canonical envelope fields; corrected the fixture, preserving strict production parsing, and all 21 frontend units passed.
- New Go shadowing and long lines were corrected in one lint pass; the unchanged full lint gate passed afterward.
- Private full browser diagnostics showed **28 expected / 1 unexpected / 0 skipped / 0 flaky / 0 infrastructure errors**: the onboarding test expected the superseded empty-state sentence. Updated only that copy assertion and added an exact creation-route check while retaining committed retry, authorization, foreign-workspace denial and shell disposal assertions. Review caught an undefined variable introduced in the new route assertion and corrected it to the exact current authorized pathname; the intervening standard run failed without an exposed diagnostic, so no specific failure cause is asserted for that run. Final source typechecking and the canonical 29-test run passed. Raw reports remained outside the repository in an OS-protected temporary directory.

## Operator Setup and Remaining Acceptance

Provision private `FLUX_LINKS.CURSOR_KEY` to API deployments as documented; rotation deliberately invalidates existing cursors and exposes first-page recovery. Actual live provider verification/recovery/OAuth acceptance remains pending plan 02-35. Requirement completion remains empty until the full requirement and provider acceptance gate; search/filtering belongs to plan 15.

## Self-Check: PASSED

Both newly created cursor files exist. Commits `cb11215` and `4dfda1c` exist and contain the declared RED/GREEN work. The exact 23-path implementation inventory is committed, no accidental deletions or untracked outputs remain, and all required gates have passing evidence. Stub/threat-surface review found no pagination stub or undeclared network/auth/schema surface; the existing later-phase redirect/analytics availability statement remains intentional.
