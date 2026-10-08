import { ZIdentityError, ZIdentityResponse } from "@flux/zod";
import { initContract } from "@ts-rest/core";
import { getSecurityMetadata } from "../utils.js";

const c = initContract();

export const identityContract = c.router({
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
