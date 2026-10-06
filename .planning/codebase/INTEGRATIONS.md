# External Integrations

**Analysis Date:** 2026-10-05

## APIs & External Services

**Transactional email:**
- Resend — sends rendered welcome-email HTML in `apps/backend/internal/lib/email/client.go` and `apps/backend/internal/lib/email/emails.go`.
  - SDK/Client: `github.com/resend/resend-go/v2` v2.28.0 in `apps/backend/go.mod`.
  - Auth: `FLUX_INTEGRATION.RESEND_API_KEY`, derived from Koanf tags and loader behavior in `apps/backend/internal/config/config.go`.
  - Sender is hardcoded to Flux at Resend's onboarding domain in `apps/backend/internal/lib/email/client.go`; provider-specific sending configuration is not exposed.
  - HTML is parsed from `templates/emails/<template>.html` on each send. React Email exports source templates from `packages/emails/src/templates/` using `packages/emails/package.json`.
  - Welcome task handler is registered in `apps/backend/internal/lib/job/job.go`; `apps/backend/internal/lib/job/email_tasks.go` creates tasks, but no HTTP route currently enqueues them.

**Documentation CDN:**
- jsDelivr serves Scalar API Reference JavaScript in `apps/backend/static/openapi.html` using an unpinned `@scalar/api-reference` CDN URL.
  - SDK/Client: browser script; not a workspace dependency.
  - Auth: none.
  - `/docs` renders the page and `/static/openapi.json` supplies the contract via `apps/backend/internal/router/system.go`.

**Domain integrations:**
- No market data, broker, exchange, AI model, payment, or object-storage clients detected in `apps/backend/` or `packages/`. Treat integrations described in `spec.md` as requirements rather than wired services.

## Data Storage

**Databases:**
- PostgreSQL — native pgx v5.9.2 pooling in `apps/backend/internal/database/database.go`; database server version is not pinned for production.
  - Connection: `FLUX_DATABASE.HOST`, `FLUX_DATABASE.PORT`, `FLUX_DATABASE.USER`, `FLUX_DATABASE.PASSWORD`, `FLUX_DATABASE.NAME`, `FLUX_DATABASE.SSL_MODE` from `apps/backend/internal/config/config.go`.
  - Client: `pgxpool.Pool`, exposed through `server.Server.DB` in `apps/backend/internal/server/server.go`; no ORM detected.
  - Tern v2.4.1 applies embedded SQL and records version in `schema_version` via `apps/backend/internal/database/migrator.go`.
  - Startup migrates unless primary environment is `local` in `apps/backend/cmd/flux/main.go`; `apps/backend/internal/database/migrations/001_setup.sql` has no application DDL.
  - CLI migrations use `FLUX_DB_DSN` in `apps/backend/taskfile.yml`; this variable is separate from runtime database configuration.
  - Integration helper uses `postgres:15-alpine` in `apps/backend/internal/testing/container.go`; this image is not a production requirement.

**File Storage:**
- Local filesystem only: static documentation in `apps/backend/static/` and exported templates in `apps/backend/templates/emails/`, read through relative paths by `apps/backend/internal/router/system.go` and `apps/backend/internal/lib/email/client.go`.
- No upload persistence, S3-compatible client, or remote file storage implementation detected in `apps/backend/` or `packages/`.

**Caching:**
- Redis client is initialized and pinged by `apps/backend/internal/server/server.go`, but application cache reads/writes are not implemented.
- Asynq uses the same Redis address for queue persistence in `apps/backend/internal/lib/job/job.go`: concurrency 10; queue weights critical/default/low are 6/3/1.
- Welcome jobs use default queue, up to three retries, and 30-second timeout in `apps/backend/internal/lib/job/email_tasks.go`.
- Redis connection configuration currently exposes address only in `apps/backend/internal/config/config.go`; password, username, TLS, and database selection are not configured by this struct.
- HTTP rate limiting uses Echo's in-memory store rather than Redis in `apps/backend/internal/router/router.go`.

## Authentication & Identity

**Auth Provider:**
- Clerk — backend SDK v2.6.0 in `apps/backend/go.mod`.
  - `apps/backend/internal/service/auth.go` sets the process-wide SDK key from `Config.Auth.SecretKey`.
  - Auth variable: `FLUX_AUTH.SECRET_KEY` from `apps/backend/internal/config/config.go`.
  - `apps/backend/internal/middleware/auth.go` wraps Clerk header authorization, reads session claims, and stores user ID, organization role, and permissions in Echo context.
  - Middleware exists but is not applied to the current system routes in `apps/backend/internal/router/router.go` and `apps/backend/internal/router/system.go`; `/api/v1` group has no registered domain handlers.
  - OpenAPI declares bearer JWT and `x-service-token` schemes in `packages/openapi/src/index.ts`; declaration alone does not implement a service-token verifier.
  - No frontend sign-in implementation is present in `apps/frontend/`.

## Monitoring & Observability

