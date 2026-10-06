# Phase 1 Multi-Source Coverage Audit

The goal is: “Operators and developers have a boring, reliable foundation with final process and architecture boundaries for every later product slice.” This revision retains **22 plans, 44 tasks and 18 waves**. It stabilizes the existing operator/runtime/contract foundation without adding workspace/link/product schema, redirect resolution, click transport, attribution, analytics, billing or product UI from later phases. Locked D-01 through D-19 remain implemented; deferred product ideas are excluded.

| Source | ID | Feature or constraint | Plans | Status | Notes |
|--------|----|-----------------------|-------|--------|-------|
| GOAL | Phase 1 | Operators and developers have a reliable foundation with final process/architecture boundaries for later slices | 01-01–01-22 | COVERED | 22 plans, 44 tasks, 18 waves; no later product-domain behavior |
| REQ | PLAT-01 | Independent API/redirector/worker/migrator roles; no implicit API workers/migrations | 01-04, 01-09, 01-10 | COVERED |  |
| REQ | PLAT-02 | Explicit deterministic migration and empty-database/re-run proof | 01-03, 01-04 | COVERED |  |
| REQ | PLAT-03 | Separate role live/ready with sanitized diagnostics | 01-11 | COVERED |  |
| REQ | PLAT-04 | Bounded SIGTERM drain and complete cleanup | 01-03, 01-09, 01-12, 01-18 | COVERED |  |
| REQ | PLAT-05 | Exact matching local/CI toolchain and service dependencies | 01-01, 01-02, 01-19, 01-20, 01-22 | COVERED |  |
| REQ | PLAT-06 | Failing format/lint/types/generation/migration/test/build CI gates | 01-07, 01-08, 01-19, 01-20, 01-21, 01-22 | COVERED |  |
| REQ | PLAT-07 | Correlated safe HTTP/async logs/traces and bounded metrics | 01-13, 01-14, 01-15, 01-16, 01-17, 01-18 | COVERED |  |
| REQ | PLAT-08 | Versioned TS/Zod authority with generated Go/OpenAPI and drift detection | 01-05, 01-06, 01-07, 01-08 | COVERED |  |
| REQ | SAFE-01 | Redaction across logs/traces/metrics/errors/generated artifacts | 01-11, 01-13, 01-14, 01-15, 01-16, 01-17, 01-18 | COVERED |  |
| REQ | SAFE-08 | Actionable redacted dependency/secret CI scans | 01-21, 01-22 | COVERED |  |
| CONTEXT | D-01 | Thin role roots with explicit resource ownership | 01-09, 01-10, 01-17 | COVERED |  |
| CONTEXT | D-02 | No implicit API migrations/consumers; explicit operator migrator | 01-04, 01-09, 01-10 | COVERED |  |
| CONTEXT | D-03 | Preserve Echo/sound packages and narrow incremental extraction | 01-02, 01-03, 01-09, 01-15–01-17 | COVERED |  |
| CONTEXT | D-04 | Separate Asynq worker startup from HTTP without later event-stream work | 01-03, 01-09, 01-10 | COVERED |  |
| CONTEXT | D-05 | Every long-running role exposes live/ready, including worker management listener | 01-09–01-11 | COVERED |  |
| CONTEXT | D-06 | Coarse public health and redacted detailed operational diagnostics | 01-11, 01-14–01-18 | COVERED |  |
| CONTEXT | D-07 | SIGINT/SIGTERM remove readiness, stop intake, bounded drain, joined cleanup | 01-03, 01-12, 01-17, 01-18 | COVERED |  |
| CONTEXT | D-08 | Partial-startup unwind and optional integration independence | 01-03, 01-09, 01-12, 01-17, 01-18 | COVERED |  |
| CONTEXT | D-09 | TS/Zod to canonical OpenAPI to generated Go boundary | 01-05, 01-06, 01-08, 01-11 | COVERED |  |
| CONTEXT | D-10 | Awaited fail-closed generation, all statuses and stale served artifact rejection | 01-05, 01-06, 01-08, 01-11, 01-20, 01-22 | COVERED |  |
| CONTEXT | D-11 | Embedded documentation/email; alternate-directory regression | 01-07, 01-08, 01-22 | COVERED |  |
| CONTEXT | D-12 | Official verified supported exact toolchain, services and tools; Bun lock authority | 01-01, 01-06, 01-13, 01-18–01-22 | COVERED |  |
| CONTEXT | D-13 | Root checks actually invoke Go and every implemented TS workspace | 01-20, 01-22 | COVERED |  |
| CONTEXT | D-14 | Specified behavior regressions without blanket coverage percentage | 01-02–01-18, 01-20–01-22 | COVERED |  |
| CONTEXT | D-15 | Real pinned PostgreSQL/Redis; narrow failure injection only | 01-01, 01-03, 01-04, 01-10–01-12, 01-16, 01-18, 01-20, 01-22 | COVERED |  |
| CONTEXT | D-16 | Injected OTel/OTLP with correlated safe logs and collector controls | 01-13–01-18 | COVERED |  |
| CONTEXT | D-17 | Retire direct vendor SDK hooks; forwarding behind collector | 01-14–01-18 | COVERED |  |
| CONTEXT | D-18 | Typed wrapped configuration errors; roots choose exit behavior | 01-03, 01-04, 01-09, 01-10, 01-14, 01-17 | COVERED |  |
| CONTEXT | D-19 | Compatible external keys, explicit migration notes and secret-free diagnostics | 01-03, 01-04, 01-09, 01-14–01-18, 01-22 | COVERED |  |
| RESEARCH | R-01 | Coupled server decomposition and explicit role dependency graphs | 01-09, 01-10 | COVERED |  |
| RESEARCH | R-02 | Reverse cleanup registration, joined/idempotent close and pool ping-failure unwind | 01-03, 01-09, 01-12, 01-17 | COVERED |  |
| RESEARCH | R-03 | Remove package-global job email dependency | 01-03 | COVERED |  |
| RESEARCH | R-04 | Worker management health isolated from public HTTP API | 01-09–01-11 | COVERED |  |
| RESEARCH | R-05 | Per-request allocation and safe typed validation fallback | 01-02 | COVERED |  |
| RESEARCH | R-06 | Bounded concurrent readiness/cancellation; optional dependency policy | 01-11, 01-18 | COVERED |  |
| RESEARCH | R-07 | Every actual health response status and generated transport equality | 01-05, 01-06, 01-11 | COVERED |  |
| RESEARCH | R-08 | Fix runtime contracts export and preserve NodeNext .js imports | 01-05 | COVERED |  |
| RESEARCH | R-09 | Awaited atomic deterministic generation; Git-independent drift checks | 01-05, 01-06, 01-08 | COVERED |  |
| RESEARCH | R-10 | Embed docs/email; alternate-directory behavior tests | 01-07, 01-12, 01-22 | COVERED |  |
| RESEARCH | R-11 | One frozen Bun lock after clean install/build comparison | 01-01 | COVERED |  |
| RESEARCH | R-12 | Actual root Go/TS formatting/lint/types/tests/build/generation/migration commands | 01-19, 01-20, 01-22 | COVERED |  |
| RESEARCH | R-13 | Pinned real PostgreSQL/Redis; clean migration and repeat no-op | 01-01, 01-04, 01-10–01-12, 01-16, 01-18, 01-20, 01-22 | COVERED |  |
| RESEARCH | R-14 | Active HTTP/job SIGTERM subprocess drain and startup-failure injection | 01-03, 01-09, 01-12, 01-17 | COVERED |  |
| RESEARCH | R-15 | Vendor-neutral role providers, OTLP and queue/retry correlation | 01-13–01-18 | COVERED |  |
| RESEARCH | R-16 | No URI/IP/identity/email/SQL/provider-error telemetry; bounded labels | 01-13–01-18 | COVERED |  |
| RESEARCH | R-17 | Optional bounded export, complete cleanup and collector redaction | 01-13, 01-17, 01-18 | COVERED |  |
| RESEARCH | R-18 | Least-privilege pinned GitHub Actions, timeouts and lock-keyed caches | 01-22 | COVERED |  |
| RESEARCH | R-19 | Go/JS vulnerability analysis and redacted worktree/history scans | 01-19, 01-21, 01-22 | COVERED |  |
| RESEARCH | R-20 | Official module/release verification and package legitimacy before pinning | 01-01, 01-06, 01-13, 01-18, 01-19, 01-21, 01-22 | COVERED |  |
| RESEARCH | R-21 | No new ORM/DI/microservices/dual schema authority or later product domains | 01-01–01-22 | COVERED |  |
| RESEARCH | R-22 | Read-only/uninitialized Git must not disable generation/tests/scanning | 01-08, 01-21, 01-22 | COVERED |  |

