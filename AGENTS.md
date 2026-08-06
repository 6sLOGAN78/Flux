# AGENTS.md — AI Agent Operating System & Protocol

> **Universal Directive**: This document is the master operating system for every AI coding agent operating in this repository. It defines non-negotiable rules for behavior, workflow, architecture, testing, and execution. It contains **zero project specifications** and is 100% reusable.

---

## 1. Document Precedence Hierarchy

When requirements or instructions conflict, AI agents MUST enforce this top-down priority order:

1. **`AGENTS.md`** *(Universal Operating Protocol & Safety Rules)*
2. **Active Task Document** *(`tasks/*/*.md` — Executable Work Unit)*
3. **Product Intent** *(`PRODUCT.md` / PRD — Domain Scope & Non-Goals)*
4. **System Architecture** *(`docs/ARCHITECTURE.md` / HLD — Component Boundaries)*
5. **Security & Performance** *(`ai/SECURITY.md`, `ai/PERFORMANCE.md` — SLAs & Constraints)*
6. **Machine Schemas** *(`api/*.yaml`, `database/*.sql` — Authoritative Contracts)*
7. **Quality Standards** *(`ai/CODING_STANDARDS.md`, `ai/TESTING.md` — Code Guidelines)*
8. **Reference Documentation** *(`docs/*`, `adr/*`, `ops/*` — Context Deep-Dives)*

---

## 2. AI Behavior & Execution Protocol

- **Context Isolation**: Load only the current task document and its direct schema dependencies. Never ingest unrelated subsystem files.
- **Empirical Rigor**: Base all diagnoses strictly on full, un-truncated runtime error logs. Never form hypotheses without inspecting tracebacks.
- **No Assumptions**: Never guess package names, function signatures, SQL column types, or API schemas. Always verify against `api/` and `database/`.
- **Surgical Edits**: Make minimal, precise code modifications. Never reformat, rewrite, or restructure code outside the active task scope.
- **Silent Log Synthesis**: Inspect background command outputs silently and synthesize findings concisely in natural language.

---

## 3. Engineering Philosophy

- **Single Responsibility Principle (SRP)**: Every file, class, package, and document MUST answer exactly one question.
- **DRY & YAGNI**: Eliminate duplicate logic and unused abstractions. Never build speculative features.
- **Explicit Over Implicit**: Prefer explicit type declarations, structured error wrapping, and clear interface implementations over magical reflection or dynamic typing.
- **Zero Boilerplate Bloat**: Keep prompts and task specifications lightweight and high-density.

---

## 4. Planning & Workflow Protocol

All feature work MUST follow this 5-step task workflow:

```text
Step 1: Read Task & Load Schemas  ──►  Step 2: Write Failing Test (Red)
                                                  │
Step 5: Verify & Commit (Done)    ◄──  Step 4: Verify Test Passes (Green)  ◄──  Step 3: Minimal Implementation
```

1. **Task Scope Lock**: Ingest the active task file from `tasks/`. Verify inputs, outputs, and target file boundaries.
2. **Schema Alignment**: Inspect target schema contracts (`api/*.yaml`, `database/*.sql`).
3. **Task Right-Sizing**: Ensure task changes remain under ~200 lines of code. Split tasks that span multiple components.

---

## 5. Implementation Rules

- **No Symptom Masking**: Never resolve failures by swallowing exceptions, adding dummy fallbacks, or disabling broken assertions. Fix the root cause.
- **Immutable Public Contracts**: Never alter public API signatures or database column types without an explicit ADR and schema update.
- **Strict Error Wrapping**: Every internal error must be caught, contextually wrapped, and bubbled up appropriately.
- **Null & Type Safety**: Guard against null pointer dereferences and undefined properties before property access.

---

## 6. Architecture Rules

- **Strict Layered Package Structure**: All backend code MUST strictly follow the layered `internal/` package structure:
  - `internal/config/`: Application configuration & environment loader (`config.go`)
  - `internal/database/`: PostgreSQL pool initialization (`database.go`) & versioned SQL migrations (`migrations/`)
  - `internal/errs/`: Domain errors (`types.go`) & HTTP error mappers (`http.go`)
  - `internal/handler/`: Controller HTTP handlers (`base.go`, `health.go`, domain handlers)
  - `internal/lib/`: Reusable utilities (`aws/`, `email/`, `job/`, `utils/`)
  - `internal/logger/`: Structured Zerolog logger setup (`logger.go`)
  - `internal/middleware/`: HTTP middlewares (`auth.go`, `global.go`, `rate_limit.go`, `request_id.go`, `tracing.go`)
  - `internal/model/`: Pure domain entities & DTO payloads (`base.go`, subpackage models)
  - `internal/repository/`: SQL persistence & Redis cache implementations (`repositories.go`, domain repos)
  - `internal/router/`: Echo route registration (`router.go`, `system.go`, `v1/`)
  - `internal/server/`: Server lifecycle & dependency injection (`server.go`)
  - `internal/service/`: Framework-agnostic business logic services (`services.go`, domain services)
  - `internal/testing/`: Testcontainers-go integration test suites (`container.go`, `assertions.go`, `helpers.go`, `server.go`)
  - `internal/validation/`: Request payload validation utilities (`utils.go`)
- **Decoupled Domain Logic**: Keep core business logic pure and framework-agnostic. Separate HTTP handlers, database queries, and domain entities.
- **No Cyclic Dependencies**: Dependencies flow strictly inward: `router` ➔ `handler` ➔ `service` ➔ `repository` / `model`. Circular imports are illegal.
- **Idempotent State Operations**: Database mutations, queue message processing, and cache writes MUST be idempotent and safe for retries.

---

## 7. Testing Strategy (TDD Mandate)

- **Test-First Requirement**: Implementation code MUST NOT be written until a failing test is written and verified to fail.
- **Unit Tests**: Test pure domain logic in isolation using fast unit tests (<10ms per test).
- **Integration Tests**: Test storage repositories and API routes against real local dependencies using containerized DBs.
- **Benchmarks**: Every SLA-critical path (e.g. redirect handlers, hash encoders) MUST include a Go/JS benchmark test verifying latency budgets.

---

## 8. Review Process & Verification Commands

Before declaring any task complete, the agent MUST run and verify:

1. **Test Suite**: Run `cd apps/backend && go test -v ./...` (or `bun test`). Result MUST be 100% PASS.
2. **Static Analysis**: Run linter (`golangci-lint run` / `bun run lint`). Result MUST be 0 warnings/errors.
3. **Diff Check**: Review git diff to confirm no unexpected files or trailing whitespaces were touched.

---

## 9. Definition of Done (DoD)

A task is **DONE** only when ALL of the following criteria are verified:

- [ ] Failing test written and empirically confirmed to fail (`Red`).
- [ ] Minimal implementation written making the test pass (`Green`).
- [ ] Full test suite executes cleanly with 0 failures.
- [ ] Linter executes with 0 warnings or errors.
- [ ] Code strictly respects target file boundaries declared in the task document.
- [ ] Zero duplicated PRD, HLD, or schema text added to source code comments.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) |
| **Previous** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) |
| **Next** | [PRODUCT.md](file:///home/logan78/Desktop/plan/PRODUCT.md) |
| **Children** | [tasks/01_core/task_100_bootstrap_backend.md](file:///home/logan78/Desktop/plan/tasks/01_core/task_100_bootstrap_backend.md) |
| **Dependencies** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
