# Phase 2: Tenant-Safe Link Control Plane - Pattern Map

**Mapped:** 2026-10-08
**Scope:** Current implementation, 02-CONTEXT.md, 02-RESEARCH.md and 02-UI-SPEC.md. Historical Phase 1 maps are not implementation evidence.
**Files analyzed:** 22 responsibility groups below; proposed names within unspecified directories are planner directions, not existing files.
**Analogs found:** 18 / 22 groups have foundation matches (5 exact foundation groups, 13 partial/role matches); 4 are greenfield. A foundation match does not imply existing domain behavior.

## File Classification

Backend paths below are relative to `apps/backend/`; brace lists include every named responsibility.

| New/modified file direction | Role | Data flow | Closest current analog | Quality |
|---|---|---|---|---|
| `internal/service/identity.go`, `auth.go` | service | request-response | `internal/service/auth.go`; `app/api.go` injection | partial |
| `internal/middleware/{auth,workspace,origin}.go` | middleware | request-response | `middleware/auth.go`, `global.go` | role-match |
| `internal/service/{workspace,team,link}.go`, permission/lifecycle helpers | service | CRUD | `service/services.go` wiring only | partial |
| `internal/repository/{user,workspace,invitation,link,mutation,audit}.go`, registry | service | CRUD | `repository/repositories.go`; `testing/transaction.go` | partial |
| `internal/model/{user,workspace,invitation,link}.go` if needed | model | transform | `model/base.go` naming only | partial |
| `internal/database/migrations/002_*.sql` onward; `migrator_test.go` | migration/test | batch | `001_setup.sql`, `database/migrator.go` | role-match |
| `internal/handler/{identity,workspace,team,link}.go`, `handlers.go` | controller | request-response | `handler/base.go` | exact adapter |
| `internal/router/{router,product}.go` | route | request-response | `router/router.go` | exact registration |
| `internal/errs/{http,types}.go`, `validation/` additions | utility | transform | `errs/http.go`, `validation/utils.go`, `middleware/global.go` | exact boundary |
| `internal/app/{api,worker}.go`, role tests | provider | event-driven | same files; `app/roles_test.go` | exact ownership |
| `internal/config/config.go`, samples and development instructions | config | transform | existing role-scoped config | role-match |
| `internal/safety/{url,transport}.go`, tests | utility | request-response | none | none |
| `internal/lib/job/invitation_tasks.go`, `invitation_handlers.go`, `job.go` | service | event-driven | `job/email_tasks.go`, `handlers.go` | role-match |
| Invitation delivery repository/dispatcher and encryption helpers | service/utility | event-driven | queue constructor only; no durable outbox | none |
| `packages/emails/src/templates/invitation.tsx`; email adapters/assets | component/service | file-I/O | `welcome.tsx`, email package | role-match |
| `packages/zod/src/{identity,workspace,team,link,error}.ts`, exports/tests | model/test | transform | `packages/zod/src/health.ts` | exact schema |
| `packages/openapi/src/contracts/{identity,workspace,team,link}.ts`, index/security | config | request-response | `contracts/health.ts`, `utils.ts` | role-match |
| `internal/transport/`, codegen config, generated JSON/HTML, `scripts/generate.ts` | config | file-I/O | current generator/manifest | role-match |
| Backend domain/HTTP/worker `*_test.go`, shared product fixtures | test | CRUD/event-driven | `testing/container.go`, `handler/base_test.go` | role-match |
| `apps/frontend/{package.json,tsconfig.json,next.config.ts,proxy.ts,app/layout.tsx,app/globals.css}` | config/provider | request-response | strict workspace manifests only | partial |
| Frontend auth/onboarding/chooser/Links/create/detail/Team/invitation pages; `components/`, `lib/` | component/hook/store | request-response | none; frontend empty | none |
| Frontend browser/provider tests, Playwright config; root gate/scanner adaptations | test/config | event-driven/batch | root gate plumbing only; no browser suite | none for browser behavior |

## Pattern Assignments

### 1. API composition, registries and role ownership

**Analog:** `apps/backend/internal/app/api.go:53–60,125–140,201–247,291–302`.

```go
func NewAPI(ctx context.Context, cfg *config.Config) (*API, error) {
    runtime, err := newRole(ctx, config.RoleAPI, cfg, defaultRoleFactories())
    if err != nil { return nil, err }
    return &API{runtime}, nil
}
// Current defaultRoleRouter, lines 300–302:
services := &service.Services{Job: srv.Job}
return router.NewRouter(srv, &handler.Handlers{OpenAPI: handler.NewOpenAPIHandler(srv)}, services), nil
```

