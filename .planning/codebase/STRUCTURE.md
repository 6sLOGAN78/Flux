# Codebase Structure

**Analysis Date:** 2026-10-05

## Directory Layout

```text
flux/
├── apps/
│   └── backend/                  # Standalone Go module and runtime assets
│       ├── cmd/flux/main.go       # Backend composition root
│       ├── internal/
│       │   ├── config/           # Environment loading and observability settings
│       │   ├── database/         # pgx pool, tern runner, embedded migrations
│       │   │   └── migrations/
│       │   ├── errs/             # API error envelopes and constructors
│       │   ├── handler/          # Health/docs handlers and generic adapters
│       │   ├── lib/
│       │   │   ├── email/        # Resend client and template helpers
│       │   │   ├── job/          # Asynq tasks and consumer
│       │   │   └── utils/        # Small shared Go helpers
│       │   ├── logger/           # Zerolog and New Relic setup
│       │   ├── middleware/       # Auth, context, tracing, errors, request policies
│       │   ├── model/            # Shared UUID/timestamp/pagination types
│       │   ├── repository/       # Empty repository registry scaffold
│       │   ├── router/           # HTTP setup and system routes
│       │   ├── server/           # Resource container and HTTP lifecycle
│       │   ├── sqlerr/           # PostgreSQL error conversion
│       │   ├── testing/          # Containers, transactions, assertions, test server
│       │   └── validation/       # Request binding and validation utilities
│       ├── static/               # Docs HTML and generated OpenAPI JSON
│       ├── templates/emails/     # Exported HTML consumed by Go
│       ├── go.mod
│       ├── go.sum
│       ├── taskfile.yml
│       └── .golangci.yml
├── packages/
│   ├── zod/src/                  # Shared schemas and pagination types
│   ├── openapi/src/              # API contracts and OpenAPI generator
│   │   └── contracts/
│   └── emails/src/templates/     # React Email authoring source
├── package.json                  # Bun workspace and Turbo entry scripts
├── turbo.json                    # JavaScript workspace task graph
├── bun.lock
├── package-lock.json
├── readme.md                     # Project/tooling explanation
├── spec.md                       # Product requirements, not implementation inventory
└── .planning/codebase/           # Codebase reference documents
```

Only `apps/backend/` contains application source. The shared packages are `packages/zod/`, `packages/openapi/`, and `packages/emails/`; frontend, redirect, analytics, and separate worker application directories described by `spec.md` are not present.

## Directory Purposes

**`apps/backend/cmd/flux/`:**
- Purpose: Keep executable startup isolated from internal packages.
- Contains: Configuration/resource/layer wiring and interrupt-driven shutdown.
- Key files: `apps/backend/cmd/flux/main.go`.

**`apps/backend/internal/`:**
- Purpose: Go-module-private application and infrastructure packages.
- Contains: Transport, service/repository scaffolds, model types, database, middleware, delivery, test helpers.
- Key files: `apps/backend/internal/server/server.go`, `apps/backend/internal/router/router.go`, `apps/backend/internal/handler/base.go`.
- Use the existing layer-oriented directories for current additions; `spec.md`'s domain package layout is not the current source tree.

**`apps/backend/internal/database/`:**
- Purpose: Manage PostgreSQL connectivity and migration execution.
- Contains: pgx pool setup, query tracers, tern runner and numbered SQL files.
- Key files: `apps/backend/internal/database/database.go`, `apps/backend/internal/database/migrator.go`, `apps/backend/internal/database/migrations/001_setup.sql`.
- The existing migration is an empty up/down template, not a domain schema.

**`apps/backend/internal/handler/` and `apps/backend/internal/router/`:**
- Purpose: HTTP transport and route assembly.
- Contains: Health/docs handlers, shared response adapters, handler registry, route registration.
- Key files: `apps/backend/internal/handler/health.go`, `apps/backend/internal/handler/handlers.go`, `apps/backend/internal/router/system.go`.

**`apps/backend/internal/lib/`:**
- Purpose: Backend delivery integrations and reusable small helpers.
- Contains: Redis-backed Asynq jobs, Resend email rendering/sending, general utilities.
- Key files: `apps/backend/internal/lib/job/job.go`, `apps/backend/internal/lib/job/email_tasks.go`, `apps/backend/internal/lib/email/client.go`, `apps/backend/internal/lib/utils/utils.go`.

