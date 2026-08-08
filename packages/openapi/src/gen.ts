import { generateOpenApi } from "@ts-rest/open-api";
import { apiContract } from "./contracts";
import * as fs from "node:fs";
import * as path from "node:path";

const openApiDocument = generateOpenApi(
  apiContract,
  {
    info: {
      title: "Flux Platform REST API",
      version: "2.0.0",
      description: "Production-grade URL shortening, dynamic link routing, custom domains, marketing campaigns, SaaS multi-tenancy, billing, webhooks, multi-channel notifications, and real-time analytics API.",
      contact: {
        name: "Flux API Engineering Team",
        email: "support@flux.dev",
        url: "https://flux.dev",
      },
      license: {
        name: "MIT",
        url: "https://opensource.org/licenses/MIT",
      },
    },
    servers: [
      {
        url: "http://localhost:8080",
        description: "Local Development Server",
      },
      {
        url: "https://api.flux.dev",
        description: "Production Gateway",
      },
    ],
    setOperationId: true,
  },
  {
    setMissingTagsFromPath: true,
  }
);

// Inject Security Schemes & Global Security requirements
openApiDocument.components = openApiDocument.components || {};
openApiDocument.components.securitySchemes = {
  bearerAuth: {
    type: "http",
    scheme: "bearer",
    bearerFormat: "JWT",
    description: "Enter your Bearer JWT token to authorize requests",
  },
  "x-service-token": {
    type: "apiKey",
    in: "header",
    name: "X-Service-Token",
    description: "Internal service mesh authentication token",
  },
};

openApiDocument.security = [
  { bearerAuth: [] },
];

openApiDocument.tags = [
  { name: "Health", description: "System operational status and health check endpoints" },
  { name: "Links", description: "Short URL creation, retrieval, Base62 routing, category assignment, and management" },
  { name: "Categories", description: "Category management and link organization" },
  { name: "Campaigns", description: "Marketing campaigns and UTM template building" },
  { name: "Domains", description: "Custom branded domain registration and CNAME verification" },
  { name: "Workspaces", description: "Multi-tenant Organization & Workspace management with RBAC" },
  { name: "Billing", description: "Stripe subscription tiers, checkout, and metered billing" },
  { name: "OAuth", description: "OAuth 2.0 Token issuance and developer API keys" },
  { name: "Webhooks", description: "Outbound real-time HTTP callback subscription management" },
  { name: "Notifications", description: "In-app alerts and multi-channel notification center" },
  { name: "User", description: "Authenticated user context and profile management" },
  { name: "Analytics", description: "Real-time link click metrics, time-series, and performance reporting" },
];

// Post-process path operations to assign explicit OpenAPI tags
for (const [pathKey, pathObj] of Object.entries(openApiDocument.paths || {})) {
  for (const [_, operation] of Object.entries(pathObj as Record<string, any>)) {
    if (typeof operation === "object" && operation !== null) {
      if (!operation.tags || operation.tags.length === 0) {
        if (pathKey.includes("/health")) {
          operation.tags = ["Health"];
        } else if (pathKey.includes("/links")) {
          operation.tags = ["Links"];
        } else if (pathKey.includes("/categories")) {
          operation.tags = ["Categories"];
        } else if (pathKey.includes("/campaigns")) {
          operation.tags = ["Campaigns"];
        } else if (pathKey.includes("/domains")) {
          operation.tags = ["Domains"];
        } else if (pathKey.includes("/workspaces")) {
          operation.tags = ["Workspaces"];
        } else if (pathKey.includes("/billing")) {
          operation.tags = ["Billing"];
        } else if (pathKey.includes("/oauth")) {
          operation.tags = ["OAuth"];
        } else if (pathKey.includes("/webhooks")) {
          operation.tags = ["Webhooks"];
        } else if (pathKey.includes("/notifications")) {
          operation.tags = ["Notifications"];
        } else if (pathKey.includes("/me")) {
          operation.tags = ["User"];
        } else if (pathKey.includes("/analytics")) {
          operation.tags = ["Analytics"];
        } else {
          operation.tags = ["General"];
        }
      }
    }
  }
}

const targetPathBackend = path.resolve(__dirname, "../../../apps/backend/static/openapi.json");
const targetPathPackage = path.resolve(__dirname, "../openapi.json");

for (const targetPath of [targetPathBackend, targetPathPackage]) {
  const targetDir = path.dirname(targetPath);
  if (!fs.existsSync(targetDir)) {
    fs.mkdirSync(targetDir, { recursive: true });
  }
  fs.writeFileSync(targetPath, JSON.stringify(openApiDocument, null, 2), "utf-8");
  console.log(`Successfully generated OpenAPI spec at: ${targetPath}`);
}
