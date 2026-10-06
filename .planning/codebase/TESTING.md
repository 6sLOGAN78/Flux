# Testing Patterns

**Analysis Date:** 2026-10-05

## Test Framework

**Runner:**
- Go's standard `testing` package; the module declares Go `1.25.5` in `apps/backend/go.mod`.
- No dedicated runner configuration or actual `_test.go` files are detected under `apps/backend/`. Test support exists in `apps/backend/internal/testing/` as ordinary Go implementation files.
- No TypeScript unit runner, browser runner, or package test scripts are detected in `package.json`, `packages/zod/package.json`, `packages/openapi/package.json`, or `packages/emails/package.json`.
- No test suite execution or compilation is established by this documentation scan.

**Assertion Library:**
- Testify `v1.11.1` is declared in `apps/backend/go.mod`. Helpers use `require` for fatal setup/type failures and `assert` for checks that can continue: `apps/backend/internal/testing/container.go`, `apps/backend/internal/testing/assertions.go`.
- Testcontainers Go `v0.42.0` is declared in `apps/backend/go.mod`; the database helper launches `postgres:15-alpine` in `apps/backend/internal/testing/container.go`.

**Run Commands:**
The following standard Go commands apply to the module at `apps/backend/go.mod`; they are not repository-defined test tasks and currently have no detected test cases to execute.

```bash
cd apps/backend
go test ./...                         # Discover/run Go tests
go test ./... -race                   # Race detection
go test ./... -coverprofile=coverage.out # Coverage output
go tool cover -html=coverage.out      # View an existing coverage profile
```

- Watch mode: not configured in `apps/backend/taskfile.yml` or root `package.json`.
- `apps/backend/taskfile.yml` contains run/migration/tidy tasks, with no test, coverage, or lint task.
- Root `package.json` delegates lint/typecheck/format scripts to Turbo, but the TypeScript package manifests do not define matching checks. Treat build scripts in `packages/zod/package.json` and `packages/openapi/package.json` as compilation workflows, not test suites.

## Test File Organization

**Location:**
- No runnable test suites are detected. Shared backend testing infrastructure lives in `apps/backend/internal/testing/`.
- Place Go tests near the package being exercised using standard `_test.go` naming; this guidance follows the `_test.go` exclusions and `testpackage` check in `apps/backend/.golangci.yml`, not an existing test suite example.
- Alias imports of `github.com/6sLOGAN78/flux/internal/testing` when also importing Go's standard `testing` package, because the helper package declares `package testing` in all five files.

**Naming:**
- Test file naming is not demonstrated by actual test files. `apps/backend/.golangci.yml` expects `_test.go` for its test-specific exclusions.
- Helper names describe behavior: `SetupTestDB`, `SetupTest`, `CreateTestServer`, `MustMarshalJSON`, `AssertValidUUID`, `WithRollbackTransaction` in `apps/backend/internal/testing/`.

**Structure:**
```text
apps/backend/internal/testing/
├── container.go     # PostgreSQL container and migrated pool setup
├── server.go        # Minimal server assembly for tests
├── helpers.go       # Combined setup, JSON/root/pointer helpers
├── assertions.go    # Timestamp, UUID, object, and string assertions
└── transaction.go   # Commit-on-success and rollback-only callbacks
```

## Test Structure

**Suite Organization:**
- No table-driven suites, `TestMain`, subtests, benchmarks, or Testify suites are detected. Do not infer suite conventions from `readme.md` or `spec.md`.
- The actual reusable setup pattern from `apps/backend/internal/testing/helpers.go` is:

```go
func SetupTest(t *testing.T) (*TestDB, *server.Server, func()) {
    t.Helper()
    // Logger initialization omitted here.
    testDB, dbCleanup := SetupTestDB(t)
    testServer := CreateTestServer(&logger, testDB)
    cleanup := func() {
        if testDB.Pool != nil {
            testDB.Pool.Close()
        }
        dbCleanup()
    }
    return testDB, testServer, cleanup
}
```

**Patterns:**
- Mark helpers receiving `*testing.T` with `t.Helper()` to attribute failures to callers: `apps/backend/internal/testing/helpers.go`, `apps/backend/internal/testing/assertions.go`.
- `SetupTestDB` registers container termination through `t.Cleanup`; its returned cleanup closure closes the pool. Callers must arrange pool cleanup as well: `apps/backend/internal/testing/container.go`.
- `SetupTest` returns a database, minimal server, and cleanup closure. The closure closes the pool directly and also invokes `dbCleanup`, which closes it again: `apps/backend/internal/testing/helpers.go`.
- Prefer fatal assertions for prerequisites and nonfatal assertions for independent checks, matching `apps/backend/internal/testing/assertions.go`.
- `_test.go` exclusions reduce selected lint checks; configured `testifylint`, `testpackage`, and `tparallel` rules remain declared in `apps/backend/.golangci.yml`. No passing baseline is verified.

## Mocking

**Framework:** No implemented mock framework or mock types are detected. Testify is available through `apps/backend/go.mod`, but its mock package is not imported by application or helper files.

**Patterns:**
- No actual mocking example exists. The implemented seam uses direct server assembly instead of invoking production startup (`apps/backend/internal/testing/server.go`):

```go
testServer := &server.Server{
    Logger: logger,
    DB: &database.Database{
        Pool: db.Pool,
    },
    Config: db.Config,
}
```

**What to Mock:**
- No repository policy is implemented. `CreateTestServer` omits Redis, job service, logger service, and HTTP server startup, so tests depending on those fields need explicit setup; this helper is not a full external-service fixture (`apps/backend/internal/testing/server.go`, `apps/backend/internal/server/server.go`).
- Production Clerk initialization sets a process-level SDK key in `apps/backend/internal/service/auth.go`, and email jobs use a package-level client in `apps/backend/internal/lib/job/handlers.go`; no test isolation/mock seam is implemented for either.

