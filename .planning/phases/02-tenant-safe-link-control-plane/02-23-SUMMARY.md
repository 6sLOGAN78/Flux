---
phase: 02-tenant-safe-link-control-plane
plan: "23"
subsystem: invitations
tags: [go, postgres, redis, asynq, readiness, lifecycle, tdd]
requires:
  - phase: 02-22
    provides: Durable invitation worker, immutable requests and recovery fences
provides:
  - Local open-adapter, encryption-key and two-template worker readiness validation
  - Reverse cleanup of partially allocated email adapters
  - Registered actual-resource partial-startup and bounded active-delivery drain proof
affects: [02-24, 02-25, 02-30, 02-35]
tech-stack:
  added: []
  patterns: [local email readiness, atomic adapter close state, partial allocation ownership]
key-files:
  created: []
  modified:
    - apps/backend/internal/app/worker.go
    - apps/backend/internal/app/roles_test.go
    - apps/backend/internal/app/invitation_worker_test.go
    - apps/backend/internal/config/config.go
    - apps/backend/internal/config/config_test.go
    - apps/backend/internal/lib/email/client.go
    - docs/development.md
    - scripts/check.ts
key-decisions:
  - Readiness validates both embedded templates, the complete invitation configuration and an open configured email adapter locally; provider availability never gates readiness.
  - Register an email allocation before handling its factory error; retain dependencies beneath active delivery even when the caller shutdown deadline expires.
requirements-completed: []
requirement-slices: [TEN-03, TEN-07, TEN-08, SAFE-03, SAFE-04]
duration: 14min 48s
duration-basis: Measured execution from recorded plan start through verified production task commit; initial context setup and metadata closeout excluded
completed: 2026-10-10
---

# Phase 2 Plan 23: Invitation Worker Readiness and Shutdown Summary

**Worker readiness checks real owned PostgreSQL/Redis plus local keys, both templates and an open email adapter; partial email allocations unwind safely and actual invitation sends retain dependencies through bounded shutdown.**

## Performance

- Measured execution: **14min 48s**, from **2026-10-10T21:58:49Z** to **2026-10-10T22:13:37Z**. No interruptions occurred in this window.
- Tasks completed: **2/2**.
- Concrete task paths: **8**, all declared by plan 23. No task-created source files or tracked deletions.

## Accomplishments

- Preserved the worker PostgreSQL ownership and required database configuration already implemented in plan 21. Actual database and Redis readiness probes remain in the real role graph.
- Added transport-free email `CheckLocal`: cancellation, configured/open welcome and invitation transports, and local rendering of both embedded templates. Worker readiness also validates the complete invitation key ring, active key, sender and public origin. An empty client, closed adapter, malformed/missing active key or blank provider credential cannot report ready. No provider ping or email send occurs during readiness.
- Email close state uses an atomic flag, is idempotent, closes owned idle invitation connections and prevents subsequent delivery through the disposed adapter. Rendering remains independently available without configured transport.
- Worker construction registers an email allocation before returning its factory error. Existing reverse ownership releases the partial adapter before Redis/PostgreSQL; original failures stay privately unwrap-able, while startup/configuration/health diagnostics expose no key or provider values.
- Explicitly registered `TestInvitationWorkerRoleOwnership` in the existing integration inventory. Real migrated PostgreSQL, Redis and default factories exercise allocated-resource failures at database, Redis, email, consumer and consumer-start stages. Tests verify closed actual pools/clients, disposed adapters and reverse consumer/Redis/database cleanup.
- An actual queued invitation enters a bounded injected local sender. Shutdown stops intake and returns at its 300 ms caller deadline while the send retains usable PostgreSQL/Redis. Releasing the send commits Delivered and ciphertext erasure before the worker closes Redis/PostgreSQL and releases its listener. No production lifecycle rewrite or provider call was needed.
- Updated development documentation to describe existing worker database requirements, shared invitation configuration, truthful readiness and active-work cleanup. Optional API producers, consumer-free API composition, explicit migrations, independent redirectors and old roles without new authentication prerequisites remain intact.

## Task Commits

1. **Task 1: Specify keep invitation worker readiness and shutdown honest** — `0fada34` (test).
2. **Task 2: Deliver keep invitation worker readiness and shutdown honest** — `6cd652d` (feat).

Normal commits were used without hook bypass. Neither task commit deletes tracked files.

## Verification

- Meaningful focused RED: empty email client, missing active key, malformed key, partial email adapter remaining usable and whitespace-only worker provider credential. The actual ownership suite also exposed absent closed-adapter readiness. Fixture errors were not accepted as feature RED.
- Original canonical root RED unit gate failed specifically on the two new worker tests, after successful workspace/script/tool checks.
- Focused race-enabled app/config/email checks passed; the actual-resource worker suite covers **7 direct scenarios** (five startup stages, dependency probes and active drain).
- Final canonical root unit gate passed **106 selected race-enabled Go units**, **28 frontend units**, **6 email tests**, all canonical package/script suites and tool self-tests. Existing finite welcome export and invitation rendering/escaping checks remain intact.
- Final canonical integration gate passed **19 registered tests across 6 groups**, including actual HTTP health, independent process roles, PostgreSQL/Redis delivery/recovery, old welcome retry/correlation, telemetry/outage and resource cleanup checks. No skipped or missing required Go test was accepted.
- Original full native browser gate passed **40/40 on its first run**, with strict positive-count, zero-skip/flaky/infrastructure validation. No browser fixture, runner, retry or gate relaxation was used.
- Root formatting, strict lint and typechecking passed. A first lint run identified eight introducing constant/style/test issues; all were corrected in declared files, and native golangci-lint reported **0 issues** before the complete root lint passed.
- Migration checks and five-artifact generation reproducibility passed. Both OpenAPI copies, generated Go transport and both email HTML outputs remain unchanged.
- Final root production build passed. Existing `cleanFrontendBuild` validated real directory ownership, nonsymlink containment and production prerender manifest v4 before disposing generated frontend output.
- **CI=true full local scan passed**: no vulnerable imported Go package across roles/tests, clean Bun dependencies, clean worktree and full-history secrets. Existing unused GO-2026-5932 inventory advisory remains visible; no advisory exception was introduced. This is local CI-mode evidence, not a hosted workflow claim.

