# Coding Conventions

**Analysis Date:** 2026-10-05

## Naming Patterns

**Files:**
- Use lowercase Go files named for their responsibility, with underscores for multiple words: `apps/backend/internal/middleware/request_id.go`, `apps/backend/internal/lib/job/email_tasks.go`.
- Use lowercase TypeScript feature files: `packages/zod/src/health.ts`, `packages/openapi/src/contracts/health.ts`. Email templates use lowercase `.tsx` files: `packages/emails/src/templates/welcome.tsx`.
- Use numbered SQL migration files under `apps/backend/internal/database/migrations/`, following `001_setup.sql` and its tern up/down separator.

**Functions:**
- Use PascalCase for exported Go functions and methods, camelCase for private helpers: `NewHealthHandler` and `CheckHealth` in `apps/backend/internal/handler/health.go`; `generateErrorCode` in `apps/backend/internal/sqlerr/handler.go`.
- Use `New<Type>` constructors with dependencies supplied explicitly: `NewServices` in `apps/backend/internal/service/services.go`, `NewGlobalMiddlewares` in `apps/backend/internal/middleware/global.go`.
- Use camelCase TypeScript helpers and contract operations: `schemaWithPagination` in `packages/zod/src/utils.ts`, `getSecurityMetadata` in `packages/openapi/src/utils.ts`, `getHealth` in `packages/openapi/src/contracts/health.ts`.
- Use PascalCase React components: `WelcomeEmail` in `packages/emails/src/templates/welcome.tsx`.

**Variables:**
- Use short Go receiver/dependency names in small scopes (`s`, `h`, `cfg`, `ctx`, `err`) and descriptive names for longer operations: `apps/backend/internal/server/server.go`, `apps/backend/internal/handler/health.go`.
- Use camelCase TypeScript locals and props: `openApiSecurity` in `packages/openapi/src/utils.ts`, `userFirstName` in `packages/emails/src/templates/welcome.tsx`.
- Preserve existing wire field spellings rather than deriving them from identifiers. `apps/backend/internal/model/base.go` uses camelCase JSON timestamps and snake_case DB fields; `apps/backend/internal/lib/job/email_tasks.go` uses snake_case payload JSON; health checks use `response_time` in `packages/zod/src/health.ts`.

**Types:**
- Use PascalCase structs and interfaces: `HTTPError`, `FieldError`, `Action` in `apps/backend/internal/errs/http.go`; `WelcomeEmailProps` in `packages/emails/src/templates/welcome.tsx`.
- Prefix exported Zod schemas with `Z`: `ZHealthResponse` in `packages/zod/src/health.ts`. Keep nested implementation schemas private where they are not reused.
- Use typed string constants for enums: `ActionTypeRedirect` in `apps/backend/internal/errs/http.go`, `TaskWelcome` in `apps/backend/internal/lib/job/email_tasks.go`.
- Use generic response/helper types where applicable: `PaginatedResponse` in both `apps/backend/internal/model/base.go` and `packages/zod/src/utils.ts`, `Ptr[T any]` in `apps/backend/internal/testing/helpers.go`.

## Code Style

**Formatting:**
- Go formatting is exposed by the `tidy` task in `apps/backend/taskfile.yml` using `go fmt ./...`. Use Go formatting for changed Go files; sample helpers contain spacing/import-order inconsistencies, so do not copy those inconsistencies.
- TypeScript source uses two spaces, double quotes, semicolons, trailing commas in multiline objects, and arrow functions: `packages/zod/src/health.ts`, `packages/openapi/src/utils.ts`, `packages/emails/src/templates/welcome.tsx`.
- No Prettier, Biome, or ESLint configuration is detected in the source inventory. `package.json` has `format:check`, `format:fix`, `lint`, and `lint:fix` Turbo scripts, but the package manifests do not implement those tasks; these are not evidence of enforced TypeScript formatting.
- `turbo.json` defines a `format` task while the root script requests `format:check`. Check the task configuration before relying on root formatting commands.