The active router currently bypasses `NewRepositories`, `NewServices` and the handler registry constructor. Wire real repositories→narrow services→handlers here; extending unused registries alone delivers no route. Standard imports use `github.com/6sLOGAN78/flux/internal/...` before third-party dependencies. Preserve `r.own(stage, cleanup, err)` ownership and partial-startup cleanup.

`app/worker.go` currently owns email/consumer, with Redis constructed in `api.go`; it has no PostgreSQL. Invitation re-reading/dispatch requires explicitly adding worker-owned database resources and readiness/cleanup, without allocating an API consumer or making redirector depend on identity/PostgreSQL. Extend role tests with actual required resources.

`service/services.go:16–22` offers registry shape only. `repository/repositories.go:7–11` is empty. `service/auth.go:15–20` calls `clerk.SetKey` despite its comment: replace that global with injected provider clients. `middleware/auth.go:70–79` copies organization roles/permissions and logs a subject: do not reuse those behaviors for Flux capabilities or observability. Research supplies JWT/JWKS/clock verification; Flux membership remains authoritative.

### 2. Product handlers, validation and safe errors

**Analog:** `apps/backend/internal/handler/base.go:7–10,77–109`.

```go
func Handle[Req validation.Validatable, Res any](
    _ Handler, handler HandlerFunc[Req, Res], status int, newRequest func() Req,
) echo.HandlerFunc {
    return func(c echo.Context) error {
        return handleRequest(c, newRequest(), func(c echo.Context, req Req) (interface{}, error) {
            return handler(c, req)
        }, JSONResponseHandler{status: status})
    }
}
```

Supply a fresh pointer factory per invocation. Copy isolation tests from `handler/base_test.go` (`TestWrappersIsolateConcurrentRequests`, `TestWrappersDoNotRetainOmittedFields`); never capture mutable bound payloads. Resolve actor/workspace from verified context, not body/query ownership fields. Authorization must also run inside service transaction boundaries.

`validation/utils.go:18–21,36–49` requires `Validate() error`, hides bind diagnostics and produces public field errors. Use `CustomValidationErrors` for approved field messages. Separate route/header/query/body validation deliberately; generic binding alone cannot enforce Origin, JSON-only payloads or version preconditions.

`middleware/global.go:91–119,147–155` recognizes wrapped `*errs.HTTPError` through `errors.As` and serializes only before response commitment. The canonical envelope fields are `code`, `message`, `errors`, `status`, `override`, `action` (`errs/http.go:30–36`). Return typed errors; add deliberate stable conflict/precondition/key-unavailable codes instead of deriving codes from personalized text. Register products under the currently empty `/api/v1` group (`router/router.go:59–60`), retaining correlation/recovery ordering.

### 3. Explicit SQL, migrations and real fixtures

**Analog:** `apps/backend/internal/testing/transaction.go:18–38,59–65`.

```go
tx, err := db.Pool.Begin(ctx)
if err != nil { return fmt.Errorf("failed to begin transaction: %w", err) }
defer func() { resultErr = errors.Join(resultErr, rollbackTransaction(ctx, tx)) }()
if err26 := fn(tx); err26 != nil { return err26 }
if err31 := tx.Commit(ctx); err31 != nil {
    return fmt.Errorf("failed to commit transaction: %w", err31)
}
```

This is a test helper, not an implemented product repository. Production uses research's `pgx.BeginFunc`/`BeginTxFunc` with required scope, parameterized statements and one documented workspace→membership→resource lock order. SQL, owner preservation, durable request ledger, signed tenant cursors, global normalized domain/key uniqueness and audit atomicity are new domain patterns.

`testing/container.go:57–65` provides `SetupTestDB(t) (*TestDB, func())`: it creates pinned real PostgreSQL and applies application migrations. Use it for product suites; `SetupTestPostgres` intentionally does not migrate. Redis has a separate real helper. Transaction cleanup detaches canceled request context and applies a five-second timeout; retain that property.

`database/migrator.go:25–26` embeds `migrations/*.sql`. Copy Tern's `---- create above / drop below ----` separator from `001_setup.sql:4`, preserving 001 and adding 002 onward. Extend `TestMigrationEmptyDatabaseAndBinary` with existing-version→new-schema assertions, not only empty database success. Fixtures must exercise FK denials, rollback, concurrent key/owner/version races and two-workspace response isolation.

### 4. Authored contracts and generated boundaries

**Analog:** `packages/zod/src/health.ts:1–2,13–24` and `packages/openapi/src/contracts/health.ts:1–14`.

```ts
import { extendApi } from "@anatine/zod-openapi";
import { z } from "zod";
// Existing schema naming/strictness:
export const ZHealthLiveResponse = extendApi(
  z.object({ status: z.literal("alive") }).strict(),
  { title: "transport.HealthLiveResponse",
    description: "Process liveness, independent of external dependencies.",
    example: { status: "alive" } },
);
```

