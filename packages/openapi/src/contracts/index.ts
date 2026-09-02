import { initContract } from "@ts-rest/core";
import {
  ZHealthResponse,
  ZGeoClusterHealthResponse,
  ZAnycastStatusResponse,
  ZStreamMetrics,
  ZClusterFailoverStatus,
  ZLink,


  ZCreateLinkInput,
  ZUpdateLinkInput,
  ZBulkCategorizeInput,
  ZCategory,
  ZCreateCategoryInput,
  ZCampaign,
  ZCreateCampaignInput,
  ZCustomDomain,
  ZCreateDomainInput,
  ZAuthMeResponse,
  ZAnalyticsSummaryResponse,
  ZTimeseriesResponse,
  ZTopLinksResponse,
  ZReferrersResponse, ZCampaignPerformanceResponse, ZUTMPerformanceResponse,
  ZLinkMetricsResponse,
  ZAttributionResult,
  ZOrganization,
  ZWorkspace,
  ZWorkspaceMember,
  ZSubscription,
  ZCustomerPortalResponse,
  ZAPIKey,
  ZOAuthTokenResponse,
  ZWebhook,
  ZWebhookDelivery,
  ZNotification,
} from "@flux/zod";
import { z } from "zod";

const c = initContract();

export const apiContract = c.router({
  getHealth: {
    method: "GET",
    path: "/health",
    responses: {
      200: ZHealthResponse,
    },
    summary: "Health Check Endpoint",
    metadata: {
      openApiTags: ["Health"],
    },
  },
  getGeoClusterHealth: {
    method: "GET",
    path: "/health/geo-cluster",
    responses: {
      200: ZGeoClusterHealthResponse,
    },
    summary: "Get multi-region DB replication and Edge KV sync health status",
    metadata: {
      openApiTags: ["Health"],
    },
  },
  getAnycastStatus: {
    method: "GET",
    path: "/health/anycast-dns",
    responses: {
      200: ZAnycastStatusResponse,
    },
    summary: "Get Anycast BGP DNS routing and Edge TLS status",
    metadata: {
      openApiTags: ["Health"],
    },
  },
  getFailoverStatus: {
    method: "GET",
    path: "/health/failover",
    responses: {
      200: ZClusterFailoverStatus,
    },
    summary: "Get multi-region automated HA failover status",
    metadata: {
      openApiTags: ["Health"],
    },
  },




  // --- Links ---
  createLink: {
    method: "POST",
    path: "/api/v1/links",
    body: ZCreateLinkInput,
    responses: {
      201: ZLink,
      400: z.object({ error: z.string().openapi({ example: "Invalid destination URL" }) }),
    },
    summary: "Shorten a URL link",
    metadata: {
      openApiTags: ["Links"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getLinks: {
    method: "GET",
    path: "/api/v1/links",
    query: z.object({
      page: z.string().optional(),
      limit: z.string().optional(),
      search: z.string().optional(),
    }),
    responses: {
      200: z.object({
        data: z.array(ZLink),
        total: z.number(),
        page: z.number(),
        limit: z.number(),
        totalPages: z.number()
      }),
    },
    summary: "List links with pagination",
    metadata: {
      openApiTags: ["Links"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getLink: {
    method: "GET",
    path: "/api/v1/links/:shortCode",
    pathParams: z.object({
      shortCode: z.string().openapi({ description: "Base62 short code slug", example: "xyz123" }),
    }),
    responses: {
      200: ZLink,
      404: z.object({ error: z.string().openapi({ example: "Link not found" }) }),
    },
    summary: "Get shortened link details",
    metadata: {
      openApiTags: ["Links"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  updateLink: {
    method: "PATCH",
    path: "/api/v1/links/:id",
    pathParams: z.object({
      id: z.string().uuid().openapi({ description: "Link UUID", example: "123e4567-e89b-12d3-a456-426614174000" }),
    }),
    body: ZUpdateLinkInput,
    responses: {
      200: ZLink,
      404: z.object({ error: z.string().openapi({ example: "Link not found" }) }),
    },
    summary: "Update link destination URL or category assignment",
    metadata: {
      openApiTags: ["Links"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  bulkCategorizeLinks: {
    method: "POST",
    path: "/api/v1/links/bulk-categorize",
    body: ZBulkCategorizeInput,
    responses: {
      200: z.object({ success: z.boolean(), updated_count: z.number().int() }),
    },
    summary: "Bulk assign or unassign category for multiple short links",
    metadata: {
      openApiTags: ["Links"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },

  // --- Categories ---
  createCategory: {
    method: "POST",
    path: "/api/v1/categories",
    body: ZCreateCategoryInput,
    responses: {
      201: ZCategory,
      400: z.object({ error: z.string().openapi({ example: "Invalid category payload" }) }),
    },
    summary: "Create link category",
    metadata: {
      openApiTags: ["Categories"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },

  // --- Campaigns ---
  createCampaign: {
    method: "POST",
    path: "/api/v1/campaigns",
    body: ZCreateCampaignInput,
    responses: {
      201: ZCampaign,
      400: z.object({ error: z.string().openapi({ example: "Invalid campaign payload" }) }),
    },
    summary: "Create marketing campaign with UTM template",
    metadata: {
      openApiTags: ["Campaigns"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getCampaigns: {
    method: "GET",
    path: "/api/v1/campaigns",
    responses: {
      200: z.array(ZCampaign),
      401: z.object({ error: z.string().openapi({ example: "Unauthorized" }) }),
    },
    summary: "List campaigns",
    metadata: {
      openApiTags: ["Campaigns"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getCampaign: {
    method: "GET",
    path: "/api/v1/campaigns/:id",
    pathParams: z.object({ id: z.string().uuid() }),
    responses: {
      200: ZCampaign,
      404: z.object({ error: z.string().openapi({ example: "Campaign not found" }) }),
    },
    summary: "Get a specific campaign",
    metadata: {
      openApiTags: ["Campaigns"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  updateCampaign: {
    method: "PATCH",
    path: "/api/v1/campaigns/:id",
    pathParams: z.object({ id: z.string().uuid() }),
    body: ZCreateCampaignInput.partial(),
    responses: {
      200: ZCampaign,
      404: z.object({ error: z.string().openapi({ example: "Campaign not found" }) }),
    },
    summary: "Update campaign",
    metadata: {
      openApiTags: ["Campaigns"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  deleteCampaign: {
    method: "DELETE",
    path: "/api/v1/campaigns/:id",
    pathParams: z.object({ id: z.string().uuid() }),
    responses: {
      204: z.undefined(),
      404: z.object({ error: z.string().openapi({ example: "Campaign not found" }) }),
    },
    summary: "Delete campaign",
    metadata: {
      openApiTags: ["Campaigns"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },

  // --- Domains ---
  createDomain: {
    method: "POST",
    path: "/api/v1/domains",
    body: ZCreateDomainInput,
    responses: {
      201: ZCustomDomain,
      400: z.object({ error: z.string().openapi({ example: "Invalid domain" }) }),
    },
    summary: "Register custom branded domain",
    metadata: {
      openApiTags: ["Domains"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getDomains: {
    method: "GET",
    path: "/api/v1/domains",
    responses: {
      200: z.object({ data: z.array(ZCustomDomain) }),
    },
    summary: "List custom domains",
    metadata: {
      openApiTags: ["Domains"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getDomain: {
    method: "GET",
    path: "/api/v1/domains/:id",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    responses: {
      200: ZCustomDomain,
      404: z.object({ error: z.string() }),
    },
    summary: "Get custom domain by ID",
    metadata: {
      openApiTags: ["Domains"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  deleteDomain: {
    method: "DELETE",
    path: "/api/v1/domains/:id",
    pathParams: z.object({
      id: z.string().uuid(),
    }),
    body: z.any().optional(), // ts-rest requires body or no body, we'll allow empty
    responses: {
      204: z.undefined(),
      404: z.object({ error: z.string() }),
    },
    summary: "Delete custom domain",
    metadata: {
      openApiTags: ["Domains"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },

  // --- Multi-Tenant & Workspaces ---
  getWorkspaces: {
    method: "GET",
    path: "/api/v1/workspaces",
    responses: {
      200: z.array(ZWorkspace),
    },
    summary: "List tenant workspaces for active user",
    metadata: {
      openApiTags: ["Workspaces"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },

  // --- Subscriptions & Billing ---
  getSubscription: {
    method: "GET",
    path: "/api/v1/billing/subscription",
    responses: {
      200: ZSubscription,
    },
    summary: "Get active SaaS plan subscription status",
    metadata: {
      openApiTags: ["Billing"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  createCustomerPortal: {
    method: "POST",
    path: "/api/v1/billing/portal",
    body: z.object({}),
    responses: {
      200: ZCustomerPortalResponse,
    },
    summary: "Create a Stripe Customer Portal session",
    metadata: {
      openApiTags: ["Billing"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },

  // --- OAuth 2.0 & Developer Keys ---
  issueOAuthToken: {
    method: "POST",
    path: "/oauth/token",
    body: z.object({
      grant_type: z.string().openapi({ example: "client_credentials" }),
      client_id: z.string().openapi({ example: "flx_app_123" }),
      client_secret: z.string().openapi({ example: "sec_abc999" }),
    }),
    responses: {
      200: ZOAuthTokenResponse,
    },
    summary: "Exchange credentials for OAuth 2.0 Access Token",
    metadata: {
      openApiTags: ["OAuth"],
    },
  },

  // --- Webhooks ---
  getWebhooks: {
    method: "GET",
    path: "/api/v1/webhooks",
    responses: {
      200: z.array(ZWebhook),
    },
    summary: "List outbound webhooks",
  },
  updateWebhook: {
    method: "PATCH",
    path: "/api/v1/webhooks/:id",
    pathParams: z.object({ id: z.string().uuid() }),
    body: z.object({
      endpoint_url: z.string().url().optional(),
      events: z.array(z.string()).optional(),
      active: z.boolean().optional(),
    }),
    responses: {
      200: ZWebhook,
    },
    summary: "Update webhook configuration",
  },
  deleteWebhook: {
    method: "DELETE",
    path: "/api/v1/webhooks/:id",
    pathParams: z.object({ id: z.string().uuid() }),
    body: z.object({}),
    responses: {
      204: z.undefined(),
    },
    summary: "Delete a webhook",
  },
  getWebhookDeliveries: {
    method: "GET",
    path: "/api/v1/webhooks/:id/deliveries",
    pathParams: z.object({ id: z.string().uuid() }),
    responses: {
      200: z.array(ZWebhookDelivery),
    },
    summary: "List webhook deliveries",
  },

  createWebhook: {
    method: "POST",
    path: "/api/v1/webhooks",
    body: z.object({
      endpoint_url: z.string().url().openapi({ example: "https://api.acme.com/webhook" }),
      events: z.array(z.string()).openapi({ example: ["link.redirect", "conversion"] }),
    }),
    responses: {
      201: ZWebhook,
    },
    summary: "Register outbound webhook listener endpoint",
    metadata: {
      openApiTags: ["Webhooks"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },

  // --- Notifications ---
  getNotifications: {
    method: "GET",
    path: "/api/v1/notifications",
    responses: {
      200: z.array(ZNotification),
    },
    summary: "Get unread user notifications list",
    metadata: {
      openApiTags: ["Notifications"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  markNotificationsRead: {
    method: "POST",
    path: "/api/v1/notifications/mark-read",
    body: z.object({
      notification_ids: z.array(z.string().uuid()).openapi({ description: "Array of notification UUIDs to mark read" }),
    }),
    responses: {
      200: z.object({ success: z.boolean(), updated_count: z.number().int() }),
    },
    summary: "Mark notifications as read",
    metadata: {
      openApiTags: ["Notifications"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },

  // --- User & Analytics ---
  getMe: {
    method: "GET",
    path: "/api/v1/me",
    responses: {
      200: ZAuthMeResponse,
      401: z.object({ error: z.string().openapi({ example: "Unauthorized - Invalid JWT" }) }),
    },
    summary: "Get authenticated user profile",
    metadata: {
      openApiTags: ["User"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getAnalyticsSummary: {
    method: "GET",
    path: "/api/v1/analytics/summary",
    query: z.object({
      from: z.string().optional(),
      to: z.string().optional(),
    }),
    responses: {
      200: ZAnalyticsSummaryResponse,
      401: z.object({ error: z.string().openapi({ example: "Unauthorized" }) }),
    },
    summary: "Get global analytics summary",
    metadata: {
      openApiTags: ["Analytics"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getAnalyticsTimeseries: {
    method: "GET",
    path: "/api/v1/analytics/timeseries",
    query: z.object({
      from: z.string().optional(),
      to: z.string().optional(),
      interval: z.string().optional(),
    }),
    responses: {
      200: ZTimeseriesResponse,
      401: z.object({ error: z.string().openapi({ example: "Unauthorized" }) }),
    },
    summary: "Get analytics timeseries data",
    metadata: {
      openApiTags: ["Analytics"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getAnalyticsTopLinks: {
    method: "GET",
    path: "/api/v1/analytics/top-links",
    query: z.object({
      from: z.string().optional(),
      to: z.string().optional(),
      limit: z.string().optional(),
    }),
    responses: {
      200: ZTopLinksResponse,
      401: z.object({ error: z.string().openapi({ example: "Unauthorized" }) }),
    },
    summary: "Get top links",
    metadata: {
      openApiTags: ["Analytics"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getAnalyticsReferrers: {
    method: "GET",
    path: "/api/v1/analytics/referrers",
    query: z.object({
      from: z.string().optional(),
      to: z.string().optional(),
      limit: z.string().optional(),
    }),
    responses: {
      200: ZReferrersResponse,
      401: z.object({ error: z.string().openapi({ example: "Unauthorized" }) }),
    },
    summary: "Get top referrers",
    metadata: {
      openApiTags: ["Analytics"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getAnalyticsCampaigns: {
    method: "GET",
    path: "/api/v1/analytics/campaigns",
    query: z.object({
      from: z.string().optional(),
      to: z.string().optional(),
      limit: z.string().optional(),
    }),
    responses: {
      200: ZCampaignPerformanceResponse,
      401: z.object({ error: z.string().openapi({ example: "Unauthorized" }) }),
    },
    summary: "Get campaign performance metrics",
    metadata: {
      openApiTags: ["Analytics"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getAnalyticsUTM: {
    method: "GET",
    path: "/api/v1/analytics/utm",
    query: z.object({
      dimension: z.string().optional(),
      from: z.string().optional(),
      to: z.string().optional(),
      limit: z.string().optional(),
    }),
    responses: {
      200: ZUTMPerformanceResponse,
      401: z.object({ error: z.string().openapi({ example: "Unauthorized" }) }),
    },
    summary: "Get UTM performance metrics by dimension",
    metadata: {
      openApiTags: ["Analytics"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getLinkMetrics: {
    method: "GET",
    path: "/api/v1/analytics/links/:id",
    pathParams: z.object({
      id: z.string().uuid().openapi({ description: "Link UUID", example: "123e4567-e89b-12d3-a456-426614174000" }),
    }),
    responses: {
      200: ZLinkMetricsResponse,
      404: z.object({ error: z.string().openapi({ example: "Link metrics not found" }) }),
    },
    summary: "Get detailed time-series metrics for a specific link",
    metadata: {
      openApiTags: ["Analytics"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getAnalyticsAttribution: {
    method: "GET",
    path: "/api/v1/analytics/attribution",
    query: z.object({
      from: z.string().optional(),
      to: z.string().optional(),
      model: z.string().optional(),
    }),
    responses: {
      200: ZAttributionResult,
      400: z.object({ error: z.string().openapi({ example: "Bad Request" }) }),
      401: z.object({ error: z.string().openapi({ example: "Unauthorized" }) }),
    },
    summary: "Get campaign attribution",
    metadata: {
      openApiTags: ["Analytics"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },
  getAnalyticsStreamMetrics: {
    method: "GET",
    path: "/api/v1/analytics/stream-metrics",
    responses: {
      200: ZStreamMetrics,
      401: z.object({ error: z.string().openapi({ example: "Unauthorized" }) }),
    },
    summary: "Get global real-time stream ingestion and compression metrics",
    metadata: {
      openApiTags: ["Analytics"],
      openApiSecurity: [{ bearerAuth: [] }],
    },
  },


  // --- Edge Redirect Engine ---
  edgeRedirect: {
    method: "GET",
    path: "/:slug",
    pathParams: z.object({
      slug: z.string().openapi({ description: "Short URL slug", example: "openai" }),
    }),
    responses: {
      302: z.undefined().openapi({ description: "302 Redirect to target destination URL" }),
      404: z.object({ error: z.string().openapi({ example: "Link Not Found" }) }),
    },
    summary: "Execute sub-10ms multi-region edge redirect",
    metadata: {
      openApiTags: ["Redirects"],
    },
  },
});

