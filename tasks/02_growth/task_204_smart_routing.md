---id: TASK-204
title: Smart Routing Engine (Geo, Device & Time)
layer: Level 5 (Executable Task Unit)
status: Done
owner: Feature Engineer
depends_on:
  - docs/core/redirect_engine.md#part-ii
references:
  - [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md)
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - smart-routing
  - geo-routing
  - device-targeting
---

# TASK-204: Smart Routing Engine (Geo, Device & Time)

## 1. Goal
Implement conditional redirect router evaluating Geo-IP, OS device type, and time windows.

## 2. Scope & Target Boundaries
Target source files modified during this task:
- `internal/modules/redirect/router.go`

## 3. Dependencies & Prerequisites
This task relies on specifications and schema contracts in:
  - docs/core/redirect_engine.md#part-ii

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md)
  - [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md)

## 5. Acceptance Criteria
- [x] Routes visitors to targeted destinations based on matching rule evaluation.
- [x] Code follows all formatting and linting rules in [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md).
- [x] Latency and memory usage adhere to targets in [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md).

## 6. Target Deliverables
- `internal/modules/redirect/router.go`

## 7. Definition of Done (DoD)
- [x] **Step 1: Write failing unit test** verifying expected feature behavior.
- [x] **Step 2: Confirm test failure**: Run `go test -v ./internal/modules/redirect/...`.
- [x] **Step 3: Write minimal implementation** inside target files.
- [x] **Step 4: Confirm test passes**: Run `go test -v ./internal/modules/redirect/...`.
- [x] **Step 5: Run linter**: Confirm zero warnings or errors.

## 8. Testing Strategy
- **Unit Tests**: Test pure domain logic in isolation.
- **Verification Command**: `go test -v ./internal/modules/redirect/...`

## 9. Documentation Updates
- [x] Mark task status as `Done` in this document frontmatter upon completion.
- [x] Update function/package docstrings if signatures were extended.


<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/core/redirect_engine.md](file:///home/logan78/Desktop/flux/docs/core/redirect_engine.md) |
| **Previous** | [tasks/02_growth/task_203_custom_domains.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_203_custom_domains.md) |
| **Next** | [tasks/02_growth/task_205_deep_linking.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_205_deep_linking.md) |
| **Children** | None |
| **Dependencies** | [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql), [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md), [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md), [ai/TESTING.md](file:///home/logan78/Desktop/flux/ai/TESTING.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
