---
phase: 01-foundation-stability-system-boundaries
plan: "09"
subsystem: infra
tags: [go, echo, process-roles, postgres, redis, asynq, cleanup]
requires:
  - phase: 01-03
    provides: Safe configuration errors and reverse idempotent cleanup
  - phase: 01-04
    provides: Role-aware configuration and explicit migration command
  - phase: 01-06
    provides: Canonical generated health transport for subsequent route integration
  - phase: 01-07
    provides: Embedded documentation and email assets
provides:
  - Independent API, redirector and worker constructors with Run and Close
  - Explicit producer opt-in and consumer-only worker processing
  - Role listener/deadline settings with compatible FLUX_SERVER fallbacks
  - Every-stage startup failure cleanup with safe joined diagnostics
affects: [01-10, 01-11, 01-12, role-composition, observability]
tech-stack:
  added: []
  patterns: [explicit owned resource graphs, shared role-owned Redis, infrastructure-neutral cleanup aliases]
key-files:
  created: [apps/backend/internal/app/api.go, apps/backend/internal/app/redirector.go, apps/backend/internal/app/worker.go, apps/backend/internal/app/roles_test.go, apps/backend/internal/lifecycle/cleanup.go]
  modified: [apps/backend/internal/server/server.go, apps/backend/internal/config/config.go, apps/backend/internal/config/config_test.go, apps/backend/internal/app/cleanup.go, apps/backend/internal/database/migrator.go, apps/backend/internal/database/migrator_test.go, apps/backend/internal/lib/job/job.go]
key-decisions:
  - API owns PostgreSQL and only explicitly enabled Redis/producers; worker shares one role-owned Redis client with its consumer and does not allocate a producer.
  - Keep app cleanup aliases while placing the unchanged primitive in lifecycle to prevent infrastructure-to-app import cycles.
  - Preserve FLUX_SERVER port/timeouts as role fallbacks; role overrides control listener, drain and readiness deadlines, and unused numeric/duration settings are omitted before decoding.
patterns-established:
  - Register each acquired resource immediately and unwind constructor failures using an independent bounded cleanup context.
  - Server is a dependency adapter and HTTP transport; app owns infrastructure allocation and cleanup.
requirements-completed: ["PLAT-01", "PLAT-04"]
duration: 10min
completed: 2026-10-06
---

# Phase 1 Plan 9: Role Configuration and Resource Ownership Summary

**API, redirector and worker own distinct resource graphs with immediate cleanup registration, compatible role configuration and independent HTTP listeners.**

## Performance

- Started: 2026-10-06T13:47:10Z
- Completed: 2026-10-06T13:57:00Z
- Duration: approximately 10 minutes
- Tasks: 2
- Source files changed: 12

## Accomplishments

- Added `NewAPI`, `NewRedirector` and `NewWorker`, each exposing `Run(ctx)`, `Close(ctx)`, its Echo transport and bound address. API constructs PostgreSQL and optionally Redis with enqueue-only producers when `FLUX_API.PRODUCER_ENABLED` is true. It never invokes migrations, creates an email client, or starts a consumer. Redirector constructs only logging and system HTTP in this phase. Worker constructs Redis, a configured email adapter, consumer-only Asynq processing and a management listener, with no PostgreSQL, enqueue producer, public API or docs routes.
- Server construction receives explicitly owned resources instead of allocating PostgreSQL, Redis or jobs. HTTP setup accepts the role address and existing timeout settings; shutdown drains HTTP, allowing app cleanup to close dependent resources afterwards. Failed HTTP drain attempts a forced close before returning its cause.
- Every acquired resource registers cleanup immediately. Constructor failures unwind earlier resources with an independent bounded context. HTTP/listeners close before consumers, Redis, PostgreSQL and logger resources. Cleanup attempts every registered closer, retains private causes and returns safe resource labels, including when multiple closers fail.
- Added `FLUX_API.*`, `FLUX_REDIRECTOR.*` and `FLUX_WORKER.*` listen-address/drain-timeout/readiness-timeout settings. Existing `FLUX_SERVER.PORT` and HTTP timeout names remain supported as fallbacks. Default addresses are `:8080`, `:8081` and worker management `127.0.0.1:8082`; explicit role addresses override legacy ports. Drain defaults to 30 seconds; readiness defaults to the existing observability timeout. Invalid addresses, ports or deadlines fail safely. Only used dependencies require configuration: unused Clerk, Resend, Redis and monitoring credentials do not block roles. Worker requires Redis and its current email adapter credentials. Other roles' numeric/duration settings are omitted before decoding.

## Task Commits

1. **Task 1: Specify and prove the owned behavior** — `fb8e88c` (test, RED).
2. **Task 2: Implement role configuration and resource ownership** — `51c6747` (feat, GREEN).

Normal Git commits were used. No tracked files were deleted.

## Verification

