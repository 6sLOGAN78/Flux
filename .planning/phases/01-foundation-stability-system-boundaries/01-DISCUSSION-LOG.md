# Phase 1: Foundation Stability & System Boundaries - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-10-05
**Phase:** 1-Foundation Stability & System Boundaries
**Areas discussed:** Process ownership, Lifecycle and health, Contract and asset authority, Reproducible quality gates, Observability and configuration
**Mode:** `--auto`; every selected answer was the recommended default.

---

## Process Ownership

| Option | Description | Selected |
|--------|-------------|----------|
| Thin independent composition roots | API, redirector, worker, and migrator own only required resources | ✓ |
| Runtime mode flag | One binary selects a role through configuration | |
| Keep coupled server | Defer separation until product features arrive | |

**Auto-selected choice:** Thin independent composition roots.
**Notes:** Preserve Echo and sound packages; extract incrementally around ownership and testability rather than rewriting the repository tree.

---

## Lifecycle and Health

| Option | Description | Selected |
|--------|-------------|----------|
| Role-specific live and ready endpoints | Liveness is process-only; readiness checks required role dependencies | ✓ |
| One aggregate status endpoint | Continue combining health and dependency detail | |
| Provider-managed health only | Delegate all checks to deployment infrastructure | |

**Auto-selected choice:** Role-specific liveness/readiness with ordered bounded shutdown.
**Notes:** Handle SIGINT and SIGTERM, sanitize public states, unwind partial startup, attempt all cleanup, and close dependents before shared resources.

---

## Contract and Asset Authority

| Option | Description | Selected |
|--------|-------------|----------|
| Deterministic OpenAPI boundary | Keep TS/Zod authoring, generate one artifact, and generate/verify Go boundaries | ✓ |
| Go annotations as second source | Maintain independent Go and TypeScript contract sources | |
| Manual synchronization | Continue updating generated/served files by convention | |

**Auto-selected choice:** One deterministic OpenAPI boundary from the existing contracts.
**Notes:** Generation must fail on write errors and drift. Assets must be embedded or resolved from explicit validated roots, never the current working directory.

---

## Reproducible Quality Gates

| Option | Description | Selected |
|--------|-------------|----------|
| Pinned toolchain and behavior gates | One supported stack with real unit/integration/lifecycle checks | ✓ |
| Coverage threshold | Use a broad line-coverage number as the principal gate | |
| Build and lint only | Defer runtime regression tests | |

**Auto-selected choice:** One pinned supported toolchain with behavior-focused quality gates.
**Notes:** Bun is the supported workspace package manager. CI must exercise backend and all packages, clean migrations, role lifecycle, contract equality, alternate working directories, validation safety, and real PostgreSQL/Redis paths.

---

## Observability and Configuration

| Option | Description | Selected |
|--------|-------------|----------|
| OpenTelemetry/OTLP boundary | Vendor-neutral instrumentation through a collector | ✓ |
| Direct New Relic instrumentation | Keep provider SDKs in product infrastructure paths | |
| Logs only | Defer traces and metrics | |

**Auto-selected choice:** OpenTelemetry and OTLP collector boundary, with typed configuration errors and compatibility-first naming.
**Notes:** Composition roots decide startup failure. Instrumentation and diagnostics redact secrets and sensitive data; exact patch versions remain an implementation-time verification.

## the agent's Discretion

- Exact package names and extraction order.
- Exact supported patch versions after official verification.
- OpenAPI-to-Go generator, CI job layout, and per-asset embed/config choice.
- Test file organization and bounded metric naming.

## Deferred Ideas

None. Product-domain work remains in Phases 2–6.
