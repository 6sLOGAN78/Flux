import { expect, type Page, test } from "@playwright/test";
import { providerAssetCache } from "./provider-assets";

const publicProviderAsset = providerAssetCache();

test("removal of another member retains creator links and invalidates their tab from fresh SQL authority", async ({
  page,
  context,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  const headers = {
    Authorization: `Bearer ${fixture.token}`,
    Origin: "http://127.0.0.1:3100",
    "Idempotency-Key": "browser-other-removal",
  };
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers,
    data: { name: "Other removal tenant" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  expect(
    (
      await request.post(
        `${process.env.FLUX_BROWSER_FIXTURE}/restore-member?workspace=${workspace.id}`,
      )
    ).status(),
  ).toBe(204);
  const member = await (
    await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/member-client`)
  ).json();
  const linked = await request.post(`${fixture.api}/api/v1/workspaces/${workspace.id}/links`, {
    headers: {
      ...headers,
      Authorization: `Bearer ${member.token}`,
      "Idempotency-Key": "browser-removal-creator",
    },
    data: { destination: "https://example.com/creator", title: "Creator retained" },
  });
  expect(linked.status()).toBe(201);
  const { link } = await linked.json();
  const second = await context.newPage();
  let release!: () => void;
  let captured!: () => void;
  const ready = new Promise<void>((resolve) => {
    captured = resolve;
  });
  const delayed = new Promise<void>((resolve) => {
    release = resolve;
  });
  try {
    for (const [tab, client] of [
      [page, fixture.client],
      [second, member.client],
    ] as const) {
      await installProviderTransport(tab, client);
      await tab.route("**/api/v1/**", async (route) => {
        const url = new URL(route.request().url());
        const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
        if (tab === second && url.searchParams.get("search") === "Creator") {
          captured();
          await delayed;
        }
        await route.fulfill({ response }).catch(() => undefined);
      });
    }
    await second.goto(`/workspaces/${workspace.id}/links`);
    await expect(second.getByText("Creator retained", { exact: true }).first()).toBeVisible();
    await second.getByLabel("Search links", { exact: true }).fill("Creator");
    await second.getByRole("button", { name: "Search links", exact: true }).click();
    await ready;
    await page.goto(`/workspaces/${workspace.id}/team`);
    const team = page.getByRole("main", { name: "Team", exact: true });
    await team
      .getByRole("row")
      .filter({ hasText: "colleague@example.test" })
      .getByRole("button", { name: "Remove member", exact: true })
      .click();
    await page
      .getByRole("dialog", { name: "Remove member?", exact: true })
      .getByRole("button", { name: "Remove member", exact: true })
      .click();
    await expect(team.getByRole("status").filter({ hasText: "Member removed." })).toBeVisible();
    await expect(second).toHaveURL(/\/workspaces\?access=changed/);
    release();
    await expect(second.getByText("Creator retained", { exact: true })).toHaveCount(0);
    await expect(second.getByText("Other removal tenant", { exact: true })).toHaveCount(0);
    await expect(
      second.getByRole("heading", { name: "Choose a workspace", exact: true }),
    ).toBeVisible();
    const me = await request.get(`${fixture.api}/api/v1/me`, {
      headers: { Authorization: `Bearer ${member.token}` },
    });
    expect(me.status()).toBe(200);
    expect(
      (await me.json()).workspaces.some((item: { id: string }) => item.id === workspace.id),
    ).toBe(false);
    const denied = await request.get(
      `${fixture.api}/api/v1/workspaces/${workspace.id}/links/${link.id}`,
      { headers: { Authorization: `Bearer ${member.token}` } },
    );
    expect(denied.status()).toBe(404);
    const retained = await request.get(
      `${fixture.api}/api/v1/workspaces/${workspace.id}/links/${link.id}`,
      { headers },
    );
    expect(retained.status()).toBe(200);
    expect((await retained.json()).link).toEqual(link);
  } finally {
    release?.();
    await second.close();
  }
});

test("removal confirms safely, preserves failure state, then self removal scrubs both tabs and late links", async ({
  page,
  context,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const headers = {
    Authorization: `Bearer ${fixture.token}`,
    Origin: "http://127.0.0.1:3100",
    "Idempotency-Key": "browser-removal-workspace",
  };
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers,
    data: { name: "Removal browser tenant" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  expect(
    (
      await request.post(
        `${process.env.FLUX_BROWSER_FIXTURE}/restore-member?workspace=${workspace.id}`,
      )
    ).status(),
  ).toBe(204);
  let mutations = 0;
  let fail = false;
  const routeAPI = async (target: Page) => {
    await target.route("**/api/v1/**", async (route) => {
      const url = new URL(route.request().url());
      if (route.request().method() === "DELETE") {
        mutations++;
        if (fail) {
          await route.fulfill({ status: 503, contentType: "application/json", body: "{}" });
          return;
        }
      }
      const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
      await route.fulfill({ response });
    });
  };
  await routeAPI(page);
  await page.goto(`/workspaces/${workspace.id}/team`);
  const team = page.getByRole("main", { name: "Team", exact: true });
  const own = team.getByRole("row").filter({ hasText: "local@example.test" });
  const trigger = own.getByRole("button", { name: "Remove member", exact: true });
  await trigger.click();
  const dialog = page.getByRole("dialog", { name: "Remove member?", exact: true });
  await expect(dialog.getByRole("button", { name: "Keep member", exact: true })).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(dialog).toHaveCount(0);
  await expect(trigger).toBeFocused();
  expect(mutations).toBe(0);
  await trigger.click();
  await dialog.getByRole("button", { name: "Remove member", exact: true }).click();
  await expect(team.getByRole("alert")).toHaveText("Promote another owner first");
  expect(mutations).toBe(1);
  await team.getByRole("button", { name: "Reload team", exact: true }).click();
  const colleague = team.getByRole("row").filter({ hasText: "colleague@example.test" });
  await colleague.getByLabel("New role").selectOption("owner");
  await colleague.getByRole("button", { name: "Change role", exact: true }).click();
  await page
    .getByRole("dialog", { name: "Change role?", exact: true })
    .getByRole("button", { name: "Change role", exact: true })
    .click();
  await expect(colleague.locator("td").nth(1)).toHaveText("Owner");
  fail = true;
  await trigger.click();
  await dialog.getByRole("button", { name: "Remove member", exact: true }).click();
  await expect(team.getByRole("alert")).toHaveText(
    "We couldn't remove the member. Reload team before trying again.",
  );
  expect(mutations).toBe(2);
  await expect(trigger).toBeDisabled();
  fail = false;
  await team.getByRole("button", { name: "Reload team", exact: true }).click();
  const second = await context.newPage();
  await installProviderTransport(second, fixture.client);
  await routeAPI(second);
  await second.goto(`/workspaces/${workspace.id}/links`);
  await expect(second.getByRole("heading", { name: "Links", exact: true })).toBeVisible();
  let release!: () => void;
  let fetched!: () => void;
  const ready = new Promise<void>((resolve) => {
    fetched = resolve;
  });
  const delayed = new Promise<void>((resolve) => {
    release = resolve;
  });
  try {
    await second.route("**/api/v1/workspaces/*/links?*", async (route) => {
      const url = new URL(route.request().url());
      const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
      fetched();
      await delayed;
      await route.fulfill({ response }).catch(() => {});
    });
    await second.getByLabel("Search links", { exact: true }).fill("retained");
    await second.getByRole("button", { name: "Search links", exact: true }).click();
    await ready;
    await page.bringToFront();
    await trigger.click();
    await expect(
      dialog.getByText("This member will lose access to this workspace.", { exact: true }),
    ).toBeVisible();
    await dialog.getByRole("button", { name: "Remove member", exact: true }).click();
    await expect(page).toHaveURL(/\/workspaces\?access=changed/);
    await expect(second).toHaveURL(/\/workspaces\?access=changed/);
    release();
    for (const tab of [page, second]) {
      await expect(tab.getByText("Removal browser tenant", { exact: true })).toHaveCount(0);
      await expect(tab.getByRole("table")).toHaveCount(0);
      await expect(
        tab.getByRole("heading", { name: "Choose a workspace", exact: true }),
      ).toBeVisible();
      await expect(tab.getByText("You have been signed out.", { exact: true })).toHaveCount(0);
    }
    expect(mutations).toBe(3);
  } finally {
    release();
    await second.close();
  }
});

test("roles safely confirm promotion then self demotion and preserve Links", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-role-changes",
    },
    data: { name: "Role change tenant" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  expect(
    (
      await request.post(
        `${process.env.FLUX_BROWSER_FIXTURE}/restore-member?workspace=${workspace.id}`,
      )
    ).status(),
  ).toBe(204);
  let mutations = 0;
  let fail = false;
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    if (route.request().method() === "PATCH") {
      mutations++;
      if (fail) {
        await route.fulfill({ status: 503, contentType: "application/json", body: "{}" });
        return;
      }
    }
    const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/team`);
  const team = page.getByRole("main", { name: "Team", exact: true });
  const own = team.getByRole("row").filter({ hasText: "local@example.test" });
  const colleague = team.getByRole("row").filter({ hasText: "colleague@example.test" });
  await expect(own.getByText("Promote another owner first", { exact: true })).toBeVisible();
  await own.getByLabel("New role").selectOption("admin");
  await own.getByRole("button", { name: "Change role", exact: true }).click();
  await page
    .getByRole("dialog", { name: "Change role?", exact: true })
    .getByRole("button", { name: "Change role", exact: true })
    .click();
  await expect(team.getByRole("alert")).toHaveText("Promote another owner first");
  await expect(own.locator("td").nth(1)).toHaveText("Owner");
  await team.getByRole("button", { name: "Reload team", exact: true }).click();
  await colleague.getByLabel("New role").selectOption("owner");
  const trigger = colleague.getByRole("button", { name: "Change role", exact: true });
  await trigger.click();
  const dialog = page.getByRole("dialog", { name: "Change role?", exact: true });
  await expect(
    dialog.getByRole("button", { name: "Keep current role", exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(dialog).toHaveCount(0);
  await expect(trigger).toBeFocused();
  expect(mutations).toBe(1);
  await trigger.click();
  fail = true;
  await dialog.getByRole("button", { name: "Change role", exact: true }).click();
  await expect(team.getByRole("alert")).toHaveText(
    "We couldn't change the role. Reload team before trying again.",
  );
  expect(mutations).toBe(2);
  await expect(colleague.locator("td").nth(1)).toHaveText("Member");
  fail = false;
  await team.getByRole("button", { name: "Reload team", exact: true }).click();
  await colleague.getByLabel("New role").selectOption("owner");
  await trigger.click();
  await dialog.getByRole("button", { name: "Change role", exact: true }).click();
  await expect(team.getByRole("status").filter({ hasText: "Role changed." })).toBeVisible();
  await expect(colleague.locator("td").nth(1)).toHaveText("Owner");
  await own.getByLabel("New role").selectOption("member");
  await own.getByRole("button", { name: "Change role", exact: true }).click();
  await expect(
    dialog.getByText("This change removes the member's current management permissions.", {
      exact: true,
    }),
  ).toBeVisible();
  await dialog.getByRole("button", { name: "Change role", exact: true }).click();
  await expect(
    team.getByText("You do not have permission to view Team.", { exact: true }),
  ).toBeVisible();
  await expect(team.getByRole("table")).toHaveCount(0);
  await page.getByRole("link", { name: "Links", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Links", exact: true })).toBeVisible();
  expect(mutations).toBe(4);
});

test("team inspects actual members with responsive roles and truthful retry then scrubs removed access", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-team-inspect",
    },
    data: { name: "Team inspection tenant" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  await new Promise((resolve) => setTimeout(resolve, 70));
  let fail = false;
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    if (fail && url.pathname.endsWith("/members")) {
      await route.fulfill({ status: 503, contentType: "application/json", body: "{}" });
      return;
    }
    const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/links`);
  await expect(page.getByRole("heading", { name: "Links", exact: true })).toBeVisible();
  await page.getByRole("link", { name: "Team", exact: true }).click({ timeout: 10000 });
  await expect(page.getByRole("table", { name: "Team members" })).toBeVisible();
  await expect(page.getByRole("table").locator("tbody td").nth(1)).toHaveText("Owner");
  const identity = await page.getByRole("table").locator("tbody td").first().textContent();
  expect(identity).toContain("@");
  await expect(page.getByRole("form", { name: "Invite member" })).toBeVisible();
  fail = true;
  await page.getByRole("button", { name: "Reload team", exact: true }).click();
  const team = page.getByRole("main", { name: "Team", exact: true });
  await expect(team.getByRole("alert")).toHaveText("We couldn't load team. Try again.");
  await expect(page.getByRole("table")).toBeVisible();
  await expect(page.getByText("No members", { exact: true })).toHaveCount(0);
  fail = false;
  await page.getByRole("button", { name: "Retry loading team", exact: true }).click();
  await expect(team.getByRole("alert")).toHaveCount(0);
  await page.setViewportSize({ width: 320, height: 720 });
  await expect(page.getByRole("list", { name: "Team members" })).toBeVisible();
  await expect(page.getByRole("list", { name: "Team members" }).locator("dd").nth(1)).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(
    true,
  );
  const removed = await request.post(
    `${process.env.FLUX_BROWSER_FIXTURE}/restore-remove?workspace=${workspace.id}`,
  );
  expect(removed.status()).toBe(204);
  await page.getByRole("button", { name: "Reload team", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Choose a workspace", exact: true }),
  ).toBeVisible();
  await expect(page.getByText("Team inspection tenant", { exact: true })).toHaveCount(0);
  await expect(page.getByRole("list", { name: "Team members" })).toHaveCount(0);
});

test("team fresh role matrix disposes stale management access while retaining valid Links", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-team-role-matrix",
    },
    data: { name: "Team role tenant" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  let hold = false;
  let release: (() => void) | undefined;
  let held: (() => void) | undefined;
  const arrived = new Promise<void>((resolve) => {
    held = resolve;
  });
  const released = new Promise<void>((resolve) => {
    release = resolve;
  });
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
    if (hold && url.pathname.endsWith("/members")) {
      held?.();
      await released;
    }
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/team`);
  await expect(page.getByRole("table", { name: "Team members" })).toBeVisible();
  for (const role of ["admin", "member", "viewer", "owner"]) {
    const changed = await request.post(
      `${process.env.FLUX_BROWSER_FIXTURE}/restore-role?workspace=${workspace.id}&role=${role}`,
    );
    expect(changed.status()).toBe(204);
    if (role === "admin" || role === "member") {
      await page.getByRole("button", { name: "Reload team", exact: true }).click();
    } else {
      await page.goto(`/workspaces/${workspace.id}/team`);
    }
    if (role === "owner" || role === "admin") {
      await expect(page.getByRole("table").locator("tbody td").nth(1)).toHaveText(
        role === "owner" ? "Owner" : "Admin",
      );
    } else {
      await expect(
        page.getByText("You do not have permission to view Team.", { exact: true }),
      ).toBeVisible();
      await expect(page.getByRole("table", { name: "Team members" })).toHaveCount(0);
      await expect(
        page
          .getByRole("navigation", { name: "Workspace", exact: true })
          .getByRole("link", { name: "Team", exact: true }),
      ).toHaveCount(0);
      await page.getByRole("link", { name: "Links", exact: true }).click();
      await expect(page.getByRole("heading", { name: "Links", exact: true })).toBeVisible();
      if (role === "viewer")
        await expect(
          page.getByText("You have view-only access. Ask an owner or admin to change your role."),
        ).toBeVisible();
    }
  }
  hold = true;
  await page.getByRole("button", { name: "Reload team", exact: true }).click();
  await arrived;
  expect(
    (
      await request.post(
        `${process.env.FLUX_BROWSER_FIXTURE}/restore-role?workspace=${workspace.id}&role=member`,
      )
    ).status(),
  ).toBe(204);
  await page.evaluate(() => window.dispatchEvent(new Event("focus")));
  await expect(
    page.getByText("You do not have permission to view Team.", { exact: true }),
  ).toBeVisible();
  release?.();
  await expect(page.getByRole("table", { name: "Team members" })).toHaveCount(0);
});

test("custom-key rejects invalid input and shows actual server collision with a fresh changed submission", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const headers = {
    Authorization: `Bearer ${fixture.token}`,
    Origin: "http://127.0.0.1:3100",
    "Idempotency-Key": "browser-custom-workspace",
  };
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers,
    data: { name: "Custom key tenant" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  const seeded = await request.post(`${fixture.api}/api/v1/workspaces/${workspace.id}/links`, {
    headers: { ...headers, "Idempotency-Key": "browser-custom-seed" },
    data: { destination: "https://example.com", customKey: "taken-key" },
  });
  expect(seeded.status()).toBe(201);
  const submissions: { key: string | undefined; body: string | null }[] = [];
  await page.route("**/api/v1/**", async (route) => {
    if (route.request().method() === "POST" && route.request().url().endsWith("/links")) {
      submissions.push({
        key: route.request().headers()["idempotency-key"],
        body: route.request().postData(),
      });
    }
    const response = await route.fetch({
      url: `${fixture.api}${new URL(route.request().url()).pathname}`,
    });
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/links/new`);
  await expect(page.getByText("Managed hostname: go.flux.test", { exact: true })).toBeVisible();
  await page
    .getByLabel("Destination URL", { exact: true })
    .fill("https://example.com/custom-browser");
  for (const value of ["a/b", "a%2fb", "a b", "aéz", "api"]) {
    await page.getByLabel("Custom short key", { exact: true }).fill(value);
    await page.getByRole("button", { name: "Create link", exact: true }).click();
    await expect(page.getByRole("main").getByRole("alert")).toHaveText(
      "Enter a valid custom short key.",
    );
    await expect(page.getByLabel("Custom short key", { exact: true })).toHaveValue(value);
  }
  expect(submissions).toHaveLength(0);
  await page.getByLabel("Custom short key", { exact: true }).fill("TAKEN-KEY");
  await expect(page.getByLabel("Custom short key", { exact: true })).toHaveValue("taken-key");
  await page.getByRole("button", { name: "Create link", exact: true }).click();
  await expect(page.getByRole("main").getByRole("alert")).toHaveText(
    "This short key is unavailable. Choose another key or generate one.",
  );
  await expect(page.getByText("Link created.", { exact: true })).toHaveCount(0);
  await page.getByLabel("Custom short key", { exact: true }).fill("available-key");
  await page.getByRole("button", { name: "Create link", exact: true }).click();
  await expect(page.getByText("Link created.", { exact: true })).toBeVisible();
  await expect(page.getByText("https://go.flux.test/available-key", { exact: true })).toBeVisible();
  expect(submissions).toHaveLength(2);
  expect(submissions[0]?.key).toBeTruthy();
  expect(submissions[1]?.key).not.toBe(submissions[0]?.key);
});

test("destination rejects reserved addresses with exact safe feedback", async ({
  page,
  request,
}) => {
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
    const response = await route.fetch({
      url: `${fixture.api}${new URL(route.request().url()).pathname}`,
    });
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/links/new`);
  await expect(page.getByText("Managed hostname: go.flux.test", { exact: true })).toBeVisible();
  await page.getByLabel("Destination URL", { exact: true }).fill("ftp://example.com");
  await page.getByRole("button", { name: "Create link", exact: true }).click();
  await expect(page.getByRole("main").getByRole("alert")).toHaveText(
    "Enter a valid HTTP or HTTPS URL.",
  );
  await page.getByLabel("Destination URL", { exact: true }).fill("http://[fec0::1]/private");
  await page.getByRole("button", { name: "Create link", exact: true }).click();
  await expect(page.getByRole("main").getByRole("alert")).toHaveText(
    "This destination isn't allowed. Choose a different public destination.",
  );
  await expect(page).toHaveURL(new RegExp(`/workspaces/${workspace.id}/links/new$`));
  await expect(page.getByLabel("Destination URL", { exact: true })).toHaveValue(
    "http://[fec0::1]/private",
  );
  await page
    .getByLabel("Destination URL", { exact: true })
    .fill("https://example.com/a%2Fb?q=a%2Bb&x=1#section");
  await page.getByRole("button", { name: "Create link", exact: true }).click();
  await expect(
    page.getByRole("link", { name: "Open destination (opens in a new tab)" }),
  ).toHaveAttribute("href", "https://example.com/a%2Fb?q=a%2Bb&x=1#section");
  await expect(
    page.getByRole("link", { name: "Open destination (opens in a new tab)" }),
  ).toHaveAttribute("rel", "noreferrer noopener");
  await page.evaluate(() =>
    Object.defineProperty(navigator, "clipboard", {
      value: { writeText: () => Promise.reject(new Error("denied")) },
      configurable: true,
    }),
  );
  await page.getByRole("button", { name: "Copy short URL", exact: true }).click();
  await expect(
    page.getByText("Couldn't copy. Select and copy the short URL below.", { exact: true }),
  ).toBeVisible();
  expect(await page.locator("time[datetime]").count()).toBe(2);
});

test("custom-key uncertain creation retries the identical committed request", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const response = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-destination-uncertain",
    },
    data: { name: "Destination retry tenant" },
  });
  expect(response.status()).toBe(201);
  const { workspace } = await response.json();
  const submissions: { key: string | undefined; body: string | null }[] = [];
  await page.route("**/api/v1/**", async (route) => {
    const response = await route.fetch({
      url: `${fixture.api}${new URL(route.request().url()).pathname}`,
    });
    if (
      route.request().method() === "POST" &&
      route.request().url().endsWith(`/workspaces/${workspace.id}/links`)
    ) {
      submissions.push({
        key: route.request().headers()["idempotency-key"],
        body: route.request().postData(),
      });
      expect(response.status()).toBe(201);
      if (submissions.length === 1) {
        await route.abort("failed");
        return;
      }
    }
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/links/new`);
  await page.getByLabel("Destination URL", { exact: true }).fill("https://example.com/uncertain");
  await page.getByLabel("Custom short key", { exact: true }).fill("Browser_Launch");
  await expect(page.getByLabel("Custom short key", { exact: true })).toHaveValue("browser_launch");
  await page.getByRole("button", { name: "Create link", exact: true }).click();
  await expect(page.getByRole("main").getByRole("alert")).toHaveText(
    "We couldn't confirm link creation. Retry to check the same request.",
  );
  await expect(page.getByText("Link created.", { exact: true })).toHaveCount(0);
  await page.getByRole("button", { name: "Retry creation", exact: true }).click();
  await expect(page.getByText("Link created.", { exact: true })).toBeVisible();
  expect(submissions).toHaveLength(2);
  expect(submissions[0]?.key).toBeTruthy();
  expect(submissions[1]).toEqual(submissions[0]);
  expect(JSON.parse(submissions[0]!.body!).customKey).toBe("browser_launch");
  await expect(
    page.getByText("https://go.flux.test/browser_launch", { exact: true }),
  ).toBeVisible();
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

test("invitation-queue: actual API queue, expiry, duplicate and draft boundaries", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-invitation-workspace-01",
    },
    data: { name: "Invitation Team" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/team`);
  const form = page.getByRole("form", { name: "Invite member" });
  await expect(form.getByLabel("Role", { exact: true })).toHaveValue("member");
  await form.getByLabel("Email address", { exact: true }).fill("future+tag@example.test");
  await form.getByRole("button", { name: "Invite member", exact: true }).click();
  await expect(
    page.getByRole("region", { name: "Invitations", exact: true }).getByRole("status"),
  ).toHaveText("Invitation queued");
  await expect(page.getByRole("region", { name: "Invitations", exact: true })).toContainText(
    "future+tag@example.test",
  );
  await expect(page.getByRole("region", { name: "Invitations", exact: true })).toContainText(
    "Expires",
  );
  await expect(page.getByRole("list", { name: "Workspace invitations" })).toContainText(
    "Delivery status: Queued",
  );
  const time = page.getByRole("list", { name: "Workspace invitations" }).locator("time");
  const expiresAt = await time.getAttribute("datetime");
  expect(expiresAt).toMatch(/^\d{4}-\d{2}-\d{2}T/);
  await expect(time).toHaveText(
    await page.evaluate(
      (value) =>
        new Intl.DateTimeFormat(undefined, {
          dateStyle: "full",
          timeStyle: "long",
        }).format(new Date(value)),
      expiresAt as string,
    ),
  );
  await form.getByLabel("Email address", { exact: true }).fill("future+tag@example.test");
  await form.getByRole("button", { name: "Invite member", exact: true }).click();
  await expect(
    page.getByRole("region", { name: "Invitations", exact: true }).getByRole("alert"),
  ).toHaveText("An invitation is already pending for this email. Resend or revoke it.");
  // A draft uses the existing native safe-first workspace-switch boundary.
  const other = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-invitation-workspace-02",
    },
    data: { name: "Other invitation Team" },
  });
  expect(other.status()).toBe(201);
  const second = (await other.json()).workspace;
  await page.reload();
  await form.getByLabel("Email address", { exact: true }).fill("draft@example.test");
  await page.getByLabel("Switch workspace").selectOption(second.id);
  await page.getByRole("button", { name: "Open workspace", exact: true }).click();
  const dialog = page.getByRole("dialog", { name: "Discard unsaved changes?" });
  await expect(dialog.getByRole("button", { name: "Stay", exact: true })).toBeFocused();
  await dialog.getByRole("button", { name: "Stay", exact: true }).click();
  await expect(form.getByLabel("Email address", { exact: true })).toHaveValue("draft@example.test");
  await page.getByRole("button", { name: "Open workspace", exact: true }).click();
  await dialog.getByRole("button", { name: "Discard", exact: true }).click();
  await expect(page).toHaveURL(new RegExp(`/workspaces/${second.id}/links$`));
  await page.getByRole("link", { name: "Team", exact: true }).click();
  await expect(form.getByLabel("Email address", { exact: true })).toHaveValue("");
  await expect(page.getByRole("region", { name: "Invitations", exact: true })).not.toContainText(
    "future+tag@example.test",
  );
});

