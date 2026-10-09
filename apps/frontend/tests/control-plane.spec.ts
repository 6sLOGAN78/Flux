import { expect, type Page, test } from "@playwright/test";

test("destination rejects reserved addresses with exact safe feedback", async ({ page, request }) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const response = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-destination-01",
    },
    data: { name: "Destination policy tenant" },
  });
  expect(response.status()).toBe(201);
  const { workspace } = await response.json();
  await page.route("**/api/v1/**", async (route) => {
    const response = await route.fetch({ url: `${fixture.api}${new URL(route.request().url()).pathname}` });
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/links/new`);
  await expect(page.getByText("Managed hostname: go.flux.test", { exact: true })).toBeVisible();
  await page.getByLabel("Destination URL", { exact: true }).fill("http://[fec0::1]/private");
  await page.getByRole("button", { name: "Create link", exact: true }).click();
  await expect(page.getByRole("main").getByRole("alert")).toHaveText(
    "This destination isn't allowed. Choose a different public destination.",
  );
  await expect(page).toHaveURL(new RegExp(`/workspaces/${workspace.id}/links/new$`));
  await expect(page.getByLabel("Destination URL", { exact: true })).toHaveValue("http://[fec0::1]/private");
  await page.getByLabel("Destination URL", { exact: true }).fill("https://example.com/a%2Fb?q=a%2Bb&x=1#section");
  await page.getByRole("button", { name: "Create link", exact: true }).click();
  await expect(page.getByRole("link", { name: "Open destination (opens in a new tab)" })).toHaveAttribute(
    "href", "https://example.com/a%2Fb?q=a%2Bb&x=1#section",
  );
  await expect(page.getByRole("link", { name: "Open destination (opens in a new tab)" })).toHaveAttribute("rel", "noreferrer noopener");
  await page.evaluate(() => Object.defineProperty(navigator, "clipboard", { value: { writeText: () => Promise.reject(new Error("denied")) }, configurable: true }));
  await page.getByRole("button", { name: "Copy short URL", exact: true }).click();
  await expect(page.getByText("Couldn't copy. Select and copy the short URL below.", { exact: true })).toBeVisible();
  expect(await page.locator("time[datetime]").count()).toBe(2);
});

test("link-create commits a generated link and opens escaped detail", async ({ page, request }) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const response = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-link-create-01",
    },
    data: { name: "Link browser tenant" },
  });
  expect(response.status()).toBe(201);
  const { workspace } = await response.json();
  await page.route("**/api/v1/**", async (route) => {
    const response = await route.fetch({
      url: `${fixture.api}${new URL(route.request().url()).pathname}`,
    });
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/links/new`);
  await expect(page.getByText("Managed hostname: go.flux.test", { exact: true })).toBeVisible();
  await page.getByLabel("Destination URL", { exact: true }).fill("https://example.com/campaign");
  await page.getByLabel("Title", { exact: true }).fill("<script>protected</script>");
  await page.getByRole("button", { name: "Create link", exact: true }).click();
  await expect(page).toHaveURL(new RegExp(`/workspaces/${workspace.id}/links/[0-9a-f-]{36}`));
  await expect(page.getByText("<script>protected</script>", { exact: true })).toBeVisible();
  await expect(page.getByText("https://example.com/campaign", { exact: true })).toBeVisible();
  await expect(page.getByText(/https:\/\/go.flux.test\/[a-z2-7]{20}/)).toBeVisible();
  await expect(
    page.getByText("Link management is available. Redirects and analytics are not available yet."),
  ).toBeVisible();
  expect(await page.locator("main script").count()).toBe(0);
});

