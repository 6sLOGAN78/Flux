---
phase: 02-tenant-safe-link-control-plane
plan: "02"
subsystem: auth
tags: [clerk, jwt, bearer, echo, playwright, openapi]
requires:
  - phase: 02-01
    provides: Native Clerk browser UI, bearer API adapter and executed-browser gate
provides:
  - SDK-verified bearer-only GET /api/v1/me with exact issuer and authorized-party checks
  - Per-request active-session verification with one bounded two-second deadline
  - Signed production-router browser fixture and race-enabled claims/revocation/concurrency corpus
  - Canonical identity schemas/contracts and three regenerated Go/OpenAPI outputs
affects: [02-03, 02-04, 02-05, 02-35]
tech-stack:
  added: []
  patterns: [Explicit SDK client injection, bounded active-session lookup, test-binary-only signed transport]
key-files:
  created:
    - apps/backend/internal/service/auth_test.go
    - apps/backend/internal/handler/product.go
    - packages/zod/src/identity.ts
    - packages/openapi/src/contracts/identity.ts
  modified:
    - apps/backend/internal/service/auth.go
    - apps/backend/internal/app/api.go
    - apps/backend/internal/router/router.go
    - apps/backend/internal/config/config.go
    - packages/zod/src/index.ts
    - packages/openapi/src/contracts/index.ts
    - packages/openapi/openapi.json
    - apps/backend/static/openapi.json
    - apps/backend/internal/transport/health.gen.go
    - scripts/check.ts
    - apps/frontend/tests/control-plane.spec.ts
    - packages/openapi/src/gen.test.ts
key-decisions:
  - "Keep provider clients explicitly injected, fixed-endpoint and bounded; never initialize the global Clerk key or authorize from organization claims."
  - "Return only authenticated:true until durable identity mapping in 02-03; local signed SDK fixtures do not satisfy live provider acceptance."
requirements-completed: []
requirements-addressed: [TEN-01, SAFE-02, SAFE-03]
duration: 286min
completed: 2026-10-08
---

# Phase 2 Plan 2: Active Bearer Boundary Summary

**Clerk SDK signature verification, exact issuer/authorized party checks, and bounded current-session lookup protect the generated `/api/v1/me` endpoint through the real Go router.**

## Performance

- Started: approximately 2026-10-08T12:41:00Z
- Completed: 2026-10-08T17:27:00Z
- Elapsed: approximately 286 minutes, including a multi-hour usage interruption and resumption; this is not uninterrupted execution time.
- Tasks: 2
- Product/test/config files created or modified across task commits: 16 (14 declared plus two authorized test adaptations)

## Accomplishments

- Removed `clerk.SetKey` from the service constructor. Production uses explicitly keyed SDK JWKS/session/user clients, a fixed provider endpoint, disabled proxy inheritance, rejected redirects and bounded HTTP operations. Config retains `FLUX_AUTH.SECRET_KEY` and adds issuer/authorized-party fields through the existing environment decoder. Missing/invalid authentication configuration fails closed without making health depend on the provider.
- `Authenticate` uses the pinned Clerk SDK cryptographic verifier and requires the exact configured issuer, an exact nonempty allowed authorized party, valid subject/session identifiers, explicit exp/nbf/iat and coherent current times. Every accepted request rereads provider session state within one two-second deadline shared with JWKS lookup. Revoked, pending, missing or mismatched sessions return safe 401; provider failures/timeouts return safe 503. Provider organization claims supply no authority.
- The explicitly injected ProductHandler exposes only GET `/api/v1/me` and serializes canonical generated `{authenticated:true}` after verification. Ambient cookies are ignored, duplicate/malformed authorization is rejected, and no-store is set before the rate limiter as well as the handler. No durable user UUID or tenant route is fabricated.
- Authored identity Zod/ts-rest modules declare the response, safe error envelope, bearer security, status coverage and no-store policy. Both OpenAPI JSON copies and the Go transport were regenerated from that authority; the existing health aliases and generation manifest remain intact.
- The external-package test fixture owns signed RSA tokens, loopback SDK/JWKS/session transport, real pinned migrated PostgreSQL and the production Echo router. Its ordinary integration invocation performs a bounded signed HTTP probe and exits; only the explicit test-binary flag holds it ready for Playwright. The root browser gate starts it before consumption, accepts only a sanitized loopback readiness URL, and performs bounded graceful/fallback cleanup. No production auth mode, login route, signing-key environment variable or bypass was introduced.

## Task Commits

1. Task 1 executable browser specification/bootstrap RED: `002237a` — `test(02-02): bootstrap signed production-router browser RED`
2. Task 2 signed actual-HTTP behavioral RED: `8490b79` — `test(02-02): specify signed bearer claims revocation and deadlines`
3. Task 2 implementation GREEN: `da0053d` — `feat(02-02): verify active bearer sessions at the Go boundary`

## Verification

