# Plan 01-21 dependency prerequisite repair

The real scanner checkpoint reported 72 Bun advisories and three Go module advisories. No exception or ignore baseline was added.

## Changes and evidence

- Updated Echo to v4.15.3 and go-jose/v3 to v3.0.5 using their published fixes. Broader package analysis including tests exposed the archive dependency; updated moby/go-archive to v0.3.0 and klauspost/compress to v1.18.7, with required transitive updates. Go checksums remain authoritative.
- Added 17 exact overrides for existing vulnerable JavaScript dependencies, resolved from official npm registry metadata and verified by Bun's integrity-checked lock. A fresh Bun audit reports zero advisories. No new direct package was introduced.
- Removed unused tsc-alias and its paired concurrently watcher from the Zod/OpenAPI workspaces. Their authored imports already use relative ESM paths. This removes the unpatched braces ancestry; builds and watch commands use the existing TypeScript compiler.
- `bun install --frozen-lockfile` passed without changes. Actual `bun run check:fast` passed every stage, including all workspaces, meaningful tests, pinned lint/type checks, role builds, and exact generated-byte verification.
- Fresh `go build ./...` and the full `go test -race ./...` passed, including PostgreSQL, Redis, collector and role subprocess regressions.

## Scanner scope correction still owned by 01-21

The initial module-only scan of cmd/api cannot distinguish unrelated packages in a module and omits test-only package exposure. Use the pinned govulncheck's package-level analysis over `./...` with `-test`, retaining module inventory diagnostics separately and failing every vulnerable imported package. This expands analyzed targets to all roles and tests; it is stricter than default symbol reachability.

Actual package analysis after patches reports zero vulnerable imports. It still reports GO-2026-5932 in the module inventory: the legacy `golang.org/x/crypto/openpgp` family is unmaintained and has no fixed version. `go list -deps -test ./...` proves none of those packages is imported. Flux needs other x/crypto packages. Do not claim the module inventory is clean, hard-code an advisory ignore, or hide its diagnostic. A regression must show that adding a vulnerable import changes the result to failure.

Primary references: [Go advisory](https://vuln.go.dev/ID/GO-2026-5932.json), [govulncheck scope and test flags](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck), [Echo fix](https://github.com/labstack/echo/releases/tag/v4.15.3), [go-jose advisory](https://vuln.go.dev/ID/GO-2026-4945.json). Exact JavaScript versions and integrity values are recorded in package.json and bun.lock.

Plan 01-21 remains incomplete until its wrapper scope correction, regressions and actual combined scans pass. No completion summary or tracking advancement is implied by this prerequisite report.
