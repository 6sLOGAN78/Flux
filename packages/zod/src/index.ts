import { z } from "zod";
import { extendZodWithOpenApi } from "@anatine/zod-openapi";

// Extend Zod with OpenAPI metadata capabilities
extendZodWithOpenApi(z);

// --- Entity Schemas ---

export const ZLink = z.object({
  id: z.string().uuid().openapi({ description: "Unique link identifier (UUID)", example: "123e4567-e89b-12d3-a456-426614174000" }),
  shortCode: z.string().min(1).max(20).openapi({ description: "Custom or generated Base62 short code", example: "xyz123" }),
  destinationUrl: z.string().url().openapi({ description: "Target URL for redirection", example: "https://example.com/target-page" }),
  tenantId: z.string().uuid().optional().openapi({ description: "Tenant identifier for multi-tenant isolation" }),
  createdAt: z.string().datetime().openapi({ description: "Link creation timestamp in ISO 8601 format" }),
  updatedAt: z.string().datetime().openapi({ description: "Link last update timestamp in ISO 8601 format" }),
}).openapi({ description: "Shortened Link entity" });

export const ZCreateLinkInput = z.object({
  destinationUrl: z.string().url().openapi({ description: "Destination URL to shorten", example: "https://example.com/long-url" }),
  customCode: z.string().min(3).max(20).optional().openapi({ description: "Optional custom short code slug", example: "my-custom-alias" }),
}).openapi({ description: "Input payload for creating a shortened link" });

export const ZUser = z.object({
  id: z.string().openapi({ description: "Unique user identifier", example: "usr_998127" }),
  email: z.string().email().openapi({ description: "User email address", example: "user@example.com" }),
  name: z.string().optional().openapi({ description: "User full name", example: "Jane Doe" }),
  createdAt: z.string().datetime().openapi({ description: "Account creation timestamp" }),
}).openapi({ description: "User account entity" });

export const ZAuthMeResponse = z.object({
  user_id: z.string().openapi({ description: "Authenticated user identifier", example: "usr_998127" }),
  email: z.string().email().openapi({ description: "User email address", example: "user@example.com" }),
  status: z.string().openapi({ description: "Session status", example: "authenticated" }),
}).openapi({ description: "Authenticated user session response" });

export const ZHealthResponse = z.object({
  status: z.string().openapi({ description: "Service operational status", example: "ok" }),
  database: z.string().openapi({ description: "Database connectivity status", example: "connected" }),
}).openapi({ description: "System health check response" });

export const ZAnalyticsSummaryResponse = z.object({
  totalClicks: z.number().int().nonnegative().openapi({ description: "Total click count across all links", example: 42150 }),
  uniqueVisitors: z.number().int().nonnegative().openapi({ description: "Total unique visitor IP count", example: 31200 }),
  activeLinks: z.number().int().nonnegative().openapi({ description: "Total active short links", example: 128 }),
}).openapi({ description: "Global analytics summary statistics" });

export const ZLinkMetricsResponse = z.object({
  linkId: z.string().uuid().openapi({ description: "Target link UUID" }),
  shortCode: z.string().openapi({ description: "Short code slug", example: "xyz123" }),
  totalClicks: z.number().int().nonnegative().openapi({ description: "Total click count for this link", example: 1540 }),
  clicksByDate: z.array(z.object({
    date: z.string().openapi({ example: "2026-08-08" }),
    clicks: z.number().int().nonnegative().openapi({ example: 320 }),
  })).openapi({ description: "Time-series daily click breakdown" }),
}).openapi({ description: "Detailed analytics metrics for a specific link" });

// --- Exported Inferred Types ---
export type Link = z.infer<typeof ZLink>;
export type CreateLinkInput = z.infer<typeof ZCreateLinkInput>;
export type User = z.infer<typeof ZUser>;
export type AuthMeResponse = z.infer<typeof ZAuthMeResponse>;
export type HealthResponse = z.infer<typeof ZHealthResponse>;
export type AnalyticsSummaryResponse = z.infer<typeof ZAnalyticsSummaryResponse>;
export type LinkMetricsResponse = z.infer<typeof ZLinkMetricsResponse>;

