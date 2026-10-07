# Plan 01-20 prerequisite source quality repair

Completed: 2026-10-07. This records the prerequisite recovery needed to execute the configured source checks. It is neither another plan nor a plan completion summary; plan 01-20 remains owned by its executor.

## Scope and commits

- `d73564a` — `fix(01-20): satisfy authored TypeScript quality gates`: 10 authored schema, contract, email and generation source/test files.
- `af3e83a` — `fix(01-20): satisfy backend source quality gates`: 76 authored backend Go source/test files, including one new rollback regression test file.

Both commits used normal Git commits, individually enumerated source paths, and no hook bypass. Neither commit deletes a file. The runner's package.json, turbo.json, scripts/check.ts, scripts/check.test.ts and backend/taskfile.yml remain outside these commits. No lint configuration, dependency, tool pin, generated output or phase counter changed.

The parent explicitly authorized restoring only the uncommitted taskfile `migrate:check` body to its committed behavior: run `migrate` twice in the backend directory. Routing that task back into the new root migration-test stage recursively invoked the very test exercising the task. The taskfile restoration remains with the plan 01-20 executor's pending runner changes.

## Diagnostic reduction and implementation

The original unrestricted report contained 862 findings across 77 files; the preceding automatic fixes left 687. Fresh unrestricted repair reports reduced the findings to 385, 172, 41 and finally **zero**. The final invocation set both `--max-same-issues=0` and `--max-issues-per-linter=0`, so this result is not a truncated diagnostic count. The existing `gomodguard` deprecation notice remains a tool warning; its rule remains enabled.

The repair adds exported/package documentation, checked file-response assertions, descriptive policy constants, explicit enum handling, typed test context keys, non-shadowing local variables, keyed struct literals, protobuf getters, current supported APIs, and source formatting. Five production concerns were separated into small helpers: role-section configuration validation, telemetry endpoint validation, readiness probe/result processing, safe log-field extraction, and validation message/custom-error formatting. Their existing regression suites remain intact.

Transaction test helpers retain successful-callback commit behavior. Deferred rollback now reports real cleanup failures, ignores expected `pgx.ErrTxClosed`, and uses a bounded cancellation-independent context that preserves callback context values. This context belongs only to test transaction cleanup; it does not change process-role shutdown budgets. Three unit tests prove canceled callback cleanup, retained rollback causes and already-closed transaction behavior.

Local preview values now return fresh maps; the JSON inspection helper returns serialization/output failures instead of silently dropping them. OpenTelemetry's no-op deprecated log buffer option was removed after checking the pinned SDK source. The collector fixture uses supported pre-start file copying, and PostgreSQL's combined wait uses its supported deadline option.

## Preservation and automatic-fix review

- All **86 existing top-level Go test names** remain present; the only additions are the three rollback tests, for 89 total. Integration classification names remain unchanged.
- Automatic field reordering was checked against the original `e861be7` field order; positional struct literals were converted to named fields with their original semantic values, including same-type contract fixtures and failure tables.
- Two unsafe automatic edits changed captured context assignments from `=` to `:=`: the migrator's shared exit context and its ownership test's captured comparison context. Both were restored before acceptance. An AST comparison of assignments against the original source found no remaining conversion of those original assignments into local declarations.
- Cleanup, lifecycle and job-stop tests retain their original exact error-instance identity assertions. Automatic replacements with cause matching alone were reverted; narrowly annotated direct comparisons preserve the stronger idempotency contract.
- Actual HTTP response headers are copied through the supported recorder API; all original JSON content-type and schema assertions remain.
- Trace-state/baggage removal, collector privacy assertions, queue retry/replacement semantics, role ownership, dependency readiness, partial-startup cleanup, shutdown order and the shared migrator exit deadline remain exercised by fresh race-enabled tests.
- Isolated regeneration compares canonical OpenAPI, served OpenAPI, generated Go transport code and exported email HTML by bytes. The generator now carries each artifact path together with its bytes, and generator tests use explicit output paths or checked bounds instead of non-null assertions.

