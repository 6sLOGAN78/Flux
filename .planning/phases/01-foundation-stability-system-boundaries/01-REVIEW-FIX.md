---
phase: 01-foundation-stability-system-boundaries
fixed_at: 2026-10-08T00:34:41Z
review_path: .planning/phases/01-foundation-stability-system-boundaries/01-REVIEW.md
iteration: 1
findings_in_scope: 7
fixed: 7
skipped: 0
status: all_fixed
---

# Phase 01: Code Review Fix Report

**Fixed at:** 2026-10-08T00:34:41Z
**Source review:** .planning/phases/01-foundation-stability-system-boundaries/01-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 7
- Fixed: 7
- Skipped: 0

## Fixed Issues

### CR-01: Collector exports private resource and scope schema URLs

**Status:** fixed
**Files modified:** `deploy/otel-collector.yaml`, `apps/backend/internal/app/observability_test.go`
**Commit:** d0a1194
**Applied fix:** Clear resource and scope schema URLs in every signal. Extend the existing dirty sender fixture with six envelope URLs, outside the attribute maps. Both accessors are supported by the exact digest-pinned Collector 0.162.0; the test validates configuration using that binary and checks outgoing protobuf plus debug output.
**Verification:** Expanded fixture failed before the configuration change with retained private input. `TestCollectorRedactsAllSignals` passed afterward and again under the race detector. No collector version, privacy allowlist, or span-link decision changed.

### CR-02: Root check rewrites stale artifacts before its drift gate

**Status:** fixed: requires human verification
**Files modified:** `scripts/check.ts`, `scripts/check.test.ts`, `packages/emails/package.json`, `packages/openapi/src/gen.test.ts`
**Commit:** 3c036a4
**Applied fix:** Run isolated generation comparison before tools, builds, or tests in full/fast checks. Email builds only compile; explicit generation/export remains responsible for publishing. OpenAPI CLI tests use temporary output paths and assert that checked outputs remain untouched.
**Verification:** All 16 real root CLI fixtures passed: full and fast modes each rejected changed/missing canonical JSON, served JSON, generated Go transport, and email HTML at the initial gate. Every fixture asserted preservation of the initially corrupted/missing file, other artifact bytes, and authored schema bytes. Root runner suite passed 9/9 at this step; OpenAPI CLI tests passed; scripts typecheck and email build passed without tracked asset changes. The full/fast fixture group took approximately 65 seconds, stopping before any later stage.

### CR-03: Fresh CI requires Task but does not install or pin it

**Status:** fixed
**Files modified:** `tools.lock.json`, `scripts/install-tools.ts`, `scripts/check.ts`, `scripts/check.test.ts`, `.github/workflows/ci.yml`, `docs/development.md`
**Commit:** 199d3bf
**Applied fix:** Add official standalone Task 3.54.0 with archive and executable SHA-256 pins for Linux/macOS x64/arm64. Include Task in installer/version verification, prepend verified tools for every root subprocess, publish that path to later CI steps, and document direct-test prerequisites.
**Verification:** Downloaded all four official archives and checked archive hashes against the upstream checksum manifest; independently computed each binary hash. Shared installer and configuration verification passed. A new temporary install directory with no Task received the verified official binary while host PATH excluded the existing Task; launching Task using only the temporary directory returned 3.54.0. Existing other verified tool binaries were copied into that temporary directory to keep this focused proof scoped. A regression also executes Task through the root runner with a host PATH containing no executables.

### CR-04: API changes database passwords containing spaces

**Status:** fixed: requires human verification
**Files modified:** `apps/backend/internal/database/database.go`, `apps/backend/internal/database/migrator.go`, `apps/backend/internal/app/roles_test.go`, `scripts/check.ts`
**Commit:** 3f78610
**Applied fix:** Share one PostgreSQL URL builder using `url.UserPassword`, encoded path/RawPath, host-port joining, and query values. Register the new real-database API/migrator parity test as an integration test.
**Verification:** The regression creates disposable PostgreSQL role/database names and a password containing spaces and reserved URL characters. Before the fix, migration succeeded but API construction failed. Afterward both migration and actual API construction/query succeeded with the same unchanged identity; the test passed under the race detector. No credential or driver diagnostic is exposed on failure.

### CR-05: Scanner and installer timeouts leave descendants running and can wait indefinitely

**Status:** fixed: requires human verification
**Files modified:** `scripts/subprocess.ts` (new), `scripts/subprocess.test.ts` (new), `scripts/scan.ts`, `scripts/install-tools.ts`
**Commit:** 730c0f2
**Applied fix:** Share bounded capture with POSIX group creation/termination on supported Linux/macOS platforms, a 100 ms final drain deadline, stream destruction/unref fallback, output limits, and fixed private-error boundaries. Stderr is drained/count-limited without retention. New helper/test files are explicitly required by this finding's shared-helper and regression recommendation.
**Verification:** Original scanner/installer implementations each settled after roughly 2000 ms despite a 30 ms deadline for an inherited-pipe child. All five new regressions passed: nested timeout and output overflow for both wrappers, actual descendant termination, and bounded settlement when an intentionally detached pipe holder escapes the group (fixture cleanup terminates that escaped process). Existing bounded scanner/CLI regressions and installer self-test passed. Scripts typecheck and lint passed. Linux behavior was executed; macOS uses the same POSIX process-group API but was not executed here.

### WR-01: Exported OpenAPI document contains unresolved schema references

**Status:** fixed: requires human verification
**Files modified:** `packages/openapi/src/index.ts`, `packages/openapi/src/gen.ts`, `packages/openapi/src/gen.test.ts`
**Commit:** 65b2360
**Applied fix:** Preserve generated components when adding security schemes; serialize that complete exported document and remove duplicate generation/recovery.
**Verification:** Exported-reference regression failed before the change and passed afterward. All 12 OpenAPI tests passed, including recursive local-reference resolution, exported-document JSON semantic equality, and equality with the checked canonical bytes. Workspace build/typecheck and real `generate:check` passed; canonical artifacts were unchanged.

### WR-02: HTTPError.Is treats every HTTP error code as equal

**Status:** fixed: requires human verification
**Files modified:** `apps/backend/internal/errs/http.go`, `apps/backend/internal/errs/http_test.go` (new)
**Commit:** 1f56165
**Applied fix:** Compare stable codes after nil-safe receiver/target type checks. Add the explicitly requested regression test file.
**Verification:** Differing codes, wrapped differing errors, and typed-nil target/receiver cases failed before the fix. All matching/differing/wrapped/nil/unrelated-type cases passed afterward under the race detector.

## Verification Boundaries

Re-read each modified source section and performed the applicable compiler/type checks. Final changed-scope Go lint (`internal/errs`, `internal/database`, `internal/app`) reported zero issues; Biome formatting/lint and scripts/OpenAPI typechecks passed. Go checks used `GOTOOLCHAIN=go1.26.8`; Node was 22.23.3 and Bun was 1.3.14. A fresh frozen Bun install included the checked patch directory.

Logic/state findings retain the required human-verification label even where meaningful automated regressions passed. This report does not assert full-suite, hosted-CI, re-review/security, or phase-goal completion; those gates belong to the parent workflow. Seven final commits each contain one finding; test-style adjustments were folded into their respective commits on the isolated unpublished branch. No checks or policy allowlists were weakened.

---

_Fixed: 2026-10-08T00:34:41Z_
_Fixer: the agent (gsd-code-fixer)_
_Iteration: 1_
