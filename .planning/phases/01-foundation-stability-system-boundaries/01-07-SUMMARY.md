---
phase: 01-foundation-stability-system-boundaries
plan: "07"
subsystem: infra
tags: [go, embed, scalar, csp, sri, react-email, templates]
requires:
  - phase: 01-05
    provides: Canonical deterministic authored OpenAPI and identical served artifact
  - phase: 01-06
    provides: Verified generated Go health transport boundary
provides:
  - Embedded documentation HTML and canonical JSON served at existing URLs
  - Verified exact Scalar viewer release with SRI and constrained CSP
  - Embedded deterministic welcome template with safe render-only email seam
affects: [01-08, 01-09, 01-10, 01-11, 01-12, foundation-quality-gates]
tech-stack:
  added: []
  patterns: [package-owned embedded FS, closed template allowlist, render-only email seam, integrity-pinned viewer]
key-files:
  created: [apps/backend/static/assets.go, apps/backend/templates/assets.go, apps/backend/internal/handler/openapi_test.go, apps/backend/internal/lib/email/client_test.go]
  modified: [apps/backend/static/openapi.html, apps/backend/templates/emails/welcome.html, packages/emails/src/templates/welcome.tsx, apps/backend/internal/handler/openapi.go, apps/backend/internal/router/system.go, apps/backend/internal/lib/email/client.go]
key-decisions:
  - Use the official Scalar 1.73.0 standalone UMD bundle, verified against its checksum-validated npm tarball, with sha384 SRI and hash-only script CSP.
  - Use a closed Template switch and html/template ParseFS with missingkey=error; rendering requires no provider client and sending follows successful rendering.
patterns-established:
  - Documentation and email ownership is exposed through package-level embed.FS exports independent of process working directory.
  - Email HTML remains an export of the locked React Email CLI with Go interpolation tokens preserved.
requirements-completed: [PLAT-08, PLAT-06]
duration: 5min
completed: 2026-10-06
---

# Phase 1 Plan 7: Embedded Documentation and Email Assets Summary

**Embedded docs preserve canonical JSON bytes and existing URLs, while deterministic email exports render escaped data from any working directory.**

## Performance

- Started: 2026-10-06T13:31:55Z
- Completed: 2026-10-06T13:37:10Z
- Duration: approximately 5 minutes
- Tasks: 2
- Source/artifact files changed: 10

## Accomplishments

- Embedded `openapi.html` and the existing generated `openapi.json` in the static package. `/docs` serves the HTML; `/static/openapi.json` serves unchanged canonical bytes with JSON content type. Both responses revalidate through `Cache-Control: no-cache` and set `nosniff`. Replaced the filesystem-relative static route with the exact embedded document route.
- Pinned the official Scalar 1.73.0 standalone bundle. Verified the registry tarball's sha512 integrity and exact equality between its standalone file and the CDN response before calculating sha384 SRI. The docs response permits only that script hash, restricts connections to the same origin, and denies frames, objects, and base URI changes. Disabled default external fonts and telemetry. Inline styles remain allowed as required by Scalar.
- Embedded `emails/*.html` and introduced `Client.Render`, which neither constructs nor calls Resend. Only the existing `TemplateWelcome` enum value maps to a template path. Unknown names fail before transport use. `html/template` escapes text, rejects missing required data, and returns no partially rendered body on errors. `SendEmail` calls the same rendering seam and reports an unconfigured transport instead of panicking.
- Removed wall-clock copyright output and the fictitious street address; the stable footer uses Flux. Exported with the already installed, lockfile-pinned React Email 6.3.3 CLI, preserving `{{.UserFirstName}}`.

## Task Commits

1. **Task 1 RED: Require docs assets after changing directory** — `03f64ed` (test).
2. **Task 1 GREEN: Embed docs and pin integrity-verified viewer** — `2d0a35d` (feat).
3. **Task 2 RED: Require safe directory-independent email rendering** — `f3ccf53` (test).
4. **Task 2 GREEN: Embed deterministic email templates and safe rendering** — `20ccc85` (feat).

Normal Git commits were used. No tracked files were deleted.

## Verification

- RED: The docs suite failed because the embedded JSON serving seam was absent. The email suite failed because the render-only seam was absent. Each failure was recorded before its implementation commit.
- GREEN: `go test ./internal/handler ./internal/lib/email -run 'Test(OpenAPI|Embedded|Render|Template|AlternateDirectory)' -count=1` passed.
- Actual Echo handlers were invoked after `t.Chdir(t.TempDir())`; responses retained expected HTML, headers, and byte-identical canonical JSON. Email tests rendered after chdir, checked HTML escaping and token execution, rejected eight unknown/path-shaped names before sending, and rejected missing required data.
- `go test -race ./internal/handler ./internal/lib/email -count=1`, handler/router race checks, and `go vet ./internal/handler ./internal/router ./internal/lib/email ./static ./templates` passed.
- `go build ./...` passed for the backend. `node node_modules/typescript/bin/tsc --project packages/emails/tsconfig.json --noEmit` passed for the changed email source.
- Two independent temporary CLI exports matched each other and the checked welcome HTML exactly. HTML sha256: `537a01d6349fdfa619169ff30276f444be3d92f0f9085ca4b903b2b1309db734`.
- Playwright drove the installed Google Chrome against a temporary fixture serving the exact docs HTML, canonical JSON, and response CSP. Both documented health operations rendered, with no page or browser console errors. Go regressions separately exercised the real handlers.
- Official [Scalar HTML integration documentation](https://scalar.com/products/api-references/integrations/html-js) and Context7 guidance confirmed the standalone bundle and CSP style requirements. [Published Scalar 1.73.0 metadata](https://registry.npmjs.org/@scalar/api-reference/1.73.0) identified the official repository and checksum-verified tarball. Context7 and installed React Email 6.3.3 source confirmed CLI export uses default props and preserves the Go tokens.
- Whitespace, production stub, and threat-surface scans found no unexpected outputs, production placeholders, or new unplanned trust boundaries.

## Decisions Made

Used a single standalone UMD file so SRI and CSP cover the complete viewer script without follow-up JavaScript chunk downloads. The existing viewer remains CDN-loaded; the docs HTML, contract, and email templates are embedded. Viewer updates require changing its exact version, verified SRI, and matching CSP together.

Kept constructors and sending callers compatible. `Render` is independent of provider credentials and configuration; missing template keys fail explicitly rather than silently producing broken emails.

## Deviations from Plan

None — both bounded five-file tasks followed the required RED/GREEN sequence and exact ownership union.

## Issues Encountered

The Playwright-managed Chromium executable was absent. Used the installed Google Chrome executable through the existing Playwright Python package for the browser verification; no browser or npm package was installed. No authentication gates occurred.

## User Setup Required

None.

## Next Plan Readiness

Plan 01-08 can verify embedded asset drift and deterministic email export through root quality gates. Plans 01-09 through 01-12 can narrow role construction while preserving the embedded FS packages and documentation URLs. Phase 1 remains in execution; no subsequent plan was executed here.

## Self-Check: PASSED

All ten owned source/artifact files and this summary exist. Commits `03f64ed`, `2d0a35d`, `f3ccf53`, and `20ccc85` exist without tracked deletions. Both tasks passed their acceptance checks, independent email exports matched the checked asset, and no generated repository files remain untracked.
