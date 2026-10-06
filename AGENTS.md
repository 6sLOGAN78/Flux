<!-- GSD:project-start source:PROJECT.md -->
## Project

**Flux**

Flux is the temporary internal codename for a production-grade, multi-tenant link attribution and marketing analytics SaaS. It gives marketers, growth teams, developers, agencies, and affiliate teams branded short links, reliable redirects, asynchronous click tracking, conversion attribution, analytics, developer APIs, and later partner and billing capabilities through an original Go-first implementation.

The repository is brownfield: it contains a useful Go/Echo service foundation and TypeScript packages for API contracts and email templates, but the product domains, database schema, frontend, redirect data plane, and executable test suites are not implemented yet.

**Core Value:** A marketer can create a link, send a visitor through a fast and reliable redirect, record the click asynchronously, attribute a later conversion and revenue to that click, and see the result in analytics.

### Constraints

- **Architecture**: Go-first modular architecture with independently deployable binaries where useful — scale API, redirects, and workers separately without premature microservices.
- **Existing code**: Preserve sound boilerplate conventions and adapt incrementally — do not rewrite functioning code solely to match the target folder tree.
- **Redirect performance**: Cached redirect lookup targets p95 under 30 ms server processing; uncached targets p95 under 100 ms — redirect availability cannot depend synchronously on analytics.
- **Reliability**: ClickHouse, analytics workers, webhooks, and email may fail without preventing redirects — the redirect response is the critical path.
- **Tenancy**: Every workspace-owned entity contains or resolves to a workspace identifier, with server-side authorization and isolation tests from the first product schema.
- **Data**: PostgreSQL is authoritative for transactional entities; Redis is never authoritative for link configuration; analytics may be eventually consistent.
- **Security**: Strict destination URL validation, SSRF defenses, secure sessions, CSRF protection where applicable, API-key hashing, webhook signing, rate limits, and parameterized SQL are mandatory.
- **Privacy**: Minimize collected data, avoid unnecessary raw network identifiers, support configurable retention later, and do not use invasive fingerprinting as primary identity.
- **Money**: Store monetary values in integer minor units with explicit currency — never use floating-point commission or revenue arithmetic.
- **Developer experience**: Versioned REST API, OpenAPI, idempotency, pagination, predictable errors, request IDs, rate-limit metadata, and SDK-friendly contracts are product requirements.
- **Quality**: Major domain logic requires automated tests; external input must be validated; no placeholder production paths or silently swallowed errors.
- **Brand**: Flux is temporary; no Dub-derived branding or protected assets may be used.
<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->
## Technology Stack

