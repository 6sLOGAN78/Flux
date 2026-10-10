import { expect } from "@playwright/test";
import { test } from "node:test";
import {
  canRemoveMember,
  removalResult,
  invitationPage,
  queuedInvitation,
  invitationStatus,
  invitationTime,
} from "./team";

const invitation = {
  id: "00000000-0000-4000-8000-000000000001",
  workspaceId: "00000000-0000-4000-8000-000000000002",
  email: "a.b+tag@example.test",
  role: "member",
  status: "Queued",
  expiresAt: "2026-10-17T11:00:00Z",
};

test("invitation projections reject foreign scope, private credentials and guessed statuses", () => {
  const page = { items: [invitation], nextAfter: null };
  expect(invitationPage(page, invitation.workspaceId)).toEqual(page);
  expect(queuedInvitation({ invitation }, invitation.workspaceId)).toEqual(invitation);
  for (const item of [
    { ...invitation, workspaceId: invitation.id },
    { ...invitation, token: "private-canary" },
    { ...invitation, status: "Sent" },
    { ...invitation, expiresAt: "invalid" },
  ]) {
    expect(() => invitationPage({ ...page, items: [item] }, invitation.workspaceId)).toThrow();
    expect(() => queuedInvitation({ invitation: item }, invitation.workspaceId)).toThrow();
  }
  expect(() =>
    queuedInvitation(
      { invitation: { ...invitation, status: "Delivered" } },
      invitation.workspaceId,
    ),
  ).toThrow();
});

test("invitation display preserves canonical states and gives full local expiry including timezone", () => {
  for (const [status, label] of [
    ["Queued", "Queued"],
    ["Delivered", "Delivered"],
    ["Failed", "Delivery failed"],
    ["Accepted", "Accepted"],
    ["Expired", "Expired"],
    ["Revoked", "Revoked"],
  ] as const)
    expect(invitationStatus(status)).toBe(label);
  expect(invitationTime(invitation.expiresAt)).toBe(
    new Date(invitation.expiresAt).toLocaleString(undefined, {
      dateStyle: "full",
      timeStyle: "long",
    }),
  );
});

test("removal controls deny unknown roles and match owner/admin target restrictions", () => {
  for (const actor of ["owner", "admin", "member", "viewer", "org:owner", "unknown", ""]) {
    for (const target of ["owner", "admin", "member", "viewer", "unknown", ""]) {
      const allowed =
        ["owner", "admin", "member", "viewer"].includes(target) &&
        (actor === "owner" || (actor === "admin" && ["member", "viewer"].includes(target)));
      expect(canRemoveMember(actor, target)).toBe(allowed);
    }
  }
});

test("committed removal rejects foreign identity, malformed response and private snapshots", () => {
  const workspaceId = "00000000-0000-4000-8000-000000000001";
  const removedUserId = "00000000-0000-4000-8000-000000000002";
  const valid = { workspaceId, removedUserId, selfRemoved: true };
  expect(removalResult(valid, workspaceId, removedUserId)).toEqual(valid);
  for (const value of [
    { ...valid, workspaceId: removedUserId },
    { ...valid, removedUserId: workspaceId },
    { ...valid, selfRemoved: "true" },
    { ...valid, role: "owner" },
    { ...valid, email: "private@example.test" },
    { workspaceId, removedUserId },
  ])
    expect(() => removalResult(value, workspaceId, removedUserId)).toThrow();
});
