import { extendApi } from "@anatine/zod-openapi";
import { z } from "zod";

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

export const ZWorkspace = extendApi(
  z
    .object({
      id: extendApi(z.string().uuid(), { "x-go-type": "string" }),
      name: extendApi(
        z.string().refine((value) => [...value].length >= 1 && [...value].length <= 100),
        {
          minLength: 1,
          maxLength: 100,
        },
      ),
      role: z.enum(["owner", "admin", "member", "viewer"]),
    })
    .strict(),
  { title: "transport.Workspace" },
);

export const ZWorkspaceResponse = extendApi(
  z.object({ workspace: ZWorkspace, managedHost: z.string().optional() }).strict(),
  {
    title: "transport.WorkspaceResponse",
  },
);
export const ZWorkspacesResponse = extendApi(
  z.object({ workspaces: z.array(ZWorkspace) }).strict(),
  {
    title: "transport.WorkspacesResponse",
  },
);
export const ZCreateWorkspaceRequest = extendApi(
  z.object({ name: z.string().min(1).max(400) }).strict(),
  {
    title: "transport.CreateWorkspaceRequest",
    description:
      "Explicit workspace name; the server trims Unicode whitespace and validates 1–100 Unicode characters.",
  },
);
export type WorkspaceResponse = z.infer<typeof ZWorkspaceResponse>;

export const ZIdentityResponse = extendApi(
  z
    .object({
      authenticated: z.literal(true),
      workspaces: z.array(ZWorkspace),
      lastWorkspace: ZWorkspace.nullable(),
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

export type IdentityResponse = z.infer<typeof ZIdentityResponse>;

export const ZWorkspacePreferenceRequest = extendApi(
  z.object({ workspaceId: extendApi(z.string().uuid(), { "x-go-type": "string" }) }).strict(),
  {
    title: "transport.WorkspacePreferenceRequest",
    description: "Selection must be reauthorized against current Flux membership.",
  },
);

export const ZCreateLinkRequest = extendApi(
  z
    .object({
      destination: z.string().min(1).max(8192),
      customKey: z
        .string()
        .regex(/^[a-zA-Z0-9][a-zA-Z0-9_-]{2,63}$/)
        .optional(),
      title: extendApi(
        z.string().refine((value) => [...value].length <= 200),
        { maxLength: 200 },
      ).optional(),
    })
    .strict(),
  {
    title: "transport.CreateLinkRequest",
    description:
      "Public HTTP(S) destination, optional title of at most 200 Unicode characters and optional ASCII customKey, lowercased before validation and hashing. Keys use 3–64 letters, digits, underscores or hyphens and start with a letter or digit; system paths are reserved. Creation never fetches the destination.",
  },
);
export const ZLinkSuspension = extendApi(
  z
    .object({
      actorId: extendApi(z.string().uuid(), { "x-go-type": "string" }),
      reason: z.string(),
      at: extendApi(z.string().datetime({ offset: true }), { "x-go-type": "string" }),
    })
    .strict(),
  { title: "transport.LinkSuspension" },
);
export const ZLink = extendApi(
  z
    .object({
      id: extendApi(z.string().uuid(), { "x-go-type": "string" }),
      workspaceId: extendApi(z.string().uuid(), { "x-go-type": "string" }),
      shortUrl: z.string().url(),
      destination: z.string().url(),
      title: extendApi(
        z.string().refine((value) => [...value].length <= 200),
        { maxLength: 200 },
      ),
      creator: z
        .object({
          id: extendApi(z.string().uuid(), { "x-go-type": "string" }),
          email: extendApi(z.string().email(), { "x-go-type": "string" }),
        })
        .strict(),
      createdAt: extendApi(z.string().datetime({ offset: true }), { "x-go-type": "string" }),
      updatedAt: extendApi(z.string().datetime({ offset: true }), { "x-go-type": "string" }),
      version: z.string().regex(/^[1-9][0-9]*$/),
      lifecycle: z.enum(["active", "disabled", "archived", "deleted"]),
      suspension: ZLinkSuspension.nullable(),
    })
    .strict(),
  { title: "transport.Link" },
);
export const ZLinkResponse = extendApi(z.object({ link: ZLink }).strict(), {
  title: "transport.LinkResponse",
});
export type LinkResponse = z.infer<typeof ZLinkResponse>;

export const ZLinksResponse = extendApi(
  z
    .object({ items: z.array(ZLink).max(100), nextCursor: z.string().max(2048).nullable() })
    .strict(),
  { title: "transport.LinksResponse" },
);
export type LinksResponse = z.infer<typeof ZLinksResponse>;
