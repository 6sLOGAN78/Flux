---
phase: 02-tenant-safe-link-control-plane
plan: "30"
subsystem: testing
tags: [tenancy, postgres, redis, authorization, privacy, playwright, concurrency]
requires:
  - phase: 02-29
    provides: Delivered scoped control-plane mutations, immutable audit and recovery boundaries
provides:
  - Registered real PostgreSQL and HTTP tenant-denial coverage with composite FK and concurrent actor-removal proofs
  - Actual Redis forged-envelope replay coverage and operational exporter privacy controls
  - Two-tab held-response proofs after workspace switch, removal and native signout
affects: [02-31, 02-32, 02-34, 02-35]
tech-stack:
  added: []
  patterns: [positive scoped fixture controls, immutable initial-route metadata classification, fresh authorized selector verification]
key-files:
  created:
    - apps/backend/internal/handler/tenant_isolation_test.go
    - apps/backend/internal/observability/privacy_test.go
    - apps/frontend/tests/tenant-denial.spec.ts
    - apps/frontend/tests/privacy.spec.ts
  modified:
    - apps/backend/internal/app/invitation_worker_test.go
    - scripts/check.ts
key-decisions:
  - Verify delivered security without fabricating a failing business test or changing secure production behavior.
  - Distinguish exact unchanged client-request route metadata and fresh authorized workspace choices from stale protected DTO data.
patterns-established:
  - New container-backed top-level suites are explicitly registered and still subject to native compiled test discovery.
requirements-completed: []
requirement-slices: [TEN-01, TEN-02, TEN-03, TEN-04, TEN-05, TEN-06, TEN-07, TEN-08, LINK-01, LINK-02, LINK-03, LINK-05, LINK-06, LINK-07, LINK-08, LINK-09, SAFE-02, SAFE-03, SAFE-04]
duration: 74min 13s
duration-basis: Sum of two reliably captured active windows; unknown pre-quota elapsed, quota downtime and metadata close-out excluded
completed: 2026-10-12
---

# Phase 2 Plan 30: Tenant Boundary Verification Summary

**Real PostgreSQL, HTTP, Redis, exporter and two-tab browser proofs verify delivered tenant isolation, fresh mutation authority and private-data disposal.**

## Performance

- Measured windows: **2026-10-11T21:02:55Z–21:13:39Z**, 10min 44s, and **2026-10-12T13:07:00Z–14:10:29Z**, 63min 29s. Total **74min 13s**. Unknown elapsed, interruption downtime and metadata work are excluded.
- Tasks: **2/2**. Source/test paths changed: **6**, all declared. PLAN classification and execution metadata are separate.

## Accomplishments

- Registered `TestControlPlaneTenantDenials` uses migrated PostgreSQL, real signed-provider HTTP and two disjoint durable actors. Eight subtests cover missing scope, both directions of foreign workspace/link/member/invitation denial, cursor transplantation, private search, scoped ledger replay, composite foreign keys, current locked actor removal and absent future routes. Three real FK violations prove SQLSTATE **23503** with valid mandatory fields; seven owned tables have positive rows and no null workspace. The concurrent HTTP mutation waits on the real workspace lock, then observes committed membership removal and creates no ledger effect.
- Exporter coverage exercises ten private fields through real telemetry sanitization with positive span, log and metric captures. Operational HTTP logs and foreign DTO/error bodies reject private fields and identifiers. No future analytics, cache or redirect execution was invented.
- Actual Redis processes workspace-, invitation- and delivery-forged tuples, then processes the same envelopes again with distinct task IDs after completion. Provider calls remain zero and retry diagnostics disclose no envelope identifiers. Existing stale, terminal and lease recovery coverage remains intact.
- Both browser tabs first render an authorized owned row and hold a real DTO containing its ID. Native switch, membership invalidation and signout dispose old records before release; after completion, the entire body including scripts contains no old link ID, title or destination. Operational console/error captures contain no private canaries. Only explicit aborted-request failures are accepted by the held transport; other handler errors fail.
- Initial Next hydration scripts can retain the old workspace parameter supplied in the original authorized request URL. Their exact bytes are recorded before the held DTO and must remain identical or disappear; newly introduced locations fail. After switch, the old workspace ID is allowed only in one exact selector option backed by a fresh authorized `/me` response. It is absent from the active main and every other live element. Removal/signout permit no old workspace ID in live elements. This does **not** claim raw full-document erasure of initial request metadata.

## Task Commits

1. **Task 1: Specify tenant-denial and privacy proofs** — `c5493c8` (test). Own unpublished `ce5bbd3` was replaced by a narrowly authorized normal amendment to construct a synthetic field at runtime while preserving its exact effective value.
2. **Task 2: Verify forged replay and final privacy gates** — `6374814` (test).

Exact path staging and normal commits were used. No production source, architecture, runtime pin, dependency manifest, scanner rule, test deadline or provider configuration changed. No push, tracked deletion or plan 31 execution occurred. Unrelated `.serena/project.yml` remains unstaged.

