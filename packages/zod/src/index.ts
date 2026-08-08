import { z } from "zod";
import { extendZodWithOpenApi } from "@anatine/zod-openapi";

// Extend Zod with OpenAPI metadata capabilities
extendZodWithOpenApi(z);

// --- Link & Base Schemas ---

export const ZLink = z.object({
  id: z.string().uuid().openapi({ description: "Unique link identifier (UUID)", example: "123e4567-e89b-12d3-a456-426614174000" }),
  shortCode: z.string().min(1).max(20).openapi({ description: "Custom or generated Base62 short code", example: "xyz123" }),
  destinationUrl: z.string().url().openapi({ description: "Target URL for redirection", example: "https://example.com/target-page" }),
  categoryId: z.string().uuid().nullable().optional().openapi({ description: "Optional link category UUID" }),
  tenantId: z.string().uuid().optional().openapi({ description: "Tenant identifier for multi-tenant isolation" }),
  createdAt: z.string().datetime().openapi({ description: "Link creation timestamp in ISO 8601 format" }),
  updatedAt: z.string().datetime().openapi({ description: "Link last update timestamp in ISO 8601 format" }),
}).openapi({ description: "Shortened Link entity" });

export const ZCreateLinkInput = z.object({
  destinationUrl: z.string().url().openapi({ description: "Destination URL to shorten", example: "https://example.com/long-url" }),
  customCode: z.string().min(3).max(20).optional().openapi({ description: "Optional custom short code slug", example: "my-custom-alias" }),
  categoryId: z.string().uuid().optional().openapi({ description: "Optional category UUID assignment" }),
  title: z.string().max(100).optional().openapi({ description: "Optional title for the link", example: "Product Launch Page" }),
  description: z.string().max(255).optional().openapi({ description: "Optional description", example: "Summer sale landing page" }),
}).openapi({ description: "Input payload for creating a shortened link" });

export const ZUpdateLinkInput = z.object({
  destinationUrl: z.string().url().optional().openapi({ description: "Updated destination URL" }),
  categoryId: z.string().uuid().nullable().optional().openapi({ description: "Updated category UUID or null to unassign" }),
  title: z.string().max(100).optional().openapi({ description: "Updated link title" }),
  description: z.string().max(255).optional().openapi({ description: "Updated link description" }),
}).openapi({ description: "Input payload for updating a link" });

export const ZBulkCategorizeInput = z.object({
  linkIds: z.array(z.string().uuid()).openapi({ description: "Array of link UUIDs to categorize" }),
  categoryId: z.string().uuid().nullable().openapi({ description: "Category UUID to assign, or null to unassign" }),
}).openapi({ description: "Input payload for bulk link categorization" });

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

// --- QR Code & Advanced Customization ---

export const ZQRCustomization = z.object({
  fgColor: z.string().openapi({ description: "Foreground hex color", example: "#0f172a" }),
  bgColor: z.string().openapi({ description: "Background hex color", example: "#ffffff" }),
  logoUrl: z.string().url().optional().openapi({ description: "Center logo icon URL" }),
  dotStyle: z.string().openapi({ description: "QR matrix module pattern style", example: "circle" }),
}).openapi({ description: "QR Code styling options" });

// --- A/B Testing & Traffic Splitter ---

export const ZABVariant = z.object({
  id: z.string().openapi({ description: "Variant identifier" }),
  destinationUrl: z.string().url().openapi({ description: "Variant destination URL" }),
  weight: z.number().min(0).max(100).openapi({ description: "Traffic percentage weight", example: 50 }),
  clicks: z.number().int().nonnegative().openapi({ description: "Total clicks on variant" }),
  conversions: z.number().int().nonnegative().openapi({ description: "Total conversions on variant" }),
}).openapi({ description: "A/B testing destination variant" });

// --- Multi-Tenant SaaS & RBAC ---

export const ZOrganization = z.object({
  id: z.string().uuid().openapi({ description: "Organization UUID" }),
  name: z.string().openapi({ description: "Organization name", example: "Acme Corp" }),
  slug: z.string().openapi({ description: "Unique organization slug", example: "acme" }),
  billingEmail: z.string().email().openapi({ description: "Organization billing email", example: "billing@acme.com" }),
  createdAt: z.string().datetime().openapi({ description: "Creation timestamp" }),
}).openapi({ description: "Multi-tenant Organization entity" });

