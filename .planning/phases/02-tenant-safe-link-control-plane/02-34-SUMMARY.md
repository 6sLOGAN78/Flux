---
phase: 02-tenant-safe-link-control-plane
plan: "34"
subsystem: testing
tags: [clerk, resend, playwright, authentication, private-evidence]
requires:
  - phase: 02-33
    provides: Complete native browser collection and frozen quality gates
provides:
  - Executable private provider prerequisite, automated acceptance and strict final acceptance commands
  - Twelve runnable provider browser definitions and an isolated production Go/PostgreSQL/Redis harness
  - Closed 22-row evidence protocol with explicit human-only pending steps
affects: [02-35]
tech-stack:
  added: []
  patterns: [native annotation evidence, private run-specific artifacts, owned provider egress fault]
key-files:
  created:
    - apps/frontend/tests/provider.config.ts
    - apps/frontend/tests/acceptance.test.ts
    - apps/frontend/tests/provider-auth.spec.ts
    - apps/frontend/tests/provider-session.spec.ts
    - apps/frontend/tests/provider-invitation.spec.ts
  modified:
    - apps/frontend/playwright.config.ts
    - apps/frontend/package.json
    - package.json
    - scripts/check.ts
    - scripts/check.test.ts
    - docs/development.md
key-decisions:
  - Keep actual provider execution separate from the unchanged 65-case local project and fail missing configuration before collection.
  - Bind the closed acceptance rows to positive native execution and private run/configuration evidence; only six enumerated human rows may remain pending in automated mode.
  - Exercise production provider timeout by detaching only the test-owned API egress network while PostgreSQL and Redis remain healthy.
requirements-completed: []
requirement-slices: [TEN-01, TEN-02, TEN-03, TEN-06, SAFE-02, SAFE-04]
duration: 41min 02s
duration-basis: Sum of captured active windows; stalled interruption and metadata close-out excluded
completed: 2026-10-12
---

# Phase 2 Plan 34: Private Provider Acceptance Preparation Summary

**Three safe provider validation commands, twelve real-provider browser definitions and a private native evidence protocol are prepared for the plan35 configuration checkpoint.**

## Performance

- Tasks: **2/2**. Source/documentation paths changed: **11**, all declared. Five source files were created across RED/GREEN commits.
- Captured active windows: **2026-10-12T20:36:24Z–20:54:00Z** (**17min 36s**) and **2026-10-12T21:02:17Z–21:25:43Z** (**23min 26s**), totaling **41min 02s**.
- The stalled interval **20:54:00Z–21:02:17Z** and metadata close-out are excluded. No tests are attributed to that stalled interval.

## Accomplishments

- Root `provider:check`, `provider:automated:check`, `provider:acceptance:check` and `test:e2e:provider` now have executable implementations. The first existing `test:e2e` remains intact. Missing prerequisites fail nonzero before provider collection or startup and disclose safe variable names only.
- Configuration checks reuse exact production environment names, validate matching development publishable-key/issuer/authorized-origin shapes, identical frontend/backend secret configuration, canonical decoded AES keys and cursor key, explicitly approved distinct test accounts/recipients, and an operator-owned private directory outside the repository. An actual configured run compares authenticated BAPI JWKS with the configured issuer JWKS under bounded requests.
- The provider project defines actual email/password signup and login controls, verification/recovery challenge controls, Google/GitHub redirects, token refresh and real expiry, revocation/signout, seven-day session policy, bounded quota/latency sampling, fail-closed provider timeout, cookie/key inspection, and the named-owner/invitation/last-workspace/Links/membership-loss journey. It uses no SDK login shortcut to claim factor proof.
- Its disposable harness compiles production API, migrator and worker binaries, reuses the existing exact PostgreSQL/Redis image pins, and owns its containers, networks and private runtime files. The API has normal authentication and real dependencies. Timeout testing detaches only its egress network; database/cache readiness is checked before requiring 503 and recovery. No endpoint substitution, host firewall change or production authentication bypass was added.
- The invitation journey prepares old-key encrypted work, rotates the active key while retaining decryption keys, starts the production worker for one approved-recipient delivery, checks terminal ciphertext erasure, and verifies new-key writes using a second queued invitation revoked before another send. Actual inbox input is read only from the private directory during the later authorized run.
- Every required evidence row must appear exactly once and reference positive non-skipped native test evidence. Automated mode permits pending status only for the six exact human rows with closed reasons and canonical manual steps. Strict final mode rejects all pending rows. Successful runs produce unique private 0600 proof/record files in an owned nonsymlink 0700 directory; reports contain no tokens, email addresses or screenshots/traces.
- Development instructions document private inputs, operator recipient authorization, fixed commands, bounded development-instance load, private inbox handoff, proof promotion and remaining human steps. A digest binds the record to its private proof; this is execution-integrity checking, not a cryptographic guarantee against an operator rewriting both files.

