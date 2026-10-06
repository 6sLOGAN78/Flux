---
phase: 01-foundation-stability-system-boundaries
plan: "17"
subsystem: infra
tags: [go, opentelemetry, lifecycle, privacy, vendor-neutral]
requires:
  - phase: 01-12
    provides: Readiness-first serial drain and bounded reverse-order cleanup
  - phase: 01-13
    provides: Owned bounded trace, metric and log providers
  - phase: 01-14
    provides: Compatible settings mapping and sanitizing logger
  - phase: 01-15
    provides: Injected HTTP telemetry
  - phase: 01-16
    provides: Injected database and job telemetry
provides:
  - Role-owned providers injected into HTTP, logs, database and jobs
  - Provider-last cleanup including partial startup and cleanup failures
  - Migrator PostgreSQL close and provider flush under one shared exit deadline
  - Parsed binary/module/authored-import regression excluding New Relic SDKs
affects: [01-18, 01-19, 01-20]
tech-stack:
  added: []
  patterns: [provider-first registration, provider-last flush, shared one-shot exit context]
key-files:
  created: [apps/backend/internal/app/vendor_test.go, apps/backend/cmd/migrator/main_test.go]
  modified: [apps/backend/internal/app/api.go, apps/backend/internal/app/roles_test.go, apps/backend/internal/server/server.go, apps/backend/cmd/migrator/main.go, apps/backend/go.mod, apps/backend/go.sum, apps/backend/internal/database/migrator.go, apps/backend/internal/logger/logger.go, apps/backend/internal/logger/logger_test.go, apps/backend/internal/middleware/middlewares.go, apps/backend/internal/middleware/rate_limit.go, apps/backend/internal/observability/redaction.go, apps/backend/cmd/api/main.go, apps/backend/cmd/redirector/main.go, apps/backend/cmd/worker/main.go, apps/backend/cmd/flux/main.go]
key-decisions:
  - Register each role's telemetry owner before infrastructure and inject its APIs without SDK globals; flush after the established serial HTTP/job/dependency cleanup.
  - Keep legacy configuration keys and inert constructor signatures, remove vendor-typed compatibility methods and the server's obsolete LoggerService field.
  - Preserve exact migration versions as bounded log-only fields and shutdown deadline diagnostics as a fixed allowed stage.
requirements-completed: [PLAT-07, SAFE-01]
duration: 9min
completed: 2026-10-06
---

# Phase 1 Plan 17: Role-Owned Telemetry and Vendor Retirement Summary

**API, redirector, worker and migrator own injected telemetry, flush after dependency cleanup, and compile without New Relic SDK dependencies.**

## Performance

- Started: 2026-10-06T15:25:54Z
- Implementation completed: 2026-10-06T15:34:32Z
- Duration: approximately 9 minutes
- Tasks: 2
- Implementation/test/dependency files: 18

## Accomplishments

- The existing shared role factory creates exactly one `observability.New` owner from the complete `TelemetrySettings` mapping, registers its cleanup first, and constructs the safe logger with its injected log API. Server receives that owner; API PostgreSQL, API producer and worker consumer constructors receive it explicitly. Redirector and worker management HTTP use the existing safe request/correlation/tracing/logging/recovery middleware without allocating additional infrastructure or exposing API documentation.
- Reverse cleanup retains readiness-first HTTP/job drain and releases listeners, producers/consumers, Redis and PostgreSQL before provider shutdown. Partial allocation registers returned closers before propagating its startup error. All cleanup actions are attempted despite failures; safe wrappers retain private causes. No SDK-global provider or vendor application is initialized.
- Migrator constructs its provider before PostgreSQL work, creates a safe migration span and correlated logs, retains exact start/end versions, closes the one-shot PostgreSQL connection, ends the span, then flushes providers. PostgreSQL close and flush obtain the identical lazily created exit context using the existing migrator timeout. They cannot acquire successive fresh deadlines. Migration work retains its separate cancellation/operation deadline.
- Removed all vendor SDK module requirements/checksums, authored imports, vendor-typed logger methods and dormant middleware hooks. Compatible `new_relic` configuration still binds and emits a constant deprecation warning through the safe logger. The inert `LoggerService` log-sink adapter remains only for source-compatible legacy constructors and has no vendor dependency or ownership.
- `TestForbiddenVendorDependencies` parses successful JSON from `go list -deps -json` for all four roles and `go list -m -json all`, checking import/module/replacement paths and dependency errors. It also parses every authored Go import independently of build tags. Command failures, empty/malformed output and resolution errors fail the assertion. Separate adversarial fixtures prove hidden build-tag imports and forbidden transitive/replacement paths are rejected.

## Task Commits

1. **Task 1 RED: Specify provider ownership and vendor exclusion** — `db55047` (`test(01-17): specify role provider ownership and vendor exclusion`).
2. **Task 2 GREEN: Implement and verify role telemetry** — `1bd20b7` (`feat(01-17): wire owned telemetry and retire vendor SDK hooks`).

RED preceded GREEN. Neither commit deleted tracked files.

## Verification

