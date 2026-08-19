---
id: TASK-621
title: Dynamic Smart Routing Rules Engine & Mobile Deep Linking
layer: Level 5 (Executable Task Unit)
status: Completed
owner: Frontend Engineer
depends_on:
  - tasks/07_frontend_core_ops/task_613_link_detail_qr_studio.md
references:
  - api/openapi_v2_growth.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - smart-routing
  - geo-targeting
  - deep-linking
  - device-rules
---

# TASK-621: Dynamic Smart Routing Rules Engine & Mobile Deep Linking

## 1. Goal
Build the Smart Routing configuration interface (`/routing`) supporting geo-location targeting rules (by ISO country code), device/OS routing (iOS, Android, Desktop), and mobile deep link app scheme configuration (Universal Links / App Links).

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/growth/SmartRoutingPage.tsx`
- `apps/frontend/src/components/routing/GeoRoutingRuleBuilder.tsx`
- `apps/frontend/src/components/routing/DeviceRuleSelector.tsx`
- `apps/frontend/src/components/routing/DeepLinkConfigCard.tsx`

## 3. Dependencies & Prerequisites
- `@flux/openapi`

## 4. Referenced Architecture & Product Specs
- [docs/growth/smart_routing.md](file:///home/logan78/Desktop/flux/docs/growth/smart_routing.md)
- [docs/growth/deep_linking.md](file:///home/logan78/Desktop/flux/docs/growth/deep_linking.md)

## 5. Acceptance Criteria
- [x] Rule builder permits adding ordered condition sets (e.g. `If Country == 'GB' -> Destination URL`).
- [x] Deep linking card supports setting iOS Bundle ID, Android Package Name, and fallback web URLs.
- [x] Visual flow diagram displays priority order and fallback route clearly.

## 6. Target Deliverables
- `apps/frontend/src/pages/growth/SmartRoutingPage.tsx`
- `apps/frontend/src/components/routing/GeoRoutingRuleBuilder.tsx`

## 7. Definition of Done (DoD)
- [x] Step 1: Write rule validation unit test.
- [x] Step 2: Confirm test failure.
- [x] Step 3: Implement routing interface.
- [x] Step 4: Confirm tests pass.
- [x] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/growth/SmartRoutingPage.test.tsx`
