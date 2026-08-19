---
id: TASK-643
title: Outbound Webhooks Manager, Delivery Log & Event Payload Inspector
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Frontend Engineer
depends_on:
  - tasks/10_frontend_saas_developer/task_642_developer_api_keys_oauth.md
references:
  - api/openapi_v3_saas.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - webhooks
  - event-bus
  - payloads
  - retries
---

# TASK-643: Outbound Webhooks Manager, Delivery Log & Event Payload Inspector

## 1. Goal
Implement the Webhooks Manager page (`/developer/webhooks`) supporting outbound webhook endpoint registration, event trigger subscriptions (`link.created`, `click.recorded`), delivery history inspection, and manual re-delivery triggers.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/developer/WebhooksPage.tsx`
- `apps/frontend/src/components/webhooks/WebhookEndpointList.tsx`
- `apps/frontend/src/components/webhooks/CreateWebhookModal.tsx`
- `apps/frontend/src/components/webhooks/WebhookDeliveryHistory.tsx`

## 3. Dependencies & Prerequisites
- `@flux/openapi` (`createWebhook`, `ZWebhook`)

## 4. Referenced Architecture & Product Specs
- [docs/saas/webhooks_event_bus.md](file:///home/logan78/Desktop/flux/docs/saas/webhooks_event_bus.md)
- [tasks/03_saas/task_304_webhook_engine.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_304_webhook_engine.md)

## 5. Acceptance Criteria
- [ ] Creation modal validates HTTPS URLs and allows multi-selecting event subscription checkboxes.
- [ ] Delivery history shows timestamp, response status code (200, 500), latency ms, and expandable JSON payload inspector.
- [ ] Retry button sends test event and provides immediate feedback.

## 6. Target Deliverables
- `apps/frontend/src/pages/developer/WebhooksPage.tsx`
- `apps/frontend/src/components/webhooks/WebhookDeliveryHistory.tsx`

## 7. Definition of Done (DoD)
- [ ] Step 1: Write delivery status badge and URL validation unit tests.
- [ ] Step 2: Confirm test failure.
- [ ] Step 3: Implement Webhooks page.
- [ ] Step 4: Confirm tests pass.
- [ ] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/developer/WebhooksPage.test.tsx`
