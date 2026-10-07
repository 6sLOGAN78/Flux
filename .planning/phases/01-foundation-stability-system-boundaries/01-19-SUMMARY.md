---
phase: 01-foundation-stability-system-boundaries
plan: "19"
subsystem: testing
tags: [bun, biome, golangci-lint, staticcheck, govulncheck, gitleaks, sha256]
requires:
  - phase: 01-18
    provides: Existing pinned runtime, generator, module and collector provenance
provides:
  - Exact official quality tool release and executable hashes for Linux/macOS x64/arm64
  - Fail-closed installation, configuration verification and executable integrity regression fixtures
  - Strict Bun script type checking and compatible Biome/Go lint configurations
affects: [01-20, 01-21, 01-22]
tech-stack:
  added: [Biome 2.5.15, golangci-lint 2.14.0, Staticcheck 2026.2.1, govulncheck v1.8.0, Gitleaks 8.30.1]
  patterns: [Checksum before execution, atomic verified publication, isolated deterministic source builds]
key-files:
  created: [scripts/install-tools.ts, scripts/tsconfig.json, biome.json]
  modified: [tools.lock.json, apps/backend/.golangci.yml]
key-decisions:
  - "Build official govulncheck v1.8.0 source with Go 1.26.8 and a checksum-locked isolated module graph because upstream publishes no standalone release binaries."
  - "Retain effective Go checks through v2 staticcheck aliases and goimports formatter migration; preserve repository findings for the assigned quality-runner plan."
patterns-established:
  - "Quality binaries live in ignored tmp/tools; --install-root accepts a separate writable proof directory."
  - "Downloaded archives and extracted/built executables each have independent SHA256 pins; cached executable integrity is checked before every version probe."
requirements-completed: [PLAT-05, PLAT-06]
duration: 18min
completed: 2026-10-07
---

# Phase 01 Plan 19: Reproducible Quality Tooling Summary

**Five exact official quality tools install only after source/archive and executable SHA256 verification, with deterministic govulncheck builds and compatible Biome/Go lint configurations.**

## Performance