## Dependency Graph and Ownership

| Plan | Needs | Creates | Wave | Owned files | Checkpoint |
|------|-------|---------|------|-------------|------------|
| 01-01 | Existing scaffold | Pin toolchain and real dependency services | 1 | 8 | No |
| 01-02 | 01-01 | Isolate requests and validate safely | 2 | 5 | No |
| 01-03 | 01-01 | Return configuration errors and own cleanup | 2 | 9 | No |
| 01-04 | 01-03 | Run explicit deterministic migrations | 3 | 6 | No |
| 01-05 | 01-01 | Author deterministic canonical health contracts | 2 | 8 | No |
| 01-06 | 01-05 | Generate checked Go health transport | 3 | 4 | No |
| 01-07 | 01-05, 01-06 | Embed documentation and email assets | 4 | 10 | No |
| 01-08 | 01-06, 01-07 | Reject stale generated artifacts | 5 | 2 | No |
| 01-09 | 01-04, 01-06, 01-07 | Define role configuration and resource ownership | 5 | 7 | No |
| 01-10 | 01-04, 01-09 | Wire independent thin role binaries | 6 | 6 | No |
| 01-11 | 01-06, 01-07, 01-09, 01-10 | Expose role-specific live and ready | 7 | 7 | No |
| 01-12 | 01-09, 01-11 | Drain on SIGTERM and close every resource | 8 | 6 | No |
| 01-13 | 01-12 | Define safe telemetry and propagation providers | 9 | 8 | No |
| 01-14 | 01-01, 01-06, 01-09, 01-13 | Configure compatible observability and safe logging | 10 | 4 | No |
| 01-15 | 01-02, 01-11, 01-14 | Instrument safe HTTP correlation | 11 | 8 | No |
| 01-16 | 01-03, 01-15 | Propagate safe database and queued-job context | 12 | 6 | No |
| 01-17 | 01-04, 01-12, 01-14, 01-16 | Wire role telemetry and retire vendor hooks | 13 | 9 | No |
| 01-18 | 01-01, 01-08, 01-14, 01-17 | Prove collector redaction and outage independence | 14 | 5 | No |
| 01-19 | 01-18 | Pin and install reproducible quality tooling | 15 | 5 | No |
| 01-20 | 01-01, 01-05, 01-08, 01-10, 01-19 | Run complete backend and workspace quality gates | 16 | 10 | No |
| 01-21 | 01-19, 01-20 | Fail safely on dependency and secret findings | 17 | 5 | No |
| 01-22 | 01-21 | Encode full least-privilege CI and runbook | 18 | 2 | No |

