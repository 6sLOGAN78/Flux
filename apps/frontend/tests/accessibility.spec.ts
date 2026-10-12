import { expect, type Page, test } from "@playwright/test";
import { providerAssetCache } from "./provider-assets";
const publicProviderAsset = providerAssetCache();

const providerAssets = new WeakMap<Page, Set<Promise<void>>>();

// Only the browser harness substitutes FAPI responses. The real SDK and native
// UI still render; local transport proofs never count as live factor acceptance.
export const installProviderTransport = async (
  page: Page,
  sessionClient?: Record<string, unknown>,
  completedSessionClient?: Record<string, unknown>,
) => {
  let pending = providerAssets.get(page);
  if (!pending) {
    pending = new Set();
    providerAssets.set(page, pending);
  }
  const assets = pending;
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
  let signIn: Record<string, unknown> | null = null;
  let signUp: Record<string, unknown> | null = null;
  const client: Record<string, unknown> = sessionClient ?? {
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
    if (
      /\/sessions\/[^/]+\/(?:remove|end)$/.test(url.pathname) ||
      (route.request().method() === "DELETE" &&
        ["/v1/client/sessions", "/v1/client"].includes(url.pathname))
    ) {
      client.sessions = [];
      client.last_active_session_id = null;
    }
    if (url.pathname.startsWith("/npm/")) {
      // Fetch provider-owned SDK assets, not fake application auth controls.
      const asset = (async () => {
        const pinnedURL = `https://cdn.jsdelivr.net${url.pathname.replace("@clerk/clerk-js@6/", "@clerk/clerk-js@6.38.1/").replace("@clerk/ui@1/", "@clerk/ui@1.39.1/")}${url.search}`;
        const response = await publicProviderAsset(pinnedURL, async () => {
          const fetched = await route.fetch({
            url: pinnedURL,
            timeout: 15000,
            maxRedirects: 0,
            headers: { accept: "*/*" },
          });
          try {
            const contentType = fetched.headers()["content-type"] ?? "";
            if (Number(fetched.headers()["content-length"] ?? 0) > 8 * 1024 * 1024)
              throw new Error("Provider asset size limit");
            return { status: fetched.status(), contentType, body: await fetched.body() };
          } finally {
            await fetched.dispose();
          }
        });
        await route.fulfill({
          status: response.status,
          contentType: response.contentType,
          body: response.body,
        });
      })();
      assets.add(asset);
      try {
        await asset;
      } finally {
        assets.delete(asset);
      }
      return;
    }
    if (url.pathname.includes("/client/sign_ins") && route.request().method() === "POST") {
      signIn = {
        object: "sign_in",
        id: "si_local",
        status: "needs_first_factor",
        supported_identifiers: ["email_address"],
        supported_first_factors: [
          { strategy: "password" },
          {
            strategy: "reset_password_email_code",
            email_address_id: "idn_local",
            safe_identifier: "local@example.test",
          },
        ],
        supported_second_factors: [],
        first_factor_verification: null,
        second_factor_verification: null,
        identifier: "local@example.test",
        created_session_id: null,
        abandon_at: Date.now() + 600000,
      };
      client.sign_in = signIn;
      const form = new URLSearchParams(route.request().postData() ?? "");
      if (
        (url.pathname.includes("attempt_first_factor") || form.has("password")) &&
        completedSessionClient
      ) {
        Object.assign(client, completedSessionClient);
        signIn.status = "complete";
        signIn.created_session_id = completedSessionClient.last_active_session_id;
        signIn.first_factor_verification = {
          status: "verified",
          strategy: "password",
          attempts: 1,
          expire_at: null,
          error: null,
        };
        client.sign_in = signIn;
      }
    }
    if (url.pathname.includes("/client/sign_ups") && route.request().method() === "POST") {
      signUp = {
        object: "sign_up",
        id: "su_local",
        status: "missing_requirements",
        required_fields: ["email_address", "password"],
        optional_fields: [],
        missing_fields: [],
        unverified_fields: ["email_address"],
        email_address: "local@example.test",
        username: null,
        first_name: null,
        last_name: null,
        phone_number: null,
        web3_wallet: null,
        external_account: null,
        external_account_strategy: null,
        has_password: true,
        unsafe_metadata: {},
        created_session_id: null,
        created_user_id: null,
        abandon_at: Date.now() + 600000,
        legal_accepted_at: null,
        locale: "en",
        timezone: null,
        verifications: {
          email_address: {
            status: "unverified",
            strategy: "email_code",
            supported_strategies: ["email_code"],
            next_action: "needs_attempt",
            attempts: 0,
            expire_at: Date.now() + 600000,
            error: null,
          },
          phone_number: null,
          web3_wallet: null,
          external_account: null,
        },
      };
      client.sign_up = signUp;
    }
    const sessions = client.sessions as Record<string, unknown>[] | undefined;
    if (url.pathname.includes("/tokens") && sessions?.[0]) {
      const protocol = sessions[0].id === "sess_browser_roles" ? "/member-client" : "/client";
      const refreshed = await page.request.get(`${process.env.FLUX_BROWSER_FIXTURE}${protocol}`);
      expect(refreshed.status()).toBe(200);
      const issued = await refreshed.json();
      expect(issued.client.sessions[0].id).toBe(sessions[0].id);
      sessions[0].last_active_token = issued.client.sessions[0].last_active_token;
    }
    const response =
      url.pathname.includes("/tokens") && sessions?.[0]
        ? sessions[0].last_active_token
        : /\/client\/sessions\/[^/]+\/touch$/.test(url.pathname) && sessions?.[0]
          ? sessions[0]
          : url.pathname.includes("/client/sign_ups")
            ? signUp
            : url.pathname === "/v1/environment"
              ? environment
              : url.pathname.includes("/client/sign_ins")
                ? signIn
                : client;
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

export const finishTenantTransport = async (page: Page) => {
  await page.goto("about:blank", { waitUntil: "commit" });
  const assets = providerAssets.get(page);
  while (assets?.size) await Promise.all([...assets]);
  await page.unrouteAll({ behavior: "wait" });
  providerAssets.delete(page);
};

test.beforeEach(async ({ request }) => {
  expect((await request.post(`${process.env.FLUX_BROWSER_FIXTURE}/restore-reset`)).status()).toBe(
    204,
  );
});
test.afterEach(async ({ page }) => finishTenantTransport(page));

async function accessibleWorkspace(
  page: Page,
  request: import("@playwright/test").APIRequestContext,
) {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, {
    ...fixture.client,
    cookie_expires_at: Date.now() + 30 * 86400000,
  });
  const headers = {
    Authorization: `Bearer ${fixture.token}`,
    Origin: "http://127.0.0.1:3100",
    "Idempotency-Key": crypto.randomUUID(),
  };
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers,
    data: { name: "Keyboard workspace" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
    await route.fulfill({ response });
  });
  return { fixture, workspace, headers };
}

