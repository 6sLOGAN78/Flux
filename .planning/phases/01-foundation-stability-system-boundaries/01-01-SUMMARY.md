---
phase: 01-foundation-stability-system-boundaries
plan: "01"
subsystem: infra
tags: [go, node, bun, postgres, redis, compose, testcontainers]
requires: []
provides:
  - Exact supported runtime pins and one frozen Bun dependency graph
  - Digest-pinned loopback-only PostgreSQL and Redis local services
  - Matching real PostgreSQL and Redis test factories
affects: [01-04, foundation-quality-gates, integration-tests, ci]
tech-stack:
  added: [Go 1.26.8, Node 22.23.3, Bun 1.3.14, PostgreSQL 17.11, Redis 8.10.2]
  patterns: [immutable service image parity, loopback-only dependencies, registered test cleanup]
key-files:
  created: [.tool-versions, compose.yaml]
  modified: [package.json, apps/backend/go.mod, apps/backend/internal/testing/container.go]
key-decisions:
  - Retain verified installed Node 22.23.3 LTS and Bun 1.3.14; move unsupported Go 1.25 to supported Go 1.26.8.
  - Separate raw SetupTestPostgres infrastructure from migrating SetupTestDB; retain migration failures explicitly.
patterns-established:
  - Compose and test helpers use identical exact image tags and immutable registry index digests.
  - Dependency fixture ports bind loopback; cleanup is registered before fallible setup continues.
requirements-completed: [PLAT-05]
duration: 9min
completed: 2026-10-06
---

# Phase 1 Plan 1: Reproducible Toolchain and Dependency Services Summary

**Verified Go/Node/Bun pins, one frozen Bun graph, and matching digest-pinned PostgreSQL/Redis services and test fixtures.**

## Performance

- Duration: approximately 9 minutes
- Completed: 2026-10-06
- Tasks: 2
- Source files changed: 7 (including two intentional lockfile deletions)

## Accomplishments

- Pinned Go 1.26.8, Node 22.23.3, and Bun 1.3.14 in `.tool-versions`; aligned the module directive, package-manager version, and Node engine.
- Added only PostgreSQL and Redis to Compose, with health checks, loopback ports, named volumes, and explicitly disposable local PostgreSQL credentials. The unauthenticated local Redis service is bound only to loopback.
- Updated PostgreSQL fixture readiness to wait for the final server rather than an initialization log alone. Added raw PostgreSQL and real Redis fixtures, immediate cleanup registration, and loopback-only random test ports.
- Preserved the root Bun lock unchanged; removed npm and email-local locks only after a clean install, actual package compilation, and actual email export passed.

## Task Commits

1. **Task 1: Specify and prove the owned behavior** — `2107759` (chore)
2. **Task 2: Implement pin toolchain and real dependency services** — `1208340` (chore)

Both commits used normal Git hooks. The deletions of `package-lock.json` and `packages/emails/bun.lock` are intentional.

## Release and Registry Evidence

