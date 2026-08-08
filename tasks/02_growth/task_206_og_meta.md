---id: TASK-206
title: Dynamic Open Graph Scraping & Social Preview
layer: Level 5 (Executable Task Unit)
status: Done
owner: Feature Engineer
depends_on:
  - docs/growth/dynamic_og_metadata.md
references:
  - [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md)
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - og-meta
  - social-preview
  - scraper
---

# TASK-206: Dynamic Open Graph Scraping & Social Preview

## 1. Goal
Implement OG tag scraper and social crawler meta preview HTML renderer.

## 2. Scope & Target Boundaries
Target source files modified during this task:
- `internal/modules/ogmeta/`

## 3. Dependencies & Prerequisites
This task relies on specifications and schema contracts in:
  - docs/growth/dynamic_og_metadata.md

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md)
  - [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md)

## 5. Acceptance Criteria
- [x] Detects social bots (Slack, Twitter) and renders dynamic HTML meta preview tags.
- [x] Code follows all formatting and linting rules in [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md).
- [x] Latency and memory usage adhere to targets in [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md).

## 6. Target Deliverables
- `internal/modules/ogmeta/scraper.go`

## 7. Definition of Done (DoD)
- [x] **Step 1: Write failing unit test** verifying expected feature behavior.
- [x] **Step 2: Confirm test failure**: Run `go test -v ./internal/modules/ogmeta/...`.
- [x] **Step 3: Write minimal implementation** inside target files.
- [x] **Step 4: Confirm test passes**: Run `go test -v ./internal/modules/ogmeta/...`.
- [x] **Step 5: Run linter**: Confirm zero warnings or errors.

## 8. Testing Strategy
- **Unit Tests**: Test pure domain logic in isolation.
- **Verification Command**: `go test -v ./internal/modules/ogmeta/...`

## 9. Documentation Updates
- [x] Mark task status as `Done` in this document frontmatter upon completion.
- [x] Update function/package docstrings if signatures were extended.


<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/growth/dynamic_og_metadata.md](file:///home/logan78/Desktop/flux/docs/growth/dynamic_og_metadata.md) |
| **Previous** | [tasks/02_growth/task_205_deep_linking.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_205_deep_linking.md) |
| **Next** | [tasks/02_growth/task_207_advanced_qr.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_207_advanced_qr.md) |
| **Children** | None |
| **Dependencies** | [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql), [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md), [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md), [ai/TESTING.md](file:///home/logan78/Desktop/flux/ai/TESTING.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
