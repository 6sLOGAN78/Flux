---id: TASK-104
title: Analytics Click Event Collector
layer: Level 5 (Executable Task Unit)
status: Done
owner: Feature Engineer
depends_on:
  - api/asyncapi_events.yaml#ClickEvent
references:
  - [analytics_pipeline.md](file:///home/logan78/Desktop/flux/docs/core/analytics_pipeline.md)
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - analytics
  - click-collector
  - kafka
---

# TASK-104: Analytics Click Event Collector

## 1. Goal
Implement async click event producer writing event payloads to Redis Stream / Kafka.

## 2. Scope & Target Boundaries
Target source files modified during this task:
- `internal/modules/analytics/`

## 3. Dependencies & Prerequisites
This task relies on specifications and schema contracts in:
  - api/asyncapi_events.yaml#ClickEvent

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md)
  - [analytics_pipeline.md](file:///home/logan78/Desktop/flux/docs/core/analytics_pipeline.md)

## 5. Acceptance Criteria
- [x] Click events published asynchronously without blocking redirect response path.
- [x] Code follows all formatting and linting rules in [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md).
- [x] Latency and memory usage adhere to targets in [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md).

## 6. Target Deliverables
- `internal/modules/analytics/collector.go`

## 7. Definition of Done (DoD)
- [x] **Step 1: Write failing unit test** verifying expected feature behavior.
- [x] **Step 2: Confirm test failure**: Run `go test -v ./internal/modules/analytics/...`.
- [x] **Step 3: Write minimal implementation** inside target files.
- [x] **Step 4: Confirm test passes**: Run `go test -v ./internal/modules/analytics/...`.
- [x] **Step 5: Run linter**: Confirm zero warnings or errors.

## 8. Testing Strategy
- **Unit Tests**: Test pure domain logic in isolation.
- **Verification Command**: `go test -v ./internal/modules/analytics/...`

## 9. Documentation Updates
- [x] Mark task status as `Done` in this document frontmatter upon completion.
- [x] Update function/package docstrings if signatures were extended.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/core/analytics_pipeline.md](file:///home/logan78/Desktop/flux/docs/core/analytics_pipeline.md) |
| **Previous** | [tasks/01_core/task_103_redirect_handler.md](file:///home/logan78/Desktop/flux/tasks/01_core/task_103_redirect_handler.md) |
| **Next** | [tasks/01_core/task_105_dashboard_apis.md](file:///home/logan78/Desktop/flux/tasks/01_core/task_105_dashboard_apis.md) |
| **Children** | None |
| **Dependencies** | [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql), [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md), [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md), [ai/TESTING.md](file:///home/logan78/Desktop/flux/ai/TESTING.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
