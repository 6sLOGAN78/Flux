# Hosted CI secret-scan follow-up

[Run 37746551091](https://github.com/6sLOGAN78/Flux/actions/runs/37746551091) failed at the final `scan` stage on f38fcb7. Every preceding quality stage passed. The earlier hosted pass covered implementation before the final verification report was committed; planning files are included in secret scans and therefore need their own current acceptance evidence.

Direct `CI=true bun run scan` reproduced two worktree findings and their historical counterparts. Private inspection identified the Sourcegraph rule at lines 9 and 529 of 01-VERIFICATION.md: both values are the public implementation commit referenced by the successful hosted run. No credential was found at these locations. A SourceGraph helper name elsewhere in the report supplies the case-insensitive fragment keyword.

The [exact upstream Gitleaks 8.30.1 rule](https://github.com/gitleaks/gitleaks/blob/v8.30.1/config/gitleaks.toml) accepts bare 40-character hex in a fragment containing the provider keyword. Refine only this rule: all three prefixed formats remain recognized anywhere; legacy bare hex requires an adjacent provider-named assignment, including SRC_ACCESS_TOKEN. Keep entropy checking and inherit every other pinned default rule. No SHA exemption, credential/path/commit allowlist, baseline, advisory exception, history rewrite or skipped history is introduced.

Add the recognized public rule identifier to the scanner's closed diagnostic vocabulary. Move the direct sanitized scanner command before `bun run check` in CI, so findings are visible before the root runner suppresses arbitrary subprocess diagnostics. The full root check still includes its final scan, covering files generated during its other stages. Permissions, immutable action/tool pins, timeouts and all quality gates remain enforced.

Regression evidence:

- A new real-pinned-tool fixture creates a Git commit, writes its public SHA alongside an unrelated SourceGraph helper mention, and commits the evidence. It failed under the previous rule and passes in both worktree and full history with the refinement.
- Runtime-generated legacy SOURCEGRAPH_ACCESS_TOKEN and SRC_ACCESS_TOKEN assignments plus all three prefixed formats fail in the worktree. Each is committed and removed; full-history scanning continues failing while the worktree is clean. Captured diagnostics contain none of the generated credentials.
- All 13 scanner tests passed under CI=true, including malformed reports, missing/shallow history, deleted GitHub credentials, malicious metadata and real vulnerable Go imports. Biome, scripts typecheck and workflow permission/timeout/order validation passed.
- Actual repository dependency, worktree and full-history scans passed after the refinement. The unrelated unimported Go module inventory advisory remains visible under the existing generic package-exposure policy.

Final acceptance uses the full local command and the current hosted workflow, rather than reusing the earlier hosted pass. This follow-up does not start Phase 2 or alter the completed foundation's product scope.

## Application integration diagnostic follow-up

[Run 37751432635](https://github.com/6sLOGAN78/Flux/actions/runs/37751432635) passed the direct sanitized scanner and all preceding full-check stages, then failed `backend:integration:internal/app`. The complete local check passed on the same source. Its hosted failure did not identify a test because the root runner discarded output from unsuccessful subprocesses.

Preserve that privacy boundary while extracting only failed Go test names already present in the repository-discovered stage manifest. Never expose event output, package names, arbitrary subtest labels or unknown identifiers. Non-JSON diagnostics retain the existing generic failure. The real subprocess regression verifies a failed child with private stdout/stderr and unrecognized metadata produces only the known top-level test identifier in the root failure message.

This change improves the evidence needed to diagnose the integration failure; it does not change application behavior, skip a test, retry a failed CI gate or relax a deadline. Acceptance still requires the current hosted workflow.
