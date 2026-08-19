---
id: TASK-623
title: Custom Branded Domains, Live DNS Checker & SSL Manager
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Frontend Engineer
depends_on:
  - tasks/06_frontend_foundation/task_604_ui_component_primitives.md
references:
  - api/openapi_v2_growth.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - custom-domains
  - dns
  - ssl
  - cname
---

# TASK-623: Custom Branded Domains, Live DNS Checker & SSL Manager

## 1. Goal
Build the Custom Branded Domains page (`/domains`) with domain setup modal, real-time DNS CNAME/TXT verification badge, ACME SSL certificate status tracker, and custom root domain redirect settings.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/growth/DomainsPage.tsx`
- `apps/frontend/src/components/domains/DomainSetupModal.tsx`
- `apps/frontend/src/components/domains/DNSVerificationCard.tsx`
- `apps/frontend/src/components/domains/SSLStatusBadge.tsx`

## 3. Dependencies & Prerequisites
- `@flux/openapi` (`createDomain`, `ZCustomDomain`)

## 4. Referenced Architecture & Product Specs
- [docs/growth/custom_domains_ssl.md](file:///home/logan78/Desktop/flux/docs/growth/custom_domains_ssl.md)
- [tasks/02_growth/task_203_custom_domains.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_203_custom_domains.md)

## 5. Acceptance Criteria
- [ ] Registration modal generates verification CNAME and TXT challenge records for copy-pasting into Cloudflare/Route53.
- [ ] Verify DNS button triggers backend lookup and updates verification status in real time.
- [ ] SSL certificate status pill displays `active` (green), `pending` (amber), or `expired` (red).

## 6. Target Deliverables
- `apps/frontend/src/pages/growth/DomainsPage.tsx`
- `apps/frontend/src/components/domains/DomainSetupModal.tsx`

## 7. Definition of Done (DoD)
- [ ] Step 1: Write DNS record helper unit tests.
- [ ] Step 2: Confirm test failure.
- [ ] Step 3: Implement Domains page.
- [ ] Step 4: Confirm tests pass.
- [ ] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/growth/DomainsPage.test.tsx`
