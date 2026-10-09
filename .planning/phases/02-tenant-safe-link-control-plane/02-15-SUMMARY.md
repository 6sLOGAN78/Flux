---
phase: 02-tenant-safe-link-control-plane
plan: "15"
subsystem: links
tags: [go, postgres, nextjs, search, tenancy, pagination, tdd]
requires:
  - phase: 02-14
    provides: Signed tenant/query-bound cursor pagination and real link creation CTA
provides:
  - Bounded literal key/title/destination search with closed lifecycle filters
  - Exact normalized search/lifecycle cursor binding under fresh SQL membership
  - Native search controls, scoped page resets, draft recovery and stale-response cancellation
affects: [02-27, 02-32, 02-35]
tech-stack:
  added: []
  patterns: [parameterized escaped ILIKE, Unicode whitespace normalization, generation-bound requests]
key-files:
  created: []
  modified:
    - apps/backend/internal/repository/link.go
    - apps/backend/internal/service/link.go
    - apps/backend/internal/service/cursor.go
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/frontend/components/link-list.tsx
    - apps/frontend/lib/links.ts
    - apps/frontend/lib/links.test.ts
    - apps/frontend/tests/control-plane.spec.ts
    - packages/zod/src/identity.ts
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/openapi.json
    - apps/backend/static/openapi.json
    - apps/backend/internal/transport/health.gen.go
key-decisions:
  - Validate raw search as UTF-8 with at most 200 Unicode characters before trimming Unicode whitespace; preserve exact interior text and case for signed query binding.
  - Use fixed SQL parameters and explicit exclamation-mark pattern escaping so percent, underscore, backslash and SQL metacharacters remain literal text.
  - Commit search/filter changes explicitly, invalidate older generations immediately and reset scoped successful page history while retaining drafts during transient failures.
requirements-completed: []
requirements-addressed: [LINK-03, LINK-06, TEN-06, TEN-07, TEN-08, SAFE-03]
duration: 20min
completed: 2026-10-09
---

# Phase 2 Plan 15: Scoped Link Search and Lifecycle Filters Summary

**Literal parameterized PostgreSQL search and closed lifecycle filters deliver real scoped results, signed normalized query continuations and recoverable native browser controls.**

## Performance

- Tasks: 2/2.
- Recorded executable implementation/verification interval: approximately 17:54 UTC–18:13:51 UTC on 2026-10-09, about 20 minutes. Initial context preparation preceded this interval and is excluded; an exact dispatch timestamp was not recorded.
- Actual implementation/test inventory: **14 modified paths**, exhaustively listed above; no new source files. This summary and STATE/ROADMAP tracking are additional documentation paths.

## Accomplishments

- GET links accepts optional literal search and the closed `nondeleted`, `active`, `disabled`, `archived`, `deleted` states. Omitted state remains nondeleted. Raw search must be valid UTF-8, contain no NUL, and use at most 200 Unicode characters before surrounding Unicode whitespace is trimmed. Unknown, duplicate and invalid query fields fail closed; page and cursor limits remain unchanged.
- Repository SQL keeps fixed workspace, limit, timestamp/UUID seek, search and state parameters with fixed ordering. Explicit `ESCAPE '!'` escapes `!`, `%` and `_`; backslashes remain literal under that explicit escape character. Key, title and destination use case-insensitive substring matching. No user-controlled field names, sorting or SQL text are interpolated. Real PostgreSQL remains the authority, with fresh transactional membership checks for every query.
- Existing HMAC cursors bind the exact normalized search text and lifecycle before decoding or signing a continuation. Changed case/interior text, changed lifecycle and another workspace reject valid signed tokens with `400 CURSOR_INVALID`; trimming surrounding whitespace preserves the same effective query. Existing signing-key validation, 2048-byte cursor bounds, roles and startup independence are preserved.
- Native labeled Search links, explicit submit, Clear search, Lifecycle select and Clear filters controls issue real queries. Committed changes immediately abort/invalidate previous generations and reset the cursor/history scope. Transient errors retain typed search drafts and prior rows; retry uses the committed query. Loading/error states do not assert a successful empty collection. Required no-matching and deleted-empty copy is exact. Existing viewer read-only presentation and real creation anchors/flows remain intact.
- `ZLinkListQuery` in the owned canonical identity schema is consumed by the authored identity contract. Both OpenAPI copies and generated Go transport were regenerated together. Search limits, strict fields, closed states, cursor errors and existing rate-limit metadata are documented. Collection success 8 MiB, other success 64 KiB and error 8 KiB limits and all their tests are unchanged.

## Task Commits

1. Task 1 — Specify scoped literal search and lifecycle filters: `8c25dc7` (`test`).
2. Task 2 — Deliver scoped literal link search and lifecycle filters: `8dd9ef3` (`feat`).

## Verification

