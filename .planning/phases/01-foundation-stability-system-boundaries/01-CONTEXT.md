# Phase 1: Foundation Stability & System Boundaries - Context

**Gathered:** 2026-10-05
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 1 turns the existing Go/TypeScript scaffold into a reproducible and independently operable foundation. It establishes API, redirector, worker, and migrator process boundaries; deterministic migrations and contracts; role-specific lifecycle and health behavior; executable quality gates; and vendor-neutral observability. It does not add workspaces, links, redirects, tracking, conversion, attribution, analytics, or product UI behavior assigned to later phases.

</domain>

<decisions>
## Implementation Decisions

### Process Ownership

- **D-01:** Create thin API, redirector, worker, and migrator composition roots over shared domain and infrastructure packages. Each role constructs, starts, drains, and closes only the resources it owns.
- **D-02:** API replicas must never start workers or run migrations implicitly. Migrations are an explicit release or operator action.
- **D-03:** Preserve Echo and sound existing packages. Extract narrow dependencies from the broad `server.Server` container only where resource ownership, process isolation, or testability requires it; do not perform a wholesale target-tree rewrite.
- **D-04:** Separate Asynq worker startup from HTTP startup now, while leaving product-specific queues and replayable event-stream design to their later phases.

### Lifecycle and Health

- **D-05:** Every long-running role exposes liveness that reflects the process itself and readiness that checks only dependencies required for that role.
- **D-06:** Public health responses expose sanitized component states. Detailed dependency errors belong in redacted structured logs and traces.
- **D-07:** Handle both SIGINT and SIGTERM. Shutdown stops intake, drains HTTP and workers within configured deadlines, closes shared resources after their dependents, attempts every cleanup action, and reports combined cleanup failures.
- **D-08:** Partial startup failure must unwind every resource already created. Optional integrations cannot accidentally become readiness requirements.

### Contract and Asset Authority

- **D-09:** Retain the existing TypeScript/Zod contract-authoring path, generate one OpenAPI artifact deterministically, and generate or verify Go transport boundaries from that artifact. Handwritten Go and TypeScript schemas must not become competing authorities.
- **D-10:** Contract generation must await writes, fail on any output error, cover every documented response status, and make CI reject stale generated or served artifacts.
- **D-11:** Required static and email assets are embedded in their owning binary/package where practical. Any non-embedded asset root is explicit, validated at startup, and independent of the current working directory.

### Reproducible Quality Gates

- **D-12:** Pin one supported Go, Node, Bun, PostgreSQL, Redis, lint, migration, and code-generation toolchain across local development and CI. Bun is the supported JavaScript workspace package manager; conflicting lockfile ownership is removed only after a clean reproducibility check.
- **D-13:** Root checks must actually invoke backend and every workspace package. A green orchestration command cannot hide missing package scripts or omit the Go service.
- **D-14:** Phase acceptance is behavior-based rather than driven by a blanket coverage percentage. Required regression coverage includes clean migration, independent role startup, readiness/liveness, SIGTERM drain order, partial-startup cleanup, request object isolation, safe validation errors, generated-contract equality, alternate working directory, and real PostgreSQL/Redis paths.
- **D-15:** Use real dependency integration tests through pinned containers where behavior depends on PostgreSQL or Redis; reserve mocks for narrow failure injection and unit boundaries.

### Observability and Configuration

- **D-16:** OpenTelemetry APIs and OTLP form the vendor-neutral instrumentation boundary for traces and metrics. Structured logs carry the same request/correlation context; exports route through a collector with redaction and cardinality controls.
- **D-17:** Direct New Relic-specific instrumentation may be retired or isolated behind the OpenTelemetry export boundary. Product code must not depend on a monitoring vendor SDK.
- **D-18:** Configuration packages return typed, wrapped errors. Composition roots decide whether startup exits; libraries do not call fatal process termination.
- **D-19:** Preserve existing external configuration names during stabilization or provide explicit compatibility and migration documentation before changing them. Secret values and raw provider errors never enter public diagnostics.

### the agent's Discretion

- Exact internal package names and the order of small extraction refactors, provided the four process roles and ownership rules above remain explicit.
- Exact supported patch versions at implementation time, after verifying current official support and compatibility with the repository.
- Choice of OpenAPI-to-Go generator and CI runner layout, provided there is one deterministic contract boundary and all required checks run.
- Whether individual assets are embedded or provided through an explicit asset root, based on update frequency and binary ownership.
- Exact test filenames, fixture structure, and metric names within the behavioral and cardinality constraints above.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Product Intent and Phase Contract

