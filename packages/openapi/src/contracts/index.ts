import { initContract } from "@ts-rest/core";
import {
  ZHealthResponse,
  ZLink,
  ZCreateLinkInput,
  ZAuthMeResponse,
  ZAnalyticsSummaryResponse,
  ZLinkMetricsResponse,
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
});

