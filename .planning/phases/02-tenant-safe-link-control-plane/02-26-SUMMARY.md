---
phase: 02-tenant-safe-link-control-plane
plan: "26"
subsystem: invitations
tags: [clerk, nextjs, privacy, postgres, recovery, tdd]
requires:
  - phase: 02-25
    provides: Scoped invitation rotation and revocation
  - phase: 02-24
    provides: Fresh verified-recipient inspection and explicit atomic acceptance
provides:
  - Immediate account-switch disposal with session-bound invitation responses
  - Explicit invitation reload after provider failure without access grants
  - Executable memory-only recovery and secret-free browser continuation proof
affects: [02-30, 02-31, 02-34, 02-35]
tech-stack:
  added: []
  patterns: [disposable arrival memory, synchronous request cancellation, post-commit focus restoration]
key-files:
  created:
    - apps/frontend/lib/invitation.test.ts
    - apps/backend/internal/service/invitation_test.go
  modified:
    - apps/frontend/app/invitations/page.tsx
    - apps/frontend/components/invitation-panel.tsx
    - apps/frontend/tests/control-plane.spec.ts
    - apps/backend/internal/handler/product_test.go
    - apps/backend/internal/testing/product.go
key-decisions:
  - Invitation credentials cross only authorized inspect and explicit accept POST bodies; full navigation discards memory and requires a new invitation.
  - Switch account synchronously cancels pending work and clears authorized review before provider signout completes; late responses must match the current session.
  - Browser protocol issues fresh signed three-minute JWTs per scenario and SDK token refresh without changing session lifetime or production verification.
requirements-completed: []
requirement-slices: [TEN-03, TEN-04, TEN-05, TEN-06, TEN-07, TEN-08, SAFE-02, SAFE-03]
duration: 28min 36s
duration-basis: Measured resumed verification through GREEN commit self-check; earlier interrupted implementation, quota interruption and restart downtime are unmeasured and excluded
completed: 2026-10-11
---

# Phase 2 Plan 26: Invitation Authentication and Recovery Privacy Summary

**Invitation review stays bound to the signed-in account, account switching disposes pending access immediately, and browser navigation loses credentials safely without persistence or unintended membership effects.**

## Performance

- Resumed measured window: **28min 36s**, from **2026-10-11T13:00:33.264Z** through **2026-10-11T13:29:09.629Z**.
- Tasks: **2/2**. The existing RED commit and partial GREEN were preserved.
- Source/test paths: **7**, including **2 created files**; the PLAN clarification is separate metadata.
- Earlier executor time and quota/restart downtime are excluded.

## Accomplishments

- Switching accounts synchronously aborts invitation work, clears the old review and actions, and shows safe switching status before provider signout completes. Inspection and acceptance results are fenced by request generation and current session. Signout failure exposes approved safe copy and explicit retry without restoring old authorization.
- Provider 503 shows only the approved loading-failure copy and an explicit Reload invitation action. No review, membership or acceptance action is granted before successful inspection.
- Seven native invitation privacy cases cover provider failure/reload; held inspection versus held signout; refresh and full navigation; in-place native modal sign-in with a renewed Client; signed-in private inspection and explicit acceptance without unrelated workspace creation; accepted/terminal safe states; and the real SDK's stale-Client cookie-renewal navigation.
- Stale Client touch receives a fixed safe local 303 continuation without the credential or fragment. The fresh document has no arrival memory, sends neither inspection nor acceptance, and requests a new invitation. No storage, cookie, arbitrary return URL or production auth override was introduced.
- Unit proof checks both canonical credential POST operations use no-store, omit cookies, reject redirects and suppress referrers. Browser assertions check URL, headers, unrelated bodies, referrers, HTML, logs, history state, local/session storage and cookies for secret exposure.
- Registered real PostgreSQL HTTP cases deny inspection and acceptance for unverified, unavailable, mismatched, missing-primary and banned profiles without details or membership; removed accepted actors cannot inspect or rejoin. Existing role-preserving repetition, transactional cleanup, concurrency, expiry, revocation and one-effect audit cases remain intact.
- Invitation mutation focus now returns after committed rendering and reload, avoiding a focus target that React is about to remove. Fresh short-lived fixture JWT issuance lets every scenario and SDK refresh exercise production verification independently of total suite duration.

## Task Commits

1. **Task 1: Specify invitation privacy and account recovery boundaries** — `8557537` (preserved RED).
2. **Task 2: Protect invitation authentication and recovery boundaries** — `5cd3150` (GREEN).

No hook bypass, blanket staging or tracked deletion was used.

## Verification

