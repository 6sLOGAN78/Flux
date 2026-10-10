import { extendApi } from "@anatine/zod-openapi";
import { z } from "zod";

export const ZCreateInvitationRequest = extendApi(
  z
    .object({
      email: z.string().min(3).max(1024),
      role: extendApi(z.enum(["admin", "member", "viewer"]), {
        "x-enum-varnames": ["InviteGrantAdmin", "InviteGrantMember", "InviteGrantViewer"],
      }),
    })
    .strict(),
  {
    title: "transport.CreateInvitationRequest",
    description:
      "Server trims and casefolds email, max 254 bytes, preserving dot/plus aliases; owner grants admin/member/viewer, admin grants member/viewer.",
  },
);
export const ZInvitation = extendApi(
  z
    .object({
      id: extendApi(z.string().uuid(), { "x-go-type": "string" }),
      workspaceId: extendApi(z.string().uuid(), { "x-go-type": "string" }),
      email: extendApi(z.string().email().max(254), { "x-go-type": "string" }),
      role: extendApi(z.enum(["admin", "member", "viewer"]), {
        "x-enum-varnames": ["InviteAdmin", "InviteMember", "InviteViewer"],
      }),
      status: extendApi(z.enum(["Queued", "Expired", "Accepted", "Revoked"]), {
        "x-enum-varnames": ["InviteQueued", "InviteExpired", "InviteAccepted", "InviteRevoked"],
      }),
      expiresAt: extendApi(z.string().datetime({ offset: true }), { "x-go-type": "string" }),
    })
    .strict(),
  { title: "transport.Invitation" },
);
export const ZInvitationResponse = extendApi(z.object({ invitation: ZInvitation }).strict(), {
  title: "transport.InvitationResponse",
});
export const ZInvitationsResponse = extendApi(
  z
    .object({
      items: z.array(ZInvitation).max(25),
      nextAfter: extendApi(z.string().uuid().nullable(), { "x-go-type": "string" }),
    })
    .strict(),
  { title: "transport.InvitationsResponse" },
);
export type InvitationsResponse = z.infer<typeof ZInvitationsResponse>;

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

export const ZTeamMember = extendApi(
  z
    .object({
      id: extendApi(z.string().uuid(), { "x-go-type": "string" }),
      workspaceId: extendApi(z.string().uuid(), { "x-go-type": "string" }),
      email: extendApi(z.string().email().max(320), { "x-go-type": "string" }),
      role: extendApi(z.enum(["owner", "admin", "member", "viewer"]), {
        "x-enum-varnames": ["TeamOwner", "TeamAdmin", "TeamMember", "TeamViewer"],
      }),
    })
    .strict(),
  { title: "transport.TeamMember" },
);
export const ZMembersResponse = extendApi(
  z
    .object({
      items: z.array(ZTeamMember).max(25),
      nextAfter: extendApi(z.string().uuid().nullable(), {
        "x-go-type": "string",
        description:
          "Last user UUID continuation; null means authoritative exhaustion. Never grants workspace authority.",
      }),
    })
    .strict(),
  { title: "transport.MembersResponse" },
);
export type MembersResponse = z.infer<typeof ZMembersResponse>;
export const ZChangeMemberRoleRequest = extendApi(
  z
    .object({
      role: extendApi(z.enum(["owner", "admin", "member", "viewer"]), {
        "x-enum-varnames": ["GrantOwner", "GrantAdmin", "GrantMember", "GrantViewer"],
      }),
    })
    .strict(),
  { title: "transport.ChangeMemberRoleRequest" },
);
export const ZMemberResponse = extendApi(
  z
    .object({
      member: ZTeamMember,
      actorRole: extendApi(z.enum(["owner", "admin", "member", "viewer"]), {
        "x-enum-varnames": [
          "CapabilityOwner",
          "CapabilityAdmin",
          "CapabilityMember",
          "CapabilityViewer",
        ],
      }),
    })
    .strict(),
  { title: "transport.MemberResponse" },
);
export const ZMemberListQuery = z
  .object({ after: extendApi(z.string().uuid(), { "x-go-type": "string" }).optional() })
  .strict();

export const ZRemoveMemberRequest = extendApi(z.object({}).strict(), {
  title: "transport.RemoveMemberRequest",
});
export const ZMemberRemovalResponse = extendApi(
  z
    .object({
      removedUserId: extendApi(z.string().uuid(), { "x-go-type": "string" }),
      workspaceId: extendApi(z.string().uuid(), { "x-go-type": "string" }),
      selfRemoved: z.boolean(),
    })
    .strict(),
  { title: "transport.MemberRemovalResponse" },
);

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
    .object({
      items: z.array(ZLink).max(100),
      nextCursor: extendApi(z.string().min(1).max(2048).nullable(), {
        description:
          "Opaque signed workspace/query-bound continuation; null means authoritative exhaustion.",
      }),
    })
    .strict(),
  { title: "transport.LinksResponse" },
);
export type LinksResponse = z.infer<typeof ZLinksResponse>;

// Canonical query contract: Go normalizes surrounding Unicode whitespace before
// binding the exact search text and lifecycle into its signed cursor.
export const ZLinkListQuery = extendApi(
  z
    .object({
      limit: z
        .string()
        .regex(/^[0-9]{1,3}$/)
        .refine((value) => Number(value) >= 1 && Number(value) <= 100)
        .optional(),
      cursor: z.string().min(1).max(2048).optional(),
      search: extendApi(
        z.string().refine((value) => [...value].length <= 200 && !value.includes("\0")),
        {
          maxLength: 200,
          description:
            "Literal case-insensitive substring across short key, title or destination; surrounding Unicode whitespace is trimmed. Maximum 200 Unicode characters.",
        },
      ).optional(),
      state: z.enum(["nondeleted", "active", "disabled", "archived", "deleted"]).optional(),
    })
    .strict(),
  { title: "transport.LinkListQuery" },
);
