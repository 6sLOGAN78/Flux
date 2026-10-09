---
phase: 02-tenant-safe-link-control-plane
plan: "09"
subsystem: workspace
tags: [tenancy, cancellation, cross-tab, accessibility, nextjs]
requires:
  - phase: "02-08"
    provides: Membership-authorized bootstrap and committed workspace selection
provides:
  - Generation-guarded workspace requests and tenant subtree disposal
  - Signal-only cross-tab invalidation and focus authorization refresh
  - Native unsaved-name confirmation consumed by onboarding
affects: [02-13, 02-14, 02-16, 02-35]
tech-stack:
  added: []
  patterns: [abort plus request generation, scoped subtree unmount, signal-only invalidation, native safe-first dialog]
key-files:
  created:
    - apps/frontend/components/workspace-switcher.tsx
    - apps/frontend/components/workspace-chooser.tsx
    - apps/frontend/components/confirm-dialog.tsx
  modified:
    - apps/frontend/components/app-shell.tsx
    - apps/frontend/lib/workspace.ts
    - apps/frontend/lib/workspace.test.ts
    - apps/frontend/tests/control-plane.spec.ts
    - apps/backend/internal/handler/product_test.go
    - apps/frontend/app/onboarding/page.tsx
    - apps/frontend/app/workspaces/page.tsx
    - apps/frontend/app/workspaces/[workspaceId]/links/page.tsx
key-decisions:
  - Broadcast only a fixed invalidation signal; server bootstrap and committed selection remain the sole workspace authority.
  - Dispose the scoped subtree before switching, signout or access-loss recovery; keep valid same-workspace content during transient refresh failure.
  - Consume native unsaved confirmation in the existing onboarding name flow; future link drafts must use the same confirmation and scoped disposal boundaries.
requirements-completed: []
requirements-addressed: [TEN-06, TEN-07, TEN-08, SAFE-02, SAFE-03]
duration: 23min
completed: 2026-10-09
---

# Phase 2 Plan 9: Workspace Switching Summary

**Workspace changes abort pending requests, reject late generations and dispose tenant content before committed selection, with signal-only cross-tab invalidation and native unsaved-change confirmation.**

## Performance

- Tasks: 2/2. Approximately 23 minutes including setup, RED evidence and verification.
- Implementation and tests: 11 unique paths, including eight declared paths and three necessary page consumers. Existing backend selection service already met the locking and fresh-membership requirements and was preserved.

## Accomplishments

- `WorkspaceRequests` combines AbortController with generation checks. Superseding requests, switching, invalidation, session loss and unmount abort the previous transport and invalidate its callbacks even if transport completion ignores cancellation. The deterministic test releases an older promise and proves it cannot publish tenant content.
- Links consumes `WorkspaceSwitcher`. Its authorized snapshot owns the scoped subtree; scrubbing unmounts old content, selections and descendant state before PUT selection. Only the committed Go response supplies the new route identity. Session and workspace identity guard all renders. Current role capabilities remain closed and Go remains authoritative.
- BroadcastChannel carries exactly `"invalidate"`, never workspace IDs, names, roles, credentials or content. Object messages are ignored. Receivers scrub immediately and rebootstrap through the chooser. Membership loss clears current content, navigates to the chooser, focuses its heading and shows “Your workspace access changed. Choose an available workspace.” Signout clears local private state before the provider operation and emits invalidation.
- Focus refresh rechecks server bootstrap and the scoped workspace. Transient refresh failures retain valid same-workspace content and provide Retry. Scoped 403/404 responses confirm membership through fresh bootstrap instead of treating resource absence as revocation. Removed membership cannot leave former identifiers in the rendered chooser or workspace navigation.
- Native `ConfirmDialog` opens with Stay focused, closes safely on Escape, returns focus to the invoking control, and requires explicit Discard. The existing onboarding name flow consumes it when leaving for workspace selection; Discard clears the name and retry draft without submitting or creating a workspace. Session changes also clear that draft and confirmation.
- Backend actual-HTTP proof confirms a missing resource does not erase valid workspace membership or its authorized selection. Existing locked selection/removal races and all-role selection remain intact.

## Task Commits

1. **Specify switching and cancellation:** `6b85ab2` — three browser cases and deterministic/helper/actual-HTTP specifications before production changes. The complete initial browser selection failed; a bounded isolated private report confirmed authorized Switch Alpha rendered and the absent Switch workspace control timed out. Fixture startup failure was not accepted as RED.
2. **Deliver scoped switching and cross-tab recovery:** `3a4daf1` — generation guard, consumed switcher/chooser, native dialog/onboarding consumer and strengthened browser completion and refresh assertions.

## Verification