**Linting:**
- Backend rules are declared in `apps/backend/.golangci.yml` with configuration version `2` and a three-minute timeout; no installed golangci-lint version or passing lint result is established by this map.
- The configured checks include error handling (`errcheck`, `errorlint`, `nilerr`), security (`gosec`, `bidichk`), complexity (`cyclop`, `gocognit`, `funlen`), naming (`revive`, `stylecheck`), imports (`goimports`), context/resource use (`noctx`, `bodyclose`, `sqlclosecheck`), and tests (`testifylint`, `testpackage`, `tparallel`).
- Configured function limits are 100 lines and 50 statements, excluding comments; cyclomatic maximum is 30 with package average 10; cognitive threshold is 20: `apps/backend/.golangci.yml`.
- `nolint` directives must name specific checks and normally explain the suppression; `funlen`, `gocognit`, and `lll` are explanation exceptions in `apps/backend/.golangci.yml`.
- Tests matching `_test.go` have selected exclusions, including `errcheck`, `funlen`, `gosec`, and `noctx`: `apps/backend/.golangci.yml`. Helpers under `apps/backend/internal/testing/` are ordinary `.go` files and do not match that exclusion.
- TypeScript strictness is configured in `packages/zod/tsconfig.json`, `packages/openapi/tsconfig.json`, and `packages/emails/tsconfig.json`: `strict`, `noUncheckedIndexedAccess`, `noImplicitOverride`, `isolatedModules`, and `verbatimModuleSyntax`.

## Import Organization

**Order:**
1. Go standard-library imports first, then a blank line and application/external imports: `apps/backend/internal/server/server.go`.
2. Existing Go files vary in whether project imports precede external imports. Use the configured import formatter rather than hand-maintaining inconsistent grouping: `apps/backend/.golangci.yml`.
3. TypeScript generally imports external dependencies before relative/local modules: `packages/openapi/src/gen.ts`, `packages/openapi/src/contracts/health.ts`. There is no detected import-order lint rule.

**Path Aliases:**
- Backend imports use the module prefix `github.com/6sLOGAN78/flux/internal/...`, declared in `apps/backend/go.mod`.
- TypeScript `@/*` resolves into each package's `src` directory via `packages/zod/tsconfig.json`, `packages/openapi/tsconfig.json`, and `packages/emails/tsconfig.json`; `packages/openapi/src/contracts/health.ts` uses `@/utils.js`.
- Preserve `.js` suffixes in relative TypeScript imports for NodeNext ESM output: `packages/zod/src/index.ts`, `packages/openapi/src/contracts/index.ts`.
- Cross-package imports use workspace package names such as `@flux/zod` in `packages/openapi/src/contracts/health.ts` and `workspace:*` in `packages/openapi/package.json`.

## Error Handling

**Patterns:**
- Return Go errors and wrap failures with operation context and `%w`: `apps/backend/internal/server/server.go`, `apps/backend/internal/lib/job/handlers.go`, `apps/backend/internal/testing/transaction.go`.
- Construct API errors through `apps/backend/internal/errs/types.go`; the JSON shape in `apps/backend/internal/errs/http.go` includes code, message, status, override, field errors, and optional action.
- Let the Echo error boundary serialize handler failures. `GlobalErrorHandler` in `apps/backend/internal/middleware/global.go` recognizes typed HTTP/Echo errors, maps SQL errors through `apps/backend/internal/sqlerr/handler.go`, logs the original failure, and avoids writing an already committed response.
- Match wrapped errors with `errors.As`/`errors.Is`, as in `apps/backend/internal/sqlerr/handler.go`.
- Generic endpoint wrappers in `apps/backend/internal/handler/base.go` centralize validation, timing, logging, and JSON/file/no-content responses. Health and documentation handlers also use direct Echo methods; do not assume every route passes through the wrapper (`apps/backend/internal/router/system.go`).
- Request payloads handled by wrappers implement `validation.Validatable` and `Validate() error`: `apps/backend/internal/validation/utils.go`. Its bind-error string parsing and custom-error type assertions assume particular shapes; handle new error types deliberately instead of assuming arbitrary errors are supported.
- The TypeScript generator logs asynchronous write failures with `console.error` in `packages/openapi/src/gen.ts`; it does not explicitly set an exit code on failure.

## Logging

**Framework:** Zerolog for backend logging with optional New Relic integration (`apps/backend/internal/logger/logger.go`); console logging in the OpenAPI generator (`packages/openapi/src/gen.ts`).