test("accessibility link form identifies invalid fields and focuses a blocking summary by keyboard", async ({
  page,
  request,
}) => {
  const { workspace } = await accessibleWorkspace(page, request);
  await page.goto(`/workspaces/${workspace.id}/links/new`);
  await expect(page.getByRole("button", { name: "Create link", exact: true })).toBeEnabled();
  const destination = page.getByLabel("Destination URL", { exact: true });
  await destination.focus();
  await page.keyboard.type("invalid");
  await page.keyboard.press("Enter");
  await expect(
    page.getByRole("region", { name: "Create link", exact: true }).getByRole("alert"),
  ).toBeFocused();
  await expect(destination).toHaveAttribute("aria-invalid", "true");
  await expect(destination).toHaveAccessibleDescription(/HTTP or HTTPS/);
  await expect(destination).toHaveAttribute("autocomplete", "url");
  await destination.focus();
  await page.keyboard.press("ControlOrMeta+A");
  await page.keyboard.type("https://example.com/keyboard");
  await page.keyboard.press("Tab");
  await expect(page.getByLabel("Title", { exact: true })).toBeFocused();
  await page.keyboard.type("Keyboard link");
  await page.keyboard.press("Tab");
  await expect(page.getByLabel("Custom short key", { exact: true })).toBeFocused();
  await page.keyboard.press("Tab");
  await page.keyboard.press("Enter");
  await expect(page.getByText("Link created.", { exact: true })).toBeVisible();
  await expect(page.getByRole("heading", { level: 1 })).toHaveCount(1);
  await expect(
    page.getByRole("main").getByRole("heading", { name: "Link details", exact: true }),
  ).toBeFocused();
  expect(
    await page
      .locator("[tabindex]")
      .evaluateAll((nodes) => nodes.some((node) => Number(node.getAttribute("tabindex")) > 0)),
  ).toBe(false);
});

