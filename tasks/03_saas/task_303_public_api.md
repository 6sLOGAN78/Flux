---id: TASK-303
title: Public REST API Platform & OAuth 2.0
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Feature Engineer
depends_on:
  - api/openapi_v3_saas.yaml
references:
  - [public_api_oauth.md](file:///home/logan78/Desktop/flux/docs/saas/public_api_oauth.md)
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - public-api
  - oauth2
  - api-tokens
---

# TASK-303: Public REST API Platform & OAuth 2.0

## 1. Goal
Implement OAuth 2.0 authorization server and public API rate-limited endpoints.

## 2. Scope & Target Boundaries
Target source files modified during this task:
- `internal/modules/publicapi/`

## 3. Dependencies & Prerequisites
This task relies on specifications and schema contracts in:
  - api/openapi_v3_saas.yaml

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md)
  - [public_api_oauth.md](file:///home/logan78/Desktop/flux/docs/saas/public_api_oauth.md)

## 5. Acceptance Criteria
- [ ] Issues OAuth tokens (`flx_live_...`) and validates public API requests.
- [ ] Code follows all formatting and linting rules in [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md).
- [ ] Latency and memory usage adhere to targets in [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md).

## 6. Target Deliverables
- `internal/modules/publicapi/oauth.go`

## 7. Definition of Done (DoD)
- [ ] **Step 1: Write failing unit test** verifying expected feature behavior.
- [ ] **Step 2: Confirm test failure**: Run `go test -v ./internal/modules/publicapi/...`.
- [ ] **Step 3: Write minimal implementation** inside target files.
- [ ] **Step 4: Confirm test passes**: Run `go test -v ./internal/modules/publicapi/...`.
- [ ] **Step 5: Run linter**: Confirm zero warnings or errors.

## 8. Testing Strategy
- **Unit Tests**: Test pure domain logic in isolation.
- **Verification Command**: `go test -v ./internal/modules/publicapi/...`

## 9. Documentation Updates
- [ ] Mark task status as `Done` in this document frontmatter upon completion.
- [ ] Update function/package docstrings if signatures were extended.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/saas/public_api_oauth.md](file:///home/logan78/Desktop/flux/docs/saas/public_api_oauth.md) |
| **Previous** | [tasks/03_saas/task_302_stripe_billing.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_302_stripe_billing.md) |
| **Next** | [tasks/03_saas/task_304_webhook_engine.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_304_webhook_engine.md) |
| **Children** | None |
| **Dependencies** | [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql), [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md), [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md), [ai/TESTING.md](file:///home/logan78/Desktop/flux/ai/TESTING.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
