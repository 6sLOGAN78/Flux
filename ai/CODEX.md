# CODEX.md — OpenAI Codex CLI Agent Adapter

> **Master Protocol Directive**: OpenAI Codex CLI agents MUST follow [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md).

## Operational Instructions
- **Context Isolation**: Load `AGENTS.md` + active task document + target schemas.
- **Single Responsibility**: Ensure every file answers exactly one question.
- **No Symptom Masking**: Fix root causes; never swallow exceptions or comment out broken assertions.
- **Definition of Done**: Unit tests pass 100% and linter returns 0 warnings.
