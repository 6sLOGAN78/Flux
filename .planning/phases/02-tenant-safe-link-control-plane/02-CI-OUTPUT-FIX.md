# Phase 2 CI output and scanner timing repair

## Trigger and resulting behavior

Hosted run 37778710983 on the completed 02-01 commit passed frontend/browser and backend checks, then failed the final secret scan. Reproduction identified generated Next deployment-local preview/action keys in the disposable frontend build tree. Private browser diagnostic reports also contained test-only signed credentials locally; five owned reports were moved to a protected temporary directory outside the repository.

The quality runner now removes only verified Next output at `apps/frontend/build` after all behavior/build checks and immediately before the final full-worktree scan. Cleanup refuses symlinked or unrecognized output, verifies the canonical path and version-4 preview metadata, and preserves authored files. Independent `bun run build` remains available to produce deployment output. No scanner rule, directory allowlist, credential exception, history scope or production scan deadline changed.

A recorded script-suite failure also showed Bun's default five-second timeout expiring in the real vulnerable-import regression. That test now bounds its scanner subprocess at 60 seconds and allows 65 seconds including cleanup. Its required nonzero scanner result and imported-vulnerability assertion remain unchanged; timeout alone cannot pass it.

## Verification

- Cleanup regression proves generated output removal, source preservation, missing-output handling and refusal of unrecognized output and symlinks.
- Root runner suite: 14 tests passed.
- Real vulnerable-import regression passed with the unchanged vulnerability assertion.
- Script lint, formatting and type checking passed.
- Full `bun run check` passed after the final changes: canonical generation, all workspace/backend checks, nine browser cases, real container integration groups, migrations, builds, dependency scanning, full-worktree Gitleaks and complete-history Gitleaks.
- Broader Phase 2 requirements and live provider acceptance remain pending their subsequent plans; this repair does not mark Phase 2 complete.
