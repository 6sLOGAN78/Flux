<!-- refreshed: 2026-10-05 -->
# Architecture

**Analysis Date:** 2026-10-05

## System Overview

```text
HTTP clients
    |
    v
Echo router + global middleware
apps/backend/internal/router/router.go
apps/backend/internal/middleware/
    |
    +--> /status --> handler/health.go --> PostgreSQL + Redis probes
    +--> /docs   --> handler/openapi.go --> static/openapi.html
    +--> /static --> apps/backend/static/

Process composition: apps/backend/cmd/flux/main.go
    |
    v
Server resource container: apps/backend/internal/server/server.go
    +--> database/database.go --> pgxpool --> PostgreSQL
    +--> Redis client --> Redis
    +--> lib/job/job.go --> Asynq workers --> lib/email/client.go --> Resend
    +--> logger/logger.go --> zerolog + optional New Relic

Build-time assets (independent of Go compilation)
packages/zod/src/ --> packages/openapi/src/ --> backend/static/openapi.json
packages/emails/src/templates/ --> backend/templates/emails/welcome.html
```

Only the API foundation is implemented. `apps/backend/internal/router/system.go` registers health, documentation, and static-file routes. `apps/backend/internal/router/router.go` creates an empty `/api/v1` group. Link management, redirect serving, analytics, frontend applications, and separate worker binaries have no implementation under `apps/`; `spec.md` describes product requirements rather than executable architecture.

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| Composition root | Load config, initialize resources and layers, start HTTP, handle interrupt | `apps/backend/cmd/flux/main.go` |
| Server | Own database, Redis, jobs, logger, configuration, HTTP lifecycle | `apps/backend/internal/server/server.go` |
| Router | Install cross-cutting middleware and register routes | `apps/backend/internal/router/router.go`, `apps/backend/internal/router/system.go` |
| Health handler | Probe PostgreSQL and Redis; return dependency status | `apps/backend/internal/handler/health.go` |
| Handler adapters | Bind/validate typed requests, log and trace execution, serialize responses | `apps/backend/internal/handler/base.go` |
| Service registry | Initialize Clerk and expose job service | `apps/backend/internal/service/services.go`, `apps/backend/internal/service/auth.go` |
| Repository registry | Empty injection scaffold; no application queries | `apps/backend/internal/repository/repositories.go` |
| Database | Configure pgx pool and query instrumentation; execute embedded tern migrations | `apps/backend/internal/database/database.go`, `apps/backend/internal/database/migrator.go` |
| Background jobs | Create and process welcome-email tasks in Redis-backed queues | `apps/backend/internal/lib/job/job.go`, `apps/backend/internal/lib/job/email_tasks.go` |
| Email delivery | Render Go HTML templates and send through Resend | `apps/backend/internal/lib/email/client.go` |
| API contracts | Generate OpenAPI from ts-rest routes and shared Zod response schemas | `packages/openapi/src/index.ts`, `packages/zod/src/health.ts` |

## Pattern Overview

**Overall:** Layered Go service scaffold with manual dependency injection and TypeScript asset/contract workspaces.

**Key Characteristics:**
- Wire dependencies explicitly in `apps/backend/cmd/flux/main.go`; registry constructors collect handlers, services, and repositories.
- `apps/backend/internal/server/server.go` is a shared concrete resource container passed into most packages, rather than narrow dependency interfaces.
- Middleware centralizes request identity, tracing, logging, CORS, recovery, throttling, and error translation in `apps/backend/internal/middleware/`.
- The active health route calls infrastructure directly; the repository layer in `apps/backend/internal/repository/repositories.go` contains no database operations.
- TypeScript contracts in `packages/openapi/src/contracts/` document the Go API; they do not generate its handlers or enforce runtime request validation.

## Layers

**Process and infrastructure:**
- Purpose: Configure and own long-lived clients and server lifecycle.
- Location: `apps/backend/cmd/flux/`, `apps/backend/internal/server/`, `apps/backend/internal/config/`, `apps/backend/internal/database/`, `apps/backend/internal/logger/`.
- Contains: Startup, environment parsing, pgx pooling, embedded migration runner, logging/APM setup.
- Depends on: PostgreSQL, Redis, New Relic SDK, Go HTTP server.
- Used by: Handler, service, middleware, and repository constructors.