**Error Tracking:**
- Optional New Relic APM — initialized only with a nonempty configured license in `apps/backend/internal/logger/logger.go`; failure to create the agent leaves instrumentation disabled.
- License variable: `FLUX_OBSERVABILITY.NEW_RELIC.LICENSE_KEY` derived from `apps/backend/internal/config/observability.go` and `apps/backend/internal/config/config.go`.
- HTTP transactions use `nrecho` in `apps/backend/internal/middleware/tracing.go`; middleware adds request ID, real IP, user agent, user ID when available, response status, and wrapped errors.
- PostgreSQL query tracing uses `nrpgx5` in `apps/backend/internal/database/database.go`; Redis hooks use `nrredis` in `apps/backend/internal/server/server.go`.
- Health-check custom events are recorded in `apps/backend/internal/handler/health.go`; rate-limit instrumentation is in `apps/backend/internal/middleware/rate_limit.go`.
- No Sentry client detected in `apps/backend/go.mod` or workspace package manifests.

**Logs:**
- Zerolog to stdout; production JSON when configured format is `json`, console output otherwise in `apps/backend/internal/logger/logger.go`.
- Logs include service/environment and request/trace context through `apps/backend/internal/middleware/context.go`, `apps/backend/internal/middleware/request_id.go`, and `apps/backend/internal/middleware/global.go`.
- `local` environment enables pgx query logging via pgx-zerolog in `apps/backend/internal/database/database.go`.
- `AppLogForwardingEnabled` is declared in `apps/backend/internal/config/observability.go`; the custom logger creation in `apps/backend/internal/logger/logger.go` does not consume that setting or attach a forwarding writer.
- `/status` pings PostgreSQL and Redis in `apps/backend/internal/handler/health.go`. Database failure determines HTTP 503; Redis failure is reported in checks without changing overall health status.

## CI/CD & Deployment

**Hosting:**
- Not detected: no hosting provider or deployment manifests in the scanned repository. Runtime is a plain Go `net/http.Server` started by `apps/backend/cmd/flux/main.go`.
- Build-time generation and local orchestration use `package.json`, `turbo.json`, and `apps/backend/taskfile.yml`; backend HTTP server uses configurable port and timeouts in `apps/backend/internal/server/server.go`.

**CI Pipeline:**
- Not detected: no workflow directory or CI configuration found. Root scripts in `package.json` expose build/lint/typecheck/format commands but are not an automated pipeline.

## Environment Configuration

**Required env vars:**
- Names below are inferred directly from tags and the env loader in `apps/backend/internal/config/config.go`: prefix `FLUX_`, delimiter `.`, no replacement of ordinary underscores with dots.
- `FLUX_PRIMARY.ENV`.
- `FLUX_SERVER.PORT`, `FLUX_SERVER.READ_TIMEOUT`, `FLUX_SERVER.WRITE_TIMEOUT`, `FLUX_SERVER.IDLE_TIMEOUT`, `FLUX_SERVER.CORS_ALLOWED_ORIGINS`.
- `FLUX_DATABASE.HOST`, `FLUX_DATABASE.PORT`, `FLUX_DATABASE.USER`, `FLUX_DATABASE.NAME`, `FLUX_DATABASE.SSL_MODE`, `FLUX_DATABASE.MAX_OPEN_CONNS`, `FLUX_DATABASE.MAX_IDLE_CONNS`, `FLUX_DATABASE.CONN_MAX_LIFETIME`, `FLUX_DATABASE.CONN_MAX_IDLE_TIME`.
- `FLUX_DATABASE.PASSWORD` is supported but lacks a `required` validation tag.
- `FLUX_REDIS.ADDRESS`, `FLUX_AUTH.SECRET_KEY`, `FLUX_INTEGRATION.RESEND_API_KEY`.
- Optional observability subtree: `FLUX_OBSERVABILITY.LOGGING.LEVEL`, `FLUX_OBSERVABILITY.LOGGING.FORMAT`, `FLUX_OBSERVABILITY.NEW_RELIC.LICENSE_KEY`, and remaining tags in `apps/backend/internal/config/observability.go`. Defaults apply when the entire observability pointer is absent; a partial subtree is not merged with defaults by `apps/backend/internal/config/config.go`.
- Database pool settings are validated in configuration but not assigned to pgx pool options in `apps/backend/internal/database/database.go`.
- `FLUX_DB_DSN` is needed only by the explicit Tern CLI task in `apps/backend/taskfile.yml`.

**Secrets location:**
- `apps/backend/.env` and `apps/backend/.env.sample` exist; contents are not read. Godotenv autoload is imported in `apps/backend/internal/config/config.go`; `.gitignore` excludes `.env`.
- Deployment secret manager configuration: not detected; `apps/backend/internal/config/config.go` expects process environment inputs.
- Use variable names only in maps and never copy configuration values from secret files.

## Webhooks & Callbacks

**Incoming:**
- None implemented: `apps/backend/internal/router/system.go` registers only GET `/status`, `/docs`, and static assets; `apps/backend/internal/router/router.go` creates an empty `/api/v1` group.
- No Clerk lifecycle webhook, Resend delivery webhook, or external ingestion endpoint detected in `apps/backend/internal/handler/`.

**Outgoing:**
- Resend SDK sends emails from `apps/backend/internal/lib/email/client.go`; this is a provider API call, not an application-defined webhook callback.
- New Relic agent emits telemetry from `apps/backend/internal/logger/logger.go` when enabled.
- No configurable outbound webhook targets or callback dispatcher detected in `apps/backend/internal/`.

---

*Integration audit: 2026-10-05*
