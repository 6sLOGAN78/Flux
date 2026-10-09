// Presentation hints follow the closed Flux role matrix. Go reauthorizes effects.
export const workspaceCapabilities = (role: string) => {
  const valid = ["owner", "admin", "member", "viewer"].includes(role);
  return {
    read: valid,
    create: valid && role !== "viewer",
    team: role === "owner" || role === "admin",
  };
};
import { ZWorkspaceResponse } from "@flux/zod";
import type { createAPI } from "./api";

// Only the committed response supplies route identity; caller selection is input.
export const selectWorkspace = async (
  api: ReturnType<typeof createAPI>,
  workspaceId: string,
  signal: AbortSignal,
) => {
  const result = ZWorkspaceResponse.parse(
    await api.request("/me/last-workspace", {
      method: "PUT",
      body: { workspaceId },
      signal,
    }),
  );
  if (result.workspace.id !== workspaceId) throw new Error("Workspace selection failed");
  return result.workspace;
};
