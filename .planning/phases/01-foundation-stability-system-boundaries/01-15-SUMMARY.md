---
phase: 01-foundation-stability-system-boundaries
plan: "15"
subsystem: api
tags: [go, echo, opentelemetry, correlation, privacy, metrics]
requires:
  - phase: 01-13
    provides: Injected telemetry APIs, closed attribute policies and traceparent-only propagation
  - phase: 01-14
    provides: Sanitizing stdout writer and context-correlated log APIs
provides:
  - Injected HTTP server spans with sanitized parent and Link contexts
  - Bounded UUID request and correlation response/log/span identity
  - Closed route, method, role and status-class HTTP metric labels
  - Private request, handler, readiness, error and panic telemetry
affects: [01-16, 01-17, 01-18]
tech-stack:
  added: []
  patterns: [injected HTTP API instrumentation, ingress metadata replacement, fixed operational labels]
key-files:
  created: [apps/backend/internal/middleware/tracing_test.go]
  modified: [apps/backend/internal/middleware/tracing.go, apps/backend/internal/middleware/context.go, apps/backend/internal/middleware/request_id.go, apps/backend/internal/middleware/global.go, apps/backend/internal/handler/base.go, apps/backend/internal/handler/health.go, apps/backend/internal/router/router.go, apps/backend/internal/server/server.go]
key-decisions:
  - Use the existing closed HTTP route policy after Echo routing; unknown routes collapse to unmatched and metric status codes represent their hundred-class.
  - Retain the inert legacy tracing constructor argument until plan 01-17, while the router explicitly injects Server.Telemetry; role provider construction remains deferred to its owned plan.
patterns-established:
  - RequestID validates canonical nonzero UUID headers with a 36-byte precheck, rejects multiple values, replaces input and attaches context plus response headers.
  - HTTP tracing strips all incoming tracestate and baggage headers before the traceparent-only propagator clears inherited baggage/span state.
requirements-completed: [PLAT-07, SAFE-01]
duration: 11min
completed: 2026-10-06
---

# Phase 1 Plan 15: Safe HTTP Correlation Summary

**Echo requests emit injected server spans, fixed-label metrics and private operational logs with shared validated UUID and trace identity.**

## Performance

- Started: 2026-10-06T14:56:59Z
- Completed: 2026-10-06T15:07:00Z
- Duration: approximately 11 minutes
- Tasks: 2
- Implementation/test files: 9, including one minimal prerequisite seam

## Accomplishments

- Replaced owned HTTP/Echo/handler vendor hooks with explicitly injected OTel tracer/meter APIs. The router instruments early CORS and rate-limit responses within request identity, context logging and safe recovery. Provider creation and lifecycle ownership remain assigned to plan 01-17.
- Validated request and correlation headers as canonical, nonzero UUIDs. Missing, invalid, oversized, noncanonical and multiple-value inputs are replaced before the request proceeds; request headers, response headers, context, spans and logs use the validated identity. Invalid correlation falls back to the accepted/generated request UUID.
- Removed all incoming tracestate and baggage headers before extraction. The traceparent-only propagator also clears inherited baggage and span context, accepts valid standard parent IDs/flags, and rejects invalid incoming parents. Constructor-supplied Links are bounded to 16, reconstructed with empty TraceState and sanitized attributes before span creation.
- Recorded one `http.request` server span and the `flux.http.requests` counter plus `flux.http.duration` histogram per request. Closed method/route/role values and status classes bound metric series. Matched Echo route templates are read after routing and checked against the shared route policy; unknown/unapproved routes collapse to `unmatched`. IDs never become metric labels.
- Removed full URI/query, IP, user agent, identity, arbitrary header, filename and opaque provider/driver error fields from HTTP telemetry. Recovery suppresses Echo's default stack output and replaces panic text with a constant. Known application validation errors preserve their public envelopes and field errors; unrecognized Echo messages use standard HTTP status text.
- Preserved fresh request factories and existing handler isolation tests. Readiness records allowlisted dependency-check events and safe classifications while retaining its established provider-free checks and timeout behavior.

## Task Commits

1. **Task 1 RED: Specify and prove owned behavior** — `928cd28` (`test(01-15): specify safe injected HTTP telemetry and correlation`).
2. **Task 2 GREEN: Implement safe HTTP instrumentation** — `2617afd` (`feat(01-15): instrument private HTTP spans metrics and correlated logs`).

RED preceded GREEN. No tracked files were deleted.

## Verification

- RED failed on the absent injected telemetry/Link constructor before implementation; the existing handler suite passed.
- Final `go test -race ./internal/middleware ./internal/handler -count=1` passed: middleware 1.085s, handler 4.772s. Real Echo requests use an actual synchronous SDK span exporter, manual metric reader and buffered sanitized stdout writer. Assertions directly inspect accepted parent IDs, server span kind, response/log/span identity, empty span/parent/Link TraceState, removed baggage, invalid/duplicate header replacement and invalid-parent rejection.
- Forty varied query strings plus forty distinct unknown paths and three status/error requests produce exactly five counter label sets for 83 requests. Captured spans, metrics and logs contain no supplied credential/PII/propagation/provider-error markers. Provider errors, panics, public validation and unrecognized Echo errors are exercised through actual responses.
- `go test -race ./... -count=1` passed: app 42.123s, database 10.339s, handler 7.148s, middleware 2.215s. Existing process, container-backed persistence, health and lifecycle tests passed. The final additional middleware regressions/public-message tightening were verified by the focused race suite afterward.
- `go vet ./internal/middleware ./internal/handler ./internal/router`, `go build ./cmd/...`, `bun run build` and `git diff --check` passed. Bun replayed two successful cached workspace builds.
- Context7 official OTel documentation and installed Go API documentation confirmed the injected tracer/meter, span Link and metric reader APIs. No packages were installed or changed.
- Stub and threat-surface scans found no goal-blocking placeholders or undeclared network/auth/schema surfaces.

## Decisions Made

Use the existing closed route allowlist and hundred-class response status metric values rather than accepting arbitrary route strings or per-request identities. Preserve the inert legacy constructor call while the router passes a typed Telemetry object; the minimal Server field enables later role composition without introducing provider initialization in this stage.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added the minimal injected server telemetry seam**
- Found during: Task 2 integration.
- Issue: The owned router needs a role-supplied telemetry object, but Server had no field for it. The existing middleware registry still passes its inert legacy vendor adapter until plan 01-17.
- Fix: Added only `Telemetry *observability.Telemetry` and its import to `server.go`; router passes it explicitly. The tracing constructor tolerates the inert old argument temporarily, creates no vendor application and uses injected APIs when supplied. No role provider creation, registry redesign or lifecycle change was made.
- Files: `apps/backend/internal/server/server.go` plus the planned tracing/router files.
- Verification: Full backend race suite, final focused race suite, vet and command build passed.
- Commit: `2617afd`.

## Issues Encountered

None unresolved. Empty/noop telemetry objects are the intended optional-export behavior; the actual role provider injection belongs to plan 01-17.

## User Setup Required

None for this HTTP boundary.

## Next Phase Readiness

Ready for database/job instrumentation in plan 01-16, role/provider composition and vendor retirement in plan 01-17, and actual collector/protobuf transport validation in plan 01-18. This summary proves the local HTTP span/metric/stdout boundary and preserves prior regressions; it does not claim the later end-to-end transport proof.

## Self-Check: PASSED

- Created test and SUMMARY files exist; all eight changed production files exist.
- RED `928cd28` and GREEN `2617afd` commits exist in order.
- No unexpected deletions, goal-blocking stubs or generated/untracked runtime artifacts remain.
