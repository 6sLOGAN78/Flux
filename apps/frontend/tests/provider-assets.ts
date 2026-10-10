type Asset = { status: number; contentType: string; body: Buffer };

// Cache only public, pinned provider JavaScript within this test process.
export function providerAssetCache() {
  const entries = new Map<string, Promise<Asset>>();
  let bytes = 0;
  return async (url: string, fetch: () => Promise<Asset>): Promise<Asset> => {
    const parsed = new URL(url);
    if (
      parsed.origin !== "https://cdn.jsdelivr.net" ||
      parsed.username ||
      parsed.password ||
      parsed.hash ||
      parsed.search.length > 512 ||
      !/^\/npm\/@clerk\/(?:clerk-js@6\.38\.1|ui@1\.39\.1)\/dist\/[A-Za-z0-9_.-]+\.js$/.test(
        parsed.pathname,
      )
    )
      throw new Error("Unapproved provider asset");
    let pending = entries.get(url);
    if (!pending) {
      if (entries.size >= 64) throw new Error("Provider asset cache limit");
      pending = fetch().then((asset) => {
        if (
          asset.status !== 200 ||
          !/^(?:application|text)\/javascript(?:;|$)/i.test(asset.contentType) ||
          asset.body.length === 0 ||
          asset.body.length > 8 * 1024 * 1024 ||
          bytes + asset.body.length > 32 * 1024 * 1024
        )
          throw new Error("Invalid provider asset");
        bytes += asset.body.length;
        return { ...asset, body: Buffer.from(asset.body) };
      });
      entries.set(url, pending);
      pending.catch(() => {
        if (entries.get(url) === pending) entries.delete(url);
      });
    }
    const asset = await pending;
    return { ...asset, body: Buffer.from(asset.body) };
  };
}