## Languages
- Go 1.25.5 module language directive — HTTP backend, middleware, database access, jobs, email delivery, and testing helpers in `apps/backend/go.mod` and `apps/backend/internal/`.
- TypeScript — reusable schemas and REST contracts in `packages/zod/src/` and `packages/openapi/src/`; email template TSX in `packages/emails/src/templates/welcome.tsx`.
- SQL — Tern migration files in `apps/backend/internal/database/migrations/`; `001_setup.sql` contains only migration scaffold comments, without application tables.
- HTML — exported email template in `apps/backend/templates/emails/welcome.html` and API reference page in `apps/backend/static/openapi.html`.
- YAML/JSON — task runner, linting, compiler, and workspace configuration in `apps/backend/taskfile.yml`, `apps/backend/.golangci.yml`, `turbo.json`, and package manifests.
## Runtime
- Go backend executable: `apps/backend/cmd/flux/main.go`; `apps/backend/go.mod` declares Go 1.25.5.
- Node.js >=22 declared in root `package.json`; Bun scripts and package management support the TypeScript workspaces.
- No frontend application runtime is implemented: `apps/frontend/` is empty. React dependencies in `package.json` are not evidence of a running dashboard.
- Bun 1.2.13 declared by `package.json` `packageManager`; workspace globs cover `apps/*` and `packages/*`.
- Lockfiles present: `bun.lock`, `packages/emails/bun.lock`, and `package-lock.json`. The repository contains both Bun and npm lockfiles; use the declared Bun manager for workspace operations.
- Go modules: `apps/backend/go.mod` and `apps/backend/go.sum`.
## Frameworks
- Echo v4.15.2 — HTTP routing and middleware; `apps/backend/internal/router/router.go`, declared in `apps/backend/go.mod`.
- React 19.1.0 in email workspace; React/React DOM ^19.2.6 in root manifest — implemented UI usage is the email component in `packages/emails/src/templates/welcome.tsx`.
- React Email ^6.3.3 and `@react-email/components` 0.0.34 — preview and HTML export in `packages/emails/package.json`.
- Zod ^3.24.2 — schema validation/types in `packages/zod/src/health.ts`; `@anatine/zod-openapi` ^2.2.7 attaches documentation metadata.
- `@ts-rest/core` and `@ts-rest/open-api` ^3.52.1 — REST contracts and OpenAPI generation in `packages/openapi/src/contracts/health.ts` and `packages/openapi/src/index.ts`.
- Go standard `testing` — backend helpers under `apps/backend/internal/testing/`.
- Testify v1.11.1 — assertions used in `apps/backend/internal/testing/container.go` and `apps/backend/internal/testing/assertions.go`.
- Testcontainers Go v0.42.0 — PostgreSQL integration helper starts `postgres:15-alpine` in `apps/backend/internal/testing/container.go`.
- No runnable `_test.go` suites or JavaScript test runner configuration detected under `apps/backend/` or `packages/`; distinguish helper infrastructure from actual test coverage.
- Turborepo ^2.9.6 — root orchestration in `package.json` and `turbo.json`; build dependencies use `^build`, dev is persistent and uncached.
- TypeScript ^6.0.3 at root; ^5.8.2 in `packages/zod/package.json` and `packages/openapi/package.json`. These are manifest ranges, not a claim of installed resolution.
- `tsc-alias` ^1.8.12 — rewrites emitted aliases during schema/contract compilation; `packages/zod/package.json`, `packages/openapi/package.json`.
- `tsx` ^4.19.3 — contract generator execution in `packages/openapi/package.json`.
- `concurrently` ^9.1.2 and `wait-on` ^8.0.3 — compiler/watch coordination; OpenAPI dev waits for `packages/zod/dist/index.js`.
- Taskfile v3 schema — backend tasks `run`, `migrations:new`, `migrations:up`, and `tidy` in `apps/backend/taskfile.yml`; Task executable version is not pinned.
- GolangCI-Lint configuration is present at `apps/backend/.golangci.yml`; linter executable version is not pinned by `apps/backend/go.mod`.
## Key Dependencies
- `github.com/jackc/pgx/v5` v5.9.2 — PostgreSQL pool and driver in `apps/backend/internal/database/database.go`; no ORM is used.
- `github.com/jackc/tern/v2` v2.4.1 — embedded migrations and `schema_version` tracking in `apps/backend/internal/database/migrator.go`.
- `github.com/redis/go-redis/v9` v9.14.1 — Redis client and health checks in `apps/backend/internal/server/server.go` and `apps/backend/internal/handler/health.go`.
- `github.com/hibiken/asynq` v0.26.0 — Redis-backed worker and welcome-email jobs in `apps/backend/internal/lib/job/job.go` and `email_tasks.go`.
- `github.com/clerk/clerk-sdk-go/v2` v2.6.0 — authentication setup and reusable header authorization middleware in `apps/backend/internal/service/auth.go` and `apps/backend/internal/middleware/auth.go`.
- `github.com/resend/resend-go/v2` v2.28.0 — transactional email transport in `apps/backend/internal/lib/email/client.go`.
- Koanf v2.3.4 plus env provider v1.1.0 — hierarchical environment configuration in `apps/backend/internal/config/config.go`.
- Godotenv v1.5.1 — automatic dotenv loading imported by `apps/backend/internal/config/config.go`.
- Go Playground validator v10.30.2 — configuration and request validation in `apps/backend/internal/config/config.go` and `apps/backend/internal/validation/utils.go`.
- Zerolog v1.35.1 — structured logs in `apps/backend/internal/logger/logger.go`; pgx-zerolog adapter is declared in `apps/backend/go.mod`.
- New Relic Go agent v3.43.3, Echo integration v1.1.5, pgx integration v1.3.4, Redis integration v1.1.2, pkgerrors integration v1.1.0 — optional instrumentation in `apps/backend/internal/logger/logger.go`, `apps/backend/internal/middleware/tracing.go`, `apps/backend/internal/database/database.go`, and `apps/backend/internal/server/server.go`.
## Configuration
- `apps/backend/internal/config/config.go` loads environment names prefixed `FLUX_`, removes that prefix, lowercases the remainder, and uses `.` as the nesting delimiter. Nested names therefore use dots, for example `FLUX_DATABASE.HOST`; underscores inside field names remain underscores.
- Required sections: primary environment, server timeouts/CORS, PostgreSQL connection and pool settings, Clerk secret key, Redis address, Resend API key; see structs and validation tags in `apps/backend/internal/config/config.go`.
- Observability defaults and custom validation live in `apps/backend/internal/config/observability.go`; blank New Relic license disables the agent in `apps/backend/internal/logger/logger.go`.
- `apps/backend/.env` and `apps/backend/.env.sample` are present; contents are not inspected. `.gitignore` excludes `.env`.
- No project skills are present under `.agents/` or `.codex/`.
- `package.json` and `turbo.json`: workspace build/dev/cleanup orchestration. Root scripts also name lint/typecheck/format tasks; inspect each package for actual script support rather than assuming every root task executes checks.
- `packages/zod/tsconfig.json`, `packages/openapi/tsconfig.json`, `packages/emails/tsconfig.json`: strict NodeNext ESM, ES2022, declarations/source maps, `dist/` output; email config uses `react-jsx`.
- `packages/openapi/src/gen.ts`: writes generated JSON to `packages/openapi/openapi.json` and `apps/backend/static/openapi.json`.
- `packages/emails/package.json`: `export` writes HTML to `apps/backend/templates/emails/`; this is a separate task from the root build.
- `apps/backend/taskfile.yml`: `task run` executes `go run ./cmd/flux`; migrations CLI consumes `FLUX_DB_DSN` separately from the application's structured configuration.
## Platform Requirements
- Go compatible with `apps/backend/go.mod`, Bun 1.2.13, and Node >=22 per `package.json`; Task and Tern CLIs are needed only for the commands in `apps/backend/taskfile.yml` that invoke them.
- Accessible PostgreSQL and Redis for `apps/backend/internal/server/server.go`; Docker-compatible container runtime for `apps/backend/internal/testing/container.go`.
- Run the backend with `apps/backend/` as working directory: static documents and email templates use relative paths in `apps/backend/internal/router/system.go` and `apps/backend/internal/lib/email/client.go`.
- Go HTTP executable from `apps/backend/cmd/flux/main.go` plus runtime files in `apps/backend/static/` and `apps/backend/templates/emails/`; migration SQL is embedded by `apps/backend/internal/database/migrator.go`.
- PostgreSQL, Redis, Clerk, and Resend configuration are wired in `apps/backend/internal/config/config.go`; New Relic is optional.
- Hosting provider, Dockerfile, Compose stack, Kubernetes manifests, and CI deployment workflow: not detected. `spec.md` contains proposed platform architecture; it is not an implementation manifest.
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

