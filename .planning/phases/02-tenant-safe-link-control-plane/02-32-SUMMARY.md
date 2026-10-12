---
phase: 02-tenant-safe-link-control-plane
plan: "32"
subsystem: ui
tags: [responsive, reflow, contrast, playwright, forced-colors, reduced-motion]
requires:
  - phase: 02-31
    provides: Native keyboard forms, modal focus and truthful loading feedback
provides:
  - Desktop sidebar, tablet navigation and mobile workspace reflow
  - Readable enlarged text, long destinations and bounded scrollable dialogs
  - Native viewport, contrast-frame and accessibility-media verification
affects: [02-33, 02-35]
tech-stack:
  added: []
  patterns: [relative typography, labeled mobile cards, immediate contrast-safe color changes]
key-files:
  created: []
  modified:
    - apps/frontend/app/globals.css
    - apps/frontend/components/app-shell.tsx
    - apps/frontend/components/link-list.tsx
    - apps/frontend/components/team-list.tsx
    - apps/frontend/tests/accessibility.spec.ts
key-decisions:
  - Preserve exact destination link names and show the new-tab notice as adjacent visible text.
  - Keep foreground and background colors immediate; optional feedback uses opacity only and reduced motion disables it.
patterns-established:
  - Read-only native animation-frame samples verify contrast through real button enablement, rather than waiting for a final color.
requirements-completed: []
requirement-slices: [TEN-01, TEN-02, TEN-03, TEN-04, TEN-05, TEN-06, TEN-07, TEN-08, LINK-01, LINK-02, LINK-03, LINK-05, LINK-06, LINK-07, LINK-08, LINK-09, SAFE-02, SAFE-03, SAFE-04]
duration: 43min 09s
duration-basis: Captured continuous active window through original canonical full GREEN; prior dispatch idle time and metadata close-out excluded
completed: 2026-10-12
---

# Phase 2 Plan 32: Responsive Native Control Plane Summary

**Workspace screens now reflow from a 224px desktop sidebar to tablet navigation and mobile cards, with enlarged readable text, scrollable native dialogs and contrast-safe feedback.**

## Performance

- Started: **2026-10-12T18:48:22Z**. Original canonical full GREEN observed: **2026-10-12T19:31:31Z**. Duration: **43min 09s**.
- Tasks: **2/2**. Source/test files modified: **5**, all declared. PLAN classification and execution metadata are separate.
- No source files were created. Existing detail, chooser and confirmation components receive the shared CSS refinement without unnecessary edits.

## Accomplishments

- Desktop workspace navigation uses a fixed, independently scrollable **224px** sidebar; main content has **1200px** maximum width and **24px** inset, with a **64px** section header. Account navigation sits at the sidebar bottom. Tablet uses a workspace/account topbar and horizontal native navigation. Below **768px**, navigation stacks, content uses **16px** inset and action buttons span the available width while retaining native 44px targets.
- Authentication/chooser retain the existing **448px** panel. The existing dedicated create form receives a centered **640px** maximum width. Four relative typography sizes preserve the approved 16/14/20/28px defaults and 400/600 weights while responding to text enlargement. Heading/workspace names and full destinations wrap; definition-list values retain their complete text.
- Links and Team preserve semantic desktop tables and expose the same authorized records as labeled mobile cards. The Links breakpoint now agrees with the approved 768px boundary. Rows, token spacing, neutral surfaces and card borders remain original native UI. Destination anchors preserve their complete accessible URL name and safe target/rel/href, with a visible adjacent new-tab notice.
- Shared native dialog bounds and scrolling preserve safe-first keyboard focus, Escape and trigger return in a short viewport. Focus scroll margins and native forced-color outlines retain visible controls. Reduced motion disables feedback and animations, with no positional animation, smooth scrolling or shimmer added.
- Original Flux text branding, textual lifecycle states, local full timestamps with timezone and the management-only notice remain visible. No redirect, analytics, editing, preview fetching, provider auth bypass or new product capability was introduced.

## Task Commits

1. **Task 1: Specify viewport, enlargement and media behavior** — `7bfc934` (test).
2. **Task 1/2: Capture transient native button contrast regression** — `4a805e8` (test).
3. **Task 2: Deliver responsive shell and readable feedback** — `14d17c3` (feat).

Exact path staging and normal commits were used. No tracked files were deleted. Unrelated `.serena/project.yml`, `README.md` and `.planning/quick/` changes were preserved and excluded.

## Verification

