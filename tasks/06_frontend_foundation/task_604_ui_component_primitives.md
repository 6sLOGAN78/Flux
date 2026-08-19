---
id: TASK-604
title: Notion & Dub Minimalist UI Primitives & Form Controls
layer: Level 5 (Executable Task Unit)
status: Completed
owner: Frontend Engineer
depends_on:
  - tasks/06_frontend_foundation/task_600_frontend_scaffold_env.md
references:
  - PRODUCT.md
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - ui
  - design-system
  - minimalist
  - button
  - modal
  - table
  - notion-dub
---

# TASK-604: Notion & Dub Minimalist UI Primitives & Form Controls

## 1. Goal
Implement accessible, reusable UI component primitives adhering to the **Notion & Dub.co minimalist design language**: clean monochrome zinc base, subtle hairline borders (`border-zinc-200` / `border-zinc-800`), simple solid black/white buttons (`bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900`), and single purposeful functional accent for active states.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/components/ui/Button.tsx`
- `apps/frontend/src/components/ui/Input.tsx`
- `apps/frontend/src/components/ui/Modal.tsx`
- `apps/frontend/src/components/ui/DataTable.tsx`
- `apps/frontend/src/components/ui/Badge.tsx`
- `apps/frontend/src/components/ui/Tabs.tsx`

## 3. Dependencies & Prerequisites
- `clsx` and `tailwind-merge` in `apps/frontend/package.json`

## 4. Referenced Architecture & Product Specs
- `.agents/skills/design-taste-frontend/SKILL.md`
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md)
- [PRODUCT.md](file:///home/logan78/Desktop/flux/PRODUCT.md)

## 5. Acceptance Criteria
- [x] Buttons:
  - Primary: Solid black `bg-zinc-900 text-white hover:bg-zinc-800 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200`.
  - Secondary/Outline: Subtle border `border-zinc-200 hover:bg-zinc-100 dark:border-zinc-800 dark:hover:bg-zinc-900`.
  - Single-line labels with tactile active states (`scale-[0.98]`).
- [x] Inputs & Controls: Clean flat inputs with `border-zinc-200 dark:border-zinc-800`, subtle focus ring (`ring-zinc-900/10 dark:ring-zinc-100/10`), label on top, error message below.
- [x] Modal: Lightweight backdrop blur (`backdrop-blur-xs`), subtle border, focus trapped, clean `Escape` close.
- [x] Badges: Minimalist pill badges with subtle zinc/emerald/blue backgrounds (no loud neon gradients).

## 6. Target Deliverables
- `apps/frontend/src/components/ui/` primitive component library

## 7. Definition of Done (DoD)
- [x] Step 1: Write component unit tests for Button, Input, Modal, and DataTable.
- [x] Step 2: Confirm test failure.
- [x] Step 3: Implement Notion & Dub minimalist UI primitives.
- [x] Step 4: Confirm tests pass with `bun test`.
- [x] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/components/ui/Button.test.tsx`
