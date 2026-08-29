# Frontend ↔ Backend Contracts

| ID | Frontend | Backend | Method | Endpoint | Status | Notes |
|----|----------|---------|--------|----------|--------|-------|
| CONTRACT-001 | `useGetLinks` | `LinksHandler.GetLinks` | GET | `/api/v1/links` | WORKING | Paginates from Postgres. |
| CONTRACT-002 | `useCreateLink` | `LinksHandler.CreateLink` | POST | `/api/v1/links` | WORKING | Creates link. Tenant is passed as nil. |
| CONTRACT-003 | `useUpdateLink` | `LinksHandler.UpdateLink` | PATCH | `/api/v1/links/:id` | WORKING | |
| CONTRACT-004 | `useBulkCategorize`| `LinksHandler.bulkCategorize`| POST | `/api/v1/links/bulk-categorize`| BROKEN | Endpoint not registered in `v1.go`. |
| CONTRACT-005 | `useAnalyticsQuery`| `AnalyticsHandler.GetSummary`| GET | `/api/v1/analytics/summary` | BROKEN | Returns 500 error (`nil` provider). |
| CONTRACT-006 | `CampaignsPage` | *Unimplemented* | POST | `/api/v1/campaigns` | BROKEN | Frontend uses local mocked array. |
| CONTRACT-007 | `BillingPage` | *Unimplemented* | GET | `/api/v1/billing/subscription`| BROKEN | Frontend uses local mocked array. |
| CONTRACT-008 | `WorkspacesPage` | *Unimplemented* | GET | `/api/v1/workspaces` | BROKEN | Frontend uses local mocked array. |

**Observation**: The `@flux/openapi` contract defines all endpoints perfectly, but the Go Echo router (`v1.go`) only implements `Links` and `Analytics` (which crashes).

## API Authorization (DEC-002)
| Category | Endpoint | Access Level | Description |
|----------|----------|--------------|-------------|
| System | `GET /:slug` | PUBLIC | Shortlink redirect engine |
| Authenticated | `GET /api/v1/me` | AUTHENTICATED | Get current user info |
| Authenticated | `GET /api/v1/links/*` | AUTHENTICATED | Links CRUD |
| Authenticated | `POST,PATCH,DELETE /api/v1/links/*`| AUTHENTICATED | Links CRUD |
| Authenticated | `GET /api/v1/analytics/*` | AUTHENTICATED | Analytics data |
