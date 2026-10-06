---
phase: 01-foundation-stability-system-boundaries
plan: "12"
subsystem: infra
tags: [go, lifecycle, sigterm, echo, asynq, readiness, testcontainers]
requires:
  - phase: 01-03
    provides: Safe reverse-order cleanup and aggregated private error causes
  - phase: 01-09
    provides: Role-owned resource graphs and partial-startup unwind
  - phase: 01-10
    provides: Independently built role commands
  - phase: 01-11
    provides: Generated role-specific liveness and readiness
provides:
  - Shared SIGINT/SIGTERM coordinator with readiness-first shutdown and one overall deadline
  - Worker intake stop and active-job drain with reachable unready management listener
  - Real role subprocess, active HTTP, active Redis job, deadline and partial-startup regressions
affects: [operational-lifecycle, deployment, 01-13]
tech-stack:
  added: []
  patterns: [atomic readiness gate, serial ownership pipeline, bounded caller exit, single supervisor signal context]
key-files:
  created: [apps/backend/internal/app/lifecycle.go, apps/backend/internal/app/lifecycle_test.go]
  modified: [apps/backend/internal/app/api.go, apps/backend/internal/app/worker.go, apps/backend/internal/handler/health.go, apps/backend/internal/lib/job/job.go, apps/backend/cmd/api/main.go, apps/backend/cmd/redirector/main.go, apps/backend/cmd/worker/main.go, apps/backend/cmd/flux/main.go]
key-decisions:
  - Preserve serial cleanup after a bounded caller returns so an uncooperative dependent cannot race shared-resource closure; process commands return failure and exit at the deadline.
  - Bound Asynq idle polling by the configured worker drain budget while retaining the default interval for longer budgets.
patterns-established:
  - All long-running commands use app.NotifyContext; RoleRuntime.Run owns the single drain/cleanup context and commands do not grant another budget.
  - Readiness gates bypass probes during shutdown and recheck before returning an in-flight probe response.
requirements-completed: ["PLAT-04"]
duration: 11min
completed: 2026-10-06
---

# Phase 1 Plan 12: Bounded Role Shutdown Summary

**SIGINT and SIGTERM mark roles unready before stopping intake, drain active HTTP and Redis-backed jobs, and close owned resources in reverse order under one overall deadline.**

## Performance

- Started: 2026-10-06T14:18:18Z
- Completed: 2026-10-06T14:28:53Z
- Duration: approximately 11 minutes
- Tasks: 2
- Source files changed: 10

## Accomplishments

- Added a shared `signal.NotifyContext` coordinator for interrupt and SIGTERM, an atomic readiness gate and idempotent shutdown results. The gate flips before intake stop; drain and cleanup callbacks receive the same context. Stage/resource errors expose fixed safe names while preserving internal causes through `errors.Is` and joined errors.
- API and redirector call the existing HTTP `Shutdown` adapter, which force-closes connections on deadline. Worker first calls Asynq `Stop`, waits for its active consumer shutdown, then drains management HTTP and closes consumer/producer resources before role-owned Redis. Management `/ready` returns generated 503 `not_ready` and `/live` remains 200 while a job drains.
- Preserved existing cleanup aliases and reverse-order resource stack. HTTP shutdown remains registered in cleanup so worker-drain failures and partial startup still attempt listener/HTTP release. Removed the commands' redundant second timeout; runtime owns shutdown and commands decide the process exit status.
- Added real PostgreSQL/Redis partial-startup tests for constructed API producer resources and worker listener/consumer resources, plus deterministic fault-only tests proving every failing closer is attempted, all causes remain discoverable, readiness precedes stop, and the same deadline reaches each stage.

## Task Commits

1. **Task 1: Specify and prove the owned behavior** — `e8d5ace` (`test(01-12): specify readiness-first bounded lifecycle order`).
2. **Task 2: Implement drain on SIGTERM and close every resource** — `7ca3e0c` (`feat(01-12): coordinate readiness-first bounded role shutdown`).

RED precedes GREEN. No tracked files were deleted.

## Verification