**HTTP transport:**
- Purpose: Turn requests into HTTP responses and apply shared behavior.
- Location: `apps/backend/internal/router/`, `apps/backend/internal/middleware/`, `apps/backend/internal/handler/`, `apps/backend/internal/validation/`.
- Contains: Route registration, global error handler, auth adapter, typed handler helpers, health/docs handlers.
- Depends on: Echo, server resources, `apps/backend/internal/errs/`, `apps/backend/internal/sqlerr/`.
- Used by: `apps/backend/cmd/flux/main.go` through `router.NewRouter`.

**Application/data scaffold:**
- Purpose: Provide extension points for application behavior and persistence.
- Location: `apps/backend/internal/service/`, `apps/backend/internal/repository/`, `apps/backend/internal/model/`.
- Contains: Clerk setup, job reference, empty repository registry, shared UUID/timestamp/pagination model types.
- Depends on: Server resources and job infrastructure.
- Used by: Composition and handler constructors; registry parameters are currently unused where no feature needs them.

**Background delivery:**
- Purpose: Execute queued welcome-email deliveries.
- Location: `apps/backend/internal/lib/job/`, `apps/backend/internal/lib/email/`.
- Contains: Asynq producer/task definition, consumer, Resend adapter, HTML template renderer.
- Depends on: Redis, backend template files, Resend configuration.
- Used by: Server startup and `service.Services.Job`; no routed feature enqueues a task.

**Shared TypeScript packages:**
- Purpose: Maintain schema/contract and email source assets.
- Location: `packages/zod/`, `packages/openapi/`, `packages/emails/`.
- Contains: Zod response schemas, ts-rest contracts, OpenAPI generation, React Email templates.
- Depends on: Workspace imports and package-specific build commands.
- Used by: OpenAPI generation and email export; no frontend app consumes them in this tree.

## Data Flow

### Primary Request Path

1. Main wires the Echo router to the HTTP server (`apps/backend/cmd/flux/main.go:23`, `apps/backend/internal/server/server.go:78`).
2. Echo applies throttling, CORS/security, request ID, tracing, context logger, request logging, and recovery (`apps/backend/internal/router/router.go:23`).
3. `GET /status` dispatches directly to `HealthHandler.CheckHealth` (`apps/backend/internal/router/system.go:10`, `apps/backend/internal/handler/health.go:25`).
4. The handler probes PostgreSQL and Redis sequentially with independent five-second contexts, builds a status map, and returns JSON (`apps/backend/internal/handler/health.go`).
5. Database failure yields HTTP 503; Redis failure appears in the checks but does not change overall health (`apps/backend/internal/handler/health.go`).

### Background Email Flow

1. `NewWelcomeEmailTask` serializes recipient/name, selects the default queue, three retries, and a 30-second task timeout (`apps/backend/internal/lib/job/email_tasks.go:18`).
2. A caller can enqueue through `JobService.Client`; no current HTTP handler invokes this producer (`apps/backend/internal/lib/job/job.go`).
3. The Asynq server dispatches `email:welcome` to the consumer (`apps/backend/internal/lib/job/job.go:41`, `apps/backend/internal/lib/job/handlers.go:20`).
4. The consumer decodes payload and calls `SendWelcomeEmail`; the email client renders `templates/emails/welcome.html` and sends through Resend (`apps/backend/internal/lib/email/emails.go`, `apps/backend/internal/lib/email/client.go:26`).

### Contract and Asset Flow

1. Define runtime schemas in `packages/zod/src/health.ts` and export through `packages/zod/src/index.ts`.
2. Attach response schemas to ts-rest contracts in `packages/openapi/src/contracts/health.ts`; compose routes in `packages/openapi/src/contracts/index.ts`.
3. `packages/openapi/src/index.ts` produces the OpenAPI object; `packages/openapi/src/gen.ts:21` writes both `packages/openapi/openapi.json` and `apps/backend/static/openapi.json`.
4. React Email export configured in `packages/emails/package.json` renders TSX templates into `apps/backend/templates/emails/` for Go rendering.

