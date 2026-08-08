# Subsystem Specification: White Label Engine

> **Source**: Extracted from `prompt/4/06.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Enterprise)**.
> You MUST implement **Part 4 (Flux Enterprise) — File 06.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART IV — FLUX ENTERPRISE
## CHAPTER 6 — ENTERPRISE WHITE-LABEL ENGINE

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART IV — Flux Enterprise: Chapter 6 — Enterprise White-Label Engine** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Allow enterprise clients to fully white-label Flux under their own corporate identity, custom CSS/themes, custom dashboard subdomains, and white-label SSL certificates.

Key Features:
- Custom Dashboard Domain (e.g. `analytics.acme.com` serving the Flux Dashboard UI).
- White-label CSS theme & logo customization.
- Removal of all "Powered by Flux" branding footers.

2. Database Schema Design

```sql
CREATE TABLE IF NOT EXISTS white_label_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID UNIQUE NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    dashboard_domain VARCHAR(255) UNIQUE,
    brand_name VARCHAR(100) NOT NULL,
    logo_url TEXT,
    favicon_url TEXT,
    primary_color VARCHAR(10) DEFAULT '#0052FF',
    custom_css TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

3. Implementation Step-by-Step

Step 1: Write Custom Domain Routing Middleware serving white-labeled Next.js frontend assets.
Step 2: Implement Dynamic CSS Injection Engine serving tenant theme variables (`--primary-color`).
Step 3: Build Caddy ACME challenge handler for white-label dashboard SSL provisioning.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Previous** | [docs/enterprise/saml_scim_sso.md](file:///home/logan78/Desktop/flux/docs/enterprise/saml_scim_sso.md) |
| **Next** | [docs/enterprise/abuse_detection.md](file:///home/logan78/Desktop/flux/docs/enterprise/abuse_detection.md) |
| **Children** | [task_401_attribution_engine.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_401_attribution_engine.md), [task_402_funnel_analytics.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_402_funnel_analytics.md), [task_403_revenue_analytics.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_403_revenue_analytics.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
