/**
 * Cloudflare Worker Multi-Region Edge Redirect Engine
 * High-performance edge redirect execution with KV lookups & non-blocking analytics streaming.
 * @see docs/core/redirect_engine.md#part-iii
 * @see ai/PERFORMANCE.md
 */

export default {
  /**
   * Fetch event handler for Cloudflare Worker edge environment.
   * @param {Request} request
   * @param {object} env - Environment bindings (LINK_KV, ANALYTICS_QUEUE)
   * @param {object} ctx - ExecutionContext (waitUntil)
   * @returns {Promise<Response>}
   */
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const slug = url.pathname.slice(1);

    if (!slug) {
      return Response.redirect("https://flux.dev", 302);
    }

    // 1. Edge KV Lookup (<2ms)
    let targetUrl = null;
    if (env && env.LINK_KV) {
      targetUrl = await env.LINK_KV.get(`slug:${slug}`);
      if (!targetUrl) {
        targetUrl = await env.LINK_KV.get(slug);
      }
    }

    if (!targetUrl) {
      return new Response("Link Not Found", { status: 404 });
    }

    // 2. Async Non-Blocking Click Event Queue (<1ms)
    if (env && env.ANALYTICS_QUEUE && ctx && typeof ctx.waitUntil === "function") {
      ctx.waitUntil(
        env.ANALYTICS_QUEUE.send({
          slug,
          timestamp: new Date().toISOString(),
          ip: request.headers.get("cf-connecting-ip") || request.headers.get("x-forwarded-for"),
          country: request.headers.get("cf-ipcountry"),
          userAgent: request.headers.get("user-agent"),
          referrer: request.headers.get("referer"),
        })
      );
    }

    // 3. Instant 302 Redirect
    return Response.redirect(targetUrl, 302);
  },
};