// Presentation fixtures exercise canonical states only. Actual worker delivery
// will be exercised in 02-21; this test cannot stand in for a provider send.
test("invitation-queue: canonical status presentation fixture, no worker delivery claim", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const headers = {
    Authorization: `Bearer ${fixture.token}`,
    Origin: "http://127.0.0.1:3100",
    "Idempotency-Key": "browser-status-workspace",
  };
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers,
    data: { name: "Status presentation Team" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  const queued = await request.post(
    `${fixture.api}/api/v1/workspaces/${workspace.id}/invitations`,
    {
      headers: { ...headers, "Idempotency-Key": "browser-status-invite-01" },
      data: { email: "status@example.test", role: "member" },
    },
  );
  expect(queued.status()).toBe(201);
  const { invitation } = await queued.json();
  expect(invitation.status).toBe("Queued");
  let projectedStatus = "Queued";
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname.endsWith("/invitations") && route.request().method() === "GET") {
      await route.fulfill({
        json: { items: [{ ...invitation, status: projectedStatus }], nextAfter: null },
      });
    } else {
      const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
      await route.fulfill({ response });
    }
  });
  await page.goto(`/workspaces/${workspace.id}/team`);
  const panel = page.getByRole("region", { name: "Invitations", exact: true });
  for (const [status, label] of [
    ["Queued", "Queued"],
    ["Delivered", "Delivered"],
    ["Failed", "Delivery failed"],
    ["Accepted", "Accepted"],
    ["Expired", "Expired"],
    ["Revoked", "Revoked"],
  ]) {
    projectedStatus = status as string;
    await panel.getByRole("button", { name: "Reload invitations" }).click();
    await expect(panel.getByRole("list")).toContainText(`Delivery status: ${label}`);
    await expect(panel.locator("time")).toHaveAttribute("datetime", invitation.expiresAt);
    await expect(panel).not.toContainText("Sent");
  }
});

