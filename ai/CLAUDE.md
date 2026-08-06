# CLAUDE.md — Claude Code CLI Agent Adapter

> **Master Protocol Directive**: All Claude Code sessions MUST follow the rules, precedence hierarchy, and TDD workflow defined in [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md).

## Operational Instructions
- **Context Isolation**: Load only `AGENTS.md` + active task file from `tasks/` + target schema from `api/` or `database/`.
- **Precedence Hierarchy**: `AGENTS.md` ➔ Task Spec ➔ `PRODUCT.md` ➔ `ARCHITECTURE.md` ➔ `SECURITY/PERFORMANCE` ➔ Schemas.
- **TDD Requirement**: Write failing test first ➔ Confirm failure ➔ Write minimal code ➔ Pass test ➔ Run linter.
- **Verification Command**: Execute exact test command specified in task document.
