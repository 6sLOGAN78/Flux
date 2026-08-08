import { z } from "zod";
import { extendZodWithOpenApi } from "@anatine/zod-openapi";

// Extend Zod with OpenAPI metadata capabilities
extendZodWithOpenApi(z);

// --- Link & Base Schemas ---

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
  title: z.string().max(100).optional().openapi({ description: "Optional title for the link", example: "Product Launch Page" }),
  description: z.string().max(255).optional().openapi({ description: "Optional description", example: "Summer sale landing page" }),
}).openapi({ description: "Input payload for creating a shortened link" });

// --- Category Schemas ---

export const ZCategory = z.object({
  id: z.string().uuid().openapi({ description: "Category UUID", example: "987e6543-e89b-12d3-a456-426614174000" }),
  name: z.string().min(1).max(100).openapi({ description: "Category name", example: "Marketing" }),
  color: z.string().openapi({ description: "Hex color code", example: "#3b82f6" }),
  description: z.string().optional().openapi({ description: "Optional category description" }),
  createdAt: z.string().datetime().openapi({ description: "Category creation timestamp" }),
  updatedAt: z.string().datetime().openapi({ description: "Category last update timestamp" }),
}).openapi({ description: "Category entity" });

export const ZCreateCategoryInput = z.object({
  name: z.string().min(1).max(100).openapi({ description: "Category name", example: "Marketing" }),
  color: z.string().openapi({ description: "Hex color code", example: "#3b82f6" }),
  description: z.string().max(255).optional().openapi({ description: "Category description" }),
}).openapi({ description: "Input payload for creating a category" });

// --- Campaign Schemas ---

export const ZCampaign = z.object({
  id: z.string().uuid().openapi({ description: "Campaign UUID" }),
  name: z.string().openapi({ description: "Campaign title name", example: "Q3 Growth Campaign" }),
  utm_source: z.string().optional().openapi({ description: "Default UTM Source", example: "twitter" }),
  utm_medium: z.string().optional().openapi({ description: "Default UTM Medium", example: "cpc" }),
  utm_campaign: z.string().optional().openapi({ description: "Default UTM Campaign", example: "q3_launch" }),
  status: z.string().openapi({ description: "Campaign operational status", example: "active" }),
}).openapi({ description: "Marketing campaign entity" });

export const ZCreateCampaignInput = z.object({
  name: z.string().min(1).max(150).openapi({ description: "Campaign name", example: "Q3 Growth Campaign" }),
  utm_source: z.string().optional().openapi({ description: "UTM Source", example: "twitter" }),
  utm_medium: z.string().optional().openapi({ description: "UTM Medium", example: "cpc" }),
  utm_campaign: z.string().optional().openapi({ description: "UTM Campaign", example: "q3_launch" }),
  utm_term: z.string().optional().openapi({ description: "UTM Term" }),
  utm_content: z.string().optional().openapi({ description: "UTM Content" }),
}).openapi({ description: "Input payload for creating a campaign" });

// --- Custom Domain Schemas ---

export const ZCustomDomain = z.object({
  id: z.string().uuid().openapi({ description: "Custom domain UUID" }),
  domain: z.string().openapi({ description: "Branded hostname", example: "link.acme.com" }),
  verification_token: z.string().openapi({ description: "DNS challenge verification token", example: "flux-verify=abc123" }),
  is_verified: z.boolean().openapi({ description: "CNAME verification status", example: true }),
  ssl_status: z.string().openapi({ description: "ACME TLS/SSL certificate status", example: "active" }),
}).openapi({ description: "Custom branded domain entity" });

export const ZCreateDomainInput = z.object({
  domain: z.string().openapi({ description: "Domain hostname to configure", example: "link.acme.com" }),
  custom_root_redirect: z.string().url().optional().openapi({ description: "Root domain fallback URL" }),
}).openapi({ description: "Input payload for adding a custom domain" });

// --- User & Auth Schemas ---

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
export type Category = z.infer<typeof ZCategory>;
export type CreateCategoryInput = z.infer<typeof ZCreateCategoryInput>;
export type Campaign = z.infer<typeof ZCampaign>;
export type CreateCampaignInput = z.infer<typeof ZCreateCampaignInput>;
export type CustomDomain = z.infer<typeof ZCustomDomain>;
export type CreateDomainInput = z.infer<typeof ZCreateDomainInput>;
export type User = z.infer<typeof ZUser>;
export type AuthMeResponse = z.infer<typeof ZAuthMeResponse>;
export type HealthResponse = z.infer<typeof ZHealthResponse>;
export type AnalyticsSummaryResponse = z.infer<typeof ZAnalyticsSummaryResponse>;
export type LinkMetricsResponse = z.infer<typeof ZLinkMetricsResponse>;
