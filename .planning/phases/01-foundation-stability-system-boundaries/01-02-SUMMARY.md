---
phase: 01-foundation-stability-system-boundaries
plan: "02"
subsystem: api
tags: [go, echo, validation, race-tests, request-isolation]
requires:
  - phase: 01-01
    provides: Supported Go toolchain and dependency fixtures
provides:
  - Per-invocation request factories for JSON, file, and no-content adapters
  - Sanitized typed binding and validation errors with checked wrapped-error handling
  - Executable concurrent isolation and safe-error regressions
affects: [foundation-quality-gates, product-handlers, tenant-isolation]
tech-stack:
  added: []
  patterns: [per-request factories, checked errors.As, safe public error envelopes]
key-files:
  created: [apps/backend/internal/handler/base_test.go, apps/backend/internal/validation/utils_test.go]
  modified: [apps/backend/internal/handler/base.go, apps/backend/internal/validation/utils.go, apps/backend/internal/testing/transaction.go]
key-decisions:
  - Request factories must return a fresh value on each invocation; no existing registered product consumers require compatibility.
  - Binding failures use Invalid request; validation failures use Validation failed and retain recognized authored field errors.
patterns-established:
  - Synchronize actual JSON decoding and callback observation to deterministically prove concurrent request contamination without depending on scheduling.
requirements-completed: ["PLAT-05"]
duration: 3min
completed: 2026-10-06
---

# Phase 1 Plan 2: Request Isolation and Safe Validation Summary

**Fresh requests for all generic Echo adapters and safe typed HTTP 400 errors for malformed, ordinary, wrapped, and nil validation inputs.**

## Performance

- Duration: approximately 3 minutes
- Started: 2026-10-06T05:57:47Z
- Completed: 2026-10-06T06:00:09Z
- Tasks: 2
- Source files changed: 5

## Accomplishments

- `Handle`, `HandleFile`, and `HandleNoContent` invoke a request factory inside the returned handler, preventing mutable request reuse and omitted-field retention.
- Binding errors return `errs.HTTPError` with `BAD_REQUEST`, status 400, and `Invalid request`, without parsing or returning binder diagnostics.
- Validation safely recognizes wrapped validator/custom errors, including pointer custom errors; ordinary errors retain no raw diagnostic text. Empty or nil error collections still reject the request rather than silently succeeding. Nil payloads and nil validator fields cannot trigger helper panics.
- Corrected `WithTransaction` documentation to describe committing successful callbacks. Its behavior and the rollback-only fixture helper are preserved.

## Task Commits

1. **Task 1: Specify and prove the owned behavior** — `e631a0c` (test, RED).
2. **Task 2: Implement isolate requests and validate safely** — `430c033` (feat, GREEN).

Both commits ran normal Git hooks. No files were deleted.

## Verification

- RED: `go test -race ./internal/handler ./internal/validation -count=1` failed for the intended behavior. All three adapters returned another request's marker and retained an omitted field. Arbitrary/wrapped validation errors panicked, malformed binding leaked diagnostics, and empty typed validation errors incorrectly succeeded.
- GREEN: the same scoped race command passed: handler 1.026s, validation 1.019s. Tests exercise actual Echo JSON binding, concurrent distinct markers, successive omitted fields, arbitrary binder strings, malformed JSON, validator/custom/ordinary/wrapped errors, nil collections/pointers/fields, and valid payloads.
- `go vet ./internal/handler ./internal/validation ./internal/testing` passed.
- `go test ./internal/testing` compiled the helper package successfully; it has no test files, so this is not a claim of transaction integration coverage.
- `git diff --check` passed. Source scans confirmed no binder string-split indexing or unchecked custom-error assertions remain, every wrapper invokes its factory per call, and transaction documentation matches its existing semantics.
- Stub scan found no placeholder production behavior. No new endpoint, file access, authentication path, or schema trust boundary was introduced.
- Library API guidance was checked through Context7 against official Echo v4 and go-playground validator sources.

## Decisions Made

Followed the planned factory API and retained the existing error envelope. Generic binder and validation failures have stable separate messages; explicitly authored custom field errors remain useful to clients. Race regressions synchronize decode completion to make contamination failures deterministic.

## Deviations from Plan

None in implementation or file ownership. The two task blocks repeat final GREEN acceptance criteria, while Task 1 explicitly requires preimplementation failures. Task 1 therefore recorded and committed RED evidence; all shared final acceptance checks passed after Task 2. No downstream fragment was required or changed.

## TDD Gate Compliance

RED `e631a0c` precedes GREEN `430c033`; failures were behavioral, not compile errors. The only GREEN test adjustment supplies the planned request factory API. No separate refactor was needed.

## Issues Encountered

None. Existing migration failure remains assigned to plan 01-04 and does not affect these request tests.

## User Setup Required

None.

## Next Phase Readiness

Ready for plan 01-03. Future consumers must supply a factory returning a fresh validatable request for each invocation. Broader lifecycle, migration, contract, and quality-gate coverage remains with the assigned Phase 1 plans.

## Self-Check: PASSED

Both created test files exist; both recorded task commits exist; scoped race and vet checks pass; changes remain within the five owned source files.