- `spec.md` §§36–38, 47, 51–58, 63–66, 70–73 — process architecture, configuration, observability, jobs, testing, CI/deployment, coding standards, documentation, and the first GSD objective.
- `.planning/PROJECT.md` — core value, constraints, brownfield context, and locked architecture decisions.
- `.planning/REQUIREMENTS.md` — Phase 1 requirements `PLAT-01` through `PLAT-08`, `SAFE-01`, and `SAFE-08`.
- `.planning/ROADMAP.md` §Phase 1 — phase goal, boundary, requirement allocation, and observable success criteria.

### Current Codebase Evidence

- `.planning/codebase/STACK.md` — implemented toolchain, dependency, configuration, build, and runtime baseline.
- `.planning/codebase/ARCHITECTURE.md` — current composition root, coupled resource container, request paths, contract pipeline, and extension points.
- `.planning/codebase/CONCERNS.md` — source-confirmed lifecycle, health, validation, asset, OpenAPI, global-state, and test gaps that Phase 1 must cover.

### Research Decisions

- `.planning/research/SUMMARY.md` §§Recommended Stack, Architecture Approach, Phase 1 — consolidated stack, boundary, sequencing, and Phase 1 recommendations.
- `.planning/research/STACK.md` — current supported versions, adoption timing, CI/toolchain, OpenTelemetry, and dependency recommendations.
- `.planning/research/ARCHITECTURE.md` — composition roots, ownership boundaries, data flow, and staged brownfield evolution.
- `.planning/research/PITFALLS.md` — phase-mapped failure modes and prevention strategies relevant to lifecycle, tenancy preparation, observability, and replay-safe foundations.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `apps/backend/cmd/flux/main.go`: existing composition root to split into role-specific thin entry points.
- `apps/backend/internal/server/server.go`: current resource construction and HTTP lifecycle logic; useful behavior to decompose around explicit ownership rather than discard.
- `apps/backend/internal/database/database.go` and `migrator.go`: pgx pool and embedded Tern migration foundation for explicit migration and integration checks.
- `apps/backend/internal/router/`, `middleware/`, and `handler/`: reusable Echo routing, error, request ID, logging, and system-route scaffolding after correctness fixes.
- `apps/backend/internal/lib/job/`: existing Asynq producer/consumer wiring to move under the worker composition root.
- `apps/backend/internal/testing/`: Testcontainers and assertion helpers that can seed the first executable suites after correcting misleading transaction-helper semantics.
- `packages/zod/` and `packages/openapi/`: existing schema and OpenAPI authoring pipeline to make deterministic and contract-enforced.
- `packages/emails/` and `apps/backend/templates/`: existing source/export assets to make working-directory independent.

### Established Patterns

- Manual constructor injection is already preferred; narrow dependency ownership should extend this pattern instead of introducing a service locator or DI framework.
- Echo middleware centralizes cross-cutting HTTP behavior, while business logic is expected to live in service/domain code.
- PostgreSQL access uses pgx and migrations use embedded Tern files; no ORM rewrite is wanted.
- TypeScript workspaces use Bun/Turborepo, strict TypeScript, Zod, and ts-rest/OpenAPI generation.
- Configuration is environment-driven and validated at startup, but current nested dot naming and fatal exits require compatibility-conscious repair.

### Integration Points

- Process startup and signal handling connect through `apps/backend/cmd/flux/main.go` and `apps/backend/internal/server/server.go`.
- Liveness/readiness replace the current combined `/status` behavior in `apps/backend/internal/router/system.go` and `handler/health.go`.
- Worker separation connects at `apps/backend/internal/lib/job/job.go`; the API may retain a producer client without owning consumers.
- Contract enforcement connects `packages/zod/src/`, `packages/openapi/src/gen.ts`, generated OpenAPI artifacts, and Go transport code/CI.
- Observability replaces or isolates direct New Relic hooks across logger, middleware, database, and Redis setup with an OpenTelemetry boundary.
- Root quality orchestration connects `package.json`, `turbo.json`, package scripts, `apps/backend/taskfile.yml`, Go commands, and the new CI workflow.

</code_context>

<specifics>
## Specific Ideas

- The desired foundation is deliberately “boring”: explicit process roles, deterministic startup and shutdown, reproducible builds, observable failure, and real regression coverage before product domains are added.
- Preserve useful boilerplate and make ownership seams visible. Folder symmetry with `spec.md` is less important than correct lifecycle boundaries.
- A redirector binary exists as an independently operable shell in Phase 1, while actual redirect behavior remains Phase 3 scope.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope. Workspace identity, links, redirect behavior, click-event transport, analytics storage, and product UI remain assigned to later roadmap phases.

</deferred>

---

*Phase: 1-Foundation Stability & System Boundaries*
*Context gathered: 2026-10-05*

