---
phase: 02-tenant-safe-link-control-plane
plan: "33"
subsystem: testing
tags: [playwright, quality-gates, go, scanning, generation]
requires:
  - phase: 02-32
    provides: Responsive control plane and 65 native browser cases
provides:
  - Native browser collection bound to complete successful execution
  - Mandatory delivered frontend workspace and rejection of unsuccessful Go subtests
  - Frozen installation, full quality gates and ignored-path secret regression evidence
affects: [02-34, 02-35]
tech-stack:
  added: []
  patterns: [native collection identity comparison, private fail-closed reports]
key-files:
  created: []
  modified:
    - scripts/check.ts
    - scripts/check.test.ts
    - scripts/scan.test.ts
    - docs/development.md
key-decisions:
  - Compare native Playwright spec and project identities with the same selection before accepting execution.
  - Require the delivered frontend and reject Go skip or fail events even below a passing parent.
requirements-completed: []
requirement-slices: [TEN-01, TEN-02, TEN-03, TEN-04, TEN-05, TEN-06, TEN-07, TEN-08, LINK-01, LINK-02, LINK-03, LINK-05, LINK-06, LINK-07, LINK-08, LINK-09, SAFE-02, SAFE-03, SAFE-04]
duration: 17min 18s
duration-basis: Sum of captured active windows; unmeasured pre-quota completion tail and interruption excluded
completed: 2026-10-12
---

# Phase 2 Plan 33: Complete Frozen Quality Gates Summary

**The root gate now binds native browser collection to execution, requires the frontend workspace and refuses skipped or failed Go subtests beneath passing parents.**

## Performance

- Tasks: **2/2**. Modified source/documentation files: **4**, all declared; no source files created.
- Known active windows: **2026-10-12T19:37:48Z–19:50:26Z** (**12min 38s**) and **2026-10-12T20:28:40Z–20:33:20Z** (**4min 40s**), totaling **17min 18s**.
- The full run continued beyond the first window during a quota interruption. Its retained session later returned actual exit-zero completion; its precise completion timestamp is unknown and that tail is excluded from duration. Metadata close-out is excluded.

## Accomplishments

- A bounded native `--list` JSON stage uses the same project/grep selectors as execution. Positive unique spec/project identities must match the successful execution report exactly. Omitted, substituted, duplicate, empty, malformed or globally erroneous evidence fails. Discovery command failure has no fallback acceptance.
- Existing positive execution totals, zero skips/retries/flaky cases, single successful results and aggregate consistency remain mandatory. Reports stay private; public diagnostics disclose stable stage classifications only. The existing browser and stage deadlines are unchanged.
- Root workspace discovery explicitly requires `apps/frontend` alongside all three libraries. Full stages continue direct format/lint/typecheck/unit/build execution, compiler-verified Go discovery, race-enabled selections, migration verification and security scanning without Turbo acceptance.
- Go execution now rejects any skip or fail event, including subtests beneath a successful parent, while preserving all required top-level success checks.
- Real pinned Gitleaks tests prove Git-ignored frontend cache and test-result credentials fail worktree scanning and remain blocking after deletion from committed history. No scanner implementation, rule, exception or secret exclusion changed.
- Development instructions now describe all four workspaces, five generated outputs and the complete collection/execution and cleanup contract.

## Task Commits

1. **Task 1: Capture omitted browser and unsuccessful Go evidence** — `2ec927d` (test).
2. **Task 1: Capture missing delivered frontend discovery** — `ee5a2fa` (test).
3. **Task 2: Enforce complete evidence and document verified gates** — `b7beab2` (fix).

Exact staging and normal commits were used. No tracked files were deleted. Unrelated `.serena/project.yml`, `README.md` and `.planning/quick/` changes were preserved.

## Verification

