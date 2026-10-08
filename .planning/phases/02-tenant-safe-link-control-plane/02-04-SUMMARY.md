---
phase: 02-tenant-safe-link-control-plane
plan: "04"
subsystem: auth
tags: [clerk, csrf, cors, bearer, playwright, nextjs, postgresql]
requires:
  - phase: 02-03
    provides: Committed durable identity and registered signed real-PostgreSQL browser/HTTP fixtures
provides:
  - Installed bearer-only Actor boundary for versioned product requests
  - Exact browser mutation origin, JSON and 64KiB body enforcement
  - Private no-store responses, correlation and safe recovery metadata
  - Canonical authenticated account screen with cancellation, recovery and sign-out clearing
  - Native CSS tokens wired through the root layout
affects: [02-05, 02-06, 02-09, 02-30, 02-35]
tech-stack:
  added: []
  patterns: [Verified Actor middleware, Same-origin memory-only tokens, Abort and session-bound identity rendering]
key-files:
  created:
    - apps/backend/internal/middleware/auth_test.go
    - apps/backend/internal/middleware/origin.go
    - apps/frontend/app/page.tsx
    - apps/frontend/app/globals.css
  modified:
    - apps/backend/internal/middleware/auth.go
    - apps/backend/internal/middleware/global.go
    - apps/backend/internal/router/router.go
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/frontend/app/layout.tsx
    - apps/frontend/lib/api.ts
    - apps/frontend/lib/api.test.ts
    - apps/frontend/tests/control-plane.spec.ts
    - apps/frontend/next.config.ts
    - apps/frontend/package.json
    - bun.lock
key-decisions:
  - "Protect the installed versioned router boundary before future mutations become reachable; a test-only mutation probe verifies enforcement without inventing a production endpoint."
  - "Use the canonical @flux/zod workspace export for account response validation, and suppress canceled or previous-session results before rendering identity."
requirements-completed: []
requirements-addressed: [TEN-01, SAFE-02, SAFE-03]
duration: 24min
completed: 2026-10-08
---

# Phase 2 Plan 4: Browser Session Boundaries Summary

**Versioned product requests now install verified Actors, reject unsafe browser mutations, and expose a native account screen with bounded token requests, safe session recovery and private-data clearing.**

## Performance

- Resumed execution after the predecessor was interrupted; preserved its uncommitted tests.
- Duration: approximately 24 minutes of resumed work; restart overhead and predecessor work are not separately measured.
- Tasks: 2/2.
- Implementation/test/config ownership: 16 concrete files across task commits (11 declared files actually touched plus five necessary narrow wiring/test files).

## Accomplishments

- Replaced legacy header authorization, provider organization roles/permissions and identity logging with explicitly injected, verified Actor middleware. `/me` consumes that Actor and continues resolving committed internal UUIDs through the existing identity service. Provider organization claims never grant Flux workspace authority.
- Installed protection across the existing `/api/v1/` router boundary. Browser mutations reject missing, null, foreign and suffix-spoofed origins; require a single JSON content type and explicit bearer; and bound the complete body to 64KiB before reaching handlers. Only local development origins may use HTTP. Preflights use exact origin comparison with explicit methods/headers and no credentialed wildcard allowance.
- All versioned private responses, including early preflight/rate-limit errors, use `Cache-Control: no-store` and retain request correlation. Provider unavailability returns `Retry-After: 1`; rate limiting returns `Retry-After: 1` and `X-RateLimit-Limit: 20`. Header spelling is case-insensitive; these indicate the current process-local limiter and do not claim distributed quota enforcement.
- Preserved the previously implemented fixed HTTPS Clerk SDK endpoints, no proxy inheritance, explicit client injection and two-second provider deadline. No provider transport override, alternate key or auth bypass was introduced into production.
- Added the actual root account UI using `useAuth().getToken()` and canonical `ZIdentityResponse` validation from the normal `@flux/zod` export. Tokens stay within each request invocation; requests omit ambient credentials, disable caching, reject redirects and use the fixed same-origin rewrite. HTTP 429 and server failures map to safe recoverable errors without raw response bodies.
- Account requests cancel on teardown/session change. Ready identity is tagged with its session so another session cannot render old data. Sign-out immediately aborts requests and removes private state; an unsuccessful provider sign-out remains locally cleared and offers an explicit retry. Expiry and unavailability use the approved safe copy. No workspace, redirect or analytics capability is fabricated.
- Wired the native CSS token stylesheet into the root layout, including typography, original neutral/blue surfaces, 44px controls, visible focus, 320px reflow, reduced motion and forced colors. Existing provider-native sign-in/sign-up/recovery screens remain intact.
- Validated the fixed build-time API rewrite: HTTPS or loopback HTTP only, without credentials, path, query or fragment. Malformed configuration produces a closed diagnostic. The root and rewritten API paths receive no-store headers.

## Task Commits

1. Task 1 RED: `9ee1346` — `test(02-04): specify browser session boundaries and safe recovery`.
2. Task 2 GREEN: `071552a` — `feat(02-04): enforce browser session boundaries and recover safely`.

## Verification

