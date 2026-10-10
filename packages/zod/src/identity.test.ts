import assert from "node:assert/strict";
import { describe, test } from "node:test";
import { ZInvitation } from "./identity.js";

describe("safe invitation state contract", () => {
  const invitation = {
    id: "00000000-0000-4000-8000-000000000001",
    workspaceId: "00000000-0000-4000-8000-000000000002",
    email: "a.b+tag@example.test",
    role: "member",
    expiresAt: "2026-10-17T11:00:00Z",
  };
  test("accepts all durable public invitation states", () => {
    for (const status of ["Queued", "Delivered", "Failed", "Accepted", "Expired", "Revoked"])
      assert.equal(ZInvitation.safeParse({ ...invitation, status }).success, true);
  });
  test("rejects guessed states, unsafe recipients and private delivery fields", () => {
    for (const value of [
      { ...invitation, status: "Sent" },
      { ...invitation, status: "Queued", email: "a@b" },
      { ...invitation, status: "Queued", token: "private-canary" },
      { ...invitation, status: "Queued", ciphertext: "private-canary" },
      { ...invitation, status: "Queued", keyId: "fixture" },
    ])
      assert.equal(ZInvitation.safeParse(value).success, false);
  });
});