- Browser RED: one executed case failed on expected 200 versus actual 404 from the running production router. Private report had zero infrastructure errors; missing fixture/constructor failures were not accepted as RED.
- HTTP RED: the signed active-session actual-HTTP subtest failed on expected 200 versus 404 before implementation.
- Race-enabled signed HTTP proof passed: three top-level unit groups, **20 named claim/session scenarios**, four additional credential subcases, immediate revocation of the same valid token and **10 concurrent per-request session checks**. Includes exact issuer/party, missing identity/time claims, invalid signature/path identifiers, expiry/future timestamps, pending/revoked/mismatched/not-found sessions, safe outage and bounded timeout behavior. Organization claims do not alter the response.
- `bun run test:e2e -- --project=local --grep bearer`: **2 actual browser cases passed**, including a real Clerk JS `session.getToken()` signed request accepted by Go and safe denial/outage/cookie-only cases. Root orchestrator performed the successful final run after fixture lifecycle repair.
- Full local browser suite: **9 passed**, preserving all seven original sign-in/native verification/accessibility cases; no retries, skips or flaky results. Root orchestrator performed this successful final run serially.
- `bun run test:unit`: **passed all 22 test stages**, including **80 race-enabled Go unit tests**, compiled manifest checks, all workspace suites, scripts and tool self-test. Final run completed with `Checks passed`.
- OpenAPI unit suite: **12 passed** with existing health privacy checks retained and `/me` bearer/status/no-store assertions added.
- Real pinned PostgreSQL `TestBearerBrowserFixture` integration passed with race detection. Existing `TestRoleHealthActualHTTP` regression passed with race detection, preserving the router's nil-service compatibility.
- Owned Go lint across service/handler/router/config/app: **0 issues**. `bun run generate:check` passed. Frontend build/typecheck, script typecheck, owned TypeScript checks and final frontend formatting/lint passed. `git diff --check` passed.
- No packages were installed or runtime/tool/database pins changed.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Supply the missing bearer browser cases**
- Found during Task 1: the existing browser spec contained seven sign-in cases and no bearer cases despite the plan assumption.
- Added narrowly necessary signed SDK/Go and denial browser cases to `apps/frontend/tests/control-plane.spec.ts`, explicitly authorized by the orchestrator; all original cases remain.
- Commits: `002237a`, `da0053d`.

**2. [Rule 3 - Blocking] Adapt the health-only OpenAPI test baseline**
- Found during Task 2: `packages/openapi/src/gen.test.ts` required exactly two health paths and globally prohibited the field name `error`, which legitimately belongs to the standard product error envelope.
- Explicitly authorized adaptation includes the new `/me` path and its security/status/cache policy; the original forbidden-diagnostics assertion now applies to the complete unchanged health schemas and health routes. No scanner or health privacy control was removed.
- Verification: all 12 OpenAPI tests and the root unit gate passed.
- Commit: `da0053d`.

## Issues Encountered

- Usage limits interrupted Task 2; both RED commits and uncommitted implementation were preserved. Execution resumed without restarting RED or discarding shared changes.
- Three fixture teardown attempts reached the executor fix limit: waiting for native controls left lazy provider chunks pending, `unrouteAll(wait)` invalidated live routes, and global networkidle timed out with Next development traffic. The executor returned a checkpoint. The root orchestrator then repaired fixture lifecycle by tracking owned provider asset promises, bounding asset fetches to 15 seconds, navigating away to stop producers, draining those promises without swallowing failures, then removing routes. Targeted and full browser gates passed after that structural repair; it is included in `da0053d`.
- An overlapping frontend unit build/browser development run collided in shared Next output. Serial frontend build and final browser/unit verification passed; no gate was relaxed.
- An earlier concurrent direct scripts run had 51/52 passing with an existing pinned OpenPGP scanner fixture timing out at Bun's five-second default. The final serial root unit gate passed unchanged. No scanner/test timeout or vulnerability classification was modified.

## External Acceptance and Next Plan Readiness

- Live Clerk factors, recovery, social sign-in, production cookie/session/key-rotation behavior and provider quota/latency remain pending plan 02-35. Local signed SDK interception proves the Go boundary; it cannot count as those external acceptance rows.
- Durable issuer/subject-to-Flux-user mapping and verified profile persistence belong to 02-03. Current `/me` deliberately proves authentication only, exactly as this plan specifies.
- TEN-01, SAFE-02 and SAFE-03 are addressed here but are not marked wholly complete: identity mapping, mutation/session policies and wider tenant isolation require subsequent plans.

## Known Stubs

None in production. The minimal authenticated response is the planned observable skeleton, not a fabricated identity. Empty profile fields/collections and synthetic provider responses are confined to the explicit test fixture. No new unplanned production trust boundary was found during the threat-surface scan.

## Self-Check: PASSED

All 16 product/test/config paths exist; all four created files exist. Commits `002237a`, `8490b79` and `da0053d` exist. No tracked files were deleted. Private diagnostics and generated runtime outputs remain ignored by Git.
