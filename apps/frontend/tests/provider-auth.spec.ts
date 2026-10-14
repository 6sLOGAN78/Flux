import { expect, test } from "@playwright/test";
import { providerAPI, providerLogin, providerRow } from "./provider.config";

test("real provider password signup and mandatory verification boundary", async ({
  page,
}, info) => {
  await page.goto("/sign-up");
  await expect(
    page.getByRole("heading", { name: "Create your account", exact: true }),
  ).toBeVisible();
  await page.getByLabel(/email address/i).fill(process.env.FLUX_PROVIDER_SIGNUP_EMAIL!);
  await page.getByLabel(/^password$/i).fill(process.env.FLUX_PROVIDER_SIGNUP_PASSWORD!);
  await page.getByRole("button", { name: /^continue$/i }).click();
  await expect(page.getByText(/verify your email/i).first()).toBeVisible();
  const denied = await page.evaluate(async () => {
    const clerk = (
      window as unknown as {
        Clerk: {
          session: unknown;
          client: {
            signUp: {
              status: string;
              attemptEmailAddressVerification(input: { code: string }): Promise<unknown>;
            };
          };
        };
      }
    ).Clerk;
    try {
      await clerk.client.signUp.attemptEmailAddressVerification({ code: "000000" });
      return false;
    } catch {
      return !clerk.session && clerk.client.signUp.status !== "complete";
    }
  });
  expect(denied).toBe(true);
  const unauthenticated = await page.evaluate(
    async () => (await fetch("/api/v1/me", { cache: "no-store" })).status,
  );
  expect(unauthenticated).toBe(401);
  providerRow(info, "email_password_signup");
  providerRow(info, "email_verification", "human_pending");
});

test("real provider password login reaches production Go identity", async ({ page }, info) => {
  const token = await providerLogin(page);
  const identity = await providerAPI(page, "/me", "GET", undefined, token);
  info.annotations.push({
    type: "flux-provider-api-status",
    description: JSON.stringify({ endpoint: "identity", status: identity.status }),
  });
  expect(identity.status).toBe(200);
  providerRow(info, "password_login");
});

test("real provider recovery control reaches authorized inbox challenge", async ({
  page,
}, info) => {
  await page.goto("/sign-in");
  await page.getByLabel(/email address/i).fill(process.env.FLUX_PROVIDER_TEST_EMAIL!);
  await page.getByRole("button", { name: /^continue$/i }).click();
  await page.getByRole("link", { name: /forgot password/i }).click();
  await page.getByRole("button", { name: "Reset your password", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Reset password", exact: true })).toBeVisible();
  await expect(
    page.getByRole("textbox", { name: "Enter verification code", exact: true }),
  ).toBeVisible();
  await expect
    .poll(() =>
      page.evaluate(
        () =>
          (
            window as unknown as {
              Clerk?: {
                client?: { signIn?: { firstFactorVerification?: { strategy?: string } } };
              };
            }
          ).Clerk?.client?.signIn?.firstFactorVerification?.strategy,
      ),
    )
    .toBe("reset_password_email_code");
  expect(
    await page.evaluate(
      () => !(window as unknown as { Clerk?: { session?: unknown } }).Clerk?.session,
    ),
  ).toBe(true);
  providerRow(info, "password_recovery", "human_pending");
});

for (const [label, id, origin] of [
  ["Google", "google_oauth", "accounts.google.com"],
  ["GitHub", "github_oauth", "github.com"],
] as const) {
  test(`real provider ${label} control starts the configured OAuth flow`, async ({
    page,
  }, info) => {
    await page.goto("/sign-in");
    const control = page.getByRole("button", { name: new RegExp(label, "i") });
    await expect(control).toBeVisible();
    await control.click();
    await page.waitForURL((url) => url.hostname === origin || url.hostname.endsWith(`.${origin}`));
    expect(new URL(page.url()).protocol).toBe("https:");
    providerRow(info, id, "human_pending");
  });
}
