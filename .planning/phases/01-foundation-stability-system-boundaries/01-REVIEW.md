---
phase: 01-foundation-stability-system-boundaries
reviewed: 2026-10-08T06:09:27Z
depth: standard
iteration: 1
reviewed_commit: 8f4f15740500ec8bdf787a62400e7a8a1418a9e7
files_reviewed: 20
files_reviewed_list:
  - .github/workflows/ci.yml
  - apps/backend/internal/app/observability_test.go
  - apps/backend/internal/app/roles_test.go
  - apps/backend/internal/database/database.go
  - apps/backend/internal/database/migrator.go
  - apps/backend/internal/errs/http.go
  - apps/backend/internal/errs/http_test.go
  - deploy/otel-collector.yaml
  - docs/development.md
  - packages/emails/package.json
  - packages/openapi/src/gen.test.ts
  - packages/openapi/src/gen.ts
  - packages/openapi/src/index.ts
  - scripts/check.test.ts
  - scripts/check.ts
  - scripts/install-tools.ts
  - scripts/scan.ts
  - scripts/subprocess.test.ts
  - scripts/subprocess.ts
  - tools.lock.json
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
status: clean
---

# Phase 01: Follow-up Code Review Report

**Reviewed:** 2026-10-08T06:09:27Z  
**Depth:** standard  
**Iteration:** 1  
**Files Reviewed:** 20  
**Status:** clean

## Narrative Findings (AI reviewer)

### Summary

Reviewed the explicit 20-file fix scope, the seven original findings, and the changed database, collector, contract-generation, CI/tool-installation and subprocess boundaries. All seven original findings are resolved in the actual source. No new BLOCKER or WARNING finding was identified in this follow-up scope.

The original 124-file review remains preserved in `01-REVIEW-INITIAL.md`; this report records the bounded follow-up rather than claiming a second review of unchanged files. `clean` applies to this code-review scope. Phase-goal verification, the threat audit and later product functionality remain separate gates.

### Original finding closure

| Finding | Resolution checked in source | Regression/evidence assessed |
| --- | --- | --- |
| CR-01 — collector schema URL disclosure | Trace, log and metric transforms each clear resource and scope schema URLs before export. The exact pinned collector accessors are exercised by the real collector test. | Dirty sender fixtures now inject private-marker URLs into all six envelope fields; outgoing protobuf and collector debug bytes are checked for retention. The fix report records the regression failing before the fix and passing afterward, including race execution. |
| CR-02 — root check repairs initial artifact drift | Full/fast stages begin with isolated `generate:check`; email builds only compile, and the OpenAPI CLI test uses temporary outputs and preserves both checked documents. | Sixteen real root-CLI cases cover changed/missing files in every artifact category in both modes, verify rejection before later stages, and assert that initial bytes/missing files and authored source remain unchanged. |
| CR-03 — missing Task prerequisite | Official Task 3.54.0 archive/binary pins cover the four supported platforms; the shared installer validates and provisions Task. Root subprocesses prepend the installed tools directory and CI publishes it through `GITHUB_PATH`. Direct-test documentation identifies the prerequisite. | The no-host-PATH runner regression exercises discovery. This review independently verified the installed Task binary against the manifest SHA-256 and ran its version command with a PATH containing no executables. |
| CR-04 — inconsistent PostgreSQL credential encoding | API pools and migrator connections call the same URL builder. Userinfo uses `url.UserPassword`, the database component has an encoded `RawPath`, and query options use `url.Values`. | The real PostgreSQL regression uses spaces and reserved characters in role, password and database names, migrates, constructs an actual API and checks its resulting connection identity. It is explicitly registered as integration coverage. |
| CR-05 — unbounded scanner/installer descendant handling | Both wrappers call the shared helper, which starts a POSIX group, kills it on timeout/overflow, and bounds final pipe draining/settlement. Private diagnostics remain suppressed. | Tests cover nested inherited-pipe descendants for timeout and overflow in both wrappers, descendant termination, and bounded settlement when a pipe holder escapes its group. This review additionally exercised missing/empty executable failures; both failed with the closed diagnostic and settled in approximately 7 ms total. |
| WR-01 — invalid exported OpenAPI | The builder preserves generated components while adding security schemes; the serializer consumes that same exported document. The duplicate schema-recovery generation is removed. | Recursive local-reference resolution and exported/canonical document equality regressions cover the previously broken boundary; the CLI test asserts explicit outputs and unchanged checked bytes. |
| WR-02 — every HTTP error compares equal | `HTTPError.Is` now checks nil-safe target/receiver values and stable-code equality. | Tests cover same/different codes, wrapped same/different codes, typed-nil receiver/target and unrelated types. |

### Verification boundaries

Read the changed source and regressions, traced their callers and error/public-output boundaries, and inspected the supplied targeted RED/GREEN and final-quality evidence. The actual post-fix local full quality gate and hosted CI success are recorded in `01-FINAL-CHECK.md`; they supplement the source review and do not substitute for it. The generic fixer-template “requires human verification” labels do not identify an additional unresolved finding.

This follow-up did not repeat the full quality suite. Its independent probes were limited to the installed Task digest/version and subprocess spawn-failure handling. POSIX process-group execution was tested on Linux by the implementation regressions; macOS was reviewed through the shared POSIX implementation and manifest pins, without a claim of executed macOS coverage.

No implementation files were edited and no commits were made during this follow-up. The initial review archive was preserved.

---

_Reviewer: gsd-code-reviewer_  
_Depth: standard_
