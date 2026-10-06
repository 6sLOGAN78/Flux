---
phase: 01-foundation-stability-system-boundaries
plan: "04"
subsystem: database
tags: [go, postgres, tern, migrations, config, testcontainers]
requires:
  - phase: 01-01
    provides: Verified Go toolchain and digest-pinned PostgreSQL
  - phase: 01-03
    provides: Typed safe configuration failures and reverse cleanup stack
provides:
  - PostgreSQL-only one-shot migrator with observed start/end schema versions
  - Explicit non-interactive migration and repeated CI migration commands
  - Role-aware configuration preserving existing FLUX environment names
  - Real pinned PostgreSQL migration, executable, failure, and cleanup regressions
affects: [01-09, 01-10, 01-11, 01-12, foundation-quality-gates]
tech-stack:
  added: []
  patterns: [safe typed migration errors, explicit process roles, observed schema versions, bounded cleanup]
key-files:
  created: [apps/backend/internal/database/migrator_test.go, apps/backend/cmd/migrator/main.go]
  modified: [apps/backend/internal/config/config.go, apps/backend/internal/config/config_test.go, apps/backend/internal/database/migrator.go, apps/backend/internal/database/migrations/001_setup.sql, apps/backend/taskfile.yml]
key-decisions:
  - Preserve Migrate's error-only compatibility facade and expose MigrateWithResult for one-shot callers.
  - Make the existing bootstrap migration an explicit harmless SELECT 1 without product tables.
  - Migrator validates PostgreSQL and common logging/timeout configuration independently of HTTP, Redis, Clerk, Resend, and vendor credentials.
patterns-established:
  - MigrationError exposes stable operation names while retaining private causes through Unwrap.
  - Migration connections register immediate cleanup and close with an independent bounded context after success or failure.
requirements-completed: [PLAT-02, PLAT-01]
duration: 9min
completed: 2026-10-06
---

# Phase 1 Plan 4: Explicit Deterministic Migrations Summary

**PostgreSQL-only one-shot migrations report actual schema versions, preserve safe wrapped failures, and run repeatedly through a non-interactive CI task.**

## Performance

- Started: 2026-10-06T13:15:22Z
- Completed: 2026-10-06
- Duration: approximately 9 minutes
- Tasks: 2
- Source files changed: 7

## Accomplishments

- Added typed API, redirector, worker, and migrator roles. `LoadConfig` retains API-compatible behavior; `LoadConfigForRole` keeps existing FLUX prefix, dot nesting, and underscore names. Migrator accepts only primary/PostgreSQL settings with safe logging defaults and validates database ports, logging format/level, and operation timeout. HTTP, Redis, Clerk, Resend, and monitoring license settings can be omitted.
- Added `MigrateWithResult`, returning observed start/end versions, while retaining the existing `Migrate` facade for current API/test-helper callers. Correct URL userinfo encoding supports special-character passwords. Connect, construction, load, version reads, migration, and close failures retain causes behind safe operation messages. All allocated connections register cleanup immediately and close independently of operation cancellation.
- Added `cmd/migrator`, which configures a local logger, handles SIGINT/SIGTERM, bounds migration duration using the existing observability timeout, completes cleanup before `main` selects process exit status, and starts no HTTP or queue consumer.
- Added explicit `migrate`, repeated `migrate:check`, and a compatible `migrations:up` alias. Migration tasks consume structured FLUX database keys, require no confirmation, and invoke the embedded migration binary rather than exposing a connection string to CLI arguments.

## Task Commits

1. **Task 1: Specify and prove the owned behavior** — `68331c8` (test, RED).
2. **Task 2: Implement run explicit deterministic migrations** — `121a098` (feat, GREEN).

Normal Git hooks ran. No tracked files were deleted.

## Verification