test("link-create dirty draft confirms workspace switch and disposes fields", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const workspaces = [];
  for (const [index, name] of ["Link draft Alpha", "Link draft Beta"].entries()) {
    const response = await request.post(`${fixture.api}/api/v1/workspaces`, {
      headers: {
        Authorization: `Bearer ${fixture.token}`,
        Origin: "http://127.0.0.1:3100",
        "Idempotency-Key": `browser-link-draft-${index}`,
      },
      data: { name },
    });
    expect(response.status()).toBe(201);
    workspaces.push((await response.json()).workspace);
  }
  await page.route("**/api/v1/**", async (route) => {
    const response = await route.fetch({
      url: `${fixture.api}${new URL(route.request().url()).pathname}`,
    });
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspaces[0].id}/links/new`);
  await page
    .getByLabel("Destination URL", { exact: true })
    .fill("https://example.com/private-draft");
  await page.getByLabel("Switch workspace", { exact: true }).selectOption(workspaces[1].id);
  await page.getByRole("button", { name: "Open workspace", exact: true }).click();
  await expect(page.getByRole("dialog", { name: "Discard unsaved changes?" })).toBeVisible();
  await page.getByRole("button", { name: "Stay", exact: true }).click();
  await expect(page.getByLabel("Destination URL", { exact: true })).toHaveValue(
    "https://example.com/private-draft",
  );
  await page.getByRole("button", { name: "Open workspace", exact: true }).click();
  await page.getByRole("button", { name: "Discard", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Link draft Beta", exact: true })).toBeVisible();
  await expect(page.getByLabel("Destination URL", { exact: true })).toHaveCount(0);
  expect(await page.getByRole("main").innerHTML()).not.toContain("private-draft");
});

test("link-create missing detail preserves authorized workspace and removal disposes content", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const response = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-link-missing-01",
    },
    data: { name: "Link missing tenant" },
  });
  expect(response.status()).toBe(201);
  const { workspace } = await response.json();
  await page.route("**/api/v1/**", async (route) => {
    const response = await route.fetch({
      url: `${fixture.api}${new URL(route.request().url()).pathname}`,
    });
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/links/00000000-0000-4000-8000-000000000000`);
  await expect(page.getByRole("main").getByRole("alert")).toHaveText("Link not found");
  await expect(
    page.getByRole("heading", { name: "Link missing tenant", exact: true }),
  ).toBeVisible();
  const removed = await request.post(
    `${process.env.FLUX_BROWSER_FIXTURE}/restore-remove?workspace=${workspace.id}`,
  );
  expect(removed.status()).toBe(204);
  await page.evaluate(() => window.dispatchEvent(new Event("focus")));
  await expect(
    page.getByRole("heading", { name: "Choose a workspace", exact: true }),
  ).toBeVisible();
  await expect(page.getByRole("navigation", { name: "Workspace" })).toHaveCount(0);
});

