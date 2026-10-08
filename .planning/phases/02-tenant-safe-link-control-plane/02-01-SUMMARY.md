---
phase: 02-tenant-safe-link-control-plane
plan: "01"
subsystem: auth
tags: [nextjs, clerk, playwright, bearer, privacy]
requires:
  - phase: 01-production-ready-service-foundation
    provides: Pinned runtimes, explicit root quality gates, Bun tests and dependency scanners
provides:
  - Native Clerk sign-in/sign-up, password recovery and verification UI with safe account state
  - Real browser stage with strict nonzero, completed, non-skipped and nonflaky result validation
  - Fixed-origin token-injected Go API fetch adapter with bounded private failures and cancellation
affects: [02-02, 02-03, 02-35]
tech-stack:
  added: [next@16.4.0, "@clerk/nextjs@7.9.12", "@clerk/testing@2.2.44", "@playwright/test@1.64.0"]
  patterns: [Provider-native authentication, test-only provider transport interception, explicit per-request bearer]
key-files:
  created:
    - apps/frontend/package.json
    - apps/frontend/app/layout.tsx
    - apps/frontend/app/sign-in/[[...sign-in]]/page.tsx
    - apps/frontend/app/sign-up/[[...sign-up]]/page.tsx
    - apps/frontend/lib/api.ts
    - apps/frontend/lib/api.test.ts
    - apps/frontend/tests/control-plane.spec.ts
    - apps/frontend/playwright.config.ts
    - apps/frontend/next.config.ts
    - apps/frontend/tsconfig.json
  modified: [package.json, bun.lock, scripts/check.ts, scripts/check.test.ts]
key-decisions:
  - "Local browser evidence uses real Clerk SDK/UI with transport interception only in tests; live verification, recovery and OAuth acceptance remain pending final plan 02-35."
  - "Reuse already-locked @types/node 22.19.19 for Next preflight, preserve runtime and React graphs, and keep generated Next outputs outside tracked source."
requirements-completed: []
requirements-addressed: [TEN-01, SAFE-02, SAFE-03]
duration: 24min
completed: 2026-10-08
---

# Phase 2 Plan 1: Native Authentication and Browser Gate Summary

**Next/Clerk native authentication routes, a strict executed-browser gate, and a bounded bearer-only API adapter establish the first runnable frontend capability.**

## Performance

- Started: 2026-10-08T12:15:00Z
- Completed: 2026-10-08T12:39:00Z
- Tasks: 2
- Product/test/config files created or modified: 14

## Accomplishments

- Original Flux wrappers expose provider-native email/password, Google/GitHub, recovery and mandatory email-verification controls. Signed-in provider state exposes account controls without returning workspace data before Go identity mapping.
- Missing provider configuration fails closed with a safe retry notice. Provider failure/loading controls retain safe original copy. Initial routes are `/sign-in` and `/sign-up`; the protected root/shared CSS remain assigned to session refinement.
- The API factory accepts the actual Clerk `getToken` callback and same-origin origin. Each invocation obtains a fresh in-memory bearer; fetch omits cookies, cache, redirects and referrer. Strict path handling, 64 KiB response bounds, safe errors, timeout and caller cancellation are implemented, with no automatic mutation retry.
- Browser provisioning and argument forwarding are connected to the root gate. JSON aggregate counts and every individual result must prove successful completed execution; empty, missing, skipped, flaky, failed and retry reports fail closed. Existing Bun and Go guards remain active.

## Task Commits

1. Task 1 specification and actual rendered RED: `736dfe9` — `test(02-01): specify native sign-in and enforce real browser results`
2. Task 2 API behavioral RED: `a7ae0d7` — `test(02-01): specify bearer adapter privacy and cancellation boundaries`
3. Task 2 implementation GREEN: `d3ed555` — `feat(02-01): deliver native Clerk auth and private bearer API adapter`

## Verification