- RED: `go test ./internal/config ./internal/database -run 'Test(Config|Migration)' -count=1` failed on the missing planned role loader, migration result/error types, and migration-stage interface. This was a missing-interface compile failure, not an executed behavioral failure.
- GREEN scoped behavior checks passed; final `go test -race ./internal/config ./internal/database -run 'Test(Config|Migration)' -count=1 -v` passed (config 1.028s, database 7.442s).
- The real Testcontainers suite uses `postgres:17.11-alpine@sha256:b0f9560a2de083e2cc7382e75f808c7381a32852a7ec49117deedb300e552b24` with loopback-only port binding. It observes no initial version table, then migration result `0 -> 1`, exact embedded migration count in PostgreSQL, and repeated result `1 -> 1`.
- Built the actual migrator binary, removed the test ledger, and ran it from an unrelated temporary working directory with PostgreSQL-only FLUX configuration: first run reports `0 -> 1`; second reports `1 -> 1`; both exit zero. A connection failure exits nonzero without the marker password or provider address in output.
- `task --dir <backend> migrate:check` passed with stdin at EOF, proving the CI path requires no interactive confirmation.
- A real PostgreSQL trigger injects migration failure with a secret marker; `errors.As` still retrieves `pgconn.PgError`, while public diagnostics omit the marker. Querying `pg_stat_activity` confirms zero leaked one-shot connections after successful and failing runs.
- Narrow failure tests cover load, start-version, apply, end-version, canceled connection, and cleanup stages, with safe messages and retained causes. Cleanup receives an independent non-canceled bounded context. Existing configuration stage/environment compatibility tests remain passing.
- `go vet ./internal/config ./internal/database ./cmd/migrator`, `go build ./cmd/...`, and `git diff --check` passed. A transitive `go list -deps ./cmd/migrator` scan found no router, handler, service, or Asynq package dependencies.
- Context7 documentation from official Tern/pgx repositories was consulted, and pinned Tern v2.4.1 source was checked for the embedded migration APIs. No dependencies were installed or changed.
- Production stub and threat-surface scans found no blocking stubs or additional network/auth/file-access trust boundaries outside the plan register.

## Decisions Made

Preserve the existing migration function signature through a compatibility facade so this bounded slice requires no edits to the current API composition root or test helpers. The migrator uses its own logger and never initializes optional vendor telemetry. Its existing observability health-check timeout currently supplies the operation deadline; operators can override that duration through the preserved environment key.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Make the existing bootstrap migration executable**

- Found during: Task 2, resolving the prerequisite recorded by plan 01-01.
- Issue: `001_setup.sql` contained only comments; Tern rejects it with `no sql in forward migration step`, preventing both explicit migration and migrating test fixtures.
- Fix: Added explicit harmless `SELECT 1` statements to the forward/down bootstrap. No product schema was introduced and invalid migration errors remain failures.
- Files: `apps/backend/internal/database/migrations/001_setup.sql`, the only additional source file beyond the six-file ownership union, explicitly authorized by the orchestrator.
- Verification: Real empty-database and repeated migrations pass with exact version 1; actual migration failures remain typed and secret-safe.
- Commit: `121a098`.

## TDD Gate Compliance

RED `68331c8` precedes GREEN `121a098`. Both plan tasks were completed in this sequence; no separate refactor commit was needed.

## Issues Encountered

None remaining. No authentication gates occurred. Migration construction failures are wrapped in production; behavioral failure injection covers the remaining mutable migration stages and real database apply errors.

## User Setup Required

None for verification. Operators use the existing FLUX primary/PostgreSQL environment keys when invoking the command; `migrations:up` now shares that structured configuration instead of the former separate `FLUX_DB_DSN` CLI path.

## Next Plan Readiness

Explicit migrations and migrating fixtures now work. Subsequent process-role plans can adopt the role loader and must remove the legacy API's implicit migration call in their assigned slice. This summary completes plan 01-04's allocated PLAT-01/PLAT-02 behavior; independent API/worker/redirector composition remains assigned to later plans. Phase 1 remains in execution.

## Self-Check: PASSED

Both created files and all seven source outputs exist. Both task commits exist, RED precedes GREEN, no source deletions were committed, and final scoped race/vet/build/import/whitespace checks passed.