- `bun run test:unit`: passed all **87 discovered race-enabled Go top-level unit tests**, all four workspace suites, script regressions and tool self-tests, with dependency-ordered builds. Final frontend verification separately passed **16 tests** after the final cleanup refinements.
- Integration: the initial `bun run test:integration` stopped at unchanged `TestTracestateHTTPRedisOTLP`. An isolated race-enabled run passed. The full registered gate, invoked through the same exported `runChecks`/`runCommand` runner with additional safe failure-label capture, then passed **17 top-level tests across six groups**. Real PostgreSQL/Redis and race-enabled actual-HTTP assertions completed; no new unregistered top-level container test was added.
- `bun run test:e2e -- --project=local --grep switching`: final **4 completed, zero skipped/flaky**. Cases cover authorized switch/signout, real-Go two-tab membership revocation with a held older authorized response, native unsaved confirmation and same-workspace transient/404 recovery.
- The held old response is explicitly released and its route completion awaited before final absence/focus assertions. A finally block releases the hold and drains the second tab's owned provider assets. Both tabs lose workspace navigation and the revoked UUID from rendered main content.
- `bun run test:e2e -- --project=local`: **19 completed, zero skipped/flaky**, including the previous 15 auth, restore, workspace, keyboard and responsive cases.
- Root format, lint, typecheck and deterministic generated-contract checks passed. Final frontend format/lint checks passed. Final root production build passed after browser development execution; the existing validated `cleanFrontendBuild` helper then removed only its recognized generated output.
- `CI=true bun run scan`: Go imported-package and Bun dependency checks, worktree secrets and complete Git-history secrets passed. Inventory-only GO-2026-5932 remains visible; no advisory allowlist or scanner gate change was introduced.
- `git diff --check` passed; no tracked files were deleted. Browser reports/provider diagnostics stayed private outside the repository; no runtime, tool or contract pins changed.

## Deviations from Plan

**1. [Rule 3 - Blocking] Three consumed page paths**
- `apps/frontend/app/workspaces/page.tsx` consumes the declared chooser rather than retaining an independent implementation.
- `apps/frontend/app/workspaces/[workspaceId]/links/page.tsx` consumes the declared switching boundary, which owns disposal of the tenant subtree.
- `apps/frontend/app/onboarding/page.tsx` consumes native confirmation in an actual existing dirty-input flow and clears its retry draft on discard/session change. No link-creation feature or production test endpoint was fabricated.
- Commit: `3a4daf1`. These narrow consumers were authorized by the orchestrator and necessary to make the declared components reachable.

**2. [Rule 1 - Bug] Verification and reachable recovery refinements**
- Whole-body HTML assertions initially included Next's initial route scripts and authorized switcher option IDs. Assertions now inspect actual tenant main/navigation content; removed identifiers remain forbidden in the chooser. The held-route test now explicitly awaits completion and always releases/drains its second tab.
- Invalidation navigates to the chooser after immediate scrub, allowing the same still-authorized workspace to be reopened instead of leaving a switcher mounted in chooser mode on an unchanged route. An optional-snapshot narrowing error was corrected without suppressions.
- Commit: `3a4daf1`. Final targeted and full browser suites passed.

## Deferred Issues

- The unchanged telemetry integration test failed once, then isolated and full registered execution passed without source changes. The standard initial runner retained only its safe failing test name, so the original underlying cause is unavailable. The orchestrator reports a previous similar transient. This remains an existing verification concern recorded in `deferred-items.md`, not a claimed repair.

## Known Stubs

- `apps/frontend/components/app-shell.tsx:27`: existing Team button remains disabled until plan 02-16. Existing first-link creation remains disabled until 02-14. Neither prevents workspace switching.
- Link drafts, search, cursors and link dialogs are not implemented in this bounded slice. Future consumers must live within the disposed scoped subtree and use the native unsaved confirmation before switching; no fake production draft/data path was added.

## Decisions Made

- Cross-tab communication invalidates state only; browser storage/messages cannot authorize or select workspaces.
- Preserve valid same-tenant data on transient refresh failure while confirmed access loss immediately disposes the old subtree.
- Use the existing onboarding name draft as the real native-confirmation consumer; leave link creation and Team operations to their approved plans.

## Next Phase Readiness

- Scoped switching and chooser recovery are consumed and verified. Canonical identity/workspace contracts, locked Go preference selection and closed role capabilities remain unchanged.
- Live provider-factor acceptance remains pending final plan 02-35. Requirement IDs are addressed here but are not globally marked complete before that final gate.

## Self-Check: PASSED

- All 11 changed implementation/test paths exist, all three created components are consumed, and both task commits resolve in Git.
- Final required unit, registered integration and targeted browser verification completed positively; full browser, build, formatting, lint, typecheck, generation and scans passed as recorded above.