export const ZWorkspace = z.object({
  id: z.string().uuid().openapi({ description: "Workspace UUID" }),
  organizationId: z.string().uuid().openapi({ description: "Parent organization UUID" }),
  name: z.string().openapi({ description: "Workspace name", example: "Marketing Team" }),
  slug: z.string().openapi({ description: "Workspace slug", example: "marketing" }),
  isDefault: z.boolean().openapi({ description: "Default workspace flag", example: true }),
  createdAt: z.string().datetime().openapi({ description: "Creation timestamp" }),
}).openapi({ description: "Tenant Workspace entity" });

export const ZWorkspaceMember = z.object({
  id: z.string().uuid().openapi({ description: "Member UUID" }),
  workspaceId: z.string().uuid().openapi({ description: "Workspace UUID" }),
  userId: z.string().uuid().openapi({ description: "User UUID" }),
  role: z.enum(["owner", "admin", "editor", "viewer"]).openapi({ description: "RBAC role", example: "admin" }),
}).openapi({ description: "Workspace member role mapping" });

// --- Subscriptions & Billing ---

export const ZSubscription = z.object({
  id: z.string().uuid().optional().openapi({ description: "Subscription UUID" }),
  organizationId: z.string().uuid().openapi({ description: "Organization UUID" }),
  stripeCustomerId: z.string().openapi({ description: "Stripe customer identifier", example: "cus_123" }),
  planTier: z.enum(["free", "pro", "business"]).openapi({ description: "Subscription tier level", example: "pro" }),
  status: z.enum(["active", "past_due", "canceled", "trialing"]).openapi({ description: "Subscription billing status", example: "active" }),
}).openapi({ description: "Stripe subscription state entity" });

// --- Public API & OAuth 2.0 ---

export const ZAPIKey = z.object({
  id: z.string().uuid().openapi({ description: "API Key UUID" }),
  workspaceId: z.string().uuid().openapi({ description: "Workspace UUID" }),
  name: z.string().openapi({ description: "Key label name", example: "Production Backend Key" }),
  keyPrefix: z.string().openapi({ description: "Key prefix", example: "flx_live_" }),
  scopes: z.array(z.string()).openapi({ description: "Granted permission scopes", example: ["links:read", "links:write"] }),
  rateLimitPerMin: z.number().int().openapi({ description: "Requests per minute limit", example: 100 }),
}).openapi({ description: "Developer API Key entity" });

export const ZOAuthTokenResponse = z.object({
  access_token: z.string().openapi({ description: "Bearer Access Token", example: "flx_oauth_abc123" }),
  token_type: z.string().openapi({ example: "Bearer" }),
  expires_in: z.number().int().openapi({ example: 3600 }),
  scope: z.string().optional().openapi({ example: "links:read links:write" }),
}).openapi({ description: "OAuth 2.0 Access Token Response" });

// --- Webhooks ---

export const ZWebhook = z.object({
  id: z.string().uuid().openapi({ description: "Webhook UUID" }),
  workspaceId: z.string().uuid().openapi({ description: "Workspace UUID" }),
  url: z.string().url().openapi({ description: "Target callback URL", example: "https://api.acme.com/webhooks" }),
  events: z.array(z.string()).openapi({ description: "Subscribed event triggers", example: ["link.created", "click.recorded"] }),
  isActive: z.boolean().openapi({ description: "Active status", example: true }),
}).openapi({ description: "Registered Webhook Endpoint entity" });

// --- Exported Inferred Types ---
export type Link = z.infer<typeof ZLink>;
export type CreateLinkInput = z.infer<typeof ZCreateLinkInput>;
export type UpdateLinkInput = z.infer<typeof ZUpdateLinkInput>;
export type BulkCategorizeInput = z.infer<typeof ZBulkCategorizeInput>;
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
export type QRCustomization = z.infer<typeof ZQRCustomization>;
export type ABVariant = z.infer<typeof ZABVariant>;
export type Organization = z.infer<typeof ZOrganization>;
export type Workspace = z.infer<typeof ZWorkspace>;
export type WorkspaceMember = z.infer<typeof ZWorkspaceMember>;
export type Subscription = z.infer<typeof ZSubscription>;
export type APIKey = z.infer<typeof ZAPIKey>;
export type OAuthTokenResponse = z.infer<typeof ZOAuthTokenResponse>;
export type Webhook = z.infer<typeof ZWebhook>;
