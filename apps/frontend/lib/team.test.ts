import { expect } from "@playwright/test";
import { test } from "node:test";
import { canRemoveMember, removalResult } from "./team";

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