**State Management:**
- PostgreSQL connection state is held by `database.Database.Pool`; the sole migration `apps/backend/internal/database/migrations/001_setup.sql` contains no schema statements.
- Redis stores Asynq queues; the HTTP rate limiter is process-local memory (`apps/backend/internal/router/router.go`, `apps/backend/internal/lib/job/job.go`).
- Request identity, user claims, and contextual logger use Echo context; logging/APM also uses request `context.Context` (`apps/backend/internal/middleware/context.go`, `apps/backend/internal/middleware/auth.go`).

## Key Abstractions

**Server resource container:**
- Purpose: Share configured infrastructure and own lifecycle.
- Examples: `apps/backend/internal/server/server.go`, `apps/backend/internal/handler/base.go`.
- Pattern: Explicit constructor injection of `*server.Server`.

**Typed handler/response adapters:**
- Purpose: Wrap a request implementing `Validate() error` with bind, validation, metrics, and response output.
- Examples: `apps/backend/internal/handler/base.go`, `apps/backend/internal/validation/utils.go`.
- Pattern: Generic `Handle`, `HandleNoContent`, and `HandleFile` functions backed by `ResponseHandler`; the active health/docs handlers bypass these wrappers.

**Application HTTP error:**
- Purpose: Standard JSON error envelope with status/code/message, field errors, and optional action.
- Examples: `apps/backend/internal/errs/http.go`, `apps/backend/internal/errs/types.go`, `apps/backend/internal/sqlerr/handler.go`.
- Pattern: Construct typed errors and return them; central middleware maps wrapped database/Echo errors using `errors.As`.

**Contract router:**
- Purpose: Compose HTTP documentation from typed schemas.
- Examples: `packages/openapi/src/contracts/index.ts`, `packages/openapi/src/contracts/health.ts`.
- Pattern: `initContract().router(...)`; OpenAPI security metadata exists in `packages/openapi/src/utils.ts` but the health route is public.

## Entry Points

**Backend executable:**
- Location: `apps/backend/cmd/flux/main.go`.
- Triggers: `go run ./cmd/flux` or `task run` from `apps/backend/`.
- Responsibilities: Configuration, optional non-local migrations, resource/layer wiring, HTTP startup, interrupt-driven shutdown.

**OpenAPI generator:**
- Location: `packages/openapi/src/gen.ts`.
- Triggers: Package `gen` script in `packages/openapi/package.json`.
- Responsibilities: Render and write API documentation assets; separate from the package `build` script.

**Email preview/export:**
- Location: `packages/emails/package.json`, `packages/emails/src/templates/welcome.tsx`.
- Triggers: Package `dev` and `export` scripts.
- Responsibilities: Preview on port 3001 and export backend HTML templates.

## Architectural Constraints

- **Threading:** Go serves concurrent requests; main starts HTTP in a goroutine. Asynq runs ten concurrent workers inside the same process (`apps/backend/cmd/flux/main.go`, `apps/backend/internal/lib/job/job.go`).
- **Global state:** Clerk uses SDK-wide `clerk.SetKey`; jobs use a package-global `emailClient`; email preview data is a package map; logger initialization sets zerolog globals (`apps/backend/internal/service/auth.go`, `apps/backend/internal/lib/job/handlers.go`, `apps/backend/internal/lib/email/preview.go`, `apps/backend/internal/logger/logger.go`).
- **Circular imports:** No import cycle detected among inspected source packages. Preserve the direction from router/handler/service toward server/infrastructure (`apps/backend/internal/router/router.go`, `apps/backend/internal/server/server.go`).
- **Working directory:** Docs, static files, and emails use paths relative to the backend directory. Launch with `apps/backend/` as working directory or supply those assets there (`apps/backend/internal/router/system.go`, `apps/backend/internal/handler/openapi.go`, `apps/backend/internal/lib/email/client.go`).
- **Lifecycle:** PostgreSQL must connect during construction; Redis ping failure logs and continues, while an Asynq start error is fatal. Shutdown closes HTTP, database, and jobs but does not close the separate Redis client (`apps/backend/internal/server/server.go`).
- **Schema source:** Go payloads and TypeScript schemas are maintained separately; generated OpenAPI describes only health HTTP 200 despite a possible 503 response (`apps/backend/internal/handler/health.go`, `packages/openapi/src/contracts/health.ts`).
- **Project skills:** No project skill directories detected at `.codex/skills/` or `.agents/skills/`; no additional local architecture constraints are available.