test("switching commits selection and clears old content before the next workspace", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const workspaces = [];
  for (const [index, name] of ["Switch Alpha", "Switch Beta"].entries()) {
    const response = await request.post(`${fixture.api}/api/v1/workspaces`, {
      headers: {
        Authorization: `Bearer ${fixture.token}`,
        Origin: "http://127.0.0.1:3100",
        "Idempotency-Key": `browser-switch-000${index}`,
      },
      data: { name },
    });
    expect(response.status()).toBe(201);
    workspaces.push((await response.json()).workspace);
  }
  await page.route("**/api/v1/**", async (route) => {
    const response = await route.fetch({
      url: `${fixture.api}${new URL(route.request().url()).pathname}`,
    });
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspaces[0].id}/links`);
  await expect(page.getByRole("heading", { name: "Switch Alpha", exact: true })).toBeVisible();
  await page.getByLabel("Switch workspace", { exact: true }).selectOption(workspaces[1].id);
  await page.getByRole("button", { name: "Open workspace", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Switch Beta", exact: true })).toBeVisible();
  expect(await page.getByRole("main").innerHTML()).not.toContain(workspaces[0].id);
  expect(await page.getByRole("navigation", { name: "Workspace" }).innerHTML()).not.toContain(
    workspaces[0].id,
  );
  await page.getByRole("button", { name: "Sign out", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Sign in to Flux", exact: true })).toBeVisible();
  expect(await page.getByRole("main").innerHTML()).not.toContain(workspaces[1].id);
  await expect(page.getByRole("navigation", { name: "Workspace" })).toHaveCount(0);
});

test("switching two tabs revocation discards a held older authorized response", async ({
  page,
  context,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  const response = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-switch-race-01",
    },
    data: { name: "Revoked Switch" },
  });
  expect(response.status()).toBe(201);
  const { workspace } = await response.json();
  const second = await context.newPage();
  let release = () => {};
  const held = new Promise<void>((resolve) => {
    release = resolve;
  });
  let observed = () => {};
  const started = new Promise<void>((resolve) => {
    observed = resolve;
  });
  let finished = () => {};
  const completed = new Promise<void>((resolve) => {
    finished = resolve;
  });
  let hold = false;
  try {
    for (const tab of [page, second]) {
      await installProviderTransport(tab, fixture.client);
      await tab.route("**/api/v1/**", async (route) => {
        const response = await route.fetch({
          url: `${fixture.api}${new URL(route.request().url()).pathname}`,
        });
        if (tab === page && hold && new URL(route.request().url()).pathname === "/api/v1/me") {
          hold = false;
          observed();
          await held;
          await route.fulfill({ response }).catch(() => {});
          finished();
          return;
        }
        await route.fulfill({ response });
      });
      await tab.goto(`/workspaces/${workspace.id}/links`);
      await expect(tab.getByRole("heading", { name: "Revoked Switch", exact: true })).toBeVisible();
    }
    hold = true;
    await page.evaluate(() => window.dispatchEvent(new Event("focus")));
    await started;
    const removed = await request.post(
      `${process.env.FLUX_BROWSER_FIXTURE}/restore-remove?workspace=${workspace.id}`,
    );
    expect(removed.status()).toBe(204);
    await second.evaluate(() => {
      const channel = new BroadcastChannel("flux.workspace-invalidation");
      channel.postMessage("invalidate");
      channel.close();
      window.dispatchEvent(new Event("focus"));
    });
    for (const tab of [page, second]) {
      await expect(
        tab.getByRole("heading", { name: "Choose a workspace", exact: true }),
      ).toBeVisible();
      await expect(
        tab.getByText("Your workspace access changed. Choose an available workspace.", {
          exact: true,
        }),
      ).toBeVisible();
      expect(await tab.getByRole("main").innerHTML()).not.toContain(workspace.id);
      await expect(tab.getByRole("navigation", { name: "Workspace" })).toHaveCount(0);
    }
    release();
    await completed;
    await expect(
      page.getByRole("heading", { name: "Choose a workspace", exact: true }),
    ).toBeFocused();
    expect(await page.getByRole("main").innerHTML()).not.toContain(workspace.id);
  } finally {
    release();
    await finishProviderTransport(second);
  }
});

test("switching unsaved workspace name offers safe Stay and Discard without submitting", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  let posts = 0;
  await page.route("**/api/v1/**", async (route) => {
    if (route.request().method() === "POST") posts++;
    const response = await route.fetch({
      url: `${fixture.api}${new URL(route.request().url()).pathname}`,
    });
    await route.fulfill({ response });
  });
  await page.goto("/onboarding");
  await page.getByLabel("Workspace name", { exact: true }).fill("Unsaved workspace");
  const choose = page.getByRole("button", { name: "Choose a workspace", exact: true });
  await choose.click();
  await expect(
    page.getByRole("dialog", { name: "Discard unsaved changes?", exact: true }),
  ).toBeVisible();
  await expect(page.getByRole("button", { name: "Stay", exact: true })).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(choose).toBeFocused();
  await expect(page.getByLabel("Workspace name", { exact: true })).toHaveValue("Unsaved workspace");
  await choose.click();
  await page.getByRole("button", { name: "Discard", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Choose a workspace", exact: true }),
  ).toBeVisible();
  expect(posts).toBe(0);
});

test("switching same-workspace transient refresh retains authorized content and missing resource confirms membership", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const response = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-switch-refresh-01",
    },
    data: { name: "Refresh Switch" },
  });
  expect(response.status()).toBe(201);
  const { workspace } = await response.json();
  let unavailable = false;
  let missing = false;
  await page.route("**/api/v1/**", async (route) => {
    const path = new URL(route.request().url()).pathname;
    if (unavailable && path === "/api/v1/me") {
      await route.fulfill({ status: 503, body: "{}" });
      return;
    }
    if (missing && path === `/api/v1/workspaces/${workspace.id}`) {
      await route.fulfill({ status: 404, body: "{}" });
      return;
    }
    const response = await route.fetch({ url: `${fixture.api}${path}` });
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/links`);
  await expect(page.getByRole("heading", { name: "Refresh Switch", exact: true })).toBeVisible();
  unavailable = true;
  await page.evaluate(() => window.dispatchEvent(new Event("focus")));
  await expect(page.getByRole("main").getByRole("alert")).toHaveText(
    "Workspace could not be loaded. Try again.",
  );
  await expect(page.getByRole("heading", { name: "Refresh Switch", exact: true })).toBeVisible();
  unavailable = false;
  missing = true;
  await page.evaluate(() => window.dispatchEvent(new Event("focus")));
  await expect(page.getByRole("main").getByRole("alert")).toHaveCount(0);
  await expect(page.getByRole("heading", { name: "Refresh Switch", exact: true })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Choose a workspace", exact: true })).toHaveCount(
    0,
  );
});

