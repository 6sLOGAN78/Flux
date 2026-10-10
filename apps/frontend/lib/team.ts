import { ZMemberRemovalResponse } from "@flux/zod";
import type { createAPI } from "./api";

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
