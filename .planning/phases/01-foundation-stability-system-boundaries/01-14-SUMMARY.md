---
phase: 01-foundation-stability-system-boundaries
plan: "14"
subsystem: infra
tags: [go, observability, zerolog, opentelemetry, redaction, configuration]
requires:
  - phase: 01-13
    provides: Injected safe telemetry providers and shared attribute/correlation policies
provides:
  - Compatible dot-nested optional OTLP settings with typed safe validation
  - Shared sanitizing stdout JSON and injected OTel log bridge
  - Validated trace/span and UUID request/correlation logging
affects: [01-15, 01-16, 01-17, 01-18]
tech-stack:
  added: []
  patterns: [shared pre-sink allowlist, injected log API, safe legacy configuration warning]
key-files:
  created: [apps/backend/internal/logger/logger_test.go]
  modified: [apps/backend/internal/logger/logger.go, apps/backend/internal/config/observability.go, apps/backend/internal/config/config_test.go]
key-decisions:
  - Preserve legacy New Relic consumer signatures as inert adapters until plan 01-17, with no vendor initialization; inject the OTel log API through LoggerService.LogSink or NewLogger.
  - Accept existing json/console configuration values while always emitting sanitized structured JSON; reject uncontrolled messages and error details before either sink.
patterns-established:
  - ObservabilityConfig.TelemetrySettings maps optional export and fixed queue/deadline bounds into config-independent observability.Settings.
  - WithContext hooks reattach validated context IDs after event fields, and the shared writer sanitizes before stdout or OTel emission.
requirements-completed: [PLAT-07, SAFE-01]
duration: 7min
completed: 2026-10-06
---

# Phase 1 Plan 14: Compatible Observability and Safe Logging Summary

**Optional OTLP settings preserve external configuration names, while stdout JSON and injected OTel logs share allowlisted operational fields and validated correlation IDs.**

## Performance

- Started: 2026-10-06T14:46:42Z
- Completed: 2026-10-06T14:53:45Z
- Duration: approximately 7 minutes
- Tasks: 2
- Implementation/test files: 4

## Accomplishments

- Added `FLUX_OBSERVABILITY.OTLP.*` enabled/endpoint/sampling/queue/batch/deadline settings and mapping to the existing Settings interface. Preserved dot-nested names and legacy `new_relic` decoding. Non-default legacy credentials/settings cause a constant deprecation warning without printing values or initializing vendor clients.
- Validation returns typed wrapped `ConfigError` values with safe operation-stage diagnostics. Enabled endpoints reject invalid scheme/authority, user credentials, query tokens, fragments and invalid ports; queues, batching, sampling, environment and export deadlines use the provider bounds. Disabled export ignores stale endpoint values; blank endpoints remain valid and introduce no readiness dependency.
- Replaced direct vendor setup, global zerolog formatting/stack mutation and raw SQL console printing with one synchronized sanitizing writer. Both stdout and injected OTel emission use `observability.SanitizeAttributes`. Unknown messages become a bounded operation name, unknown errors become `operation failed`, and arbitrary identity/token/DSN/SQL/stack fields disappear.
- Context hooks preserve validated UUID request/correlation metadata despite conflicting event fields. Standard nonzero lowercase trace/span pairs become stdout fields and OTel record context, reconstructed without tracestate or baggage. Deterministic attribute ordering prioritizes correlation within the 16-attribute bound.
- Preserved temporary `LoggerService`, `GetApplication`, `Shutdown` and `WithTraceContext` compatibility signatures so adjacent consumers compile. These adapters do not create or own a New Relic application. New telemetry ownership and later instrumented consumers remain assigned to plans 01-15 through 01-17.

## Task Commits

1. **Task 1 RED: Specify owned behavior** — `5828ba4` (`test(01-14): specify compatible telemetry settings and safe correlated logs`).
2. **Task 2 GREEN: Implement settings and safe sinks** — `b4d5ea7` (`feat(01-14): configure optional telemetry and sanitize correlated log sinks`).

RED preceded GREEN. No tracked files were deleted.

## Verification

- RED failed on the absent planned OTLP settings, mapping and logger/context interfaces before implementation.
- `go test -race ./internal/config ./internal/logger ./internal/observability -count=1` passed. Actual captured stdout JSON and records exported through the injected OTel SDK prove shared trace/span/UUID identity and rejected secret markers, synthetic PII, credentials, SQL and stack fields. Tests also capture the actual pgx compatibility constructor and exercise concurrent child-context loggers.
- `go test -race ./... -count=1` passed: app 40.093s, database 6.309s, handler 4.249s, logger 1.021s, observability 1.074s. Existing process, container-backed persistence, transport and lifecycle regressions passed.
- `go vet ./internal/config ./internal/logger`, `go build ./cmd/...`, `bun run build` and `git diff --check` passed. Bun replayed two successful cached workspace builds.
- Checked the official OTel documentation through Context7 and the exact installed zerolog context API through `go doc`; used pinned modules already installed by plan 01-13. No dependency changes were necessary.
- Stub and threat-surface scans found no goal-blocking placeholders or undeclared surfaces. Inert vendor compatibility methods are intentional until plan 01-17; optional nil OTel sinks are the specified disabled-export behavior. No application endpoint, schema, auth path or new transport implementation was added.

## Decisions Made

Keep the existing New Relic type signatures temporarily while removing all direct vendor initialization in this slice. Composition can inject an owned OTel log API through `LoggerService.LogSink` or `NewLogger`, without setting global SDK providers. Existing console configuration still binds, with consistently safe JSON output.

## Deviations from Plan

None - plan executed within the specified ownership and compatibility contract.

## Issues Encountered

The installed zerolog API attaches hook-visible context with `Context.Ctx`, while `Logger.WithContext` returns a Go context. Corrected the call and verified actual exporter correlation before GREEN. No unresolved issue remains.

## User Setup Required

None for this stage-local configuration/logging boundary.

## Next Phase Readiness

Ready for plan 01-15 HTTP instrumentation. Database/job instrumentation belongs to 01-16, role/provider ownership and vendor retirement to 01-17, and the complete collector transport proof to 01-18. This plan proves configuration, stdout and injected log-export boundaries without claiming those later integrations.

## Self-Check: PASSED

- All four implementation/test files and this SUMMARY exist.
- RED commit `5828ba4` and GREEN commit `b4d5ea7` exist and contain the owned task files.
- No goal-blocking stubs, unexpected deletions or untracked runtime artifacts remain.
