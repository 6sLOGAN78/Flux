import { describe, test, expect, mock } from "bun:test";
import worker from "./redirect.js";

describe("Cloudflare Worker Edge Redirect Engine", () => {
  const mockKV = {
    get: mock(async (key) => {
      if (key === "slug:openai" || key === "openai") {
        return "https://openai.com";
      }
      return null;
    }),
  };

  const mockQueue = {
    send: mock(async (data) => {}),
  };

  const createEnv = () => ({
    LINK_KV: mockKV,
    ANALYTICS_QUEUE: mockQueue,
  });

  const createCtx = () => {
    const promises = [];
    return {
      waitUntil: mock((promise) => {
        promises.push(promise);
      }),
      promises,
    };
  };

  test("should redirect root path to default fallback https://flux.dev with 302", async () => {
    const request = new Request("https://flux.dev/");
    const env = createEnv();
    const ctx = createCtx();

    const response = await worker.fetch(request, env, ctx);
    expect(response.status).toBe(302);
    expect(response.headers.get("Location")).toBe("https://flux.dev");
  });

  test("should redirect matching slug in KV to target URL with 302 and queue analytics event", async () => {
    const request = new Request("https://flux.dev/openai", {
      headers: {
        "cf-connecting-ip": "1.2.3.4",
        "cf-ipcountry": "US",
        "user-agent": "Mozilla/5.0",
        referer: "https://google.com",
      },
    });
    const env = createEnv();
    const ctx = createCtx();

    const response = await worker.fetch(request, env, ctx);
    expect(response.status).toBe(302);
    expect(response.headers.get("Location")).toBe("https://openai.com");

    expect(ctx.waitUntil).toHaveBeenCalled();
    expect(mockQueue.send).toHaveBeenCalledWith(
      expect.objectContaining({
        slug: "openai",
        ip: "1.2.3.4",
        country: "US",
        userAgent: "Mozilla/5.0",
        referrer: "https://google.com",
      })
    );
  });

  test("should return 404 when slug is not found in KV", async () => {
    const request = new Request("https://flux.dev/nonexistent");
    const env = createEnv();
    const ctx = createCtx();

    const response = await worker.fetch(request, env, ctx);
    expect(response.status).toBe(404);
    const body = await response.text();
    expect(body).toBe("Link Not Found");
  });
});