- **Duration:** 18min
- **Started:** 2026-10-07T15:48:07Z
- **Completed:** 2026-10-07T16:06:24Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Preserved existing generator/OTel/collector provenance and added exact pins from official [Biome releases](https://github.com/biomejs/biome/releases/tag/%40biomejs/biome%402.5.15), [golangci-lint releases](https://github.com/golangci/golangci-lint/releases/tag/v2.14.0), [Staticcheck releases](https://github.com/dominikh/go-tools/releases/tag/2026.2.1), [Gitleaks releases](https://github.com/gitleaks/gitleaks/releases/tag/v8.30.1), and [govulncheck source](https://go.googlesource.com/vuln/+/refs/tags/v1.8.0).
- Verified all sixteen official prebuilt archives/binaries against GitHub-published SHA256 metadata and checksum assets. Measured extracted executable hashes independently. Cross-built and measured govulncheck hashes for Linux/macOS x64/arm64.
- Implemented `--verify`, `--verify-config`, `--self-test`, and `--install-root PATH` with official-origin validation, exact versions, 120-second download deadlines, download/expansion size limits, 300-second subprocess deadlines, bounded captured diagnostics and atomic publication after verification.
- Kept two-space/double-quote/semicolon formatting and recommended Biome checks. Migrated obsolete Go linter aliases into staticcheck's complete check set, retained type checking as built-in behavior, and moved goimports to the v2 formatter section.
- Type-checked all scripts with the existing locked Node declarations and Bun-compatible module semantics, without adding package dependencies.

## Task Commits

1. **Task 1: Specify and prove the owned behavior** — `5047dee` (`test`): happy-path/integrity/version/download regressions recorded RED with `Installation not implemented` before implementation.
2. **Task 2: Implement pin and install reproducible quality tooling** — `1648dcb` (`feat`): official pins, installer, configurations, expanded timeout/cleanup regressions and passing verification.

## Files Created/Modified

- `scripts/install-tools.ts` — Verified installation, safe archive member extraction, deterministic source build, CLI modes and embedded self-tests.
- `scripts/tsconfig.json` — Strict, dependency-free checking of Bun scripts through existing workspace Node typings.
- `biome.json` — Authored TS/script formatting and recommended lint rules; excludes generated OpenAPI artifacts checked by the existing generator.
- `tools.lock.json` — Official exact release URLs, published archive checksums, measured binary hashes and the isolated govulncheck build graph.
- `apps/backend/.golangci.yml` — Compatible v2 configuration preserving security, correctness and style checks.

## Verification

- `bun scripts/install-tools.ts --verify`: PASS for all five actual Linux x64 binaries.
- `bun scripts/install-tools.ts --verify-config`: PASS through exact Biome and golangci-lint binaries; Go linter enumeration also succeeds.
- `bun scripts/install-tools.ts --self-test`: PASS. Corrupt downloads, corrupt cached executables and wrong extracted-binary hashes fail before execution; version `1.2.30` is rejected for `1.2.3`; failed downloads and rejected versions never publish; read-only module cache cleanup succeeds; stderr cannot corrupt JSON stdout; a real sleeping subprocess is terminated by a 30ms test deadline.
- Full installation and configuration verification in a fresh temporary `--install-root`: PASS, with no residual `.install-*` directories and removal of the proof directory afterward.
- Independently built govulncheck executables from separate directories and caches match SHA256 `f417d39b256c1a73c83d64fbcd8b063547d4cf67e151502e2340456003feecc8`.
- `tsc -p scripts/tsconfig.json`: PASS for all current scripts.
- Biome formatting of all owned TS/JSON files and linting of the installer with `--error-on-warnings`: PASS with zero warnings.
- `git diff --check`: PASS. No tracked files deleted; installed binaries remain in the existing ignored `tmp/` directory.

Linux x64 binaries were actually executed. Other platform archives were downloaded and hashed, and govulncheck was cross-built for those targets; this environment did not execute macOS or arm64 binaries. Unsupported OS/architecture combinations fail explicitly.

## Decisions Made

The official [Go source README](https://go.googlesource.com/vuln/+/refs/tags/v1.8.0/README.md) distributes govulncheck through source builds. Use that official exact source, Go module sums and source ZIP SHA256, a locked temporary module graph, the project Go 1.26.8 compiler, CGO disabled, fixed architecture settings, `-trimpath`, `-buildvcs=false`, and an empty build ID. Compare the resulting binary hash before execution. No application go.mod/go.sum or Bun dependency graph changed.

The root manifest does not declare ESM. Use ESNext/Bundler semantics for Bun-only scripts instead of changing the root package owned by the next plan. Reuse the installed OpenAPI workspace's Node declarations and declare only the existing generator's narrow Bun subprocess API.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Official govulncheck releases have no prebuilt binary assets**
- **Found during:** Task 2 official release resolution.
- **Issue:** The plan assumed every tool had an official standalone binary download. GitHub's old release listing is also behind the official Google source tags.
- **Fix:** Resolve official v1.8.0 from Google's source repository and produce a standalone executable through a checksum-verified, exact-compiler source build in isolated temporary module/build caches. Record measured hashes for all four supported targets and prove independent-directory reproducibility. The orchestrator confirmed this implementation choice within the already-authorized scope.
- **Files modified:** `scripts/install-tools.ts`, `tools.lock.json`.
- **Verification:** Fresh installation, module integrity verification, version probe and independent SHA256 comparison all PASS.
- **Committed in:** `1648dcb`.

## Issues Encountered

Go prints download progress to stderr alongside JSON metadata on stdout. The installer now captures stdout separately and drains/bounds stderr without exposing it. Go's downloaded module directories are read-only; cleanup now changes permissions only inside the installer's own temporary tree without following symlinks. Both cases have regression checks. A Biome check with every analysis tool disabled processes no files; configuration validation uses the formatter on the actual owned configuration instead.

## User Setup Required

None. Run the installer through Bun with the supported Go executable on PATH. No credentials or external service configuration required.

## Next Phase Readiness

Plan 01-20 can call the installer and use the exact binaries in `tmp/tools`; plan 01-21 owns scans and plan 01-22 owns full CI. Repository-wide Go/TS lint findings and quality orchestration are deliberately left to their assigned plan; this plan proves installation and configuration compatibility and zero-warning lint for its new installer.

## Self-Check: PASSED

All five claimed files exist. Both task commits exist. No goal-blocking production stubs or additional unmodeled threat surfaces were introduced. The existing Go lint exclusion mentioning TODO is a configuration pattern, not a production stub.
