---id: TASK-108
title: Redis Sliding Window Rate Limiting Middleware
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Feature Engineer
depends_on:
  - ai/PERFORMANCE.md
references:
  - [ai/SECURITY.md](file:///home/logan78/Desktop/plan/ai/SECURITY.md)
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - rate-limiter
  - redis
  - sliding-window
---

# TASK-108: Redis Sliding Window Rate Limiting Middleware

## 1. Goal
Implement rate limiting middleware enforcing IP and API key request quotas.

## 2. Scope & Target Boundaries
Target source files modified during this task:
- `internal/middleware/ratelimit.go`

## 3. Dependencies & Prerequisites
This task relies on specifications and schema contracts in:
  - ai/PERFORMANCE.md

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md)
  - [ai/SECURITY.md](file:///home/logan78/Desktop/plan/ai/SECURITY.md)

## 5. Acceptance Criteria
- [ ] Returns HTTP 429 Too Many Requests when sliding window quota is exceeded.
- [ ] Code follows all formatting and linting rules in [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/plan/ai/CODING_STANDARDS.md).
- [ ] Latency and memory usage adhere to targets in [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md).

## 6. Target Deliverables
- `internal/middleware/ratelimit.go`

## 7. Definition of Done (DoD)
- [ ] **Step 1: Write failing unit test** verifying expected feature behavior.
- [ ] **Step 2: Confirm test failure**: Run `go test -v ./internal/middleware/...`.
- [ ] **Step 3: Write minimal implementation** inside target files.
- [ ] **Step 4: Confirm test passes**: Run `go test -v ./internal/middleware/...`.
- [ ] **Step 5: Run linter**: Confirm zero warnings or errors.

## 8. Testing Strategy
- **Unit Tests**: Test pure domain logic in isolation.
- **Verification Command**: `go test -v ./internal/middleware/...`

## 9. Documentation Updates
- [ ] Mark task status as `Done` in this document frontmatter upon completion.
- [ ] Update function/package docstrings if signatures were extended.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md) |
| **Previous** | [tasks/01_core/task_107_qr_generator.md](file:///home/logan78/Desktop/plan/tasks/01_core/task_107_qr_generator.md) |
| **Next** | [tasks/02_growth/task_201_clickhouse_pipeline.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_201_clickhouse_pipeline.md) |
| **Children** | None |
| **Dependencies** | [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql), [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md), [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/plan/ai/CODING_STANDARDS.md), [ai/TESTING.md](file:///home/logan78/Desktop/plan/ai/TESTING.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