- RED reproduced four exact-preflight denials returning 204 instead of 403, real migrated-PostgreSQL mutation handlers being reached without the required boundary, 429/503 client failures receiving a generic error, and a nonlocal HTTP rewrite being accepted. Cancellation during token refresh already worked and was retained.
- Actual browser RED completed five session-matching cases: two existing cases passed, three new missing-root cases failed, with zero skipped/flaky cases and zero infrastructure errors. Signed provider data and diagnostic JSON were captured only in protected OS temporary storage.
- `bun run test:unit` passed all workspace builds/tests, script/tool self-tests and **82 discovered race-enabled Go unit top-level tests**. The frontend API suite passed **10 tests**, including safe status mapping, fixed rewrite validation, fresh tokens, actual fetch cancellation, no late dispatch after token refresh cancellation, provider timeout, oversized bodies and private error suppression.
- Final race-enabled `TestProductActualHTTP` passed **26 test pass events**: the dispatcher and its **25 subtests**, retaining the identity/concurrency/rollback/privacy corpus, eight mutation denial/allowed cases, their boundary parent, and recovery/rate-header proof. Only the test router registers the mutation probe. Denials preserve no-store/correlation and do not leak the signed token; exactly one allowed mutation reaches its handler with Actor and without legacy authority.
- Exact-preflight middleware tests passed, including a valid preflight with an exact origin, no credential allowance and no-store.
- Required `bun run test:e2e -- --project=local --grep session` passed **five actual browser cases**. The complete local suite passed **13 cases**, retaining all ten prior native provider/bearer/identity flows and adding account/sign-out, 401 expiry and 503 retry. The happy path also verifies 320px page reflow and keyboard focus. The final complete run placed browser outputs inside a protected OS temporary directory.
- A recovery assertion initially matched both the application alert and Next's route announcer. Private diagnostics confirmed a strict-locator violation; scoping the assertion to the main content preserved the exact required message. No application recovery behavior or assertion was weakened.
- Final frontend production build, format, lint and typecheck passed after the final session guard and rewrite changes. Owned backend packages reported **zero lint issues**; `git diff --check` passed.
- `bun install --ignore-scripts` and `bun install --frozen-lockfile --ignore-scripts` passed for the existing local workspace dependency. Runtime, React, SDK and tool pins were unchanged.
- The root's verified frontend-output cleanup ran immediately before the final secret scan: **worktree clean; full history clean**. No scanner rules or privacy exceptions changed. No unexpected tracked-file deletions occurred.

## Decisions Made

- Boundary enforcement belongs to the installed router, so future handlers cannot accidentally omit browser/session safeguards. The probe remains test-only and uses the registered real-PG dispatcher.
- Go remains the authentication/authorization authority. A public client-rendered shell does not need a second Next authentication proxy; every private API call passes through verified Go middleware.
- Canonical workspace exports own response validation and types; no handwritten identity DTO or generated artifact changes were needed.
- Provider-default seven-day maximum session lifetime and SDK refresh remain the approved policy. The long-lived FAPI HttpOnly credential differs from the short-lived application JWT; actual Secure/SameSite/cookie attributes, production refresh/key rotation, recovery-triggered session behavior, OAuth and provider quota evidence remain plan 02-35 acceptance, not claims made by local fixtures.

## Deviations from Plan

### Necessary Narrow Adjustments

1. **[Rule 3 - Blocking] Registered actual-HTTP mutation proof:** `apps/backend/internal/handler/product_test.go` extends the existing registered real-PG dispatcher. This avoids an unregistered container suite or invented production mutation endpoint. Extracted its assertion helper to preserve the project's complexity limit.
2. **[Rule 2 - Missing Critical] Exact installed CORS enforcement:** `apps/backend/internal/middleware/global.go` owns the existing CORS boundary and safe error response; updating it was necessary to reject unsafe preflights before middleware short-circuiting and to attach recovery headers.
3. **[Rule 3 - Blocking] Actual rewrite validation:** `apps/frontend/next.config.ts` owns the runtime rewrite and response headers; policy tests against the actual configuration required this narrow change rather than an unused validator.
4. **[Rule 3 - Blocking] Canonical schema dependency:** `apps/frontend/package.json` and `bun.lock` now declare only `@flux/zod: workspace:*`. The build previously could not resolve its canonical import. Existing root dependency ordering automatically builds Zod before the frontend; no runner or recursive prebuild modification was needed.

These five additional concrete paths bring actual ownership to 16. The plan-listed config test, internal sample environment and proxy paths did not need changes: fixed provider transport/configuration was already implemented, adjacent runbook documentation belongs to plan 02-05, and Go protects the actual API. No unrelated production refactor or other plan was executed.

## Known Stubs

None in the delivered session boundary or account screen. No missing product endpoint is presented as implemented. Pending workspace/team/link capabilities remain their own plans.

## External Acceptance and Next Plan Readiness

Local fixtures exercise the real SDK/UI, signed Go authentication and migrated PostgreSQL; they cannot prove live provider factors or cookie/session policy. Final provider acceptance remains pending plan 02-35. TEN-01, SAFE-02 and SAFE-03 are addressed by this slice but intentionally remain globally incomplete until their other phase plans and required external evidence are complete. No new user setup is required for these local proofs.

The bounded browser/session foundation is ready for plan 02-05 and subsequent authoritative workspace features. No new threat surface outside this plan's browser/provider trust boundaries was introduced.

## Self-Check: PASSED

- All four created implementation/test files and this summary exist.
- RED `9ee1346` and GREEN `071552a` exist in history in the required order.
- Required local behavior is verified, no unexpected deletions or production stubs remain, and live provider limitations are explicit.
