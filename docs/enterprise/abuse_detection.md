# Subsystem Specification: Abuse Detection

> **Source**: Extracted from `prompt/4/07.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Enterprise)**.
> You MUST implement **Part 4 (Flux Enterprise) — File 07.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART IV — FLUX ENTERPRISE
## CHAPTER 7 — ENTERPRISE ADMIN, ABUSE DETECTION & AUTOMATED MALWARE SCANNING

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART IV — Flux Enterprise: Chapter 7 — Enterprise Admin, Abuse Detection & Automated Malware Scanning** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Build an automated security moderation and anti-abuse platform to protect Flux from hosting phishing, malware, or illegal short URL destinations.

Key Features:
- Real-time Destination URL Scanning via Google Safe Browsing API & VirusTotal API.
- Automated Abuse Detection Engine (flagging high-frequency link creation or spam patterns).
- Administrative Quarantine & Link Takedown Workflow.

2. Database Schema Design

```sql
CREATE TABLE IF NOT EXISTS security_scans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    is_safe BOOLEAN DEFAULT TRUE,
    threat_type VARCHAR(100), -- 'malware', 'phishing', 'social_engineering'
    threat_provider VARCHAR(50), -- 'google_safe_browsing', 'virustotal'
    scanned_at TIMESTAMPTZ DEFAULT NOW()
);
```

3. Implementation Step-by-Step

Step 1: Build URL Safety Verification Worker calling Google Safe Browsing v4 API before new links are activated.
Step 2: Implement automatic quarantine handler (`status = 'quarantined'`) returning 451 Unavailable For Legal Reasons on flagged links.
Step 3: Build Admin Security Portal with one-click link restoration or global domain ban triggers.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Previous** | [docs/enterprise/white_label_engine.md](file:///home/logan78/Desktop/flux/docs/enterprise/white_label_engine.md) |
| **Next** | [docs/global/edge_redirect_workers.md](file:///home/logan78/Desktop/flux/docs/global/edge_redirect_workers.md) |
| **Children** | [task_401_attribution_engine.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_401_attribution_engine.md), [task_402_funnel_analytics.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_402_funnel_analytics.md), [task_403_revenue_analytics.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_403_revenue_analytics.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
