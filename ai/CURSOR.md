# CURSOR.md — Cursor IDE Agent Adapter

> **Master Protocol Directive**: All Cursor IDE agent sessions MUST follow the rules, precedence hierarchy, and TDD workflow defined in [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md).

## Operational Instructions
- **Context Isolation**: Ingest only the active task file from `tasks/` and target schema files.
- **Rules Enforcement**: Enforce all code guidelines in [`ai/CODING_STANDARDS.md`](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md) and testing rules in [`ai/TESTING.md`](file:///home/logan78/Desktop/flux/ai/TESTING.md).
- **TDD Requirement**: Follow 5-step TDD cycle (`Red` ➔ `Green` ➔ `Verify` ➔ `Lint`).