## Task Commits

1. **Task 1: Specify complete private provider evidence boundaries** — `c8fdf59` (test).
2. **Task 2: Prepare provider configuration, executable journeys and strict evidence checks** — `8d125d9` (feat).

Exact staging and normal commits were used. No tracked files were deleted. Unrelated `.serena/project.yml`, `README.md` and `.planning/quick/` changes were preserved.

## Verification

- Genuine RED at the new validator boundary executed **5 tests: 4 failed, 1 passed**. The initial runnable boundary accepted missing/unknown/incomplete evidence and unsupported proof; these failures represent absent new validation, not a pre-existing product defect or missing import/script failure.
- Focused provider verification passed **15/15** after implementation: **11 validator cases** plus **4 root/provider protocol cases**. Positive coherent configuration and proof fixtures precede controlled missing/failed/unknown/duplicate/pending, skipped/retried/local/SDK evidence, stale run/digest/instance, mismatched JWKS and malformed-key/unauthorized-recipient mutations. Synthetic configuration fixtures do not contact real providers.
- The writer protocol case executes a real native Playwright **Node-only test fixture**, serializes its actual annotations into private artifacts, validates their digest/run/config binding and 0600 modes, and rejects missing/duplicate/unknown annotations, failed native results and an unsafe private directory without creating additional files. Its annotations are explicitly fixture data, not live provider acceptance. The writer also independently validates the native report before serialization.
- Native collection retained **65 local browser definitions** and collected **12 provider definitions** separately. Provider definition collection is preparation evidence only. Clean-environment probes for all three validation modes fail safely; no missing configuration is converted to a skip or zero-pass success.
- Original **CI=true `bun run check` passed**, returning actual exit **0** and “Checks passed”: **65 local browser cases**, **109 race-enabled Go unit tests across 16 packages**, **21 integration top-level tests across 6 groups**, **42 frontend units**, **6 email tests**, **15 OpenAPI tests**, schema/script/tool suites, all format/lint/type checks, workspace/backend builds, migrations, all **five** generated artifacts and native security scans.
- After that full run, the narrow writer protocol guard was added within owned runner paths. All affected original script suites then passed **69/69 across four files**, including native execution, real frozen-root generation drift and pinned scanner/history regressions. Script TypeScript and pinned Biome lint/format passed. The local browser configuration and production feature code did not change after the successful full run.
- After the final local browser run, the original production frontend build passed. Exported `cleanFrontendBuild` validated the owned nonsymlink production manifest and removed output; `apps/frontend/build` is absent.
- An offline capability probe compiled the static production API and executed it as uid65534 in the existing pinned PostgreSQL image with **network disabled** and no configuration. It reached the expected configuration failure, proving executable/image compatibility; its owned temporary binary and ephemeral container were removed. The image CA bundle check also passed. This does not prove the configured harness startup or live provider journeys.
- Original **CI=true `bun run scan` passed after the GREEN commit**: all backend roles/tests had no vulnerable imported Go packages, Bun dependencies were clean, and complete worktree/full history secret scans were clean. The existing unused-module inventory advisory remained visible. Final summary/tracking commits receive another original scan during close-out.
- `git diff --check` passed. No dependency, lockfile, runtime pin, generated artifact, production endpoint or scanner exception changed.

