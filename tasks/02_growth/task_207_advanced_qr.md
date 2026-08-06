---id: TASK-207
title: Advanced Custom QR Microservice
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Feature Engineer
depends_on:
  - docs/core/qr_service.md#part-ii
references:
  - [openapi_v2_growth.yaml](file:///home/logan78/Desktop/plan/api/openapi_v2_growth.yaml)
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - styled-qr
  - vector-qr
  - microservice
---

# TASK-207: Advanced Custom QR Microservice

## 1. Goal
Implement logo embedding, color gradient matrix, and vector SVG output for QR codes.

## 2. Scope & Target Boundaries
Target source files modified during this task:
- `internal/modules/qr/advanced.go`

## 3. Dependencies & Prerequisites
This task relies on specifications and schema contracts in:
  - docs/core/qr_service.md#part-ii

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md)
  - [openapi_v2_growth.yaml](file:///home/logan78/Desktop/plan/api/openapi_v2_growth.yaml)

## 5. Acceptance Criteria
- [ ] Renders styled QR codes with embedded center logos and custom brand colors.
- [ ] Code follows all formatting and linting rules in [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/plan/ai/CODING_STANDARDS.md).
- [ ] Latency and memory usage adhere to targets in [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md).

## 6. Target Deliverables
- `internal/modules/qr/advanced.go`

## 7. Definition of Done (DoD)
- [ ] **Step 1: Write failing unit test** verifying expected feature behavior.
- [ ] **Step 2: Confirm test failure**: Run `go test -v ./internal/modules/qr/...`.
- [ ] **Step 3: Write minimal implementation** inside target files.
- [ ] **Step 4: Confirm test passes**: Run `go test -v ./internal/modules/qr/...`.
- [ ] **Step 5: Run linter**: Confirm zero warnings or errors.

## 8. Testing Strategy
- **Unit Tests**: Test pure domain logic in isolation.
- **Verification Command**: `go test -v ./internal/modules/qr/...`

## 9. Documentation Updates
- [ ] Mark task status as `Done` in this document frontmatter upon completion.
- [ ] Update function/package docstrings if signatures were extended.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/core/qr_service.md](file:///home/logan78/Desktop/plan/docs/core/qr_service.md) |
| **Previous** | [tasks/02_growth/task_206_og_meta.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_206_og_meta.md) |
| **Next** | [tasks/02_growth/task_208_traffic_splitter.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_208_traffic_splitter.md) |
| **Children** | None |
| **Dependencies** | [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql), [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md), [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/plan/ai/CODING_STANDARDS.md), [ai/TESTING.md](file:///home/logan78/Desktop/plan/ai/TESTING.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
