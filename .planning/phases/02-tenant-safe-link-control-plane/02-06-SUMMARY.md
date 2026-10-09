---
phase: 02-tenant-safe-link-control-plane
plan: "06"
subsystem: workspace
tags: [postgresql, tenancy, idempotency, echo, nextjs]
requires:
  - phase: "02-05"
    provides: Durable internal identity and reusable signed real-PostgreSQL product fixtures
provides:
  - Atomic named workspace, owner membership, and identity-scoped bootstrap ledger
  - Freshly authorized workspace list and summary with tenant transaction lock discipline
  - Reachable named onboarding and authorized Links empty state
affects: [02-07, 02-08, 02-14, 02-35]
tech-stack:
  added: []
  patterns: [workspace-first authorization locks, canonical SHA-256 bootstrap replay, durable user provenance]
key-files:
  created:
    - apps/backend/internal/database/migrations/003_workspaces.sql
    - apps/backend/internal/repository/workspace.go
    - apps/backend/internal/service/workspace.go
    - apps/frontend/app/onboarding/page.tsx
    - apps/frontend/app/workspaces/[workspaceId]/links/page.tsx
    - apps/frontend/components/empty-state.tsx
  modified:
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/router/router.go
    - apps/backend/internal/service/services.go
    - apps/backend/internal/testing/browser_fixture_test.go
    - apps/frontend/app/page.tsx
    - apps/frontend/package.json
    - apps/frontend/tests/control-plane.spec.ts
    - packages/zod/src/identity.ts
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/src/gen.test.ts
    - packages/openapi/openapi.json
    - apps/backend/static/openapi.json
    - apps/backend/internal/transport/health.gen.go
key-decisions:
  - Bootstrap serializes on durable actor identity before a workspace exists; replay releases that lock and takes the normal workspace-first authorization lock.
  - Browser execution prepares the canonical schema dependency explicitly instead of depending on a preceding unit build.
  - Keep the visible first-link control disabled with an accessible explanation until plan 02-14 wires actual link creation.
requirements-completed: []
requirements-addressed: [TEN-02, TEN-05, TEN-07, TEN-08, SAFE-03]
duration: 26min
completed: 2026-10-09
---

# Phase 2 Plan 6: Named Workspace Bootstrap Summary

**PostgreSQL atomically commits a Unicode-named workspace, its owner, and a SHA-256 retry ledger; fresh Flux membership authorizes each workspace response.**

## Performance

- Duration: approximately 26 minutes, including verification and private browser diagnostics.
- Tasks: 2/2.
- Implementation files: 20 (14 declared paths and six narrowly necessary supporting paths).

## Accomplishments

- Migration 003 adds workspaces, the closed owner/admin/member/viewer membership matrix, tenant-owned mutation and immutable audit tables, and bootstrap requests whose committed result resolves to a workspace. Actor and creator provenance references durable users; membership audit targets are protected historical snapshots. Memberships and tenant tables use composite workspace/id primary keys. Ledger retention has a database-enforced minimum of 24 hours.
- Parameterized queries commit workspace, owner, and canonical validated-payload digest together. Unchanged retries replay one committed workspace after fresh authorization; changed payloads conflict. Reads require durable actor identity, and scoped authorization locks the workspace before reading membership. Ordinary writes can reuse the shared lock and role gate in their transaction; membership mutations must take the exclusive lock.
- Authenticated GET/POST workspace collection and GET scoped summary extend the existing exported identity contract owners. All three generated outputs match. Existing bearer, Origin, JSON, body-limit, safe-error and no-store middleware remain installed.
- The account provides a reachable onboarding entry. Required-name onboarding preserves typed input and its session-scoped retry key after uncertain delivery, navigates only after the committed API response, and renders Links only after a new server membership check. Foreign workspace content is denied; focus changes trigger revalidation.

## Task Commits

1. **Specify atomic named workspace behavior:** `8a79e0c` — actual PostgreSQL signed HTTP and browser specifications. HTTP RED reached the absent POST route (404 versus expected 201); the absent onboarding browser specification failed.
2. **Deliver atomic named workspace behavior:** `cb71b8f` — migration, repository/service, canonical routes/transports, onboarding, authorized empty state, and executable security/recovery cases.

## Verification

