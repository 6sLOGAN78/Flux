---
phase: 01-foundation-stability-system-boundaries
plan: "18"
subsystem: infra
tags: [opentelemetry, collector, protobuf, privacy, redis, integration]
requires:
  - phase: 01-01
    provides: Pinned local PostgreSQL and Redis environment
  - phase: 01-08
    provides: Canonical generated OpenAPI and email artifacts
  - phase: 01-14
    provides: Compatible optional OTLP configuration and sanitized stdout
  - phase: 01-17
    provides: Role-owned telemetry with provider-last bounded cleanup
provides:
  - Digest-pinned opt-in loopback collector with three value-aware redaction pipelines
  - Real HTTP-to-Redis execution/retry proof with actual pre/post collector OTLP protobuf captures
  - All-role collector outage verification and final 200-request single-series cardinality proof
affects: [01-19, 01-20, 01-21, 01-22]
tech-stack:
  added: [OpenTelemetry Collector Contrib 0.162.0]
  patterns: [independent dirty signal probes, decoded protobuf lineage, exact binary configuration validation]
key-files:
  created: [deploy/otel-collector.yaml, docs/observability.md, apps/backend/internal/app/observability_test.go]
  modified: [compose.yaml, tools.lock.json]
key-decisions:
  - Preserve sanitized application links before the collector; drop complete link and event collections at the collector because this pinned OTTL release has no individual Link context.
  - Keep the collector optional and vendor routing outside application processes; assert final cumulative metrics rather than accepting an early partial batch.
requirements-completed: ["PLAT-07", "SAFE-01", "PLAT-04"]
duration: 11min
completed: 2026-10-07
---

# Phase 1 Plan 18: Collector Privacy and Outage Independence Summary

**Digest-pinned collector redaction preserves HTTP/Redis retry trace lineage, strips private fields from all three OTLP signals, and leaves required operations independent of monitoring outages.**

## Performance

- Recovery session resumed: 2026-10-07T15:28:35Z
- GREEN completed: 2026-10-07T15:39:28Z
- Recovery duration: approximately 11 minutes; earlier interrupted work is excluded from this active-time metric.
- Tasks: 2, including the existing RED task.
- Implementation/test/configuration files: 5.

## Accomplishments

- Recovered the interrupted executor's five-file implementation without recreating RED or reverting prior work. Added stronger independent collector probes for secret values under allowed attribute keys, invalid scalar types, private metric exemplars, preservation of valid UUIDs and numeric points, and explicit nonzero Link span IDs.
- Compose's optional observability profile pins official contrib 0.162.0 at manifest digest `sha256:39923a8e431bd1f57be82411999d389fcfe40857492e4365456d97a4c1f74be6`. Both host receiver ports are loopback-only, configuration mounts read-only, and memory is capped. All three pipelines apply memory limiter, fail-closed value-aware transform, and batching before the local debug exporter. No vendor credentials are embedded.
- `TestTracestateHTTPRedisOTLP` sends actual HTTP through the existing middleware and real PostgreSQL/Redis role graphs, runs a real Asynq worker against a local fake email HTTP transport, inspects persisted initial/retry/completed payloads, and verifies four execution/retry spans. Legacy private metadata/headers are replaced through public enqueue/RevokeTask APIs. The test checks HTTP/job baggage stripping and explicitly decodes every protobuf Span and Link tracestate before and after the actual collector, requiring empty values and preserved trace IDs, sampled flags, parents and UUID lineage. Application captures require nonempty valid links; collector captures require links removed.
- Marker scans cover safe buffered stdout, complete OTLP wire payloads for logs/traces/metrics, collector debug output, queued/retried metadata, safe provider failures, and existing generated OpenAPI/email artifacts. Failure messages do not print synthetic private values. A separate deliberately dirty three-signal protobuf probe verifies collector defenses without relying on the application's prior sanitization.
- Disconnected and slow collector tests retain API/worker/redirector required-dependency readiness, bound shutdown, attempt every owned infrastructure/provider closer, and release listeners. The actual migrator subprocess still commits the PostgreSQL migration under both outages and exits within its bounded budget; final telemetry failure can return exit 1 after migration success. Narrow fault exporters prove deadlines and all three provider shutdown attempts.
- The final cardinality capture requires all 200 HTTP requests in the cumulative counter, exactly one fixed series, `unmatched` route, and no request/correlation UUID labels. The runbook explains endpoint/configuration names, defaults and bounds, disabled export, deprecation compatibility, queue replacement semantics, production collector routing and the link/event tradeoff.

