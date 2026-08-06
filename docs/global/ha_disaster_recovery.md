# Subsystem Specification: Ha Disaster Recovery

> **Source**: Extracted from `prompt/5/05.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Global Scale)**.
> You MUST implement **Part 5 (Flux Global Scale) — File 05.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART V — FLUX GLOBAL SCALE
## CHAPTER 5 — GLOBAL HIGH AVAILABILITY, CIRCUIT BREAKERS & DISASTER RECOVERY

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART V — Flux Global Scale: Chapter 5 — Global High Availability, Circuit Breakers & Disaster Recovery** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Implement global High Availability (HA), Circuit Breakers, Automated Regional Failover, and Chaos Engineering protocols to guarantee 99.999% platform uptime.

Key Features:
- Circuit Breaker pattern (Hystrix / Gobreaker) preventing cascading database failures.
- Automated Health Probes evacuating unhealthy regions in <5 seconds.
- Multi-Region Disaster Recovery (DR) testing and automated point-in-time database snapshotting.

2. Implementation Step-by-Step

Step 1: Implement Circuit Breaker middleware around external database and API calls (Tripping on 50% error rate over 10s window).
Step 2: Build global DNS health checker (Route53 / Cloudflare Traffic Manager) rerouting Anycast traffic away from degraded regions.
Step 3: Conduct automated chaos engineering tests validating system resilience during simulated region failure.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/plan/docs/ARCHITECTURE.md) |
| **Previous** | [docs/global/anycast_dns_tls.md](file:///home/logan78/Desktop/plan/docs/global/anycast_dns_tls.md) |
| **Next** | [tasks/01_core/task_100_bootstrap_backend.md](file:///home/logan78/Desktop/plan/tasks/01_core/task_100_bootstrap_backend.md) |
| **Children** | [task_501_edge_redirects.md](file:///home/logan78/Desktop/plan/tasks/05_global/task_501_edge_redirects.md), [task_502_geo_replication.md](file:///home/logan78/Desktop/plan/tasks/05_global/task_502_geo_replication.md), [task_503_anycast_dns.md](file:///home/logan78/Desktop/plan/tasks/05_global/task_503_anycast_dns.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/plan/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