test("accessibility delayed team loading exposes inert skeleton and keyboard cancellation with retry", async ({
  page,
  request,
}) => {
  const { workspace } = await accessibleWorkspace(page, request);
  let release: (() => void) | undefined;
  const held = new Promise<void>((resolve) => {
    release = resolve;
  });
  let intercepted = 0;
  await page.route(`**/api/v1/workspaces/${workspace.id}/members`, async (route) => {
    intercepted += 1;
    if (intercepted === 1) await held;
    await route.fallback();
  });
  try {
    await page.goto(`/workspaces/${workspace.id}/team`);
    const team = page.getByRole("region", { name: "Team", exact: true });
    await expect(team).toHaveAttribute("aria-busy", "true");
    await expect(team.locator(".loading-rows")).toHaveAttribute("inert", "");
    await expect(team.locator(".loading-rows")).toHaveAttribute("aria-hidden", "true");
    await expect(
      team.getByText("This is taking longer than expected. You can retry.", { exact: true }),
    ).toBeVisible();
    await team.getByRole("button", { name: "Cancel loading", exact: true }).focus();
    await page.keyboard.press("Enter");
    await expect(team).toHaveAttribute("aria-busy", "false");
    const retry = team.getByRole("button", { name: "Retry loading team", exact: true });
    await retry.focus();
    release?.();
    await page.keyboard.press("Enter");
    await expect(team.getByText("No more members.", { exact: true })).toBeVisible();
    expect(intercepted).toBe(2);
  } finally {
    release?.();
  }
});

test("accessibility native destructive and reason dialogs trap keyboard focus and return safely", async ({
  page,
  request,
}) => {
  const { workspace, fixture, headers } = await accessibleWorkspace(page, request);
  const created = await request.post(`${fixture.api}/api/v1/workspaces/${workspace.id}/links`, {
    headers: { ...headers, "Idempotency-Key": crypto.randomUUID() },
    data: { destination: "https://example.com/dialog", title: "Keyboard dialog" },
  });
  expect(created.status()).toBe(201);
  const { link } = await created.json();
  await page.goto(`/workspaces/${workspace.id}/links/${link.id}`);
  const archive = page
    .getByRole("region", { name: "Link details", exact: true })
    .locator('[aria-label="Lifecycle actions"]')
    .getByRole("button", { name: "Archive link", exact: true });
  await archive.focus();
  await page.keyboard.press("Enter");
  const dialog = page.getByRole("dialog");
  await expect(dialog.getByRole("button", { name: "Keep link", exact: true })).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(dialog.getByRole("button", { name: "Archive link", exact: true })).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(dialog.getByRole("button", { name: "Keep link", exact: true })).toBeFocused();
  await archive.evaluate((element) => element.focus());
  await expect(dialog.getByRole("button", { name: "Keep link", exact: true })).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(archive).toBeFocused();
  const suspend = page.getByRole("button", { name: "Suspend link", exact: true });
  await suspend.focus();
  await page.keyboard.press("Enter");
  await expect(dialog.getByRole("button", { name: "Suspend link", exact: true })).toBeDisabled();
  await expect(dialog.getByLabel("Reason", { exact: true })).toHaveAttribute(
    "aria-invalid",
    "true",
  );
  await page.keyboard.press("Shift+Tab");
  await expect(dialog.getByLabel("Reason", { exact: true })).toBeFocused();
  await page.keyboard.type("Reviewed by keyboard");
  await page.keyboard.press("Tab");
  await page.keyboard.press("Tab");
  await page.keyboard.press("Enter");
  await expect(page.getByText("Link updated.", { exact: true })).toBeVisible();
  await expect(page.getByText("Link updated.", { exact: true })).toHaveAttribute("role", "status");
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
  await page.getByRole("button", { name: "Copy short URL", exact: true }).focus();
  await page.keyboard.press("Enter");
  await expect(page.getByText("Short URL copied.", { exact: true })).toHaveAttribute(
    "role",
    "status",
  );
});