- RED: the prescribed scoped race command failed because the planned `Lifecycle` interface did not exist; committed before implementation.
- GREEN: `go test -race ./internal/app -run 'Test(Lifecycle|SIGTERM|PartialStartup)' -count=1` passed (16.96 seconds before adding the final partial-startup regression).
- Final `go test -race ./... -count=1` passed all packages, including app (32.23 seconds), handler (3.99 seconds), existing role constructors/commands, health behavior, cleanup aggregation, migrations and contracts.
- Actual API, redirector and worker commands are built with `-race` and launched from unrelated directories. API/redirector reach readiness, receive SIGTERM, exit successfully within the configured bound and release their listeners; PostgreSQL backend connections return to zero.
- Active HTTP proof uses separate race-instrumented subprocess fixtures with the actual API/redirector composition, real pinned PostgreSQL for API, and only a deterministic `/active` handler injection. The flushed `started` response proves the handler entered before SIGTERM; the `completed` response proves it drains before process exit. In-process actual Echo/HTTP tests independently prove deadline force-close cancels an active handler. Plain partially received headers are not treated as entered-handler evidence.
- Actual worker proof uses pinned Redis 8.10.2 and a local HTTP email transport held during delivery. After SIGTERM, management reports 503 readiness and 200 liveness; a second task remains unclaimed. Releasing the first produces exactly one processed job, one pending task, zero registered consumers and clean exit. A separately stalled delivery with a 300 ms budget produces safe deadline diagnostics and a bounded nonzero process exit.
- Real PostgreSQL 17.11 and Redis 8.10.2 partial-startup failures release constructed pools, clients and listeners. Narrow factory/closer faults are used only for deterministic error injection; pinned helper image references and module sums are unchanged.
- `go vet ./internal/app ./internal/lib/job ./internal/handler ./internal/server ./cmd/...`, all five command builds, `go mod verify`, `bun run build` and `git diff --check` passed. Bun's existing workspace build replayed two successful cached package builds.
- Consulted Context7's official Asynq documentation and exact installed v0.26.0 source for `Stop`, `Shutdown`, shared Redis ownership and idle polling. No dependencies were installed or changed. Stub and threat-surface scans found no goal-blocking placeholders, library fatal calls, secret-bearing lifecycle diagnostics or undeclared production endpoints.

## Decisions Made

One bounded caller wraps the serial ownership pipeline. Cooperative closers complete in reverse order despite errors. An arbitrary closer that ignores cancellation cannot safely be forcibly interrupted while also guaranteeing later shared resources remain valid: the caller returns a deadline failure, the command exits, and the sequential pipeline only advances when that closer returns. Tests release an injected uncooperative closer and prove later callbacks then run in order. This is an explicit limitation for noncooperative callbacks, not a claim that all closers finished before forced process termination.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Minimal lifecycle prerequisites outside the owned file union**
- Existing health handlers had no shutdown gate; jobs exposed only combined shutdown; commands created a second drain context.
- Added `HealthHandler.SetReadinessGate`, `JobService.StopIntake`/`Drain`, and shared signal wiring in API, redirector, worker and compatibility Flux commands. These six prerequisite files were explicitly authorized by the orchestrator; no domain or architecture expansion occurred.
- Verification: real process tests, active work tests, existing role tests and full race suite.
- Commit: `7ca3e0c`.

**2. [Rule 1 - Bug] Idle Asynq stop could consume the entire shutdown budget**
- The first full race run showed existing one-second worker shutdown tests timing out: Asynq's default idle polling sleeps for 0.5–1.5 seconds before `Stop` completes, delaying resource cleanup.
- Set consumer `TaskCheckInterval` to `min(1 second, configured worker drain timeout / 10)`. Longer budgets retain the default interval; short budgets reserve time for active drain and cleanup. Test budgets were not increased.
- File: `apps/backend/internal/lib/job/job.go`; verification: final full race suite passed.
- Commit: `7ca3e0c`.

The existing redirector constructor and server HTTP shutdown adapter already satisfy their interfaces and remain unchanged. Common cleanup remains infrastructure-neutral to avoid import cycles.

## Issues Encountered

Initial process-test assumptions about partially received HTTP headers and Asynq completed-task retention were corrected: net/http rejects incomplete new requests during shutdown, and nonretained completed jobs are counted by `Processed`, not `Completed`. Final active-handler evidence uses an entered-handler response marker. No deferred issues or authentication gates remain.

## User Setup Required

None. Existing role drain-timeout configuration controls supervisor shutdown.

## Next Plan Readiness

Ready for 01-13. No later plan was executed. Product routes and redirect data-plane resources remain their previously declared later-phase scope.

## Self-Check: PASSED

All ten changed source files and this summary exist. RED `e8d5ace` and GREEN `7ca3e0c` exist, in order, without tracked deletions. Final full race, vet, build, module and workspace checks passed. Intentional source files are committed and no generated runtime files remain untracked.