| Component | Selection | Official evidence and verification |
|---|---|---|
| Go | 1.26.8 | [Official download metadata](https://go.dev/dl/?mode=json) lists 1.27.1 and 1.26.8. [Release policy](https://go.dev/doc/devel/release) supports the two newest major releases. Downloaded and executed `GOTOOLCHAIN=go1.26.8 go version` through Go's checksum-verified toolchain mechanism; backend `go version` then selects 1.26.8 automatically from `go.mod`. Official Linux amd64 archive SHA256: `d0f743b33e8d8945e6b1f432edd15785c70507121d6e2a723b21285eddf8b57b`. |
| Node | 22.23.3 | [Official release metadata](https://nodejs.org/dist/index.json) contains v22.23.3 with LTS name Jod. [Release schedule](https://nodejs.org/en/about/previous-releases) confirms the supported LTS line. Installed binary reports v22.23.3 and successfully runs workspace tooling. |
| Bun | 1.3.14 | [Official stable release](https://github.com/oven-sh/bun/releases/tag/bun-v1.3.14), published 2026-05-13; installed binary reports 1.3.14. Frozen install succeeds without modifying the existing root lock. |
| PostgreSQL | 17.11-alpine | [Official support policy](https://www.postgresql.org/support/versioning/) lists 17.11 as the current supported 17 minor, supported through 2029-11-08. Registry manifest and live `SELECT version()` verified 17.11. |
| Redis | 8.10.2-alpine | [Official security release](https://github.com/redis/redis/releases/tag/8.10.2), published 2026-09-17. Registry manifest and live `INFO server` verified 8.10.2. |

Verified Docker Hub multi-platform index digests, identical in Compose and fixture constants:

- PostgreSQL: `sha256:b0f9560a2de083e2cc7382e75f808c7381a32852a7ec49117deedb300e552b24`
- Redis: `sha256:3811787313eba226a2ef38658c6ccb91cd5e110edc89c37767de373120a0e5a0`

`docker buildx imagetools inspect` confirmed linux/amd64 manifests and official image provenance before configuration changed. Docker 29.1.3 and Compose 5.0.0 were available.

## Verification

- Baseline frozen install/build succeeded; uncached build confirmed that Turbo executes Zod and OpenAPI while emails has no build script.
- A clean `git archive` snapshot at `/tmp/flux-01-01-8Zbx9b` installed 552 packages with `bun install --frozen-lockfile`, compiled Zod and OpenAPI through `bun run build -- --force`, compiled emails through root `tsc --project packages/emails/tsconfig.json`, and ran `bun run --filter @flux/emails export` successfully. The selected root runtime manifest was copied into that snapshot and all checks passed again before duplicate locks were deleted.
- After deletion, root `bun install --frozen-lockfile && bun run build -- --force` passed with no lock regeneration.
- `docker compose config --quiet` passed. A dedicated `flux-plan01-01-proof` project started both pinned services with `--wait`; live PostgreSQL and Redis version queries passed. Inspection showed only `127.0.0.1:15432` and `127.0.0.1:16379` host bindings.
- `go test ./internal/testing` compiled the final helper successfully with Go 1.26.8 (repository package has no committed test files yet).
- A temporary external-package suite in the clean snapshot ran `go test -race ./internal/testing -run TestPinned -count=1 -v`: PostgreSQL version and parameterized write/read passed; Redis version, write/read, TTL, and repeated cleanup passed; actual container inspection confirmed loopback bindings. Both tests passed in 4.219 seconds. Temporary tests remain in the snapshot, outside owned source files; permanent phase regression suites belong to later plans.
- `git diff --check` and changed-file stub scans passed.

## Decisions Made

Kept supported installed Node/Bun versions to minimize toolchain churn. Selected the supported Go 1.26 patch because installed Go 1.25 was outside the official support window. PostgreSQL 17 preserves a mature supported database line; exact image digests protect against mutable tags.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Separate infrastructure readiness from application migration execution**

- Found during: Task 2 real PostgreSQL fixture smoke.
- Issue: The existing `001_setup.sql` has no forward SQL. Tern returns `loading database migrations: no sql in forward migration step`, so the migrating helper cannot currently serve as a raw infrastructure proof.
- Adjustment: Added explicit `SetupTestPostgres` and kept `SetupTestDB` as its migrating wrapper; migration errors still fail. Registered pool/container cleanup before migration failure. No migration files or product schema changed.
- Files: `apps/backend/internal/testing/container.go`.
- Verification: Raw PostgreSQL race smoke passes; original migrating path's exact failure is recorded for plan 01-04 in `deferred-items.md`.
- Commit: `1208340`.

**2. [Rule 3 - Blocking] Verify email compilation directly while its build script is absent**

- Found during: Task 1 clean build evidence.
- Issue: Root Turbo reports only two build tasks despite three implemented TypeScript packages.
- Adjustment: Ran the frozen root compiler against the email tsconfig and its real export command explicitly. Package quality-script changes remain in their owning plan.
- Verification: Email TS compilation and HTML export both passed before lock deletion.
- Source edits: None.

## Deferred Issues

- Plan 01-04 must make the existing migration executable; migrating fixtures remain correctly failing until then.
- Root/package quality scripts and remaining exact lint/migration/generation pins belong to subsequent Phase 1 plans. This summary completes only plan 01-01's allocated PLAT-05 slice.

## User Setup Required

None. Verification stopped only its dedicated Compose project afterward, preserving its containers and named volumes (`flux-plan01-01-proof_postgres_data`, `flux-plan01-01-proof_redis_data`). No user services or volumes were deleted.

## Next Plan Readiness

Subsequent Phase 1 plans can use the supported Go toolchain, frozen Bun graph, and real pinned dependencies. Phase 1 remains in execution; no phase transition occurred.

## Self-Check: PASSED

Created files and both task commits exist. Runtime manifests align, Compose/helper image references match, duplicate locks are absent, the root lock remains intact, and all scoped final checks passed. The pre-existing migration failure is explicitly deferred rather than hidden.
