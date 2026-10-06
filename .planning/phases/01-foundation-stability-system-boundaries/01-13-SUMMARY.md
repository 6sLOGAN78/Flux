---
phase: 01-foundation-stability-system-boundaries
plan: "13"
subsystem: infra
tags: [go, opentelemetry, otlp, redaction, correlation, privacy]
requires:
  - phase: 01-12
    provides: Bounded role lifecycle and cleanup ownership
provides:
  - Config-independent injected OTel trace, metric and log providers with bounded asynchronous exporters
  - Value-aware operational allowlist and safe provider error classification
  - Traceparent-only propagation and UUID request/correlation metadata
affects: [01-14, 01-15, 01-16, 01-17, 01-18]
tech-stack:
  added: [OpenTelemetry Go v1.47.0, OTLP HTTP log exporter v0.23.0, OTLP protobuf v1.11.0]
  patterns: [fixed instrumentation scope, pre-export sanitization, bounded independent provider shutdown]
key-files:
  created: [apps/backend/internal/observability/telemetry.go, apps/backend/internal/observability/telemetry_test.go, apps/backend/internal/observability/redaction.go, apps/backend/internal/observability/redaction_test.go, apps/backend/internal/observability/correlation.go]
  modified: [apps/backend/go.mod, apps/backend/go.sum, tools.lock.json]
key-decisions:
  - Pin the official v1.47.0 stable API/SDK and trace/metric HTTP family with its v0.23.0 log HTTP exporter, retaining existing New Relic modules through plan 01-17.
  - Expose fixed-scope API objects and sanitize every export surface with closed operational values, suppress metric correlation labels and all exemplars, and bound all independent provider shutdowns.
patterns-established:
  - observability.New(ctx, Settings, role) owns Tracer, Meter, Logger, Propagator and Shutdown without SDK globals or config imports.
  - Only accepted traceparent identity and sampled flag survive propagation; SpanContext and Links are reconstructed with empty TraceState.
requirements-completed: [PLAT-07, SAFE-01]
duration: 13min
completed: 2026-10-06
---

# Phase 1 Plan 13: Safe Telemetry Providers Summary

**Injected OTel traces, metrics and logs export only approved operational fields, preserve safe trace correlation, and flush through bounded asynchronous providers.**

## Performance

- Started: 2026-10-06T14:31:00Z
- Completed: 2026-10-06T14:44:00Z
- Duration: approximately 13 minutes
- Tasks: 2
- Implementation and dependency files: 8

## Accomplishments