**Patterns:**
- Use contextual loggers from `middleware.GetLogger(c)` and attach structured operation fields: `apps/backend/internal/handler/health.go`, `apps/backend/internal/handler/base.go`.
- Use fluent `Str`, `Int`, `Dur`, and `Err` fields followed by a short message. Record validation/handler durations through the central wrapper: `apps/backend/internal/handler/base.go`.
- Request logging records identifiers, route/request metadata, latency, status, and severity selected from response status: `apps/backend/internal/middleware/global.go`.
- Guard optional tracing/logger services before recording custom events: `apps/backend/internal/handler/health.go`.
- `PrintJSON` in `apps/backend/internal/lib/utils/utils.go` uses `fmt.Println` as a utility; operational backend code generally uses the structured logger.

## Comments

**When to Comment:**
- Explain exported Go helper responsibilities and non-obvious implementation assumptions: `apps/backend/internal/testing/helpers.go`, `apps/backend/internal/sqlerr/err.go`.
- Add rationale for middleware quirks and fallback behavior, such as status extraction and Redis startup behavior: `apps/backend/internal/middleware/global.go`, `apps/backend/internal/server/server.go`.
- Keep comments aligned with behavior. `WithTransaction` in `apps/backend/internal/testing/transaction.go` commits successful callbacks despite its rollback-oriented summary; use implementation behavior as the authority.

**JSDoc/TSDoc:**
- Not detected in the TypeScript implementation samples (`packages/zod/src/utils.ts`, `packages/openapi/src/utils.ts`, `packages/emails/src/templates/welcome.tsx`). Go uses ordinary doc comments.
- Inline configuration explanations are present in `packages/openapi/tsconfig.json` and `packages/emails/tsconfig.json`; no project-local skill conventions are detected under `.codex/skills/` or `.agents/skills/`.

## Function Design

**Size:** Prefer focused constructors and domain helpers, following `apps/backend/internal/service/auth.go` and `apps/backend/internal/sqlerr/handler.go`; consult declared limits in `apps/backend/.golangci.yml`. Long orchestration functions exist in `apps/backend/internal/handler/base.go` and `apps/backend/internal/testing/container.go`, so the limits are configuration intentions rather than a verified clean baseline.

**Parameters:**
- Pass `context.Context` to database/background work; HTTP handlers accept `echo.Context`: `apps/backend/internal/testing/transaction.go`, `apps/backend/internal/lib/job/handlers.go`, `apps/backend/internal/handler/health.go`.
- Use explicit dependency pointers in constructors, chiefly `*server.Server`: `apps/backend/internal/handler/health.go`, `apps/backend/internal/service/auth.go`.
- TypeScript optional options use destructuring with defaults: `getSecurityMetadata` in `packages/openapi/src/utils.ts`; React templates destructure typed props in `packages/emails/src/templates/welcome.tsx`.

**Return Values:**
- Use `(value, error)` for fallible Go factories and `error` for effects; test setup adds an explicit cleanup closure: `apps/backend/internal/server/server.go`, `apps/backend/internal/testing/helpers.go`.
- Use `echo.HandlerFunc` factories for middleware/handler wrappers: `apps/backend/internal/handler/base.go`.
- Schema helpers return typed `z.ZodSchema` values; metadata helpers return inferred objects: `packages/zod/src/utils.ts`, `packages/openapi/src/utils.ts`.

## Module Design

**Exports:**
- Keep application packages under `apps/backend/internal/`; dependencies are collected through `Server`, `Services`, `Handlers`, and `Repositories` constructors in their respective packages. `apps/backend/internal/repository/repositories.go` is an empty scaffold, not an implemented data-access pattern.
- Export reusable TypeScript schemas, contracts, and helpers through named exports: `packages/zod/src/health.ts`, `packages/openapi/src/contracts/health.ts`, `packages/openapi/src/utils.ts`.
- Email templates provide both a named component and default export for React Email, plus `PreviewProps`: `packages/emails/src/templates/welcome.tsx`.

**Barrel Files:**
- Add shared schema exports to `packages/zod/src/index.ts`, which also installs the Zod OpenAPI extension.
- Add new API contracts to the `c.router` aggregation in `packages/openapi/src/contracts/index.ts`; `packages/openapi/src/index.ts` generates the complete document from that router.
- Respect package export maps in `packages/zod/package.json` and `packages/openapi/package.json`; the latter's runtime contract path is singular `dist/contract/index.js`, while source/build directory naming is plural `contracts`, so do not treat the map as validated.

---

*Convention analysis: 2026-10-05*
