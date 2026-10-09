---
phase: 02-tenant-safe-link-control-plane
plan: "08"
subsystem: workspace
tags: [postgresql, tenancy, preferences, nextjs, contracts]
requires:
  - phase: "02-07"
    provides: Fresh workspace authorization and native Links shell
provides:
  - Membership-derived bootstrap and safe last-workspace restoration
  - Workspace-locked preference selection for all current member roles
  - Native authorized chooser with session cancellation and safe sign-out
affects: [02-09, 02-10, 02-35]
tech-stack:
  added: []
  patterns: [membership-joined preferences, workspace-first selection locks, session-bound chooser]
key-files:
  created:
    - apps/backend/internal/database/migrations/004_workspace_preferences.sql
    - apps/frontend/app/workspaces/page.tsx
  modified:
    - apps/backend/internal/repository/workspace.go
    - apps/backend/internal/service/workspace.go
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/frontend/app/page.tsx
    - apps/frontend/lib/workspace.ts
    - apps/frontend/tests/control-plane.spec.ts
    - packages/zod/src/identity.ts
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/openapi.json
    - apps/backend/static/openapi.json
    - apps/backend/internal/transport/health.gen.go
key-decisions:
  - Derive restored selection from the same current membership snapshot as the chooser; a stored preference never grants authority or exposes removed workspace identifiers.
  - Select under the existing workspace-first shared lock and recheck membership before saving the preference; creation commits its initial preference atomically.
  - Keep provider-only unit bootstrap substitution strictly test-local; all product and browser fixtures use real migrated PostgreSQL and production middleware.
requirements-completed: []
requirements-addressed: [TEN-06, TEN-07, TEN-08, SAFE-02, SAFE-03]
duration: 15min
completed: 2026-10-09
---

# Phase 2 Plan 8: Authorized Workspace Restore Summary

**Workspace restoration uses fresh PostgreSQL membership, while locked preference selection and a native chooser prevent stale or forged workspace values from exposing tenant data.**

## Performance

- Tasks: 2/2. Resumed execution and verification: approximately 15 minutes; the interrupted predecessor's elapsed interval is excluded.
- Implementation: 22 unique paths, comprising 14 declared paths and eight necessary supporting paths. The predecessor's RED commit and saved implementation were preserved.

## Accomplishments

- Migration 004 creates an identity-owned nullable workspace preference. GET `/api/v1/me` returns current memberships and restores only a preference joined to that same membership snapshot. Removed preferences return `null` and expose neither the former UUID nor name. Empty memberships serialize as an array.
- PUT `/api/v1/me/last-workspace` validates the strict canonical JSON body, acquires the existing shared workspace lock, rechecks membership and upserts the actor's preference transactionally. Owner/admin/member/viewer may select; foreign, unknown and removed membership receive safe denial. Workspace creation saves its initial preference in the existing atomic transaction. Preferences never authorize subsequent requests.
- Root routes fresh authorized restoration to Links or the chooser. The chooser lists only server-authorized workspace summaries, commits Open workspace before routing, supports Create workspace and an empty state, clears private state on session changes/focus revalidation, cancels pending requests, and renders a safe local signed-out screen. Browser storage and cross-tab messages supply no authority.
- Canonical Zod/ts-rest modules remain the authored contract source. Both OpenAPI documents and generated Go transport were regenerated together. Existing fixed bearer/Origin/JSON/no-store protections and production authentication remain intact.
- Real PostgreSQL tests verify foreign selection, malformed/unknown/trailing JSON, every role, a selection blocked behind removal that rechecks after commit, and stale bootstrap privacy. Migration proof verifies 003→004, 004 down to 003, latest upgrade and existing full rollback/binary cases.

## Task Commits

1. **Specify restore only a current member workspace:** `79a16e6` — predecessor captured meaningful actual-PG missing bootstrap fields and actual-browser missing chooser behavior; no infrastructure failure was counted as RED.
2. **Deliver restore only a current member workspace:** `94de69b` — preference migration, membership-derived bootstrap, locked mutation, root/chooser, regenerated contracts and preserved security/session proof.

## Verification

