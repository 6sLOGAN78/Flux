import { expect, test } from "@playwright/test";
import {
  ZIdentityResponse,
  ZInvitationResponse,
  ZMembersResponse,
  ZWorkspaceResponse,
} from "@flux/zod";
import {
  providerAPI,
  providerControl,
  providerLogin,
  providerRow,
  readProviderInvitation,
} from "./provider.config";

test("real provider named owner, encrypted rotation, one delivery and invited tenant disposal", async ({
  page,
  browser,
}, info) => {
  test.setTimeout(120000);
  await providerLogin(page);
  await page.goto("/onboarding");
  await page.getByLabel("Workspace name", { exact: true }).fill("Provider acceptance workspace");
  const created = page.waitForResponse(
    (response) => response.url().endsWith("/api/v1/workspaces") && response.status() === 201,
  );
  await page.getByRole("button", { name: "Create workspace", exact: true }).click();
  const workspace = ZWorkspaceResponse.parse(await (await created).json()).workspace;
  expect(workspace.role === "owner").toBe(true);
  const ownership = await providerControl("owner-proof", { workspace: workspace.id });
  expect(ownership.owners === 1 && ownership.workspaces === 1).toBe(true);
  providerRow(info, "named_workspace_owner");
  await expect(
    page.getByRole("link", { name: "Create your first link", exact: true }),
  ).toBeVisible();
  providerRow(info, "links_empty_state");
  expect(
    (await providerAPI(page, "/me/last-workspace", "PUT", { workspaceId: workspace.id })).status,
  ).toBe(200);
  await page.goto("/");
  await expect(page).toHaveURL(new RegExp(`/workspaces/${workspace.id}/links$`));
  providerRow(info, "last_used_workspace");

  const queued = await providerAPI(page, `/workspaces/${workspace.id}/invitations`, "POST", {
    email: process.env.FLUX_PROVIDER_INVITEE_EMAIL,
    role: "member",
  });
  expect(queued.status).toBe(201);
  const invite = ZInvitationResponse.parse(queued.body).invitation;
  const before = await providerControl("delivery-proof", {
    workspace: workspace.id,
    invitation: invite.id,
  });
  expect(before.queued === true && before.old_key === true && before.erased === false).toBe(true);
  await providerControl("rotate-start-worker");
  await expect
    .poll(
      async () => {
        const result = await providerControl("delivery-proof", {
          workspace: workspace.id,
          invitation: invite.id,
        });
        return result.delivered === true && result.erased === true && result.old_key === true;
      },
      { timeout: 30000 },
    )
    .toBe(true);
  await providerControl("stop-worker");
  providerRow(info, "invitation_delivery");

  // Prove new writes use the rotated key without sending a second email.
  const second = await providerAPI(page, "/workspaces", "POST", {
    name: "Provider rotation cancellation",
  });
  expect(second.status).toBe(201);
  const rotationWorkspace = ZWorkspaceResponse.parse(second.body).workspace;
  const newQueued = await providerAPI(
    page,
    `/workspaces/${rotationWorkspace.id}/invitations`,
    "POST",
    { email: process.env.FLUX_PROVIDER_INVITEE_EMAIL, role: "viewer" },
  );
  expect(newQueued.status).toBe(201);
  const newInvite = ZInvitationResponse.parse(newQueued.body).invitation;
  const after = await providerControl("delivery-proof", {
    workspace: rotationWorkspace.id,
    invitation: newInvite.id,
  });
  expect(after.queued === true && after.new_key === true && after.erased === false).toBe(true);
  expect(
    (
      await providerAPI(
        page,
        `/workspaces/${rotationWorkspace.id}/invitations/${newInvite.id}`,
        "DELETE",
        { version: newInvite.version },
      )
    ).status,
  ).toBe(200);
  providerRow(info, "invitation_key_rotation");

  const hash = await readProviderInvitation();
  const invited = await browser.newContext({ baseURL: "http://localhost:3000" });
  try {
    const first = await invited.newPage();
    await providerLogin(first, "INVITEE");
    await first.goto(`/invitations${hash}`);
    await expect(
      first.getByRole("button", { name: "Accept invitation", exact: true }),
    ).toBeVisible();
    expect(new URL(first.url()).hash === "").toBe(true);
    const identity = ZIdentityResponse.parse((await providerAPI(first, "/me")).body);
    expect(identity.workspaces.some((entry) => entry.id === workspace.id)).toBe(false);
    await first.getByRole("button", { name: "Accept invitation", exact: true }).click();
    await expect(first).toHaveURL(new RegExp(`/workspaces/${workspace.id}/links$`));
    await expect(
      first.getByRole("link", { name: "Create your first link", exact: true }),
    ).toBeVisible();
    providerRow(info, "invitation_onboarding");
    const other = await invited.newPage();
    await other.goto(`/workspaces/${workspace.id}/links`);
    await expect(
      other.getByRole("link", { name: "Create your first link", exact: true }),
    ).toBeVisible();
    const members = ZMembersResponse.parse(
      (await providerAPI(page, `/workspaces/${workspace.id}/members`)).body,
    );
    const member = members.items.find((entry) => entry.id === identity.user.id);
    expect(member?.role === "member").toBe(true);
    expect(
      (
        await providerAPI(
          page,
          `/workspaces/${workspace.id}/members/${identity.user.id}`,
          "DELETE",
          {},
        )
      ).status,
    ).toBe(200);
    // Both mounted tabs revalidate on focus; stale last-used preference grants no access.
    for (const tab of [first, other]) {
      await tab.bringToFront();
      await tab.evaluate(() => window.dispatchEvent(new Event("focus")));
      await expect(tab).toHaveURL(/\/workspaces(?:\?access=changed)?$/);
      expect((await tab.locator("main").innerText()).includes(workspace.name)).toBe(false);
    }
    const fresh = ZIdentityResponse.parse((await providerAPI(first, "/me")).body);
    expect(
      fresh.lastWorkspace === null && !fresh.workspaces.some((entry) => entry.id === workspace.id),
    ).toBe(true);
    providerRow(info, "membership_loss_tabs");
  } finally {
    await invited.close();
  }
});
