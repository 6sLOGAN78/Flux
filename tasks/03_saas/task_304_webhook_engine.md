---id: TASK-304
title: Webhook Delivery Engine & Event Bus
layer: Level 5 (Executable Task Unit)
status: Done
owner: Feature Engineer
depends_on:
  - api/asyncapi_events.yaml
references:
  - [webhooks_event_bus.md](file:///home/logan78/Desktop/flux/docs/saas/webhooks_event_bus.md)
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - webhooks
  - hmac-sha256
  - event-bus
---

# TASK-304: Webhook Delivery Engine & Event Bus

## 1. Goal
Implement outbound webhook event bus with HMAC-SHA256 signatures and backoff retries.

## 2. Scope & Target Boundaries
Target source files modified during this task:
- `internal/modules/webhook/`

## 3. Dependencies & Prerequisites
This task relies on specifications and schema contracts in:
  - api/asyncapi_events.yaml

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md)
  - [webhooks_event_bus.md](file:///home/logan78/Desktop/flux/docs/saas/webhooks_event_bus.md)

## 5. Acceptance Criteria
- [x] Delivers signed webhook payloads to subscriber URLs with execution retry logs.
- [x] Code follows all formatting and linting rules in [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md).
- [x] Latency and memory usage adhere to targets in [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md).

## 6. Target Deliverables
- `internal/modules/webhook/deliver.go`

## 7. Definition of Done (DoD)
- [x] **Step 1: Write failing unit test** verifying expected feature behavior.
- [x] **Step 2: Confirm test failure**: Run `go test -v ./internal/modules/webhook/...`.
- [x] **Step 3: Write minimal implementation** inside target files.
- [x] **Step 4: Confirm test passes**: Run `go test -v ./internal/modules/webhook/...`.
- [x] **Step 5: Run linter**: Confirm zero warnings or errors.

## 8. Testing Strategy
- **Unit Tests**: Test pure domain logic in isolation.
- **Verification Command**: `go test -v ./internal/modules/webhook/...`

## 9. Documentation Updates
- [x] Mark task status as `Done` in this document frontmatter upon completion.
- [x] Update function/package docstrings if signatures were extended.


<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/saas/webhooks_event_bus.md](file:///home/logan78/Desktop/flux/docs/saas/webhooks_event_bus.md) |
| **Previous** | [tasks/03_saas/task_303_public_api.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_303_public_api.md) |
| **Next** | [tasks/03_saas/task_305_notifications.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_305_notifications.md) |
| **Children** | None |
| **Dependencies** | [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql), [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md), [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md), [ai/TESTING.md](file:///home/logan78/Desktop/flux/ai/TESTING.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
