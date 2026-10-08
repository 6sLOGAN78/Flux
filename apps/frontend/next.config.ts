import type { NextConfig } from "next";

const readAPIOrigin = () => {
  try {
    return new URL(process.env.FLUX_API_ORIGIN ?? "http://127.0.0.1:8080");
  } catch {
    throw new Error("Invalid API origin configuration");
  }
};
const apiOrigin = readAPIOrigin();
if (
  apiOrigin.username ||
  apiOrigin.password ||
  apiOrigin.pathname !== "/" ||
  apiOrigin.search ||
  apiOrigin.hash ||
  !(
    apiOrigin.protocol === "https:" ||
    (apiOrigin.protocol === "http:" &&
      ["127.0.0.1", "localhost", "[::1]"].includes(apiOrigin.hostname))
  )
) {
  throw new Error("Invalid API origin configuration");
}
const config: NextConfig = {
  distDir: "build",
  poweredByHeader: false,
  agentRules: false,
  async headers() {
    return ["/", "/api/v1/:path*"].map((source) => ({
      source,
      headers: [{ key: "Cache-Control", value: "no-store" }],
    }));
  },
  async rewrites() {
    return [{ source: "/api/v1/:path*", destination: `${apiOrigin.origin}/api/v1/:path*` }];
  },
};
export default config;
