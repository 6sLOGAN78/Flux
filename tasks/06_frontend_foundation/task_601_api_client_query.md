---
id: TASK-601
title: Type-Safe ts-rest API Client & TanStack Query Hooks
layer: Level 5 (Executable Task Unit)
status: Done
owner: Frontend Engineer
depends_on:
  - packages/openapi/src/contracts/index.ts
  - tasks/06_frontend_foundation/task_600_frontend_scaffold_env.md
references:
  - packages/openapi/src/contracts/index.ts
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - api-client
  - ts-rest
  - react-query
  - hooks
---

# TASK-601: Type-Safe ts-rest API Client & TanStack Query Hooks

## 1. Goal
Initialize `@ts-rest/core` client with dynamic Authorization Bearer token injection (from Clerk) and encapsulate queries/mutations into idiomatic TanStack React Query v5 hooks.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/api/client.ts`
- `apps/frontend/src/api/queryClient.ts`
- `apps/frontend/src/hooks/useLinksQuery.ts`
- `apps/frontend/src/hooks/useAnalyticsQuery.ts`

## 3. Dependencies & Prerequisites
- `@flux/openapi` contract routes
- `@tanstack/react-query` v5

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md)
- [docs/ARCHITECTURE.md#contract-first-api--schema-pipeline](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md#contract-first-api--schema-pipeline)

## 5. Acceptance Criteria
- [x] API client attaches `Authorization: Bearer <token>` automatically when an auth token is present in the React session.
- [x] React Query hooks expose `useGetLinks`, `useCreateLink`, `useGetAnalyticsSummary`, and `useLinkMetrics`.
- [x] Cache invalidation occurs automatically on mutations (`createLink`, `updateLink`, `createCategory`).

## 6. Target Deliverables
- `apps/frontend/src/api/client.ts`
- `apps/frontend/src/api/queryClient.ts`
- `apps/frontend/src/hooks/useLinksQuery.ts`

## 7. Definition of Done (DoD)
- [x] Step 1: Write unit tests verifying client request construction and headers.
- [x] Step 2: Confirm test failure.
- [x] Step 3: Implement client and query hooks.
- [x] Step 4: Confirm test passes with `bun test`.
- [x] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/api/client.test.ts`
