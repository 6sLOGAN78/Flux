---
phase: 01-foundation-stability-system-boundaries
plan: "20"
subsystem: testing
tags: [bun, go, race, biome, golangci-lint, staticcheck, turbo, quality-gates]
requires:
  - phase: 01-19
    provides: Checksum-verified pinned quality tools and compatible configurations
  - phase: 01-08
    provides: Isolated byte-for-byte generation verification
provides:
  - Executable format, lint, typecheck, test and build commands in all three implemented workspaces
  - Complete fast/full backend and workspace quality orchestration with bounded subprocess execution
  - Compiled Go test discovery compared with complete unit/integration selections and race-enabled execution
affects: [01-21, 01-22]
tech-stack:
  added: []
  patterns: [Uncached explicit quality stages, compiled test-list equality, fail-closed test execution counts]
key-files:
  created: [scripts/check.ts, scripts/check.test.ts, packages/zod/src/health.test.ts, packages/emails/src/templates/welcome.test.tsx]
  modified: [package.json, turbo.json, packages/zod/package.json, packages/openapi/package.json, packages/emails/package.json, apps/backend/taskfile.yml]
key-decisions:
  - "Discover all implemented workspaces, build dependencies before consumers, compare source-selected Go tests against Go's compiled listing, and require successful execution of every selected top-level test."
  - "Keep backend Taskfile migrate:check as two direct embedded migrator invocations; the root migration gate runs its existing pinned-PostgreSQL regression to avoid recursive task/test dispatch."
patterns-established:
  - "Root check:fast runs 43 actual stages; check adds six container/migration/scanner stages for a 49-stage manifest."
  - "All 89 current Go tests belong to exactly one execution group: 76 unit and 13 integration; both run with the race detector and cache disabled."
requirements-completed: [PLAT-05, PLAT-06]
duration: 46min
completed: 2026-10-07
---

# Phase 01 Plan 20: Complete Backend and Workspace Quality Gates Summary

**An uncached 43-stage fast gate verifies Go and all three TypeScript workspaces, while a 49-stage full manifest adds real container regressions, migration verification and the next plan's scanner dispatch.**

## Performance

- **Duration:** 46min, including the prerequisite repair checkpoint and recovery.
- **Recorded start:** 2026-10-07T16:11:43Z
- **Completed:** 2026-10-07T16:57:18Z
- **Tasks:** 2
- **Owned files modified:** 10
- **Coordinated prerequisite repair:** 86 additional authored source/test files; separately committed and documented in `01-20-LINT-REPAIR.md`.

## Accomplishments

- Added strict health-schema acceptance/rejection tests, email text escaping and Go-token preservation tests, and two real locked-CLI email exports that must match each other and the embedded HTML by bytes.
- Every implemented workspace now defines real format, lint, typecheck, test and build commands. Email build includes deterministic export. Reused existing locked Node declarations for test compilation without adding dependencies.
- Root quality execution invokes Go formatting, module integrity, vet, unrestricted staticcheck/golangci-lint, TypeScript formatting/lint/typechecking, all authored Bun suites, tool integrity/self-tests, all Go unit groups with race detection, all-role binaries and uncached isolated generation comparison.
- Discovery rejects missing scripts, zero source tests, omitted required workspaces/backend, duplicate test names and missing classified integration names. Go's compiled JSON test listing must equal the selected source manifest; zero executed or skipped selected tests cannot count as success.
- Full dispatch includes all 13 container-backed regressions, clean/repeated migration and real migrator-binary verification, and the root scanner command. Eight injected orchestration tests prove stage completeness and error propagation. Scanner execution remains owned by 01-21; the complete real CI gate remains owned by 01-22.
- Subprocesses have explicit deadlines, a 16 MiB capture ceiling, process-group termination and stable diagnostics without raw captured tool output. Turbo declares build ordering and outputs, while verification tasks disable caching. Empty `apps/frontend` receives no invented scripts or UI.

## Task Commits

1. **Task 1 RED:** `7a9bbe5` — `test(01-20): specify workspace checks and schema email behavior`.
2. **Task 1 GREEN:** `6ec6a53` — `feat(01-20): run checks tests and builds in every workspace`.
3. **Task 2 RED:** `e861be7` — `test(01-20): require complete quality stage dispatch and failure propagation`.
4. **Task 2 GREEN:** `0a33933` — `feat(01-20): enforce complete backend and workspace quality gates`.

Prerequisite recovery commits: `d73564a` (authored TypeScript quality), `af3e83a` (backend source quality), and `eb1664f` (repair evidence). All commits used normal hooks; no tracked file was deleted.

## Files Created/Modified

- `packages/zod/package.json`, `packages/openapi/package.json`, `packages/emails/package.json` — actual workspace checks, tests and builds using existing dependencies and exact installed tools.
- `packages/zod/src/health.test.ts` — documented state acceptance and rejection of unknown states, absent fields and private diagnostics.
- `packages/emails/src/templates/welcome.test.tsx` — escaping, preserved template tokens, deterministic render/export and checked artifact equality.
- `scripts/check.ts` — discovery, complete explicit stage manifest, compiled test-list equality, bounded execution and failure propagation.
- `scripts/check.test.ts` — successful full dispatch, fast/full scope comparison, missing/zero/omitted checks, test-list mismatch, skipped tests, stage failures and real subprocess timeout regressions.
- `package.json` — canonical full/fast checks, stage commands, generation, verified tool installation and future scanner entrypoints.
- `turbo.json` — correct build/test/typecheck dependency ordering, build outputs and uncached quality tasks.
- `apps/backend/taskfile.yml` — discoverable quality aliases while preserving direct migration behavior and the existing independent binary targets.

