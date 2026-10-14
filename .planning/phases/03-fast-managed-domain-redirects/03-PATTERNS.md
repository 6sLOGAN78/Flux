# Phase 3 implementation patterns

Inspected 2026-10-14. New paths below are intentional additions, not claims that modules already exist. Preserve the existing concrete server resource container; do not rewrite adjacent product modules.

| New/change target | Closest existing analogue | Contract to preserve |
|---|---|---|
| internal/redirect/{identity,repository,resolver,handler}.go (new) | safety/url.go; repository/link.go; handler/health.go | Pure normalization/safety; narrow parameterized minimal projection; explicit constructor dependencies and bounded context. No auth dependency on public GET. |
| internal/app/redirector.go, api.go | NewAPI/newRole and defaultRoleRouter in api.go | Explicit resource ownership; startup unwind; real listener; health-only sibling worker; no docs/API routing leak. |
| internal/config/config.go | normalizeAPIPolicies, excludeUnownedConfig, validateRoleSections | Dot-nested FLUX names; role validation only for owned resources; redirect policy must not require CursorKey/Clerk/Resend. |
| internal/{repository,service,handler}/link_edit.go (new) | link.go Suspension/Transition and product.go adapters | Fresh membership before replay, workspace/actor/link lock order, mutation ledger exact response snapshot and canonical hash. New method files avoid enlarging the existing large modules. |
| PATCH route/schemas | identity.ts ZLinkTransitionRequest and ts-rest transitionLink | One authored Zod+contract authority; generated transport/health.gen.go and both OpenAPI copies; strict unknown field handling. |
| link edit form | components/link-detail.tsx lifecycle/suspension behavior; lib/links.ts LifecycleRequests | Immutable version+idempotency intent, session/workspace request cancellation, explicit review after conflict, native labels/status. |
| redirect/cache.go + embedded cache.lua (new) | lib/job Redis client ownership; testing/container.go | Existing go-redis version; real Redis script proof; decimal-string versions and mandatory primary guard; tiny closed script operations and same hash slot. |
| migrations/014_redirect_outbox.sql (new) | 012_link_lifecycle_audit.sql, 013_link_suspension_audit.sql | Tern up/down separator, constrained schema and transactional invariants; no manual schema push or migration at every HTTP startup. |
| redirect/outbox.go and relay.go (new) | repository/mutation.go and app/lifecycle.go | Short claim transaction, leased token ack, cancellation, one shutdown deadline; no Redis network inside locked SQL transaction. |
| internal/middleware/redirect_proxy.go (new) | auth.go strict input boundary; config/config_test.go | Explicit IPExtractor and closed policy, no trust inferred from private network or arbitrary headers. |
| redirect metrics | observability/telemetry.go; middleware/tracing.go | Injected provider, fixed route name and safe attributes; collector must preserve allowed fields and remove planted identifiers. |
| integration cases | app/roles_test.go, handler/product_test.go | SetupTestDB/SetupTestRedis real resources, t.Cleanup, race detector. New top-level groups added to scripts/check.ts exact registry. |
| browser redirect journey | internal/testing/browser_fixture_test.go; tests/control-plane.spec.ts | Test-only signed provider transport below app authorization; real Go/PG and actual public redirector, no production login bypass. |
| load acceptance | app/observability_test.go; scripts/check.ts stage catalog | Actual transport/telemetry, positive sample evidence, gate rejects missing tests; no `t.Skip` to disguise prerequisites. |

Key existing signatures: `NewRedirector(ctx context.Context, cfg *config.Config) (*Redirector, error)`; `RoleRuntime.Run(ctx) error`, `Close(ctx) error`, `Address() string`; `config.NormalizeLinkHost(string) (string,error)`; `safety.ValidateDestination(string, config.LinksConfig) (string,error)`; `testing.SetupTestDB(t)` and `SetupTestRedis(t)` return resource plus cleanup. Inspect exact return types and fixture helpers before implementation. `Handle` adapters require fresh per-invocation factories; new edit adapter may follow current strict product JSON decoder instead.

Ownership rule: every executor stages only declared files and its summary. Respect README.md, .serena/project.yml, unrelated quick plans and private .playwright-cli outputs. Shared configuration/router/root runner files have explicit dependency edges; no simultaneous wave edits to them.
