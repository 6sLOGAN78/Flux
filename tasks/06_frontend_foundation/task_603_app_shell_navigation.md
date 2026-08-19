---
id: TASK-603
title: Minimalist App Shell, Clean Sidebar, cmdk & Workspace Switcher
layer: Level 5 (Executable Task Unit)
status: Completed
owner: Frontend Engineer
depends_on:
  - tasks/06_frontend_foundation/task_602_auth_clerk_sso.md
references:
  - PRODUCT.md
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - app-shell
  - sidebar
  - cmdk
  - navigation
  - layout
  - minimalist
---

# TASK-603: Minimalist App Shell, Clean Sidebar, cmdk & Workspace Switcher

## 1. Goal
Build the Notion & Dub-style dashboard shell (`AppLayout`): clean flat sidebar with hairline border (`border-r border-zinc-200 dark:border-zinc-800`), minimalist header with breadcrumbs, tenant Workspace Switcher, and keyboard command palette (`Cmd+K` / `Ctrl+K`).

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/components/layout/AppLayout.tsx`
- `apps/frontend/src/components/layout/Sidebar.tsx`
- `apps/frontend/src/components/layout/Header.tsx`
- `apps/frontend/src/components/layout/WorkspaceSwitcher.tsx`
- `apps/frontend/src/components/layout/CommandPalette.tsx`

## 3. Dependencies & Prerequisites
- `lucide-react`
- `react-router-dom` v7

## 4. Referenced Architecture & Product Specs
- Design Taste Skill: `.agents/skills/design-taste-frontend/SKILL.md`
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md)
- [PRODUCT.md](file:///home/logan78/Desktop/flux/PRODUCT.md)

## 5. Acceptance Criteria
- [x] Sidebar: Minimalist flat navigation items (`hover:bg-zinc-100 dark:hover:bg-zinc-900`), active item indicator, collapsible on mobile (`<768px`).
- [x] Command palette (`Cmd+K`): Fast fuzzy search with clean search input and subtle shortcut key hints (`⌘K`, `ESC`).
- [x] Workspace switcher: Dub-style clean dropdown with organization logos/monograms and instant workspace switching.

## 6. Target Deliverables
- `apps/frontend/src/components/layout/AppLayout.tsx`
- `apps/frontend/src/components/layout/Sidebar.tsx`
- `apps/frontend/src/components/layout/CommandPalette.tsx`

## 7. Definition of Done (DoD)
- [x] Step 1: Write component unit test for Sidebar navigation and Command Palette shortcut.
- [x] Step 2: Confirm test failure.
- [x] Step 3: Implement layout components.
- [x] Step 4: Confirm tests pass with `bun test`.
- [x] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/components/layout/AppLayout.test.tsx`
