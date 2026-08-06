# INDEX.md — Master Documentation Sitemap

## Root Specifications
- [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) — AI Agent Operating Protocol
- [PRODUCT.md](file:///home/logan78/Desktop/flux/PRODUCT.md) — Product Vision & Matrix
- [ROADMAP.md](file:///home/logan78/Desktop/flux/ROADMAP.md) — Tier Evolution Roadmap
- [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) — Protask Monorepo System Architecture
- [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) — Security & Auth
- [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md) — Performance SLAs
- [ai/DEPLOYMENT.md](file:///home/logan78/Desktop/flux/ai/DEPLOYMENT.md) — Production Deployment

## Applications & Monorepo Packages
- [`apps/backend/`](file:///home/logan78/Desktop/flux/apps/backend) — Go 1.24 Echo REST API & Asynq Worker
- [`apps/frontend/`](file:///home/logan78/Desktop/flux/apps/frontend) — Vite React 19 Client App
- [`packages/zod/`](file:///home/logan78/Desktop/flux/packages/zod) — Shared Zod Schemas & Inferred TS Types
- [`packages/openapi/`](file:///home/logan78/Desktop/flux/packages/openapi) — `@ts-rest` Contracts & OpenAPI Spec Generator
- [`packages/emails/`](file:///home/logan78/Desktop/flux/packages/emails) — React Email Transactional Templates

## Subsystem Specifications (`docs/`)
- [Link Management](file:///home/logan78/Desktop/flux/docs/core/link_management.md)
- [Redirect Engine](file:///home/logan78/Desktop/flux/docs/core/redirect_engine.md)
- [Analytics Pipeline](file:///home/logan78/Desktop/flux/docs/core/analytics_pipeline.md)
- [QR Service](file:///home/logan78/Desktop/flux/docs/core/qr_service.md)
- [Multi-Tenant RBAC](file:///home/logan78/Desktop/flux/docs/saas/multi_tenant_rbac.md)
- [Attribution Engine](file:///home/logan78/Desktop/flux/docs/enterprise/attribution_engine.md)
- [Edge Redirect Workers](file:///home/logan78/Desktop/flux/docs/global/edge_redirect_workers.md)

## 🏷️ AI Agent RAG Tag Lookup Matrix
| Tag / Keyword | Target Task File | Primary Specification |
| :--- | :--- | :--- |
| `bootstrap`, `echo` | [TASK-100](file:///home/logan78/Desktop/flux/tasks/01_core/task_100_bootstrap_backend.md) | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| `base62` | [TASK-102](file:///home/logan78/Desktop/flux/tasks/01_core/task_102_base62_encoder.md) | [link_management.md](file:///home/logan78/Desktop/flux/docs/core/link_management.md) |
| `jwt`, `auth`, `clerk` | [TASK-101](file:///home/logan78/Desktop/flux/tasks/01_core/task_101_auth_service.md) | [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| `redirect`, `redis` | [TASK-103](file:///home/logan78/Desktop/flux/tasks/01_core/task_103_redirect_handler.md) | [redirect_engine.md](file:///home/logan78/Desktop/flux/docs/core/redirect_engine.md) |
| `clickhouse` | [TASK-201](file:///home/logan78/Desktop/flux/tasks/02_growth/task_201_clickhouse_pipeline.md) | [time_series_analytics.md](file:///home/logan78/Desktop/flux/docs/growth/time_series_analytics.md) |

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | None |
| **Previous** | None |
| **Next** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
| **Children** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md), [PRODUCT.md](file:///home/logan78/Desktop/flux/PRODUCT.md), [ROADMAP.md](file:///home/logan78/Desktop/flux/ROADMAP.md), [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md), [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/CODING_STANDARDS.md](file:///home/logan78/Desktop/flux/ai/CODING_STANDARDS.md), [ai/TESTING.md](file:///home/logan78/Desktop/flux/ai/TESTING.md), [ai/DEPLOYMENT.md](file:///home/logan78/Desktop/flux/ai/DEPLOYMENT.md), [CONTRIBUTING.md](file:///home/logan78/Desktop/flux/CONTRIBUTING.md), [CHANGELOG.md](file:///home/logan78/Desktop/flux/CHANGELOG.md) |
| **Dependencies** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