test("invitation-queue: fresh SQL roles restrict invitation form and grants", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-invite-capability",
    },
    data: { name: "Invitation capability Team" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
    await route.fulfill({ response });
  });
  for (const role of ["owner", "admin", "member", "viewer"]) {
    expect(
      (
        await request.post(
          `${process.env.FLUX_BROWSER_FIXTURE}/restore-role?workspace=${workspace.id}&role=${role}`,
        )
      ).status(),
    ).toBe(204);
    await page.goto(`/workspaces/${workspace.id}/team`);
    const form = page.getByRole("form", { name: "Invite member" });
    if (role === "owner" || role === "admin") {
      await expect(form.getByLabel("Role", { exact: true })).toHaveValue("member");
      await expect(form.getByLabel("Role", { exact: true }).locator("option")).toHaveText(
        role === "owner" ? ["Admin", "Member", "Viewer"] : ["Member", "Viewer"],
      );
    } else {
      await expect(
        page.getByText("You do not have permission to view Team.", { exact: true }),
      ).toBeVisible();
      await expect(form).toHaveCount(0);
      await expect(page.getByRole("region", { name: "Invitations", exact: true })).toHaveCount(0);
    }
  }
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
  await expect(navigation.getByRole("link", { name: "Team", exact: true })).toHaveAttribute(
    "href",
    /\/team$/,
  );
  await expect(navigation.getByRole("button", { name: "Team", exact: true })).toHaveCount(0);
  await expect(
    page.getByText("Create a managed-domain link to organize its destination and status.", {
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    page.getByRole("link", { name: "Create your first link", exact: true }),
  ).toHaveAttribute("href", `${new URL(page.url()).pathname}/new`);
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

test("library lists real newest-first rows with recovery and immediate membership disposal", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const headers = {
    Authorization: `Bearer ${fixture.token}`,
    Origin: "http://127.0.0.1:3100",
    "Idempotency-Key": "browser-library-workspace",
  };
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers,
    data: { name: "Library browser tenant" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  const links = [];
  for (const index of [0, 1]) {
    const result = await request.post(`${fixture.api}/api/v1/workspaces/${workspace.id}/links`, {
      headers: { ...headers, "Idempotency-Key": `browser-library-link-${index}` },
      data: {
        destination: `https://example.com/library-${index}?a=1&b=2`,
        title: `Library title ${index}`,
      },
    });
    expect(result.status()).toBe(201);
    links.push((await result.json()).link);
  }
  let fail = false;
  let release: (() => void) | undefined;
  const held = new Promise<void>((resolve) => {
    release = resolve;
  });
  let hold = true;
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname.endsWith(`/workspaces/${workspace.id}/links`)) {
      if (hold) await held;
      if (fail) {
        await route.fulfill({ status: 503, json: { message: "PRIVATE-LIBRARY-FAILURE" } });
        return;
      }
      const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
      const body = await response.json();
      // Presentation-only canonical archived state; lifecycle mutation is a later plan.
      if (response.ok() && body.items?.[0]) body.items[0].lifecycle = "archived";
      await route.fulfill({ response, json: body });
      return;
    }
    const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/links`);
  await expect(page.getByText("Loading links…", { exact: true })).toBeVisible();
  hold = false;
  release?.();
  const table = page.getByRole("table", { name: "Links library" });
  await expect(table).toBeVisible();
  await expect(table.getByRole("row")).toHaveCount(3);
  await expect(table.getByRole("row").nth(1)).toContainText("Library title 1");
  await expect(table.getByRole("link", { name: "Library title 1", exact: true })).toHaveAttribute(
    "href",
    `/workspaces/${workspace.id}/links/${links[1].id}`,
  );
  await expect(
    table.getByRole("link", { name: links[1].destination, exact: true }),
  ).toHaveAttribute("href", links[1].destination);
  await expect(table.locator("time").first()).toHaveAttribute("datetime", links[1].createdAt);
  await expect(table.getByRole("row").nth(1)).toContainText("Archived");
  await expect(table.getByRole("row").nth(2)).toContainText("Active");
  fail = true;
  await page.getByRole("button", { name: "Reload links", exact: true }).click();
  await expect(
    page.getByRole("region", { name: "Links", exact: true }).getByRole("alert"),
  ).toHaveText("We couldn't load links. Try again.");
  await expect(table).toBeVisible();
  await expect(page.getByText("PRIVATE-LIBRARY-FAILURE")).toHaveCount(0);
  fail = false;
  await page.getByRole("button", { name: "Retry loading links", exact: true }).click();
  await expect(
    page.getByRole("region", { name: "Links", exact: true }).getByRole("alert"),
  ).toHaveCount(0);
  await page.setViewportSize({ width: 375, height: 812 });
  await expect(table).toBeHidden();
  const cards = page.getByRole("list", { name: "Links library" });
  await expect(cards).toBeVisible();
  await expect(cards.getByRole("listitem")).toHaveCount(2);
  await expect(cards.getByRole("listitem").first()).toContainText("Library title 1");
  await expect(cards.getByRole("listitem").first()).toContainText("Created");
  const removed = await request.post(
    `${process.env.FLUX_BROWSER_FIXTURE}/restore-remove?workspace=${workspace.id}`,
  );
  expect(removed.status()).toBe(204);
  await page.getByRole("button", { name: "Reload links", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Choose a workspace", exact: true }),
  ).toBeVisible();
  await expect(page.getByText("Library title 1", { exact: true })).toHaveCount(0);
});

test("library empty viewer presentation and initial failure never claim empty before success", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-library-empty",
    },
    data: { name: "Empty viewer library" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  let fail = true;
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    if (fail && url.pathname.endsWith(`/workspaces/${workspace.id}/links`)) {
      await route.fulfill({ status: 503, json: {} });
      return;
    }
    const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
    // Test-local presentation projection only; actual viewer SQL authority is
    // independently verified by the registered real-PG HTTP library test.
    if (url.pathname.endsWith(`/workspaces/${workspace.id}`)) {
      const body = await response.json();
      body.workspace.role = "viewer";
      await route.fulfill({ response, json: body });
    } else await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/links`);
  await expect(
    page.getByRole("region", { name: "Links", exact: true }).getByRole("alert"),
  ).toHaveText("We couldn't load links. Try again.");
  await expect(page.getByRole("heading", { name: "No links yet" })).toHaveCount(0);
  fail = false;
  await page.getByRole("button", { name: "Retry loading links", exact: true }).click();
  await expect(page.getByRole("heading", { name: "No links yet" })).toBeVisible();
  await expect(
    page.getByText("You have view-only access. Ask an owner or admin to change your role.", {
      exact: true,
    }),
  ).toBeVisible();
  await expect(page.getByRole("link", { name: "Create your first link" })).toHaveCount(0);
  await expect(page.getByRole("link", { name: "Create link", exact: true })).toHaveCount(0);
});

