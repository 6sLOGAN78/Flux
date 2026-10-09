import {
  ZIdentityError,
  ZIdentityResponse,
  ZWorkspaceResponse,
  ZWorkspacesResponse,
  ZCreateWorkspaceRequest,
} from "@flux/zod";
import { initContract } from "@ts-rest/core";
import { z } from "zod";
import { getSecurityMetadata } from "../utils.js";

const c = initContract();

export const identityContract = c.router({
  createWorkspace: {
    summary: "Create a named workspace and owner atomically",
    description:
      "Requires exact allowed Origin, JSON and Idempotency-Key. Canonical trimmed Unicode name is hashed with SHA-256; identity-scoped retries replay the committed workspace after fresh membership authorization, conflicting content returns 409. Ledger retention is at least 24 hours. All responses are no-store.",
    path: "/api/v1/workspaces",
    method: "POST",
    metadata: getSecurityMetadata(),
    headers: z.object({ "idempotency-key": z.string().regex(/^[A-Za-z0-9_-]{16,128}$/) }),
    body: ZCreateWorkspaceRequest,
    responses: {
      201: ZWorkspaceResponse,
      400: ZIdentityError,
      401: ZIdentityError,
      403: ZIdentityError,
      404: ZIdentityError,
      409: ZIdentityError,
      413: ZIdentityError,
      415: ZIdentityError,
      429: ZIdentityError,
      503: ZIdentityError,
    },
  },
  listWorkspaces: {
    summary: "List current workspace memberships",
    path: "/api/v1/workspaces",
    method: "GET",
    metadata: getSecurityMetadata(),
    responses: {
      200: ZWorkspacesResponse,
      401: ZIdentityError,
      429: ZIdentityError,
      503: ZIdentityError,
    },
  },
  getWorkspace: {
    summary: "Read a freshly authorized workspace summary",
    description:
      "Workspace and durable actor scope are required; foreign, unknown or removed memberships return the same 404. Provider organization claims confer no authority. All responses are no-store.",
    path: "/api/v1/workspaces/:workspaceId",
    method: "GET",
    metadata: getSecurityMetadata(),
    pathParams: z.object({ workspaceId: z.string().uuid() }),
    responses: {
      200: ZWorkspaceResponse,
      401: ZIdentityError,
      404: ZIdentityError,
      429: ZIdentityError,
      503: ZIdentityError,
    },
  },
  getMe: {
    summary: "Resolve the current internal user",
    description:
      "Requires an explicit bearer, exact issuer and authorized party, and a currently active provider session. First mapping verifies the provider primary email and commits a unique issuer/subject UUID in PostgreSQL; email never merges identities. All responses use Cache-Control: no-store. Cookies and provider organization claims grant no access.",
    path: "/api/v1/me",
    method: "GET",
    metadata: getSecurityMetadata(),
    responses: {
      200: ZIdentityResponse,
      401: ZIdentityError,
      429: ZIdentityError,
      503: ZIdentityError,
    },
  },
});