## Naming Patterns
- Use lowercase Go files named for their responsibility, with underscores for multiple words: `apps/backend/internal/middleware/request_id.go`, `apps/backend/internal/lib/job/email_tasks.go`.
- Use lowercase TypeScript feature files: `packages/zod/src/health.ts`, `packages/openapi/src/contracts/health.ts`. Email templates use lowercase `.tsx` files: `packages/emails/src/templates/welcome.tsx`.
- Use numbered SQL migration files under `apps/backend/internal/database/migrations/`, following `001_setup.sql` and its tern up/down separator.
- Use PascalCase for exported Go functions and methods, camelCase for private helpers: `NewHealthHandler` and `CheckHealth` in `apps/backend/internal/handler/health.go`; `generateErrorCode` in `apps/backend/internal/sqlerr/handler.go`.
- Use `New<Type>` constructors with dependencies supplied explicitly: `NewServices` in `apps/backend/internal/service/services.go`, `NewGlobalMiddlewares` in `apps/backend/internal/middleware/global.go`.
- Use camelCase TypeScript helpers and contract operations: `schemaWithPagination` in `packages/zod/src/utils.ts`, `getSecurityMetadata` in `packages/openapi/src/utils.ts`, `getHealth` in `packages/openapi/src/contracts/health.ts`.
- Use PascalCase React components: `WelcomeEmail` in `packages/emails/src/templates/welcome.tsx`.
- Use short Go receiver/dependency names in small scopes (`s`, `h`, `cfg`, `ctx`, `err`) and descriptive names for longer operations: `apps/backend/internal/server/server.go`, `apps/backend/internal/handler/health.go`.
- Use camelCase TypeScript locals and props: `openApiSecurity` in `packages/openapi/src/utils.ts`, `userFirstName` in `packages/emails/src/templates/welcome.tsx`.
- Preserve existing wire field spellings rather than deriving them from identifiers. `apps/backend/internal/model/base.go` uses camelCase JSON timestamps and snake_case DB fields; `apps/backend/internal/lib/job/email_tasks.go` uses snake_case payload JSON; health checks use `response_time` in `packages/zod/src/health.ts`.
- Use PascalCase structs and interfaces: `HTTPError`, `FieldError`, `Action` in `apps/backend/internal/errs/http.go`; `WelcomeEmailProps` in `packages/emails/src/templates/welcome.tsx`.
- Prefix exported Zod schemas with `Z`: `ZHealthResponse` in `packages/zod/src/health.ts`. Keep nested implementation schemas private where they are not reused.
- Use typed string constants for enums: `ActionTypeRedirect` in `apps/backend/internal/errs/http.go`, `TaskWelcome` in `apps/backend/internal/lib/job/email_tasks.go`.
- Use generic response/helper types where applicable: `PaginatedResponse` in both `apps/backend/internal/model/base.go` and `packages/zod/src/utils.ts`, `Ptr[T any]` in `apps/backend/internal/testing/helpers.go`.
## Code Style
- Go formatting is exposed by the `tidy` task in `apps/backend/taskfile.yml` using `go fmt ./...`. Use Go formatting for changed Go files; sample helpers contain spacing/import-order inconsistencies, so do not copy those inconsistencies.
- TypeScript source uses two spaces, double quotes, semicolons, trailing commas in multiline objects, and arrow functions: `packages/zod/src/health.ts`, `packages/openapi/src/utils.ts`, `packages/emails/src/templates/welcome.tsx`.
- No Prettier, Biome, or ESLint configuration is detected in the source inventory. `package.json` has `format:check`, `format:fix`, `lint`, and `lint:fix` Turbo scripts, but the package manifests do not implement those tasks; these are not evidence of enforced TypeScript formatting.
- `turbo.json` defines a `format` task while the root script requests `format:check`. Check the task configuration before relying on root formatting commands.
- Backend rules are declared in `apps/backend/.golangci.yml` with configuration version `2` and a three-minute timeout; no installed golangci-lint version or passing lint result is established by this map.
- The configured checks include error handling (`errcheck`, `errorlint`, `nilerr`), security (`gosec`, `bidichk`), complexity (`cyclop`, `gocognit`, `funlen`), naming (`revive`, `stylecheck`), imports (`goimports`), context/resource use (`noctx`, `bodyclose`, `sqlclosecheck`), and tests (`testifylint`, `testpackage`, `tparallel`).
- Configured function limits are 100 lines and 50 statements, excluding comments; cyclomatic maximum is 30 with package average 10; cognitive threshold is 20: `apps/backend/.golangci.yml`.
- `nolint` directives must name specific checks and normally explain the suppression; `funlen`, `gocognit`, and `lll` are explanation exceptions in `apps/backend/.golangci.yml`.
- Tests matching `_test.go` have selected exclusions, including `errcheck`, `funlen`, `gosec`, and `noctx`: `apps/backend/.golangci.yml`. Helpers under `apps/backend/internal/testing/` are ordinary `.go` files and do not match that exclusion.
- TypeScript strictness is configured in `packages/zod/tsconfig.json`, `packages/openapi/tsconfig.json`, and `packages/emails/tsconfig.json`: `strict`, `noUncheckedIndexedAccess`, `noImplicitOverride`, `isolatedModules`, and `verbatimModuleSyntax`.
## Import Organization
- Backend imports use the module prefix `github.com/6sLOGAN78/flux/internal/...`, declared in `apps/backend/go.mod`.
- TypeScript `@/*` resolves into each package's `src` directory via `packages/zod/tsconfig.json`, `packages/openapi/tsconfig.json`, and `packages/emails/tsconfig.json`; `packages/openapi/src/contracts/health.ts` uses `@/utils.js`.
- Preserve `.js` suffixes in relative TypeScript imports for NodeNext ESM output: `packages/zod/src/index.ts`, `packages/openapi/src/contracts/index.ts`.
- Cross-package imports use workspace package names such as `@flux/zod` in `packages/openapi/src/contracts/health.ts` and `workspace:*` in `packages/openapi/package.json`.
## Error Handling
- Return Go errors and wrap failures with operation context and `%w`: `apps/backend/internal/server/server.go`, `apps/backend/internal/lib/job/handlers.go`, `apps/backend/internal/testing/transaction.go`.
- Construct API errors through `apps/backend/internal/errs/types.go`; the JSON shape in `apps/backend/internal/errs/http.go` includes code, message, status, override, field errors, and optional action.
- Let the Echo error boundary serialize handler failures. `GlobalErrorHandler` in `apps/backend/internal/middleware/global.go` recognizes typed HTTP/Echo errors, maps SQL errors through `apps/backend/internal/sqlerr/handler.go`, logs the original failure, and avoids writing an already committed response.
- Match wrapped errors with `errors.As`/`errors.Is`, as in `apps/backend/internal/sqlerr/handler.go`.
- Generic endpoint wrappers in `apps/backend/internal/handler/base.go` centralize validation, timing, logging, and JSON/file/no-content responses. Health and documentation handlers also use direct Echo methods; do not assume every route passes through the wrapper (`apps/backend/internal/router/system.go`).
- Request payloads handled by wrappers implement `validation.Validatable` and `Validate() error`: `apps/backend/internal/validation/utils.go`. Its bind-error string parsing and custom-error type assertions assume particular shapes; handle new error types deliberately instead of assuming arbitrary errors are supported.
- The TypeScript generator logs asynchronous write failures with `console.error` in `packages/openapi/src/gen.ts`; it does not explicitly set an exit code on failure.
## Logging
- Use contextual loggers from `middleware.GetLogger(c)` and attach structured operation fields: `apps/backend/internal/handler/health.go`, `apps/backend/internal/handler/base.go`.
- Use fluent `Str`, `Int`, `Dur`, and `Err` fields followed by a short message. Record validation/handler durations through the central wrapper: `apps/backend/internal/handler/base.go`.
- Request logging records identifiers, route/request metadata, latency, status, and severity selected from response status: `apps/backend/internal/middleware/global.go`.
- Guard optional tracing/logger services before recording custom events: `apps/backend/internal/handler/health.go`.
- `PrintJSON` in `apps/backend/internal/lib/utils/utils.go` uses `fmt.Println` as a utility; operational backend code generally uses the structured logger.
## Comments
- Explain exported Go helper responsibilities and non-obvious implementation assumptions: `apps/backend/internal/testing/helpers.go`, `apps/backend/internal/sqlerr/err.go`.
- Add rationale for middleware quirks and fallback behavior, such as status extraction and Redis startup behavior: `apps/backend/internal/middleware/global.go`, `apps/backend/internal/server/server.go`.
- Keep comments aligned with behavior. `WithTransaction` in `apps/backend/internal/testing/transaction.go` commits successful callbacks despite its rollback-oriented summary; use implementation behavior as the authority.
- Not detected in the TypeScript implementation samples (`packages/zod/src/utils.ts`, `packages/openapi/src/utils.ts`, `packages/emails/src/templates/welcome.tsx`). Go uses ordinary doc comments.
- Inline configuration explanations are present in `packages/openapi/tsconfig.json` and `packages/emails/tsconfig.json`; no project-local skill conventions are detected under `.codex/skills/` or `.agents/skills/`.
## Function Design
- Pass `context.Context` to database/background work; HTTP handlers accept `echo.Context`: `apps/backend/internal/testing/transaction.go`, `apps/backend/internal/lib/job/handlers.go`, `apps/backend/internal/handler/health.go`.
- Use explicit dependency pointers in constructors, chiefly `*server.Server`: `apps/backend/internal/handler/health.go`, `apps/backend/internal/service/auth.go`.
- TypeScript optional options use destructuring with defaults: `getSecurityMetadata` in `packages/openapi/src/utils.ts`; React templates destructure typed props in `packages/emails/src/templates/welcome.tsx`.
- Use `(value, error)` for fallible Go factories and `error` for effects; test setup adds an explicit cleanup closure: `apps/backend/internal/server/server.go`, `apps/backend/internal/testing/helpers.go`.
- Use `echo.HandlerFunc` factories for middleware/handler wrappers: `apps/backend/internal/handler/base.go`.
- Schema helpers return typed `z.ZodSchema` values; metadata helpers return inferred objects: `packages/zod/src/utils.ts`, `packages/openapi/src/utils.ts`.
## Module Design
- Keep application packages under `apps/backend/internal/`; dependencies are collected through `Server`, `Services`, `Handlers`, and `Repositories` constructors in their respective packages. `apps/backend/internal/repository/repositories.go` is an empty scaffold, not an implemented data-access pattern.
- Export reusable TypeScript schemas, contracts, and helpers through named exports: `packages/zod/src/health.ts`, `packages/openapi/src/contracts/health.ts`, `packages/openapi/src/utils.ts`.
- Email templates provide both a named component and default export for React Email, plus `PreviewProps`: `packages/emails/src/templates/welcome.tsx`.
- Add shared schema exports to `packages/zod/src/index.ts`, which also installs the Zod OpenAPI extension.
- Add new API contracts to the `c.router` aggregation in `packages/openapi/src/contracts/index.ts`; `packages/openapi/src/index.ts` generates the complete document from that router.
- Respect package export maps in `packages/zod/package.json` and `packages/openapi/package.json`; the latter's runtime contract path is singular `dist/contract/index.js`, while source/build directory naming is plural `contracts`, so do not treat the map as validated.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

