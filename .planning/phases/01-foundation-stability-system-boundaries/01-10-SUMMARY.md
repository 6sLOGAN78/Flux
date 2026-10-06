---
phase: 01-foundation-stability-system-boundaries
plan: "10"
subsystem: infra
tags: [go, process-roles, echo, postgres, redis, asynq, taskfile]
requires:
  - phase: 01-04
    provides: Role-aware configuration and explicit one-shot migration command
  - phase: 01-09
    provides: Independently owned app constructors with Run and Close
provides:
  - Thin API, redirector and worker executable composition roots
  - Legacy Flux command delegates exclusively to the API role
  - Independent role run/build Taskfile targets
  - Real binary startup, configuration, dependency isolation and teardown regressions
affects: [01-11, 01-12, independent-deployment, role-startup]
tech-stack:
  added: []
  patterns: [role-specific composition roots, process exit decisions in main, fixture-bounded subprocess verification]
key-files:
  created: [apps/backend/cmd/api/main.go, apps/backend/cmd/redirector/main.go, apps/backend/cmd/worker/main.go]
  modified: [apps/backend/cmd/flux/main.go, apps/backend/taskfile.yml, apps/backend/internal/app/roles_test.go]
key-decisions:
  - Keep each executable thin and explicit over its existing app constructor; Flux is an API-only compatibility shim.
  - Use the existing Resend SDK RESEND_BASE_URL override for a local fake email transport in worker subprocess tests without adding production configuration.
patterns-established:
  - LoadConfigForRole precedes role construction; main owns exit status and signal-context cancellation, and Run/Close release owned resources before exit.
  - Build all role commands, launch them from unrelated working directories with isolated environment configuration, and verify process/listener/dependency teardown.
requirements-completed: ["PLAT-01"]
duration: 6min
completed: 2026-10-06
---

# Phase 1 Plan 10: Independent Role Binaries Summary

**API, redirector and worker binaries own their existing app role graphs; the legacy Flux command starts only API resources and migrations remain explicit.**

## Performance

- Started: 2026-10-06T13:59:27Z
- Completed: 2026-10-06T14:05:00Z
- Duration: approximately 6 minutes
- Tasks: 2
- Source files changed: 6

## Accomplishments

- Added thin `cmd/api`, `cmd/redirector` and `cmd/worker` mains. Each loads role-specific configuration, calls its existing app constructor, runs its role, closes it under the configured drain deadline, and returns safe failure diagnostics with nonzero exit status. Main owns signal handling and process exit decisions. SIGINT/SIGTERM contexts are wired; active-work SIGTERM acceptance remains assigned to plan 01-12.
- Converted `cmd/flux` to a documented API compatibility shim. It no longer calls migrations or the old service assembly. API and Flux do not require Redis, email or auth credentials by default; redirector requires no transactional/queue infrastructure; worker owns Redis and its email adapter. The existing explicit migrator was retained unchanged.
- Added `run:<role>` and `build:<role>` Taskfile targets for API, redirector, worker and migrator, plus a combined `build` target. Task commands anchor their working directory to the backend. `BIN_DIR` is overrideable and defaults to ignored `build/bin` output. Existing `run`, migration and migration compatibility targets remain usable.
- Extended `TestRoleBinaryStartup` to build all four role commands and Flux; launch the long-running binaries from temporary working directories; test role-specific ports and permitted/documentation route exposure; reject malformed role configuration with prompt nonzero exits and marker-secret-safe diagnostics; keep empty PostgreSQL unmigrated during API/Flux startup; process an actual Redis welcome-email task through a local fake HTTP email transport; repeat the explicit migrator with exact version results; and perform bounded subprocess teardown with listener, PostgreSQL connection and worker consumer-registration checks.

## Task Commits

1. **Task 1: Specify and prove the owned behavior** — `893a690` (`test(01-10): specify independent role binary startup`).
2. **Task 2: Implement wire independent thin role binaries** — `19f5d3b` (`feat(01-10): wire independent thin role binaries`).

## Verification

- RED: `go test ./internal/app -run TestRoleBinaryStartup -count=1` failed because `cmd/api`, `cmd/redirector` and `cmd/worker` did not exist. The RED commit preceded implementation.
- GREEN: `go test ./internal/app -run TestRoleBinaryStartup -count=1 -v` passed every binary startup subtest in approximately 11 seconds.
- Final: `go test -race ./internal/app -run TestRole -count=1` passed in approximately 19 seconds, including the existing resource-graph/failure-cleanup tests and binary startup suite. The binary suite performs actual combined Taskfile builds into a temporary directory and checks each role's run/build target command.
- `go build ./cmd/api ./cmd/redirector ./cmd/worker ./cmd/migrator ./cmd/flux` passed.
- `go vet ./internal/app ./cmd/api ./cmd/redirector ./cmd/worker ./cmd/migrator ./cmd/flux`, `go mod verify`, and `git diff --check` passed. Backend toolchain reported Go 1.26.8.
- Tests used the existing exact digest-pinned PostgreSQL 17.11 and Redis 8.10.2 container helpers, only where the role requires them. Disposable services and subprocess resources were torn down. The worker test contacts only the local fake email server through the SDK's documented `RESEND_BASE_URL` override, also verified against the installed v2.28.0 source; no provider credentials or new dependencies were needed.
- Production stub scan and scoped threat-surface review found no goal-blocking stubs or newly introduced endpoint/auth/file-access/schema trust boundaries. Existing role constructors own the listeners.

## Decisions Made

Keep explicit, small composition roots within the owned command files rather than introduce another shared abstraction. Preserve the tested constructor resource boundaries, retain the existing migration command, and use the SDK-supported local email endpoint override solely in verification.

## Deviations from Plan

None — plan executed within the six-file ownership union. Additional binary coverage includes the legacy Flux compatibility command, and Taskfile builds execute into temporary output to keep the repository free of generated artifacts.

## TDD Gate Compliance

RED `893a690` precedes GREEN `19f5d3b`. Both tasks completed in this sequence; no separate refactor commit was needed.

## Issues Encountered

None remaining. No authentication gates occurred.

## User Setup Required

None for verification. Operators supply the existing role-owned FLUX environment configuration when running the individual commands or Taskfile targets.

## Next Plan Readiness

Independent role command startup is complete. Sanitized liveness/readiness routes remain assigned to 01-11; unready-first active-work SIGTERM drain acceptance remains assigned to 01-12. This summary completes plan 01-10's PLAT-01 behavior, without claiming later health/drain acceptance or overall phase completion. Execution stops after this plan.

## Self-Check: PASSED

All six owned source files and the three created command entrypoints exist. Both task commits exist, RED precedes GREEN, no tracked source deletions were committed, and the final race/vet/build/checksum/whitespace checks passed. No generated runtime files remain untracked.