- `bun run test:unit`: passed **87 discovered race-enabled Go top-level unit tests**, all **four workspace suites**, script regressions and tool self-tests; all four dependency-ordered production builds passed.
- `bun run test:integration`: passed **17 registered top-level tests across six groups** with real pinned PostgreSQL/Redis and the race detector. Restore cases remain subtests of registered `TestProductActualHTTP`; no new unregistered container test was added.
- `bun run test:e2e -- --project=local --grep restore`: **1 completed, zero skipped/flaky**. Authorized chooser selection and root restoration succeed; forged storage/message values are ignored; membership removal returns to the chooser without former UUID/name.
- Targeted existing sign-out test: **1 completed, zero skipped/flaky** after fixing local signed-out rendering. Full `bun run test:e2e -- --project=local`: **15 completed, zero skipped/flaky**. Prior identity stability, sign-out, expiry/recovery, provider controls, keyboard and responsive assertions remain intact.
- `bun run typecheck`, `bun run lint`, `bun run format:check`, `bun run generate:check`, and `go build ./cmd/...`: passed on final source. Canonical generated outputs match exactly.
- Following browser dev execution, production Next output was rebuilt successfully and removed only through the existing validated `cleanFrontendBuild` helper. `CI=true bun run scan`: imported-package dependency scan, Bun scan, complete worktree and full Git-history secret scans all passed. The unused inventory advisory GO-2026-5932 remains visible; no vulnerable package is imported by a role or test.
- `git diff --check`: passed. No tracked file deletions or repository-local private browser reports were introduced. Runtime/tool/provider pins and scanner enforcement are unchanged.

## Deviations from Plan

**1. [Rule 3 - Blocking] Eight necessary supporting paths**
- `apps/backend/internal/router/router.go`: register the declared selection adapter in the existing versioned router.
- `apps/backend/internal/database/migrator_test.go`: adapt existing migration proof to 004 with explicit forward/down checks.
- `apps/backend/internal/testing/product.go`: test-only reset/removal helpers operate on actual PostgreSQL below application layers; no production test endpoint or auth bypass.
- `apps/backend/internal/service/auth_test.go`: inject the actual workspace service into the legacy real-PG fixture. Provider-only container-free units retain the production router/middleware with an explicit test-local canonical empty-bootstrap probe; signed, revoked, cookie-only and concurrent session checks remain unchanged.
- `apps/backend/internal/transport/contract_test.go`, `apps/frontend/lib/api.test.ts`, and `packages/openapi/src/gen.test.ts`: update strict canonical proof for required bootstrap fields and the new implemented route.
- `apps/frontend/lib/api.ts`: admit PUT through the existing fixed-origin bearer request helper without changing its security behavior.
- Commit: `94de69b`. These are consumed wiring and affected regression adaptations, without broader refactoring or additional product behavior.

**2. [Rule 1 - Bug] Safe chooser sign-out and race-test completion**
- Full browser evidence initially showed 14 passing cases and one sign-out failure: private identity cleared but the signed-out heading never rendered. Chooser now renders local signed-out state and avoids the competing post-sign-out root navigation. All original assertions remain; targeted and full browser checks passed.
- New removal-race HTTP assertions were moved from the worker goroutine onto the test goroutine so fatal assertions cannot strand the result channel. Lint and full integration passed. Initial optional-snapshot narrowing and line-length findings were also corrected without suppressing checks.
- Commit: `94de69b`.

## Known Stubs and Remaining Acceptance

- No new production stubs were found in this plan's modified paths. Existing disabled Team availability remains assigned to **02-16**; first-link creation remains assigned to **02-14**. Their existing truthful states were preserved.
- Local browser evidence uses the real provider SDK/UI with test-only signed transport plus production Go router and real PostgreSQL. Live provider factors/session/cookie acceptance remains pending **02-35**. No external authentication gate occurred.
- This bounded D-03 slice does not complete phase-wide TEN/SAFE requirements; requirement traceability remains unchanged.

## Self-Check: PASSED

Both created files and all implementation paths exist. RED `79a16e6` and GREEN `94de69b` exist in Git. Exact staging, no tracked deletions, final positive executed counts, regenerated bytes and private-output boundaries were checked before SDK state advancement.
