---
phase: 01-foundation-stability-system-boundaries
plan: "21"
subsystem: security
tags: [bun, govulncheck, gitleaks, redaction, dependency-audit, tdd]
requires:
  - phase: 01-19
    provides: Checksum-verified exact standalone quality tools
  - phase: 01-20
    provides: Root scanner commands and full quality-stage dispatch
provides:
  - Fail-closed bounded private scanner capture with sanitized findings
  - All backend roles and tests checked for vulnerable imported Go packages
  - Root Bun lock audit and redacted worktree/full-history secret checks
  - Real vulnerable-import and deleted-secret mutation regressions
affects: [01-22, ci, security]
tech-stack:
  added: []
  patterns: [Private bounded scanner reports, Closed public metadata, Package exposure with explicit module inventory]
key-files:
  created: [scripts/scan.ts, scripts/scan.test.ts, .gitleaks.toml]
  modified: [tools.lock.json]
key-decisions:
  - "Analyze every backend package and test with govulncheck -scan=package -test ./...; fail every vulnerable import while explicitly retaining unused module inventory advisories."
  - "Keep reports in bounded memory, drain stderr without retention, and publish recognized identifiers plus manifest-backed package names and safe file locations."
patterns-established:
  - "No advisory ignores, secret baselines, directory suppressions or inline allow comments; real mutations live only in disposable runtime trees."
requirements-completed: [SAFE-08, PLAT-06]
duration: 36min
completed: 2026-10-07
---

# Phase 01 Plan 21: Safe Dependency and Secret Scanners Summary

**Pinned Go, Bun and Gitleaks checks fail safely on vulnerable imports, lock advisories, secrets and execution errors, with private bounded reports and real mutation coverage.**

## Performance

- **Duration:** 36min elapsed, including the coordinated prerequisite repair checkpoint.
- **Started:** 2026-10-07T22:47:02Z
- **Completed:** 2026-10-07T23:23:00Z
- **Tasks:** 2
- **Owned source/config files changed:** 4

## Accomplishments

- Real govulncheck v1.8.0 runs `-json -scan=package -test ./...` across the entire backend, covering every role and test dependency. Protocol, exact scanner version, source/package scan mode and Go1.26.8 SBOM are validated. Every finding with an imported package or symbol fails; malformed or incomplete reports fail closed.
- Bun1.3.14 audits the root lock with `audit --json`; every advisory fails. Registry/service errors, malformed JSON and inconsistent nonzero empty reports fail safely. The already-wired root `scan`, `scan:dependencies`, `scan:secrets` and full `check` entrypoints are preserved.
- Gitleaks8.30.1 runs redacted worktree and `--all --full-history` scans, without an ignore baseline or directory suppression. Local trees without Git still receive a worktree scan and an explicit history-unavailable diagnostic. CI fails for absent or shallow history. Inline allow comments and `.gitleaksignore` cannot bypass checks.
- Capture has a 240-second deadline, an 8MiB combined output budget and at most 100 public findings per stage. Raw reports stay in memory; stderr is drained without retention. Unknown metadata becomes stable opaque identifiers; package labels are checked against owned manifests and file labels against safe existing paths. Versions cannot carry arbitrary prerelease text.
- Temporary runtime fixtures prove real worktree detection, detection after a secret is removed from Git history, and failure after adding a legacy OpenPGP import. No token-shaped fixture was committed and no test dependency was installed.

## Task Commits

1. **Task 1: Specify and prove the owned behavior** — `5371e9d` (`test`): six expected preimplementation failures. Additional scope/redaction RED — `14a62eb` (`test`): four expected failures before correction.
2. **Task 2: Implement safe dependency and secret findings** — `213da88` (`feat`): initial wrappers and eight passing regressions. Final accepted scope/protocol/redaction correction — `887f6c0` (`fix`): twelve passing scanner regressions and the actual combined scan.

Coordinated prerequisite commits: `6da9ca5` (`fix`), `98ef471` (`docs`). Their details and verification are recorded in `01-21-DEPENDENCY-REPAIR.md`.

## Files Created/Modified

- `scripts/scan.ts` — Pinned scanner verification, bounded capture, validated Go protocol, Bun audit, redacted worktree/history analysis and safe CLI exit propagation.
- `scripts/scan.test.ts` — Twelve behavior regressions covering findings, failure injection, metadata redaction, scope, history, real tools and command propagation.
- `.gitleaks.toml` — Pinned default secret rules with no fixture, directory or credential allowlists.
- `tools.lock.json` — Exact scanner commands, official evidence, package/test analysis scope and output/history policy. Existing executable checksums remain unchanged.
- `package.json` — Existing scanner entrypoints from 01-20 verified and consumed; this executor did not rewrite the coordinated dependency overrides.

