---
phase: 01-foundation-stability-system-boundaries
plan: "05"
subsystem: api
tags: [zod, ts-rest, openapi, node, bun, contracts]
requires:
  - phase: 01-01
    provides: Verified Node/Bun toolchain and frozen existing dependencies
provides:
  - Strict sanitized live/readiness schemas and complete 200/503 contracts
  - Deterministic awaited OpenAPI generation with staged publication
  - Node/Bun-resolvable built contracts export and generator regression suite
affects: [01-06, 01-11, foundation-quality-gates]
tech-stack:
  added: []
  patterns: [sorted deterministic serialization, module-relative output paths, awaited staging and cleanup]
key-files:
  created: [packages/openapi/src/gen.test.ts]
  modified: [packages/zod/src/health.ts, packages/zod/src/index.ts, packages/openapi/src/contracts/health.ts, packages/openapi/src/gen.ts, packages/openapi/package.json, packages/openapi/openapi.json, apps/backend/static/openapi.json]
key-decisions:
  - Restore named component schemas from the same authored ts-rest contract inside the owned generator, preserving existing operation metadata and security schemes.
  - Use node:test with existing Node typings so the regression suite compiles and runs under both Node and Bun without new dependencies.
patterns-established:
  - Every generated destination is staged before publication; staging errors preserve existing artifacts and all temporary cleanup is awaited.
  - Health schemas expose only literal or coarse enum states and named component checks, with strict rejection of extra diagnostic fields.
requirements-completed: [PLAT-08]
duration: 6min
completed: 2026-10-06
---

# Phase 1 Plan 5: Deterministic Canonical Health Contracts Summary

**Strict Zod live/readiness contracts generate matching OpenAPI 3.0.2 artifacts with awaited writes, observable failures, and Node/Bun runtime exports.**

## Performance

- Started: 2026-10-06T13:07:11Z
- Completed: 2026-10-06
- Duration: approximately 6 minutes
- Tasks: 2
- Source/artifact files changed: 8

## Accomplishments

- Replaced `/status` documentation with `/live` HTTP 200 and `/ready` HTTP 200/503. Kept `info.version` 1.0.0 and OpenAPI 3.0.2.
- Exported `ZHealthLiveResponse`/`HealthLiveResponse` and `ZHealthReadyResponse`/`HealthReadyResponse`, with named `transport.HealthLiveResponse` and `transport.HealthReadyResponse` components. Payloads have only `alive`, `ready|not_ready`, and `{name, state}` checks; unknown diagnostic fields fail schema validation. Examples are static synthetic data.
- Exported deterministic serialization and generation. Object keys sort recursively; array order remains intact; binary file conversion remains supported. Default paths resolve from `import.meta.url`, and importing the generator has no write side effects.
- Staged all outputs before sequential atomic renames, awaited writes and cleanup, preserved failures, and set CLI exit status nonzero on failure. Canonical and served JSON contain identical bytes.
- Fixed the contracts export to `dist/contracts/index.js`, preserved NodeNext `.js` imports, and added a runnable package test script using installed dependencies.

## Task Commits

1. **Task 1: Specify and prove the owned behavior** — `c4e52ca` (test)
2. **Task 2: Implement deterministic canonical health contracts** — `acfdc5f` (feat)

Both commits used normal Git hooks. No tracked files were deleted.

## Verification

- RED: both package builds passed, then all 9 initial tests failed against missing health schemas, missing generator APIs, working-directory-dependent CLI output, and the invalid package export.
- GREEN: `bun run --filter @flux/zod build && bun run --filter @flux/openapi build && bun test packages/openapi/src/gen.test.ts` passed; final suite has 10 passing tests.
- `node --test packages/openapi/dist/gen.test.js` passed all 10 tests against compiled outputs under Node 22.23.3. Bun 1.3.14 also passed runtime export resolution.
- Tests cover strict health payloads, named schema references and all statuses, stable serialization across insertion order, repeated matching output bytes, invalid staging paths at each output index, directly injected asynchronous write failures at each index, rename failure cleanup, alternate working directory CLI generation, nonzero subprocess failures, and both runtime exports.
- `bun run --filter @flux/openapi gen`, canonical/served `cmp`, and `git diff --check` passed. Stub scan found no production placeholders.
- Promise filesystem behavior was checked against the [official Node 22 filesystem documentation](https://nodejs.org/docs/latest-v22.x/api/fs.html); installed ts-rest source confirmed named schema generation from titles.

## Decisions Made

Recovered named schemas from the same authored contract within `gen.ts`, preserving existing operations and security schemes. Used Node's built-in test API to avoid adding Bun typings or dependencies solely for tests.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Preserve generated named schema components**

- Found during: Task 2, naming health transport schemas.
- Issue: The existing unowned `src/index.ts` document builder replaces generated `components` with security schemes. Named response `$ref` values would therefore point at absent schemas.
- Fix: The owned generator restores generated schemas from the same ts-rest authoring router while preserving the existing OpenAPI operation metadata and security schemes. No handwritten transport authority or unowned source edits were added.
- Files: `packages/openapi/src/gen.ts`.
- Verification: Both readiness status schemas and liveness response references resolve to their named generated components.
- Commit: `acfdc5f`.

## Issues Encountered

Context7 MCP and the `ctx7` CLI were unavailable. Used official documentation and installed dependency source without installing packages. No authentication gates occurred.

## Publication Limits

Each rename is atomic on its destination filesystem; publication across two directories is not a filesystem transaction. A later rename failure can leave an earlier destination updated, while the command still fails nonzero and cleans temporary stages. Staging/write failures leave all existing destinations untouched.

## User Setup Required

None.

## Next Plan Readiness

Plan 01-06 can generate Go transport types from the canonical artifact, and plan 01-11 can integrate the routes. This summary completes plan 01-05's TS/Zod portion of PLAT-08; downstream Go generation and handler integration remain assigned to their plans. Phase 1 remains in execution.

## Self-Check: PASSED

All eight owned source/artifact files exist. Both task commits exist and have no file deletions. Builds, Bun/Node regressions, generator execution, matching artifacts, and whitespace checks passed. No new dependencies, production stubs, or unplanned trust boundaries were introduced.