**`apps/backend/internal/testing/`:**
- Purpose: Reusable Go test infrastructure rather than production behavior.
- Contains: PostgreSQL container fixture, simplified test server, transaction helpers, assertions and JSON helpers.
- Key files: `apps/backend/internal/testing/container.go`, `apps/backend/internal/testing/server.go`, `apps/backend/internal/testing/transaction.go`.
- No `_test.go` files are detected in the application tree.

**`packages/zod/`:**
- Purpose: Export shared validation schemas and TypeScript types.
- Contains: Health response schema, pagination utility, barrel initialization.
- Key files: `packages/zod/src/index.ts`, `packages/zod/src/health.ts`, `packages/zod/src/utils.ts`.

**`packages/openapi/`:**
- Purpose: Define ts-rest contracts and render OpenAPI JSON assets.
- Contains: Contracts, operation metadata mapping, security metadata helper, generator.
- Key files: `packages/openapi/src/contracts/index.ts`, `packages/openapi/src/contracts/health.ts`, `packages/openapi/src/index.ts`, `packages/openapi/src/gen.ts`.

**`packages/emails/`:**
- Purpose: Author and preview HTML email layouts in React.
- Contains: Welcome email TSX and React Email commands.
- Key files: `packages/emails/src/templates/welcome.tsx`, `packages/emails/package.json`.
- Export output belongs in `apps/backend/templates/emails/`, as configured by the package export script.

## Key File Locations

**Entry Points:**
- `apps/backend/cmd/flux/main.go`: Executable composition and lifecycle.
- `apps/backend/internal/router/router.go`: Echo setup and versioned group scaffold.
- `apps/backend/internal/router/system.go`: Active `/status`, `/docs`, `/static` endpoints.
- `packages/openapi/src/gen.ts`: Generate JSON documentation in two locations.
- `packages/emails/src/templates/welcome.tsx`: Current email preview/export component.

**Configuration:**
- `package.json`: Root Bun workspace patterns and Turbo scripts.
- `turbo.json`: JavaScript dependency task ordering and caching.
- `apps/backend/go.mod`: Independent Go module and dependencies; backend has no package manifest to expose Go tasks to Turbo.
- `apps/backend/taskfile.yml`: Go run, formatting/module maintenance, migration commands.
- `apps/backend/.golangci.yml`: Go linter configuration.
- `apps/backend/internal/config/config.go`: `FLUX_` environment configuration mapping and validation.
- `apps/backend/internal/config/observability.go`: Logging/APM/health settings and defaults.
- `packages/zod/tsconfig.json`, `packages/openapi/tsconfig.json`, `packages/emails/tsconfig.json`: Package-local TypeScript compilation settings.
- `.gitignore`, `apps/backend/.gitignore`: Generated output and environment exclusions.

**Core Logic:**
- `apps/backend/internal/server/server.go`: HTTP, database, Redis, jobs ownership.
- `apps/backend/internal/handler/base.go`: Generic validated request handling and response strategies.
- `apps/backend/internal/middleware/global.go`: Request policies, logging and error serialization.
- `apps/backend/internal/middleware/auth.go`: Optional Clerk authorization adapter.
- `apps/backend/internal/service/services.go`: Application service registry.
- `apps/backend/internal/repository/repositories.go`: Empty persistence registry.
- `apps/backend/internal/model/base.go`: UUID, timestamps and pagination representation.
- `apps/backend/internal/sqlerr/handler.go`: Database error-to-HTTP mapping.

**Testing:**
- `apps/backend/internal/testing/helpers.go`: Shared test setup and helpers.
- `apps/backend/internal/testing/container.go`: Disposable PostgreSQL and migrations fixture.
- `apps/backend/internal/testing/assertions.go`: Shared assertions.
- `apps/backend/internal/testing/transaction.go`: Transaction/rollback wrappers.
- No application test suites or frontend/E2E test directories are detected under `apps/` or `packages/`.

## Naming Conventions

**Files:**
- Go files use lowercase names with underscores for compound names: `apps/backend/internal/middleware/request_id.go`, `apps/backend/internal/lib/job/email_tasks.go`.
- Registry files use plural layer names: `apps/backend/internal/handler/handlers.go`, `apps/backend/internal/service/services.go`, `apps/backend/internal/repository/repositories.go`.
- TypeScript modules use lowercase feature names and `index.ts` barrels: `packages/zod/src/health.ts`, `packages/openapi/src/contracts/index.ts`.
- React Email templates use lowercase `.tsx` filenames and PascalCase component exports: `packages/emails/src/templates/welcome.tsx` exports `WelcomeEmail`.
- SQL migrations use numbered prefixes: `apps/backend/internal/database/migrations/001_setup.sql`.
- Existing human documentation uses lowercase filenames (`readme.md`, `spec.md`); codebase maps use uppercase names in `.planning/codebase/`.

