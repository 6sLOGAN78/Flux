import { initContract } from "@ts-rest/core";
import {
  ZHealthResponse,
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
  ZLinkMetricsResponse,
  ZOrganization,
  ZWorkspace,
  ZWorkspaceMember,
  ZSubscription,
  ZAPIKey,
  ZOAuthTokenResponse,
  ZWebhook,
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
  createWebhook: {
    method: "POST",
    path: "/api/v1/webhooks",
    body: z.object({
      url: z.string().url().openapi({ example: "https://api.acme.com/webhook" }),
      events: z.array(z.string()).openapi({ example: ["link.created", "click.recorded"] }),
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

