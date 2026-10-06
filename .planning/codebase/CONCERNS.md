# Codebase Concerns

**Analysis Date:** 2026-10-05

## Tech Debt

**Infrastructure boilerplate is the implemented scope:**
- Issue: The implemented router exposes `/status`, `/static`, and `/docs`; `/api/v1` is an empty group. `Repositories` is empty and the only SQL migration contains comments. Product functionality described in `spec.md` is requirements, not implementation.
- Files: `apps/backend/internal/router/router.go`, `apps/backend/internal/router/system.go`, `apps/backend/internal/repository/repositories.go`, `apps/backend/internal/database/migrations/001_setup.sql`, `spec.md`.
- Impact: Planning cannot assume link, workspace, analytics, domain, or attribution behavior exists.
- Fix approach: Keep reusable infrastructure and add vertical product slices with schema, repositories, services, authenticated routes, contracts, and tests.

**Configuration exposes controls that runtime does not apply:**
- Issue: Database pool size and lifetime settings are required but never assigned to `pgxpool.Config`. Health-check enablement, interval, timeout, and selected checks are not used by the health handler. `SlowQueryThreshold` and `AppLogForwardingEnabled` are declared without corresponding runtime use.
- Files: `apps/backend/internal/config/config.go`, `apps/backend/internal/config/observability.go`, `apps/backend/internal/database/database.go`, `apps/backend/internal/handler/health.go`, `apps/backend/internal/logger/logger.go`.
- Impact: Operators cannot tune pool capacity or health behavior using the declared interface; logging/forwarding flags do not implement their advertised control.
- Fix approach: Assign supported pool settings with explicit units and semantics, consume health-check configuration, and either implement logging controls or remove them.

**Build and quality commands lack complete workspace coverage:**
- Issue: Root scripts invoke Turbo lint, formatting, and typecheck tasks, while inspected package manifests declare build/dev/clean or email export tasks instead. The Go backend has a Taskfile, not a JavaScript package manifest, so root Turbo build does not build the API.
- Files: `package.json`, `turbo.json`, `packages/zod/package.json`, `packages/openapi/package.json`, `packages/emails/package.json`, `apps/backend/taskfile.yml`.
- Impact: Root commands alone cannot establish backend build or workspace lint/typecheck coverage.
- Fix approach: Add explicit package quality scripts and a documented backend build/check entry point; verify orchestration actually schedules each expected package.

**Configuration loader terminates the process instead of returning errors:**
- Issue: `LoadConfig` has an error return but calls `logger.Fatal()` on load, unmarshal, and validation failures.
- Files: `apps/backend/internal/config/config.go`, `apps/backend/cmd/flux/main.go`.
- Impact: Callers cannot recover, test failure paths conventionally, or control cleanup and startup diagnostics.
- Fix approach: Return wrapped errors from the configuration layer; make process-exit decisions in `main`.

## Known Bugs

Findings below follow directly from source inspection. No running service or test-suite reproduction is claimed.

**Redis failure still reports overall healthy:**
- Symptoms: `/status` can return HTTP 200 and overall `healthy` while `checks.redis.status` is `unhealthy`.
- Files: `apps/backend/internal/handler/health.go` (`CheckHealth`).
- Trigger: PostgreSQL ping succeeds and Redis ping fails. Only the PostgreSQL error branch sets `isHealthy = false`.
- Workaround: Consumers must inspect each check; set aggregate health consistently with the intended Redis dependency policy.

**Termination signal bypasses graceful shutdown:**
- Symptoms: The graceful cleanup path listens only for `os.Interrupt`.
- Files: `apps/backend/cmd/flux/main.go`.
- Trigger: A supervisor sends SIGTERM, the common termination signal in service/container environments.
- Workaround: Use SIGINT until signal handling includes SIGTERM; register both and test draining behavior.

**OpenAPI contract package runtime export points to the wrong directory:**
- Symptoms: `@flux/openapi/contracts` advertises types under `dist/contracts/index.d.ts` but runtime JavaScript under `dist/contract/index.js`.
- Files: `packages/openapi/package.json`, `packages/openapi/src/contracts/index.ts`, `packages/openapi/tsconfig.json`.
- Trigger: A consumer resolves the declared runtime export after compilation; source directory is plural `contracts`.
- Workaround: Align runtime export with `dist/contracts/index.js`.