- Verified the official [v1.47.0 release](https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.47.0), its [module family](https://github.com/open-telemetry/opentelemetry-go/blob/v1.47.0/versions.yaml), [Go exporter documentation](https://opentelemetry.io/docs/languages/go/exporters/) and [protobuf origin](https://github.com/open-telemetry/opentelemetry-proto-go/releases/tag/v1.11.0). All selected exact paths were checked with `go list -m -versions`; module sums, module-file sums and source origins are recorded in `tools.lock.json`. A temporary API/exporter probe compiled before provider tests were authored; it was removed afterward.
- Added config-independent `Settings`, injected `Exporters` and owned `Telemetry`. Fixed-scope tracer/meter/logger APIs avoid SDK global registration. Role resources contain only `flux.<role>`, a closed role and a validated environment. Empty endpoint/exporter configuration produces optional no-op instrumentation. Explicit HTTP endpoints select all three official OTLP exporters with export deadlines and retries disabled.
- Trace/log queues, batches, attribute count/length, span events/links, periodic metrics batches and metric cardinality are bounded. Root sampling is explicit. Metric labels exclude request/correlation IDs, exemplars are disabled and scrubbed defensively, and uncontrolled instrument names are discarded at export.
- Value-aware allowlists remove arbitrary attributes, raw SQL/arguments, network identifiers, destinations, credential/body/identity fields and opaque error strings. Span names, event names, status descriptions, scope metadata, metric descriptions/units, log bodies/event names/severity text and resources are covered alongside attributes. Unknown errors produce the constant `operation failed`; wrapped causes remain available internally through `errors.Is`.
- Extraction passes only `traceparent` to OTel and clears baggage plus the prior span context. Serialization removes stale case-insensitive traceparent/tracestate/baggage entries, then writes only a validated clean traceparent. Every exported span, parent and Link context is reconstructed with empty TraceState. UUID request/correlation metadata is validated or replaced.
- Shutdown attempts every independent provider, safely joins failures and remains bounded by the caller deadline and an internal export-timeout budget. Repeated calls observe the original shutdown rather than closing providers twice.

## Task Commits

1. **Task 1: Bootstrap official compatible OTel modules** — `cdd415d` (`chore(01-13): pin verified compatible OpenTelemetry module family`).
2. **Task 2 RED: Specify injected providers and propagation** — `8d50fe7` (`test(01-13): specify safe telemetry export and propagation boundaries`).
3. **Task 2 GREEN: Implement and prove providers** — `af3beb6` (`feat(01-13): add bounded private telemetry providers and safe propagation`).

RED preceded GREEN. No tracked files were deleted.

## Verification

- `go mod verify`, selected SDK/exporter package builds, compiled exporter/API probe and `go list -m all` passed.
- RED failed on the absent planned provider/redaction/propagation interfaces before implementation.
- `go test -race ./internal/observability -count=1` passed; final expanded run took 1.082 seconds. In-memory exporters prove role resources, allowed operational fields, rejected secret markers on all export surfaces, safe trace/parent/Link state and log trace/span correlation. Additional tests prove metric exemplar/resource/scope sanitization, invalid metadata/settings rejection, explicit root sampling, optional providers and bounded queues under a blocked exporter.
- Actual trace/metric/log SDK provider shutdown tests prove every exporter is closed and all private causes remain discoverable without exposing their messages. Deterministic delayed shutdown tests prove aggregate completed failures and caller/internal deadlines, including an uncooperative closer and idempotent repeat calls.
- `go test -race ./... -count=1` passed after completing module graph checksums: app 37.757s, database 13.090s, handler 5.340s, observability 1.065s. Existing real PostgreSQL/Redis and lifecycle suites continue to pass with the selected module family.
- `go vet ./internal/observability`, `go build ./cmd/...`, `bun run build` and `git diff --check` passed. Bun replayed two successful cached workspace builds.
- Stub and threat-surface scans found no goal-blocking placeholders or undeclared production surfaces. Remote OTLP transport is the planned exporter trust boundary; no application endpoint, authentication path or database schema was added.

## Decisions Made

Use the official stable v1.47.0 logs API/SDK with the matching experimental v0.23.0 HTTP log exporter. Existing New Relic requirements remain pinned for their later planned retirement. The allowlist includes closed route/operation/metric vocabularies: future instrumentation must explicitly extend policy when introducing new operational names.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Completed the upgraded module graph and import classification**
- Found during: Task 2 overall verification.
- Issue: Minimal module bootstrap compiled the selected OTel APIs but the existing backend required additional checksum entries and direct-import bookkeeping after MVS upgrades to mapstructure, testify and otelhttp.
- Fix: Downloaded the resolved graph, ran `go mod tidy`, and retained the existing `nrredis-v9` requirement/checksums explicitly despite tidy treating it as unused. No alternative package or dependency name was substituted.
- Files: `apps/backend/go.mod`, `apps/backend/go.sum`.
- Verification: `go mod verify`, full backend race suite and all command builds passed.
- Commit: `af3beb6`.

## Issues Encountered

The installed HeaderCarrier API exposes carrier methods rather than Header.Del; propagation uses case-insensitive map-key removal so stale metadata is removed consistently from both standard carrier implementations.

## User Setup Required

None for this local provider boundary.

## Next Phase Readiness

Plan 01-14 maps configuration to Settings and wires stdout logs; 01-15 instruments HTTP; 01-16 instruments database/jobs; 01-17 injects role ownership and retires vendor consumers. The actual HTTP → Redis retry → OTLP protobuf → collector proof belongs solely to 01-18. This summary claims local exporter boundaries and regression compatibility, not that later integration is complete. No blocker remains for plan 01-14.

## Self-Check: PASSED

All five created observability source/test files and dependency evidence files exist. Commits `cdd415d`, `8d50fe7` and `af3beb6` exist in Git history. RED and GREEN gate commits are in order.
