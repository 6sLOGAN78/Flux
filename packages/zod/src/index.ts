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
  campaignId: z.string().uuid().optional().openapi({ description: "Optional campaign UUID assignment" }),
  title: z.string().max(100).optional().openapi({ description: "Optional title for the link", example: "Product Launch Page" }),
  description: z.string().max(255).optional().openapi({ description: "Optional description", example: "Summer sale landing page" }),
  utmSource: z.string().max(255).optional(),
  utmMedium: z.string().max(255).optional(),
  utmCampaign: z.string().max(255).optional(),
  utmTerm: z.string().max(255).optional(),
  utmContent: z.string().max(255).optional(),
}).openapi({ description: "Input payload for creating a shortened link" });

export const ZUpdateLinkInput = z.object({
  destinationUrl: z.string().url().optional().openapi({ description: "Updated destination URL" }),
  categoryId: z.string().uuid().nullable().optional().openapi({ description: "Updated category UUID or null to unassign" }),
  campaignId: z.string().uuid().nullable().optional().openapi({ description: "Updated campaign UUID or null to unassign" }),
  title: z.string().max(100).nullable().optional().openapi({ description: "Updated title" }),
  description: z.string().max(255).nullable().optional().openapi({ description: "Updated description" }),
  utmSource: z.string().max(255).nullable().optional(),
  utmMedium: z.string().max(255).nullable().optional(),
  utmCampaign: z.string().max(255).nullable().optional(),
  utmTerm: z.string().max(255).nullable().optional(),
  utmContent: z.string().max(255).nullable().optional(),
}).openapi({ description: "Input payload for updating a shortened link" });

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
  hostname: z.string().openapi({ description: "Branded hostname", example: "link.acme.com" }),
  verification_token: z.string().optional().openapi({ description: "DNS challenge verification token (only returned on creation)", example: "flux-verify=abc123" }),
  status: z.string().openapi({ description: "Verification status", example: "active" }),
}).openapi({ description: "Custom branded domain entity" });

export const ZCreateDomainInput = z.object({
  hostname: z.string().openapi({ description: "Domain hostname to configure", example: "link.acme.com" }),
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
  total_clicks: z.number().openapi({ description: "Total recorded click events", example: 45000 }),
  unique_visitors: z.number().openapi({ description: "Total unique visitors", example: 12000 }),
}).openapi({ description: "High-level platform analytics summary" });

export const ZTimeseriesDataPoint = z.object({
  timestamp: z.string(),
  clicks: z.number(),
  unique_visitors: z.number(),
});

export const ZTimeseriesResponse = z.object({
  data: z.array(ZTimeseriesDataPoint)
});

export const ZTopLink = z.object({
  link_id: z.string(),
  short_code: z.string(),
  clicks: z.number(),
});

export const ZTopLinksResponse = z.object({
  data: z.array(ZTopLink)
});

export const ZReferrerStat = z.object({
  referrer: z.string(),
  clicks: z.number(),
});

export const ZReferrersResponse = z.object({
  data: z.array(ZReferrerStat)
});

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
  plan: z.enum(["free", "pro", "business", "enterprise"]).openapi({ description: "Billing plan tier", example: "enterprise" }),
  status: z.string().openapi({ description: "Stripe subscription status", example: "active" }),
  currentPeriodEnd: z.string().datetime().openapi({ description: "Billing cycle renewal timestamp" }),
  maxLinks: z.number().openapi({ description: "Maximum active links allowed" }),
  analyticsRetention: z.number().openapi({ description: "Days to retain analytics data" }),

}).openapi({ description: "Stripe Subscription details" });

export const ZCustomerPortalResponse = z.object({
  url: z.string().url().openapi({ description: "Stripe Customer Portal URL", example: "https://billing.stripe.com/p/session/..." }),
}).openapi({ description: "URL to redirect the user to the Stripe Customer Portal" });

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
export type TimeseriesResponse = z.infer<typeof ZTimeseriesResponse>;
export type TopLinksResponse = z.infer<typeof ZTopLinksResponse>;
export type ReferrersResponse = z.infer<typeof ZReferrersResponse>;
export type LinkMetricsResponse = z.infer<typeof ZLinkMetricsResponse>;
export type QRCustomization = z.infer<typeof ZQRCustomization>;
export type ABVariant = z.infer<typeof ZABVariant>;
export type Organization = z.infer<typeof ZOrganization>;
export type Workspace = z.infer<typeof ZWorkspace>;
export type WorkspaceMember = z.infer<typeof ZWorkspaceMember>;
export type Subscription = z.infer<typeof ZSubscription>;
export type CustomerPortalResponse = z.infer<typeof ZCustomerPortalResponse>;
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

