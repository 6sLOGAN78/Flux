---
phase: 01-foundation-stability-system-boundaries
plan: "11"
subsystem: infra
tags: [go, echo, readiness, liveness, openapi, privacy, testcontainers]
requires:
  - phase: 01-06
    provides: Canonical generated health transport and schema authority
  - phase: 01-07
    provides: Embedded local email assets
  - phase: 01-09
    provides: Explicit role resource ownership and readiness deadline configuration
  - phase: 01-10
    provides: Independently runnable API, redirector and worker binaries
provides:
  - Process-only /live and bounded role-specific /ready on all long-running listeners
  - Required-only concurrent checks with safe classified failure logs and coarse generated payloads
  - Actual role HTTP contract equality and worker provider-free readiness regressions
affects: [01-12, role-health, operational-probes, privacy]
tech-stack:
  added: []
  patterns: [explicit injected readiness checks, shared request-derived deadline, allowlisted diagnostic classification]
key-files:
  created: [apps/backend/internal/handler/health_test.go]
  modified: [apps/backend/internal/handler/health.go, apps/backend/internal/router/router.go, apps/backend/internal/router/system.go, apps/backend/internal/app/api.go, apps/backend/internal/app/redirector.go, apps/backend/internal/app/worker.go]
key-decisions:
  - Each role injects its required checks into health registration; API probes PostgreSQL and enabled producer Redis, worker probes Redis and local email adapter configuration/assets, redirector declares no Phase 1 external dependencies.
  - Remove public /status and keep /live and /ready outside the customer limiter; emit safe failure classifications and timing instead of arbitrary driver/provider error strings.
patterns-established:
  - Only the request handler aggregates and logs probe results; concurrent probes share one context and buffered result channel.
  - Compare real role HTTP status/body with generated Go encoding and schemas resolved from the canonical OpenAPI response reference.
requirements-completed: ["PLAT-03", "SAFE-01"]
duration: 7min
completed: 2026-10-06
---

# Phase 1 Plan 11: Role Liveness and Readiness Summary

**API, redirector and worker expose generated, sanitized health responses with process-only liveness and concurrent required-dependency readiness under one request deadline.**

## Performance

- Started: 2026-10-06T14:08:22Z
- Completed: 2026-10-06T14:15:15Z
- Duration: approximately 7 minutes
- Tasks: 2
- Source files changed: 7

## Accomplishments

- Added explicit `ReadinessCheck` declarations injected by role composition, independent of global-container inference. API requires PostgreSQL and, only when producers are enabled, queue Redis. Worker requires queue Redis and local email-adapter configuration plus successful embedded-template rendering; it never sends readiness email or contacts the provider. The Phase 1 redirector declares no external dependencies and returns an empty JSON checks array.
- Registered `/live` and `/ready` on all role listeners, including the worker's existing management-only listener. `/live` always returns generated HTTP 200 `alive` without executing checks. `/ready` runs independent probes concurrently using one deadline derived from the incoming request and configured role timeout, preserves deterministic component order, observes request cancellation, and returns HTTP 200 or 503 with generated `ready|not_ready` states.
- Removed the previous public `/status` route and its dependency topology, timings, environment, timestamp and raw error response fields. Customer rate limiting explicitly skips the two health routes; customer requests still exhaust their quota. Probe failure logs occur once per failed component with allowlisted failure classification, component and duration; arbitrary driver/provider text is omitted from both responses and logs.
- Added regressions for actual listener HTTP responses from API, API with producers, redirector and worker. Every observed health status/body is compared with exact expected state, generated Go JSON encoding and the canonical OpenAPI operation's response schema. Worker remains management-only; no docs/product/status routes are exposed, and a local provider request counter remains zero throughout readiness checks.

## Task Commits

1. **Task 1: Specify and prove the owned behavior** — `a10b5d9` (`test(01-11): specify role liveness and bounded readiness`).
2. **Task 2: Implement expose role-specific live and ready** — `e1fe54f` (`feat(01-11): expose sanitized role liveness and readiness`).

Normal atomic Git commits were used. No tracked files were deleted.

## Verification

- RED: prescribed `go test -race ./internal/handler ./internal/router ./internal/app -run 'Test(Live|Ready|Role)' -count=1` failed on the absent planned `ReadinessCheck`, `NewReadinessHandler` and `RegisterHealthRoutes` interfaces; the existing app role tests passed. RED was committed before implementation.
- GREEN: the same prescribed scoped race suite passed: handler approximately 7.8 seconds, app approximately 23.8 seconds, router compiled with no package-local tests. This preserves existing role-constructor allowlists, startup-failure cleanup, real binary startup and dependency teardown tests.
- `go test -race ./internal/handler -count=1` passed the complete handler package, preserving its existing request-factory and asset regressions.
- `go vet ./internal/handler ./internal/router ./internal/app`, `go build ./cmd/api ./cmd/redirector ./cmd/worker ./cmd/migrator ./cmd/flux`, `go mod verify` and `git diff --check` passed.
- Unit checks prove zero liveness probe calls, one classified failure log without connection-string/provider/stack markers, simultaneous stalled probes sharing the same request deadline/context, request cancellation, and health success after exhausting customer quota.
- Real pinned PostgreSQL 17.11 and Redis 8.10.2 fixtures establish ready baselines. Closing API PostgreSQL proves not-ready behavior; narrow closed-client probe injection independently proves required API/worker Redis failures while preserving the active consumer's ordered resource ownership. Removing worker adapter credentials proves local-adapter not-ready behavior. Liveness stays 200 during dependency failures.
- Consulted Context7 Echo documentation for post-routing middleware and the installed exact Echo v4.15.2 `RateLimiterConfig.Skipper` implementation. No dependencies or generated artifacts changed. Production stub/secret-pattern and threat-surface scans found no goal-blocking placeholders or unplanned trust boundaries; the new health routes are declared by this plan.

## Decisions Made

Keep role-owned readiness declarations separate from handler registry compatibility construction. Probe only resources actually required by the role. Local email rendering validates required assets without provider requests. Safe structured failure classifications preserve useful operational context without exporting unknown error text.

## Deviations from Plan

None — all production/test edits remain within the seven-file ownership union. No prerequisite edits or dependency installs were required.

## TDD Gate Compliance

RED `a10b5d9` precedes GREEN `e1fe54f`. Both tasks completed in sequence; no refactor commit was needed.

## Issues Encountered

Initial local email health rendering used the wrong template key; corrected it to the existing embedded template's `UserFirstName`. Early failure injection closed the worker's active shared Redis client before consumer shutdown and exposed an upstream Asynq subscriber panic; another API test exposed non-idempotent repeated Redis close. The final tests substitute a separate closed probe client and restore the active client before role teardown, preserving real baseline infrastructure and ordered cleanup without production workarounds. All final checks pass. No authentication gates occurred.

## User Setup Required

None for verification. Operators use existing role readiness timeout configuration and `/live` or `/ready` for supervisor probes.

## Next Plan Readiness

Ready for 01-12's unready-first active-work shutdown/drain slice. This plan completes the allocated PLAT-03 and health-diagnostic SAFE-01 behavior; overall privacy and lifecycle phase acceptance remain subject to their adjacent fragments. The dependency-free redirector is intentional Phase 1 scope, not a product stub. No later plan was executed.

## Self-Check: PASSED

All seven owned source files exist. Both task commits exist and contain no tracked deletions. RED precedes GREEN; scoped and full-handler race checks, vet, binary builds, module verification and whitespace checks passed. No generated runtime files remain untracked.
