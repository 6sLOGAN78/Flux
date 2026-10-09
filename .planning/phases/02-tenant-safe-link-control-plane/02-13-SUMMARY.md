---
phase: 02-tenant-safe-link-control-plane
plan: "13"
subsystem: links
tags: [go, postgres, nextjs, tenancy, tdd]
requires:
  - phase: 02-12
    provides: Canonical link creation/detail and custom-key retry behavior
provides:
  - Freshly authorized bounded newest-first nondeleted link collection
  - Responsive real library rows with safe retry and membership disposal
  - Canonical generated list response and bounded browser collection transport
affects: [02-14, 02-15, 02-27, 02-32, 02-35]
tech-stack:
  added: []
  patterns: [timestamp-UUID ordering, partial tenant index, endpoint-specific response budget]
key-files:
  created:
    - apps/backend/internal/database/migrations/007_link_listing.sql
    - apps/frontend/components/link-list.tsx
  modified:
    - apps/backend/internal/repository/link.go
    - apps/backend/internal/service/link.go
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/router/router.go
    - apps/frontend/app/workspaces/[workspaceId]/links/page.tsx
    - apps/frontend/lib/api.ts
    - apps/frontend/lib/api.test.ts
    - apps/frontend/tests/control-plane.spec.ts
    - packages/zod/src/identity.ts
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/openapi.json
    - apps/backend/static/openapi.json
    - apps/backend/internal/transport/health.gen.go
key-decisions:
  - Reject supplied cursors and filters until their complete scoped refinements; nextCursor remains null without exhaustion or pagination claims.
  - Preserve migration 006 custom keys and introduce listing as forward migration 007.
  - Allow only UUID-scoped collection GET responses up to 8 MiB; retain 64 KiB for other successful responses and 8 KiB for errors.
requirements-completed: []
requirements-addressed: [LINK-03, LINK-06, TEN-06, TEN-07, TEN-08, SAFE-03]
duration: 20min
completed: 2026-10-09
---

# Phase 2 Plan 13: Authorized Link Library Summary

**Fresh SQL membership authorizes a bounded, deterministic tenant link library, rendered as desktop rows and mobile cards with real loading, retry, viewer and access-loss behavior.**

## Performance

- Tasks: 2/2. Resumed interrupted execution without repeating or replacing the committed RED task.
- Duration: approximately 20 minutes of resumed execution; the earlier interruption and usage-limit wait are excluded.
- Actual implementation/test paths: 16 across RED and GREEN. Fifteen changed in GREEN; the previously committed backend test remains intact. The declared `lib/links.ts` required no change. Summary/tracking are additional documentation paths.

## Accomplishments

- Registered production `GET /api/v1/workspaces/:workspaceId/links` with default 25 and allowed 1–100 limit. Query parsing rejects duplicates, invalid bounds, unsupported cursors, search/state/foreign filters and oversized input. No total, fake cursor or pagination-completion claim is emitted.
- Repository begins the existing transaction, freshly requires read capability under the shared workspace lock, and applies parameterized workspace/nondeleted predicates before ordering by `(created_at DESC, id DESC)`. A matching partial workspace/timestamp/UUID index supports the query. Current viewers may read; removed membership returns safe 404 without rows.
- Canonical Zod and ts-rest authority defines a strict bounded `items`/nullable `nextCursor` response. Both OpenAPI copies and generated Go transport were regenerated together; no handwritten DTO was introduced.
- The consumed workspace Links page now displays real semantic desktop tables and labeled mobile cards. Full accessible title/detail links, short URL, destination links, local full timestamps and textual state are available. Initial loading and failure never claim empty; successful empty results do. Same-tenant temporary failure retains rows and offers retry. Read-only viewers receive the approved helper and no mutation CTA. Existing abort/generation and workspace subtree ownership discard requests/data on membership denial or switching.
- Existing creation, committed detail navigation, custom keys, canonical errors and idempotency behavior remain intact. Signed pagination and creation CTA refinement remain assigned to plan 14; search/filtering to plan 15; lifecycle mutations to later plans.
- Collection GET transport has an endpoint-specific 8 MiB success ceiling. Go validates destination length at 8192 UTF-8 bytes; worst-case six-byte JSON escapes across 100 destinations need about 4.92 MiB. Current creation title is at most 200 runes, managed hostname at most 253 bytes, short key at most 64 ASCII characters, creator email at most 320 characters; metadata remains well within the remaining allowance. Current production creates no suspension reason; later suspension operations must maintain bounded inputs. Noncollection success remains 64 KiB, error body remains 8 KiB, and cancellation, deadline, token and privacy safeguards remain intact.

## Task Commits

1. Task 1 — Specify authorized library and recovery: `4b46930` (`test`, preceding agent).
2. Task 2 — Deliver authorized newest-first library: `556ad75` (`feat`).

## Verification