test("accessibility workspace skip link and invitation form retain native labels and busy semantics", async ({
  page,
  request,
}) => {
  const { workspace } = await accessibleWorkspace(page, request);
  expect(
    (
      await request.post(
        `${process.env.FLUX_BROWSER_FIXTURE}/restore-member?workspace=${workspace.id}`,
      )
    ).status(),
  ).toBe(204);
  await page.goto(`/workspaces/${workspace.id}/team`);
  await expect(page.getByRole("button", { name: "Invite member", exact: true })).toBeEnabled();
  const skip = page.getByRole("link", { name: "Skip to main content", exact: true });
  await skip.focus();
  await page.keyboard.press("Enter");
  await expect(page.getByRole("main")).toBeFocused();
  await expect(page.getByRole("link", { name: "Team", exact: true })).toHaveAttribute(
    "aria-current",
    "page",
  );
  const colleague = page.getByRole("row").filter({ hasText: "colleague@example.test" });
  await colleague.getByLabel("New role").focus();
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("Tab");
  const change = colleague.getByRole("button", { name: "Change role", exact: true });
  await expect(change).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(page.getByRole("button", { name: "Keep current role", exact: true })).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(change).toBeFocused();
  const remove = colleague.getByRole("button", { name: "Remove member", exact: true });
  await remove.focus();
  await page.keyboard.press("Enter");
  await expect(page.getByRole("button", { name: "Keep member", exact: true })).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(remove).toBeFocused();
  await expect(page.getByLabel("Email address", { exact: true })).toHaveAttribute(
    "autocomplete",
    "email",
  );
  await page.getByRole("button", { name: "Invite member", exact: true }).focus();
  await page.keyboard.press("Enter");
  await expect(
    page.getByRole("region", { name: "Invitations", exact: true }).getByRole("alert"),
  ).toBeFocused();
  await expect(page.getByLabel("Email address", { exact: true })).toHaveAttribute(
    "aria-invalid",
    "true",
  );
  await page.getByLabel("Email address", { exact: true }).focus();
  await page.keyboard.type("keyboard@example.test");
  await page.keyboard.press("Tab");
  await expect(page.getByLabel("Role", { exact: true })).toBeFocused();
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("Tab");
  await expect(page.getByRole("button", { name: "Invite member", exact: true })).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(page.getByText("Invitation queued", { exact: true })).toHaveAttribute(
    "role",
    "status",
  );
  const revoke = page.getByRole("button", { name: "Revoke", exact: true });
  await revoke.focus();
  await page.keyboard.press("Enter");
  await expect(page.getByRole("button", { name: "Keep invitation", exact: true })).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(revoke).toBeFocused();
});

test("accessibility held link creation keeps busy controls disabled and announces only committed success", async ({
  page,
  request,
}) => {
  const { workspace } = await accessibleWorkspace(page, request);
  let release: (() => void) | undefined;
  const held = new Promise<void>((resolve) => {
    release = resolve;
  });
  let submitted = 0;
  await page.route(`**/api/v1/workspaces/${workspace.id}/links`, async (route) => {
    if (route.request().method() === "POST") {
      submitted += 1;
      await held;
    }
    await route.fallback();
  });
  try {
    await page.goto(`/workspaces/${workspace.id}/links/new`);
    await expect(page.getByRole("button", { name: "Create link", exact: true })).toBeEnabled();
    await page.getByLabel("Destination URL", { exact: true }).focus();
    await page.keyboard.type("https://example.com/held");
    await page.keyboard.press("Enter");
    const form = page.getByRole("region", { name: "Create link", exact: true }).locator("form");
    await expect(form).toHaveAttribute("aria-busy", "true");
    const submit = page.getByRole("button", { name: "Creating link…", exact: true });
    await expect(submit).toBeDisabled();
    await expect(submit).toHaveAccessibleDescription(
      "Wait for link creation to finish before submitting again.",
    );
    await expect(page.getByLabel("Destination URL", { exact: true })).toBeDisabled();
    await page.keyboard.press("Enter");
    await page.keyboard.press("Enter");
    await expect.poll(() => submitted).toBe(1);
    await expect(page.getByText("Link created.", { exact: true })).toHaveCount(0);
    release?.();
    await expect(page.getByText("Link created.", { exact: true })).toHaveAttribute(
      "role",
      "status",
    );
    expect(submitted).toBe(1);
  } finally {
    release?.();
  }
});