## Source annotations

The 86 changed authored files contain **61 new, explained source annotations**, with **70 named rule applications**. Eleven annotations are in production source and 50 are in tests. No configuration exclusion or blanket unqualified `nolint` was added. The configured `nolintlint` verifies that annotations name rules, explain their purpose and are used.

| Named rule | Applications | Scope and rationale |
| --- | ---: | --- |
| `gocognit` | 30 | Individual regression tests/helpers retain complete ordered protocols and failure assertions, including recursive schema validation, concurrency isolation, lifecycle, propagation, collector privacy and migration proofs. Production complexity findings were refactored. |
| `gocyclo` | 4 | Four complete integration proofs: tracestate over HTTP/Redis/OTLP, collector redaction, role binary startup and clean/repeated migrations. |
| `cyclop` | 5 | The same four integration functions, plus the transport test package's aggregate complexity from its recursive schema/generator proof. |
| `testpackage` | 14 | File-specific access to private fixtures, failure injection, provider/exporter boundaries and migration/cleanup helpers; includes the new private rollback boundary tests. |
| `errorlint` | 3 | Exact error-object identity is required by cleanup, lifecycle and job-stop idempotency tests, including nil on successful job shutdown. |
| `revive` | 6 | Preserve five existing exported compatibility type names and the existing logger-first contextual helper signature. |
| `nonamedreturns` | 3 | Deferred cleanup must update migrator exit status, partial-startup return state and migration cleanup errors. |
| `spancheck` | 1 | pgx transfers its query span from `TraceQueryStart` to `TraceQueryEnd`, where the span is always ended. |
| `gosec` | 1 | The lazily shared migrator exit context is canceled by the outer defer after database close and telemetry flush. |
| `fatcontext` | 2 | That shared exit-context assignment and its ownership test's captured context comparison. |
| `reassign` | 1 | A serial compatibility-logger test captures stdout and restores it during cleanup. |

## Fresh verification

All commands used the pinned Go 1.26.8, golangci-lint 2.14.0, staticcheck 2026.2.1 and Biome 2.5.15 tools already verified by plan 01-19.

| Check | Result |
| --- | --- |
| `go build ./...` | Passed |
| `go vet ./...` | Passed |
| `staticcheck -checks=all ./...` | Passed |
| Full unrestricted `golangci-lint run ./...` | Passed, zero diagnostics |
| `golangci-lint fmt --diff` | Passed, empty diff |
| Final `go test -race -count=1 ./...` | Passed all packages, including real PostgreSQL, Redis and collector paths |
| Focused fresh migrator, actual health HTTP and database race tests | Passed after context/header/task corrections |
| `bun run --filter '@flux/*' build` | Passed all three implemented workspaces, including email export |
| `bun run --filter '@flux/*' typecheck` | Passed all three implemented workspaces |
| `bun test` | 62 passed, zero failed across eight source/built test files |
| Scripts `tsc --project scripts/tsconfig.json --noEmit` | Passed |
| Biome check over repair-owned authored source | Passed without changes |
| `bun scripts/generate.ts --check` | Passed, exact bytes for all four artifact categories |
| `git diff --check` | Passed |

Temporary reports and logs are under ignored `tmp/`, including `repair-lint-acceptance.json`, `repair-race-acceptance.log`, `repair-goimports-diff.log`, `repair-bun-tests.log` and `repair-generate2.log`.

The excluded plan-owned `packages/emails/src/templates/welcome.test.tsx` still needs its import organization adjustment by the plan 01-20 executor: the `@react-email/components` import follows the Node imports. An accidental directory-wide formatter edit to that file was restored to its committed bytes, and the parent was notified. This file is outside the source repair commits. The root runner's actual `check:fast` acceptance and the later scanner/full-phase gates remain pending their owning plans.

## Self-check: PASSED

Both source commits exist, all 86 committed source paths exist, no committed source file was deleted, the acceptance lint report contains zero issues, and the final full Go race process exited successfully. The pending runner files remain unstaged for their owning executor. Phase state and plan counters were not advanced.
