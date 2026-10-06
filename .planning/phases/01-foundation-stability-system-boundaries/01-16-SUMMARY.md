---
phase: 01-foundation-stability-system-boundaries
plan: "16"
subsystem: infra
tags: [go, pgx, asynq, redis, opentelemetry, privacy, correlation]
requires:
  - phase: 01-13
    provides: Injected private telemetry providers and safe propagation
  - phase: 01-14
    provides: Safe stdout and injected log sinks
  - phase: 01-15
    provides: Validated HTTP correlation context
provides:
  - SQL-free injected PostgreSQL query spans and outcomes
  - Optional versioned welcome-job metadata with safe traceparent and UUID IDs
  - Safe legacy task migration and real Redis retry propagation
affects: [01-17, 01-18]
tech-stack:
  added: []
  patterns: [injected query tracer, versioned queue metadata, public-API legacy task migration]
key-files:
  created: [apps/backend/internal/lib/job/correlation_test.go]
  modified: [apps/backend/internal/database/database.go, apps/backend/internal/database/database_test.go, apps/backend/internal/lib/job/email_tasks.go, apps/backend/internal/lib/job/handlers.go, apps/backend/internal/lib/job/job.go]
key-decisions:
  - Migrate unsafe legacy jobs by enqueueing a canonical sanitized replacement before revoking the original; never mutate Asynq's concurrently persisted shared payload or headers.
  - Inject the narrow WelcomeEmailSender boundary and optional telemetry objects while preserving existing constructor callers and worker drain behavior.
patterns-established:
  - NewWelcomeEmailTaskContext persists only version 1 metadata, canonical valid UUIDs and cleaned traceparent IDs/flags; old constructors omit absent metadata.
  - Database tracing ignores pgx SQL/arguments and records fixed dependency, operation and outcome fields without driver error text.
requirements-completed: [PLAT-07, SAFE-01]
duration: 14min
completed: 2026-10-06
---

# Phase 1 Plan 16: Safe Database and Queue Correlation Summary

**PostgreSQL spans exclude SQL and arguments, while real Redis welcome-job execution and retry preserve validated trace/correlation IDs without private propagation state or recipient telemetry.**

## Performance

- Duration: approximately 14 minutes
- Completed: 2026-10-06T15:22:36Z
- Tasks: 2
- Implementation/test files: 6

## Accomplishments

- Replaced New Relic and local SQL-text tracing in the owned database constructor with an injected OTel query tracer. It never inspects SQL, parameters, command tags or driver error strings. Spans contain fixed PostgreSQL operation/dependency fields, safe correlation and success/error outcomes; context is reconstructed without baggage or tracestate. The existing construction/ping cleanup behavior remains intact.
- Added optional versioned metadata to welcome payloads and `NewWelcomeEmailTaskContext`. Existing `to`/`first_name` payloads and `NewWelcomeEmailTask` remain compatible. Serialization retains only canonical nonzero UUIDs and valid standard traceparent IDs/flags; invalid/unknown versions and private fields are dropped.
- Worker executions create injected consumer spans, clean parent/Link contexts, bounded retry counts and correlated logs. Email addresses and names remain in the operational delivery payload and never enter spans or logs. Error wrappers preserve private causes through Unwrap but expose only constant stage/classification messages to Asynq's persisted LastErr and logger. The Asynq logger adapter never formats untrusted library messages or provider arguments; handler panic text is also suppressed.
- Legacy unsafe metadata or headers are canonicalized through public enqueue/revoke APIs. A deterministic `safe:` task ID preserves lineage; the replacement keeps queue, remaining retry budget, timeout, deadline and retention. The original is revoked only after successful enqueue or an identical existing replacement is confirmed. Enqueue/lookup/conflict failure returns a safe error and leaves the original recoverable. Consumer migration clients share the role-owned Redis connection; the old combined constructor closes its separately owned inspector.
- Preserved the existing StopIntake/Drain and budget-aware TaskCheckInterval behavior from plan 01-12. Constructor additions are optional and source-compatible for adjacent role factories. Global vendor retirement and role provider ownership remain assigned to 01-17.

## Task Commits

1. **Task 1 RED: Specify owned behavior** — `098d954` (`test(01-16): specify private database and Redis retry correlation`).
2. **Task 2 GREEN: Implement and verify owned behavior** — `474f0e7` (`feat(01-16): propagate private database and queued job telemetry`).

