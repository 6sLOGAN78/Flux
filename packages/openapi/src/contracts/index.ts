import { initContract } from "@ts-rest/core";
import { ZHealthResponse, ZLink, ZCreateLinkInput } from "@flux/zod";
import { z } from "zod";

const c = initContract();

export const apiContract = c.router({
  getHealth: {
    method: "GET",
    path: "/health",
    responses: {
      200: ZHealthResponse,
    },
    summary: "Health Check Endpoint",
  },
  createLink: {
    method: "POST",
    path: "/api/v1/links",
    body: ZCreateLinkInput,
    responses: {
      201: ZLink,
      400: z.object({ error: z.string() }),
    },
    summary: "Shorten a URL link",
  },
  getLink: {
    method: "GET",
    path: "/api/v1/links/:shortCode",
    pathParams: z.object({
      shortCode: z.string(),
    }),
    responses: {
      200: ZLink,
      404: z.object({ error: z.string() }),
    },
    summary: "Get shortened link details",
  },
});
