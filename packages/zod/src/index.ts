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
  userId: z.string().openapi({ description: "User UUID" }),
  email: z.string().email().openapi({ description: "User email" }),
  status: z.string().openapi({ description: "Authentication status string", example: "authenticated" }),
}).openapi({ description: "Current user profile authentication status" });

// --- Analytics & System Schemas ---

export const ZHealthResponse = z.object({
  status: z.string().openapi({ description: "Service status", example: "ok" }),
  database: z.string().openapi({ description: "PostgreSQL pool status", example: "connected" }),
}).openapi({ description: "Health check status payload" });

export const ZAnalyticsSummaryResponse = z.object({
  totalLinks: z.number().int().openapi({ description: "Total shortened links", example: 1250 }),
  totalClicks: z.number().int().openapi({ description: "Total recorded click events", example: 45000 }),
  activeDomains: z.number().int().openapi({ description: "Total verified custom domains", example: 12 }),
  topReferrer: z.string().openapi({ description: "Top referrer domain", example: "twitter.com" }),
}).openapi({ description: "High-level platform analytics summary" });

export const ZLinkMetricsResponse = z.object({
  linkId: z.string().uuid().openapi({ description: "Target link UUID" }),
  shortCode: z.string().openapi({ description: "Link short code slug" }),
  totalClicks: z.number().int().openapi({ description: "Total click count" }),
  clicksByDate: z.array(z.object({
    date: z.string().openapi({ description: "Date in YYYY-MM-DD format", example: "2026-08-08" }),
    clicks: z.number().int().openapi({ description: "Click count for date", example: 340 }),
  })).openapi({ description: "Daily click breakdown" }),
}).openapi({ description: "Detailed analytics metrics for a specific link" });

// --- Advanced Feature Schemas ---

export const ZQRCustomization = z.object({
  colorDark: z.string().openapi({ description: "Hex code for dark modules", example: "#000000" }),
  colorLight: z.string().openapi({ description: "Hex code for background", example: "#ffffff" }),
  logoUrl: z.string().url().optional().openapi({ description: "Optional embedded center logo URL" }),
}).openapi({ description: "QR Code styling parameters" });

export const ZABVariant = z.object({
  destinationUrl: z.string().url().openapi({ description: "Target destination URL" }),
  weight: z.number().min(0).max(100).openapi({ description: "Traffic allocation percentage weight", example: 50 }),
}).openapi({ description: "A/B test traffic distribution variant" });

export const ZOrganization = z.object({
  id: z.string().uuid().openapi({ description: "Organization UUID" }),
  name: z.string().openapi({ description: "Organization name", example: "Acme Corp" }),
  slug: z.string().openapi({ description: "Organization slug", example: "acme-corp" }),
}).openapi({ description: "Multi-tenant Organization entity" });

export const ZWorkspace = z.object({
  id: z.string().uuid().openapi({ description: "Workspace UUID" }),
  orgId: z.string().uuid().openapi({ description: "Parent Organization UUID" }),
  name: z.string().openapi({ description: "Workspace name", example: "Marketing Team" }),
}).openapi({ description: "Workspace scope entity" });

export const ZWorkspaceMember = z.object({
  id: z.string().uuid().openapi({ description: "Membership UUID" }),
  workspaceId: z.string().uuid().openapi({ description: "Workspace UUID" }),
  userId: z.string().uuid().openapi({ description: "User UUID" }),
  role: z.enum(["owner", "admin", "member", "viewer"]).openapi({ description: "RBAC role level", example: "admin" }),
}).openapi({ description: "Workspace member role mapping" });

export const ZSubscription = z.object({
  id: z.string().uuid().openapi({ description: "Subscription UUID" }),
  orgId: z.string().uuid().openapi({ description: "Organization UUID" }),
  plan: z.enum(["free", "pro", "enterprise"]).openapi({ description: "Billing plan tier", example: "enterprise" }),
  status: z.string().openapi({ description: "Stripe subscription status", example: "active" }),
  currentPeriodEnd: z.string().datetime().openapi({ description: "Billing cycle renewal timestamp" }),
}).openapi({ description: "Stripe Subscription details" });