## System Overview
```text
```
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
- Wire dependencies explicitly in `apps/backend/cmd/flux/main.go`; registry constructors collect handlers, services, and repositories.
- `apps/backend/internal/server/server.go` is a shared concrete resource container passed into most packages, rather than narrow dependency interfaces.
- Middleware centralizes request identity, tracing, logging, CORS, recovery, throttling, and error translation in `apps/backend/internal/middleware/`.
- The active health route calls infrastructure directly; the repository layer in `apps/backend/internal/repository/repositories.go` contains no database operations.
- TypeScript contracts in `packages/openapi/src/contracts/` document the Go API; they do not generate its handlers or enforce runtime request validation.
## Layers
- Purpose: Configure and own long-lived clients and server lifecycle.
- Location: `apps/backend/cmd/flux/`, `apps/backend/internal/server/`, `apps/backend/internal/config/`, `apps/backend/internal/database/`, `apps/backend/internal/logger/`.
- Contains: Startup, environment parsing, pgx pooling, embedded migration runner, logging/APM setup.
- Depends on: PostgreSQL, Redis, New Relic SDK, Go HTTP server.
- Used by: Handler, service, middleware, and repository constructors.
- Purpose: Turn requests into HTTP responses and apply shared behavior.
- Location: `apps/backend/internal/router/`, `apps/backend/internal/middleware/`, `apps/backend/internal/handler/`, `apps/backend/internal/validation/`.
- Contains: Route registration, global error handler, auth adapter, typed handler helpers, health/docs handlers.
- Depends on: Echo, server resources, `apps/backend/internal/errs/`, `apps/backend/internal/sqlerr/`.
- Used by: `apps/backend/cmd/flux/main.go` through `router.NewRouter`.
- Purpose: Provide extension points for application behavior and persistence.
- Location: `apps/backend/internal/service/`, `apps/backend/internal/repository/`, `apps/backend/internal/model/`.
- Contains: Clerk setup, job reference, empty repository registry, shared UUID/timestamp/pagination model types.
- Depends on: Server resources and job infrastructure.
- Used by: Composition and handler constructors; registry parameters are currently unused where no feature needs them.
- Purpose: Execute queued welcome-email deliveries.
- Location: `apps/backend/internal/lib/job/`, `apps/backend/internal/lib/email/`.
- Contains: Asynq producer/task definition, consumer, Resend adapter, HTML template renderer.
- Depends on: Redis, backend template files, Resend configuration.
- Used by: Server startup and `service.Services.Job`; no routed feature enqueues a task.
- Purpose: Maintain schema/contract and email source assets.
- Location: `packages/zod/`, `packages/openapi/`, `packages/emails/`.
- Contains: Zod response schemas, ts-rest contracts, OpenAPI generation, React Email templates.
- Depends on: Workspace imports and package-specific build commands.
- Used by: OpenAPI generation and email export; no frontend app consumes them in this tree.
## Data Flow
### Primary Request Path
### Background Email Flow
### Contract and Asset Flow
- PostgreSQL connection state is held by `database.Database.Pool`; the sole migration `apps/backend/internal/database/migrations/001_setup.sql` contains no schema statements.
- Redis stores Asynq queues; the HTTP rate limiter is process-local memory (`apps/backend/internal/router/router.go`, `apps/backend/internal/lib/job/job.go`).
- Request identity, user claims, and contextual logger use Echo context; logging/APM also uses request `context.Context` (`apps/backend/internal/middleware/context.go`, `apps/backend/internal/middleware/auth.go`).
## Key Abstractions
- Purpose: Share configured infrastructure and own lifecycle.
- Examples: `apps/backend/internal/server/server.go`, `apps/backend/internal/handler/base.go`.
- Pattern: Explicit constructor injection of `*server.Server`.
- Purpose: Wrap a request implementing `Validate() error` with bind, validation, metrics, and response output.
- Examples: `apps/backend/internal/handler/base.go`, `apps/backend/internal/validation/utils.go`.
- Pattern: Generic `Handle`, `HandleNoContent`, and `HandleFile` functions backed by `ResponseHandler`; the active health/docs handlers bypass these wrappers.
- Purpose: Standard JSON error envelope with status/code/message, field errors, and optional action.
- Examples: `apps/backend/internal/errs/http.go`, `apps/backend/internal/errs/types.go`, `apps/backend/internal/sqlerr/handler.go`.
- Pattern: Construct typed errors and return them; central middleware maps wrapped database/Echo errors using `errors.As`.
- Purpose: Compose HTTP documentation from typed schemas.
- Examples: `packages/openapi/src/contracts/index.ts`, `packages/openapi/src/contracts/health.ts`.
- Pattern: `initContract().router(...)`; OpenAPI security metadata exists in `packages/openapi/src/utils.ts` but the health route is public.
## Entry Points
- Location: `apps/backend/cmd/flux/main.go`.
- Triggers: `go run ./cmd/flux` or `task run` from `apps/backend/`.
- Responsibilities: Configuration, optional non-local migrations, resource/layer wiring, HTTP startup, interrupt-driven shutdown.
- Location: `packages/openapi/src/gen.ts`.
- Triggers: Package `gen` script in `packages/openapi/package.json`.
- Responsibilities: Render and write API documentation assets; separate from the package `build` script.
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
### Shared mutable request captured by handler adapters
### Dependency status does not match overall health
## Error Handling
- Use `errs.HTTPError` constructors for client-facing application failures (`apps/backend/internal/errs/types.go`).
- Wrap infrastructure failures with context using `%w` or pkg/errors (`apps/backend/internal/server/server.go`, `apps/backend/internal/lib/email/client.go`).
- Convert PostgreSQL constraint/no-row errors via `sqlerr.HandleError` and normalize unknown errors to a generic HTTP 500 (`apps/backend/internal/sqlerr/handler.go`, `apps/backend/internal/middleware/global.go`).
- Echo recovery handles panics; validation currently assumes specific binding-error text and supported validation-error types (`apps/backend/internal/middleware/global.go`, `apps/backend/internal/validation/utils.go`).
- Queue consumers return delivery failures for Asynq retry handling (`apps/backend/internal/lib/job/handlers.go`).
## Cross-Cutting Concerns
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, `.github/skills/`, or `.codex/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
