---
phase: 02-tenant-safe-link-control-plane
plan: "24"
subsystem: invitations
tags: [go, postgres, clerk, nextjs, privacy, transactions, tdd]
requires:
  - phase: 02-23
    provides: Durable invitation delivery with honest worker readiness and shutdown
provides:
  - Fresh verified-recipient invitation inspection without membership effects
  - Explicit atomic acceptance preserving existing roles and denying removed-actor replay
  - Fragment-free memory-only invitation arrival with native modal sign-in
affects: [02-25, 02-26, 02-30, 02-35]
tech-stack:
  added: []
  patterns: [workspace-actor-invitation locks, post-profile wall-clock expiry, memory-only arrival]
key-files:
  created:
    - apps/frontend/app/invitations/page.tsx
    - apps/frontend/lib/invitation.ts
  modified:
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/repository/delivery.go
    - apps/backend/internal/repository/invitation.go
    - apps/backend/internal/router/router.go
    - apps/backend/internal/service/invitation.go
    - apps/backend/internal/testing/browser_fixture_test.go
    - apps/backend/internal/transport/contract_test.go
    - apps/backend/internal/transport/health.gen.go
    - apps/backend/static/openapi.json
    - apps/frontend/app/layout.tsx
    - apps/frontend/next.config.ts
    - apps/frontend/tests/control-plane.spec.ts
    - packages/openapi/openapi.json
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/src/gen.test.ts
    - packages/zod/src/identity.ts
key-decisions:
  - Fresh trusted verified email is checked under workspace, durable actor and invitation locks; wall-clock expiry is sampled after the profile request and again at acceptance.
  - Acceptance preserves existing membership roles and atomically records accepted actor, preference, audit and terminal delivery ciphertext erasure.
  - Arrival secrets stay in disposable browser memory; local native modal proof models a freshly renewed Client independently of session expiry.
requirements-completed: []
requirement-slices: [TEN-03, TEN-04, TEN-05, TEN-06, TEN-07, TEN-08, SAFE-02, SAFE-03]
duration: 16min 52s
duration-basis: Resumed measured verification window from first native diagnostic through production commit self-check; prior interrupted implementation and restart downtime were not measured and are excluded
completed: 2026-10-11
---

# Phase 2 Plan 24: Explicit Verified Invitation Acceptance Summary

**Authenticated invitees review only invitations matching a fresh verified provider email, then explicitly join through one atomic PostgreSQL effect without creating an unrelated workspace.**

## Performance

- Resumed measured window: **16min 52s**, **2026-10-11T05:30:13Z** through **2026-10-11T05:47:05Z**. The existing RED commit and interrupted implementation were preserved; this duration does not claim their unmeasured execution time.
- Tasks complete: **2/2**.
- Source/test/artifact paths: **19**, including two created frontend files. Six necessary concrete additions are recorded in the updated plan; its scope notes are metadata, outside this source count.

## Accomplishments

- Canonical Zod and ts-rest authority now describe authenticated body-only POST inspection and acceptance. Both OpenAPI copies and the generated Go transport were regenerated together. Strict bounded JSON, unknown-field/trailing/query rejection, existing bearer/Origin/JSON protections, no-store and safe errors remain enforced.
- Preliminary digest lookup exposes no invitation details; recipient projection follows a fresh configured trusted provider primary verified email check. Wrong recipients, malformed/foreign credentials, expired/revoked invitations, unavailable or unverified profiles, different accepted actors and removed accepted actors are denied safely.
- Workspace → durable actor → invitation exclusive locks serialize membership effects with competing workspace mutations. Pending invitations require the allowed non-owner role and seven-day expiry. Acceptance inserts membership only if absent, preserves existing roles, marks the exact accepted actor, records immutable audit, saves workspace preference and erases retained delivery ciphertext/reconciliation flags in the same transaction. Same-actor repetition requires current membership and creates no second membership or audit event.
- A provider-latency regression verifies expiry after the fresh profile lookup using PostgreSQL wall clock; the effect also rechecks expiry before updating acceptance. Deferred commit failure rolls back membership, invitation status, preference, audit and ciphertext cleanup together.
- The head bootstrap captures a bounded arrival fragment and removes it with replaceState before provider execution. The page consumes that closure into module memory and renders fresh session-bound inspection. Secrets never enter storage, cookies, HTML, request URLs or referrers. Only an explicit Accept button performs the effect and opens the invited Links page.
- Registered real PostgreSQL HTTP subcases cover expiry/revocation, deferred commit rollback, profile-delay expiry, winning workspace-locked expiry/revoke against waiting acceptance, eight concurrent accepts with one membership/audit, Failed ciphertext cleanup, existing admin role preservation, strict input and safe foreign denial.
- Native browser cases prove signed-in review, no membership on inspection, no unrelated workspace creation, committed preference, privacy, explicit acceptance, in-place modal password sign-in and wrong-recipient denial.

## Task Commits

1. **Task 1: Specify inspect and explicitly accept a verified-recipient invitation** — `a2e06f4` (existing RED, preserved).
2. **Task 2: Deliver inspect and explicitly accept a verified-recipient invitation** — `fd7897c` (GREEN).

No hook bypass, blanket staging, reset or tracked deletion was used.

## Verification

