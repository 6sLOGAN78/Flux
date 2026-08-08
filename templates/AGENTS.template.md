---
id: TEMPLATE-AGENTS
title: AI Agent Operational Protocol & Entrypoint
layer: System / Protocol
status: Active
owner: Principal AI Architect
references:
  - INDEX.md
  - ARCHITECTURE.md
---

# [Document Name] — AI Agent Operating Protocol

## Purpose
Define the mandatory operating guidelines, execution steps, tool usage rules, safety boundaries, and precedence hierarchy for AI Coding Agents working in this repository.

## Scope
Applies to all automated agents, subagents, and AI assistant sessions operating on the codebase.

## Sections
- **1. Agent Identity & Role Definition**: Core persona and responsibility constraints.
- **2. Master Precedence Hierarchy**: Strict 8-tier conflict resolution order.
- **3. Task Execution Protocol**: Step-by-step workflow (Read Task ➔ Load Schemas ➔ TDD Cycle ➔ Verify).
- **4. Safety & Operating Boundaries**: Prohibited actions, file boundaries, commit rules.
- **5. Context Window Optimization**: Context isolation strategies to minimize token waste.
- **6. Verification & Definition of Done**: Verification commands required before declaring task completion.

## Cross References
- [Master Sitemap](file:///home/logan78/Desktop/flux/docs/INDEX.md)
- [System Architecture](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md)
- [Tasks Registry](file:///home/logan78/Desktop/flux/tasks/)

## Acceptance Criteria
- [ ] Agent can resolve any requirement conflict using Section 2 precedence rules.
- [ ] Agent follows TDD cycle without skipping failing test verification steps.
- [ ] Zero unverified code claims permitted.

## Navigation
[Top](#purpose) | [Precedence](#sections) | [Verification](#acceptance-criteria)
