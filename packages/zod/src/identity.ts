import { extendApi } from "@anatine/zod-openapi";
import { z } from "zod";

export const ZIdentityResponse = extendApi(
  z
    .object({
      authenticated: z.literal(true),
      user: z
        .object({
          id: extendApi(z.string().uuid(), { "x-go-type": "string" }),
          email: extendApi(z.string().email().max(320), { "x-go-type": "string" }),
        })
        .strict(),
    })
    .strict(),
  {
    title: "transport.IdentityResponse",
    description:
      "Committed internal UUID and verified primary email for an active bearer session. Provider identity is never merged by email.",
  },
);

export const ZIdentityError = extendApi(
  z
    .object({
      code: z.string(),
      message: z.string(),
      status: z.number().int(),
      override: z.boolean(),
      errors: z.array(z.object({ field: z.string(), error: z.string() }).strict()).nullable(),
      action: z
        .object({ type: z.literal("redirect"), message: z.string(), value: z.string() })
        .strict()
        .nullable(),
    })
    .strict(),
  {
    title: "transport.IdentityError",
    description: "Safe authentication failure without token or provider diagnostics.",
  },
);

export type IdentityResponse = z.infer<typeof ZIdentityResponse>;
