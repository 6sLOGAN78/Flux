---id: TASK-202
title: Campaign Management & UTM Parameter Builder
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Feature Engineer
depends_on:
  - api/openapi_v2_growth.yaml
references:
  - [campaign_utm_builder.md](file:///home/logan78/Desktop/plan/docs/growth/campaign_utm_builder.md)
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - utm-builder
  - campaign
  - analytics
---

# TASK-202: Campaign Management & UTM Parameter Builder

## 1. Goal
Implement UTM parameter builder and campaign link grouping service.

## 2. Scope & Target Boundaries
Target source files modified during this task:
- `internal/modules/campaign/`

## 3. Dependencies & Prerequisites
This task relies on specifications and schema contracts in:
  - api/openapi_v2_growth.yaml

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md)
  - [campaign_utm_builder.md](file:///home/logan78/Desktop/plan/docs/growth/campaign_utm_builder.md)

## 5. Acceptance Criteria
- [ ] Appends sanitized UTM parameters and groups link analytics by campaign.
- [ ] Code follows all formatting and linting rules in [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/plan/ai/CODING_STANDARDS.md).
- [ ] Latency and memory usage adhere to targets in [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md).

## 6. Target Deliverables
- `internal/modules/campaign/utm.go`

## 7. Definition of Done (DoD)
- [ ] **Step 1: Write failing unit test** verifying expected feature behavior.
- [ ] **Step 2: Confirm test failure**: Run `go test -v ./internal/modules/campaign/...`.
- [ ] **Step 3: Write minimal implementation** inside target files.
- [ ] **Step 4: Confirm test passes**: Run `go test -v ./internal/modules/campaign/...`.
- [ ] **Step 5: Run linter**: Confirm zero warnings or errors.

## 8. Testing Strategy
- **Unit Tests**: Test pure domain logic in isolation.
- **Verification Command**: `go test -v ./internal/modules/campaign/...`

## 9. Documentation Updates
- [ ] Mark task status as `Done` in this document frontmatter upon completion.
- [ ] Update function/package docstrings if signatures were extended.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/growth/campaign_utm_builder.md](file:///home/logan78/Desktop/plan/docs/growth/campaign_utm_builder.md) |
| **Previous** | [tasks/02_growth/task_201_clickhouse_pipeline.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_201_clickhouse_pipeline.md) |
| **Next** | [tasks/02_growth/task_203_custom_domains.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_203_custom_domains.md) |
| **Children** | None |
| **Dependencies** | [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql), [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md), [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/plan/ai/CODING_STANDARDS.md), [ai/TESTING.md](file:///home/logan78/Desktop/plan/ai/TESTING.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
