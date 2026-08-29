# Feature Registry

| ID       | Feature                        | Description                                           | Status    | Phase | Frontend                 | Backend                        | DB/ML                | Tests | Notes                                                  |
|----------|--------------------------------|-------------------------------------------------------|-----------|-------|--------------------------|--------------------------------|----------------------|-------|--------------------------------------------------------|
| FEAT-001 | Link Shortening CRUD           | Create, read, update, delete short links              | WORKING   | 1     | `LinksListPage.tsx`      | `links.go`                     | Postgres (`links`)   | N/A   | Basic functionality works.                             |
| FEAT-002 | Edge Redirect                  | Redirect from short code to destination url           | PARTIAL   | 1     | N/A                      | `redirect.go`                  | Postgres (`links`)   | N/A   | Works via DB. Redis cache is `nil`.                    |
| FEAT-003 | Click Analytics (Summary)      | Aggregated platform performance summary               | BROKEN    | 2     | `AnalyticsPage.tsx`      | `analytics.go`                 | ClickHouse (Schema)  | N/A   | Endpoint exists but returns 500 (provider is `nil`).   |
| FEAT-004 | Event Ingestion                | Track clicks on redirect                              | MISSING   | 2     | N/A                      | `redirect.go`                  | ClickHouse (Schema)  | N/A   | Redirect handler does not publish click events.        |
| FEAT-005 | Redis Cache-Aside              | Sub-millisecond redirect caching                      | MISSING   | 1     | N/A                      | `repository/redirect.go`       | Redis                | N/A   | Code exists but is passed as `nil` in `server.go`.     |
| FEAT-006 | Custom Domains                 | CNAME/TXT verification & custom branding              | MISSING   | 3     | `DomainsPage.tsx` (mock) | `modules/domain/verifier.go`   | Missing Postgres     | N/A   | Logic exists but not wired to API or DB.               |
| FEAT-007 | Marketing Campaigns & UTM      | Multi-channel UTM tracking and builders               | MISSING   | 3     | `CampaignsPage.tsx` (mock)| `modules/campaign/utm.go`      | Missing Postgres     | N/A   | UI uses local state. Logic not wired to API.           |
| FEAT-008 | Multi-Touch Attribution        | Attribution calculation (First/Last/Linear)           | MISSING   | 4     | `AttributionPage.tsx`    | `modules/attribution`          | Missing Postgres     | N/A   | UI uses local state. Backend disconnected.             |
| FEAT-009 | Smart Routing (Geo/Device)     | Dynamic routing rules                                 | MISSING   | 4     | `SmartRoutingPage.tsx`   | `modules/redirect/router.go`   | Missing Postgres     | N/A   | Not implemented in DB schema or handler.               |
| FEAT-010 | SaaS Billing (Stripe)          | Tier enforcement and billing portal                   | MISSING   | 5     | `BillingPage.tsx` (mock) | `modules/billing/stripe.go`    | Missing Postgres     | N/A   | UI uses `INITIAL_PLAN` mock state.                     |
| FEAT-011 | Webhooks                       | Asynchronous event delivery                           | MISSING   | 5     | `WebhooksPage.tsx`       | `modules/webhook/deliver.go`   | Missing Postgres     | N/A   | Not wired.                                             |
| FEAT-012 | Multi-Tenant Workspaces (RBAC) | Workspaces and role permissions                       | MISSING   | 5     | `WorkspacesPage.tsx`     | `modules/tenant/rbac.go`       | Missing Postgres     | N/A   | Backend `links.go` passes `nil` for tenant ID.         |
| FEAT-013 | AI / Anomaly Detection         | CTR predictions, bot detection, threat scanning       | MISSING   | 6     | `AIInsightsPage.tsx`     | `modules/ai`                   | Missing DB/Model     | N/A   | Pure vaporware UI and disconnected logic.              |
| FEAT-014 | Base62 Encoding                | Generate short codes via Base62                       | BROKEN    | 1     | N/A                      | `pkg/base62`                   | N/A                  | N/A   | Links service uses random string instead of base62.    |

## Authentication
| ID | Feature | Status | Notes |
|----|---------|--------|-------|
| FEAT-015 | Native JWT Authentication | WORKING | `users` table, `/auth/signup`, `/auth/login` implemented. Replaces fake demo auth. |