test("pagination traverses real pages and returns through remembered cursors", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const headers = {
    Authorization: `Bearer ${fixture.token}`,
    Origin: "http://127.0.0.1:3100",
    "Idempotency-Key": "browser-pagination-workspace",
  };
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers,
    data: { name: "Pagination tenant" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  for (let index = 0; index < 26; index++) {
    // Respect the production request budget while creating real records.
    await new Promise((resolve) => setTimeout(resolve, 70));
    const result = await request.post(`${fixture.api}/api/v1/workspaces/${workspace.id}/links`, {
      headers: { ...headers, "Idempotency-Key": `browser-pagination-link-${index}` },
      data: { destination: "https://example.com/pagination", title: `Pagination ${index}` },
    });
    expect(result.status()).toBe(201);
  }
  let invalid = false;
  const queries: string[] = [];
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname.endsWith(`/workspaces/${workspace.id}/links`)) {
      queries.push(url.search);
      if (invalid && url.searchParams.has("cursor")) url.searchParams.set("cursor", "tampered");
    }
    const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/links`);
  const table = page.getByRole("table", { name: "Links library" });
  await expect(table.getByRole("row")).toHaveCount(26);
  await expect(table.getByRole("row").nth(1)).toContainText("Pagination 25");
  const next = page.getByRole("button", { name: "Next page", exact: true });
  const previous = page.getByRole("button", { name: "Previous page", exact: true });
  await expect(previous).toBeDisabled();
  await next.click();
  await expect(table.getByRole("row")).toHaveCount(2);
  await expect(table.getByRole("row").nth(1)).toContainText("Pagination 0");
  await expect(page.getByText("No more links.", { exact: true })).toBeVisible();
  await expect(next).toBeDisabled();
  await previous.click();
  await expect(table.getByRole("row")).toHaveCount(26);
  await expect(previous).toBeDisabled();
  invalid = true;
  await next.click();
  await expect(
    page.getByRole("region", { name: "Links", exact: true }).getByRole("alert"),
  ).toHaveText("This page is no longer available. Return to the first page.");
  await page.getByRole("button", { name: "Return to first page", exact: true }).click();
  await expect(table.getByRole("row")).toHaveCount(26);
  await expect(previous).toBeDisabled();
  expect(queries.filter((query) => query.includes("cursor=")).length).toBeGreaterThanOrEqual(2);
});

test("empty library creation CTA opens the real committed create flow", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: {
      Authorization: `Bearer ${fixture.token}`,
      Origin: "http://127.0.0.1:3100",
      "Idempotency-Key": "browser-empty-cta-workspace",
    },
    data: { name: "Empty creation tenant" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/links`);
  await expect(page.getByText("No links yet", { exact: true })).toBeVisible();
  await page.getByRole("link", { name: "Create your first link", exact: true }).click();
  await page.getByLabel("Destination URL", { exact: true }).fill("https://example.com/first-cta");
  await page.getByLabel("Title", { exact: true }).fill("Created from empty library");
  await page.getByRole("button", { name: "Create link", exact: true }).click();
  await expect(page.getByText("Link created.", { exact: true })).toBeVisible();
  await expect(page.getByText("Created from empty library", { exact: true })).toBeVisible();
});