- Preserved meaningful RED: provider-503 recovery lacked the required Reload invitation action. Incorrect fixture signout matching and pinned public asset failures were excluded from RED.
- Focused service race test passed after correcting only its new lint assertions.
- Final formatting, complete strict lint and typechecking passed.
- Final canonical unit gate passed **107 selected race-enabled Go tests**, **30 frontend tests**, **6 email tests**, canonical contracts and all script/tool suites. Compiled inventory enforcement remained enabled.
- Final canonical integration gate passed **20 registered tests across 6 groups**, including real PostgreSQL HTTP and dependency/process/worker/telemetry boundaries.
- Preserved pre-interruption target evidence was **7/7**. After the proven blocking repairs, the affected canonical target passed **8/8** (seven privacy cases plus invitation management).
- Final original full native browser gate passed **48/48**. Zero retries, skips, flaky results or fixture errors were accepted.
- Production build, migration checks and all **five generated-artifact** reproducibility checks passed. No contracts or generated outputs were changed.
- The existing cleanFrontendBuild guard validated owned nonsymlink production output and prerender manifest v4 before disposal.
- **CI=true full scan passed**: no vulnerable imported Go package across all roles/tests, clean Bun dependencies, clean worktree secrets and clean full-history secrets. Unused GO-2026-5932 remained visible as inventory evidence without an exception.

## Decisions Made

The plan's inspect-only secret wording is clarified to include the already-canonical explicit acceptance POST body. Both operations need the credential; no additional transport or persistence is allowed.

Memory survives the selected in-place modal path, but deliberately does not survive refresh or full document navigation. Losing memory requests a new invitation instead of persisting a credential to survive provider navigation.

Local signed provider transport models native SDK protocol behavior and does not count as actual verification, recovery, Google/GitHub or production session/cookie acceptance.

## Deviations from Plan

1. **[Rule 1 - Bug] New service-test lint violations.** Full lint found two long lines and two errors.As assertions requiring require.ErrorAs in the new test. Wrapped lines and used the mandated assertion; focused race and complete final lint/unit gates passed. File: apps/backend/internal/service/invitation_test.go. Commit: `5cd3150`.
2. **[Rule 3 - Blocking fixture protocol] Native signout transport.** The pinned SDK uses DELETE /v1/client/sessions for single-session signout. Updated the held signout predicate and existing test transport to clear the Client collection for its actual endpoint, preserving Session touch resources and native UI execution. File: apps/frontend/tests/control-plane.spec.ts. Commit: `5cd3150`.
3. **[Rule 1 - Bug] Post-revocation focus race in the affected invitation flow.** The full browser diagnostic found Reload invitations inactive. A requestAnimationFrame callback could focus the still-connected Revoke trigger before React removed it. A post-commit effect waits for busy/loading/refreshing to finish and keeps the same request-generation fence before focusing the surviving trigger or Reload. File: apps/frontend/components/invitation-panel.tsx. Target 8/8 and final full 48/48 passed. Commit: `5cd3150`.
4. **[Rule 3 - Blocking fixture lifetime] Static JWT expired during the expanded full suite.** The held protocol minted /client's JWT once with a three-minute expiry, then reused it across every scenario. The private diagnostic took 196 seconds and later API creates returned 401. Added narrowly recorded ownership of apps/backend/internal/testing/product.go to issue fresh signed JWTs per /client request; SDK token refresh requests obtain a matching-session token via the existing test-only protocol in apps/frontend/tests/control-plane.spec.ts. Original three-minute TTL, session lifetime, issuer/subject/session claims and revoked/expired/foreign denial tests are preserved. Final full browser and complete unit/integration gates passed. Commit: `5cd3150`.

The PLAN records the concrete fixture addition and credential-body clarification. No new production endpoint, schema, library, runtime flag or auth bypass was added.

## Issues Encountered

The first resumed full browser run failed behind the original runner's sanitized error wrapper. A private native runner copy retained browser stdout in a mode-0600 report within a mode-0700 OS temporary directory; only repository-root resolution and the first runCommand close callback were adapted. Fixture startup, limits, native report validation and teardown remained unchanged.

That diagnostic recorded 43 passes and 5 failures with zero skips/flaky/global errors: the first focus failure and later library/auth failures after the static JWT's lifetime. All seven invitation privacy cases passed in that failing run. Both proven causes received one bounded repair; the affected target and final original full gate then passed. Failed runs are not counted as completed verification.

## Known Stubs

None in the delivered path. Empty initial UI state is replaced by real authenticated inspection or explicit safe denial/recovery; existing unrelated future redirect/analytics copy is outside this plan.

## Remaining Acceptance and Next Plan Readiness

Ready for **02-27**; it was not started. Broad requirements remain pending, so requirements-completed is intentionally empty.

Actual configured provider verification/recovery/social factors, session rotation/cookies and delivery evidence remain required in **02-35**. Local injected SDK protocol and controlled email fixtures do not satisfy that gate. No live factors or real emails were used or claimed.

## Self-Check: PASSED

- Both created tests and all seven source/test paths exist.
- Preserved RED `8557537` and GREEN `5cd3150` exist in Git and retain RED-before-GREEN ordering.
- GREEN contains no tracked deletions; every required final gate passed with positive executed counts.
- No repository runner diagnostics or generated output changes remain.
- Unrelated .serena/project.yml remains untouched and unstaged.

