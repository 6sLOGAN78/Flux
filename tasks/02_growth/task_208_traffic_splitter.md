---id: TASK-208
title: A/B Testing & Weighted Traffic Splitter
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Feature Engineer
depends_on:
  - docs/growth/ab_testing.md
references:
  - [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md)
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - ab-testing
  - traffic-splitter
  - experiments
---

# TASK-208: A/B Testing & Weighted Traffic Splitter

## 1. Goal
Implement weighted random traffic distribution engine for link A/B testing variants.

## 2. Scope & Target Boundaries
Target source files modified during this task:
- `internal/modules/abtest/`

## 3. Dependencies & Prerequisites
This task relies on specifications and schema contracts in:
  - docs/growth/ab_testing.md

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md)
  - [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md)

## 5. Acceptance Criteria
- [ ] Splits traffic according to variant weights (e.g. 50/50, 80/20) with sticky visitor sessions.
- [ ] Code follows all formatting and linting rules in [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md).
- [ ] Latency and memory usage adhere to targets in [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md).

## 6. Target Deliverables
- `internal/modules/abtest/splitter.go`

## 7. Definition of Done (DoD)
- [ ] **Step 1: Write failing unit test** verifying expected feature behavior.
- [ ] **Step 2: Confirm test failure**: Run `go test -v ./internal/modules/abtest/...`.
- [ ] **Step 3: Write minimal implementation** inside target files.
- [ ] **Step 4: Confirm test passes**: Run `go test -v ./internal/modules/abtest/...`.
- [ ] **Step 5: Run linter**: Confirm zero warnings or errors.

## 8. Testing Strategy
- **Unit Tests**: Test pure domain logic in isolation.
- **Verification Command**: `go test -v ./internal/modules/abtest/...`

## 9. Documentation Updates
- [ ] Mark task status as `Done` in this document frontmatter upon completion.
- [ ] Update function/package docstrings if signatures were extended.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/growth/ab_testing.md](file:///home/logan78/Desktop/flux/docs/growth/ab_testing.md) |
| **Previous** | [tasks/02_growth/task_207_advanced_qr.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_207_advanced_qr.md) |
| **Next** | [tasks/02_growth/task_209_job_queue.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_209_job_queue.md) |
| **Children** | None |
| **Dependencies** | [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql), [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md), [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md), [ai/TESTING.md](file:///home/logan78/Desktop/flux/ai/TESTING.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