## Anti-Patterns

### Package-global job dependency

**What happens:** `InitHandlers` overwrites a shared `emailClient` used by all job service instances (`apps/backend/internal/lib/job/handlers.go`).
**Why it's wrong:** Independent server/test instances cannot isolate delivery dependencies; shared clients make concurrent initialization fragile.
**Do this instead:** Store the email dependency on each job service and inject it through its constructor, extending the instance-based pattern in `apps/backend/internal/lib/job/job.go`.

### Shared mutable request captured by handler adapters

**What happens:** The adapters capture `req` outside the returned request closure, then bind into it for every request (`apps/backend/internal/handler/base.go`).
**Why it's wrong:** Pointer payloads can be reused across concurrent HTTP requests or retain omitted fields.
**Do this instead:** Allocate a fresh payload inside each returned handler, using a factory or equivalent request-local construction before `BindAndValidate` in `apps/backend/internal/validation/utils.go`.

### Dependency status does not match overall health

**What happens:** Redis failure records an unhealthy check while overall status stays healthy (`apps/backend/internal/handler/health.go`).
**Why it's wrong:** Redis-backed background jobs can be unavailable while health reports HTTP 200.
**Do this instead:** Define readiness requirements explicitly and align overall status with required dependencies in `apps/backend/internal/handler/health.go`; do not assume its present response is job readiness.

## Error Handling

**Strategy:** Return errors through Echo and translate centrally in `apps/backend/internal/middleware/global.go`; startup uses panic/fatal logging in `apps/backend/cmd/flux/main.go` and `apps/backend/internal/config/config.go`.

**Patterns:**
- Use `errs.HTTPError` constructors for client-facing application failures (`apps/backend/internal/errs/types.go`).
- Wrap infrastructure failures with context using `%w` or pkg/errors (`apps/backend/internal/server/server.go`, `apps/backend/internal/lib/email/client.go`).
- Convert PostgreSQL constraint/no-row errors via `sqlerr.HandleError` and normalize unknown errors to a generic HTTP 500 (`apps/backend/internal/sqlerr/handler.go`, `apps/backend/internal/middleware/global.go`).
- Echo recovery handles panics; validation currently assumes specific binding-error text and supported validation-error types (`apps/backend/internal/middleware/global.go`, `apps/backend/internal/validation/utils.go`).
- Queue consumers return delivery failures for Asynq retry handling (`apps/backend/internal/lib/job/handlers.go`).

## Cross-Cutting Concerns

**Logging:** Zerolog request-scoped logging includes request ID, method/path/IP and optional trace metadata; New Relic instruments Echo, pgx, and Redis when configured (`apps/backend/internal/logger/logger.go`, `apps/backend/internal/middleware/context.go`, `apps/backend/internal/middleware/tracing.go`, `apps/backend/internal/database/database.go`, `apps/backend/internal/server/server.go`).

**Validation:** Koanf/environment config is checked using go-playground validator; typed HTTP payloads must implement `Validate() error`; Zod schemas describe shared TypeScript contracts independently (`apps/backend/internal/config/config.go`, `apps/backend/internal/validation/utils.go`, `packages/zod/src/health.ts`).

**Authentication:** Clerk service sets the SDK key; `RequireAuth` validates header authorization and sets subject/organization role/permissions. The current router does not attach that middleware to any route (`apps/backend/internal/service/auth.go`, `apps/backend/internal/middleware/auth.go`, `apps/backend/internal/router/router.go`).

---

*Architecture analysis: 2026-10-05*
