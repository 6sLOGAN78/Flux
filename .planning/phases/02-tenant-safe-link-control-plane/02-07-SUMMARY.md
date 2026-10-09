---
phase: 02-tenant-safe-link-control-plane
plan: "07"
subsystem: workspace
tags: [postgresql, tenancy, idempotency, permissions, nextjs]
requires:
  - phase: "02-06"
    provides: Atomic workspace bootstrap and authorized Links empty state
provides:
  - Fresh membership authorization before bootstrap hash comparison or replay
  - Consumed bootstrap ledger helpers with actor-scoped expired-record cleanup
  - Closed workspace capability matrix and native authorized Links shell
affects: [02-08, 02-10, 02-16, 02-17, 02-35]
tech-stack:
  added: []
  patterns: [authorize before replay, scoped retention cleanup, closed capability matrix]
key-files:
  created:
    - apps/backend/internal/repository/mutation.go
    - apps/backend/internal/service/permissions.go
    - apps/backend/internal/service/permissions_test.go
    - apps/backend/internal/service/workspace_test.go
    - apps/frontend/components/app-shell.tsx
    - apps/frontend/lib/workspace.ts
    - apps/frontend/lib/workspace.test.ts
  modified:
    - apps/backend/internal/repository/repositories.go
    - apps/backend/internal/repository/workspace.go
    - apps/backend/internal/service/workspace.go
    - apps/backend/internal/service/services.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/database/migrator_test.go
    - apps/frontend/app/workspaces/[workspaceId]/links/page.tsx
    - apps/frontend/package.json
    - apps/frontend/tests/control-plane.spec.ts
key-decisions:
  - Authorize current membership before comparing a bootstrap retry hash; removed actors receive the same safe 404 for matching and changed requests.
  - Consume ledger helpers only in existing bootstrap operations; cleanup is actor-scoped and removes records only after their database-enforced retention window.
  - Present authorized Team availability without inventing a Team route before plan 02-16.
requirements-completed: []
requirements-addressed: [TEN-02, TEN-05, TEN-07, TEN-08, SAFE-03]
duration: 15min
completed: 2026-10-09
---

# Phase 2 Plan 7: Workspace Retry and Capability Refinements Summary

**Workspace retries authorize current membership before revealing hash conflicts, while PostgreSQL cleanup respects actor scope and the documented role matrix protects transactional capabilities.**

## Performance

- Duration: approximately 15 minutes, including verification and private failure diagnosis.
- Tasks: 2/2.
- Implementation files: 16 unique paths, consisting of 12 declared paths and four necessary supporting paths.

## Accomplishments

- Bootstrap lookup/record/hash helpers are consumed by the existing workspace repository. Same-key creation remains serialized on durable user identity. Replay releases that actor lock, takes the existing workspace-first authorization path, then compares the hash. Removed membership denies matching and changed retries without revealing ledger content. Authorized changed payloads still return 409; atomic rollback and concurrent single creation remain verified.
- Actor-scoped cleanup deletes only records whose `retain_until` has passed. Database minimum retention remains 24 hours; expired keys may start a new operation without deleting the original workspace. Foreign actors' expired records remain untouched. No generic unconsumed mutation implementation or future endpoint was added.
- Services consume the explicit Users/Workspaces repository registry. The closed capability matrix grants read to all four roles, ordinary writes to owner/admin/member, and Team capabilities to owner/admin. Role-change policy allows owners all recognized targets and restricts admins to member/viewer current and proposed roles. Workspace locking remains shared for ordinary writes and exclusive for Team operations.
- Actual PostgreSQL tests cover every role/capability pair, rejected unknown roles, a write blocked behind an exclusive workspace lock that observes committed demotion, historical creator/audit identity surviving membership removal, and a transaction-created composite FK probe rejecting a foreign workspace/membership pair with 23503. The probe rolls back entirely; future dependent production entities remain assigned to their own slices.
- The reachable Links page renders native header/navigation/main landmarks only after current server authorization. Links has `aria-current`; owners/admins see truthful disabled Team availability without an absent route. Loading, errors, retry, session binding, focus revalidation, foreign denial and the existing empty state remain intact.
- Migration proof now explicitly upgrades the existing 002 identity prefix to 003 and checks all five workspace tables. The original empty database, 001-prefix, repeated no-op, actual binary, rollback, privacy and resource tests remain intact.

