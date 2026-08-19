---
id: TASK-632
title: Conversion Funnels Builder & Unit Economics ROAS Dashboard
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Frontend Engineer
depends_on:
  - tasks/09_frontend_analytics_attribution/task_631_attribution_engine_studio.md
references:
  - api/openapi_v4_enterprise.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - funnels
  - revenue
  - roas
  - cac
  - ltv
---

# TASK-632: Conversion Funnels Builder & Unit Economics ROAS Dashboard

## 1. Goal
Implement the Conversion Funnels & Revenue Analytics dashboard (`/funnels`) with multi-step funnel creation, drop-off rate visualization, and comprehensive marketing unit economics (Ad Spend, CAC, ROAS, LTV, and LTV:CAC ratios).

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/analytics/FunnelsPage.tsx`
- `apps/frontend/src/components/funnels/FunnelVisualizer.tsx`
- `apps/frontend/src/components/funnels/CreateFunnelModal.tsx`
- `apps/frontend/src/components/revenue/UnitEconomicsCards.tsx`

## 3. Dependencies & Prerequisites
- `@flux/zod` (`ZFunnelAnalysisResult`, `ZRevenueSummaryResult`)

## 4. Referenced Architecture & Product Specs
- [docs/enterprise/revenue_analytics.md](file:///home/logan78/Desktop/flux/docs/enterprise/revenue_analytics.md)
- [tasks/04_enterprise/task_402_funnel_analytics.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_402_funnel_analytics.md)

## 5. Acceptance Criteria
- [ ] Funnel visualizer renders sequential conversion bars showing drop-off percentage between consecutive steps.
- [ ] Unit economics cards display Total Spend, Attributed Revenue, CAC ($), ROAS (x.x), and LTV:CAC health indicators (>3.0 = healthy).
- [ ] Funnel creation modal permits adding re-orderable steps mapped to specific campaign link IDs.

## 6. Target Deliverables
- `apps/frontend/src/pages/analytics/FunnelsPage.tsx`
- `apps/frontend/src/components/funnels/FunnelVisualizer.tsx`

## 7. Definition of Done (DoD)
- [ ] Step 1: Write funnel metric formatting unit tests.
- [ ] Step 2: Confirm test failure.
- [ ] Step 3: Implement Funnels page.
- [ ] Step 4: Confirm tests pass.
- [ ] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/analytics/FunnelsPage.test.tsx`