- `bun run test:integration`: passed all **17 registered top-level tests across six groups**, using real pinned PostgreSQL/Redis and race-enabled execution. The product dispatcher's workspace cases cover trimming, invalid names/unknown JSON fields, canonical retries, hash conflict, concurrent same-key single creation, one owner, retention/hash size, commit-time rollback, identity-scoped keys, foreign summary/list isolation, viewer write denial, and summary/replay denial after membership removal. Existing signed identity/session cases remain intact.
- `bun run test:e2e -- --project=local --grep workspace`: **1 completed, zero skipped/flaky**. The browser sends real signed requests to the Go/PostgreSQL fixture, deliberately loses a committed creation response, retries with the same key and typed name, opens authorized Links, and denies a foreign workspace.
- `bun run test:e2e -- --project=local`: **14 completed, zero skipped/flaky**, preserving the previous 13 browser cases.
- `bun run test:unit`: passed all **83 discovered race-enabled Go top-level unit tests**, all four workspace package suites, and shared tooling/self-tests. Its dependency-ordered prerequisites also passed all four workspace production builds, including Next.js. Canonical tests verify exact new paths, bearer metadata, required retry header, and 100 versus 101 Unicode-character response validation.
- `bun run lint`, `bun run format:check`, `bun run generate:check`, and `go build ./cmd/...`: passed. Backend vet/staticcheck/golangci and workspace TypeScript production compilation remained enforced.
- Existing validated production-build cleanup ran after browser testing; `CI=true bun run scan`: dependency, complete worktree and full Git-history scans passed. The visible unused OpenPGP module inventory advisory GO-2026-5932 remains; no vulnerable package is imported by any backend role or test. No runtime, dependency, scanner pin or scan exception changed.

## Decisions Made

Bootstrap actor serialization applies only before a workspace exists. Replay releases the actor lock before taking the scoped workspace lock, preserving workspace-first ordering and current membership authorization. Browser package execution builds its canonical dependency itself, so stale ignored outputs cannot substitute for current authored contracts.

## Deviations from Plan

### Auto-fixed Blocking Issues

**1. [Rule 3 - Blocking] Required composition and reachability wiring outside the 14 declared paths**
- Workspace services needed explicit production/test injection and route registration. Updated only `internal/service/services.go`, `internal/router/router.go`, and the existing external `internal/testing/browser_fixture_test.go`; the signed fixture protocol and auth corpus remain unchanged.
- Added the account's onboarding entry in `apps/frontend/app/page.tsx` so the delivered flow is reachable.
- Commit: `cb71b8f`.

**2. [Rule 3 - Blocking] Standalone browser command consumed stale canonical schema output**
- Generation intentionally builds in isolation and does not refresh workspace runtime dist. The browser initially could not complete navigation with the newly authored response export absent from stale local dist.
- `apps/frontend/package.json` now builds `packages/zod` before Playwright; standalone browser execution passed. The existing runner and report/privacy gates are unchanged.
- Commit: `cb71b8f`.

**3. [Rule 3 - Blocking] Existing canonical path inventory rejected the newly implemented routes**
- Extended `packages/openapi/src/gen.test.ts` to keep its exact inventory assertion with the two new paths, plus contract/header/Unicode checks. Original health and identity assertions remain intact.
- Commit: `cb71b8f`.

The six supporting paths above explain the exact 20-file implementation scope. Existing migration/role version tests already derive the complete migration corpus, so no version assertion needed weakening or replacement.

## Issues Encountered

The first combined integration run failed the existing `TestTracestateHTTPRedisOTLP`. The exact isolated test passed without edits, and the subsequent complete registered integration run passed. No unrelated observability behavior was changed. New source lint findings and the browser's broad error locator were corrected before final verification.

## Known Stubs

- `apps/frontend/components/empty-state.tsx:7`: “Create your first link” is intentionally visible and disabled, with an accessible current-availability explanation. Plan **02-14 must enable it and wire actual link creation**. This plan implements named workspace bootstrap and authorized onboarding/empty-state presentation; D-04's complete actionable link-creation journey remains pending that slice. No absent production endpoint or link data source is fabricated.

## External Acceptance and Requirement Limits

The local signed provider fixture proves browser→Go→PostgreSQL boundaries, not live provider factors, deployed cookie/session behavior or hosted CI. Live provider acceptance remains assigned to **02-35**. Requirement traceability is deliberately unchanged until complete phase proof; this slice does not close the broader tenancy/security requirements or deliver invitations, link mutations, redirects or analytics.

## Self-Check: PASSED

All six created implementation files and the generated outputs exist. Both task commits exist, contain no tracked-file deletions, and verification passed before state advancement. No untracked implementation or runtime report remains.
