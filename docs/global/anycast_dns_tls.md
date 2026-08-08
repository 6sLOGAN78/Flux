# Subsystem Specification: Anycast Dns Tls

> **Source**: Extracted from `prompt/5/03.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Global Scale)**.
> You MUST implement **Part 5 (Flux Global Scale) — File 03.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART V — FLUX GLOBAL SCALE
## CHAPTER 3 — ANYCAST DNS INFRASTRUCTURE & AUTOMATED EDGE TLS TERMINATION

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART V — Flux Global Scale: Chapter 3 — Anycast DNS Infrastructure & Automated Edge TLS Termination** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Build an Anycast DNS routing setup combined with dynamic Caddy / Envoy reverse proxy mesh to terminate TLS at global edge nodes for custom user domains.

Key Features:
- Anycast IP DNS routing directing traffic to nearest edge reverse proxy.
- Automated Edge ACME TLS certificate issuance and distribution.
- Zero-downtime certificate renewal and TLS 1.3 optimization.

2. Implementation Step-by-Step

Step 1: Configure Anycast BGP routing across primary edge data centers.
Step 2: Deploy Envoy / Caddy reverse proxy mesh with shared Redis ACME certificate store.
Step 3: Implement dynamic SNI callback querying active custom domain registry in <2ms.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Previous** | [docs/global/geo_db_replication.md](file:///home/logan78/Desktop/flux/docs/global/geo_db_replication.md) |
| **Next** | [docs/global/ha_disaster_recovery.md](file:///home/logan78/Desktop/flux/docs/global/ha_disaster_recovery.md) |
| **Children** | [task_501_edge_redirects.md](file:///home/logan78/Desktop/flux/tasks/05_global/task_501_edge_redirects.md), [task_502_geo_replication.md](file:///home/logan78/Desktop/flux/tasks/05_global/task_502_geo_replication.md), [task_503_anycast_dns.md](file:///home/logan78/Desktop/flux/tasks/05_global/task_503_anycast_dns.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
