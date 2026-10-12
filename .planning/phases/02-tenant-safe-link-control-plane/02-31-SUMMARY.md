---
phase: 02-tenant-safe-link-control-plane
plan: "31"
subsystem: ui
tags: [accessibility, keyboard, react, playwright, native-dialog, truthful-feedback]
requires:
  - phase: 02-30
    provides: Verified tenant disposal and authenticated control-plane workflows
provides:
  - Persistent described fields, invalid states and focused blocking feedback
  - Native safe-first dialog keyboard proof, shell skip link and route heading focus
  - Stable inert loading rows with delayed cancellation and explicit retry
  - Committed status and clipboard announcements with duplicate-submit proof
affects: [02-32, 02-33, 02-35]
tech-stack:
  added: []
  patterns: [shared native field semantics, stable loading geometry, scoped keyboard locators]
key-files:
  created:
    - apps/frontend/components/field.tsx
    - apps/frontend/components/inline-notice.tsx
    - apps/frontend/components/loading-rows.tsx
    - apps/frontend/tests/accessibility.spec.ts
  modified:
    - apps/frontend/components/confirm-dialog.tsx
    - apps/frontend/components/app-shell.tsx
    - apps/frontend/components/link-form.tsx
    - apps/frontend/components/link-detail.tsx
    - apps/frontend/components/team-list.tsx
    - apps/frontend/components/invitation-panel.tsx
    - apps/frontend/app/globals.css
key-decisions:
  - Preserve native modal behavior and existing request generations, mutation idempotency and fresh permission refresh.
  - Keep the workspace h1 and focus the current main screen heading on pathname changes.
patterns-established:
  - Keyboard tests target actual background action groups separately from nested modal actions.
requirements-completed: []
requirement-slices: [TEN-01, TEN-02, TEN-03, TEN-04, TEN-05, TEN-06, TEN-07, TEN-08, LINK-01, LINK-02, LINK-03, LINK-05, LINK-06, LINK-07, LINK-08, LINK-09, SAFE-02, SAFE-03, SAFE-04]
duration: 37min 56s
duration-basis: Captured active window through canonical full GREEN; metadata close-out excluded
completed: 2026-10-12
---

# Phase 2 Plan 31: Native Keyboard Accessibility Summary

**Described fields, focused errors, safe native dialogs and cancellable inert loading states now support actual keyboard workflows without claiming uncommitted success.**

## Performance

- Started: **2026-10-12T14:22:03Z**. Canonical full GREEN observed: **2026-10-12T14:59:59Z**. Measured active duration: **37min 56s**.
- Tasks: **2/2**. Source/test files: **11**, all declared. PLAN requirement classification and execution metadata are separate.

## Accomplishments

- Shared `Field` renders persistent labels, associated hints and invalid field descriptions. Link inputs use URL/off autocomplete; invitation email uses email autocomplete and native validity with a focused blocking summary. Link validation identifies destination, title or custom key without changing server authority.
- The existing header/nav/main and single workspace h1 remain intact. A keyboard-visible skip link focuses main. Pathname changes focus the current main screen heading using a negative tabindex; the browser checks reject positive tabindex.
- Existing `showModal`, Escape handling, contextual safe-first focus and connected-trigger restoration are preserved. Reason validity uses the existing Unicode normalization/codepoint/NUL policy. Native keyboard tests cover archive/suspension, role changes, member removal and invitation revocation. The inert background cannot take focus while the modal is open.
- Shared loading rows have stable geometry, inert/aria-hidden decoration and textual status. After five seconds, contextual delayed copy offers cancellation and explicit retry. Link form/detail, Team and invitations wire these controls into their existing abort/generation boundaries. Team busy state also includes mutation work.
- Disabled creation, invitation and role controls have readable explanations. Four button styles, 44px native control targets, forced-colors/reduced-motion support and scrollable bounded dialogs use the approved neutral/blue system. No new library or motion effect was added.
- Clipboard feedback remains polite and only follows the actual clipboard result. Creation stays disabled while pending; a held real request proves one POST and no success before its committed response. Existing uncertain retry payloads, fresh SQL capabilities, known committed results and tenant disposal behavior are unchanged.
- Invitation load failure no longer displays an unverified empty-state claim. Existing valid same-workspace records remain available during transient refresh failures; the full tenant-denial suite continues to prove access-loss disposal.

## Task Commits

