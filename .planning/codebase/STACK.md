# Technology Stack

**Analysis Date:** 2026-10-05

## Languages

**Primary:**
- Go 1.25.5 module language directive — HTTP backend, middleware, database access, jobs, email delivery, and testing helpers in `apps/backend/go.mod` and `apps/backend/internal/`.

**Secondary:**
- TypeScript — reusable schemas and REST contracts in `packages/zod/src/` and `packages/openapi/src/`; email template TSX in `packages/emails/src/templates/welcome.tsx`.
- SQL — Tern migration files in `apps/backend/internal/database/migrations/`; `001_setup.sql` contains only migration scaffold comments, without application tables.
- HTML — exported email template in `apps/backend/templates/emails/welcome.html` and API reference page in `apps/backend/static/openapi.html`.
- YAML/JSON — task runner, linting, compiler, and workspace configuration in `apps/backend/taskfile.yml`, `apps/backend/.golangci.yml`, `turbo.json`, and package manifests.

## Runtime

**Environment:**
- Go backend executable: `apps/backend/cmd/flux/main.go`; `apps/backend/go.mod` declares Go 1.25.5.
- Node.js >=22 declared in root `package.json`; Bun scripts and package management support the TypeScript workspaces.
- No frontend application runtime is implemented: `apps/frontend/` is empty. React dependencies in `package.json` are not evidence of a running dashboard.

**Package Manager:**
- Bun 1.2.13 declared by `package.json` `packageManager`; workspace globs cover `apps/*` and `packages/*`.
- Lockfiles present: `bun.lock`, `packages/emails/bun.lock`, and `package-lock.json`. The repository contains both Bun and npm lockfiles; use the declared Bun manager for workspace operations.
- Go modules: `apps/backend/go.mod` and `apps/backend/go.sum`.

## Frameworks

**Core:**
- Echo v4.15.2 — HTTP routing and middleware; `apps/backend/internal/router/router.go`, declared in `apps/backend/go.mod`.
- React 19.1.0 in email workspace; React/React DOM ^19.2.6 in root manifest — implemented UI usage is the email component in `packages/emails/src/templates/welcome.tsx`.
- React Email ^6.3.3 and `@react-email/components` 0.0.34 — preview and HTML export in `packages/emails/package.json`.
- Zod ^3.24.2 — schema validation/types in `packages/zod/src/health.ts`; `@anatine/zod-openapi` ^2.2.7 attaches documentation metadata.
- `@ts-rest/core` and `@ts-rest/open-api` ^3.52.1 — REST contracts and OpenAPI generation in `packages/openapi/src/contracts/health.ts` and `packages/openapi/src/index.ts`.

**Testing:**
- Go standard `testing` — backend helpers under `apps/backend/internal/testing/`.
- Testify v1.11.1 — assertions used in `apps/backend/internal/testing/container.go` and `apps/backend/internal/testing/assertions.go`.
- Testcontainers Go v0.42.0 — PostgreSQL integration helper starts `postgres:15-alpine` in `apps/backend/internal/testing/container.go`.
- No runnable `_test.go` suites or JavaScript test runner configuration detected under `apps/backend/` or `packages/`; distinguish helper infrastructure from actual test coverage.

**Build/Dev:**
- Turborepo ^2.9.6 — root orchestration in `package.json` and `turbo.json`; build dependencies use `^build`, dev is persistent and uncached.
- TypeScript ^6.0.3 at root; ^5.8.2 in `packages/zod/package.json` and `packages/openapi/package.json`. These are manifest ranges, not a claim of installed resolution.
- `tsc-alias` ^1.8.12 — rewrites emitted aliases during schema/contract compilation; `packages/zod/package.json`, `packages/openapi/package.json`.
- `tsx` ^4.19.3 — contract generator execution in `packages/openapi/package.json`.
- `concurrently` ^9.1.2 and `wait-on` ^8.0.3 — compiler/watch coordination; OpenAPI dev waits for `packages/zod/dist/index.js`.
- Taskfile v3 schema — backend tasks `run`, `migrations:new`, `migrations:up`, and `tidy` in `apps/backend/taskfile.yml`; Task executable version is not pinned.
- GolangCI-Lint configuration is present at `apps/backend/.golangci.yml`; linter executable version is not pinned by `apps/backend/go.mod`.

## Key Dependencies

