---
id: TASK-610
title: Public Landing Hero with Redirect Simulator & Pricing Matrix
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Frontend Engineer
depends_on:
  - tasks/06_frontend_foundation/task_603_app_shell_navigation.md
  - tasks/06_frontend_foundation/task_604_ui_component_primitives.md
references:
  - PRODUCT.md
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - landing
  - pricing
  - public
  - marketing
---

# TASK-610: Public Landing Hero with Redirect Simulator & Pricing Matrix

## 1. Goal
Implement public marketing landing page (`/`) featuring an interactive instant shortener simulator, <10ms edge latency live counter, and pricing comparison matrix page (`/pricing`) with Stripe tier checkout triggers.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/public/LandingPage.tsx`
- `apps/frontend/src/pages/public/PricingPage.tsx`
- `apps/frontend/src/components/public/HeroRedirectSimulator.tsx`
- `apps/frontend/src/components/public/PricingMatrix.tsx`

## 3. Dependencies & Prerequisites
- `@flux/openapi`
- `react-router-dom` v7

## 4. Referenced Architecture & Product Specs
- `.agents/skills/design-taste-frontend/SKILL.md` (Anti-Slop directives)
- [PRODUCT.md](file:///home/logan78/Desktop/flux/PRODUCT.md)

## 5. Acceptance Criteria
- [ ] Landing hero headline and value prop fit within initial desktop viewport with max 2 lines.
- [ ] Interactive shortener simulates live Base62 encoding and displays instant preview without requiring login.
- [ ] Pricing matrix features Free, Pro, and Enterprise tiers with Monthly/Annual discount toggle.

## 6. Target Deliverables
- `apps/frontend/src/pages/public/LandingPage.tsx`
- `apps/frontend/src/pages/public/PricingPage.tsx`

## 7. Definition of Done (DoD)
- [ ] Step 1: Write page render unit tests.
- [ ] Step 2: Confirm test failure.
- [ ] Step 3: Implement Landing and Pricing pages.
- [ ] Step 4: Confirm tests pass.
- [ ] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/public/LandingPage.test.tsx`