test("restore uses committed authorized selection and never trusts browser values", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-restore-0001",
    },
    data: { name: "Restore browser" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  await page.route("**/api/v1/**", async (route) => {
    const response = await route.fetch({
      url: `${fixture.api}${new URL(route.request().url()).pathname}`,
    });
    await route.fulfill({ response });
  });
  await page.goto("/workspaces");
  await expect(page.getByRole("heading", { name: "Choose a workspace" })).toBeVisible();
  await page.getByLabel("Workspace", { exact: true }).selectOption(workspace.id);
  await page.getByRole("button", { name: "Open workspace", exact: true }).click();
  await expect(page).toHaveURL(new RegExp(`/workspaces/${workspace.id}/links$`));
  await expect(page.getByRole("heading", { name: "Restore browser" })).toBeVisible();
  await page.evaluate(() => {
    localStorage.setItem("flux.workspace", "00000000-0000-4000-8000-000000000000");
    window.dispatchEvent(
      new MessageEvent("message", {
        data: { workspaceId: "00000000-0000-4000-8000-000000000000" },
      }),
    );
  });
  await page.goto("/");
  await expect(page).toHaveURL(new RegExp(`/workspaces/${workspace.id}/links$`));
  await expect(page.getByRole("heading", { name: "Restore browser" })).toBeVisible();
  const removed = await request.post(
    `${process.env.FLUX_BROWSER_FIXTURE}/restore-remove?workspace=${workspace.id}`,
  );
  expect(removed.status()).toBe(204);
  await page.goto("/");
  await expect(page).toHaveURL(/\/workspaces$/);
  await expect(page.getByRole("heading", { name: "Choose a workspace" })).toBeVisible();
  await expect(page.getByRole("option", { name: /Restore browser/ })).toHaveCount(0);
  expect(await page.getByRole("main").innerHTML()).not.toContain(workspace.id);
});

test("workspace requires an explicit name and opens authorized Links after commit", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const keys: string[] = [];
  await page.route("**/api/v1/**", async (route) => {
    const response = await route.fetch({
      url: `${fixture.api}${new URL(route.request().url()).pathname}`,
    });
    if (route.request().method() === "POST") {
      keys.push(route.request().headers()["idempotency-key"] ?? "");
      if (keys.length === 1) {
        await route.abort("failed");
        return;
      }
    }
    await route.fulfill({ response });
  });
  await page.goto("/onboarding");
  await expect(page.getByRole("heading", { name: "Create your workspace" })).toBeVisible();
  await page.getByRole("button", { name: "Create workspace", exact: true }).click();
  await expect(page.getByLabel("Workspace name")).toBeFocused();
  await page.getByLabel("Workspace name").fill("  Browser Growth 🚀  ");
  await page.getByRole("button", { name: "Create workspace", exact: true }).click();
  await expect(page.locator("#workspace-error")).toHaveText(
    "Workspace could not be created. Try again.",
  );
  await expect(page.getByLabel("Workspace name")).toHaveValue("  Browser Growth 🚀  ");
  await page.getByRole("button", { name: "Create workspace", exact: true }).click();
  await expect(page).toHaveURL(/\/workspaces\/[0-9a-f-]{36}\/links$/);
  expect(keys).toHaveLength(2);
  expect(keys[1]).toBe(keys[0]);
  await expect(page.getByRole("heading", { name: "Browser Growth 🚀" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "No links yet" })).toBeVisible();
  const navigation = page.getByRole("navigation", { name: "Workspace" });
  await expect(navigation.getByRole("link", { name: "Links", exact: true })).toHaveAttribute(
    "aria-current",
    "page",
  );
  await expect(navigation.getByRole("button", { name: "Team", exact: true })).toBeDisabled();
  await expect(navigation.getByRole("link", { name: "Team", exact: true })).toHaveCount(0);
  await expect(page.getByText("Create your first link", { exact: true })).toBeVisible();
  await expect(
    page.getByText("Link management is available. Redirects and analytics are not available yet."),
  ).toBeVisible();
  await page.goto("/workspaces/00000000-0000-4000-8000-000000000000/links");
  await expect(page.getByText("You do not have access to this workspace.")).toBeVisible();
  await expect(page.getByRole("heading", { name: "No links yet" })).toHaveCount(0);
  await expect(page.getByRole("navigation", { name: "Workspace" })).toHaveCount(0);
});

