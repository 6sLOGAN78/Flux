---
id: TASK-622
title: A/B Traffic Splitter & Statistical Significance Visualizer
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Frontend Engineer
depends_on:
  - tasks/08_frontend_growth_marketing/task_620_campaigns_utm_builder.md
references:
  - api/openapi_v2_growth.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - ab-testing
  - traffic-splitter
  - experiment
  - statistics
---

# TASK-622: A/B Traffic Splitter & Statistical Significance Visualizer

## 1. Goal
Implement the A/B Testing & Traffic Splitter management page (`/traffic-splits`) with interactive percentage allocation sliders (summing to 100%), statistical significance indicators, and one-click winner promotion.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/growth/ABTestingPage.tsx`
- `apps/frontend/src/components/abtest/VariantAllocationSlider.tsx`
- `apps/frontend/src/components/abtest/SignificanceScoreCard.tsx`

## 3. Dependencies & Prerequisites
- `@flux/zod` (`ZABVariant`)

## 4. Referenced Architecture & Product Specs
- [docs/growth/ab_testing.md](file:///home/logan78/Desktop/flux/docs/growth/ab_testing.md)
- [tasks/02_growth/task_208_traffic_splitter.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_208_traffic_splitter.md)

## 5. Acceptance Criteria
- [ ] Sliders enforce total weight constraint of 100% with intuitive auto-balancing.
- [ ] Real-time CTR comparisons highlight winning variants with statistical confidence tags (e.g. 95% confidence).
- [ ] Lock Winner button immediately sets winner allocation to 100% and disables losing variants.

## 6. Target Deliverables
- `apps/frontend/src/pages/growth/ABTestingPage.tsx`
- `apps/frontend/src/components/abtest/VariantAllocationSlider.tsx`

## 7. Definition of Done (DoD)
- [ ] Step 1: Write weight calculation unit tests.
- [ ] Step 2: Confirm test failure.
- [ ] Step 3: Implement A/B testing page.
- [ ] Step 4: Confirm tests pass.
- [ ] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/growth/ABTestingPage.test.tsx`
