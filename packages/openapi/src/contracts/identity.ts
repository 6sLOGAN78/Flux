import { ZIdentityError, ZIdentityResponse } from "@flux/zod";
import { initContract } from "@ts-rest/core";
import { getSecurityMetadata } from "../utils.js";

const c = initContract();

export const identityContract = c.router({
  getMe: {
    summary: "Verify the current bearer session",
    description:
      "Requires an explicit bearer, exact issuer and authorized party, and a currently active provider session. All responses use Cache-Control: no-store. Cookies and provider organization claims grant no access.",
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
