# Subsystem Specification: Saml Scim Sso

> **Source**: Extracted from `prompt/4/05.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Enterprise)**.
> You MUST implement **Part 4 (Flux Enterprise) — File 05.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART IV — FLUX ENTERPRISE
## CHAPTER 5 — ENTERPRISE SECURITY, SAML 2.0 / OIDC SSO, SCIM 2.0 & IP ALLOWLISTS

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART IV — Flux Enterprise: Chapter 5 — Enterprise Security, SAML 2.0 / OIDC SSO, SCIM 2.0 & IP Allowlists** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Implement enterprise security standards: SAML 2.0 / OIDC Single Sign-On (SSO), SCIM 2.0 User Provisioning, IP Address Allowlists, and SOC2 Audit Trail compliance.

Key Features:
- SAML 2.0 & OIDC SSO Integration (Okta, Azure AD, PingIdentity, OneLogin).
- SCIM 2.0 Protocol Implementation (`/scim/v2/Users`, `/scim/v2/Groups`).
- IP Allowlist Enforcement per Organization (CIDR block restrictions).

2. Database Schema Design

```sql
CREATE TABLE IF NOT EXISTS sso_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID UNIQUE NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    idp_type VARCHAR(30) NOT NULL, -- 'saml', 'oidc'
    entity_id TEXT NOT NULL,
    sso_url TEXT NOT NULL,
    certificate TEXT NOT NULL,
    enforce_sso BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ip_allowlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    cidr_block VARCHAR(45) NOT NULL, -- e.g. "192.168.1.0/24"
    description VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

3. Implementation Step-by-Step

Step 1: Integrate SAML Service Provider library (e.g. `crewjam/saml` in Go) handling ACS URL and SP metadata XML.
Step 2: Implement SCIM 2.0 REST API supporting automated user creation, updates, and de-provisioning from Okta/Azure.
Step 3: Build IP Allowlist Interceptor checking incoming client IP against parsed CIDR network trees (`net.IPNet`).


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Previous** | [docs/enterprise/ai_engine.md](file:///home/logan78/Desktop/flux/docs/enterprise/ai_engine.md) |
| **Next** | [docs/enterprise/white_label_engine.md](file:///home/logan78/Desktop/flux/docs/enterprise/white_label_engine.md) |
| **Children** | [task_401_attribution_engine.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_401_attribution_engine.md), [task_402_funnel_analytics.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_402_funnel_analytics.md), [task_403_revenue_analytics.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_403_revenue_analytics.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
