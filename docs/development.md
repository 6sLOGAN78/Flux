# Development

Flux has four Go process roles and four implemented TypeScript workspaces:
`@flux/zod`, `@flux/openapi`, `@flux/emails`, and the Next.js `@flux/frontend`.
The control plane manages tenant-scoped teams and links. Actual redirects and
analytics belong to later phases; live provider acceptance remains a separate gate.

## Acceptance contract

The API requires `FLUX_LINKS.MANAGED_HOST` to be an operator-owned DNS hostname
(for example your organization's short-link hostname); there is no default.
Set the literal dot-nested key in the backend dotenv file or deployment
environment. `apps/backend/.env.sample` documents the two link-policy
keys. Do not deploy with the test fixture hostname `go.flux.test`.
`FLUX_LINKS.BLOCKED_HOSTS` is an optional comma-separated list of DNS hostnames;
entries are trimmed, IDNA-normalized and lowercased once during API startup.
Empty entries, URLs, ports, IP literals and local/internal names fail closed.
Redirector, worker and migrator configuration does not require these API keys.

The authorized workspace response supplies the immutable managed hostname.
Creation rejects local/private/metadata, special or reserved IP destinations,
the managed hostname and its subdomains, and configured blocked names and their
subdomains. Host normalization preserves encoded path/query meaning. Validation
does not resolve DNS, fetch the destination, follow redirects or render previews.
DNS answers are not claimed public merely because a syntactic hostname passes;
any future destination processing must use the separately guarded transport.
Link management does not establish redirect or analytics availability.

Use `.tool-versions` for Go 1.26.9, Node 22.23.3, and Bun 1.3.14.
`tools.lock.json` is the source for standalone quality tools, the Go generator,
and collector version, including the required Task 3.54.0 runtime for backend
task parity tests. `bun run tools:install` provisions it without a host Task
installation; the root check prepends `tmp/tools` for all test subprocesses.
For direct backend tests, use `PATH="$PWD/tmp/tools:$PATH"` from the repository
root before changing directories. `compose.yaml` and the Testcontainers helpers use the
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
for the same runtime throughout root scripts set `GOTOOLCHAIN=go1.26.9`.
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
| Worker | PostgreSQL connection/pool settings, `FLUX_REDIS.ADDRESS`, `FLUX_INTEGRATION.RESEND_API_KEY`, invitation encryption/sender/origin settings, email adapter and queue consumer | `FLUX_WORKER.LISTEN_ADDRESS`, default `127.0.0.1:8082`; actual PostgreSQL and Redis probes, local email/key validation and both embedded templates |
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
flux_postgres env 'FLUX_REDIS.ADDRESS=127.0.0.1:6379' \
  "FLUX_INTEGRATION.RESEND_API_KEY=$FLUX_LOCAL_EMAIL_KEY" \
  'FLUX_WORKER.LISTEN_ADDRESS=127.0.0.1:8082' go run ./cmd/worker
```

Set `FLUX_LOCAL_EMAIL_KEY` privately before running a worker that sends mail.
Provision the invitation settings below privately for both API and worker before
running these commands; the worker has required PostgreSQL since plan 02-21.
Run the explicit migrator first. Worker startup never applies migrations.
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

### Offline production identity fixture

`bun run test:e2e -- --project=local --grep identity` starts the registered
`TestBrowserProductFixture` test binary with a bounded readiness/teardown
protocol. It uses digest-pinned migrated PostgreSQL, RSA-signed sessions and
injected Clerk SDK JWKS/session/profile clients through the production Go router.
The legacy `TestBearerBrowserFixture -browser-fixture` launch protocol remains
compatible. Neither fixture ships in production binaries or accepts arbitrary
runtime provider URLs. These tests establish internal identity behavior; live
Clerk verification, recovery and social factors remain the final provider gate.

Playwright output lives in a private OS temporary directory and is removed on
normal process exit. Traces, screenshots and video are disabled. Root checks
capture reports privately and require positive completed JSON results, zero
skips/retries/flaky cases, and matching aggregate counts. Before execution the
root collects the same native selection with `--list`; unique spec/project IDs
must match the executed report exactly. A self-consistent report that omits or
substitutes a collected test fails. Discovery command/report failures also fail.
Go execution rejects skipped or failed subtests even if their parent passes.
The delivered frontend is mandatory in workspace discovery. Never save signed
fixture responses or browser reports in the repository, including ignored `tmp`.

Generation stages the complete authored schema/contract package sources, including
identity, without reading stale `dist` authority. All five manifest artifacts
must exist and be nonempty before any publication. Health aliases and generated
bytes are preserved; identity consumers use the generated transport names.
The manifest includes both OpenAPI copies, Go transport, welcome HTML and
invitation HTML. Root stages invoke checks and builds directly, without Turbo
test/build cache acceptance. Git-ignored cache/test paths remain subject to
worktree and full-history secret scanning. After browser execution, rebuild the
frontend and use the validated `cleanFrontendBuild` cleanup before a separate
final scan; never add a secret-scanner exclusion for generated deployment keys.

The API also requires private `FLUX_LINKS.CURSOR_KEY`: standard Base64 of exactly
32 random bytes (`openssl rand -base64 32`). Provision the same secret to all API
replicas through the deployment secret store; never expose it to the frontend.
Redirector, worker and migrator do not require this key. Rotation invalidates
existing link cursors; users can return to the first page. No fallback key exists.

Invitation encryption and delivery configuration belongs to both API and worker:
`FLUX_INVITATIONS.ACTIVE_KEY_ID`, `FLUX_INVITATIONS.ENCRYPTION_KEYS`,
`FLUX_INVITATIONS.SENDER`, and `FLUX_INVITATIONS.PUBLIC_ORIGIN`.
The encryption setting is JSON text mapping safe key IDs (1–64 letters, digits,
hyphens or underscores) to canonical standard-Base64 strings encoding exactly
32 random bytes: `{"current":"<base64-32-byte-key>"}` is a shape illustration,
not a deployable key. The ring holds at most 32 keys and 16 KiB of JSON. Provision
it privately from an external secret store; never paste real keys into logs,
commits, browser configuration or support messages. The active ID must exist.
There is no plaintext fallback. The sender is an email address and public origin
is an HTTPS origin without credentials, path, query or fragment.

AES-256-GCM uses Go's standard `NewGCMWithRandomNonce`, authenticating workspace,
invitation, delivery and key IDs. It generates a 96-bit random nonce and prepends
it to ciphertext. Count encryptions across every replica and retry/resend using
the same key; rotate before the per-key limit of 2^32 messages. Select a new
active ID while retaining previous keys until pending intents drain or are
safely re-encrypted. Never drop a key that live ciphertext still references.
Missing or malformed keys and authentication failures must fail closed.

The API atomically records invitations, encrypted delivery intents, immutable
audits and canonical-payload replay records retained for at least 24 hours.
Current SQL authority is checked before replay. Tokens are stored only as a
digest and inside the encrypted intent, never in Team responses or logs.
Invitation expiry is seven days; duplicate elapsed pending invitations expire
transactionally before replacement. Team displays scoped durable state:
Queued, Delivered, Delivery failed, Accepted, Expired or Revoked, with full local
expiry timestamps. The API never consumes invitation intents or calls the email
provider. The independent worker polls durable PostgreSQL intents, publishes
opaque references to Redis and records delivery only after acknowledgement.
Existing `FLUX_INTEGRATION.RESEND_API_KEY` remains
unchanged and is owned by the worker.

Worker readiness renders welcome and invitation templates with local synthetic
values, validates the complete encryption ring and active key, and checks its
owned adapter is configured and open. It never sends mail, performs an email
provider ping or waits for provider availability. `/ready` reports database,
Redis and email components; `/live` remains process-only. Missing or malformed
local configuration fails closed without returning key or provider diagnostics.

Shutdown marks the worker unready and stops SQL dispatch/queue intake before
draining active work. One caller deadline bounds the shutdown result. An active
send retains its PostgreSQL and Redis dependencies until it finishes; deadline
exhaustion returns a failure and the role command exits nonzero. Cleanup closes
consumer, email adapter, Redis and PostgreSQL in reverse ownership order, with
telemetry last. Partially allocated resources are released on startup failure,
including an email adapter returned alongside an error. Existing welcome jobs,
legacy retry/correlation migration, API-only optional producers, explicit
migrators and independent redirectors keep their resource boundaries.

## Phase 2 real-provider acceptance (plans 34–35)

Local signed browser fixtures prove application boundaries. They do not prove
real password factors, inbox verification, recovery, social authorization,
production cookies or actual delivery. The separate `provider` Playwright
project uses production Next, the production Go API/migrator/worker, and
disposable PostgreSQL/Redis using the existing digest pins. It never mounts a
backend dotenv file or replaces authentication transports. The default root
`test:e2e` and full gate select only `local`; all 65 local cases remain required.

Prepare private configuration before plan 35. Do not paste keys, passwords,
cookies, inbox links or codes into chat. Use an external secret store or a
mode-0600 environment file outside this checkout; existing ignored
`apps/frontend/.env.local` can hold frontend configuration. The root process
must receive the same values, for example with Bun's `--env-file` option.
Do not commit private files or add scanner exceptions.

Exact application settings:

- `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`: development-instance publishable key.
- `CLERK_SECRET_KEY` and `FLUX_AUTH.SECRET_KEY`: identical development-instance
  backend key. Prerequisites compare real BAPI and issuer JWKS; prefixes alone
  cannot establish instance identity.
- `FLUX_AUTH.ISSUER`: exact HTTPS issuer encoded by the publishable key.
- `FLUX_AUTH.AUTHORIZED_PARTIES=http://localhost:3000` and
  `FLUX_API_ORIGIN=http://127.0.0.1:8080`: this dedicated test origin and rewrite.
- `FLUX_INVITATIONS.ACTIVE_KEY_ID`, `FLUX_INVITATIONS.ENCRYPTION_KEYS`,
  `FLUX_INVITATIONS.SENDER`, `FLUX_INVITATIONS.PUBLIC_ORIGIN` and
  `FLUX_INTEGRATION.RESEND_API_KEY`: the production authorities described above.
  Use an approved verified sender and fixed HTTPS public origin. The test ring
  needs at least two distinct decoded 32-byte keys so old pending ciphertext and
  rotated writes can both be tested; this does not change the production ring policy.
- `FLUX_LINKS.MANAGED_HOST`, optional `FLUX_LINKS.BLOCKED_HOSTS` and
  `FLUX_LINKS.CURSOR_KEY`: existing operator policies, never test fallback keys.

Test-only settings:

- `FLUX_PROVIDER_TEST_EMAIL` / `FLUX_PROVIDER_TEST_PASSWORD`: an existing
  verified operator-controlled password account.
- `FLUX_PROVIDER_INVITEE_EMAIL` / `FLUX_PROVIDER_INVITEE_PASSWORD`: a different
  existing verified operator-controlled password account.
- `FLUX_PROVIDER_SIGNUP_EMAIL` / `FLUX_PROVIDER_SIGNUP_PASSWORD`: a third
  operator-controlled fresh signup identity. Passwords must be at least 12 characters.
- `FLUX_PROVIDER_AUTHORIZED_RECIPIENTS`: the explicit comma-separated approved
  inbox addresses, including all three identities. No address is inferred from
  source, browser history or environment files.
- `FLUX_PROVIDER_PRIVATE_DIR`: an existing owned mode-0700 absolute directory
  outside the repository, without symlink components. Raw Playwright media and
  traces remain disabled. Only sanitized evidence is retained there.

In the Clerk development dashboard enable password signup/login, mandatory email
verification, password recovery, Google and GitHub, and the default seven-day
maximum session lifetime. Do not enable testing-email shortcuts as factor
evidence. The suite uses real UI/password and real provider challenges; it does
not substitute helper sign-in for factors. Google/GitHub buttons must actually
reach their providers. Development factor challenges can still require the
operator's authorized inbox/OAuth/MFA interaction.

Only plan 35, after authorized inputs/recipients are confirmed, runs:

```sh
bun run provider:check
bun run test:e2e:provider
bun run provider:automated:check
# After remaining human-only rows are verified:
bun run provider:acceptance:check
```

Missing prerequisites return a safe nonzero error before browser collection or
startup; they never become skipped tests. The dedicated harness builds static
Go binaries, creates only its own pinned disposable containers/networks, runs
migrations and starts the API. Next is built and started in production mode.
The worker starts after an old-key intent is queued, with the rotated active key
and the retained ring. After its single acknowledged invitation email, it stops.
A second encrypted intent proves new-key writes and is revoked while the worker
is stopped, so that proof sends no second invitation email. Cleanup removes
only this run's ephemeral containers, networks and binaries. It cannot delete
existing databases or unrelated Docker resources.

For the actual invitation, an operator-authorized inbox integration must write
`inbox.json` (mode 0600) inside the private directory with an
`invitation_url` containing the real received invitation link. Keep this file
out of reports. The suite checks the fixed configured public origin and bounded
token fragment, then presents that fragment to the same production UI on its
local test origin. Fresh production recipient verification and token-digest
matching must succeed. A fabricated or stale link cannot count as acceptance.
Password sign-in may also require Clerk Device Trust on a fresh browser. The
dedicated password-login test consumes its initial real challenge email and
never clicks Resend automatically. Unrelated session/workspace tests bootstrap
separate real Clerk sessions with the official `clerk.signIn({ emailAddress })`
helper, avoiding repeated password challenges. Helper setup is not factor
evidence; signup, password login, recovery and OAuth keep their real UI flows.
The provider run stops at the first failed test and never retries tests.
The suite writes a mode-0600 `verification-request-<request-id>.json` in the private
directory containing `request_id`, the authorized `email`, and `requested_at`.
The authorized inbox integration must read the actual new verification email and
write `verification-response-<request-id>.json` (mode 0600) containing the same
`request_id`, a fresh ISO `received_at`, and its six-digit `code`. The suite checks
request identity and freshness, enters that code through Clerk's UI, and removes
both files after consumption or timeout. Do not use Clerk testing codes or disable
Device Trust to satisfy this check. Keep the integration and all codes outside
the repository and reports.
The executor can operate an already-authorized inbox integration; no unsolicited
mail or unapproved recipient is allowed. The root test control endpoint is
loopback-only, protected by a per-run random bearer and limited to closed,
UUID-scoped assertions against this run's disposable database.

The outage test detaches only the API's own provider-egress network, retains
healthy PostgreSQL/Redis, requires a safe 503 within the production deadline,
then reconnects and requires successful recovery. Production endpoints and
timeouts are unchanged. A run-owned TCP relay binds only loopback and forwards
bytes to the API's internal Docker address, preserving API ingress while Docker
removes published ports during egress isolation. Capacity evidence samples actual production API requests
at 2 requests/second, requires success and p95 below two seconds, and records only
this tested rate. Clerk currently documents a development BAPI ceiling of
100 requests/10 seconds and production ceiling of 1000/10 seconds; this is not a
production load claim. Reassess the expected deployment rate/quota before launch.
See [Clerk rate limits](https://clerk.com/docs/guides/how-clerk-works/system-limits)
and [session options](https://clerk.com/docs/guides/secure/session-options).

Native collection and execution identities must match exactly, with positive
single-pass results and no skips/retries/flaky cases. Successful provider
execution prepares uniquely named `provider-proof-<run-id>.json` and
`provider-acceptance-<run-id>.md` in the private directory. During plan 35 the
executor promotes the sanitized Markdown to
`.planning/phases/02-tenant-safe-link-control-plane/02-PROVIDER-ACCEPTANCE.md`
and copies its matching proof to private `provider-proof.json`. These files
contain no raw tokens, email addresses, passwords, request bodies or cookies.
The record binds its digest/run to the successful provider execution, source
revision, current instance/origin fingerprint and a maximum age of seven days.
This is evidence integrity checking, not a cryptographic attestation against
an operator who can edit both private files.

Every required row must exist. Automated rows include signup/login, named owner,
invited onboarding, last-used selection and membership-loss tabs, first Links,
refresh/expiry/revocation/signout/max session, outage, measured capacity and actual
delivery/encryption rotation. Only the following rows may remain
`human_pending` in automated mode, each with its exact listed manual steps and
`evidence_kind=human_only`:

| Row | Allowed reason |
| --- | --- |
| email_verification, password_recovery | authorized_inbox_verification |
| google_oauth, github_oauth | oauth_account_interaction or operator_mfa |
| production_cookie_inspection | https_cookie_inspection |
| provider_key_rotation | provider_dashboard_key_rotation |

Those rows remain incomplete. HTTP development cookies cannot prove HTTPS
Secure/SameSite/HttpOnly behavior; current JWKS cannot prove rotation. Only
actual operator verification may change a human row to `human_verified`, with
the exact steps and a fresh `verified_at`. The final strict mode rejects all
pending rows. Positive SDK execution cannot satisfy a missing factor. The
provider run does not complete Phase 2 until strict acceptance and the final
phase audit pass.