## Verification

- Focused race-enabled migrated PostgreSQL suite passed **1 top-level / 8 subtests**. Focused actual Redis forged/replay case passed **1 top-level / 1 subtest**. Exporter privacy passed with **10 field cases** and positive captures; its final runtime construction was rechecked afterward.
- Original targeted browser gate passed **4** tenant-denial cases. Two original canonical full browser runs passed **57/57**, with no skipped, unexpected, flaky or global errors and unchanged deadlines/retries.
- Original unit gates passed **109 selected race-enabled Go tests**, **31 frontend**, **6 email**, **15 OpenAPI**, and schema/script/tool suites. Final original integration passed **21 registered top-level tests across 6 groups**. Compiled discovery and positive native inventories remain enforced.
- Original formatting, strict lint, types, production builds and migration verification passed. Following the final synthetic-field representation change, affected original formatting/lint and focused exporter privacy passed again. No frontend or production behavior changed afterward.
- `generate:check` verified all **five artifacts**: both OpenAPI copies, Go transport and both email templates; generated sources remain unchanged.
- Production frontend output was rebuilt successfully **after** browser canaries. Existing `cleanFrontendBuild` validated the owned nonsymlink production manifest before deletion; the output directory is absent.
- Final original **CI=true scan passed** imported Go vulnerabilities across all roles/tests, Bun dependencies, complete worktree secrets and full Git history secrets. Unused **GO-2026-5932** stays visible as inventory; no exception was added. The full chain's initial scan failure and subsequent correction are described below rather than claiming an uninterrupted full-command pass.
- `git diff --check` passed. Owned private diagnostic files were removed. No hosted CI result is claimed here.

## Decisions Made

This verification-only slice found no production business/security defect. Passing new security tests represent new coverage; fixture, typing and infrastructure failures are not business RED, and no artificial RED/GREEN production commits were created.

Scope registration preserves compiled discovery. Exact initial request metadata and currently authorized choices are classified explicitly, while all protected returned DTO fields remain subject to full-document checks. Workspace switching broadcasts invalidation, so direct authorized entry is retained to establish two simultaneously active held views.

## Deviations from Plan

**[Rule 3 - Blocking] Correct local verification fixtures and representations within declared test ownership.** Foreign actors require a durable seeded profile; held routes must target the old workspace exactly; second-tab chooser actions wait for committed navigation. The existing renewed 30-day Client-cookie model avoids the pinned SDK's top-level cookie touch while preserving bounded session/JWT and fresh token refresh. Owned secondary cleanup closes only after tracked work completes; primary teardown and deadlines are unchanged. Typed seeded records and a runtime-constructed synthetic field preserve strict checking and scanner policy. No additional paths or architecture were required.

Broad requirements were retained as requirement slices, with phase-wide completion pending plan 35.

## Issues Encountered

- After three fixture correction cycles, read-only native reports and pinned SDK/source analysis identified top-level `/v1/client/touch` returning fixture JSON. A prior claim that every final assertion had completed was retracted: a `finally` marker had only shown teardown entry, and its navigation error masked the earlier timeout. Safe document-category evidence established the actual cause before the bounded fixture correction.
- Full-body checks identified initial request-route metadata and a fresh authorized selector option, never protected link DTO fields. Exact location/source verification led to the precise assertions above; no blanket script exception was added.
- Strict TypeScript caught an untyped array in the held callback; canonical DTO types and guarded seeded records resolved it. Mechanical fixture syntax and line-length errors were corrected without rule changes.
- A cold local cache prevented the unchanged collector validation from finishing its exact image pull within its 20-second context. Preparing the immutable `tools.lock.json` image resolved it; native collector/unit/integration proofs then passed. This was infrastructure setup, not a rejected collector configuration or production RED.
- Secret scanning correctly flagged synthetic source adjacency in the new privacy fixture, including the unpublished commit. Runtime construction preserves the exact effective dirty field; only the owned latest unpublished commit was amended, with normal hooks. Worktree and full history then passed unchanged rules. Nearby synthetic source values were accidentally printed during diagnosis; no provider credentials were read or exposed, and the values were not repeated afterward.
- An earlier unpinned `bunx` formatter invocation downloaded cache data only; no manifest or lock mutation occurred. All subsequent formatting used the pinned tool.

## Known Stubs

None introduced. Synthetic provider transport and exporter capture seams are test fixtures, not production substitutes or live acceptance evidence.

## User Setup Required

None for this slice. Real provider factors/recovery/OAuth and recipient delivery acceptance remain pending plan 02-35. No live configuration, secrets or email were accessed.

## Next Phase Readiness

Plan 30 verification is complete. Phase-wide requirements remain open, and plan 31 has not started.

## Self-Check: PASSED

All six source/test paths exist and both task commits were verified as Git commit objects. Gate evidence, private diagnostic cleanup, no tracked deletions and preservation of unrelated changes were checked before SDK close-out.