**Validation helpers can panic for legitimate error shapes:**
- Symptoms: Binding errors are parsed using unchecked string-split indexes; validation errors other than `validator.ValidationErrors` are asserted directly to `CustomValidationErrors`.
- Files: `apps/backend/internal/validation/utils.go` (`BindAndValidate`, `extractValidationErrors`).
- Trigger: A custom binder returns an error without the expected comma/message format, or a `Validate()` implementation returns an ordinary error.
- Workaround: Match typed errors with checked assertions and provide a safe fallback message. The current health/docs routes do not invoke these generic helpers.

**Email template uses relative web links:**
- Symptoms: The welcome email contains `/dashboard` and `/support` links without an application origin.
- Files: `packages/emails/src/templates/welcome.tsx`, `apps/backend/templates/emails/welcome.html`.
- Trigger: A recipient follows a link from a mail client that has no application base URL.
- Workaround: Generate absolute URLs from explicit public application configuration and re-export the HTML template.

## Security Considerations

**Public health endpoint returns dependency errors verbatim:**
- Risk: Dependency failures can reveal hostnames, connection details, and infrastructure diagnostics to unauthenticated callers; actual error contents depend on the failure.
- Files: `apps/backend/internal/router/system.go`, `apps/backend/internal/handler/health.go`.
- Current mitigation: General API errors use a separate structured error handler, but health responses explicitly include `err.Error()`.
- Recommendations: Return coarse public health states and keep detailed dependency errors in server logs or restricted diagnostics.

**Auth infrastructure has no route-level policy or tenant isolation:**
- Risk: Future product routes can be registered without authentication or workspace authorization. Auth middleware populates claims but does not enforce workspace ownership.
- Files: `apps/backend/internal/router/router.go`, `apps/backend/internal/middleware/auth.go`, `apps/backend/internal/service/auth.go`, `apps/backend/internal/model/base.go`.
- Current mitigation: Clerk verification middleware exists; current registered endpoints are system endpoints, so this is a feature-boundary gap rather than a demonstrated protected-data exposure.
- Recommendations: Apply authentication to product route groups and explicit workspace permission checks in service/repository boundaries before adding tenant data.

**Proxy trust and log redaction are implicit:**
- Risk: Rate limiting and logs rely on `c.RealIP()` without explicit proxy trust configuration. Request logs include the full URI, so future query parameters containing sensitive data would enter logs.
- Files: `apps/backend/internal/router/router.go`, `apps/backend/internal/middleware/global.go`, `apps/backend/internal/middleware/tracing.go`.
- Current mitigation: Rate limiting and structured logging exist; no observed secret-bearing application endpoint is registered.
- Recommendations: Configure IP extraction for the actual proxy topology and redact sensitive query parameters before adding integrations or tracking endpoints.

## Performance Bottlenecks

**Health probes run sequential dependency round trips:**
- Problem: A health request can consume approximately two five-second timeout windows when both dependencies stall.
- Files: `apps/backend/internal/handler/health.go`.
- Cause: PostgreSQL and Redis are pinged sequentially, each with a fresh `context.Background()` timeout; request cancellation and configured health timeout are ignored.
- Improvement path: Derive context from the request, use a bounded total deadline, separate liveness/readiness semantics, and consider concurrent dependency checks.

**Email templates are parsed for every send:**
- Problem: Each welcome task opens and parses a template from disk before delivery.
- Files: `apps/backend/internal/lib/email/client.go`, `apps/backend/internal/lib/job/handlers.go`.
- Cause: `template.ParseFiles` runs inside `SendEmail` rather than template initialization.
- Improvement path: Parse/embed templates during initialization and reuse immutable templates; measure throughput before further tuning.

## Fragile Areas