Same-wave plans have disjoint ownership; every shared file is ordered by a transitive dependency. The existing graph is preserved. Plan 01-13 now owns official OTel module bootstrap plus go.mod/go.sum and tools.lock.json before provider tests; 01-14 owns compatible configuration and sanitized logger integration. 01-17 removes vendor modules after every consumer has migrated. tools.lock.json progresses through 01-06 → 01-13 → 01-18 → 01-19 → 01-21. Newly owned database/job tests in 01-03 are scoped to ping-failure/stop cleanup; database telemetry tests extend the already-ordered database test file in 01-16.

Stage completion is local: 01-05 proves TS/Zod/OpenAPI and runtime exports; 01-06 proves generated Go. 01-09 proves constructor graphs; 01-10 proves binaries; 01-11/12 prove health and active lifecycle. 01-13 uses injected provider/exporter capture; 01-14 captures sanitized sinks; 01-15 captures HTTP instrumentation; 01-16 uses real Redis with injected exporters; 01-17 verifies role provider ownership and parsed forbidden-vendor imports/modules. Only 01-18 requires the actual HTTP → real Redis enqueue/execution/retry → OTLP protobuf → pinned collector output test, including Span.trace_state, every Link.trace_state, legacy queue ingress, baggage and all sensitive signal surfaces. 01-19 validates installer/tools/configs; 01-20 executes check:fast and tests complete quality/scanner dispatch with injected runners; 01-21 executes real scans; 01-22 executes the full clean-environment gate.

Plans 01-07 and 01-20 retain ten-file unions as two explicit five-file tasks each. Documentation/email and workspace scripts/root dispatch are bounded existing-pattern seams, with expected combined context use at 40–50%; no task owns the ten-file union. All other plans own at most nine files and all plans have two tasks.

No source item is missing and no external account setup/human-only action is required. Exact supported patch versions are resolved against official sources during implementation with fail-closed availability/integrity/compatibility checks; research candidates are not new version verification claims. The workspace Git metadata is read-only and not initialized, so commit-based completion is unavailable until metadata is restored externally. This environment limitation does not waive tests, regeneration, worktree scanning or local full-gate execution.