- Existing meaningful RED: real PostgreSQL HTTP lacked the inspection route, and the native happy path lacked the invitation heading. Environment/fixture failures were excluded from RED.
- Resumed `go vet ./...` passed; the previously interrupted vet failure did not reproduce, so no historical cause is claimed.
- Focused `-race -count=1` registered invitation acceptance HTTP subtest passed against real migrated PostgreSQL.
- Final root formatting, strict lint and typechecking passed. Final lint was repeated after the introducing fixture/inventory changes.
- Canonical root unit gate passed **106 selected race-enabled Go units**, **28 frontend tests**, **6 email tests**, canonical package suites and all native script/tool tests. Exact compiled Go inventory enforcement remains enabled.
- Canonical integration gate passed **20 registered tests across six groups**, including actual HTTP/process/worker/telemetry/resource boundaries.
- Targeted native invitation browser gate passed **2/2**. The original full native browser gate passed **42/42 on its first full run**, preserving all previous 40 cases. Zero retries, skips, flaky results or fixture errors were accepted.
- Root production build and migration checks passed. Five-artifact generation reproducibility passed; no additional generator outputs were changed.
- The existing `cleanFrontendBuild` guard validated real nonsymlink ownership and production prerender manifest v4 before disposing the generated frontend output.
- **CI=true full local scan passed**: no vulnerable imported Go package across all roles/tests, clean Bun dependencies, clean worktree and full-history secret scans. Unused GO-2026-5932 remains visible as an inventory advisory; no exception was added. This is local CI-mode evidence.

## Deviations from Plan

1. **[Rule 2 - Missing critical wiring/privacy]** Added route registration in `apps/backend/internal/router/router.go`, early fragment capture in `apps/frontend/app/layout.tsx`, and arrival no-store/no-referrer headers in `apps/frontend/next.config.ts`. These are necessary to make the planned authenticated arrival reachable and private.
2. **[Rule 3 - Blocking fixture/contract compatibility]** Added the actual SQL arrival seed in `apps/backend/internal/testing/browser_fixture_test.go` during RED; retained its narrowly explained synthetic-SQL lint directive. Canonical regeneration renamed the existing Workspace role constant, requiring the same-value adaptation in `apps/backend/internal/transport/contract_test.go`. Added the two implemented paths to the exact expected inventory in `packages/openapi/src/gen.test.ts`; every previous path/assertion remains.
3. **[Rule 3 - Blocking SDK fixture protocol]** In the already-owned browser specification, return a Session resource for session touch while retaining the piggyback Client. A read-only request trace and the exact official Clerk JS 6.38.1 Session source established that the former Client response overwrote Session fields. A second finite-label capture established a distinct full-page Client touch navigation: the fixture conflated a ten-minute Client cookie with session lifetime. The completed test Client now explicitly models a renewed **30-day controlled cookie**, preserving the original session/JWT expiry/data. This is not a claim of universal Clerk lifetime. Native UI/password submission remains real SDK execution; exactly one main-document navigation before acceptance proves the selected in-place continuation. No production auth override or persistence was introduced.
4. **[Rule 1 - Bug]** Inspection initially sampled time before the fresh profile latency and could reveal an invitation that expired during that request. The controlled actual-PG regression observed 200 before the correction and now requires 404 with no workspace details. `apps/backend/internal/repository/invitation.go` samples PostgreSQL wall clock after the provider lookup and preserves the separate effect-time expiry guard; `apps/backend/internal/handler/product_test.go` retains the delayed-provider regression.

All GREEN deviations are included in `fd7897c`; the original browser seed is also in `a2e06f4`. All concrete additions are recorded in PLAN.md.

## Issues Encountered

Interrupted modal fixture fix attempts were preserved and not repeated blindly. Private native-runner copies retained failed output in a mode-0700 OS temporary directory with mode-0600 reports; original runner startup, limits, parsing and teardown were unchanged. Temporary finite-label instrumentation was removed. The earlier missing-button observation did not reproduce: the initial page correctly retained the arrival and rendered Sign in.

A Session touch resource correction alone was insufficient. The final SDK capture showed the browser at the provider Client-touch path without an application SDK instance. Exact official pinned SDK source separately identified the cookie-renewal navigation. The renewed Client fixture corrected that proven protocol state; no speculative production workaround was added.

The first full unit gate found only the stale exact OpenAPI path inventory; its two new entries were added and the complete native unit gate passed.

## Known Stubs

None in the delivered inspection/acceptance path. Empty initial UI state is replaced by real authenticated API data or explicit safe failure.

## Remaining Acceptance and Next Plan Readiness

Ready for **02-25**; it was not started. Broad requirements remain pending, so `requirements-completed` is intentionally empty.

The local fixture proves the selected freshly renewed Client's in-place modal path. **It does not prove that invitation memory survives a full document navigation.** Memory is disposable and a fresh document without an arrival has no secret to inspect or accept; stale-client Client-touch and other recovery edges remain for **02-26**. Do not turn this into cookie/storage token persistence or an arbitrary return URL.

Actual configured provider verification/recovery/social/session/cookie/delivery acceptance remains the final **02-35** gate. No real provider factors or email sends were used or claimed.

## Self-Check: PASSED

- Both created frontend files exist; all 19 source/test/artifact paths are represented by the preserved RED and GREEN commits.
- Commits `a2e06f4` and `fd7897c` exist. GREEN contains no tracked deletions.
- Required focused and complete native gates passed with positive executed counts as recorded above.
- All temporary source diagnostics were removed; generated outputs were validated and safely cleaned.
- Unrelated pre-existing `.serena/project.yml` remains unstaged and untouched.