- Browser RED: one executed assertion failed because the original sign-in heading was absent, with no infrastructure errors. A minimal real Next page served the assertion.
- API RED: seven executable tests, five meaningful failures against the initial nonfunctional API factory, before implementation.
- `bun run test:e2e -- --project=local --grep sign-in`: passed all **7** actual browser cases, no skips/retries/flakes. Covers native sign-in/social controls, password/recovery reachability, signup, email-verification step without creating a session, and 320/768/1280px keyboard/reduced-motion/overflow checks.
- API unit suite: **7 passed**, covering credential isolation, invalid targets, provider failures, safe HTTP errors, bounded/malformed bodies, deadline and cancellation.
- Browser runner regressions: **2 passed**, including actual Playwright leaf-suite shape and selector forwarding. Complete direct scripts suite: **52 passed**, zero failures.
- `bun run test:unit`: passed **21 test stages**, including the compiled Go manifest and **77 race-enabled Go unit tests**, all workspace Bun tests, script tests and tool self-test. An initial script-stage failure did not reproduce in the direct suite or complete root rerun; no gate was relaxed.
- `bun run build`: passed all frontend/shared package and backend builds.
- Full root `bun run typecheck` and final frontend typecheck/lint/format checks passed; script lint/format and `git diff --check` passed.
- Both normal frozen installation and a **clean isolated `bun install --frozen-lockfile`** passed. Node 22.23.3/Bun 1.3.14 and the existing React graph remain intact.
- Dependency scan: Bun graph clean; all Go roles/tests have no vulnerable imported packages. Existing unused `golang.org/x/crypto` inventory advisory GO-2026-5932 remains visible without suppression.
- `bun run generate:check`: passed; no canonical contracts/generated outputs were modified by this slice.

## Decisions Made

Test fixtures live only in Playwright code and intercept the provider SDK transport. Real Clerk JavaScript/UI assets are used, with exact local fixture asset versions matching the installed SDK's defaults. No application test-login route, production auth bypass, real provider secret, workspace data fixture or fake Go identity endpoint was introduced. The exported provider transport accepts a session-client input for the next plan's signed-bearer Go proof.

No full Phase 2 requirement is marked complete here: durable user mapping, session/authorization refinements and live provider acceptance span subsequent plans.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Serve real RED through a minimal Next scaffold**
- A package/config without an app cannot serve an actual rendered failing assertion. The already-owned layout, sign-in page and Next config were scaffolded in Task 1, then implemented in Task 2.
- Verification: actual missing-heading RED, followed by seven successful browser cases.
- Commits: `736dfe9`, `d3ed555`.

**2. [Rule 3 - Blocking] Resolve Next TypeScript preflight and generated outputs**
- Next attempted an automatic npm dependency install because frontend `@types/node` was not visible. Explicitly reused existing locked `@types/node@22.19.19` through Bun; no substitute library or new package identity was installed. Frozen installation and builds passed.
- Next's automatic AGENTS generation is disabled through supported `agentRules:false`. Build output uses the existing ignored `build` path; local Git excludes cover generated `next-env.d.ts` and `.next` events. Generated AGENTS output created during setup was removed. Frontend source checks use nonmutating stdin checks because the root Biome include list does not yet include the frontend and is outside ownership.
- Commits: `736dfe9`, `d3ed555`.

**3. [Rule 1 - Bug] Honor actual Playwright report shape**
- Playwright omits `suites` for leaf suites. The gate accepts an absent child array while still requiring positive aggregate counts and validating every individual completed result; regression coverage prevents an empty report from passing.
- Files: `scripts/check.ts`, `scripts/check.test.ts`.
- Commit: `d3ed555`.

## External Acceptance Still Pending

Final plan 02-35 must privately configure/verify the real Clerk instance and prove real email delivery/verification, recovery completion/session behavior, Google/GitHub OAuth and production cookie behavior. Local transport tests prove reachable native UI and failure boundaries; they are **not** evidence of completed real provider factors. No credentials were requested or printed. Production remains fail closed if configuration is absent.

## Known Stubs

None in production. Empty sessions/settings and synthetic transport responses occur only in the explicitly test-only provider fixture. The protected root/shared CSS and Go mapping are intentionally owned by following plans, rather than represented as implemented paths here.

## Self-Check: PASSED

All 14 owned product/test/config paths exist. Task commits `736dfe9`, `a7ae0d7` and `d3ed555` exist. No tracked files were deleted; generated outputs and private diagnostics remain excluded from Git.