**Generic request wrappers capture a reusable mutable request:**
- Files: `apps/backend/internal/handler/base.go`, `apps/backend/internal/validation/utils.go`.
- Why fragile: `Handle`, `HandleFile`, and `HandleNoContent` capture the supplied `req` at route setup and reuse it for each call. Pointer request values are mutated by binding; concurrent requests can race and omitted fields can retain another request's values.
- Safe modification: Allocate a fresh request per invocation using a request factory or a type-specific constructor. These wrappers have no registered domain-route usage yet, so the hazard is latent.
- Test coverage: No request isolation or race tests are present.

**Service initialization and cleanup are coupled:**
- Files: `apps/backend/internal/server/server.go`, `apps/backend/internal/database/database.go`, `apps/backend/internal/lib/job/job.go`, `apps/backend/cmd/flux/main.go`.
- Why fragile: HTTP initialization always creates a database, Redis client, and email worker. Redis ping failure logs that startup continues, but worker startup still depends on Redis. A job-start error does not clean up the already-created DB/client, and a DB ping error does not close its pool.
- Safe modification: Define mandatory versus optional dependencies, unwind partial initialization, and split HTTP/worker startup where independent operation is required.
- Test coverage: No startup-failure or shutdown tests are present.

**Shutdown ordering and resource coverage are incomplete:**
- Files: `apps/backend/internal/server/server.go`, `apps/backend/internal/lib/job/job.go`.
- Why fragile: Shutdown closes the database before stopping workers; future database-backed tasks would lose their pool during draining. The application Redis client is not closed. Early HTTP shutdown errors skip remaining cleanup.
- Safe modification: Drain producers and workers before closing shared resources, close Redis, and attempt every cleanup step while collecting errors.
- Test coverage: No active-task shutdown or resource lifecycle tests are present.

**Runtime assets depend on current working directory:**
- Files: `apps/backend/internal/lib/email/client.go`, `apps/backend/internal/handler/openapi.go`, `apps/backend/internal/router/system.go`, `apps/backend/static/openapi.html`, `apps/backend/templates/emails/welcome.html`.
- Why fragile: Asset paths are relative (`static`, `templates/emails`), unlike embedded SQL migrations. A binary started from another directory cannot reliably serve docs or render emails.
- Safe modification: Embed assets or resolve an explicit asset directory; validate required files before accepting traffic.
- Test coverage: No alternate-working-directory checks are present.

**Generated API documentation can drift or fail silently:**
- Files: `packages/openapi/src/gen.ts`, `packages/openapi/src/contracts/health.ts`, `packages/openapi/openapi.json`, `apps/backend/static/openapi.json`, `apps/backend/internal/handler/health.go`.
- Why fragile: The generator writes two separate files via callbacks and only logs write failures. The health contract declares 200 only while the handler also returns 503. Generation is an explicit `gen` task, separate from compilation.
- Safe modification: Await writes, fail the command on errors, add response statuses, and compare served artifacts against generated contracts in verification.
- Test coverage: No generated-contract consistency tests are present.

**Package-global clients obscure ownership:**
- Files: `apps/backend/internal/lib/job/handlers.go`, `apps/backend/internal/service/auth.go`, `apps/backend/internal/logger/logger.go`.
- Why fragile: Worker initialization overwrites a package-global email client; auth initialization sets a global Clerk key; logger initialization changes zerolog globals. Multiple instances/tests in one process can affect each other.
- Safe modification: Store clients on owning service instances and isolate global initialization to a documented process boundary.
- Test coverage: No multi-instance isolation tests are present.

## Scaling Limits

**Rate limits are local to each API process:**
- Current capacity: Router configures an in-memory limiter at 20 requests per second per identifier; no measured service throughput is available.
- Files: `apps/backend/internal/router/router.go`.
- Limit: Replicas do not share enforcement state; system probes and product routes share the global policy.
- Scaling path: Define route/tenant policies, use distributed enforcement where required, and keep liveness probes independent of customer quotas.

**Workers scale together with the HTTP process:**
- Current capacity: Each `JobService` configures concurrency 10 and queue weights critical/default/low of 6/3/1; these are settings, not benchmark results.
- Files: `apps/backend/internal/lib/job/job.go`, `apps/backend/internal/server/server.go`, `spec.md`.
- Limit: Every HTTP replica starts email consumers; worker throughput and API replica count cannot be independently configured through the current entry point.
- Scaling path: Provide separate worker startup and configuration when traffic warrants it, preserving the shared job definitions.