// --- Enterprise Predictive AI Schemas ---

export const ZAnomalyLog = z.object({
  id: z.string().uuid(),
  link_id: z.string().uuid(),
  anomaly_type: z.enum(["traffic_spike", "traffic_drop", "bot_surge"]),
  confidence_score: z.number(),
  summary: z.string(),
  created_at: z.string().datetime(),
});

export const ZAnomalyDetectionResult = z.object({
  link_id: z.string().uuid(),
  is_anomaly: z.boolean(),
  anomaly_type: z.enum(["traffic_spike", "traffic_drop", "bot_surge"]).optional(),
  z_score: z.number(),
  confidence_score: z.number(),
  summary: z.string(),
});

export const ZCTRPredictionResult = z.object({
  link_id: z.string().uuid(),
  historical_ctr: z.number(),
  predicted_ctr: z.number(),
  trend: z.enum(["upward", "downward", "stable"]),
  confidence: z.number(),
});

export type AnomalyLog = z.infer<typeof ZAnomalyLog>;
export type AnomalyDetectionResult = z.infer<typeof ZAnomalyDetectionResult>;
export type CTRPredictionResult = z.infer<typeof ZCTRPredictionResult>;

// --- Enterprise SSO & SCIM 2.0 Schemas ---

export const ZSSOConfig = z.object({
  id: z.string().uuid(),
  organization_id: z.string().uuid(),
  idp_type: z.enum(["saml", "oidc"]),
  entity_id: z.string(),
  sso_url: z.string().url(),
  certificate: z.string(),
  enforce_sso: z.boolean(),
});

export const ZSAMLAssertion = z.object({
  entity_id: z.string(),
  name_id: z.string().email(),
  session_index: z.string(),
  issue_instant: z.string().datetime(),
  attributes: z.record(z.string()).optional(),
});

export const ZSAMLValidationResult = z.object({
  is_valid: z.boolean(),
  user_email: z.string().email(),
  session_index: z.string(),
  attributes: z.record(z.string()).optional(),
});

export const ZSCIMUser = z.object({
  schemas: z.array(z.string()).optional(),
  id: z.string().optional(),
  externalId: z.string().optional(),
  userName: z.string().email(),
  name: z.object({ givenName: z.string(), familyName: z.string() }).optional(),
  emails: z.array(z.object({ value: z.string().email(), primary: z.boolean() })).optional(),
  active: z.boolean(),
});

export type SSOConfig = z.infer<typeof ZSSOConfig>;
export type SAMLAssertion = z.infer<typeof ZSAMLAssertion>;
export type SAMLValidationResult = z.infer<typeof ZSAMLValidationResult>;
export type SCIMUser = z.infer<typeof ZSCIMUser>;

// --- Enterprise White-Label Schemas ---

export const ZWhiteLabelConfig = z.object({
  id: z.string().uuid().optional(),
  organization_id: z.string().uuid().optional(),
  dashboard_domain: z.string().optional(),
  brand_name: z.string(),
  logo_url: z.string().url().optional(),
  favicon_url: z.string().url().optional(),
  primary_color: z.string().optional(),
  accent_color: z.string().optional(),
  custom_css: z.string().optional(),
  hide_footer: z.boolean().optional(),
  dkim_selector: z.string().optional(),
  dkim_domain: z.string().optional(),
});

export type WhiteLabelConfig = z.infer<typeof ZWhiteLabelConfig>;

// --- Enterprise Abuse & Malware Moderation Schemas ---

export const ZSecurityScan = z.object({
  id: z.string().uuid(),
  link_id: z.string().uuid(),
  url: z.string().url(),
  is_safe: z.boolean(),
  threat_type: z.string().optional(),
  threat_provider: z.string().optional(),
  scanned_at: z.string().datetime(),
});