const providerAssets = new WeakMap<Page, Set<Promise<void>>>();

// Only the browser harness substitutes FAPI responses. The real SDK and native
// UI still render; local transport proofs never count as live factor acceptance.
export const installProviderTransport = async (
  page: Page,
  sessionClient?: Record<string, unknown>,
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
    if (/\/sessions\/[^/]+\/(?:remove|end)$/.test(url.pathname)) {
      client.sessions = [];
      client.last_active_session_id = null;
    }
    if (url.pathname.startsWith("/npm/")) {
      // Fetch provider-owned SDK assets, not fake application auth controls.
      const asset = (async () => {
        const response = await route.fetch({
          url: `https://cdn.jsdelivr.net${url.pathname.replace("@clerk/clerk-js@6/", "@clerk/clerk-js@6.38.1/").replace("@clerk/ui@1/", "@clerk/ui@1.39.1/")}${url.search}`,
          timeout: 15000,
        });
        await route.fulfill({ response });
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
    const response =
      url.pathname.includes("/tokens") && sessions?.[0]
        ? sessions[0].last_active_token
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

test.beforeEach(async ({ page, request }) => {
  const reset = await request.post(`${process.env.FLUX_BROWSER_FIXTURE}/restore-reset`);
  expect(reset.status()).toBe(204);
  await installProviderTransport(page);
});

test("session root loads committed identity and signout clears private data", async ({
  page,
  request,
}) => {
  const protocol = process.env.FLUX_BROWSER_FIXTURE;
  expect(protocol).toMatch(/^http:\/\/127\.0\.0\.1:\d+$/);
  const fixture = await (await request.get(`${protocol}/client`)).json();
  await installProviderTransport(page, fixture.client);
  await page.route("**/api/v1/me", async (route) => {
    const response = await route.fetch({ url: `${fixture.api}/api/v1/me` });
    await route.fulfill({ response });
  });
  await page.goto("/");
  await expect(page).toHaveURL(/\/workspaces$/);
  await expect(
    page.getByRole("heading", { name: "Choose a workspace", exact: true }),
  ).toBeVisible();
  await expect(page.getByText("local@example.test", { exact: true })).toBeVisible();
  await page.setViewportSize({ width: 320, height: 640 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(
    true,
  );
  await page.getByRole("button", { name: "Sign out", exact: true }).focus();
  await expect(page.getByRole("button", { name: "Sign out", exact: true })).toBeFocused();
  await page.getByRole("button", { name: "Sign out", exact: true }).click();
  await expect(page.getByText("local@example.test", { exact: true })).toHaveCount(0);
  await expect(page.getByRole("heading", { name: "Sign in to Flux", exact: true })).toBeVisible();
});

for (const status of [401, 503]) {
  test(`session root handles ${status} safely and supports recovery`, async ({ page, request }) => {
    const protocol = process.env.FLUX_BROWSER_FIXTURE;
    expect(protocol).toMatch(/^http:\/\/127\.0\.0\.1:\d+$/);
    const fixture = await (await request.get(`${protocol}/client`)).json();
    await installProviderTransport(page, fixture.client);
    await page.route("**/api/v1/me", (route) =>
      route.fulfill({ status, body: "PRIVATE-PROVIDER-DIAGNOSTIC" }),
    );
    await page.goto("/");
    await expect(page.getByRole("main").getByRole("alert")).toHaveText(
      status === 401
        ? "Your session ended. Sign in to continue."
        : "Sign-in is temporarily unavailable. Try again shortly.",
    );
    await expect(page.getByText("PRIVATE-PROVIDER-DIAGNOSTIC")).toHaveCount(0);
    await expect(page.getByText("local@example.test", { exact: true })).toHaveCount(0);
    if (status === 401) {
      await expect(page.getByRole("link", { name: "Sign in", exact: true })).toBeVisible();
    } else {
      await page.unroute("**/api/v1/me");
      await page.route("**/api/v1/me", async (route) => {
        const response = await route.fetch({ url: `${fixture.api}/api/v1/me` });
        await route.fulfill({ response });
      });
      await page.getByRole("button", { name: "Retry", exact: true }).click();
      await expect(page.getByText("local@example.test", { exact: true })).toBeVisible();
    }
  });
}

const finishProviderTransport = async (page: Page) => {
  // Stop the document producing new requests, then finish only the fixture's
  // provider assets. Next development traffic is not a fixture completion signal.
  await page.goto("about:blank", { waitUntil: "commit" });
  const assets = providerAssets.get(page);
  while (assets?.size) await Promise.all([...assets]);
  await page.unrouteAll({ behavior: "wait" });
  providerAssets.delete(page);
};
test.afterEach(async ({ page }) => finishProviderTransport(page));

test("sign-in reaches native password and recovery controls using isolated transport", async ({
  page,
}) => {
  await page.goto("/sign-in");
  await page.getByLabel("Email address", { exact: true }).fill("local@example.test");
  await page.getByRole("button", { name: /^Continue/ }).click();
  await expect(page.getByLabel("Password", { exact: true })).toBeVisible();
  const recovery = page.getByRole("link", { name: "Forgot password?" });
  await expect(recovery).toBeVisible();
  await recovery.focus();
  await page.keyboard.press("Enter");
  await expect(page.getByText(/Reset your password/).first()).toBeVisible();
});

test("sign-in links to native sign-up with required password and social options", async ({
  page,
}) => {
  await page.goto("/sign-in");
  await page.getByRole("link", { name: "Sign up", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Create your account", exact: true }).first(),
  ).toBeVisible();
  await expect(page.getByLabel("Email address", { exact: true })).toBeVisible();
  await expect(page.getByLabel("Password", { exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: /Google/ })).toBeVisible();
  await expect(page.getByRole("button", { name: /GitHub/ })).toBeVisible();
});

test("sign-in sign-up reaches native mandatory email verification without creating a session", async ({
  page,
}) => {
  await page.goto("/sign-up");
  await page.getByLabel("Email address", { exact: true }).fill("local@example.test");
  await page.getByLabel("Password", { exact: true }).fill("Local test password9!");
  await page.getByRole("button", { name: /^Continue/ }).click();
  await expect(page.getByRole("heading", { name: /Verify your email/ })).toBeVisible();
  await expect(page.getByText("Your account is signed in.")).toHaveCount(0);
});

for (const width of [320, 768, 1280]) {
  test(`sign-in remains usable at ${width}px with keyboard and reduced motion`, async ({
    page,
  }) => {
    await page.setViewportSize({ width, height: 900 });
    await page.emulateMedia({ reducedMotion: "reduce" });
    await page.goto("/sign-in");
    const email = page.getByLabel("Email address", { exact: true });
    await expect(email).toBeVisible();
    await email.focus();
    await expect(email).toBeFocused();
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth),
    ).toBe(true);
  });
}

test("sign-in exposes original auth heading and native provider controls", async ({ page }) => {
  await page.goto("/sign-in");
  await expect(page.getByRole("heading", { name: "Sign in to Flux", exact: true })).toBeVisible();
  await expect(page.getByLabel("Email address", { exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: /Google/ })).toBeVisible();
  await expect(page.getByRole("button", { name: /GitHub/ })).toBeVisible();
});

test("bearer signed browser request reaches the production Go boundary", async ({
  page,
  request,
}) => {
  const protocol = process.env.FLUX_BROWSER_FIXTURE;
  expect(protocol).toMatch(/^http:\/\/127\.0\.0\.1:\d+$/);
  const fixture = await (await request.get(`${protocol}/client`)).json();
  await installProviderTransport(page, fixture.client);
  await page.goto("/sign-in");
  await expect(page.getByText("Your account is signed in.")).toBeVisible();
  const result = await page.evaluate(async ({ api }) => {
    const clerk = (window as unknown as { Clerk: { session: { getToken: () => Promise<string> } } })
      .Clerk;
    const token = await clerk.session.getToken();
    const response = await fetch(`${api}/api/v1/me`, {
      headers: { Authorization: `Bearer ${token}` },
      credentials: "omit",
      cache: "no-store",
    });
    return {
      status: response.status,
      body: await response.json(),
      cache: response.headers.get("cache-control"),
    };
  }, fixture);
  expect(result.status).toBe(200);
  expect(result.body.authenticated).toBe(true);
  expect(result.body.user.id).toMatch(/^[0-9a-f-]{36}$/);
  expect(result.body.user.email).toBe("local@example.test");
  expect(result.cache).toBe("no-store");
});

test("bearer denies revoked expired foreign and cookie-only credentials safely", async ({
  page,
  request,
}) => {
  const protocol = process.env.FLUX_BROWSER_FIXTURE;
  expect(protocol).toMatch(/^http:\/\/127\.0\.0\.1:\d+$/);
  const cases = await (await request.get(`${protocol}/cases`)).json();
  await page.goto("/sign-in");
  await expect(page.getByLabel("Email address", { exact: true })).toBeVisible();
  for (const name of ["revoked", "expired", "foreign", "outage"]) {
    const result = await page.evaluate(
      async ({ api, token }) => {
        const response = await fetch(`${api}/api/v1/me`, {
          headers: { Authorization: `Bearer ${token}` },
          credentials: "omit",
          cache: "no-store",
        });
        return { status: response.status, body: await response.text() };
      },
      { api: cases.api, token: cases[name] },
    );
    expect(result.status).toBe(name === "outage" ? 503 : 401);
    expect(result.body).not.toContain(cases[name]);
    expect(result.body).not.toContain("user_fixture");
  }
  const cookieOnly = await request.get(`${cases.api}/api/v1/me`, {
    headers: { Cookie: "__session=ambient-cookie" },
  });
  expect(cookieOnly.status()).toBe(401);
});

test("identity resolves the signed browser session to one stable internal UUID", async ({
  page,
  request,
}) => {
  const protocol = process.env.FLUX_BROWSER_FIXTURE;
  expect(protocol).toMatch(/^http:\/\/127\.0\.0\.1:\d+$/);
  const fixture = await (await request.get(`${protocol}/client`)).json();
  await installProviderTransport(page, fixture.client);
  await page.goto("/sign-in");
  await expect(page.getByText("Your account is signed in.")).toBeVisible();
  const results = await page.evaluate(async ({ api }) => {
    const clerk = (window as unknown as { Clerk: { session: { getToken: () => Promise<string> } } })
      .Clerk;
    const token = await clerk.session.getToken();
    const results = [];
    for (let i = 0; i < 3; i++) {
      const response = await fetch(`${api}/api/v1/me`, {
        headers: { Authorization: `Bearer ${token}` },
        credentials: "omit",
        cache: "no-store",
      });
      results.push({
        status: response.status,
        body: await response.json(),
        cache: response.headers.get("cache-control"),
      });
    }
    return results;
  }, fixture);
  expect(results).toHaveLength(3);
  const first = results[0];
  if (!first) throw new Error("The identity request did not produce a result");
  expect(first.status).toBe(200);
  expect(first.cache).toBe("no-store");
  expect(first.body.user.id).toMatch(
    /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
  );
  expect(results[1]).toEqual(first);
  expect(results[2]).toEqual(first);
  expect(JSON.stringify(results)).not.toContain("user_fixture");
  expect(JSON.stringify(results)).not.toContain(fixture.token);
});
