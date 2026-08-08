# PERFORMANCE.md — Flux SLAs, Latency Budgets & Caching Strategy

> **Source**: Extracted from `prompt/1/07.md`, `prompt/1/12.md`, `prompt/2/10.md`, `prompt/5/01.md`.

## 1. Latency Budgets & SLAs
- **Edge Redirect Latency Target**: <10ms global p99.
- **Core API Response Target**: <100ms p95.
- **Analytics Event Ingestion**: Sub-second click event availability in dashboard.

## 2. Caching Strategy
- **L1 (In-Memory)**: Go LRU cache / Worker memory.
- **L2 (Redis Cluster)**: Key `link:short_code` ➔ Destination URL + Routing Rules. TTL: 24 hours.
- **Cache Invalidation**: Redis Pub/Sub broadcast on link update or deletion.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) |
| **Previous** | [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Next** | [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md) |
| **Children** | None |
| **Dependencies** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
