# ANTIGRAVITY.md — Google Antigravity Agent Adapter

> **Master Protocol Directive**: All Google Antigravity AI agents MUST follow the rules, precedence hierarchy, and TDD workflow defined in [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md).

## Operational Instructions
- **Task Scope Lock**: Ingest active task file from `tasks/` and verify inputs, outputs, and target file boundaries.
- **Empirical Rigor**: Base diagnoses strictly on full runtime error logs and terminal execution output.
- **Strict Error Wrapping**: Every internal error must be contextually wrapped (`fmt.Errorf("...: %w", err)`).
- **Verification Command**: Run test and static analysis commands before declaring completion.