- Meaningful registered real-PG HTTP RED: `TestProductActualHTTP/link-search` expected 200 for literal `%` with nondeleted state and received 400 before implementation. Actual browser RED first rendered 26 real rows and navigated to the second page, then timed out waiting for the absent Search links label: **0 expected / 1 unexpected / 0 skipped / 0 flaky / 0 infrastructure errors**. The original root gate failed closed; a protected temporary copy outside the repository captured the same report without changing the canonical runner or validator.
- Actual production-router PostgreSQL GREEN covers literal percent, underscore, backslash and quoted SQL metacharacters; key/title/destination matches; case-insensitive matching; all five lifecycle filters; no matches; two-workspace row isolation and foreign-key search exclusion; normalized search continuation; changed exact search/lifecycle and foreign cursor denial; invalid/duplicate/foreign sort fields; invalid UTF-8/NUL; 200 versus 201 Unicode characters; viewer reads; and fresh membership removal with an existing valid cursor. The existing full pagination library cases remain intact.
- Boundary review added actual HTTP RED for 201 whitespace characters: expected 400, received 200 because trimming preceded validation. The corrected raw-before-trim validation passed the same registered cases and the subsequent full integration gate.
- Focused actual browser GREEN: **1 completed / 0 skipped / 0 flaky / 0 unexpected / 0 infrastructure errors**. Assertions compare the actual returned title set for literal `%` and `_`, reset a second-page cursor, retain drafts/rows on 503, recover through retry/clear controls, exercise all native lifecycle values and exact deleted/no-match copy, and hold an actual older backend response until a newer search succeeds to prove stale completion cannot repaint rows.
- Full root `bun run test:unit` passed: **95 discovered race-enabled Go unit tests**, all workspace/script/tool suites. Direct frontend units: **22 passed / 0 failed** across three files, preserving all 13 API boundary tests and adding exact Go-aligned Unicode whitespace normalization proof.
- Full root `bun run test:integration` passed: **17 registered top-level tests across six groups**. The previously deferred telemetry transient did not recur; no unrelated fix is claimed.
- Final original canonical `bun run test:e2e -- --project=local` passed: **30 completed**, with unchanged strict zero-skip/flaky/infrastructure report validation. This includes all prior 29 browser cases and the real committed creation CTA flow.
- Root format, lint, typecheck and generation drift gates passed. Lint required only introduced parser/shadow/whitespace corrections, retaining all configured checks. Migration verification passed existing latest version **7** and prefix-upgrade checks.
- The required final root production build after the last browser run passed all workspace/backend role builds. The unchanged exported `cleanFrontendBuild` validated lstat/realpath and the production prerender manifest before removing disposable frontend output.
- `CI=true bun run scan` passed: Bun dependencies clean, no vulnerable imported Go packages across all roles/tests, complete worktree and full Git history secrets clean. **GO-2026-5932** remains visibly reported as the existing unused `golang.org/x/crypto@v0.57.0` inventory advisory; no exception, suppression or gate weakening was introduced.
- Independent fresh read-only source review found no important bugs. Reviewer ran no builds/tests/servers and edited no files.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Align raw Unicode search bounds before normalization**
- Found during: Task 2 boundary review.
- Issue: 201 raw whitespace characters were accepted after trimming despite the canonical raw 200-character contract.
- Fix: Validate UTF-8/NUL and raw rune count before trimming; extend actual registered HTTP boundary proof.
- Files: `apps/backend/internal/service/link.go`, `apps/backend/internal/handler/product_test.go`.
- Verification: Observed HTTP RED, then targeted and full actual integration GREEN.
- Commit: `8dd9ef3`.

**2. [Rule 3 - Blocking] Narrow normalization proof and quality wiring**
- Found during: Task 2 implementation and unchanged lint gates.
- Issue: JavaScript trim differs from Go Unicode whitespace, the adjacent cursor comment still claimed filtering unsupported, and the extended parser exceeded the existing cognitive complexity limit.
- Fix: Use Unicode White_Space normalization and add its test in existing `apps/frontend/lib/links.test.ts`; correct only the comment in `apps/backend/internal/service/cursor.go`; extract the unchanged bounded limit parser within the owned handler. Correct introduced test variable shadows and a trailing blank line.
- Additional paths beyond the twelve declared files: only the existing helper test and cursor comment (two paths), within the orchestrator-authorized narrow wiring scope. No aggregates, runtime pins, gate files, migrations or later-plan features changed.
- Verification: 22 frontend units, root lint/format/type, 17 registered integration tests across six groups, focused and full actual browser gates.
- Commit: `8dd9ef3`.

Total deviations: two auto-fixed issues. Scope remains the bounded plan-15 search/filter slice.

## Known Stubs

None introduced. The existing truthful “Redirects and analytics are not available yet” notice describes intentionally later roadmap capabilities and does not block the search/filter goal. No placeholder data, destination fetch or mutation path was added.

## Issues Encountered

No unresolved implementation issue or authentication gate. Private failure reports stayed in protected OS temporary storage outside the repository. Existing provider verification/recovery/social/session acceptance remains pending plan 02-35; local signed fixtures do not satisfy that evidence. Full phase requirements remain uncompleted until all dependent refinements and final acceptance are delivered.

## Next Phase Readiness

Ready for plan 02-16. No plan 16 work was started. Search/filter is complete locally; editing, lifecycle mutations, custom domains, redirects, analytics and billing remain outside this slice.

## Self-Check: PASSED

- All fourteen listed modified paths exist.
- Task commits `8c25dc7` and `8dd9ef3` exist.
- All required execution gates passed with positive actual counts.
- No tracked file deletion or unexpected untracked repository file was introduced.
