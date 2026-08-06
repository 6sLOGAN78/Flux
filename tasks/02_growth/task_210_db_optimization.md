---id: TASK-210
title: DB Read Replicas & Cache Invalidation
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Feature Engineer
depends_on:
  - ai/PERFORMANCE.md
references:
  - [postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql)
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - db-replicas
  - cache-invalidation
  - l1-l2-cache
---

# TASK-210: DB Read Replicas & Cache Invalidation

## 1. Goal
Implement database read-replica splitting and Redis L1/L2 cache invalidation pub/sub.

## 2. Scope & Target Boundaries
Target source files modified during this task:
- `internal/db/replica.go`

## 3. Dependencies & Prerequisites
This task relies on specifications and schema contracts in:
  - ai/PERFORMANCE.md

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md)
  - [postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql)

## 5. Acceptance Criteria
- [ ] Directs read queries to read-replicas and invalidates Redis cache across nodes.
- [ ] Code follows all formatting and linting rules in [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/plan/ai/CODING_STANDARDS.md).
- [ ] Latency and memory usage adhere to targets in [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md).

## 6. Target Deliverables
- `internal/db/replica.go`

## 7. Definition of Done (DoD)
- [ ] **Step 1: Write failing unit test** verifying expected feature behavior.
- [ ] **Step 2: Confirm test failure**: Run `go test -v ./internal/db/...`.
- [ ] **Step 3: Write minimal implementation** inside target files.
- [ ] **Step 4: Confirm test passes**: Run `go test -v ./internal/db/...`.
- [ ] **Step 5: Run linter**: Confirm zero warnings or errors.

## 8. Testing Strategy
- **Unit Tests**: Test pure domain logic in isolation.
- **Verification Command**: `go test -v ./internal/db/...`

## 9. Documentation Updates
- [ ] Mark task status as `Done` in this document frontmatter upon completion.
- [ ] Update function/package docstrings if signatures were extended.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md) |
| **Previous** | [tasks/02_growth/task_209_job_queue.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_209_job_queue.md) |
| **Next** | [tasks/03_saas/task_301_tenant_rbac.md](file:///home/logan78/Desktop/plan/tasks/03_saas/task_301_tenant_rbac.md) |
| **Children** | None |
| **Dependencies** | [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql), [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md), [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/plan/ai/CODING_STANDARDS.md), [ai/TESTING.md](file:///home/logan78/Desktop/plan/ai/TESTING.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