## Task Commits

1. **Task 1 RED: Specify actual collector privacy and outage behavior** — `9e506bc` (`test(01-18): specify real OTLP collector privacy and outage proof`). Existing interrupted-session commit retained.
2. **Task 2 GREEN: Implement and verify collector redaction and outage independence** — `ea8df88` (`feat(01-18): prove redacted collector export and outage independence`).

RED precedes GREEN. No tracked files were deleted.

## Verification

Fresh recovery-session checks, with `/usr/local/go/bin` on PATH and Go 1.26.8:

- Final required scoped selector passed: `go test ./internal/app -run 'Test(Collector|Observability|TelemetryOutage|TelemetryLeak|TracestateHTTPRedisOTLP|Cardinality)' -count=1` — app 14.258s.
- Final scoped race selector passed: `go test -race ./internal/app -run 'Test(Collector|Observability|TelemetryOutage|TelemetryLeak|TracestateHTTPRedisOTLP|Cardinality)' -count=1` — app 16.524s.
- Expanded independent dirty probe also passed separately with `-count=1` — app 1.452s.
- `go vet ./internal/app ./internal/observability ./internal/middleware ./internal/lib/job ./cmd/migrator` passed after final Go edits.
- Cached Docker RepoDigest matches the prior officially resolved immutable manifest. Running that image's `--version` reports `otelcol-contrib version 0.162.0`; its `validate` command accepts the checked-in configuration. Each collector integration also validates its capture-exporter configuration against the same exact binary before starting it.
- `docker compose --profile observability config --quiet`, changed-file Go formatting, and `git diff --check` passed. Context7 official transform documentation confirmed value-aware `keep_keys`/context access and nil collection clearing; exact runtime validation supplies version-specific syntax proof.
- No goal-blocking stubs or untracked runtime artifacts were found. The collector network boundary is already declared in the plan threat model; no additional application endpoint, authentication path or schema change was introduced.

The parent orchestrator owns the later cumulative backend/workspace gate. This summary does not claim a fresh full-backend run.

## Decisions Made

Use the already verified official release/digest, then validate its actual binary rather than substituting a similarly named image. Keep telemetry optional and local debug export strictly after redaction.

OTTL 0.162.0 cannot individually sanitize Link context. Remove complete link/event collections at the collector, while requiring safe nonzero links in the decoded application export. Safe parent lineage and trace/correlation IDs remain available downstream; links/events intentionally do not.

## Deviations from Plan

**1. [Rule 2 - Missing critical functionality] Clear complete collector link/event collections.** The pinned transform release has no per-Link context, so attribute-only removal could leave `Link.trace_state` exposed. `deploy/otel-collector.yaml` clears links/events, and the tests separately require valid safe links before collection. The runbook documents this privacy tradeoff. Exact binary validation and independent dirty protobuf assertions passed. Commit: `ea8df88`.

All implementation changes remain within the five-file ownership union. Stronger probe assertions and final cumulative cardinality enforcement fulfill the planned behavior without new production prerequisites.

## Issues Encountered

The inherited shell PATH omitted Go; using the existing `/usr/local/go/bin` resolved this without installing anything. During probe strengthening, an initial fixture accidentally duplicated an attribute key, violating OTLP's unique-key contract; it failed the private-value scan. Replaced the existing route value to make the probe valid, then reran the independent probe and complete scoped selectors successfully. No unresolved execution blocker remains.

## User Setup Required

None required. The collector profile and endpoint configuration are optional and documented in `docs/observability.md`.

## Next Phase Readiness

Ready for 01-19 and the parent orchestrator's cumulative gates. Collector privacy, real asynchronous lineage, outage independence and cardinality checks are executable without production vendor credentials.

## Self-Check: PASSED

All five owned implementation files and this summary exist. RED `9e506bc` and GREEN `ea8df88` exist in order. Fresh scoped regular/race tests, relevant vet and exact collector/Compose validation passed; neither task commit deleted tracked files.
