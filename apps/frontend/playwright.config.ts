import { defineConfig, devices } from "@playwright/test";
import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

// Reports can contain signed test tokens even with media capture disabled.
// mkdtemp creates a private directory; only this process-owned path is removed.
const browserOutput = mkdtempSync(join(tmpdir(), "flux-playwright-"));
process.once("exit", () => rmSync(browserOutput, { recursive: true, force: true }));

// Public syntactic development key, confined to the local test server. No real
// provider secret or test-login endpoint is used by this browser harness.
export const localPublishableKey = `pk_test_${Buffer.from("fixture.clerk.accounts.dev$").toString("base64")}`;

export default defineConfig({
  testDir: "./tests",
  outputDir: browserOutput,
  timeout: 45000,
  globalTimeout: 240000,
  retries: 0,
  workers: 1,
  reporter: "json",
  expect: { timeout: 20000 },
  use: {
    baseURL: "http://127.0.0.1:3100",
    trace: "off",
    screenshot: "off",
    video: "off",
  },
  projects: [{ name: "local", use: { ...devices["Desktop Chrome"] } }],
  webServer: {
    command: "bun run dev --port 3100",
    url: "http://127.0.0.1:3100/sign-in",
    timeout: 90000,
    reuseExistingServer: false,
    stdout: "ignore",
    stderr: "ignore",
    env: { NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY: localPublishableKey, NEXT_TELEMETRY_DISABLED: "1" },
  },
});