**What NOT to Mock:**
- The supplied database setup uses a real PostgreSQL container and real pgx pool, then applies real migrations: `apps/backend/internal/testing/container.go`, `apps/backend/internal/database/migrator.go`.
- No stronger general no-mocking policy is demonstrated by test implementation; statements about mocking in `readme.md` describe tools rather than runnable coverage.

## Fixtures and Factories

**Test Data:**
- `SetupTestDB` generates a unique database name from a UUID, builds a test-mode configuration, resolves Docker host/port, retries database connection five times with two-second sleeps, verifies `Ping`, and applies migrations: `apps/backend/internal/testing/container.go`.
- The current migration `apps/backend/internal/database/migrations/001_setup.sql` contains only up/down placeholder comments, so applying migrations does not create domain tables.
- `CreateTestServer` supplies default observability configuration when absent and disables New Relic forwarding/tracing and health-check configuration: `apps/backend/internal/testing/server.go`.
- Actual reusable data helpers in `apps/backend/internal/testing/helpers.go`:

```go
func MustMarshalJSON(t *testing.T, v interface{}) []byte {
    t.Helper()
    jsonBytes, err := json.Marshal(v)
    require.NoError(t, err, "failed to marshal to JSON")
    return jsonBytes
}

func Ptr[T any](v T) *T {
    return &v
}
```

**Location:**
- No fixture directories, seed builders, snapshots, or persisted test data are detected. Helpers are in `apps/backend/internal/testing/`.
- `ProjectRoot` in `apps/backend/internal/testing/helpers.go` walks upward until it finds `go.mod`, returning `apps/backend` for backend package tests, rather than the monorepo root.
- Email `PreviewProps` in `packages/emails/src/templates/welcome.tsx` supplies a first-name preview value; it is a preview fixture, not an automated test.

## Coverage

**Requirements:** None enforced. No coverage threshold, CI workflow, or coverage-report configuration is detected alongside `apps/backend/go.mod`, `apps/backend/taskfile.yml`, or root `package.json`.

**View Coverage:**
Standard Go tooling can display a generated profile; no profile or repository wrapper is detected.

```bash
cd apps/backend
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
go tool cover -html=coverage.out
```

## Test Types

**Unit Tests:**
- Not implemented as detected `_test.go` suites. Pure mapping/error helpers in `apps/backend/internal/sqlerr/err.go` and `apps/backend/internal/sqlerr/handler.go`, validation in `apps/backend/internal/validation/utils.go`, and payload construction in `apps/backend/internal/lib/job/email_tasks.go` have no detected behavioral tests.
- TypeScript schemas and OpenAPI metadata/generation in `packages/zod/src/health.ts`, `packages/openapi/src/utils.ts`, and `packages/openapi/src/gen.ts` have no detected test files or runner configuration.

**Integration Tests:**
- Infrastructure exists for container-based database tests, but no tests invoke it: `apps/backend/internal/testing/container.go`.
- A Docker-compatible runtime and image availability are prerequisites for that helper. It exposes PostgreSQL port dynamically and waits up to 30 seconds for a readiness log: `apps/backend/internal/testing/container.go`.
- Minimal server helpers do not register routes or global middleware and do not initialize job/Redis services; endpoint tests need explicit router/dependency assembly: `apps/backend/internal/testing/server.go`, `apps/backend/internal/router/router.go`.

**E2E Tests:**
- Not used. No Playwright/Cypress configuration, browser suites, or frontend app implementation is detected; current TSX source is the email template in `packages/emails/src/templates/welcome.tsx`.

## Common Patterns

**Async Testing:**
- No async test suite example is detected. Transaction callbacks are the implemented database operation pattern in `apps/backend/internal/testing/transaction.go`:

```go
type TxFn func(tx pgx.Tx) error

func WithRollbackTransaction(ctx context.Context, db *TestDB, fn TxFn) error {
    tx, err := db.Pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)
    return fn(tx)
}
```

- Use `WithRollbackTransaction` for rollback isolation. `WithTransaction` commits on a successful callback despite its comment saying it rolls back afterward: `apps/backend/internal/testing/transaction.go`.
- No `t.Parallel`, goroutine synchronization, fake-clock, or timer-testing convention is demonstrated. Container startup uses real sleeps and contexts in `apps/backend/internal/testing/container.go`.

**Error Testing:**
- No tests assert production error behavior. Actual helper assertions follow this pattern in `apps/backend/internal/testing/assertions.go`:

```go
createdAt, ok := createdField.Interface().(time.Time)
require.True(t, ok, "CreatedAt is not a time.Time")
assert.False(t, createdAt.IsZero(), "CreatedAt should not be zero")
```

- `AssertTimestampsValid` skips fields that are absent; it does not require timestamp fields to exist. `AssertEqualExceptTime` expects structs or struct pointers and ignores fields whose declared types are `time.Time` or `*time.Time`: `apps/backend/internal/testing/assertions.go`.
- `AssertValidUUID` only checks against `uuid.Nil`; `AssertStringContains` checks every supplied substring: `apps/backend/internal/testing/assertions.go`.
- Error contract status/field/action mapping, malformed bind errors, unhealthy Redis response behavior, and transaction cleanup have no detected regression cases: `apps/backend/internal/middleware/global.go`, `apps/backend/internal/validation/utils.go`, `apps/backend/internal/handler/health.go`, `apps/backend/internal/testing/transaction.go`.

---

*Testing analysis: 2026-10-05*