- Genuine native browser RED: installed Playwright collected **two** real passing test definitions but executed **one** with grep. The report was internally consistent and the old root accepted it; the regression failed with “Missing expected rejection.” The refined root rejects it by manifest mismatch.
- Go skip/fail event regressions and omission of the delivered frontend both produced genuine missing-rejection RED before implementation. These are quality-boundary defects, not product feature or fixture failures.
- Focused guard verification passed **15/15** after the frontend correction; earlier focused scanner/guard verification passed **9/9**, including real ignored-path/history credential scanning. The complete script suites subsequently passed inside the canonical full gate.
- A fresh isolated frozen installation with package manifests, checked patch and lockfile passed; the lockfile stayed byte-identical and its temporary copy was removed. In-place `bun install --frozen-lockfile` also passed without dependency changes.
- Original **CI=true `bun run check` passed** on final immutable source: **65 browser cases**, **109 race-enabled Go unit tests across 16 packages**, **21 integration top-level tests across 6 groups**, **31 frontend units**, **6 email tests**, **15 OpenAPI tests**, schema/script/tool suites, all format/lint/type checks, workspace/backend builds, migrations and scan stages. The original retained subprocess session returned exit **0** and “Checks passed” after resume; no result was inferred from an interrupted progress line.
- Final explicit `bun run generate:check` passed all **five** artifacts: both OpenAPI copies, Go transport, welcome HTML and invitation HTML. Existing isolated authored generation, drift rejection and failure atomicity were preserved without unnecessary generator changes.
- Final explicit `bun run test:e2e -- --project=local` passed **65/65** with native collection binding on final unchanged source. An earlier standalone run also passed 65 before the final required-workspace correction; it is not substituted for this final run.
- After the final browser run, the original production frontend build passed. Exported `cleanFrontendBuild` validated its owned nonsymlink production manifest and removed output. `apps/frontend/build` is absent.
- Original **CI=true `bun run scan` passed** imported Go package scanning across roles/tests, Bun dependencies, complete worktree and full history. The existing unused-module inventory advisory remains visible, with no exception or vulnerable imported package. Final committed metadata/history receive another original CI scan during close-out.

## Issues Encountered

The first full gate rejected an optional-chain lint warning in the new manifest parser. After that run ended, the expression was corrected; separate pinned format/lint and script type checks passed, followed by the complete original full gate. No deadline increase, retry acceptance, private runner, dropped test or scanner waiver was needed.

An initial read-only Playwright collection probe used the wrong executable path and produced no collection. The corrected installed CLI successfully collected 65; the failed probe is not RED or test evidence.

## Deviations from Plan

**[Rule 1 - Bug] Bind self-consistent browser reports to native discovery.** The existing aggregate-only validator could accept an omitted executed test. The owned runner now compares unique native identities; real two-test/one-result RED and all 65 native cases verify it.

**[Rule 2 - Missing critical] Require delivered frontend and reject unsuccessful nested Go events.** Both absent frontend discovery and passing-parent concealment crossed the quality acceptance boundary. Focused regressions and full canonical execution verify the fixes within declared ownership.

Existing scanner, generator, frontend scripts, Playwright configuration and ignore policy already satisfy their relevant contract and were verified rather than rewritten. No runtime, dependency, generated artifact, architecture or production trust boundary changed. Broad requirements remain slices pending final acceptance.

## Known Stubs

None introduced. No placeholder production path or new product behavior is part of this gate refinement.

## Remaining Acceptance

Actual Clerk verification/recovery/social factors, provider session/cookie policy and real approved-recipient invitation delivery remain pending plan **02-35**. Local native SDK fixtures do not satisfy those requirements. No private provider configuration, live keys or recipient email were accessed.

## Next Phase Readiness

Plan 33 is complete; ready for **02-34** within Phase 2. No push, phase completion or transition to Phase 3 occurred.

## Self-Check: PASSED

All four modified files and three task commits exist. Ownership, absence of tracked deletions, unchanged dependency/generated graphs, successful original full and standalone verification, validated build cleanup and preservation of unrelated changes were checked. No private diagnostic capture was created.
