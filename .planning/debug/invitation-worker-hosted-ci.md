---
status: unreproduced
scope: Phase 2 execution prerequisite after plan 27, before plan 28
historical-head: 4302444
hosted-run: 38154093371
---

# Invitation worker hosted CI investigation

The historical hosted failure remains unexplained. Fresh isolated, exact app-group and original canonical full integration commands passed without changing source, fixtures, assertions, deadlines, retries or runner behavior. These passes establish current local verification; they do not establish a root cause or a repair of the historical failure.

## Historical evidence

- Hosted [run 38154093371](https://github.com/6sLOGAN78/Flux/actions/runs/38154093371), at `4302444`, failed `backend:integration:internal/app` with the sanitized name `TestInvitationWorkerRoleOwnership`.
- The integration app stage began at `2026-10-11T16:16:41.9439054Z` and the failure was reported at `2026-10-11T16:18:07.8323822Z`. The job's sanitized output retained no assertion or failing subtest detail. The run artifacts API returned no artifacts. No assertion, race, timing, container or resource cause is inferred from the elapsed time.
- The previous hosted plan 26 run `38143539391` at `12353f3` succeeded. Comparing `12353f3..4302444` found no changes in `internal/app`, `internal/lib/job`, `internal/lifecycle`, `internal/testing`, `scripts/check.ts`, `go.mod` or `go.sum`.
- Earlier tracing failures are distinct evidence and do not explain this ownership failure.

## Investigation and boundaries

Followed the existing GSD execution prerequisite and systematic debugging protocol. Read the actual worker test, resource ownership and serial cleanup implementation, job stop/drain behavior, and plan 23/27 summaries before considering any change.

The ownership suite covers five real partial-startup stages (database, Redis, email, consumer, consumer start), local readiness against actual owned dependencies, and active invitation delivery through shutdown. The readiness fixture stops intake and drains Asynq before deliberately closing Redis. The active-delivery proof retains the original 300 ms caller deadline and one-second assertion ceiling, keeps PostgreSQL/Redis usable beneath the blocked send, then verifies Delivered/ciphertext erasure and resource/listener closure after release.

The first diagnostic tested the unchanged isolated suite. The second tested the possible app-group interaction boundary using the exact registered native app command. Neither reproduced a failure, so no corrective hypothesis was confirmed and no speculative production or fixture change was made. There is no new RED/GREEN cycle or claimed fix.

Raw native output was captured in mode-0600 files inside mode-0700 OS temporary directories. Normal command umask was preserved so the nonroot collector could read its ordinary fixture configuration. Only test names, counts, durations and booleans were extracted; raw output, credentials, tokens, email addresses, provider responses and configuration values were not published or committed.

## Fresh verification

Verification ran on unchanged `4302444` on 2026-10-11 UTC, with runtime Go 1.26.9 and the existing Node 22.23.3/Bun 1.3.14 toolchain. Repository pins stayed unchanged.

1. Native `go test -race -count=1 -timeout=12m -json -run '^TestInvitationWorkerRoleOwnership$' ./internal/app` from `apps/backend`: exit 0, **1/1 top-level test and 7/7 direct subtests**, 4.08 seconds test duration. Active drain completed in 0.47 seconds. No failed/skipped events, race marker or malformed JSON.
2. Exact native registered app-group command, retaining `-race -count=1 -timeout=12m -json` and the original 13-name regular expression: exit 0, **13/13**, 84.258 seconds package duration. Ownership passed in 3.38 seconds. No failed/skipped events, race marker or malformed JSON.
3. Original unmodified `bun run test:integration`: exit 0, **20 registered tests across all 6 groups**, ending in `Checks passed`. Native inventory validation, module/tool prerequisites, race detection, positive named pass checks and original timeouts remained enabled. This is the final canonical integration gate, not a diagnostic runner substitute.
4. `git diff --check`: passed before report creation; report formatting and exact staging are checked at commit.

Full browser, unit, lint, typecheck, generation, production build and scans were not repeated because this investigation changed no executable source or fixture. Their earlier plan 27 evidence remains in `02-27-SUMMARY.md`; this report makes no fresh claim for those gates. No real email/provider writes or live configuration inspection occurred.

## Remaining uncertainty and next diagnostic

The original hosted job remains failed. Local passes cannot show which assertion failed there or establish hosted recovery. If the ownership failure recurs, capture the exact first failing native command privately under the same fixture permissions and original limits, then extract only safe assertion identifiers, numeric/boolean boundary facts and relevant stack locations. Preserve the stop/drain-before-close order and active-work dependency protection. Do not infer a cause from older telemetry failures or introduce retries, increased timeouts, assertion waivers or scanner exceptions.

This artifact records a prerequisite investigation only. It does not mark plan 28 or any requirement complete, advance STATE/ROADMAP, push a commit, or claim phase acceptance. The unrelated `.serena/project.yml` modification remains unstaged and untouched.

## Self-Check: PASSED

- The original ownership test and runner were not modified.
- Both private native reports parse successfully and contain the counts recorded above.
- The final original canonical integration process exited 0 after all six registered groups.
- Only this diagnosis report is staged for the documentation commit; no tracked deletion is included.
