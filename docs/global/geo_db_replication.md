# Subsystem Specification: Geo Db Replication

> **Source**: Extracted from `prompt/5/02.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Global Scale)**.
> You MUST implement **Part 5 (Flux Global Scale) — File 02.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART V — FLUX GLOBAL SCALE
## CHAPTER 2 — GEO-DISTRIBUTED DATABASE REPLICATION & EDGE CACHE SYNCHRONIZATION

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART V — Flux Global Scale: Chapter 2 — Geo-Distributed Database Replication & Edge Cache Synchronization** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Implement a distributed active-active database replication architecture (CockroachDB / PlanetScale) and Redis Enterprise active-active geo-replication across US, Europe, and Asia-Pacific regions.

Key Features:
- Multi-Region Database Replication with automatic failover.
- Active-Active Redis Geo-Replication with conflict-free replicated data types (CRDTs).
- Regional Cache Invalidation Bus.

2. Architecture & Data Replication Matrix

     US-East Database          EU-West Database          AP-East Database
    (CockroachDB Node)        (CockroachDB Node)        (CockroachDB Node)
           ▲                         ▲                         ▲
           │◄────── (Raft Consensus Multi-Master Replication) ──►│
           ▼                         ▼                         ▼
      US-East Redis             EU-West Redis             AP-East Redis
     (Active-Active)           (Active-Active)           (Active-Active)

3. Implementation Step-by-Step

Step 1: Configure CockroachDB / PlanetScale multi-region cluster topologies with regional locality rules (`LOCALITY "region=us-east"`).
Step 2: Set up Redis Enterprise Active-Active CRDT replication for click counters.
Step 3: Implement database failover health probes verifying 99.999% availability during regional outages.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/plan/docs/ARCHITECTURE.md) |
| **Previous** | [docs/global/edge_redirect_workers.md](file:///home/logan78/Desktop/plan/docs/global/edge_redirect_workers.md) |
| **Next** | [docs/global/anycast_dns_tls.md](file:///home/logan78/Desktop/plan/docs/global/anycast_dns_tls.md) |
| **Children** | [task_501_edge_redirects.md](file:///home/logan78/Desktop/plan/tasks/05_global/task_501_edge_redirects.md), [task_502_geo_replication.md](file:///home/logan78/Desktop/plan/tasks/05_global/task_502_geo_replication.md), [task_503_anycast_dns.md](file:///home/logan78/Desktop/plan/tasks/05_global/task_503_anycast_dns.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/plan/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