test("responsive native workspace reflows library, team and detail at 320, 768 and 1280", async ({
  page,
  request,
}) => {
  const { workspace, fixture, headers } = await accessibleWorkspace(page, request);
  const destination = `https://example.com/${"long-path".repeat(80)}`;
  const response = await request.post(`${fixture.api}/api/v1/workspaces/${workspace.id}/links`, {
    headers: { ...headers, "Idempotency-Key": crypto.randomUUID() },
    data: { destination, title: "Readable responsive link" },
  });
  expect(response.status()).toBe(201);
  const { link } = await response.json();
  for (const width of [1280, 768, 320]) {
    await page.setViewportSize({ width, height: 720 });
    await page.goto(`/workspaces/${workspace.id}/links`);
    await expect(page.getByText("No more links.", { exact: true })).toBeVisible();
    const main = page.getByRole("main");
    const geometry = await main.evaluate((element) => {
      const rect = element.getBoundingClientRect();
      return { left: rect.left, width: rect.width, padding: getComputedStyle(element).paddingLeft };
    });
    if (width === 1280) {
      expect(geometry.left).toBe(224);
      expect(geometry.width).toBeGreaterThan(448);
      expect(geometry.padding).toBe("24px");
    } else if (width === 320) expect(geometry.padding).toBe("16px");
    await expect(
      page.getByRole(width < 768 ? "list" : "table", { name: "Links library", exact: true }),
    ).toBeVisible();
    await assertPageFits(page, width);
    await page.goto(`/workspaces/${workspace.id}/team`);
    await expect(page.getByText("No more members.", { exact: true })).toBeVisible();
    await expect(
      page.getByRole(width < 768 ? "list" : "table", { name: "Team members", exact: true }),
    ).toBeVisible();
    await assertPageFits(page, width);
    await page.goto(`/workspaces/${workspace.id}/links/${link.id}`);
    await expect(page.getByText(destination, { exact: true })).toBeVisible();
    await assertPageFits(page, width);
    await expect(
      page.getByText(
        "Link management is available. Redirects and analytics are not available yet.",
        { exact: true },
      ),
    ).toBeVisible();
    expect(await page.locator("time").first().getAttribute("datetime")).toBeTruthy();
    expect(await page.locator("time").first().textContent()).toMatch(/\d.*(?:UTC|GMT)/);
  }
});

async function assertPageFits(page: Page, width: number) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
    width,
  );
  const controls = await page
    .getByRole("main")
    .locator("button, input, select, a")
    .evaluateAll((nodes) =>
      nodes
        .filter((node) => node.getClientRects().length)
        .map((node) => {
          const rect = node.getBoundingClientRect();
          return { width: rect.width, height: rect.height };
        }),
    );
  expect(controls.length).toBeGreaterThan(0);
  for (const control of controls) {
    expect(control.height).toBeGreaterThanOrEqual(44);
    expect(control.width).toBeGreaterThanOrEqual(44);
  }
}

test("responsive native text enlargement and 400 percent equivalent retain scrollable keyboard dialogs", async ({
  page,
  request,
}) => {
  const { workspace, fixture, headers } = await accessibleWorkspace(page, request);
  const response = await request.post(`${fixture.api}/api/v1/workspaces/${workspace.id}/links`, {
    headers: { ...headers, "Idempotency-Key": crypto.randomUUID() },
    data: { destination: "https://example.com/zoom", title: "Zoom keyboard link" },
  });
  expect(response.status()).toBe(201);
  const { link } = await response.json();
  await page.setViewportSize({ width: 320, height: 256 });
  await page.goto(`/workspaces/${workspace.id}/links/${link.id}`);
  await expect(page.getByRole("button", { name: "Suspend link", exact: true })).toBeEnabled();
  await page.addStyleTag({ content: ":root { font-size: 200%; }" });
  expect(await page.getByRole("main").evaluate((node) => getComputedStyle(node).fontSize)).toBe(
    "32px",
  );
  await assertPageFits(page, 320);
  const trigger = page.getByRole("button", { name: "Suspend link", exact: true });
  await trigger.focus();
  await page.keyboard.press("Enter");
  const dialog = page.getByRole("dialog");
  await expect(dialog.getByRole("button", { name: "Keep link", exact: true })).toBeFocused();
  const geometry = await dialog.evaluate((element) => ({
    height: element.getBoundingClientRect().height,
    scroll: element.scrollHeight,
    client: element.clientHeight,
    overflow: getComputedStyle(element).overflowY,
  }));
  expect(geometry.height).toBeLessThanOrEqual(224);
  expect(geometry.scroll).toBeGreaterThan(geometry.client);
  expect(geometry.overflow).toBe("auto");
  const focused = await page.evaluate(() => {
    const rect = document.activeElement!.getBoundingClientRect();
    return { top: rect.top, bottom: rect.bottom };
  });
  expect(focused.top).toBeGreaterThanOrEqual(0);
  expect(focused.bottom).toBeLessThanOrEqual(256);
  await page.keyboard.press("Escape");
  await expect(trigger).toBeFocused();
});