## Verification

- `bun test scripts/scan.test.ts`: PASS, 12 tests. Actual Go mutation imports `golang.org/x/crypto/openpgp` in a temporary module using the existing authoritative go.mod/go.sum and proves GO-2026-5932 becomes a failing imported-package finding.
- `bun test scripts/check.test.ts`: PASS, 8 tests, including scanner failure propagation into full root dispatch and stage completeness.
- `tmp/tools/biome check scripts/scan.ts scripts/scan.test.ts tools.lock.json`: PASS, zero findings or edits.
- `node_modules/.bin/tsc -p scripts/tsconfig.json`: PASS.
- Fresh actual `bun run scan`: PASS, exit zero. Bun dependencies, secrets worktree and full Git history are clean. Go reports no vulnerable imported packages in any backend role or test.
- **Module inventory is not clean:** GO-2026-5932 remains visibly reported for x/crypto v0.57.0, with no fixed version reported. Its unmaintained OpenPGP packages are not imported. This is a generic package-exposure policy, not an advisory exception; the real import mutation fails on that same advisory.
- `git diff --check`: PASS. No tracked deletions; runtime fixtures are removed and tool outputs remain ignored.

The coordinated prerequisite repair separately passed frozen Bun installation, the real fast gate and the full Go build/race suite. This plan does **not** claim the future real full CI gate or Phase 1 verification has passed; those remain 01-22/orchestrator work.

## Decisions Made

Use package-level Go exposure over all roles/tests instead of an API-only module scan or default symbol reachability. Module-only diagnostics are retained explicitly, and every imported vulnerable package fails without advisory IDs in policy code. Recognized public fields and manifest/file-backed metadata prevent scanner-controlled labels from leaking arbitrary injected credentials.

## Deviations from Plan

### Coordinated prerequisite recovery

**1. [Rule 3 - Blocking] Existing vulnerable dependency graphs prevented real scan acceptance**
- **Found during:** Task 2 actual scans.
- **Issue:** The original lock produced 72 Bun advisories and three Go module advisories. The executor committed independently passing wrappers but left the plan incomplete and returned a checkpoint.
- **Action:** The parent coordinated existing dependency repairs outside this executor's file union, with official package/advisory evidence. No package alternative, ignore baseline or vulnerability exception was introduced.
- **Repair:** `6da9ca5`, documented by `98ef471` in `01-21-DEPENDENCY-REPAIR.md`. Includes Echo/jose/archive/compress fixes, exact existing transitive overrides and removal of unused tsc-alias/concurrently.
- **Verification:** Fresh real scanner gate PASS; parent prerequisite checks recorded separately.

**2. [Rule 1 - Bug] API-only module scope and report metadata needed correction**
- **Found during:** Resume after prerequisite repair and additional RED regressions.
- **Issue:** The initial `cmd/api` module scan omitted test exposure and treated unused packages as imported vulnerabilities. Permissive metadata labels could also echo an injected marker. The pinned SBOM legitimately omits the main module's version.
- **Fix:** Analyze `./...` with `-scan=package -test`, enforce pinned config/SBOM, retain unused inventory visibly, fail all imported exposure and malformed traces, recognize absent optional module versions, and constrain public metadata with manifest/file-backed identities.
- **Files modified:** `scripts/scan.ts`, `scripts/scan.test.ts`, `tools.lock.json`.
- **Verification:** Four correction regressions recorded RED; all twelve final tests and the actual combined scan PASS.
- **Committed in:** `887f6c0`.

## Issues Encountered

Two parent-generated ignored raw OSV captures contained public advisory example tokens and were correctly flagged by the worktree scan. The parent explicitly authorized deleting only those disposable files after preserving the sanitized repair evidence. No path/rule suppression was added. Normal scanner reports remain bounded in memory.

## User Setup Required

None. Existing official tool installation and exact runtime pins are reused; live vulnerability registry access is required. CI must provide full, nonshallow Git history.

## Next Phase Readiness

Ready for 01-22's clean-environment/full CI acceptance and subsequent code/security/goal verification. No product behavior or Phase 2 transition was introduced.

## Self-Check: PASSED

All owned outputs, existing root scanner entrypoints and the prerequisite report exist. RED/GREEN/correction and prerequisite commits exist in order. Final scoped checks and actual combined scan exited zero. No goal-blocking production stubs or security surfaces outside the declared scanner/tooling trust boundaries were introduced.