export const ZScanResult = z.object({
  is_safe: z.boolean(),
  threat_type: z.string().optional(),
  threat_provider: z.string().optional(),
  reason: z.string().optional(),
});

export type SecurityScan = z.infer<typeof ZSecurityScan>;
export type ScanResult = z.infer<typeof ZScanResult>;

// --- Cloudflare Edge Redirect Worker Schemas ---

export const ZEdgeRedirectEvent = z.object({
  slug: z.string().openapi({ description: "Short URL slug executed at edge", example: "openai" }),
  timestamp: z.string().datetime().openapi({ description: "ISO 8601 timestamp of click event" }),
  ip: z.string().nullable().optional().openapi({ description: "Connecting IP address from Cloudflare header", example: "1.2.3.4" }),
  country: z.string().nullable().optional().openapi({ description: "Two-letter ISO country code", example: "US" }),
  userAgent: z.string().nullable().optional().openapi({ description: "Visitor User-Agent string" }),
  referrer: z.string().nullable().optional().openapi({ description: "HTTP Referer header" }),
}).openapi({ description: "Asynchronous click analytics event payload dispatched by Cloudflare Edge Worker" });

export const ZEdgeKVEntry = z.object({
  slug: z.string().openapi({ description: "Short link slug key", example: "openai" }),
  destinationUrl: z.string().url().openapi({ description: "Target destination URL for 302 redirect", example: "https://openai.com" }),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
}).openapi({ description: "Cloudflare Workers KV cache entry mapping slug to destination URL" });

export type EdgeRedirectEvent = z.infer<typeof ZEdgeRedirectEvent>;
export type EdgeKVEntry = z.infer<typeof ZEdgeKVEntry>;

// --- Geo-Distributed DB Replication & Edge Sync Schemas ---

export const ZRegionStatus = z.object({
  region: z.string().openapi({ description: "Regional node location code", example: "us-east" }),
  is_healthy: z.boolean().openapi({ description: "Regional node replication status health flag", example: true }),
  latency_ms: z.number().int().openapi({ description: "Regional ping latency in milliseconds", example: 12 }),
}).openapi({ description: "Regional database node status and SLA latency metrics" });

export const ZGeoClusterHealthResponse = z.object({
  regions: z.array(ZRegionStatus).openapi({ description: "Array of regional replication node statuses" }),
  sync_sla_met: z.boolean().openapi({ description: "True if all regional nodes meet <500ms sync SLA", example: true }),
}).openapi({ description: "Geo-distributed multi-region database cluster health response" });

export type RegionStatus = z.infer<typeof ZRegionStatus>;
export type GeoClusterHealthResponse = z.infer<typeof ZGeoClusterHealthResponse>;

// --- Anycast DNS & Edge TLS Schemas ---

export const ZPoPNode = z.object({
  id: z.string().openapi({ description: "Point of Presence node identifier", example: "pop-us-east-1" }),
  region: z.string().openapi({ description: "Geographic region", example: "us-east" }),
  anycast_ip: z.string().openapi({ description: "Anycast IPv4 address", example: "198.51.100.1" }),
  is_healthy: z.boolean().openapi({ description: "Health probe status", example: true }),
  bgp_state: z.enum(["advertised", "withdrawn"]).openapi({ description: "BGP route advertisement state", example: "advertised" }),
  latency_ms: z.number().int().openapi({ description: "PoP edge latency in milliseconds", example: 5 }),
}).openapi({ description: "Anycast BGP Edge Point of Presence node metadata" });

export const ZTLSCertificate = z.object({
  domain: z.string().openapi({ description: "Domain name pattern", example: "*.flux.dev" }),
  issuer: z.string().openapi({ description: "ACME Certificate Authority issuer", example: "Let's Encrypt Authority X3" }),
  status: z.enum(["active", "renewing", "expired"]).openapi({ description: "Certificate validity status", example: "active" }),
  fingerprint: z.string().openapi({ description: "SHA-256 certificate fingerprint hash" }),
  expires_at: z.string().datetime().openapi({ description: "Certificate expiration timestamp" }),
}).openapi({ description: "Automated Edge TLS Certificate details" });

