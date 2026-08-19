---
id: TASK-631
title: Multi-Touch Attribution Model Studio & Touchpoint Journey Visualizer
layer: Level 5 (Executable Task Unit)
status: Completed
owner: Frontend Engineer
depends_on:
  - tasks/09_frontend_analytics_attribution/task_630_timeseries_analytics_explorer.md
references:
  - api/openapi_v4_enterprise.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - attribution
  - linear
  - time-decay
  - u-shaped
  - enterprise
---

# TASK-631: Multi-Touch Attribution Model Studio & Touchpoint Journey Visualizer

## 1. Goal
Build the Enterprise Multi-Touch Attribution Studio (`/attribution`) allowing users to switch dynamically between 5 attribution models (First-Touch, Last-Touch, Linear, Time-Decay, Position-Based/U-Shaped), compare campaign revenue share, and inspect customer journey timelines.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/analytics/AttributionPage.tsx`
- `apps/frontend/src/components/attribution/ModelSelectorBar.tsx`
- `apps/frontend/src/components/attribution/AttributionComparisonTable.tsx`
- `apps/frontend/src/components/attribution/TouchpointTimelineFlow.tsx`

## 3. Dependencies & Prerequisites
- `@flux/zod` (`ZAttributionResult`, `ZAttributionModel`)

## 4. Referenced Architecture & Product Specs
- [docs/enterprise/attribution_engine.md](file:///home/logan78/Desktop/flux/docs/enterprise/attribution_engine.md)
- [tasks/04_enterprise/task_401_attribution_engine.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_401_attribution_engine.md)

## 5. Acceptance Criteria
- [x] Model selector tabs switch algorithm recalculations seamlessly.
- [x] Table highlights Attributed Conversions and Attributed Revenue with percentage share bars.
- [x] Visual journey flow demonstrates how weight is distributed across first touch (40%), middle touchpoints (20%), and conversion touch (40%) for position-based models.

## 6. Target Deliverables
- `apps/frontend/src/pages/analytics/AttributionPage.tsx`
- `apps/frontend/src/components/attribution/ModelSelectorBar.tsx`

## 7. Definition of Done (DoD)
- [x] Step 1: Write attribution table calculation unit tests.
- [x] Step 2: Confirm test failure.
- [x] Step 3: Implement Attribution page.
- [x] Step 4: Confirm tests pass.
- [x] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/analytics/AttributionPage.test.tsx`
