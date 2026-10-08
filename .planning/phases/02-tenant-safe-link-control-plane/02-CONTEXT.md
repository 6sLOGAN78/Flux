# Phase 2: Tenant-Safe Link Control Plane - Context

**Gathered:** 2026-10-08
**Status:** Ready for planning

<domain>
## Phase Boundary

Authenticated users can manage teams and managed-domain links inside enforceable workspace boundaries. Deliver identity mapping, workspace creation and switching, membership and invitations, documented owner/admin/member/viewer permissions, managed-domain link creation and inspection, search and cursor pagination, lifecycle transitions, durable mutation conflict handling, and administrator abuse suspension.

The phase must enforce secure sessions, appropriate CSRF protection, parameterized workspace-scoped access, explicit transactions, destination/egress safety and automated two-workspace denial tests. Preserve the independently operable Go roles and the verified foundation. Actual redirect behavior and cache invalidation belong to Phase 3; click tracking, conversions, attribution and analytics belong to later phases. Custom domains, campaign productivity, billing and enterprise identity are outside this slice.

</domain>

<decisions>
## Implementation Decisions

### Sign-in and Onboarding

- **D-01:** Use email and password for the email authentication option, with email verification and password recovery. Include Google and GitHub sign-in as specified in `spec.md` §22. The user selected email/password rather than an emailed sign-in link.
- **D-02:** A new user creating their first workspace must explicitly enter a workspace name; create the workspace and its owner membership atomically. Do not silently create a default personal workspace. Valid invitation onboarding should lead to the invited workspace rather than force creation of an unrelated workspace; retain explicit, authorized invitation acceptance.
- **D-03:** Returning users reopen their last-used workspace only while they remain a member. If that membership is no longer valid, show their available workspaces without exposing the former workspace's data. Server-side membership checks remain authoritative, including across browser tabs and previously stored workspace selections.
- **D-04:** A newly created workspace opens its Links page. The empty state prominently offers “Create your first link”; team invitations are not a mandatory step before accessing links. Do not imply that redirects or analytics are delivered by this phase.

### the agent's Discretion

The user chose to discuss only sign-in and onboarding, then explicitly chose “Capture context” with the other areas left to research and planning. Team permission details and invitation policies, custom-key rules and collision feedback, and the link library/lifecycle interaction design are delegated within the existing roadmap requirements. Research and plan these choices, document the resulting policies, and preserve all required capabilities; lack of discussion is not permission to omit them.

- Evaluate the existing Clerk integration and choose the exact supported identity/session implementation during research. Clerk is an existing asset, not a separately confirmed user lock. Never use provider organization claims as a substitute for authoritative Flux workspace membership and permissions.
- Choose session expiry/rotation behavior, account recovery presentation, invitation entry handling and workspace-selection persistence with appropriate security and denial tests.
- Design the initial dashboard shell, workspace switcher, responsive layout, forms, loading/error states and keyboard/accessibility behavior within the original product direction in `spec.md` §§48–50.
- Choose precise member permissions, owner-preservation rules, invitation expiry and acceptance behavior; short-key normalization and reservation; lifecycle transition rules and concurrency/idempotency semantics. Make these explicit and testable before execution.
- Choose internal packages, SQL/migration layout, transaction boundaries, narrowly injected dependencies and test organization without rewriting functioning foundation code.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Product Contract

- `.planning/ROADMAP.md` §Phase 2 — phase goal, MVP mode, dependency on Phase 1 and five observable success criteria.
- `.planning/REQUIREMENTS.md` — `TEN-01` through `TEN-08`, `LINK-01`, `LINK-02`, `LINK-03`, `LINK-05` through `LINK-09`, and `SAFE-02` through `SAFE-04`. Link destination/title editing and redirect invalidation are allocated to Phase 3 through `LINK-04`.
- `.planning/PROJECT.md` — core value, production constraints, brownfield preservation, privacy and original dashboard direction.
- `spec.md` §§5–6, 21–22, 48–50, 54 — long-term link operations, workspace roles, initial email/Google/GitHub authentication, Go-backed frontend, visual direction and security. The phase allocation takes precedence over the long-term feature inventory.