export const ZAnycastStatusResponse = z.object({
  pops: z.array(ZPoPNode).openapi({ description: "List of Anycast PoP nodes" }),
  active_certificates: z.number().int().openapi({ description: "Total active edge TLS certificates", example: 42 }),
}).openapi({ description: "Anycast DNS routing and Edge TLS status summary" });

export type PoPNode = z.infer<typeof ZPoPNode>;
export type TLSCertificate = z.infer<typeof ZTLSCertificate>;
export type AnycastStatusResponse = z.infer<typeof ZAnycastStatusResponse>;

// --- Global Analytics Stream & Edge Batching Schemas ---

export const ZStreamMetrics = z.object({
  total_ingested_events: z.number().int().openapi({ description: "Total ingested click events count", example: 1000000 }),
  total_bytes_processed: z.number().int().openapi({ description: "Total uncompressed raw byte size", example: 154000000 }),
  compression_ratio: z.number().openapi({ description: "Gzip stream compression ratio factor", example: 4.2 }),
}).openapi({ description: "Global analytics stream ingestion metrics and compression performance" });

export const ZClickEventBatch = z.object({
  batch_id: z.string().uuid().openapi({ description: "Batch unique UUID", example: "123e4567-e89b-12d3-a456-426614174000" }),
  timestamp: z.string().datetime().openapi({ description: "Batch creation timestamp" }),
  event_count: z.number().int().openapi({ description: "Number of events in batch", example: 100 }),
}).openapi({ description: "Edge click event batch metadata" });

export type StreamMetrics = z.infer<typeof ZStreamMetrics>;
export type ClickEventBatch = z.infer<typeof ZClickEventBatch>;

// --- Global HA & Disaster Recovery Failover Schemas ---

export const ZFailoverResult = z.object({
  failover_triggered: z.boolean().openapi({ description: "True if DNS rerouting was executed", example: true }),
  previous_region: z.string().openapi({ description: "Degraded active region code", example: "us-east" }),
  new_active_region: z.string().openapi({ description: "Rerouted backup active region code", example: "eu-west" }),
  timestamp: z.string().datetime().openapi({ description: "Failover execution timestamp" }),
}).openapi({ description: "Automated regional failover execution result payload" });

export const ZClusterFailoverStatus = z.object({
  active_region: z.string().openapi({ description: "Currently active primary region code", example: "us-east" }),
  backup_regions: z.array(z.string()).openapi({ description: "Array of registered backup region codes" }),
  is_failed_over: z.boolean().openapi({ description: "True if running on backup region", example: false }),
  region_health: z.record(z.boolean()).openapi({ description: "Map of region codes to health status flags" }),
}).openapi({ description: "Global HA multi-region failover status payload" });

export type FailoverResult = z.infer<typeof ZFailoverResult>;
export type ClusterFailoverStatus = z.infer<typeof ZClusterFailoverStatus>;






export const ZCampaignPerformance = z.object({
  campaign_id: z.string().uuid().nullable().openapi({ description: "Campaign ID (null if no campaign)" }),
  clicks: z.number().int().openapi({ description: "Total clicks" }),
  unique_visitors: z.number().int().openapi({ description: "Total unique visitors" }),
}).openapi({ description: "Campaign performance metrics" });

export const ZCampaignPerformanceResponse = z.object({
  data: z.array(ZCampaignPerformance),
}).openapi({ description: "Response containing campaign performance data" });

export const ZUTMPerformance = z.object({
  utm_value: z.string().openapi({ description: "The grouped UTM value" }),
  clicks: z.number().int().openapi({ description: "Total clicks" }),
  unique_visitors: z.number().int().openapi({ description: "Total unique visitors" }),
}).openapi({ description: "UTM performance metrics" });

export const ZUTMPerformanceResponse = z.object({
  dimension: z.string().openapi({ description: "The dimension requested (e.g. utm_source)" }),
  data: z.array(ZUTMPerformance),
}).openapi({ description: "Response containing UTM performance data" });

export type CampaignPerformance = z.infer<typeof ZCampaignPerformance>;
export type CampaignPerformanceResponse = z.infer<typeof ZCampaignPerformanceResponse>;
export type UTMPerformance = z.infer<typeof ZUTMPerformance>;
export type UTMPerformanceResponse = z.infer<typeof ZUTMPerformanceResponse>;