- Existing meaningful actual HTTP RED from `4b46930`: registered `TestProductActualHTTP/link-library` received 405 instead of expected 200. Earlier private browser RED diagnostic reported absent `Loading links…`, one failure with zero skipped/flaky/infrastructure results. These preceding-agent results were inherited, not repeated or fabricated as new execution.
- Full `bun run test:integration`: **17 registered top-level tests across six groups passed**, including actual migrated-PG production-router collection proof: empty array/null cursor, default/max bounds, 104 nondeleted rows, deterministic equal-timestamp UUID order, deleted exclusion, invalid/duplicate queries, unsupported foreign filters, foreign detail-ID denial, viewer listing and membership-removal denial. The unchanged telemetry transient did not recur; no repair is claimed.
- Full `bun run test:unit`: **92 discovered race-enabled Go unit tests**, all workspace suites, script suites and tool checks passed. Following the transport adjustment, the complete frontend unit suite passed **19 tests**, including 12 API tests and new 100-row heavily escaped response, collection overflow and unchanged noncollection ceiling proof.
- Focused root browser gate: **2 completed**, zero skipped/flaky/infrastructure errors. Full root browser regression: **27 completed**, zero skipped/flaky/infrastructure errors (25 existing plus two library tests).
- Final focused browser verification after explicit Archived/Active label coverage: **2 completed**, zero skipped/flaky/infrastructure errors. Archived coverage uses an explicitly test-only canonical state projection for presentation; actual rows, order, destinations, detail anchors, timestamp, loading/error/retry and membership denial use real backend responses. No archive mutation capability is claimed. Viewer role projection is likewise presentation proof; actual viewer SQL authorization is covered independently by real-PG HTTP tests.
- Root format, lint, typecheck and generation check passed. Migration check passed latest version **7** and preserved upgrade-prefix proof.
- All role/workspace production builds passed, including the required final root build after the last browser run. The unchanged exported `cleanFrontendBuild` validated lstat/realpath and the production manifest before removing disposable output.
- `CI=true bun run scan` passed Bun dependency, all-role/test imported Go vulnerability, worktree secret and full-history secret checks. Unused inventory advisory `GO-2026-5932` remains visible; no vulnerable package is imported. No runtime pins, scanner exceptions or gates changed.

## Deviations from Plan

1. **[Rule 3 - Blocking] Preserve migration numbering.** Plan 12 already owns `006_custom_link_keys.sql`; the stale plan13 `006_link_listing.sql` became `007_link_listing.sql`. Existing migration is preserved and migration checks pass.
2. **[Rule 3 - Blocking] Register the reachable route.** The necessary single GET registration in `router/router.go` is outside the stale declared file list; actual production-router HTTP proof requires it. Authorized by the orchestrator.
3. **[Rule 2 - Missing Critical Functionality] Bound valid collection transport correctly.** Existing 64 KiB success limit rejected potentially valid pages. Narrow changes to `lib/api.ts` and `lib/api.test.ts` add the justified collection-only limit and meaningful boundary proof. Authorized by the orchestrator; two additional actual paths are recorded.

## Issues Encountered

- Initial lint flagged two magic numbers; named query/default-limit constants resolved both, and full lint passed.
- First focused browser failure was ambiguous global alert selection, matching both the library error and Next's route announcer. Private diagnostic recorded two failures, zero skipped/flaky, no infrastructure error. Scoped Links-region assertions preserve exact error/recovery checks.
- Final Archived presentation refinement initially projected into an actual 404 response without `items`, causing one test failure and one pass. A private diagnostic identified that test callback error; restricting projection to a successful response with a row preserved the genuine denial. The final strict focused gate passed. Neither browser correction weakened the runner or production authorization.
- Existing onboarding assertions now verify the successful real empty-library message instead of the replaced disabled creation placeholder; creation/navigation/isolation assertions remain intact. Plan 14 owns the library creation CTA.

## Known Stubs

None in delivered paths. Null cursor and unsupported cursor/filter denial are intentional bounded scope, with complete refinements assigned to plans 14–15. Library creation CTA is assigned to plan 14. Existing unreachable scaffolds were not modified or represented as completed features.

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: authenticated response budget | apps/frontend/lib/api.ts | UUID-scoped collection GET alone may read up to 8 MiB; deterministic overflow/noncollection checks retain bounded resource use. |

## Next Phase Readiness

Plan 14 can refine this real library with signed tenant-bound pagination. Actual provider factor/session acceptance remains outstanding for plan 35; broad requirements remain pending and `requirements-completed` is intentionally empty. No transition or plan14 execution occurred.

## Self-Check: PASSED

All sixteen actual implementation/test paths exist; RED `4b46930` and GREEN `556ad75` exist in Git. No accidental tracked-file deletion occurred. All three generated outputs agree. Required automated checks and positive actual browser counts are recorded above.