## Task Commits

1. **Specify retry authorization and shell behavior:** `6cf2a9f` — meaningful PostgreSQL RED expected 404 for a removed actor's changed retry but observed 409. The extended workspace browser specification also failed before navigation existed.
2. **Deliver retry, capability and constraint refinements:** `51bdf0d` — consumed ledger/registry wiring, permission matrix, scoped cleanup, native shell and automated security/recovery proof.

## Verification

- `bun run test:unit`: passed all **87 discovered race-enabled Go top-level unit tests**, all **four workspace test suites**, tooling self-tests and script regressions. All four dependency-ordered workspace production builds also passed.
- `bun run test:integration`: passed all **17 registered top-level tests across six groups**, using real pinned PostgreSQL/Redis and race-enabled execution. New database cases remain subtests of registered `TestProductActualHTTP`; no unregistered container-backed test was introduced.
- Targeted real PostgreSQL `TestProductActualHTTP/workspace-bootstrap`: passed with its **six explicit subtests** after fixing test wiring; full integration subsequently passed.
- `bun run test:e2e -- --project=local --grep workspace`: **1 completed, zero skipped/flaky**. Lost committed creation response retries preserve the same key/name; authorized Links and truthful shell render, and foreign access clears workspace content/navigation.
- `bun run test:e2e -- --project=local`: **14 completed, zero skipped/flaky**, preserving the existing provider/session/browser corpus.
- `bun run lint`, `bun run format:check`, `bun run generate:check`, and `go build ./cmd/...`: passed. Canonical authored schemas and all generated bytes remain unchanged.
- Final Next production rebuild passed, then the existing validated build cleanup ran. `CI=true bun run scan`: dependency, complete worktree and full Git-history scans passed. The unused OpenPGP inventory advisory GO-2026-5932 remains visible; no vulnerable package is imported by any role or test. Runtime/tool/provider/React pins and all scan enforcement remain unchanged.
- `git diff --check`: passed. Task commits contain no tracked-file deletions.

## Deviations from Plan

**1. [Rule 3 - Blocking] Four necessary consumed-composition and reachability paths**
- `apps/backend/internal/repository/workspace.go`: consume the declared ledger helper and place hash comparison after fresh authorization in the existing operation.
- `apps/backend/internal/service/services.go`: consume the declared repository registry instead of leaving it as dead code.
- `apps/frontend/app/workspaces/[workspaceId]/links/page.tsx`: consume the declared shell and capability helper in the already reachable page.
- `apps/frontend/package.json`: include the new `components/` source in existing format/lint gates; preserve all runtime/dependency/browser commands.
- Commit: `51bdf0d`. These four paths explain the exact 16-file plan scope. Existing product adapters and registered integration dispatch suffice, so `product.go` and `scripts/check.ts` require no edits.

## Issues Encountered

The new FK proof initially attempted a temporary table referencing a permanent table, which PostgreSQL rejects. Its failed assertion exposed missing unconditional test-transaction rollback and delayed pool cleanup. The corrected transaction-created probe always rolls back, and targeted/full real-PG execution completed successfully. Frontend tests use the existing `node:test`/Playwright assertion imports after Next rejected `bun:test` types. Owned lint findings were fixed without weakening gates. No unrelated observability failure recurred.

## Known Stubs and Remaining Acceptance

- `apps/frontend/components/app-shell.tsx:24`: authorized Team is intentionally disabled with an explanation; plan **02-16** must wire the actual Team endpoint/screen. It does not pretend that Team operations already exist.
- Existing `apps/frontend/components/empty-state.tsx:7`: first-link creation remains disabled until **02-14**. D-04's complete actionable create flow remains pending; this slice preserves the authorized empty Links journey.
- Owner-preservation effects, invitations, link mutations and workspace switching arrive in later bounded plans. This plan verifies their common capability policy, without claiming those endpoints or the complete phase requirements.
- Local signed fixtures prove browser→Go→real PostgreSQL boundaries. Live provider factors/session/cookie acceptance remains assigned to **02-35**. No external auth gate occurred. Requirement traceability stays unchanged until complete phase proof.

## Self-Check: PASSED

All seven created files exist; `6cf2a9f` and `51bdf0d` exist. Exact file staging, no tracked deletions, generated-byte preservation and passing verification were checked before state advancement. No untracked implementation or runtime report remains in the repository.
