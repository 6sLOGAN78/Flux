import { initContract } from "@ts-rest/core";
import { ZHealthLiveResponse, ZHealthReadyResponse } from "@flux/zod";

const c = initContract();

export const healthContract = c.router({
  getLive: {
    summary: "Get process liveness",
    path: "/live",
    method: "GET",
    description: "Report that the process is alive without probing dependencies.",
    responses: {
      200: ZHealthLiveResponse,
    },
  },
  getReady: {
    summary: "Get process readiness",
    path: "/ready",
    method: "GET",
    description: "Report sanitized states for dependencies required by this process.",
    responses: {
      200: ZHealthReadyResponse,
      503: ZHealthReadyResponse,
    },
  },
});