Use named `Z` schemas, explicit wire spellings, `z.infer`, `.js` export suffixes and `@flux/zod` imports. Register schemas in Zod's index and routes in the contract router. Document all relevant statuses, bearer security, idempotency/version/request-ID/rate-limit headers. `getSecurityMetadata` exists; verify generated security, not merely an unused helper.

Do not copy `schemaWithPagination`: its page/total envelope contradicts approved cursor pagination. Define a strict cursor response and keep actual Go serialization aligned.

`scripts/generate.ts:16–21` enumerates both OpenAPI JSON copies, `transport/health.gen.go` and welcome HTML. Expand manifest/staging for product DTOs and invitation HTML. Existing `oapi-codegen.yaml:18–22` appends health-specific aliases; evolve it deliberately. Generated files are outputs, never hand-authored domain logic. Copy deterministic `serializeOpenAPI` (`gen.ts:29–31`) and isolated generation/drift checking (`scripts/generate.ts:149–198`). Extend `transport/contract_test.go` agreement assertions beyond its health-specific schema validator assumptions.

### 5. Invitation jobs, email and privacy

**Analog:** `job/email_tasks.go:76–96`, `job/handlers.go:110–117`.

```go
return asynq.NewTask(TaskWelcome, payload,
    asynq.MaxRetry(welcomeMaxRetries),
    asynq.Queue("default"),
    asynq.Timeout(welcomeTaskTimeout)), nil
// Safe stored retry errors:
func (e jobError) Error() string { return "job " + e.stage + ": " + observability.SafeError(e.cause) }
func (e jobError) Unwrap() error { return e.cause }
```

Reuse per-consumer dependency injection (`job.go:93`), bounded retry/timeout and cleaned correlation metadata. Invitation payloads carry workspace/invitation/delivery identities; worker verifies scoped durable state. Welcome payload's email/name is not an invitation delivery design. PostgreSQL intent, encrypted delivery material, provider result persistence and recovery polling have no analog. Add worker DB injection and tenant-forgery tests explicitly.

Invitation TSX copies the named/default export and `PreviewProps` convention from `welcome.tsx:16–20,66–70`, not its dashboard styles. Export/embedded-template registration and generator manifest must include its HTML.

## Shared Patterns

`observability/redaction.go:15–26` exposes only fixed timeout/cancellation/failure diagnostics. Preserve `%w`/`Unwrap` privately while avoiding `.Err(err)` with provider data. `redaction.go:76–111` uses a closed value allowlist; product route templates and invitation job type require approved bounded additions plus leakage tests. Never include workspace IDs, actor identity, destination/title/email/reason/token in operational labels. Audit actor/reason belongs in protected transactional records.

Root `package.json` now executes `scripts/check.ts`, not historical empty Turbo checks. Current Bun is 1.3.14 and Node 22.23.3; preserve strict workspace contracts. `check.ts:7,78–107` requires actual format/lint/typecheck/test/build scripts and nonempty tests for every discovered workspace. Add frontend package checks immediately; retain strictness while using Next-compatible TypeScript resolution rather than copying library emission settings blindly.

Register every new container-backed **top-level test name** in `integrationTests` (`check.ts:43–65`). Unregistered tests become units; filenames do not classify them. Discovery compares source names against compiled Go listings (`190–204`); execution uses race-enabled exact selections (`255–294`). Keep failure diagnostics sanitized. Browser `test:e2e` needs an explicit stage, browser provisioning and its own result handling: workspace `test` currently requires Bun's nonzero-pass output (`248–251,421–422`), so Playwright cannot replace it transparently. Extend runner tests, generated tests and scanners for Next build/cache outputs without suppressing secret/history checks.

## No Analog Found

Frontend is empty: no reusable dashboard, modal, workspace store, Clerk/Next integration or Playwright fixture exists. Use approved UI-SPEC and research pins (Next 16.4.0, Clerk 7.9.12, native React/CSS). Implement every scoped page/state; workspace switching cancels requests, clears tenant state and rejects late results, with authorization revalidation on focus/access denial. Native dialogs, semantic tables/mobile cards, explicit acceptance, conflict review and 320px/accessibility tests are new work. Go owns business decisions.

Destination URL policy/guarded transport, tenant capability/state machines, scoped CRUD/locking, durable idempotency, signed cursors and encrypted invitation outbox are also greenfield. Use RESEARCH examples and concrete tests; an empty registry is not precedent for their correctness. Creation never fetches destinations. No redirect/cache/analytics implementation or destination/title editing belongs here.

## Metadata

**Analog search scope:** backend composition, handlers/middleware/validation/errors, database/testing, jobs/observability; TypeScript schemas/contracts/email; generator/root gate/manifests. Five principal analog clusters above, with supporting cross-cutting references. No application source changed or tests claimed run; this is a documentation-only source inspection.
