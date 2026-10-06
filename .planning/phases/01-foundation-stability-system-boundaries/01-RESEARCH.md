# Phase 1 Research: Foundation Stability & System Boundaries

**Researched:** 2026-10-05
**Status:** Ready for planning
**Scope:** `PLAT-01`–`PLAT-08`, `SAFE-01`, `SAFE-08`

## Research Summary

Phase 1 should repair the existing scaffold through narrow ownership seams instead of replacing it. The current `cmd/flux` composition root constructs a broad `server.Server`; server construction couples PostgreSQL, Redis, HTTP, and Asynq; startup can leave resources open after partial failure; the signal path omits SIGTERM; health combines process and dependency state; generated contracts and runtime assets can drift; and workspace checks do not cover the Go backend. The first phase must turn these source-observed gaps into executable process and quality boundaries before product tables or routes are added.

The recommended implementation sequence is: pin and normalize the supported toolchain; introduce explicit resource groups and idempotent cleanup; add thin process roots; split liveness/readiness and shutdown behavior; make contract/assets deterministic; add OpenTelemetry and redaction; then enforce everything in CI with unit and container-backed integration tests. Preserve Echo, pgx, Tern, Redis, Asynq, Zod, and ts-rest/OpenAPI unless a concrete incompatibility appears.

## Locked Decisions from Context

- Use thin API, redirector, worker, and migrator composition roots.
- API replicas never start consumers or run migrations implicitly.
- Preserve Echo and extract incrementally; do not rewrite the repository tree.
- Liveness is process-only; readiness is role-specific and sanitized.
- SIGINT and SIGTERM use ordered, bounded drain and complete cleanup.
- Keep TS/Zod contract authoring, generate one deterministic OpenAPI artifact, and generate or verify Go transport boundaries from it.
- Embed assets or resolve them from an explicit validated root; never depend on the current working directory.
- Pin one supported toolchain and use Bun as the JavaScript workspace package manager.
- Use behavior-focused unit and integration gates instead of a blanket coverage percentage.
- OpenTelemetry/OTLP is the vendor-neutral instrumentation boundary.
- Configuration packages return typed errors and preserve external naming compatibility during stabilization.

## Current-State Findings

### Process and Resource Ownership

- `apps/backend/cmd/flux/main.go` is the only composition root.
- `apps/backend/internal/server/server.go` constructs the database, Redis, Asynq, HTTP server, and other services together.
- Every HTTP replica therefore owns worker consumers, preventing independent API/worker scaling.
- Resource initialization has early-return paths that do not close previously created clients.
- Shutdown closes resources in an order that can remove database access before workers finish, and it can skip later cleanup after an earlier error.
- Package-global logger, Clerk, and email clients obscure ownership and interfere with isolated tests.

### Health and Lifecycle

- `GET /status` probes PostgreSQL and Redis sequentially and exposes dependency errors.
- Redis failure does not consistently make aggregate status unhealthy.
- Request cancellation and configured health timeouts are not consistently applied.
- Only `os.Interrupt` is registered; ordinary container SIGTERM can bypass graceful shutdown.
- Liveness and readiness are not separated by process role.

### Contracts and Assets

- TypeScript Zod/ts-rest contracts generate OpenAPI JSON for the package and backend static directory.
- Generation uses independent asynchronous writes and can log rather than fail on an output error.
- Served contract behavior and Go handlers are not checked against each other; health 503 behavior is missing from the documented response contract.
- `@flux/openapi/contracts` contains a singular/plural runtime export mismatch.
- OpenAPI HTML/JSON and email templates depend on relative paths and the process working directory.

### Tests, CI, and Toolchain

- Testcontainers and helpers exist, but no application `*_test.go` or TypeScript test suites exercise product or lifecycle behavior.
- Root Turbo commands name lint/typecheck/format tasks that package manifests do not consistently implement; Turbo does not build the Go backend.
- Bun is declared as package manager, while npm and package-local Bun lockfiles create multiple sources of dependency resolution.
- Go, Node/Bun, service images, linter, and task tooling are not consistently pinned across a CI system because no project CI workflow exists.

### Observability and Configuration

- Zerolog and direct New Relic integrations exist across HTTP, pgx, and Redis.
- The specification requires OpenTelemetry for logs/traces/metrics and cross-process correlation.
- Request logging includes full URI values and has no explicit query redaction policy.
- Configuration returns an error but calls fatal logging internally; callers cannot test or control failure behavior.
- Existing environment naming uses a `FLUX_` prefix and dot-based nesting. Phase 1 should not silently break deployed names.

## Recommended Architecture

### Composition Roots

Create thin main packages for four roles, backed by internal constructors that return owned resources plus cleanup:

| Role | Required dependencies in Phase 1 | Readiness policy | Must not do |
|------|----------------------------------|------------------|-------------|
| API | configuration, logger/telemetry, PostgreSQL, Redis if required by current routes, HTTP | required dependencies reachable | start workers or run migrations |
| Redirector | configuration, logger/telemetry, HTTP shell; later Redis/PostgreSQL | only dependencies actually used in this phase | implement Phase 3 redirect behavior |
| Worker | configuration, logger/telemetry, Redis/Asynq, email adapter as configured | queue dependency and required adapters ready | open public API routes |
| Migrator | configuration, logger, PostgreSQL, embedded migrations | command exit status | stay resident or start HTTP |

Use a small ownership abstraction, such as a cleanup stack or role-specific application type, that records closers immediately after successful construction and unwinds them in reverse dependency order. Cleanup must attempt all closers and combine errors. Avoid rebuilding the existing `Server` container under another name.

### Lifecycle Contract

1. Parse and validate configuration without exiting inside library packages.
2. Initialize logger/telemetry and register cleanup.
3. Initialize role dependencies one at a time and register each cleanup immediately.
4. Build route/job adapters after dependencies exist.
5. Start serving/consuming.
6. On SIGINT/SIGTERM, mark readiness false, stop intake, drain active work within a configured deadline, then close dependencies in reverse order.
7. Return or log one joined error after all cleanup attempts.

### Health Contract

- `/live` returns a small stable payload and does not probe external dependencies.
- `/ready` uses the request context plus one bounded total deadline and checks required dependencies concurrently when that improves the bound.
- Each process defines its own readiness dependencies in construction; the handler does not infer them from global clients.
- Public payloads expose component name and coarse state, not raw driver/provider errors.
- Detailed errors are logged once with correlation and redaction.

### Contract and Asset Pipeline

- Retain `packages/zod` and `packages/openapi` as the authoring layer for Phase 1.
- Make generation synchronous/awaited and atomic enough that partial output cannot look successful.
- Produce one canonical OpenAPI artifact, copy or embed it deterministically for serving, and validate Go transport responses against it through generated types or contract tests.
- Add all actual health statuses and payloads to the contract.
- Fix the contracts runtime export path and fail package tests on resolution errors.
- Embed stable OpenAPI and email assets with Go `embed` where ownership and update frequency fit; otherwise expose a validated explicit asset directory.

### Observability Boundary

- Initialize an OpenTelemetry SDK per process and export OTLP to a collector-configured endpoint.
- Propagate request and correlation IDs from HTTP into worker/job metadata.
- Define a redaction function used by structured logging and error attributes.
- Avoid raw URI query strings, credentials, request bodies, event secrets, raw IPs, and arbitrary error payloads in telemetry.
- Use bounded metric labels: process role, route template, outcome/status class, dependency name. Never label by workspace, link, user, key, raw host, or error message.
- Keep direct New Relic export only behind the collector/export boundary or remove it after equivalent signals exist.

## Toolchain and CI Guidance

- Pin the Go toolchain and checked tool versions; verify exact supported patch versions at implementation time.
- Use one root Bun lockfile and explicit workspace scripts. Remove `package-lock.json` and package-local lockfile ownership only after clean install/build comparison.
- Make the root `check` path invoke Go format/vet/lint/test/build, TypeScript format/lint/typecheck/test/build, OpenAPI/email generation drift checks, and migrations.
- Use pinned PostgreSQL and Redis images for local/test parity.
- Add GitHub Actions with least permissions, dependency caching keyed by lockfiles, explicit timeouts, and no secret interpolation into logs.
- Run Go vulnerability analysis and JavaScript dependency audit against the supported lockfiles. Use a secret scanner in history/worktree mode that redacts findings in normal output.

## Implementation Order and Dependencies

1. **Toolchain and check surface** — define supported commands before moving files so every later change has a reliable feedback loop.
2. **Resource ownership primitives** — add typed construction errors, cleanup aggregation, and focused unit tests.
3. **Composition roots** — split API/worker/migrator and add the redirector shell; keep domain behavior unchanged.
4. **Lifecycle and health** — add role readiness, `/live`, `/ready`, SIGTERM, bounded drains, partial-startup tests.
5. **Contract and assets** — make generation deterministic, fix exports/status coverage, embed/configure assets, add drift checks.
6. **OpenTelemetry and redaction** — add vendor-neutral initialization after role ownership is stable so each process gets correct resource attributes.
7. **CI hardening** — enforce all checks, container tests, dependency scanning, and secret scanning.

Plans can overlap only where files do not conflict. Composition-root/lifecycle work should land before observability wiring; contract/asset work can proceed alongside ownership work after the shared check commands are established.

## Validation Architecture

### Fast Feedback Layer

- Unit-test configuration error returns, cleanup stack order/error joining, health aggregation, redaction, request factory isolation, validation fallback behavior, and role dependency declarations.
- Type-check and package-test the TypeScript contract packages.
- Verify generated artifacts by regenerating into the workspace and requiring a clean diff or checksum match.

### Integration Layer