## Deviations from Plan

None — the plan's declared scope covered the readiness, configuration, cleanup, registered tests and documentation refinements. Existing job lifecycle and PostgreSQL configuration required no reimplementation.

## Issues Encountered

The first new readiness fixture deliberately closed shared Redis beneath a running Asynq consumer, triggering an upstream PubSub nil dereference. Corrected the test protocol to stop intake and drain the actual consumer before closing resources for readiness probes. That fixture failure was excluded from RED evidence. The corrected original suite and full integration gate passed. Routine introducing lint issues were corrected without altering checks or unrelated source.

## Known Stubs

None. Modified production files contain no placeholder path, unwired data source or stub preventing the plan goal.

## Next Plan Readiness

Ready for **02-24**; it was not started. All broad requirement acceptance remains pending: invitation inspection/acceptance/resend/revoke, later tenant/security refinements and actual configured provider factors/session/delivery evidence remain assigned to their plans, including final plan **02-35**. Local injected senders and SDK fixtures do not satisfy live-provider acceptance. Phase 2 is not complete.

## Self-Check: PASSED

- All eight modified task paths exist.
- Both task commits `0fada34` and `6cd652d` exist; their combined range contains no tracked deletions.
- Registered test inventory contains 106 Go units and 19 integration tests in six groups; required native gates passed as recorded above.
- No source file was created by the tasks; the summary file is created and committed during closeout.
- Unrelated pre-existing `.serena/project.yml` changes remain unstaged and untouched.

## CI Tracing Repair Addendum — 2026-10-10

After plan closeout, hosted run `38089621064` failed the registered `TestTracestateHTTPRedisOTLP` case. Its assertion text was unavailable in the hosted runner output. Three isolated original race runs and the original registered app group (12/12) passed locally; these passes did not establish a cause for the historical failure.

**Confirmed defect:** the tracing proof waited for three collector export requests, although earlier trace, log and metric exports could satisfy that count before the collector's final asynchronous retry trace batch arrived. SDK shutdown flushes to the collector receiver and does not imply that downstream collector batches have completed. A deterministic regression using the actual digest-pinned collector held the final retry trace batch at its HTTP output while the three earlier signals arrived. The original barrier incorrectly declared completion and failed the new assertion in 1.41 seconds. This establishes the barrier defect under controlled ordering; it does not identify the hidden assertion from the historical hosted failure.

**Repair:** decode actual official OTLP protobuf exports and wait for the HTTP span, exactly four job execution spans (two initial and two retry), and nonempty log and metric evidence. Preserve the original eight-second deadline, race execution, privacy checks and complete trace/parent/link/UUID assertions. Malformed protobuf and unexpected signal paths fail immediately with value-free messages. The final lineage assertion reports only numeric counts and a boolean. No production instrumentation, provider calls, telemetry privacy policy or collector configuration changed.

**Fixture ownership:** new `apps/backend/internal/app/collector_barrier_test.go` owns the controlled HTTP sink gate and public synthetic OTLP span fixture. The gate releases on every exit; the existing collector factory owns the actual collector container, validated configuration and HTTP cleanup. `scripts/check.ts` registers `TestCollectorExportCompletenessBarrier` in the existing app integration group. Registered inventory is now 20 integration cases in six groups; Go unit inventory remains 106.

**Atomic commits:**
- `97bb1ff` — RED: reproduce incomplete collector export barrier with the actual pinned collector.
- `afb7023` — GREEN: await complete decoded evidence and retain strict assertions.

**Validation:** controlled regression plus original tracing race tests passed; the final exact registered app integration group passed 13/13 with `-race -count=1 -timeout=8m`. Native Go JSON stdout was retained in private OS temporary files (directory mode 0700, report mode 0600), with only test identifiers and counts reported. All 16 native `scripts/check.test.ts` checks passed, including compiled Go inventory and gate enforcement. Root format, lint and typecheck passed. Typecheck's production build output passed the exported `cleanFrontendBuild` ownership/manifest checks before disposal; `CI=true bun run scan` passed imported Go dependencies, Bun dependencies, worktree and full-history secret scans. Browser tests were not repeated for this test-only tracing repair. The earlier plan closeout's browser and production verification evidence remains recorded above.

**Addendum self-check: PASSED.** The new collector fixture and both modified paths exist; both repair commits exist and contain no tracked deletions. `.serena/project.yml` remains untouched and unstaged. This addendum does not advance STATE/ROADMAP: plan 24 remains next, phase acceptance and live provider verification in plan 35 remain pending, and no push or subsequent plan was performed.
