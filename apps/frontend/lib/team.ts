import {
  ZInvitationResponse,
  ZInvitationsResponse,
  ZMemberRemovalResponse,
  type InvitationsResponse,
} from "@flux/zod";
import type { createAPI } from "./api";

// Validate scope before any server projection reaches Team rendering.
export const invitationPage = (value: unknown, workspaceId: string) => {
  const result = ZInvitationsResponse.parse(value);
  if (result.items.some((item) => item.workspaceId !== workspaceId))
    throw new Error("Unexpected resource");
  return result;
};

export const queuedInvitation = (value: unknown, workspaceId: string) => {
  const { invitation } = ZInvitationResponse.parse(value);
  if (invitation.workspaceId !== workspaceId || invitation.status !== "Queued")
    throw new Error("Unexpected resource");
  return invitation;
};

export const invitationStatus = (status: InvitationsResponse["items"][number]["status"]) =>
  status === "Failed" ? "Delivery failed" : status;

export const invitationTime = (value: string) =>
  new Intl.DateTimeFormat(undefined, { dateStyle: "full", timeStyle: "long" }).format(
    new Date(value),
  );

// Presentation hints use the closed server policy; every effect is reauthorized by Go.
export const canRemoveMember = (actor: string, target: string) =>
  ["owner", "admin", "member", "viewer"].includes(target) &&
  (actor === "owner" || (actor === "admin" && (target === "member" || target === "viewer")));

export const removalResult = (value: unknown, workspaceId: string, userId: string) => {
  const result = ZMemberRemovalResponse.parse(value);
  if (result.workspaceId !== workspaceId || result.removedUserId !== userId)
    throw new Error("Unexpected resource");
  return result;
};

export const removeMember = async (
  api: ReturnType<typeof createAPI>,
  workspaceId: string,
  userId: string,
  signal: AbortSignal,
) =>
  removalResult(
    await api.request(`/workspaces/${workspaceId}/members/${userId}`, {
      method: "DELETE",
      body: {},
      idempotencyKey: crypto.randomUUID(),
      signal,
    }),
    workspaceId,
    userId,
  );