RED preceded GREEN; neither commit deleted tracked files.

## Verification

- RED failed on missing optional constructor injection, context-aware task creation and consumer interfaces before implementation.
- Focused required race command passed: job 10.579s, database 3.690s. After strengthening queue/task-option assertions, full owned-package race tests passed: job 11.223s, database 8.105s.
- Digest-pinned Redis 8.10.2 producer/consumer tests execute two actual deliveries per scenario through a local fake HTTP email transport: context metadata with valid secret-bearing tracestate/baggage, untrusted legacy metadata/headers, and old payloads with no metadata. Explicit 3-second retries, 100ms delayed polling, 10ms intake polling and 15-second test deadlines bound execution.
- Tests inspect queued and retry TaskInfo payloads/headers and LastErr, assert original legacy task removal, and verify replacement queue/retry budget/timeout/deadline/retention. Captured spans and logs preserve parent trace IDs, sampled flags, UUID correlation and retry counts; span/parent/Link TraceState is empty and private propagation/provider/recipient/name markers are absent. Operational recipient/name delivery remains correct.
- Digest-pinned PostgreSQL 17.11 executes a parameterized SELECT with distinct SQL-text/argument markers and an actual driver-error marker. Returned data/error prove those values reached PostgreSQL, while injected captured spans and stdout logs contain none of them. Query outcome and attribute allowlists are checked directly.
- `go test -race ./... -count=1` passed: app 36.591s, database 7.750s, job 10.821s, and all existing suites. `go vet ./internal/database ./internal/lib/job`, `go build ./cmd/...`, `bun run build` and `git diff --check` passed.
- Official Asynq documentation through Context7 and exact installed v0.26.0 source confirmed retry configuration and shared payload ownership. No modules or packages were installed or changed. Stub and threat-surface scans found no goal-blocking placeholders, new application endpoint, authentication path or schema.

## Decisions Made

Keep trace/log provider APIs injected and preserve optional-export behavior without SDK globals. Use a narrow email delivery interface for provider-free local transport tests; existing concrete email clients satisfy it unchanged.

Legacy migration uses the existing Redis queue boundary. Enqueue and revoke are two public API operations, so this preserves Asynq's at-least-once semantics: interruption between them can cause another migration attempt or duplicate delivery. It does not claim exactly-once delivery. A failed enqueue retains the original payload for recovery; successful migration removes private metadata from subsequent retry storage. The replacement task ID changes, and already-due scheduling becomes immediate execution while its original deadline and remaining retry budget remain enforced.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Safely migrate legacy retry payloads without concurrent mutation**
- Found during: Task 2 pinned Asynq source inspection.
- Issue: Asynq v0.26.0 shares task payload/header objects with its retry message and may persist that message concurrently during cancellation or shutdown. Scrubbing only decoded context would retain secrets in persisted retries; mutating the shared buffers would introduce a race.
- Fix: Canonical sanitized replacement enqueue followed by RevokeTask, with deterministic lineage, option/retry-budget preservation, matching-payload conflict validation and safe recoverable failure handling. This uses existing public queue APIs and introduces no Redis-internal schema dependency.
- Files: `email_tasks.go`, `handlers.go`, `job.go`, `correlation_test.go`.
- Verification: Real Redis original/removal/replacement retry inspection and full backend race suite.
- Commit: `474f0e7`.

## Issues Encountered

The SDK's in-memory exporter clears captures on shutdown; retaining test wrappers preserve exported evidence. Resend reads its base URL at package initialization, so runtime environment changes cannot provide a local in-process transport. The narrow existing-method interface supplies the local fake HTTP delivery seam without changing the email package or contacting external providers in the final tests. PostgreSQL Ping does not emit a query span, so the regression explicitly checks the two actual SQL operations.

## User Setup Required

None for this stage-local boundary.

## Next Phase Readiness

Ready for 01-17 role-owned provider injection and vendor retirement. The actual HTTP → Redis → OTLP protobuf → collector proof belongs solely to 01-18; this summary claims the owned real database/queue and injected-export boundaries only.

## Self-Check: PASSED

All six owned implementation/test files and this SUMMARY exist. RED `098d954` and GREEN `474f0e7` exist in order. No unexpected deletions, goal-blocking stubs or untracked runtime artifacts remain.