- RED: `go test ./internal/config ./internal/app -run TestRole -count=1` failed because `Config.ForRole` was missing and the existing database-to-app cleanup dependency formed an import cycle once role tests referenced the database. This was recorded as missing-interface/dependency failure before implementation, not as an executed behavioral failure.
- GREEN: the prescribed `TestRole` suite passes. Constructor spies enforce distinct allocation allowlists for default API, API producers, redirector and worker; each logger/database/Redis/producer/email/consumer/router/listener/consumer-start failure closes every earlier owned resource in reverse order. Repeated close is idempotent, joined close failures preserve causes without marker secrets, and listener cleanup is verified through immediate rebinding after consumer-start failure.
- Real integration tests use the existing digest-pinned PostgreSQL 17.11 and Redis 8.10.2 fixtures. API constructs a usable pool against empty PostgreSQL without creating `schema_version` and allocates no default Redis/jobs. Worker starts actual Asynq processing, owns no PostgreSQL or producer, closes shared Redis, and releases its listener. Redirector serves actual HTTP without PostgreSQL/Redis/jobs, returns 404 for docs and exits cleanly on context cancellation with its listener released. No email is queued or provider contacted.
- `go test -race ./internal/config ./internal/app ./internal/lib/job ./internal/database -count=1` passed, preserving existing config, cleanup, job Stop and real migration regressions.
- `go test -race ./... -count=1`, `go vet ./...` and `go build ./...` passed for the backend. After the final unused-settings decoding regression, changed-package race tests, scoped vet and the complete backend build passed again. No TypeScript source, contract or dependency changed; earlier workspace gates remain applicable.
- Existing environment binding tests still pass, and new tests exercise actual `FLUX_REDIRECTOR.*` duration/address values alongside preserved `FLUX_SERVER.READ_TIMEOUT`. Separate tests check role-specific required dependencies, defaults, overrides, invalid input and malformed unused settings.
- Context7 official Asynq guidance and the installed v0.26.0 source confirmed separate clients/servers, nonblocking `Start`, shared-client constructors and configured `ShutdownTimeout`. Shared role-owned Redis ensures failure before consumer Start can still release the underlying broker connection.
- Whitespace, production placeholder and threat-surface scans passed. Management HTTP and role resource access are covered by the plan's declared boundary; no additional unplanned endpoint, auth or schema surface was introduced.

## Decisions Made

Kept the existing server container as an adapter for handlers and middleware while moving allocation and close ownership to app. The role factory seams allow deterministic constructor failures without replacing production dependency behavior.

Used Asynq's shared Redis-client constructors for the separated producer and consumer. The app closes the shared client once after dependent job processing; the existing combined `NewJobService` remains compatible. Optional vendor services and unused Clerk initialization are omitted from the new graphs.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Extract the shared cleanup primitive to avoid composition import cycles**

- Found during: Task 1 RED and Task 2 construction.
- Issue: Both database migrations and jobs imported `app.Cleanup`; composing those packages from app produced cycles. The database package's own migration test import formed the same cycle.
- Fix: Moved the existing cleanup implementation unchanged into `internal/lifecycle`, retained app type/error aliases, and changed only cleanup imports in the migrator, its test and job package. Existing cleanup, migration and Stop regressions remain passing.
- Files: `internal/lifecycle/cleanup.go`, `internal/app/cleanup.go`, `internal/database/migrator.go`, `internal/database/migrator_test.go`, `internal/lib/job/job.go`.
- Verification: Scoped and whole-backend race suites, vet and build pass.
- Commit: `51c6747`.

**2. [Rule 3 - Blocking] Separate current job producer and consumer allocation**

- Found during: Task 2 defining the API/worker resource graphs.
- Issue: `NewJobService` always constructs both the enqueue client and consumer plus email transport, preventing the declared role boundaries and risking broker leakage if construction fails before consumer startup.
- Fix: Added narrow `NewProducer` and `NewConsumer` constructors over the role-owned shared Redis client, kept existing combined construction compatible, and made Stop handle either side while preserving safe aggregation. Worker consumer shutdown uses its configured drain timeout.
- Files: `internal/lib/job/job.go`.
- Verification: Existing Stop tests pass; real worker and constructor-spy tests prove the required resource graph.
- Commit: `51c6747`.

These prerequisite edits were explicitly authorized by the orchestrator. Five files outside the seven-file ownership union were necessary; no libraries, schema or broader architecture changed.

## TDD Gate Compliance

RED `fb8e88c` precedes GREEN `51c6747`. Both plan tasks were completed; no separate refactor commit was necessary.

## Issues Encountered

The previous migrator configuration regression expected API to reject omitted unused secrets. Updated that assertion to the new role contract; required worker/database and unknown-role validation remain covered. No authentication gates occurred.

## User Setup Required

None for verification. Production role constructors consume `LoadConfigForRole`; worker email credentials and declared dependencies must be supplied for that role.

## Next Plan Readiness

Ready for 01-10 to wire thin binaries and replace the legacy combined command. The old `cmd/flux` remains compilable; its adoption of the new graphs and removal of its own implicit migration belong to 01-10. System health routes are assigned to 01-11; redirector and worker currently own the intended HTTP shell without public product routes. Signal handling, unready-first shutdown and full process drain ordering remain assigned to 01-12. This summary covers the allocated PLAT-01/PLAT-04 slice; phase-level acceptance remains pending.

## Self-Check: PASSED

All twelve source outputs and this summary exist. Commits `fb8e88c` and `51c6747` exist, RED precedes GREEN, no tracked deletions were committed, and final scoped race/vet/build checks pass. No generated runtime files remain untracked.