- Original targeted native responsive gate passed **3/3**, both before and after the contrast correction. Actual migrated PostgreSQL/Go fixtures create real workspaces/links; real Clerk SDK transport remains test-local. Existing unchanged library regression target passed **1/1** after restoring its exact destination link name.
- The final original **CI=true `bun run check` passed** all required stages, including **65/65 browser cases** (all previous 62), **109 selected race-enabled Go unit tests**, **21 registered integration top-level tests across 6 groups**, **31 frontend units**, **6 email tests**, **15 OpenAPI tests**, schema/script/tool suites, formatting/lint/types, backend/workspace builds and migration verification.
- All **five** generated artifact comparisons passed: both OpenAPI copies, Go transport and both email templates. No runtime, dependency, generator, scanner, browser deadline, native fixture or quality-runner contract changed.
- The exact production frontend build passed again **after** browser execution, while only independent Go checks were active. Native full-run cleanup validated the owned nonsymlink production manifest and removed output before its CI security scan; `apps/frontend/build` is absent.
- CI scanning passed imported Go packages across roles/tests, Bun dependencies, the complete worktree and full Git history without exceptions. Metadata and committed task history receive a final original CI scan during close-out.
- Actual **320/768/1280px** library, Team and detail geometry has no horizontal page overflow, authorized desktop table/mobile card semantics and visible 44px controls. A **320×256 CSS-pixel viewport**, equivalent to reflow from 1280×1024 at 400% zoom, combined with **200% root text** proves 32px body text, complete long values, scrollable dialog and fully visible physically focused safe action. This is zoom-equivalent reflow, not an automated browser-chrome zoom claim.
- Native media checks verify computed normal text contrast **≥4.5**, control/focus contrast **≥3**, physical keyboard focus, forced-colors outlines, zero animation/transition under reduced motion and no smooth scroll. Read-only frame samples check contrast through actual disabled-to-enabled button changes without modifying the UI or waiting out a failing frame.
- Browser plugin was unavailable; the planned native Playwright/Chromium workflow was used. Private screenshots of all three widths, the enlarged dialog and forced colors were inspected during refinement. Final corrected source was validated by the complete native DOM/keyboard/media suite; no public screenshot/report artifact is created.

## TDD Evidence and Issues Encountered

The unchanged desktop main began at **416px**, reflecting its old centered 448px authentication panel, instead of the required **224px** sidebar boundary. This is genuine rendered absent-behavior RED. The initial root-text injection used `html`, which lost to the existing `:root` selector; the fixture was corrected to `:root` and not counted as a production defect.

The first original target's session metadata/output was accidentally discarded. Its outcome is not claimed; the retained native diagnostic ran **3 cases**, with the existing Team case passing and the new desktop/fixture cases failing. Private diagnostics preserve native startup, teardown, deadlines and report validation and capture only the first `runCommand` close output.

The media test initially assumed a Flux document title and absence of the entire Next portal. Actual source/report showed a pre-existing empty title and a normal development indicator sharing the portal with errors. Page identity now checks the actual route, Flux wordmark and rendered heading; actual error-overlay surfaces, page errors, non-informational console warnings/errors and native report global errors remain failures. Only the known exact Clerk development-key informational warning is classified separately.

The first original full browser run failed; its retained unchanged native diagnostic executed **65**, with **64 passing**, no global errors/skips/flaky cases, and one existing library failure: adding a new-tab suffix to the anchor's accessible name broke the exact full-URL contract. The owned anchor now retains its original name with adjacent visible notice; the unchanged regression target passed.

A second original full browser failure was diagnosed with another unchanged native report: **64/65**, no global errors/skips/flaky cases, **185873ms**, failing actual text contrast **1.0477** during the newly added foreground/background interpolation. A strengthened read-only animation-frame test reproduced computed **rgb(101,113,129)** on **rgb(207,218,244)** at **3.5413**, below 4.5. The fix keeps colors immediate and optional feedback opacity-only. No contrast assertion was removed or delayed. The corrected original target passed three; the final original full suite passed 65.

Both retained full diagnostic durations were below the existing 240-second browser deadline; neither failure was a budget issue. No timeout, partition, retry, skipped case, scanner exception or gate-count reduction was introduced. Final runs used immutable source and frontend checks sharing outputs were serialized.

## Deviations from Plan

**[Rule 3 - Blocking] Correct browser assumptions within the owned specification.** Root-style specificity, pre-existing title and normal development portal assumptions were corrected from source/native reports. These were fixture issues, not fabricated business RED. Broad requirement IDs remain slices, matching prior plans; phase-wide completion remains plan 35.

**[Rule 1 - Bug] Preserve destination anchor compatibility.** The new-tab notice initially changed an existing exact accessible name. `link-list.tsx` now presents it as adjacent visible text while preserving the complete URL name and original safe href/target/rel. Verified by unchanged target and all 65 cases; commit `14d17c3`.

**[Rule 1 - Bug] Eliminate transient unreadable text.** New simultaneous foreground/background transitions crossed an inaccessible intermediate contrast. Native frame sampling provides real RED in `4a805e8`; immediate colors and reduced-motion-aware opacity-only feedback fix it in `14d17c3`.

No extra source ownership, architecture, dependency or production trust boundary was needed.

## Known Stubs

None introduced. Existing empty inputs, stable loading decoration and the management-only future redirect/analytics notice are intentional states, not placeholder production paths preventing this goal.

## Remaining Acceptance and Setup

- The pre-existing empty document title remains an unrelated limitation outside this plan's declared paths; it is not reported as successful title QA.
- Local Chromium viewport/DOM/media/keyboard proofs do not claim auditory screen-reader testing, additional browser engines or live provider factor/recovery/OAuth/session acceptance.
- Live Clerk and real recipient delivery acceptance remain pending plan **02-35**. No live provider configuration, private keys or recipient email were accessed. No new user setup is required for this slice.

## Next Phase Readiness

Plan 32 is complete; ready for plan **02-33** within Phase 2. Broad requirements and actual provider acceptance remain pending. No phase completion, transition to Phase 3 or push occurred.

## Self-Check: PASSED

All five modified source/test files and three task commits exist. No source files were created or tracked files deleted. Declared ownership, native full GREEN, original generation/security gates, absent frontend output and preservation of unrelated changes were verified before SDK close-out. Private diagnostic captures are removed during close-out; final committed-history scanning follows metadata commits.
