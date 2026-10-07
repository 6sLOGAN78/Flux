# Development

Flux's foundation has four Go process roles and three implemented TypeScript
workspaces: `@flux/zod`, `@flux/openapi`, and `@flux/emails`. Product link routing,
tenant entities, and a dashboard belong to later phases.

## Acceptance contract

Use `.tool-versions` for Go 1.26.8, Node 22.23.3, and Bun 1.3.14.
`tools.lock.json` is the source for standalone quality tools, the Go generator,
and collector version. `compose.yaml` and the Testcontainers helpers use the
same immutable PostgreSQL 17.11 and Redis 8.10.2 images. Bun owns the root
`bun.lock`; the embedded Tern library in `apps/backend/go.mod` owns migrations.

From a full-history checkout with those runtimes and a running Docker daemon:

```sh
bun install --frozen-lockfile
bun run tools:install
bun scripts/install-tools.ts --verify-config
(cd apps/backend && go mod download && go mod verify)
CI=true bun run check
CI=true bun run scan
```

Install the exact runtimes using your existing version manager's `.tool-versions`
support, or the official [Go](https://go.dev/dl/), [Node](https://nodejs.org/en/download),
and [Bun](https://github.com/oven-sh/bun/releases) distributions. Confirm
`go version` from `apps/backend`, `node --version`, and `bun --version` match the
manifest. A supported older Go launcher may download the selected toolchain;
for the same runtime throughout root scripts set `GOTOOLCHAIN=go1.26.8`.
Go module commands use `GOFLAGS=-mod=readonly` and `GOWORK=off` in CI.

The root frozen install also applies `patches/ts-deepmerge@8.0.0.patch` through
Bun's checked `patchedDependencies`. Keep the patch directory with the authored
files when testing a clean copy. It supplies the existing secure merge function
as the CommonJS default required by the OpenAPI adapters; it does not change
the implementation or relax unsafe-key filtering. An old `node_modules` tree
does not prove that the frozen lock and patch install correctly.

Docker must be reachable by the user running tests. Prepare the exact images:

```sh
docker info > /dev/null
docker compose --profile observability config --quiet
docker compose --profile observability pull
```

The tests create their own disposable pinned containers and dynamic ports.
They do not require a manually started Compose stack or Clerk, Resend, or a
monitoring account. The optional local stack is for operating roles yourself:

```sh
docker compose up -d --wait postgres redis
```

Compose binds host ports to loopback. Its credentials are disposable local
defaults; `FLUX_LOCAL_POSTGRES_PORT` and `FLUX_LOCAL_REDIS_PORT` change host ports.
Set the matching application database/Redis settings when overriding them.

The full gate must execute all backend checks and every implemented workspace,
race-enabled unit and container integration tests, explicit migration tests,
all role/workspace builds, isolated regeneration, and dependency plus secret
scans. No coverage percentage replaces the required behavior assertions.
Generation checks execute directly without accepting a Turbo cache hit.

Run the existing executable gate regressions before accepting workflow changes:

```sh
bun test scripts/check.test.ts scripts/scan.test.ts scripts/generated.test.ts
```

These test omitted scripts and compiled test discovery, nonzero stages, skipped
tests, generated drift and failure atomicity, scanner failures, redaction,
temporary secret files and deleted full-history secrets. The package scanner
also runs a real temporary vulnerable OpenPGP import that must fail. Formatting,
missing workspace scripts, generated drift, and scanner findings must each
fail their actual gate in a disposable copy; never commit secret fixtures.

CI must run on pushes and pull requests with `contents: read`, full-history
checkout and no persisted checkout credentials. Actions use verified official
commit SHAs, runtime versions come from the shared manifest, tools verify both
download and executable hashes, and jobs and subprocesses have deadlines.
Obsolete runs are canceled. Untrusted code receives no production secrets.
Scanner reports stay in bounded private memory; only sanitized stage results
and advisory/location fields enter logs. Raw reports and credentials must not
be uploaded or cached. A local check does not prove a hosted CI execution.

## Full and fast feedback

`bun run check` executes the full gate. `bun run check:fast` retains formatting,
lint, type checking, race-enabled Go units, every workspace/script test, all
builds and uncached generation; it omits only the container integration,
migration and scan stage groups. CI always uses the full gate. Some fast
regressions still invoke real generators and scanners to prove failure behavior.

Individual root commands are `format:check`, `lint`, `typecheck`, `test:unit`,
`test:integration`, `build`, `generate:check`, `migrate:check`,
`scan:dependencies`, and `scan:secrets`. The root migration check creates an empty
test database and proves repeated migration and one-shot command behavior.
`bun run generate` publishes regenerated canonical/served OpenAPI, generated
Go health transport types and exported email HTML after authored changes.
`bun run generate:check` compares all four categories by bytes without modifying
the checked files. Commit authored sources and corresponding generated changes
together after review.

If a check fails, its stage identifier names the failing command. Reproduce that
command locally with disposable test configuration; keep raw provider/scanner
output private. The root runner bounds subprocess output, rejects skipped or
missing expected Go tests, and never publishes captured private diagnostics.

## Role configuration and explicit migrations

The application retains `FLUX_` names with dot nesting. Use `env` with quoted
assignments: shells cannot `export FLUX_DATABASE.HOST=...`. Underscores within a
field remain underscores. `.env` autoload still applies to the current working
directory; use an explicit environment when launching a binary elsewhere.

All roles require `FLUX_PRIMARY.ENV` (`local`, `development`, `test`, `staging`,
or `production`) and validate shared logging/optional telemetry settings.

| Role | Required configuration/resources | Listener and readiness |
| --- | --- | --- |
| API | PostgreSQL connection/pool settings; Redis address only when `FLUX_API.PRODUCER_ENABLED=true` | `FLUX_API.LISTEN_ADDRESS`, default `:8080`; database, plus Redis for the optional producer |
| Redirector | Process/HTTP settings only at this foundation stage | `FLUX_REDIRECTOR.LISTEN_ADDRESS`, default `:8081`; no external dependency probes |
| Worker | `FLUX_REDIS.ADDRESS`, `FLUX_INTEGRATION.RESEND_API_KEY`, email adapter and queue consumer | `FLUX_WORKER.LISTEN_ADDRESS`, default `127.0.0.1:8082`; Redis and local email configuration/embedded-template validation |
| Migrator | PostgreSQL connection settings; pool defaults are supplied | One-shot command exit status; no HTTP, Redis, email or auth requirement |

API startup never applies migrations or starts consumers. Producer mode creates
only an enqueue client. Worker readiness does not send email or probe Resend.
No current foundation route uses Clerk, so an auth secret is not required for
these roles. The redirector shell exposes health only; product routing follows
in its assigned later phase.

From `apps/backend`, define a disposable local PostgreSQL environment:

```sh
flux_postgres() {
  env 'FLUX_PRIMARY.ENV=local' \
    'FLUX_DATABASE.HOST=127.0.0.1' 'FLUX_DATABASE.PORT=5432' \
    'FLUX_DATABASE.USER=flux_local' 'FLUX_DATABASE.PASSWORD=flux_local_only' \
    'FLUX_DATABASE.NAME=flux' 'FLUX_DATABASE.SSL_MODE=disable' \
    'FLUX_DATABASE.MAX_OPEN_CONNS=2' 'FLUX_DATABASE.MAX_IDLE_CONNS=1' \
    'FLUX_DATABASE.CONN_MAX_LIFETIME=60' 'FLUX_DATABASE.CONN_MAX_IDLE_TIME=30' \
    "$@"
}
flux_postgres go run ./cmd/migrator
flux_postgres go run ./cmd/migrator
flux_postgres env 'FLUX_API.LISTEN_ADDRESS=127.0.0.1:8080' go run ./cmd/api
```

The first migrator applies pending embedded Tern SQL; the second is a successful
no-op. `task migrate` invokes the same binary, and `task migrations:up` remains
its compatibility alias. No separate `FLUX_DB_DSN` authority is needed.
`task migrations:new name=...` is an optional authoring helper requiring the
Tern CLI; tests and runtime migration use the locked Go library.

In separate terminals, from `apps/backend`:

```sh
env 'FLUX_PRIMARY.ENV=local' \
  'FLUX_REDIRECTOR.LISTEN_ADDRESS=127.0.0.1:8081' go run ./cmd/redirector
env 'FLUX_PRIMARY.ENV=local' 'FLUX_REDIS.ADDRESS=127.0.0.1:6379' \
  "FLUX_INTEGRATION.RESEND_API_KEY=$FLUX_LOCAL_EMAIL_KEY" \
  'FLUX_WORKER.LISTEN_ADDRESS=127.0.0.1:8082' go run ./cmd/worker
```

Set `FLUX_LOCAL_EMAIL_KEY` privately before running a worker that sends mail.
The automated tests inject local delivery/transport dependencies and require no
external email account. `go run ./cmd/flux` and `task run` remain API-only
compatibility entrypoints. `task run:api`, `run:redirector`, `run:worker`, and
`run:migrator` select the explicit roles.

Legacy `FLUX_SERVER.PORT`, HTTP timeout/CORS fields and existing database,
Redis, integration and auth names remain accepted. A role's listen address
takes precedence over the legacy port. Configure
`FLUX_<ROLE>.DRAIN_TIMEOUT` (default `30s`) and
`FLUX_<ROLE>.READINESS_TIMEOUT` (default `5s`) using Go duration strings.
Invalid settings owned by a different role do not make an independent role fail.
Legacy New Relic settings remain compatible but initialize no vendor SDK;
OTLP is optional and never a readiness dependency. See [observability](observability.md)
for the local collector, redaction, bounds and export configuration.

## Health, signals and embedded assets

Every long-running role exposes `GET /live` (process-only 200) and `GET /ready`
(200 when its required components are ready; sanitized 503 otherwise). Probe
the actual listener address, for example:

```sh
curl --fail http://127.0.0.1:8080/live
curl --fail http://127.0.0.1:8080/ready
curl --fail http://127.0.0.1:8080/static/openapi.json
```

The API serves `/docs` and the canonical embedded OpenAPI. Worker management
and redirector listeners do not serve API documentation. Dependency/provider
errors and secrets never appear in public health payloads.

Send SIGTERM or SIGINT to the built role process (not a `go run` launcher) to
stop it. Readiness first drops, intake stops, active HTTP/jobs drain within one
shared deadline, and resources close in reverse dependency order with telemetry
flushed last. The worker management endpoint can remain live but unready during
drain. Failed startup releases previously created resources; cleanup failures
or deadline exhaustion produce a nonzero process exit.

`bun run build` builds every Go command into `tmp/check-bin/` and all implemented
TypeScript workspaces. `task build` in `apps/backend` alternatively places the
four role binaries in `apps/backend/build/bin/`. Docs and email assets are
embedded, so no static/template directory needs to be copied beside a binary.
To verify working-directory independence after the root build:

```sh
FLUX_REPO_ROOT="$PWD"
(cd /tmp && env 'FLUX_PRIMARY.ENV=local' \
  'FLUX_REDIRECTOR.LISTEN_ADDRESS=127.0.0.1:8081' \
  "$FLUX_REPO_ROOT/tmp/check-bin/redirector")
```

An API or worker launched the same way still needs its role's explicit
environment. The integration suite runs the actual commands from unrelated
directories, checks API docs/template availability, tests SIGTERM while work is
active, and proves partial startup cleanup and real dependency isolation.

## CI provenance and retained diagnostics

[The workflow](../.github/workflows/ci.yml) runs on every push and pull request
on a fresh hosted Ubuntu runner with a read-only token and obsolete-run
cancellation. It reads runtime versions from `.tool-versions`, verifies installed
versions against package/Go/scanner manifests, downloads tools with checked
hashes, verifies Go modules, pulls immutable Compose images, then executes the
full root check and a separately visible sanitized scan. No production secrets,
credential caches, dependency caches or artifact uploads are configured.
If caches are introduced later, key them on OS/architecture, exact runtime/tool
versions, `bun.lock`, checked patches, `go.sum` and `tools.lock.json`; keep all
credential/config/report files outside caches and repeat integrity verification.

Official action tags and their commit objects were verified through upstream
Git refs and the GitHub API; inputs were checked against the exact `action.yml`:

| Action | Release | Verified commit |
| --- | --- | --- |
| [actions/checkout](https://github.com/actions/checkout/tree/3d3c42e5aac5ba805825da76410c181273ba90b1) | v7.0.1 | `3d3c42e5aac5ba805825da76410c181273ba90b1` |
| [actions/setup-go](https://github.com/actions/setup-go/tree/b7ad1dad31e06c5925ef5d2fc7ad053ef454303e) | v7.0.0 | `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` |
| [actions/setup-node](https://github.com/actions/setup-node/tree/820762786026740c76f36085b0efc47a31fe5020) | v7.0.0 | `820762786026740c76f36085b0efc47a31fe5020` |
| [oven-sh/setup-bun](https://github.com/oven-sh/setup-bun/tree/0c5077e51419868618aeaa5fe8019c62421857d6) | v2.2.0 | `0c5077e51419868618aeaa5fe8019c62421857d6` |

The actions' Node 24 implementation runtime is separate from the project's
pinned Node 22 toolchain. Use current hosted runners satisfying upstream runner
requirements. Full history and `persist-credentials: false` follow the official
[checkout guidance](https://github.com/actions/checkout/tree/3d3c42e5aac5ba805825da76410c181273ba90b1)
and GitHub's [secure-use guidance](https://docs.github.com/en/actions/reference/security/secure-use).

The scanner fails on every vulnerable imported Go package across all backend
roles and tests, Bun findings, secrets and scanner execution/report errors.
Unused module inventory advisories remain visibly labeled; `GO-2026-5932` is
currently inventory-only, and a real OpenPGP import mutation proves it becomes
a blocking finding when imported. There are no advisory exceptions. Logs retain
only bounded sanitized stage/advisory/location diagnostics; raw findings remain
in private memory and are never uploaded. Full Git history is mandatory in CI;
local no-Git copies explicitly report worktree-only scanning.

Local manifest, shell, clean-install and behavior checks are evidence for these
commands. Hosted workflow execution must be observed separately before claiming
a remote CI pass.