## Dependencies at Risk

**Toolchain and lockfile consistency:**
- Risk: Root/package TypeScript and React declarations differ, and Bun plus npm lockfiles coexist. This is reproducibility debt, not a verified vulnerability or incompatibility.
- Files: `package.json`, `bun.lock`, `package-lock.json`, `packages/emails/package.json`, `packages/emails/bun.lock`, `packages/zod/package.json`, `packages/openapi/package.json`.
- Impact: Installation method can select different dependency graphs; documentation and CI need a single supported install path.
- Migration plan: Use the declared Bun package manager consistently, reconcile duplicate lockfile ownership, and align versions where packages exchange runtime/types.

**Dependency vulnerability status is unverified:**
- Risk: No vulnerability audit is part of this source mapping; no specific vulnerable/deprecated package claim is established.
- Files: `apps/backend/go.mod`, `apps/backend/go.sum`, `package.json`, `bun.lock`.
- Impact: Source inspection alone cannot certify dependency security.
- Migration plan: Add dependency scanning to normal CI using the supported lockfiles/toolchains and triage concrete results.

## Missing Critical Features

**Link attribution product and tenancy:**
- Problem: Links, redirects, workspace membership/permissions, custom domains, campaigns, conversions, analytics ingestion, partner programs, API keys, and webhook delivery are specified but have no implemented routes, repositories, or schema.
- Files: `spec.md`, `apps/backend/internal/router/router.go`, `apps/backend/internal/repository/repositories.go`, `apps/backend/internal/database/migrations/001_setup.sql`.
- Blocks: End-to-end execution of the core product workflows and validation of the redirect latency/analytics isolation requirements.

**Frontend application:**
- Problem: No application UI is present under `apps`; the React code in `packages/emails` renders email templates.
- Files: `package.json`, `packages/emails/src/templates/welcome.tsx`, `spec.md`.
- Blocks: Browser-based workspace, link-management, and analytics workflows.

**Separate readiness and liveness semantics:**
- Problem: Only `/status` is registered, combining dependency diagnostics with overall health, and it does not correctly aggregate Redis failure.
- Files: `apps/backend/internal/router/system.go`, `apps/backend/internal/handler/health.go`, `spec.md`.
- Blocks: Deployment policies that distinguish an alive process from one ready to serve dependency-backed requests.

## Test Coverage Gaps

**Backend has helpers but no executable test suites:**
- What's not tested: Registered routes, health aggregation, validation error shapes, request object isolation, authentication, SQL error translation, migrations, resource cleanup, and worker retries/cancellation. No `*_test.go` files are detected in application source.
- Files: `apps/backend/internal/testing/container.go`, `apps/backend/internal/testing/server.go`, `apps/backend/internal/testing/helpers.go`, `apps/backend/internal/handler/health.go`, `apps/backend/internal/validation/utils.go`, `apps/backend/internal/handler/base.go`.
- Risk: Infrastructure regressions have no automated behavior checks; helper presence must not be interpreted as coverage.
- Priority: High.

**Transaction helper naming promises different isolation:**
- What's not tested: `WithTransaction` says it rolls back afterward but commits successful callbacks; `WithRollbackTransaction` is the helper that always rolls back.
- Files: `apps/backend/internal/testing/transaction.go`.
- Risk: Tests choosing the former based on its comment can persist changes and contaminate shared fixtures.
- Priority: Medium; correct documentation and use the rollback helper for isolated fixture writes.

**TypeScript contracts and email export:**
- What's not tested: Runtime export resolution, served/generated OpenAPI equality, 503 schema coverage, valid absolute email links, and exported HTML interpolation. No application `.test.*`/`.spec.*` suites or test scripts are detected.
- Files: `packages/openapi/package.json`, `packages/openapi/src/gen.ts`, `packages/openapi/src/contracts/health.ts`, `packages/emails/src/templates/welcome.tsx`, `packages/zod/package.json`.
- Risk: Types can compile while runtime imports, documentation, or email output remain broken.
- Priority: Medium.

---

*Concerns audit: 2026-10-05*