- Initial RED failed because `roleFactories.telemetry` did not exist. Before module cleanup, the executable vendor regression explicitly failed on `github.com/newrelic/go-agent/v3` in the module graph.
- Required scoped race checks passed: `go test -race ./internal/app ./cmd/migrator -run 'Test(Role|Lifecycle|ForbiddenVendor)' -count=1` (app 19.926s, migrator 1.022s). Existing real PostgreSQL/Redis independent-role startup and binary migration assertions remain intact.
- Final `go test -race ./... -count=1` passed: app 35.717s, migrator 1.020s, database 7.762s, job 12.121s and every other suite. This includes real command SIGTERM/worker deadline regressions, request isolation, readiness, persistence, retry and telemetry tests.
- Three long-running role tests use actual injected SDK trace/log exporters and safe buffered stdout, issue real Echo requests containing secret markers, and inspect retained spans/log records after shutdown. They prove owner injection, HTTP/log trace identity, request redaction and export of a database-cleanup span before final flush. Provider/closer spies prove all role graphs, partial startup unwind, repeated-close stability, cleanup aggregation and release of partially returned provider ownership.
- Migrator factory spies prove construction order, PostgreSQL-close-before-flush order, safe failure exit codes and identical bounded cleanup contexts. Existing subprocess tests prove actual migrations still report exact versions and exit once.
- Final `go vet` across app/server/database/logger/middleware/observability and all commands, four role binary builds, `go mod tidy`, `go mod verify`, `bun run build` and `git diff --check` passed. Bun replayed two successful cached workspace builds.
- Official OpenTelemetry documentation through Context7 confirmed explicit side-effect-free provider ownership and batch shutdown. Existing pinned modules were reused; no replacement packages were installed.
- Stub and threat-surface scans found no goal-blocking placeholders, new application endpoints, auth paths or schema changes. Activated optional OTLP export uses the already declared collector boundary.

## Decisions Made

Keep the existing role graph and lifecycle stack. Provider creation precedes safe logger and infrastructure creation; API/jobs/HTTP receive the same role object. Shared role-factory changes cover redirector and worker without modifying their thin wrappers.

Use the existing migrator observability timeout for a lazily created independent exit context shared by PostgreSQL and provider cleanup. The optional database close-context argument preserves existing callers and their established behavior.

Retain legacy external configuration and inert constructor compatibility, while deleting obsolete vendor types and calls. Preserve safe operational diagnostics with fixed configuration/deadline labels and bounded integer migration versions; these versions never become metric labels.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Removed remaining prerequisite vendor adapters**
- Found during: Task 2 authored dependency scan.
- Issue: Logger's `GetApplication`/`WithTraceContext` signatures and the middleware registry still imported New Relic; the dormant rate-limit adapter retained a vendor method call. Removing only the owned module entries would leave binaries dependent on vendor types.
- Fix: Removed unused typed methods, adjusted the logger regression to assert an inert sink, injected Server.Telemetry into the middleware registry, replaced dormant vendor events with sanitized HTTP logging, and removed Server.LoggerService ownership.
- Additional files: `internal/logger/logger.go`, `internal/logger/logger_test.go`, `internal/middleware/middlewares.go`, `internal/middleware/rate_limit.go`.
- Verification: Parsed vendor regression, full race suite, vet and four binary builds.
- Commit: `1bd20b7`.

**2. [Rule 2 - Missing critical functionality] Shared migrator exit budget and test seam**
- Found during: Task 2 lifecycle inspection.
- Issue: The prerequisite database migrator independently started a fresh five-second PostgreSQL-close budget; appending provider shutdown would allow sequential fresh deadlines. The command had no injectable seam for deterministic all-four-role ownership verification.
- Fix: Optional close-context factory in `database.MigrateWithResult`; migrator creates one lazy exit context and shares it with PostgreSQL close and provider flush. Added command-local factory tests without extracting ownership into a new service.
- Additional files: `internal/database/migrator.go`, `cmd/migrator/main_test.go`.
- Verification: Shared-context/deadline spies and existing real migration subprocess suite.
- Commit: `1bd20b7` (new test file introduced in RED `db55047`).

**3. [Rule 2 - Missing critical functionality] Safe command exits with preserved diagnostics**
- Found during: Task 2 command-boundary audit and full regression runs.
- Issue: Long-running commands and compatibility shim still used raw zerolog error serialization. Migrating to the shared sanitizing logger initially discarded exact migrator schema versions and the required worker deadline marker.
- Fix: Use the shared safe logger and `SafeError` at all command boundaries; add fixed `configuration.validate` and `error.stage: deadline` policy values plus nonnegative int32-bounded log-only migration versions. Preserved existing assertions and changed no product contract.
- Additional files: `cmd/api/main.go`, `cmd/redirector/main.go`, `cmd/worker/main.go`, `cmd/flux/main.go`, `internal/observability/redaction.go`.
- Verification: Final full backend race suite including exact versions and safe worker deadline status, vet and binary builds.
- Commit: `1bd20b7`.

## Issues Encountered

Two behavior regressions exposed by the existing binary/lifecycle tests were corrected before GREEN: dropped migration version fields and dropped shutdown deadline classification. No unresolved issue remains.

## User Setup Required

None for this owned stage. Optional collector configuration remains optional and does not gate readiness.

## Next Phase Readiness

Ready for 01-18's actual HTTP → Redis retry → OTLP protobuf → collector privacy/transport proof. This summary claims injected role/provider lifecycle and the complete vendor-free source/module/binary graph, not the later collector integration. Execution stops after 01-17.

## Self-Check: PASSED

SUMMARY, both created test files and all implementation files exist. RED `db55047` and GREEN `1bd20b7` exist in order. No unexpected deletions, goal-blocking stubs or untracked runtime artifacts remain.
