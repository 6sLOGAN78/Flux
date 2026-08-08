# TESTING.md — Testing Strategy & TDD Guidelines

## 1. Test Requirements
- Implementation code MUST NOT be written until a failing unit test is written and verified to fail (`Red`).
- Unit tests run <10ms per test (`go test -v ./...`).
- Benchmarks required for Base62 encoding and Redirect handlers (`go test -bench=.`).

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) |
| **Previous** | [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md) |
| **Next** | [ai/DEPLOYMENT.md](file:///home/logan78/Desktop/flux/ai/DEPLOYMENT.md) |
| **Children** | None |
| **Dependencies** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
