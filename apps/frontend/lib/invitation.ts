import { ZInvitationInspectionResponse, ZWorkspaceResponse } from "@flux/zod";
import { ApiError, type createAPI } from "./api";

// The head bootstrap removes the fragment before provider scripts can run.
// Consume its closure into module memory; no persistence or return URL exists.
let arrival: string | undefined;
export const invitationArrival = () => {
  if (arrival === undefined && typeof window !== "undefined") {
    const entry = window as Window & { fluxInvitationArrival?: () => string };
    arrival = entry.fluxInvitationArrival?.() ?? "";
    delete entry.fluxInvitationArrival;
  }
  return arrival ?? "";
};
export const forgetInvitation = () => {
  arrival = "";
};

export const inspectInvitation = async (
  api: ReturnType<typeof createAPI>,
  token: string,
  signal: AbortSignal,
) =>
  ZInvitationInspectionResponse.parse(
    await api.request("/invitations/inspect", { method: "POST", body: { token }, signal }),
  ).invitation;

export const acceptInvitation = async (
  api: ReturnType<typeof createAPI>,
  token: string,
  signal: AbortSignal,
) =>
  ZWorkspaceResponse.parse(
    await api.request("/invitations/accept", { method: "POST", body: { token }, signal }),
  ).workspace;

export const invitationFailure = (error: unknown) => {
  if (error instanceof ApiError && [401, 403].includes(error.status))
    return "This invitation can't be accepted with this account. Sign in with the invited, verified email address.";
  if (error instanceof ApiError && [400, 404].includes(error.status))
    return "This invitation is no longer valid. Ask an owner or admin for a new invitation.";
  return "The invitation could not be loaded. Try again.";
};
