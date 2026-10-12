import { expect, type Page, test } from "@playwright/test";
import type { LinkResponse, WorkspaceResponse } from "@flux/zod";
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

for (const boundary of ["switch", "removal", "signout"] as const) {
  test(`tenant-denial two tabs discard held link DTOs after ${boundary}`, async ({
    page,
    context,
    request,
  }) => {
    const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
    const headers = { Authorization: `Bearer ${fixture.token}`, Origin: "http://127.0.0.1:3100" };
    const workspaces: WorkspaceResponse["workspace"][] = [];
    const links: LinkResponse["link"][] = [];
    for (const index of [0, 1]) {
      const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
        headers: { ...headers, "Idempotency-Key": `tenant-denial-${boundary}-workspace-${index}` },
        data: { name: `Private ${boundary} workspace ${index}` },
      });
      expect(created.status()).toBe(201);
      const { workspace } = await created.json();
      workspaces.push(workspace);
      const result = await request.post(`${fixture.api}/api/v1/workspaces/${workspace.id}/links`, {
        headers: { ...headers, "Idempotency-Key": `tenant-denial-${boundary}-link-${index}` },
        data: {
          destination: `https://example.com/private-${boundary}-${index}`,
          title: `Private ${boundary} record ${index}`,
        },
      });
      expect(result.status()).toBe(201);
      links.push((await result.json()).link);
    }
    const [oldWorkspace, nextWorkspace] = workspaces;
    const [oldLink, nextLink] = links;
    if (!oldWorkspace || !nextWorkspace || !oldLink || !nextLink)
      throw new Error("Missing seeded tenant records");
    const second = await context.newPage();
    let release!: () => void;
    const held = new Promise<void>((resolve) => {
      release = resolve;
    });
    const telemetry: string[] = [];
    const completions: Promise<void>[] = [];
    const initialRouteScripts = new Map<Page, string[]>();
    const bootstrap = new Map<Page, { count: number; ids: string[] }>();
    const initialBootstrapCounts = new Map<Page, number>();
    try {
      for (const tab of [page, second]) {
        // Model the renewed Client cookie separately from bounded Session/JWT.
        // Near-expiry Client cookies trigger the SDK's top-level ITP touch.
        await installProviderTransport(tab, {
          ...structuredClone(fixture.client),
          cookie_expires_at: Date.now() + 30 * 86400000,
        });
        tab.on("console", (message) => telemetry.push(message.text()));
        tab.on("pageerror", (error) => telemetry.push(error.message));
        await tab.route("**/api/v1/**", async (route) => {
          const url = new URL(route.request().url());
          const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
          if (url.pathname === "/api/v1/me" && response.status() === 200) {
            const identity = await response.json();
            bootstrap.set(tab, {
              count: (bootstrap.get(tab)?.count ?? 0) + 1,
              ids: identity.workspaces.map((workspace: { id: string }) => workspace.id),
            });
          }
          await route.fulfill({ response });
        });
        await tab.goto(`/workspaces/${oldWorkspace.id}/links`);
        await expect(tab.getByRole("table", { name: "Links library" })).toContainText(
          oldLink.title,
        );
        // Next embeds the initial client-request route parameter in immutable
        // hydration scripts. Record those exact bytes before any held DTO.
        initialRouteScripts.set(
          tab,
          await tab
            .locator("script")
            .evaluateAll(
              (nodes, id) =>
                nodes.map((node) => node.textContent ?? "").filter((text) => text.includes(id)),
              oldWorkspace.id,
            ),
        );
        initialBootstrapCounts.set(tab, bootstrap.get(tab)?.count ?? 0);
        let observed!: () => void;
        const started = new Promise<void>((resolve) => {
          observed = resolve;
        });
        let finished!: () => void;
        completions.push(
          new Promise<void>((resolve) => {
            finished = resolve;
          }),
        );
        await tab.route(`**/api/v1/workspaces/${oldWorkspace.id}/links?*`, async (route) => {
          const url = new URL(route.request().url());
          const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
          // Hold a real, already authorized DTO containing the private record.
          const dto = await response.json();
          expect(dto.items.some((item: { id: string }) => item.id === oldLink.id)).toBe(true);
          observed();
          await held;
          try {
            await route.fulfill({ response });
          } catch {
            // The boundary may have aborted this exact in-flight transport.
            // Any other route error fails the proof rather than being hidden.
            expect(route.request().failure()?.errorText === "net::ERR_ABORTED").toBe(true);
          }
          finished();
        });
        await tab.getByLabel("Search links", { exact: true }).fill("Private");
        await tab.getByRole("button", { name: "Search links", exact: true }).click();
        await started;
      }
      if (boundary === "switch") {
        await page.getByLabel("Switch workspace", { exact: true }).selectOption(nextWorkspace.id);
        await page.getByRole("button", { name: "Open workspace", exact: true }).click();
        await expect(
          page.getByRole("heading", { name: nextWorkspace.name, exact: true }),
        ).toBeVisible();
        await expect(
          second.getByRole("heading", { name: "Choose a workspace", exact: true }),
        ).toBeVisible();
        await expect(second).toHaveURL(/\/workspaces\?access=changed$/);
        await second.getByLabel("Workspace", { exact: true }).selectOption(nextWorkspace.id);
        await second.getByRole("button", { name: "Open workspace", exact: true }).click();
        await expect(
          second.getByRole("heading", { name: nextWorkspace.name, exact: true }),
        ).toBeVisible();
        await page.goto(`/workspaces/${nextWorkspace.id}/links`);
        for (const tab of [page, second]) {
          await expect(tab.getByRole("table", { name: "Links library" })).toContainText(
            nextLink.title,
          );
        }
      } else if (boundary === "removal") {
        expect(
          (
            await request.post(
              `${process.env.FLUX_BROWSER_FIXTURE}/restore-remove?workspace=${oldWorkspace.id}`,
            )
          ).status(),
        ).toBe(204);
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
        }
      } else {
        await page.getByRole("button", { name: "Sign out", exact: true }).click();
        await expect(
          page.getByRole("heading", { name: "Sign in to Flux", exact: true }),
        ).toBeVisible();
        // The invalidation signal contains no provider credential or authority.
        // Finish signout through the second tab's own native provider action.
        await expect(
          second.getByRole("heading", { name: "Choose a workspace", exact: true }),
        ).toBeVisible();
        await expect(second).toHaveURL(/\/workspaces\?access=changed$/);
        await second.getByRole("button", { name: "Sign out", exact: true }).click();
        await expect(
          second.getByRole("heading", { name: "Sign in to Flux", exact: true }),
        ).toBeVisible();
      }
      // Each tab must dispose the old records before releasing the old DTO.
      for (const tab of [page, second]) {
        await expect(tab.getByText(oldLink.title, { exact: true })).toHaveCount(0);
      }
      release();
      await Promise.all(completions);
      for (const tab of [page, second]) {
        if (boundary === "signout") {
          await expect(
            tab.getByRole("heading", { name: "Sign in to Flux", exact: true }),
          ).toBeVisible();
        }
        const surface = await tab.locator("body").innerHTML();
        expect(
          [oldLink.id, oldLink.title, oldLink.destination].some((value) => surface.includes(value)),
        ).toBe(false);
        const retainedScripts = await tab
          .locator("script")
          .evaluateAll(
            (nodes, id) =>
              nodes.map((node) => node.textContent ?? "").filter((text) => text.includes(id)),
            oldWorkspace.id,
          );
        const initial = initialRouteScripts.get(tab) ?? [];
        expect(
          retainedScripts.length === 0 ||
            (retainedScripts.length === initial.length &&
              retainedScripts.every((text, index) => text === initial[index])),
        ).toBe(true);
        if (boundary === "switch") {
          const fresh = bootstrap.get(tab);
          expect((fresh?.count ?? 0) > (initialBootstrapCounts.get(tab) ?? 0)).toBe(true);
          expect(fresh?.ids.includes(oldWorkspace.id)).toBe(true);
          await expect(
            tab
              .getByLabel("Switch workspace", { exact: true })
              .locator(`option[value="${oldWorkspace.id}"]`),
          ).toHaveCount(1);
        }
        const forbiddenLocation = await tab.evaluate(
          ({ id, switching }) => {
            const body = document.body.cloneNode(true) as HTMLElement;
            body.querySelectorAll("script").forEach((node) => node.remove());
            if (switching) {
              const options = body.querySelectorAll<HTMLOptionElement>(
                "select#switch-workspace option",
              );
              options.forEach((option) => {
                if (option.getAttribute("value") === id) option.removeAttribute("value");
              });
            }
            return body.innerHTML.includes(id);
          },
          { id: oldWorkspace.id, switching: boundary === "switch" },
        );
        expect(forbiddenLocation).toBe(false);
        expect((await tab.getByRole("main").innerHTML()).includes(oldWorkspace.id)).toBe(false);
      }
      expect(
        [
          fixture.token,
          oldLink.id,
          oldLink.title,
          oldLink.destination,
          oldWorkspace.id,
          oldWorkspace.name,
          "local@example.test",
        ].some((value) => telemetry.some((entry) => entry.includes(value))),
      ).toBe(false);
    } finally {
      release();
      // The held API handlers have explicit completion barriers above. Drain
      // public assets, then close this owned secondary page directly: a forced
      // navigation can race the native provider's completed signout navigation.
      const assets = providerAssets.get(second);
      while (assets?.size) await Promise.all([...assets]);
      await second.close();
      providerAssets.delete(second);
    }
  });
}
