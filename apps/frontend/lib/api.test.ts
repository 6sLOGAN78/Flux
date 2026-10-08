import { expect } from "@playwright/test";
import { test } from "node:test";
import { ApiError, createAPI } from "./api";

const origin = "https://flux.example";
test("adapter sends fresh explicit bearer only to fixed API origin with no ambient credentials", async () => {
  let tokens = 0;
  const requests: [string, RequestInit][] = [];
  const api = createAPI(async () => `signed-${++tokens}`, {
    origin,
    fetch: async (url, init) => {
      requests.push([String(url), init ?? {}]);
      return Response.json({ userId: "durable-user" });
    },
  });
  expect(await api.request("/identity")).toEqual({ userId: "durable-user" });
  await api.request("/identity");
  expect(requests.map(([url]) => url)).toEqual([
    `${origin}/api/v1/identity`,
    `${origin}/api/v1/identity`,
  ]);
  for (const [index, [, init]] of requests.entries()) {
    expect(new Headers(init.headers).get("Authorization")).toBe(`Bearer signed-${index + 1}`);
    expect(init).toMatchObject({
      cache: "no-store",
      credentials: "omit",
      redirect: "error",
      mode: "same-origin",
      referrerPolicy: "no-referrer",
    });
  }
});

test("adapter rejects foreign, traversal, encoded, query and fragment request targets before token access", async () => {
  let called = 0;
  const api = createAPI(
    async () => {
      called++;
      return "credential";
    },
    { origin },
  );
  for (const path of [
    "https://evil.example/",
    "//evil.example",
    "/../identity",
    "/%2e%2e/identity",
    "/identity?token=secret",
    "/identity#secret",
    "/identity\\evil",
    "identity",
    "/",
  ]) {
    await expect(api.request(path)).rejects.toMatchObject({ code: "invalid_request" });
  }
  expect(called).toBe(0);
});

test("missing token fails closed and provider/private error text never escapes", async () => {
  let fetches = 0;
  for (const getToken of [
    async () => null,
    async () => {
      throw new Error("PRIVATE-TOKEN-PROVIDER");
    },
  ]) {
    const api = createAPI(getToken, {
      origin,
      fetch: async () => {
        fetches++;
        return Response.json({});
      },
    });
    try {
      await api.request("/identity");
      throw new Error("Expected rejection");
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError);
      expect(String(error)).not.toContain("PRIVATE");
    }
  }
  expect(fetches).toBe(0);
});

test("HTTP failures expose closed status errors without raw response bodies", async () => {
  for (const status of [400, 401, 403, 404, 409, 429, 500, 503]) {
    const api = createAPI(async () => "credential", {
      origin,
      fetch: async () => new Response("PRIVATE-BODY-CREDENTIAL", { status }),
    });
    try {
      await api.request("/identity");
      throw new Error("Expected rejection");
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError);
      expect(error).toMatchObject({ status });
      expect(String(error)).not.toContain("PRIVATE");
    }
  }
});

test("network errors, malformed JSON and oversized success bodies are bounded safe failures", async () => {
  for (const fetch of [
    async () => {
      throw new Error("PRIVATE-NETWORK");
    },
    async () => new Response("PRIVATE-INVALID-JSON"),
    async () => new Response("x".repeat(65537)),
  ]) {
    const api = createAPI(async () => "credential", { origin, fetch });
    await expect(api.request("/identity")).rejects.toMatchObject({
      code: "unavailable",
      message: "The request could not be completed. Try again.",
    });
  }
});

test("caller cancellation aborts actual fetch and never retries a mutation", async () => {
  let count = 0;
  const controller = new AbortController();
  const api = createAPI(async () => "credential", {
    origin,
    fetch: async (_url, init) => {
      count++;
      return new Promise<Response>((_resolve, reject) =>
        init?.signal?.addEventListener("abort", () => reject(new Error("PRIVATE-ABORT")), {
          once: true,
        }),
      );
    },
  });
  const pending = api.request("/workspaces", {
    method: "POST",
    body: { name: "Private name" },
    signal: controller.signal,
  });
  await new Promise((resolve) => setTimeout(resolve, 10));
  controller.abort();
  await expect(pending).rejects.toMatchObject({ code: "cancelled" });
  expect(count).toBe(1);
});

test("deadline also bounds an unresponsive token provider and pre-aborted calls send nothing", async () => {
  const api = createAPI(() => new Promise(() => {}), { origin, timeoutMs: 15 });
  await expect(api.request("/identity")).rejects.toMatchObject({ code: "unavailable" });
  const controller = new AbortController();
  controller.abort();
  let tokens = 0;
  const cancelled = createAPI(
    async () => {
      tokens++;
      return "credential";
    },
    { origin },
  );
  await expect(cancelled.request("/identity", { signal: controller.signal })).rejects.toMatchObject(
    { code: "cancelled" },
  );
  expect(tokens).toBe(0);
});