test("search returns actual literal matches, resets pages and preserves drafts on failures", async ({
  page,
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  await installProviderTransport(page, fixture.client);
  const headers = {
    Authorization: `Bearer ${fixture.token}`,
    Origin: "http://127.0.0.1:3100",
    "Idempotency-Key": "browser-search-workspace",
  };
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers,
    data: { name: "Search browser tenant" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  for (let index = 0; index < 26; index++) {
    await new Promise((resolve) => setTimeout(resolve, 70));
    const response = await request.post(`${fixture.api}/api/v1/workspaces/${workspace.id}/links`, {
      headers: { ...headers, "Idempotency-Key": `browser-search-link-${index}` },
      data: {
        destination: `https://example.com/search-${index}`,
        title: index === 0 ? "Literal 50% under_score" : "Search ordinary " + index,
      },
    });
    expect(response.status()).toBe(201);
  }
  let fail = false;
  let releaseSlow: (() => void) | undefined;
  const heldSlow = new Promise<void>((resolve) => {
    releaseSlow = resolve;
  });
  let observedSlow: (() => void) | undefined;
  const slowStarted = new Promise<void>((resolve) => {
    observedSlow = resolve;
  });
  let finishedSlow: (() => void) | undefined;
  const slowFinished = new Promise<void>((resolve) => {
    finishedSlow = resolve;
  });
  const queries: string[] = [];
  let returnedTitles: string[] = [];
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname.endsWith(`/workspaces/${workspace.id}/links`)) {
      queries.push(url.search);
      if (fail) {
        await route.fulfill({ status: 503, json: {} });
        return;
      }
    }
    const response = await route.fetch({ url: `${fixture.api}${url.pathname}${url.search}` });
    if (
      url.pathname.endsWith(`/workspaces/${workspace.id}/links`) &&
      url.searchParams.get("search") === "ordinary"
    ) {
      observedSlow?.();
      await heldSlow;
      // The new generation aborts this old browser transport; still deliver its
      // actual backend response when possible to prove it cannot repaint rows.
      await route.fulfill({ response }).catch(() => {});
      finishedSlow?.();
      return;
    }
    if (url.pathname.endsWith(`/workspaces/${workspace.id}/links`) && response.ok())
      returnedTitles = (await response.json()).items.map((item: { title: string }) => item.title);
    await route.fulfill({ response });
  });
  await page.goto(`/workspaces/${workspace.id}/links`);
  const table = page.getByRole("table", { name: "Links library" });
  await expect(table.getByRole("row")).toHaveCount(26);
  await page.getByRole("button", { name: "Next page", exact: true }).click();
  await expect(table.getByRole("row")).toHaveCount(2);
  const search = page.getByLabel("Search links", { exact: true });
  await search.fill("%");
  await page.getByRole("button", { name: "Search links", exact: true }).click();
  await expect(table.getByRole("row")).toHaveCount(2);
  await expect(table.getByRole("row").nth(1)).toContainText("Literal 50% under_score");
  expect(returnedTitles).toEqual(["Literal 50% under_score"]);
  expect(new URLSearchParams(queries.at(-1)).has("cursor")).toBe(false);
  await expect(page.getByRole("button", { name: "Previous page", exact: true })).toBeDisabled();
  fail = true;
  await search.fill("typed draft");
  await page.getByRole("button", { name: "Search links", exact: true }).click();
  await expect(
    page.getByRole("region", { name: "Links", exact: true }).getByRole("alert"),
  ).toHaveText("We couldn't load links. Try again.");
  await expect(search).toHaveValue("typed draft");
  await expect(table.getByRole("row").nth(1)).toContainText("Literal 50% under_score");
  fail = false;
  await page.getByRole("button", { name: "Retry loading links", exact: true }).click();
  await expect(page.getByRole("heading", { name: "No matching links", exact: true })).toBeVisible();
  await page.getByRole("button", { name: "Clear filters", exact: true }).click();
  await expect(table.getByRole("row")).toHaveCount(26);
  await search.fill("_");
  await page.getByRole("button", { name: "Search links", exact: true }).click();
  await expect(table.getByRole("row")).toHaveCount(2);
  expect(returnedTitles).toEqual(["Literal 50% under_score"]);
  await search.fill("ordinary");
  await page.getByRole("button", { name: "Search links", exact: true }).click();
  await slowStarted;
  await expect(page.getByText("Loading links…", { exact: true })).toBeVisible();
  await search.fill("%");
  await page.getByRole("button", { name: "Search links", exact: true }).click();
  await expect(table.getByRole("row")).toHaveCount(2);
  await expect(page.getByText("Loading links…", { exact: true })).toHaveCount(0);
  releaseSlow?.();
  await slowFinished;
  await expect(table.getByRole("row")).toHaveCount(2);
  await expect(table.getByRole("row").nth(1)).toContainText("Literal 50% under_score");
  await page.getByRole("button", { name: "Clear search", exact: true }).click();
  await expect(table.getByRole("row")).toHaveCount(26);
  await page.getByLabel("Lifecycle", { exact: true }).selectOption("deleted");
  await expect(page.getByRole("heading", { name: "No deleted links", exact: true })).toBeVisible();
  await expect(
    page.getByText("Deleted links appear here and can be restored.", { exact: true }),
  ).toBeVisible();
  expect(returnedTitles).toEqual([]);
  for (const state of ["active", "disabled", "archived", "nondeleted"]) {
    await page.getByLabel("Lifecycle", { exact: true }).selectOption(state);
    if (["active", "nondeleted"].includes(state))
      await expect(table.getByRole("row")).toHaveCount(26);
    else
      await expect(
        page.getByRole("heading", { name: "No matching links", exact: true }),
      ).toBeVisible();
  }
});
