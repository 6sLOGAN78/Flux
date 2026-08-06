# GEMINI.md — Gemini AI Agent Adapter

> **Master Protocol Directive**: All Gemini AI agent sessions MUST follow the rules, precedence hierarchy, and TDD workflow defined in [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md).

## Operational Instructions
- **Context Isolation**: Load only `AGENTS.md` + active task file from `tasks/` + target schemas from `api/` or `database/`.
- **Precedence Hierarchy**: `AGENTS.md` ➔ Task Spec ➔ `PRODUCT.md` ➔ `ARCHITECTURE.md` ➔ `ai/SECURITY.md` / `ai/PERFORMANCE.md` ➔ Schemas.
- **TDD Requirement**: Write failing test first ➔ Confirm failure ➔ Write minimal code ➔ Pass test ➔ Run linter.
- **Verification Command**: Execute exact test command specified in active task document.
