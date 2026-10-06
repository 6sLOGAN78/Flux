---
phase: 01-foundation-stability-system-boundaries
plan: "03"
subsystem: infra
tags: [go, config, cleanup, pgx, asynq, race-tests]
requires:
  - phase: 01-01
    provides: Supported Go toolchain and pinned dependency fixtures
provides:
  - Typed secret-safe configuration stage errors with preserved causes and FLUX key compatibility
  - Reverse-order concurrent-safe idempotent cleanup with joined named failures
  - Failed-ping pool disposal and instance-owned job email clients with complete shutdown
affects: [01-04, process-roles, startup-lifecycle, foundation-quality-gates]
tech-stack:
  added: []
  patterns: [private wrapped causes, reverse ownership cleanup, narrow failure injection]
key-files:
  created: [apps/backend/internal/config/config_test.go, apps/backend/internal/app/cleanup.go, apps/backend/internal/app/cleanup_test.go, apps/backend/internal/database/database_test.go, apps/backend/internal/lib/job/job_test.go]
  modified: [apps/backend/internal/config/config.go, apps/backend/internal/database/database.go, apps/backend/internal/lib/job/job.go, apps/backend/internal/lib/job/handlers.go]
key-decisions:
  - Configuration and cleanup errors expose stable stages/resource names while preserving private causes through Unwrap.
  - Keep Asynq Shutdown's void production API behind a narrow error-returning lifecycle seam for deterministic failure injection.
patterns-established:
  - Cleanup Push registers ownership and rejects registration once closing begins; Close attempts all callbacks once and returns a stable joined result.
requirements-completed: ["PLAT-02", "PLAT-04"]
duration: 4min
completed: 2026-10-06
---

# Phase 1 Plan 3: Configuration Errors and Resource Cleanup Summary

**Secret-safe typed configuration stages, reverse idempotent cleanup, failed-ping pool disposal, and instance-owned job clients with joined shutdown failures.**

## Performance

- Duration: approximately 4 minutes
- Started: 2026-10-06T06:02:16Z
- Completed: 2026-10-06T06:05:59Z
- Tasks: 2
- Source files changed: 9

## Accomplishments

- `LoadConfig` returns `ConfigError` for load, unmarshal, validate, and observability stages. `Error` hides provider values; `Unwrap` retains original errors for `errors.Is` and `errors.As`. Existing FLUX prefix, dot nesting, and underscore field names bind unchanged.
- Observability defaults are initialized before decoding overrides; its dedicated validator owns that stage and optional vendor credentials. Required transactional/service sections still undergo recursive validation.
- `app.Cleanup` registers ownership, closes in reverse order, attempts every closer after failure, preserves each failure through `errors.Join`, and provides safe resource labels. Concurrent and repeated closes execute each callback once; late registration returns `ErrCleanupClosed`.
- The database constructor uses a narrow ping/close seam that closes its allocated pool on initial ping failure and retains the original cause.
- Job services own separate email clients. `Stop` shuts down the consumer before closing the producer, joins safe named failures, and returns the same result without repeated effects. Existing callers remain compilable.

## Task Commits

1. **Task 1: Specify and prove the owned behavior** — `db65d82` (test, RED).
2. **Task 2: Implement return configuration errors and own cleanup** — `826afbd` (feat, GREEN).

Both commits used normal Git hooks. No files were deleted.

## Verification

- RED: the prescribed four-package race command failed to compile because the planned `ConfigError`, provider seam, `Cleanup`, ping seam, instance email fields, and error-returning `Stop` did not yet exist. This establishes missing planned interfaces; it is not represented as executed behavioral failures.
- GREEN: `go test -race ./internal/config ./internal/app ./internal/database ./internal/lib/job -run "Test(Config|Cleanup|DatabasePingFailureClosesPool|JobStopClosesAll)" -count=1 -v` passed. It executes each configuration stage, existing environment binding, secret-marker diagnostics, reverse/all/repeated/concurrent cleanup, late ownership rejection, failed-ping disposal, and consumer/producer/both/success shutdown cases.
- Expanded within owned packages: `go test -race ./internal/config ./internal/app ./internal/database ./internal/lib/job -count=1` passed, also checking successful ping retains its pool and separate jobs retain distinct email clients.
- `go vet ./internal/config ./internal/app ./internal/database ./internal/lib/job` passed.
- `go test ./...` passed for the complete backend. Packages without test files were compiled; this is not a claim of dependency integration coverage.
- `rg 'Fatal\(' apps/backend/internal/config` and the global email-client scan returned no matches. `git diff --check` passed. Created/modified production files contain no blocking stubs and introduce no new endpoint, file access, auth path, or schema trust boundary.
- Context7 documentation from official Koanf and Asynq repositories was consulted; installed source confirmed pinned Asynq v0.26.0 `Shutdown()` returns void and `Client.Close()` returns error.

## Decisions Made

Use safe error messages with privately wrapped causes instead of raw provider text. Preserve the current configuration loader signature, so no adjacent caller compile fixes are needed. Asynq cannot report a consumer shutdown error through its actual void API; the lifecycle callback seam allows that error path to be tested without a running worker. The producer's actual close error is returned normally.

## Deviations from Plan

No implementation or ownership deviations. Task 1 repeats final GREEN acceptance criteria while explicitly requiring preimplementation failures; the tests were committed RED before new APIs existed, and all shared final checks passed after Task 2. Two test files received small GREEN adjustments for robust environment isolation, explicit registration results, late-registration coverage, and the exact no-`Fatal(` scan.

## TDD Gate Compliance

RED `db65d82` precedes GREEN `826afbd`. RED failures were missing-interface compile failures, explicitly recorded above; the complete final behavior suite passed after implementation. No separate refactor commit was necessary.

## Issues Encountered

None remaining in this slice. An initial GREEN attempt used validator partial selection, which omitted nested required fields; the new incomplete-configuration test caught it. Recursive validation now excludes only the separately validated observability section.

## User Setup Required

None. Tests inject narrow failures without starting a database or worker process.

## Next Plan Readiness

Ready for plan 01-04 and subsequent role composition. Those plans must integrate the cleanup stack into partial startup and role shutdown, handle returned job-stop errors, and supply bounded drains. The existing coupled server still owns its earlier lifecycle behavior; this slice does not claim role-level integration. The existing empty migration scaffold remains assigned to plan 01-04. Requirement completion metadata covers this plan's allocated PLAT-02/PLAT-04 slice; phase-level acceptance remains pending.

## Self-Check: PASSED

All five created files exist, both task commits exist, changes are restricted to the nine declared source files, the exact acceptance scans pass, and all final race/vet/backend checks pass.
