import { expect } from "@playwright/test";
import { test } from "node:test";
import { createAPI, ApiError } from "./api";
import { inspectInvitation, acceptInvitation, invitationFailure } from "./invitation";

test("invitation credentials only cross scoped no-store POST bodies", async () => {
  const secret = "A".repeat(43);
  const requests: { path: string; options: RequestInit }[] = [];
  const api = createAPI(async () => "signed-fixture", {
    origin: "https://flux.example.test",
    fetch: async (url, options) => {
      requests.push({ path: String(url), options: options ?? {} });
      return new Response("{}", { status: 503 });
    },
  });
  for (const operation of [inspectInvitation, acceptInvitation])
    await expect(operation(api, secret, new AbortController().signal)).rejects.toBeInstanceOf(
      ApiError,
    );
  expect(requests.map((request) => new URL(request.path).pathname)).toEqual([
    "/api/v1/invitations/inspect",
    "/api/v1/invitations/accept",
  ]);
  for (const request of requests) {
    expect(request.path).not.toContain(secret);
    expect(request.options).toMatchObject({
      method: "POST",
      cache: "no-store",
      credentials: "omit",
      referrerPolicy: "no-referrer",
      redirect: "error",
      body: JSON.stringify({ token: secret }),
    });
  }
});

test("invitation failures expose only approved account-safe recovery copy", () => {
  for (const status of [401, 403])
    expect(invitationFailure(new ApiError("forbidden", status))).toBe(
      "This invitation can't be accepted with this account. Sign in with the invited, verified email address.",
    );
  for (const status of [400, 404])
    expect(invitationFailure(new ApiError("request_failed", status))).toBe(
      "This invitation is no longer valid. Ask an owner or admin for a new invitation.",
    );
  expect(invitationFailure(new ApiError("unavailable", 503))).toBe(
    "The invitation could not be loaded. Try again.",
  );
  expect(invitationFailure(new Error("private-recipient-token"))).not.toContain("private");
});