**Directories:**
- Workspace applications/packages are grouped by responsibility under `apps/` and `packages/` (`package.json`).
- Go layers use singular lowercase directories matching package names: `apps/backend/internal/handler/`, `apps/backend/internal/service/`, `apps/backend/internal/repository/`.
- Third-party delivery helpers are grouped under `apps/backend/internal/lib/email/` and `apps/backend/internal/lib/job/`.

## Where to Add New Code

**New Feature:**
- Primary code: Add request/response handling in `apps/backend/internal/handler/`, application behavior in `apps/backend/internal/service/`, and database queries in `apps/backend/internal/repository/`.
- Wire constructors into `apps/backend/internal/handler/handlers.go`, `apps/backend/internal/service/services.go`, `apps/backend/internal/repository/repositories.go` and register routes in `apps/backend/internal/router/`.
- Shared persisted/base representations belong in `apps/backend/internal/model/`; schema changes belong in `apps/backend/internal/database/migrations/`.
- TypeScript schemas belong in `packages/zod/src/` and contract declarations in `packages/openapi/src/contracts/`; update their `index.ts` exports and regenerate OpenAPI through `packages/openapi/src/gen.ts`.
- Tests: Place Go `_test.go` files beside the package being tested and reuse `apps/backend/internal/testing/`; this is placement guidance because no checked-in test suites establish an existing placement convention.

**New Component/Module:**
- HTTP middleware: `apps/backend/internal/middleware/`; assemble it in `apps/backend/internal/middleware/middlewares.go` and attach it in `apps/backend/internal/router/router.go`.
- Background task: Define payload/task constructor in `apps/backend/internal/lib/job/`, add consumer logic there, register in `JobService.Start` in `apps/backend/internal/lib/job/job.go`.
- Email: Add authoring TSX to `packages/emails/src/templates/`, export backend HTML, and add template identifiers/delivery helpers in `apps/backend/internal/lib/email/`.
- Configuration: Extend typed structures and validation in `apps/backend/internal/config/`.
- New executable: Use `apps/backend/cmd/<name>/` for a Go binary only when a separate process is required; the current entry is `apps/backend/cmd/flux/main.go`.

**Utilities:**
- Shared Go helpers: `apps/backend/internal/lib/utils/`; keep HTTP-specific validation in `apps/backend/internal/validation/` and SQL error handling in `apps/backend/internal/sqlerr/`.
- Shared TypeScript schema helpers: `packages/zod/src/utils.ts`.
- Contract/security helpers: `packages/openapi/src/utils.ts`.

## Special Directories

Repository tracking status cannot be established from the available `.git/` metadata; the statements below distinguish source/assets from ignored build output using `.gitignore`, not a verified git index.

**`apps/backend/static/`:**
- Purpose: Runtime documentation HTML and OpenAPI JSON; served by `apps/backend/internal/router/system.go`.
- Generated: `apps/backend/static/openapi.json` is generated by `packages/openapi/src/gen.ts`; `apps/backend/static/openapi.html` is a runtime source asset.
- Committed: Not verifiable; files exist and are not excluded by `.gitignore`.

**`apps/backend/templates/emails/`:**
- Purpose: HTML templates loaded at runtime by `apps/backend/internal/lib/email/client.go`.
- Generated: Yes, from `packages/emails/src/templates/` through `packages/emails/package.json`.
- Committed: Not verifiable; assets exist and are not excluded by `.gitignore`.

**`apps/backend/internal/database/migrations/`:**
- Purpose: Numbered tern up/down SQL, embedded by `apps/backend/internal/database/migrator.go`.
- Generated: No; migration skeleton creation is supported by `apps/backend/taskfile.yml`.
- Committed: Not verifiable; source directory is not excluded by `.gitignore`.

**`packages/*/dist/`, `node_modules/`, `.turbo/`:**
- Purpose: Compiled JavaScript/declarations, installed dependencies, task caches (`packages/zod/tsconfig.json`, `packages/openapi/tsconfig.json`, `turbo.json`).
- Generated: Yes.
- Committed: Excluded by `.gitignore`; edit `packages/*/src/` rather than build output.

**`.planning/codebase/`:**
- Purpose: Generated reference maps for planning/execution agents, including `.planning/codebase/ARCHITECTURE.md` and `.planning/codebase/STRUCTURE.md`.
- Generated: Yes, from source analysis.
- Committed: Not verifiable; not excluded by `.gitignore`.

---

*Structure analysis: 2026-10-05*