1. **Task 1: Specify native keyboard accessibility and delayed loading** — `986a043` (test).
2. **Task 2: Deliver accessible native forms, dialogs and loading feedback** — `f825a61` (feat).

Exact path staging and normal commits were used. No tracked files were deleted. Unrelated `.serena/project.yml` remains untouched and unstaged. No push or plan 32 execution occurred.

## Verification

- The original targeted accessibility gate passed **5/5** cases. The subsequent original canonical full browser gate passed **62/62**, including all 57 previous cases, with no skipped, unexpected, flaky or global errors accepted by its native report validator.
- Focused frontend units passed **31/31**. The final original **CI=true `bun run check` passed** formatting, strict lint, types, workspace builds/units, compiled Go discovery, **109 selected race-enabled Go unit tests**, **21 registered integration top-level tests across 6 groups**, migration verification, independent backend builds and security scans. Existing **6 email** and **15 OpenAPI** tests remain green, alongside schema/script/tool suites.
- `generate:check` passed all **five** artifact comparisons: both OpenAPI copies, Go transport and both email templates. No authored/generated contract, runtime pin, dependency manifest, runner, deadline or scanner rule changed.
- Production frontend building passed again **after** the complete browser suite. The original full runner then validated the owned nonsymlink production manifest before deleting output through `cleanFrontendBuild`; the output directory is absent.
- CI security scanning includes imported Go packages across roles/tests, Bun dependencies, the complete worktree and full Git history. No exception was added. Metadata close-out receives a final original CI scan separately.
- `git diff --check` passed. Owned private browser diagnostics and temporary formatter output were removed.

## TDD Evidence and Issues Encountered

The unchanged UI produced three executed failing cases with no global errors before production changes. The missing skip link was a genuine absent-behavior RED. The first two failures exposed specification assumptions: an unscoped alert also matched the provider, and reverse traversal from the first native dialog button was incorrectly assumed to immediately wrap to its destructive button. They were corrected rather than counted as production defects. Existing native modal security was not given an artificial RED.

The very first run executed zero cases because a copied fixture ended inside an incomplete `beforeEach`; this collection failure is explicitly **not** business RED. A private diagnostic runner initially inserted its capture into a second teardown callback as well; only the original `runCommand` close capture was retained afterward. The private copy otherwise preserved native fixture startup, teardown, deadlines and report validation.

One intermediate original run overlapped production edits and is diagnostic only. Final passing target and canonical full runs used immutable source. Two later native failures came from a background Archive locator also matching the nested modal button; actual component nesting identified the separate lifecycle action group. Physical Tab/Shift+Tab, Escape, trigger-return and inert-background assertions were preserved. Four of five cases passed before that locator correction; the corrected original target passed five and the full gate passed 62. No timeout, retry, assertion waiver or production auth substitute was introduced.

## Deviations from Plan

**[Rule 3 - Blocking] Correct bounded browser specification and diagnostic fixture boundaries.** Task 1/2 fixture parsing and native locator assumptions were corrected within the declared accessibility specification. The native provider transport was narrowly copied from the delivered tenant fixture instead of importing a specification that would register duplicate tests. Its renewed 30-day Client cookie remains independent of the bounded session/JWT and real SDK token-refresh protocol. No extra ownership or architecture was needed.

**[Rule 1 - Bug] Suppress an unverified invitation empty-state after read failure.** Found during Task 2 feedback wiring. The existing error path could also say there were no pending invitations without a successful read. The conditional now requires no error before displaying that empty state. File: `invitation-panel.tsx`; verified by the final canonical full gate; commit `f825a61`.

Broad requirements remain requirement slices, with phase-wide completion pending plan 35.

## Known Stubs

None introduced. Controlled empty inputs, inert loading decoration and the test-only signed provider transport are intentional states/fixtures, not placeholder production data.

## User Setup Required

None for this slice. Live provider factors/recovery/OAuth/session and recipient-delivery acceptance remain pending plan 02-35. No live provider configuration, credentials or recipient email were accessed.

## Next Phase Readiness

Plan 31 is complete. Rendered viewport/zoom refinement remains allocated to plan 32; local Chromium DOM/keyboard proofs are not auditory screen-reader or live-provider acceptance. Browser plugin was unavailable, so the explicitly planned native Playwright runner was used. No screenshot or additional browser-engine claim is made.

## Self-Check: PASSED

All four created files and both task commits exist. Declared ownership, gate evidence, diagnostic cleanup, absent frontend output, no tracked deletions and preservation of unrelated changes were checked before SDK close-out.
