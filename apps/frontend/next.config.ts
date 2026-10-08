import type { NextConfig } from "next";

const apiOrigin = new URL(process.env.FLUX_API_ORIGIN ?? "http://127.0.0.1:8080");
if (
  apiOrigin.username ||
  apiOrigin.password ||
  apiOrigin.pathname !== "/" ||
  apiOrigin.search ||
  apiOrigin.hash ||
  !["http:", "https:"].includes(apiOrigin.protocol)
) {
  throw new Error("Invalid API origin configuration");
}
const config: NextConfig = {
  distDir: "build",
  poweredByHeader: false,
  agentRules: false,
  async rewrites() {
    return [{ source: "/api/v1/:path*", destination: `${apiOrigin.origin}/api/v1/:path*` }];
  },
};
export default config;
