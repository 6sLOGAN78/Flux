import { expect, type Page, test } from "@playwright/test";

// Only the browser harness substitutes FAPI responses. The real SDK and native
// UI still render; local transport proofs never count as live factor acceptance.
export const installProviderTransport = async (page: Page) => {
  const attribute = (enabled = false, required = false) => ({
    enabled,
    required,
    verifications: [],
    used_for_first_factor: false,
    first_factors: [],
    used_for_second_factor: false,
    second_factors: [],
    verify_at_sign_up: false,
  });
  const environment = {
    object: "environment",
    id: "env_local",
    auth_config: {
      object: "auth_config",
      id: "auth_local",
      single_session_mode: true,
      claimed_at: null,
      reverification: false,
    },
    display_config: {
      object: "display_config",
      id: "display_local",
      application_name: "Flux",
      instance_environment_type: "development",
      sign_in_url: "/sign-in",
      sign_up_url: "/sign-up",
      home_url: "/",
      after_sign_in_url: "/",
      after_sign_up_url: "/",
      preferred_sign_in_strategy: "password",
      show_devmode_warning: false,
      captcha_public_key: null,
      captcha_widget_type: null,
      captcha_provider: "turnstile",
      theme: {},
    },
    user_settings: {
      attributes: Object.fromEntries(
        [
          "email_address",
          "phone_number",
          "username",
          "first_name",
          "last_name",
          "password",
          "web3_wallet",
          "authenticator_app",
          "backup_code",
          "passkey",
        ].map((name) => [
          name,
          name === "email_address"
            ? {
                ...attribute(true, true),
                used_for_first_factor: true,
                verifications: ["email_code"],
                verify_at_sign_up: true,
              }
            : attribute(name === "password", name === "password"),
        ]),
      ),
      social: Object.fromEntries(
        ["google", "github"].map((name) => [
          `oauth_${name}`,
          {
            enabled: true,
            required: false,
            authenticatable: true,
            strategy: `oauth_${name}`,
            name: name === "google" ? "Google" : "GitHub",
            logo_url: null,
          },
        ]),
      ),
      sign_in: { second_factor: { required: false, enabled: false } },
      sign_up: {
        allowlist_only: false,
        progressive: true,
        captcha_enabled: false,
        mode: "public",
        legal_consent_enabled: false,
      },
      actions: { delete_self: true, create_organization: false },
      enterprise_sso: { enabled: false },
      password_settings: {
        min_length: 8,
        max_length: 72,
        require_special_char: false,
        require_numbers: false,
        require_uppercase: false,
        require_lowercase: false,
        show_zxcvbn: false,
        min_zxcvbn_strength: 0,
        disable_hibp: true,
      },
      username_settings: { min_length: 4, max_length: 64 },
      passkey_settings: { allow_autofill: false, show_sign_in_button: false },
    },
    organization_settings: { enabled: false },
    commerce_settings: { billing: { enabled: false } },
    api_keys_settings: { enabled: false },
    protect_config: { enabled: false },
    maintenance_mode: false,
  };
  const client = {
    object: "client",
    id: "client_local",
    sessions: [],
    sign_in: null,
    sign_up: null,
    last_active_session_id: null,
    cookie_expires_at: null,
    created_at: 0,
    updated_at: 0,
  };
  await page.route("https://fixture.clerk.accounts.dev/**", async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname.startsWith("/npm/")) {
      // Fetch provider-owned SDK assets, not fake application auth controls.
      const response = await route.fetch({
        url: `https://cdn.jsdelivr.net${url.pathname}${url.search}`,
      });
      await route.fulfill({ response });
      return;
    }
    const response = url.pathname === "/v1/environment" ? environment : client;
    await route.fulfill({
      contentType: "application/json",
      headers: {
        "access-control-allow-origin": "http://127.0.0.1:3100",
        "access-control-allow-credentials": "true",
      },
      body: JSON.stringify({ response, client }),
    });
  });
};

test.beforeEach(async ({ page }) => {
  await installProviderTransport(page);
});

test("sign-in exposes original auth heading and native provider controls", async ({ page }) => {
  await page.goto("/sign-in");
  await expect(page.getByRole("heading", { name: "Sign in to Flux", exact: true })).toBeVisible();
  await expect(page.getByLabel("Email address", { exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: /Google/ })).toBeVisible();
  await expect(page.getByRole("button", { name: /GitHub/ })).toBeVisible();
});