test("responsive native computed contrast, forced colors and reduced motion retain visible focus", async ({
  page,
  request,
}) => {
  const { workspace } = await accessibleWorkspace(page, request);
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.name));
  await page.goto(`/workspaces/${workspace.id}/links/new`);
  const submit = page.getByRole("button", { name: "Create link", exact: true });
  await expect(submit).toBeEnabled();
  await expect(page).toHaveTitle(/Flux/);
  await expect(page.getByRole("heading", { name: "Create link", exact: true })).toBeVisible();
  await expect(page.locator("nextjs-portal")).toHaveCount(0);
  const colors = await submit.evaluate((node) => {
    const style = getComputedStyle(node);
    return { text: style.color, background: style.backgroundColor, border: style.borderColor };
  });
  expect(contrast(colors.text, colors.background)).toBeGreaterThanOrEqual(4.5);
  expect(contrast(colors.border, "rgb(255, 255, 255)")).toBeGreaterThanOrEqual(3);
  const body = await page
    .getByRole("main")
    .evaluate((node) => ({
      text: getComputedStyle(node).color,
      background: getComputedStyle(node).backgroundColor,
    }));
  expect(contrast(body.text, body.background)).toBeGreaterThanOrEqual(4.5);
  await page.getByLabel("Destination URL", { exact: true }).focus();
  await page.keyboard.press("Tab");
  const focus = await page.getByLabel("Title", { exact: true }).evaluate((node) => {
    const style = getComputedStyle(node);
    const rect = node.getBoundingClientRect();
    return {
      color: style.outlineColor,
      width: style.outlineWidth,
      kind: style.outlineStyle,
      top: rect.top,
      bottom: rect.bottom,
    };
  });
  expect(focus.width).toBe("2px");
  expect(focus.kind).toBe("solid");
  expect(contrast(focus.color, "rgb(255, 255, 255)")).toBeGreaterThanOrEqual(3);
  expect(focus.top).toBeGreaterThanOrEqual(0);
  expect(focus.bottom).toBeLessThanOrEqual(720);
  await page.emulateMedia({ forcedColors: "active", reducedMotion: "reduce" });
  expect(await page.evaluate(() => matchMedia("(forced-colors: active)").matches)).toBe(true);
  const forced = await page.getByLabel("Title", { exact: true }).evaluate((node) => {
    const style = getComputedStyle(node);
    return {
      outline: style.outlineStyle,
      width: style.outlineWidth,
      duration: style.transitionDuration,
      animation: style.animationName,
      scroll: style.scrollBehavior,
      color: style.outlineColor,
      background: style.backgroundColor,
    };
  });
  expect(forced.outline).toBe("solid");
  expect(forced.width).toBe("2px");
  expect(forced.color).not.toBe(forced.background);
  expect(forced.duration).toBe("0s");
  expect(forced.animation).toBe("none");
  expect(forced.scroll).toBe("auto");
  expect(errors).toEqual([]);
});

function contrast(first: string, second: string) {
  const luminance = (color: string) => {
    const values = color
      .match(/\d+(?:\.\d+)?/g)
      ?.slice(0, 3)
      .map((value) => {
        const channel = Number(value) / 255;
        return channel <= 0.04045 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4;
      });
    if (!values || values.length !== 3) throw new Error("Unsupported computed color");
    return values[0]! * 0.2126 + values[1]! * 0.7152 + values[2]! * 0.0722;
  };
  const a = luminance(first),
    b = luminance(second);
  return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);
}
