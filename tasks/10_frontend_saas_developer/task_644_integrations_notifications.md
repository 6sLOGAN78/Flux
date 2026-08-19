---
id: TASK-644
title: Integrations Directory & Real-Time In-App Notification Center
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Frontend Engineer
depends_on:
  - tasks/06_frontend_foundation/task_603_app_shell_navigation.md
references:
  - api/openapi_v3_saas.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - integrations
  - notifications
  - slack
  - zapier
---

# TASK-644: Integrations Directory & Real-Time In-App Notification Center

## 1. Goal
Build the Integrations Directory page (`/integrations`) with one-click connectors (Slack, Zapier, Segment, HubSpot) and the slide-over In-App Notification Center (`/notifications`) with unread counters and batch mark-as-read actions.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/developer/IntegrationsPage.tsx`
- `apps/frontend/src/pages/dashboard/NotificationsPage.tsx`
- `apps/frontend/src/components/notifications/NotificationDrawer.tsx`
- `apps/frontend/src/components/integrations/AppConnectorCard.tsx`

## 3. Dependencies & Prerequisites
- `@flux/openapi` (`getNotifications`, `markNotificationsRead`, `ZNotification`)

## 4. Referenced Architecture & Product Specs
- [docs/saas/notifications.md](file:///home/logan78/Desktop/flux/docs/saas/notifications.md)
- [tasks/03_saas/task_305_notifications.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_305_notifications.md)

## 5. Acceptance Criteria
- [ ] App connector cards show connection status (`Connected`, `Connect`) with setup modal.
- [ ] Notification drawer displays unread count badge in header and auto-updates upon clicking Mark All Read.
- [ ] Severity pills differentiate between `info` (blue), `warning` (amber), and `alert` (rose).

## 6. Target Deliverables
- `apps/frontend/src/pages/developer/IntegrationsPage.tsx`
- `apps/frontend/src/components/notifications/NotificationDrawer.tsx`

## 7. Definition of Done (DoD)
- [ ] Step 1: Write notification mark-read mutation unit tests.
- [ ] Step 2: Confirm test failure.
- [ ] Step 3: Implement Integrations and Notifications.
- [ ] Step 4: Confirm tests pass.
- [ ] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/developer/IntegrationsPage.test.tsx`