**Critical:**
- `github.com/jackc/pgx/v5` v5.9.2 — PostgreSQL pool and driver in `apps/backend/internal/database/database.go`; no ORM is used.
- `github.com/jackc/tern/v2` v2.4.1 — embedded migrations and `schema_version` tracking in `apps/backend/internal/database/migrator.go`.
- `github.com/redis/go-redis/v9` v9.14.1 — Redis client and health checks in `apps/backend/internal/server/server.go` and `apps/backend/internal/handler/health.go`.
- `github.com/hibiken/asynq` v0.26.0 — Redis-backed worker and welcome-email jobs in `apps/backend/internal/lib/job/job.go` and `email_tasks.go`.
- `github.com/clerk/clerk-sdk-go/v2` v2.6.0 — authentication setup and reusable header authorization middleware in `apps/backend/internal/service/auth.go` and `apps/backend/internal/middleware/auth.go`.
- `github.com/resend/resend-go/v2` v2.28.0 — transactional email transport in `apps/backend/internal/lib/email/client.go`.

**Infrastructure:**
- Koanf v2.3.4 plus env provider v1.1.0 — hierarchical environment configuration in `apps/backend/internal/config/config.go`.
- Godotenv v1.5.1 — automatic dotenv loading imported by `apps/backend/internal/config/config.go`.
- Go Playground validator v10.30.2 — configuration and request validation in `apps/backend/internal/config/config.go` and `apps/backend/internal/validation/utils.go`.
- Zerolog v1.35.1 — structured logs in `apps/backend/internal/logger/logger.go`; pgx-zerolog adapter is declared in `apps/backend/go.mod`.
- New Relic Go agent v3.43.3, Echo integration v1.1.5, pgx integration v1.3.4, Redis integration v1.1.2, pkgerrors integration v1.1.0 — optional instrumentation in `apps/backend/internal/logger/logger.go`, `apps/backend/internal/middleware/tracing.go`, `apps/backend/internal/database/database.go`, and `apps/backend/internal/server/server.go`.

## Configuration

**Environment:**
- `apps/backend/internal/config/config.go` loads environment names prefixed `FLUX_`, removes that prefix, lowercases the remainder, and uses `.` as the nesting delimiter. Nested names therefore use dots, for example `FLUX_DATABASE.HOST`; underscores inside field names remain underscores.
- Required sections: primary environment, server timeouts/CORS, PostgreSQL connection and pool settings, Clerk secret key, Redis address, Resend API key; see structs and validation tags in `apps/backend/internal/config/config.go`.
- Observability defaults and custom validation live in `apps/backend/internal/config/observability.go`; blank New Relic license disables the agent in `apps/backend/internal/logger/logger.go`.
- `apps/backend/.env` and `apps/backend/.env.sample` are present; contents are not inspected. `.gitignore` excludes `.env`.
- No project skills are present under `.agents/` or `.codex/`.

**Build:**
- `package.json` and `turbo.json`: workspace build/dev/cleanup orchestration. Root scripts also name lint/typecheck/format tasks; inspect each package for actual script support rather than assuming every root task executes checks.
- `packages/zod/tsconfig.json`, `packages/openapi/tsconfig.json`, `packages/emails/tsconfig.json`: strict NodeNext ESM, ES2022, declarations/source maps, `dist/` output; email config uses `react-jsx`.
- `packages/openapi/src/gen.ts`: writes generated JSON to `packages/openapi/openapi.json` and `apps/backend/static/openapi.json`.
- `packages/emails/package.json`: `export` writes HTML to `apps/backend/templates/emails/`; this is a separate task from the root build.
- `apps/backend/taskfile.yml`: `task run` executes `go run ./cmd/flux`; migrations CLI consumes `FLUX_DB_DSN` separately from the application's structured configuration.

## Platform Requirements

**Development:**
- Go compatible with `apps/backend/go.mod`, Bun 1.2.13, and Node >=22 per `package.json`; Task and Tern CLIs are needed only for the commands in `apps/backend/taskfile.yml` that invoke them.
- Accessible PostgreSQL and Redis for `apps/backend/internal/server/server.go`; Docker-compatible container runtime for `apps/backend/internal/testing/container.go`.
- Run the backend with `apps/backend/` as working directory: static documents and email templates use relative paths in `apps/backend/internal/router/system.go` and `apps/backend/internal/lib/email/client.go`.

**Production:**
- Go HTTP executable from `apps/backend/cmd/flux/main.go` plus runtime files in `apps/backend/static/` and `apps/backend/templates/emails/`; migration SQL is embedded by `apps/backend/internal/database/migrator.go`.
- PostgreSQL, Redis, Clerk, and Resend configuration are wired in `apps/backend/internal/config/config.go`; New Relic is optional.
- Hosting provider, Dockerfile, Compose stack, Kubernetes manifests, and CI deployment workflow: not detected. `spec.md` contains proposed platform architecture; it is not an implementation manifest.

---

*Stack analysis: 2026-10-05*