## Verification

- `bun test packages/zod/src/health.test.ts packages/openapi/src/gen.test.ts packages/emails/src/templates/welcome.test.tsx`: PASS, 17 tests.
- `bun run --filter '@flux/*' build` and `typecheck`: PASS for all three workspaces, including the real locked email CLI export.
- `bun test scripts/check.test.ts`: PASS, eight orchestration regressions. Full scanner dispatch is injected only; no real scan claim is made.
- `tsc -p scripts/tsconfig.json` and owned-file Biome checks: PASS.
- **Final actual `PATH=/usr/local/go/bin:$PATH bun run check:fast`: PASS**, all 43 stages including the compiled test discovery check, all 76 Go unit tests with race detection, actual workspace checks/tests/builds, all five command binaries and generation drift verification. Evidence: ignored `tmp/plan20-fast-final.log`.
- Full manifest discovery: 49 stages; exactly 89 current Go tests, 76 unit plus 13 integration, with no intersection or omitted source test. Every selected Go test must emit a successful top-level JSON event.
- Prerequisite repair independently passed full `go test -race -count=1 ./...`, including actual PostgreSQL, Redis and collector paths, plus unrestricted zero-finding lint and isolated artifact-byte verification. See `01-20-LINT-REPAIR.md`; these are separate repair acceptance results, not a claim that this plan executed the future real scanner/full CI gate.
- `git diff --check`: PASS. No generated artifacts, dependency graphs or lint rules changed in this plan.

## Decisions Made

Use explicit root stages instead of relying on Turbo to silently skip missing workspace scripts or the backend. Dependency builds run before export-consuming type checks and tests. Compile-list equality catches a source signature or build constraint that would otherwise disappear from a selected test regex. Exact integration names remain checked while new ordinary tests enter the unit group automatically.

Keep the backend's migration task directly invoking the migrator twice. Its existing integration test invokes that task, so delegating the task back into the root migration-test stage would recursively invoke the test. The root stage instead calls the existing complete migration integration proof directly.

## Deviations from Plan

### Coordinated prerequisite recovery

**1. [Rule 3 - Blocking] Existing configured source gates failed outside the ownership union**
- **Found during:** Initial actual quality checks for Task 2.
- **Issue:** 17 Go formatting failures, 862 uncapped Go lint findings across 77 files, eight TS formatting failures and six TS lint warnings blocked the required real fast gate.
- **Action:** Reported exact files/counts to the parent and returned a checkpoint after verified Task 1 and Task 2 RED. The parent explicitly coordinated source repairs outside this executor's ownership; no rules were disabled and no stage was skipped.
- **Repair:** `d73564a`, `af3e83a`, with detailed preservation, annotation and verification evidence in `01-20-LINT-REPAIR.md`, committed as `eb1664f`. All 86 prior Go tests remain, plus three new rollback regressions.
- **Verification:** Fresh final real fast gate PASS, compiled discovery equality PASS, and separately recorded full repair race/lint/build checks PASS.

**2. [Rule 1 - Bug] Avoided recursive migration task/test dispatch**
- **Found during:** Coordinated full migration regression execution before Task 2 acceptance.
- **Issue:** The drafted Taskfile alias called the root migration-test command, whose integration regression itself calls the Taskfile alias.
- **Fix:** Restore the task's committed direct migration implementation while preserving the new quality aliases. The root migration stage continues to invoke the complete existing pinned-PostgreSQL integration proof.
- **Files modified:** `apps/backend/taskfile.yml`, owned by this plan.
- **Verification:** Repair's fresh full race suite passed the migration proof; injected full dispatch and final actual fast gate PASS.
- **Committed in:** `0a33933`.

## Issues Encountered

The first checkpoint deliberately left Task 2's implementation uncommitted and the plan incomplete while the parent corrected prerequisite sources. On resume, the executor reviewed the repair report, fixed its owned email-test import organization and verified the real gate before committing GREEN. Normal iterative runner-test corrections were completed before acceptance.

## User Setup Required

None. Exact quality binaries are installed through `bun run tools:install`; use the pinned Go executable on PATH. Full integration/migration acceptance requires the already-pinned local container runtime and images.

## Next Phase Readiness

Ready for 01-21 scanner implementation and 01-22 clean-environment/full CI acceptance. The scanner entrypoints intentionally point to the next plan's implementation and are exercised through injected runners here. This is an explicit ordered dependency, not a completed production scanner or a full phase completion claim.

## Self-Check: PASSED

All ten owned artifacts and the prerequisite repair report exist. Both RED and both GREEN commits exist in order, as do all three repair commits. The final actual fast gate exited zero. No goal-blocking production stubs or security surfaces outside the declared contributor/tooling trust boundaries were introduced.