export const ZAPIKey = z.object({
  id: z.string().uuid().openapi({ description: "API Key UUID" }),
  name: z.string().openapi({ description: "Key identifier label", example: "CI/CD Token" }),
  tokenPrefix: z.string().openapi({ description: "Key display prefix", example: "flx_live_a1b2..." }),
  scopes: z.array(z.string()).openapi({ description: "Granted permission scopes", example: ["links:read", "links:write"] }),
}).openapi({ description: "Public API key credential entity" });

export const ZOAuthTokenResponse = z.object({
  accessToken: z.string().openapi({ description: "OAuth2 Bearer Token", example: "eyJhbGciOi..." }),
  tokenType: z.string().openapi({ description: "Token type header", example: "Bearer" }),
  expiresIn: z.number().int().openapi({ description: "Validity duration in seconds", example: 3600 }),
}).openapi({ description: "OAuth token issuance payload" });

// --- Webhooks ---

export const ZWebhook = z.object({
  id: z.string().uuid().openapi({ description: "Webhook UUID" }),
  workspaceId: z.string().uuid().openapi({ description: "Workspace UUID" }),
  url: z.string().url().openapi({ description: "Target callback URL", example: "https://api.acme.com/webhooks" }),
  events: z.array(z.string()).openapi({ description: "Subscribed event triggers", example: ["link.created", "click.recorded"] }),
  isActive: z.boolean().openapi({ description: "Active status", example: true }),
}).openapi({ description: "Registered Webhook Endpoint entity" });

// --- Notifications ---

export const ZNotification = z.object({
  id: z.string().uuid().openapi({ description: "Notification UUID" }),
  userId: z.string().uuid().openapi({ description: "Recipient User UUID" }),
  title: z.string().openapi({ description: "Notification title", example: "Threshold Exceeded" }),
  message: z.string().openapi({ description: "Notification body message", example: "Link xyz123 reached 10,000 clicks" }),
  type: z.enum(["info", "warning", "alert"]).openapi({ description: "Alert severity level", example: "warning" }),
  linkUrl: z.string().url().optional().openapi({ description: "Optional action link URL" }),
  isRead: z.boolean().openapi({ description: "Read status flag", example: false }),
  createdAt: z.string().datetime().openapi({ description: "Creation timestamp" }),
}).openapi({ description: "User notification item" });

// --- Enterprise Attribution Schemas ---

export const ZAttributionModel = z.enum(["first_touch", "last_touch", "linear", "time_decay", "position_based"]).openapi({ description: "Attribution model algorithm", example: "position_based" });

export const ZCampaignAttribution = z.object({
  campaign_id: z.string().uuid().openapi({ description: "Campaign UUID" }),
  campaign_name: z.string().openapi({ description: "Campaign name", example: "Summer Launch" }),
  attributed_conversions: z.number().openapi({ description: "Attributed conversion count", example: 56.8 }),
  attributed_revenue: z.number().openapi({ description: "Attributed revenue amount", example: 18080.00 }),
}).openapi({ description: "Campaign attribution metric breakdown" });

export const ZAttributionResult = z.object({
  model: ZAttributionModel,
  total_conversions: z.number().int().openapi({ description: "Total evaluated conversions", example: 142 }),
  total_attributed_revenue: z.number().openapi({ description: "Total attributed revenue", example: 45200.00 }),
  campaigns: z.array(ZCampaignAttribution).openapi({ description: "Attribution metrics per campaign" }),
}).openapi({ description: "Attribution calculation result" });

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
export type Notification = z.infer<typeof ZNotification>;
export type AttributionModel = z.infer<typeof ZAttributionModel>;
export type CampaignAttribution = z.infer<typeof ZCampaignAttribution>;
export type AttributionResult = z.infer<typeof ZAttributionResult>;

// --- Enterprise Funnel Schemas ---

export const ZFunnelStepInput = z.object({
  step_order: z.number().int().openapi({ description: "Sequential step order index", example: 1 }),
  name: z.string().openapi({ description: "Funnel step name label", example: "Landing Page" }),
  link_id: z.string().uuid().openapi({ description: "Target link UUID" }),
}).openapi({ description: "Input specification for a funnel step" });

