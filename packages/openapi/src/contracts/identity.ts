import {
  ZIdentityError,
  ZIdentityResponse,
  ZWorkspaceResponse,
  ZWorkspacesResponse,
  ZCreateWorkspaceRequest,
  ZWorkspacePreferenceRequest,
  ZCreateLinkRequest,
  ZLinkResponse,
} from "@flux/zod";
import { initContract } from "@ts-rest/core";
import { z } from "zod";
import { getSecurityMetadata } from "../utils.js";

const c = initContract();

export const identityContract = c.router({
  createLink: {
    summary: "Create a safe generated or custom-key managed-domain link",
    description:
      "Requires allowed Origin, JSON and Idempotency-Key. Fresh write membership is checked under workspace shared lock before actor-scoped create replay. SHA-256 canonical payload and committed response share a 24-hour ledger and atomic transaction. Twelve cryptographic random bytes yield lowercase unpadded base32 keys, with at most five retries only on the named global host/key constraint, including deleted rows. Optional ASCII customKey is lowercased and validated as 3–64 characters with system paths reserved; global collisions return KEY_UNAVAILABLE without owner data and changed canonical payload returns REQUEST_REUSE_CONFLICT. Public HTTP(S) destinations only; no fetch or preview. All responses are no-store.",
    path: "/api/v1/workspaces/:workspaceId/links",
    method: "POST",
    metadata: getSecurityMetadata(),
    pathParams: z.object({ workspaceId: z.string().uuid() }),
    headers: z.object({ "idempotency-key": z.string().regex(/^[A-Za-z0-9_-]{16,128}$/) }),
    body: ZCreateLinkRequest,
    responses: {
      201: ZLinkResponse,
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
  getLink: {
    summary: "Inspect a freshly authorized scoped link",
    description:
      "All current Flux members may read. Foreign, missing and removed access returns safe 404. Durable creator identity remains after membership removal. No-store; lifecycle does not prove redirect availability.",
    path: "/api/v1/workspaces/:workspaceId/links/:linkId",
    method: "GET",
    metadata: getSecurityMetadata(),
    pathParams: z.object({ workspaceId: z.string().uuid(), linkId: z.string().uuid() }),
    responses: {
      200: ZLinkResponse,
      401: ZIdentityError,
      404: ZIdentityError,
      429: ZIdentityError,
      503: ZIdentityError,
    },
  },
  selectWorkspace: {
    summary: "Select a currently authorized workspace",
    description:
      "Requires exact allowed Origin and JSON. Locks the workspace and rechecks Flux membership before committing the identity-owned preference. All roles may select. Unknown, foreign and removed memberships return the same 404. The preference never authorizes subsequent access. All responses use Cache-Control: no-store.",
    path: "/api/v1/me/last-workspace",
    method: "PUT",
    metadata: getSecurityMetadata(),
    body: ZWorkspacePreferenceRequest,
    responses: {
      200: ZWorkspaceResponse,
      400: ZIdentityError,
      401: ZIdentityError,
      403: ZIdentityError,
      404: ZIdentityError,
      413: ZIdentityError,
      415: ZIdentityError,
      429: ZIdentityError,
      503: ZIdentityError,
    },
  },
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
      "Requires an explicit bearer, exact issuer and authorized party, and a currently active provider session. First mapping verifies the provider primary email and commits a unique issuer/subject UUID in PostgreSQL; email never merges identities. Returns current Flux memberships and lastWorkspace only when authorized in the same membership snapshot, otherwise null without former workspace identifiers or names. All responses use Cache-Control: no-store. Cookies and provider organization claims grant no access.",
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
