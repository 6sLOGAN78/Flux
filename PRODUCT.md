# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users
- Primary: Engineering, Growth, and Marketing teams at tech companies and developer-first startups needing sub-10ms URL redirection, programmable dynamic routing, and deep analytics.
- Secondary: Enterprise marketing leaders and data analysts tracking multi-touch campaign attribution, conversion funnels, and revenue ROI.

## Product Purpose
Flux is a production-grade, 5-tier link infrastructure and analytics platform that turns links into programmable marketing and attribution primitives. Success means providing sub-10ms global edge redirects, granular real-time ClickHouse analytics, seamless team collaboration, and transparent attribution reporting.

## Positioning
Unlike legacy link shorteners that act as simple static redirectors, Flux provides high-performance edge-routed URL infrastructure with programmatic A/B traffic splitting, geo/device dynamic routing, deep-link app attribution, ClickHouse time-series exploration, and multi-touch revenue models (Linear, First/Last, Time-Decay, U-Shaped).

## Operating Context
- Web application accessed by desktop browsers with command-palette (`cmdk`) keyboard navigation.
- Programmatic consumption via REST API (`/api/v1`), OAuth 2.0 applications, and event-driven outbound webhooks.
- Multi-region global edge workers (Cloudflare Workers / Anycast DNS) handling click traffic with sub-10ms latency.

## Capabilities and Constraints
- **Tier 1 (Core)**: Base62 shortening, high-performance redirects, QR code generator with custom logos, rate limiting.
- **Tier 2 (Growth)**: Campaign management, visual UTM builder, custom branded domains & automated SSL, smart geo/device routing, deep linking, dynamic OG metadata, A/B testing.
- **Tier 3 (SaaS)**: Multi-tenant organizations & workspace RBAC, Stripe subscription billing, public REST API, OAuth 2.0, Webhook engine, in-app notifications.
- **Tier 4 (Enterprise)**: Multi-touch attribution engine (5 models), conversion funnels, revenue ROAS/LTV analytics, predictive AI anomaly engine, SAML 2.0 / SCIM 2.0 SSO, malware scanning.
- **Tier 5 (Global Scale)**: Cloudflare Workers edge redirect engine, geo-replicated databases (<500ms sync SLA), Anycast BGP DNS, automated multi-region failover.
- **Constraints**:
  - Flux is NOT a general-purpose CMS or website builder.
  - Raw un-aggregated click analytics logs are stored in ClickHouse, never in the OLTP Postgres database.

## Brand Commitments
- **Name**: Flux
- **Design Aesthetic**: Notion & Dub.co-inspired minimalist design language. Clean neutral monochrome base (zinc/gray palette), subtle hairline borders (`border-zinc-200` / `border-zinc-800`), minimal color saturation, high-contrast typography, and purposeful simple buttons (solid black/white primary CTAs, subtle functional accent for active states & badges) with zero visual noise.
- **Voice**: Precise, direct, technical, reliable, and fast.

## Evidence on Hand
- Complete Go Echo backend implementation in `apps/backend` with 100% passing test suite across all 5 tiers.
- End-to-end type-safe API contracts in `packages/openapi` (`@flux/openapi`) and Zod schemas in `packages/zod` (`@flux/zod`).
- Production ClickHouse schema in `database/clickhouse_analytics_schema.sql` and Postgres migrations in `apps/backend/internal/database/migrations/`.

## Product Principles
1. **Speed & Low Latency First**: Edge redirects and UI state transitions must feel instant (<10ms edge redirects, optimistic UI mutations).
2. **Developer Ergonomics**: Everything in the UI is backed by an API contract, keyboard shortcuts, and type safety.
3. **Data Integrity & Attribution Rigor**: Analytics and attribution calculations are mathematically sound and verifiable.
4. **Resilient Multi-Tenancy**: Strict tenant and workspace data isolation across all operational tiers.

## Accessibility & Inclusion
- WCAG AA contrast compliance across all dark and light surfaces.
- Full keyboard navigability with command palette and focus ring indicators.
- Respect `prefers-reduced-motion` and `prefers-reduced-transparency`.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) |
| **Previous** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
| **Next** | [ROADMAP.md](file:///home/logan78/Desktop/flux/ROADMAP.md) |
| **Children** | [ROADMAP.md](file:///home/logan78/Desktop/flux/ROADMAP.md), [docs/core/link_management.md](file:///home/logan78/Desktop/flux/docs/core/link_management.md) |
| **Dependencies** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