- Use Testcontainers with pinned PostgreSQL and Redis images.
- Run migrations from an empty database and assert the expected schema version; run again to prove idempotent no-op behavior.
- Start each process role with only its declared required dependencies.
- Inject failures at every initialization stage and assert all earlier resources are closed.
- Send SIGTERM while HTTP requests or Asynq work are active and assert readiness drops, new intake stops, active work drains within the deadline, and remaining resources close.
- Run binaries from a temporary working directory and assert docs/email assets remain accessible.
- Assert public readiness never contains connection strings, hostnames from raw driver errors, credentials, or stack traces.

### Contract and Security Layer

- Validate `/live` and `/ready` responses and all documented statuses against the generated OpenAPI artifact.
- Import `@flux/openapi/contracts` at runtime in a package test.
- Scan logs/traces produced by controlled secret-bearing inputs and assert secret markers are absent.
- Run dependency/vulnerability checks and secret scanning in CI with explicit allowlist handling for verified test fixtures only.

### Required Commands

The planner should define one canonical root command for local verification and use the same primitives in CI. At minimum, plans must name exact commands for Go formatting/vetting/testing/race testing/building, Bun install/typecheck/test/build/generation, migration integration, generation drift, dependency scanning, and secret scanning. If a tool is not installed in the repository, the plan must add and pin it before relying on the command.

## Planning Risks

- **Over-broad refactor:** splitting every package at once would hide behavioral regressions. Plan by ownership seam and require tests before moving callers.
- **False independent roles:** multiple `main` packages that all call the same coupled constructor do not satisfy `PLAT-01`.
- **Readiness deadlock:** readiness must not share request quotas or block for the sum of sequential dependency timeouts.
- **Contract dual authority:** adding Go annotations without retiring manual TS synchronization increases drift.
- **Telemetry cardinality leak:** adding workspace/link IDs or full errors to attributes creates cost and privacy failures.
- **Toolchain upgrade blast radius:** dependency/tool upgrades should be isolated early and verified before lifecycle refactors.
- **Git unavailable:** this workspace has an empty read-only `.git`; execution cannot satisfy commit-based completion until the repository metadata is restored outside the sandbox.

## Planner Checklist

- Every plan cites its requirement IDs and relevant `D-xx` decisions.
- Every task names concrete files, reads current implementation first, and includes behavior-based acceptance criteria.
- No plan introduces product schema, tenancy, link, redirect behavior, or UI work.
- The four process roles are real ownership boundaries, not aliases over the current all-in-one server.
- Contract generation, asset behavior, and root CI commands are testable from a clean checkout/environment.
- Security threat models cover public health leakage, secret/PII telemetry, CI token permissions, dependency supply chain, and unsafe cleanup diagnostics.

## Package Legitimacy Audit

The project-level stack research verified these packages against their official documentation and release repositories on 2026-10-05. The planner may include pinned installation tasks, with an implementation-time patch recheck because versions are time-sensitive.

| Package | Legitimacy | Planned use | Evidence |
|---------|------------|-------------|----------|
| `go.opentelemetry.io/otel` `v1.47.0` | Verified official OpenTelemetry Go module | API, context propagation, tracer/meter access | `.planning/research/STACK.md` and official OpenTelemetry Go docs/releases |
| `go.opentelemetry.io/otel/sdk` `v1.47.0` | Verified official OpenTelemetry Go SDK module | Per-process tracer/meter providers and resource attributes | `.planning/research/STACK.md` and official OpenTelemetry Go releases |
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace` `v1.47.0` | Verified official exporter module | OTLP trace export to the collector | `.planning/research/STACK.md` install guidance and official module docs |
| `go.opentelemetry.io/otel/exporters/otlp/otlpmetric` aligned `v1.47.0` family | Verified official exporter family; exact submodule/transport must be confirmed at implementation | OTLP metric export to the collector | Official OpenTelemetry Go exporter documentation; align the entire `v1.47` family |
| `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` `v0.71.0` | Verified official contrib instrumentation module | Standard HTTP client/server instrumentation where Echo-specific middleware is not used | `.planning/research/STACK.md` and official contrib releases |
| `github.com/oapi-codegen/oapi-codegen/v2` `v2.7.2` | Verified official oapi-codegen module | Generate/verify Go Echo transport types from canonical OpenAPI | `.planning/research/STACK.md` and official project releases |
| `github.com/testcontainers/testcontainers-go` `v0.42.0` | Already present and verified official module | PostgreSQL/Redis integration and lifecycle tests | Existing `apps/backend/go.mod` and `.planning/research/STACK.md` |

Do not invent an Echo OpenTelemetry package path. During implementation, prefer documented standard `otelhttp` adapters or verify any Echo contrib middleware exists in the official OpenTelemetry contrib repository before adding it. Installation tasks must run `go list -m -versions` or equivalent module resolution and keep all OTel core/exporter modules on one aligned release family.

---

*Phase research completed: 2026-10-05*
