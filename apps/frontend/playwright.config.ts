import { defineConfig, devices } from "@playwright/test";
import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { validateProviderConfig } from "./tests/provider.config";

const providerRun = process.env.FLUX_PROVIDER_RUN === "1";
if (providerRun) {
  const errors = validateProviderConfig(process.env);
  if (errors.length)
    throw new Error(`Provider prerequisites missing or invalid: ${errors.join(", ")}`);
}

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
  globalTimeout: providerRun ? 540000 : 240000,
  retries: 0,
  maxFailures: providerRun ? 1 : 0,
  workers: 1,
  reporter: "json",
  expect: { timeout: 20000 },
  use: {
    baseURL: "http://127.0.0.1:3100",
    trace: "off",
    screenshot: "off",
    video: "off",
  },
  projects: [
    {
      name: "local",
      testMatch: /.*\.spec\.ts/,
      testIgnore: /provider-.*\.spec\.ts/,
      use: { ...devices["Desktop Chrome"] },
    },
    {
      name: "provider",
      timeout: 90000,
      testMatch: /provider-(auth|session|invitation)\.spec\.ts/,
      use: { ...devices["Desktop Chrome"], baseURL: "http://localhost:3000" },
    },
  ],
  webServer: {
    command: providerRun ? "bun run build && bun run start --port 3000" : "bun run dev --port 3100",
    url: providerRun ? "http://localhost:3000/sign-in" : "http://127.0.0.1:3100/sign-in",
    timeout: 90000,
    reuseExistingServer: false,
    stdout: "ignore",
    stderr: "ignore",
    env: {
      NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY: providerRun
        ? process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY!
        : localPublishableKey,
      NEXT_TELEMETRY_DISABLED: "1",
    },
  },
});
