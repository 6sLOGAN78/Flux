// Presentation hints follow the closed Flux role matrix. Go reauthorizes effects.
export const workspaceCapabilities = (role: string) => {
  const valid = ["owner", "admin", "member", "viewer"].includes(role);
  return {
    read: valid,
    create: valid && role !== "viewer",
    team: role === "owner" || role === "admin",
  };
};
import { ZIdentityResponse, ZWorkspaceResponse } from "@flux/zod";
import type { createAPI } from "./api";
import { ApiError } from "./api";

// Abort transports and reject completions after *any* asynchronous boundary.
export class WorkspaceRequests {
  private generation = 0;
  private active?: AbortController;

  clear() {
    this.generation++;
    this.active?.abort();
    this.active = undefined;
  }

  begin() {
    this.clear();
    const generation = this.generation;
    const controller = new AbortController();
    this.active = controller;
    return {
      signal: controller.signal,
      current: () => generation === this.generation && !controller.signal.aborted,
    };
  }
}

export const workspaceChannel = "flux.workspace-invalidation";
export const invalidateWorkspaces = (channel: { postMessage(value: unknown): void } | null) =>
  channel?.postMessage("invalidate");

// Messages may invalidate state, but can never select a tenant or supply data.
export const listenWorkspaceInvalidation = (channel: EventTarget, clear: () => void) => {
  const listener = (event: Event) => {
    if ((event as MessageEvent<unknown>).data === "invalidate") clear();
  };
  channel.addEventListener("message", listener);
  return () => channel.removeEventListener("message", listener);
};

export const resolveWorkspace = async (
  api: ReturnType<typeof createAPI>,
  id: string,
  signal: AbortSignal,
) => {
  const bootstrap = () =>
    api.request("/me", { signal }).then((value) => ZIdentityResponse.parse(value));
  let identity = await bootstrap();
  let membership = identity.workspaces.find((workspace) => workspace.id === id);
  if (!membership) throw new ApiError("forbidden", 404);
  try {
    const result = ZWorkspaceResponse.parse(await api.request(`/workspaces/${id}`, { signal }));
    if (result.workspace.id !== id) throw new ApiError("unavailable");
    return { workspace: result.workspace, workspaces: identity.workspaces };
  } catch (error) {
    if (!(error instanceof ApiError) || ![403, 404].includes(error.status)) throw error;
    // Resource absence is not evidence of membership loss. Confirm authority.
    identity = await bootstrap();
    membership = identity.workspaces.find((workspace) => workspace.id === id);
    if (!membership) throw new ApiError("forbidden", 404);
    return { workspace: membership, workspaces: identity.workspaces };
  }
};

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
