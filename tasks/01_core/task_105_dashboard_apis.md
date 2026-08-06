---id: TASK-105
title: Dashboard Metrics REST Backend APIs
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Feature Engineer
depends_on:
  - api/openapi_v1_core.yaml
references:
  - [analytics_pipeline.md](file:///home/logan78/Desktop/plan/docs/core/analytics_pipeline.md)
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - dashboard-api
  - metrics
  - rest-api
---

# TASK-105: Dashboard Metrics REST Backend APIs

## 1. Goal
Implement link management and analytics summary REST API endpoints for user dashboard.

## 2. Scope & Target Boundaries
Target source files modified during this task:
- `internal/modules/analytics/api.go`

## 3. Dependencies & Prerequisites
This task relies on specifications and schema contracts in:
  - api/openapi_v1_core.yaml

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md)
  - [analytics_pipeline.md](file:///home/logan78/Desktop/plan/docs/core/analytics_pipeline.md)

## 5. Acceptance Criteria
- [ ] REST endpoints return correct aggregated link statistics with pagination.
- [ ] Code follows all formatting and linting rules in [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/plan/ai/CODING_STANDARDS.md).
- [ ] Latency and memory usage adhere to targets in [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md).

## 6. Target Deliverables
- `internal/modules/analytics/api.go`

## 7. Definition of Done (DoD)
- [ ] **Step 1: Write failing unit test** verifying expected feature behavior.
- [ ] **Step 2: Confirm test failure**: Run `go test -v ./internal/modules/analytics/...`.
- [ ] **Step 3: Write minimal implementation** inside target files.
- [ ] **Step 4: Confirm test passes**: Run `go test -v ./internal/modules/analytics/...`.
- [ ] **Step 5: Run linter**: Confirm zero warnings or errors.

## 8. Testing Strategy
- **Unit Tests**: Test pure domain logic in isolation.
- **Verification Command**: `go test -v ./internal/modules/analytics/...`

## 9. Documentation Updates
- [ ] Mark task status as `Done` in this document frontmatter upon completion.
- [ ] Update function/package docstrings if signatures were extended.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/core/analytics_pipeline.md](file:///home/logan78/Desktop/plan/docs/core/analytics_pipeline.md) |
| **Previous** | [tasks/01_core/task_104_analytics_ingestion.md](file:///home/logan78/Desktop/plan/tasks/01_core/task_104_analytics_ingestion.md) |
| **Next** | [tasks/01_core/task_106_analytics_enrichment.md](file:///home/logan78/Desktop/plan/tasks/01_core/task_106_analytics_enrichment.md) |
| **Children** | None |
| **Dependencies** | [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql), [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md), [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/plan/ai/CODING_STANDARDS.md), [ai/TESTING.md](file:///home/logan78/Desktop/plan/ai/TESTING.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
