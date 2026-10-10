import { test } from "node:test";
import { strict as assert } from "node:assert";
import { providerAssetCache } from "../tests/provider-assets";

const url = "https://cdn.jsdelivr.net/npm/@clerk/ui@1.39.1/dist/ui.js";
const asset = () => ({
  status: 200,
  contentType: "application/javascript; charset=utf-8",
  body: Buffer.from("genuine-fixture-bytes"),
});
test("provider assets deduplicate exact pinned URLs and preserve isolated response bytes", async () => {
  const cached = providerAssetCache();
  let calls = 0;
  const fetch = async () => {
    calls++;
    return asset();
  };
  const [first, second] = await Promise.all([cached(url, fetch), cached(url, fetch)]);
  assert.equal(calls, 1);
  assert.deepEqual(first, second);
  first.body.fill(0);
  assert.deepEqual(await cached(url, fetch), asset());
  await cached(`${url}?variant=2`, fetch);
  assert.equal(calls, 2);
});
test("provider asset failures never poison the cache or admit private/unbounded responses", async () => {
  for (const fetch of [
    async () => {
      throw new Error("fetch failed");
    },
    async () => ({ ...asset(), status: 503 }),
    async () => ({ ...asset(), contentType: "application/json" }),
    async () => ({ ...asset(), body: Buffer.alloc(8 * 1024 * 1024 + 1) }),
  ]) {
    const cached = providerAssetCache();
    await assert.rejects(cached(url, fetch));
    assert.deepEqual(await cached(url, async () => asset()), asset());
  }
  const cached = providerAssetCache();
  for (const rejected of [
    "https://fixture.clerk.accounts.dev/v1/client",
    url.replace("1.39.1", "1"),
    url.replace("cdn.jsdelivr.net", "user:password@cdn.jsdelivr.net"),
  ]) {
    await assert.rejects(
      cached(rejected, async () => {
        throw new Error("must not fetch");
      }),
    );
  }
});