export const ZFunnelQueryPayload = z.object({
  funnel_name: z.string().openapi({ description: "Funnel name label", example: "Checkout Funnel" }),
  steps: z.array(ZFunnelStepInput).min(1).openapi({ description: "Ordered funnel steps array" }),
  from: z.string().datetime().optional().openapi({ description: "Start timestamp filter" }),
  to: z.string().datetime().optional().openapi({ description: "End timestamp filter" }),
}).openapi({ description: "Payload for executing funnel evaluation query" });

export const ZFunnelStepResult = z.object({
  step_order: z.number().int().openapi({ description: "Step order index", example: 1 }),
  name: z.string().openapi({ description: "Step name", example: "Landing Page" }),
  link_id: z.string().uuid().openapi({ description: "Link UUID" }),
  visitors: z.number().int().openapi({ description: "Visitor count reaching step", example: 100 }),
  overall_conversion_pct: z.number().openapi({ description: "Conversion rate relative to step 1", example: 100.0 }),
  step_conversion_pct: z.number().openapi({ description: "Conversion rate from previous step", example: 100.0 }),
  drop_off_count: z.number().int().openapi({ description: "Visitor drop-off count", example: 0 }),
  drop_off_pct: z.number().openapi({ description: "Percentage of visitors dropping off", example: 0.0 }),
}).openapi({ description: "Calculated result metrics for a single funnel step" });

export const ZFunnelAnalysisResult = z.object({
  funnel_name: z.string().openapi({ description: "Funnel name", example: "Checkout Funnel" }),
  total_started: z.number().int().openapi({ description: "Visitors starting at step 1", example: 100 }),
  total_converted: z.number().int().openapi({ description: "Visitors completing final step", example: 33 }),
  final_conversion_pct: z.number().openapi({ description: "Overall end-to-end conversion percentage", example: 33.33 }),
  steps: z.array(ZFunnelStepResult).openapi({ description: "Step-by-step breakdown results" }),
}).openapi({ description: "Complete funnel analysis result" });

export type FunnelStepInput = z.infer<typeof ZFunnelStepInput>;
export type FunnelQueryPayload = z.infer<typeof ZFunnelQueryPayload>;
export type FunnelStepResult = z.infer<typeof ZFunnelStepResult>;
export type FunnelAnalysisResult = z.infer<typeof ZFunnelAnalysisResult>;

// --- Enterprise Revenue Analytics Schemas ---

export const ZAdSpend = z.object({
  id: z.string().uuid(),
  campaign_id: z.string().uuid(),
  campaign_name: z.string().optional(),
  date: z.string().datetime(),
  amount_spent: z.number(),
  platform: z.string(),
});

export const ZCustomerConversion = z.object({
  customer_id: z.string().uuid(),
  campaign_id: z.string().uuid(),
  revenue: z.number(),
  converted_at: z.string().datetime(),
});

export const ZCampaignRevenueMetrics = z.object({
  campaign_id: z.string().uuid(),
  campaign_name: z.string().optional(),
  spend: z.number(),
  revenue: z.number(),
  customers_acquired: z.number().int(),
  cac: z.number(),
  roas: z.number(),
  roi_pct: z.number(),
  ltv: z.number(),
  ltv_to_cac_ratio: z.number(),
});

export const ZRevenueSummaryResult = z.object({
  total_spend: z.number(),
  total_revenue: z.number(),
  total_customers: z.number().int(),
  overall_cac: z.number(),
  overall_roas: z.number(),
  overall_roi_pct: z.number(),
  overall_ltv: z.number(),
  campaigns: z.array(ZCampaignRevenueMetrics),
});

export type AdSpend = z.infer<typeof ZAdSpend>;
export type CustomerConversion = z.infer<typeof ZCustomerConversion>;
export type CampaignRevenueMetrics = z.infer<typeof ZCampaignRevenueMetrics>;
export type RevenueSummaryResult = z.infer<typeof ZRevenueSummaryResult>;