### Foundation and Current Evidence

- `.planning/phases/01-foundation-stability-system-boundaries/01-CONTEXT.md` — locked process ownership, contract authority, real-dependency testing, observability privacy and configuration compatibility.
- `.planning/phases/01-foundation-stability-system-boundaries/01-VERIFICATION.md` — completed foundation evidence; no product domain or frontend delivery is implied.
- `.planning/phases/01-foundation-stability-system-boundaries/01-CI-SCAN-REPAIR.md` — full-history scanner behavior and safe failed-test diagnostics; new planning files remain subject to scans.
- `docs/development.md` — current runtime, role, local development and quality-gate instructions.
- `.planning/codebase/STACK.md`, `.planning/codebase/ARCHITECTURE.md`, `.planning/codebase/INTEGRATIONS.md` — historical baseline only; verify claims against current source because Phase 1 changed lifecycle, telemetry, runtime assets and CI.

The Clerk documentation consulted for the email choices is https://clerk.com/docs/guides/configure/auth-strategies/sign-up-sign-in-options. Research must recheck current provider documentation before selecting SDK versions or implementation APIs. No external ADR was supplied during this discussion.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `apps/backend/internal/service/auth.go` and `apps/backend/internal/middleware/auth.go`: Clerk initialization and header-authentication scaffold. Existing role/permission context fields represent provider claims; product workspace authorization is not implemented.
- `apps/backend/internal/database/` and `apps/backend/internal/testing/`: pgx, explicit embedded Tern migrations and pinned real PostgreSQL/Redis integration helpers.
- `apps/backend/internal/handler/`, `middleware/`, `errs/` and `validation/`: request-local handler construction, centralized safe errors, correlation, logging and validation scaffolding.
- `packages/zod/`, `packages/openapi/`, `scripts/generate.ts` and `apps/backend/internal/transport/`: authored contract and checked generated-boundary pipeline.
- `apps/backend/internal/lib/job/` and `packages/emails/`: independently owned worker/producer and email template infrastructure for invitation delivery where justified.

### Established Patterns

- Go/Echo business logic and explicit constructor injection remain authoritative; the frontend calls the Go API rather than owning product logic.
- PostgreSQL is the transactional source of truth. Redis remains derived/transient, with tenant identity required at all future ownership boundaries.
- Root gates execute actual package checks, compiled Go test discovery, race-enabled unit/integration groups, deterministic generation, migration checks and dependency/worktree/full-history scans.
- Safe observability never emits raw credentials, personal values or arbitrary underlying provider diagnostics.

### Integration Points

- `apps/backend/internal/router/router.go`: the `/api/v1` group is currently empty and will receive authenticated product routes.
- `apps/backend/internal/repository/repositories.go`: persistence registry remains a scaffold; workspace-scoped queries and transactions must be implemented.
- `apps/backend/internal/app/`: API composition owns product dependencies; preserve separation from redirector, worker and migrator graphs.
- `apps/frontend/`: no running dashboard or reusable frontend component system exists. Implement the scoped product screens and include the new workspace in all applicable root quality checks.
- `scripts/check.ts`: register new real container-backed tests explicitly and ensure no source-discovered test is silently omitted.

</code_context>

<specifics>
## Specific Ideas

- The first successful workspace creation should lead directly to an actionable empty Links page, rather than a compulsory team setup wizard.
- Workspace selection should save a returning user's context without weakening current membership checks.
- Minimalist, accessible, responsive, information-dense and original visual design carries forward from the product specification. No visual reference or protected third-party assets were requested.

</specifics>

<deferred>
## Deferred Ideas

None were proposed during discussion. Redirects, analytics and the other later-phase capabilities remain outside Phase 2 as already allocated by the roadmap.

</deferred>

---

*Phase: 2-Tenant-Safe Link Control Plane*
*Context gathered: 2026-10-08*
