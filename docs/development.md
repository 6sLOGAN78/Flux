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
