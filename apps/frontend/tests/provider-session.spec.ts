import { expect, test } from "@playwright/test";
import {
  providerAPI,
  providerBAPI,
  providerControl,
  providerSession,
  providerRow,
  providerToken,
} from "./provider.config";

const claims = (token: string) =>
  JSON.parse(Buffer.from(token.split(".")[1]!, "base64url").toString("utf8")) as {
    exp: number;
    iat: number;
    sid: string;
    iss: string;
  };

test("real SDK refresh and old signed token expiry", async ({ page }, info) => {
  test.setTimeout(120000);
  const old = await providerSession(page);
  const decoded = claims(old);
  expect(decoded.iss === process.env["FLUX_AUTH.ISSUER"]).toBe(true);
  const wait = decoded.exp * 1000 - Date.now() + 1000;
  expect(wait > 0 && wait < 90000).toBe(true);
  await expect.poll(async () => (await providerToken(page)) !== old).toBe(true);
  expect((await providerAPI(page, "/me")).status).toBe(200);
  providerRow(info, "token_refresh");
  await new Promise((done) =>
    setTimeout(done, Math.max(0, decoded.exp * 1000 - Date.now() + 1000)),
  );
  expect((await providerAPI(page, "/me", "GET", undefined, old)).status).toBe(401);
  expect((await providerAPI(page, "/me")).status).toBe(200);
  providerRow(info, "token_expiry");
});

test("real active session revocation rejects the still-signed token", async ({ page }, info) => {
  const old = await providerSession(page);
  expect((await providerAPI(page, "/me", "GET", undefined, old)).status).toBe(200);
  await providerBAPI(`/v1/sessions/${claims(old).sid}/revoke`, "POST");
  expect((await providerAPI(page, "/me", "GET", undefined, old)).status).toBe(401);
  providerRow(info, "session_revocation");
});

test("real provider signout invalidates the prior bearer", async ({ page }, info) => {
  const old = await providerSession(page);
  await page.evaluate(async () =>
    (window as unknown as { Clerk: { signOut(): Promise<void> } }).Clerk.signOut(),
  );
  expect((await providerAPI(page, "/me", "GET", undefined, old)).status).toBe(401);
  providerRow(info, "session_signout");
});

test("real session maximum and measured control-plane provider capacity", async ({
  page,
}, info) => {
  const token = await providerSession(page);
  const session = await providerBAPI(`/v1/sessions/${claims(token).sid}`);
  const maximum = Number(session.expire_at) - Number(session.created_at);
  expect(Math.abs(maximum - 7 * 86400000) < 60000).toBe(true);
  providerRow(info, "session_maximum");
  // 2 requests/s leaves headroom below Clerk's documented development ceiling
  // of 100 BAPI requests/10s; each Go request rechecks active provider state.
  const requests = await Promise.allSettled(
    Array.from({ length: 10 }, async (_, sample) => {
      await new Promise((done) => setTimeout(done, sample * 500));
      const start = Date.now();
      const result = await providerAPI(page, "/me", "GET", undefined, token);
      return { status: result.status, duration: Date.now() - start };
    }),
  );
  expect(requests.every((request) => request.status === "fulfilled")).toBe(true);
  const samples = requests.map((request) => {
    if (request.status !== "fulfilled") throw new Error("Real provider capacity request failed");
    return request.value;
  });
  const durations = samples.map((sample) => sample.duration);
  durations.sort((a, b) => a - b);
  const p95 = durations[Math.ceil(durations.length * 0.95) - 1]!;
  info.annotations.push({
    type: "flux-provider-metric",
    description: JSON.stringify({ samples: samples.length, targetRps: 2, p95Ms: p95 }),
  });
  for (const sample of samples) expect(sample.status).toBe(200);
  expect(p95 < 2000).toBe(true);
  providerRow(info, "provider_quota_latency");
});

test("production Go fails closed during owned provider egress outage", async ({ page }, info) => {
  const token = await providerSession(page);
  expect((await providerAPI(page, "/me", "GET", undefined, token)).status).toBe(200);
  try {
    const readiness = await providerControl("disconnect-provider");
    expect(readiness.database_ready === true && readiness.redis_ready === true).toBe(true);
    const started = Date.now();
    expect((await providerAPI(page, "/me", "GET", undefined, token)).status).toBe(503);
    expect(Date.now() - started < 4000).toBe(true);
  } finally {
    await providerControl("reconnect-provider");
  }
  expect((await providerAPI(page, "/me")).status).toBe(200);
  providerRow(info, "provider_timeout");
});

test("real cookies and signing-key controls keep production inspection explicit", async ({
  page,
}, info) => {
  const token = await providerSession(page);
  expect((await providerAPI(page, "/me", "GET", undefined, token)).status).toBe(200);
  const jwks = await providerBAPI("/v1/jwks");
  expect(Array.isArray(jwks.keys) && jwks.keys.length > 0).toBe(true);
  const cookies = await page.context().cookies();
  expect(cookies.some((cookie) => cookie.name === "__session" && cookie.value.length > 0)).toBe(
    true,
  );
  // HTTP development cookies cannot prove production HTTPS attributes.
  providerRow(info, "production_cookie_inspection", "human_pending");
  providerRow(info, "provider_key_rotation", "human_pending");
});
