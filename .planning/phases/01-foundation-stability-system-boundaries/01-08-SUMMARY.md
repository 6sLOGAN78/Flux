---
phase: 01-foundation-stability-system-boundaries
plan: "08"
subsystem: infra
tags: [bun, openapi, oapi-codegen, react-email, generation, drift]
requires:
  - phase: 01-05
    provides: Deterministic canonical OpenAPI generator and authored health schemas
  - phase: 01-06
    provides: Checksum-pinned Go transport generator and configuration
  - phase: 01-07
    provides: Deterministic embedded email export and served OpenAPI assets
provides:
  - Root generation script with isolated write and nonrepairing check modes
  - Byte comparisons for canonical and served OpenAPI, Go transport, and email HTML
  - Fixture mutations and real subprocess failure/cleanup regressions
affects: [01-19, 01-20, 01-21, 01-22, foundation-quality-gates]
tech-stack:
  added: []
  patterns: [isolated authored builds, explicit artifact manifest, byte-authoritative drift checks, sanitized tool diagnostics]
key-files:
  created: [scripts/generate.ts, scripts/generated.test.ts]
  modified: []
key-decisions:
  - Build authored packages in a temporary tree with existing installed dependencies and redirect the workspace schema dependency to that isolated build.
  - Invoke exact installed React Email and checksum-verified oapi-codegen directly, rewriting only the copied YAML output destination.
patterns-established:
  - All generators finish before publication; check mode compares bytes without Git or artifact repair.
  - Subprocess temporary files stay inside staging and tool output never enters public diagnostics.
requirements-completed: [PLAT-08, PLAT-06]
duration: 6min
completed: 2026-10-06
---

# Phase 1 Plan 8: Generated Artifact Drift Checks Summary

**Isolated regeneration rejects stale OpenAPI, served JSON, Go transport, and email HTML without relying on Git or changing checked files.**

## Performance

- Started: 2026-10-06T13:39:20Z
- Completed: 2026-10-06T13:45:05Z
- Duration: approximately 6 minutes
- Tasks: 2
- Source files changed: 2

## Accomplishments

- Added `bun scripts/generate.ts` and `bun scripts/generate.ts --check`. Repository paths derive from the script location, so both modes work independently of the caller's working directory.
- Copied only authored source directories, package/TypeScript manifests, and backend module/generator configuration into an isolated temporary tree. Existing installed dependencies are linked for resolution; the OpenAPI package's `@flux/zod` dependency resolves to the isolated authored schema build. Checked `dist` output is never used as contract authority or modified.
- Awaited the actual package builds, OpenAPI command, checksum-verified oapi-codegen v2.8.0, and already installed React Email 6.3.3 CLI. The generator's copied YAML output is explicitly rewritten because it overrides CLI `-o`. No packages were installed.
- Compared an explicit four-artifact manifest by bytes. Missing/unreadable or changed artifacts fail with relative file paths only. Generator failures fail closed without disclosing command output or private causes. Check mode never publishes; write mode reads the complete generated manifest before publishing individual atomic replacements.
- Removed temporary stages and publication directories on success and failure. Subprocess `TMPDIR` stays in staging and optional Node compile caches are disabled.

## Task Commits

1. **Task 1 RED: Specify and prove owned behavior** — `2661837` (test).
2. **Task 2 GREEN: Implement stale artifact rejection** — `2894acc` (feat).

Both used normal Git commits. No tracked files were deleted.

## Verification

- RED: `bun test scripts/generated.test.ts` failed all 20 initial tests because `scripts/generate.ts` was absent.
- GREEN: `bun test scripts/generated.test.ts` passed all 21 tests. Fixtures require identical bytes without Git, mutate/remove each of the four output categories, verify nonrepair and source preservation, inject awaited failure at each category in both modes, and require a complete generated manifest before write publication.
- The real CLI test runs the full pinned generator pipeline from another working directory, snapshots all checked bytes, and asserts its temporary root is empty afterward.
- A real subprocess regression shadows the Go executable with a deliberately failing test-owned tool after actual schema builds/OpenAPI generation. Both CLI modes exit 1, suppress the tool's private output, retain all repository artifact bytes, and clean partial generation.
- Actual `bun scripts/generate.ts`, followed by `bun scripts/generate.ts --check`, passed with no generated-artifact diff. A final standalone `--check` also passed after cleanup hardening.
- `git diff --check`, owned-file placeholder scan, and commit deletion checks passed. The final task checkout was clean.
- Consulted Context7's official Bun subprocess documentation for awaiting `exited` and consuming captured output. Existing verified generator pins and installed package CLI metadata were reused.

## Decisions Made

The authoritative check is byte equality, independent of Git availability. Generation starts from authored source rather than cached build output. Existing dependency installations are reused with a single workspace dependency redirected into staging, so regeneration requires no package installation or new dependency.

Each file replacement is atomic on its destination filesystem. Publication of four files is not a filesystem transaction: a later publication failure may follow an earlier successful replacement, while still returning nonzero and cleaning temporary directories. Generator failures occur before any publication.

## Deviations from Plan

None — the two-file ownership slice and RED/GREEN sequence were preserved. No prerequisite sources or generated artifacts needed modification. The obsolete note that Git was unavailable did not affect byte-authoritative drift verification; normal atomic task commits were available.

## Issues Encountered

The first real-generator regression exceeded Bun's five-second default test timeout; explicit 120-second test/process bounds resolved this. A stricter cleanup regression then exposed an external Node compile-cache directory; subprocess temporary paths were confined to staging and optional compile caches disabled. The initial timeout's one test-owned orphan stage was explicitly removed. Final success and failure cleanup checks passed. No authentication gates occurred.

## User Setup Required

None.

## Next Plan Readiness

Plans 01-19 through 01-22 can expose the script through root generation commands and quality gates. Plan 01-08 is complete; phase execution remains ongoing, and no subsequent plan was executed here.

## Self-Check: PASSED

Both owned source files and this summary exist. RED commit `2661837` and GREEN commit `2894acc` exist and contain no tracked file deletions. All 21 regressions and final actual regeneration check passed, with no prerequisite artifact changes, untracked generated files, production stubs, or unplanned runtime network/authentication surfaces.
