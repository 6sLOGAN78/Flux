---
id: TASK-600
title: Frontend Environment Loader, Vite Config & Design Tokens
layer: Level 5 (Executable Task Unit)
status: Done
owner: Frontend Engineer
depends_on:
  - packages/zod/src/index.ts
references:
  - docs/ARCHITECTURE.md#frontend-libraries--styling-appsfrontend
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - frontend
  - vite
  - env
  - tailwind
  - zod
---

# TASK-600: Frontend Environment Loader, Vite Config & Design Tokens

## 1. Goal
Configure type-safe environment variable parsing (`apps/frontend/src/config/env.ts`) using Zod, create `.env.example`, configure Tailwind CSS v4 design tokens, and verify Vite build pipeline.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/config/env.ts`
- `apps/frontend/.env.example`
- `apps/frontend/src/index.css`
- `apps/frontend/vite.config.ts`

## 3. Dependencies & Prerequisites
- `@flux/zod` in `packages/zod`
- Vite and Tailwind CSS v4 in `apps/frontend/package.json`

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md)
- Design Taste Skill: `.agents/skills/design-taste-frontend/SKILL.md`

## 5. Acceptance Criteria
- [x] Zod schema validates `VITE_API_URL`, `VITE_CLERK_PUBLISHABLE_KEY`, and `VITE_APP_URL` on app startup with clear terminal errors if missing.
- [x] Tailwind CSS v4 color tokens and dark-first palette are declared in `index.css`.
- [x] `bun run build` in `apps/frontend` executes without TypeScript or CSS errors.

## 6. Target Deliverables
- `apps/frontend/src/config/env.ts`
- `apps/frontend/.env.example`
- `apps/frontend/src/index.css`

## 7. Definition of Done (DoD)
- [x] Step 1: Write unit test validating `env.ts` schema with valid and invalid values.
- [x] Step 2: Confirm test failure.
- [x] Step 3: Implement `env.ts` and `.env.example`.
- [x] Step 4: Confirm test passes with `bun test`.
- [x] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Unit Test**: `apps/frontend/src/config/env.test.ts`
- **Verification Command**: `cd apps/frontend && bun test src/config/env.test.ts`
