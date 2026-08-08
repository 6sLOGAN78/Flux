---id: TASK-501
title: Multi-Region Cloudflare Workers Edge Redirects
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Feature Engineer
depends_on:
  - docs/core/redirect_engine.md#part-iii
references:
  - [PERFORMANCE.md#slas](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md#slas)
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - cloudflare-workers
  - edge-redirects
  - sub-10ms
---

# TASK-501: Multi-Region Cloudflare Workers Edge Redirects

## 1. Goal
Implement Cloudflare Workers JavaScript edge redirect worker with KV lookup fallbacks.

## 2. Scope & Target Boundaries
Target source files modified during this task:
- `workers/redirect.js`

## 3. Dependencies & Prerequisites
This task relies on specifications and schema contracts in:
  - docs/core/redirect_engine.md#part-iii

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md)
  - [PERFORMANCE.md#slas](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md#slas)

## 5. Acceptance Criteria
- [ ] Executes redirects at edge closest to visitor with sub-10ms global latency.
- [ ] Code follows all formatting and linting rules in [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md).
- [ ] Latency and memory usage adhere to targets in [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md).

## 6. Target Deliverables
- `workers/redirect.js`

## 7. Definition of Done (DoD)
- [ ] **Step 1: Write failing unit test** verifying expected feature behavior.
- [ ] **Step 2: Confirm test failure**: Run `npm test -- workers/redirect.test.js`.
- [ ] **Step 3: Write minimal implementation** inside target files.
- [ ] **Step 4: Confirm test passes**: Run `npm test -- workers/redirect.test.js`.
- [ ] **Step 5: Run linter**: Confirm zero warnings or errors.

## 8. Testing Strategy
- **Unit Tests**: Test pure domain logic in isolation.
- **Verification Command**: `npm test -- workers/redirect.test.js`

## 9. Documentation Updates
- [ ] Mark task status as `Done` in this document frontmatter upon completion.
- [ ] Update function/package docstrings if signatures were extended.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/global/edge_redirect_workers.md](file:///home/logan78/Desktop/flux/docs/global/edge_redirect_workers.md) |
| **Previous** | [tasks/04_enterprise/task_407_abuse_malware.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_407_abuse_malware.md) |
| **Next** | [tasks/05_global/task_502_geo_replication.md](file:///home/logan78/Desktop/flux/tasks/05_global/task_502_geo_replication.md) |
| **Children** | None |
| **Dependencies** | [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql), [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md), [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md), [ai/TESTING.md](file:///home/logan78/Desktop/flux/ai/TESTING.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