## Decisions Made

Local and provider projects remain explicitly separate. Missing provider configuration fails before native discovery. Only real configured runs can produce provider acceptance; local fixtures never satisfy provider factors.

The six allowed manual rows are `email_verification`, `password_recovery`, `google_oauth`, `github_oauth`, `production_cookie_inspection` and `provider_key_rotation`. Their automated portions must execute successfully, while inaccessible inbox/social/production/dashboard actions remain explicitly incomplete until human verification.

## Deviations from Plan

**[Rule 2 - Missing critical] Bind evidence serialization to actual native annotations and secure private files.** The root writer now revalidates positive native execution before serializing closed rows. A real native protocol fixture proves serialization and denials independently of unavailable provider credentials. This adds no live acceptance claim.

**[Rule 3 - Blocking] Explicitly match browser `.spec.ts` files.** Adding the validator unit `.test.ts` exposed Playwright's default unit-file discovery and produced Node TAP instead of the expected native JSON list. Explicit local `.spec.ts` selection restores the same 65 cases and keeps provider definitions separately selected. This collection setup failure is not business RED.

The parent approved the disposable production API/worker harness and separate owned egress network after source inspection showed plan35 explicitly requires a provider timeout fault. Existing native browser/global/stage deadlines, production configuration authorities and local classifications were preserved; no additional ownership or infrastructure dependency was introduced.

## Issues Encountered

Pinned lint required explicit fixture-parser assertions and import ordering; those were corrected before final affected-suite verification. An interrupted period produced no test evidence and is excluded from duration. No private capture, browser retry, timeout increase, dropped case or scanner waiver was required.

## Known Stubs

None introduced. Actual provider execution is deliberately gated on plan35 operator inputs. No canned passed acceptance artifact was created.

## Threat Flags

| Flag | File | Description |
|---|---|---|
| threat_flag: test_control_network | scripts/check.ts | Test-only loopback control requires a per-run random bearer and closed commands; it acts only on the harness's owned Docker resources and UUID-scoped assertions. |
| threat_flag: private_evidence_files | provider.config.ts / scripts/check.ts | New private proof/inbox reads and artifact writes require an owned nonsymlink directory outside the repository, restricted permissions and bounded input sizes; public diagnostics remain sanitized. |

## Remaining Acceptance / User Setup Required

Plan **02-35** must obtain operator-authorized development Clerk/Resend configuration, actual strategy/session dashboard settings, three approved test accounts/inboxes, two invitation encryption keys, managed/public origins and an external private directory. No live keys, private provider configuration or recipient inbox were accessed, and no provider emails were sent in plan34.

The configured production harness, twelve live provider browser cases, actual invitation delivery and complete session/expiry/fault behavior have **not run yet**. Once configuration is supplied privately, the executor runs prerequisite and provider commands, promotes the matching sanitized record/private proof described in `docs/development.md`, requires automated acceptance, completes the six enumerated human steps and requires strict final acceptance. Broad requirement completion remains unset until this actual evidence and the phase audit pass.

## Self-Check: PASSED

Created provider configuration/unit/auth/session/invitation files exist. Commits `c8fdf59` and `8d125d9` exist; neither deletes tracked files. The production build output is absent, and unrelated Serena/README/quick changes are preserved. No actual provider acceptance artifact or private test credentials were written to the repository.

## Next Phase Readiness

Plan34 preparation is complete. Plan35's human configuration checkpoint and actual external acceptance remain outstanding. No phase completion, transition, push or plan35 execution is included in this plan.
