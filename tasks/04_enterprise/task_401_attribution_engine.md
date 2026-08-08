---id: TASK-401
title: Multi-Touch Attribution Engine
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Feature Engineer
depends_on:
  - docs/enterprise/attribution_engine.md
references:
  - [openapi_v4_enterprise.yaml](file:///home/logan78/Desktop/flux/api/openapi_v4_enterprise.yaml)
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - attribution-engine
  - linear
  - decay
  - u-shaped
---

# TASK-401: Multi-Touch Attribution Engine

## 1. Goal
Implement First, Last, Linear, Time-Decay, and U-Shaped (Position) attribution calculation models.

## 2. Scope & Target Boundaries
Target source files modified during this task:
- `internal/modules/attribution/`

## 3. Dependencies & Prerequisites
This task relies on specifications and schema contracts in:
  - docs/enterprise/attribution_engine.md

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md)
  - [openapi_v4_enterprise.yaml](file:///home/logan78/Desktop/flux/api/openapi_v4_enterprise.yaml)

## 5. Acceptance Criteria
- [ ] Calculates conversion credit across customer touchpoints using selected attribution model.
- [ ] Code follows all formatting and linting rules in [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md).
- [ ] Latency and memory usage adhere to targets in [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md).

## 6. Target Deliverables
- `internal/modules/attribution/calculator.go`

## 7. Definition of Done (DoD)
- [ ] **Step 1: Write failing unit test** verifying expected feature behavior.
- [ ] **Step 2: Confirm test failure**: Run `go test -v ./internal/modules/attribution/...`.
- [ ] **Step 3: Write minimal implementation** inside target files.
- [ ] **Step 4: Confirm test passes**: Run `go test -v ./internal/modules/attribution/...`.
- [ ] **Step 5: Run linter**: Confirm zero warnings or errors.

## 8. Testing Strategy
- **Unit Tests**: Test pure domain logic in isolation.
- **Verification Command**: `go test -v ./internal/modules/attribution/...`

## 9. Documentation Updates
- [ ] Mark task status as `Done` in this document frontmatter upon completion.
- [ ] Update function/package docstrings if signatures were extended.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/enterprise/attribution_engine.md](file:///home/logan78/Desktop/flux/docs/enterprise/attribution_engine.md) |
| **Previous** | [tasks/03_saas/task_307_admin_audit.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_307_admin_audit.md) |
| **Next** | [tasks/04_enterprise/task_402_funnel_analytics.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_402_funnel_analytics.md) |
| **Children** | None |
| **Dependencies** | [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql), [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md), [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md), [ai/TESTING.md](file:///home/logan78/Desktop/flux/ai/TESTING.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
