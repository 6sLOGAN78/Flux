import { expect, test } from "@playwright/test";

test("tenant-denial foreign HTTP DTO and error surfaces never disclose private fields", async ({
  request,
}) => {
  const fixture = await (await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/client`)).json();
  const foreign = await (
    await request.get(`${process.env.FLUX_BROWSER_FIXTURE}/member-client`)
  ).json();
  const headers = { Authorization: `Bearer ${fixture.token}`, Origin: "http://127.0.0.1:3100" };
  const seed = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: { ...headers, "Idempotency-Key": "privacy-identity-seed" },
    data: { name: "Foreign identity fixture" },
  });
  expect(seed.status()).toBe(201);
  const seedWorkspace = (await seed.json()).workspace;
  expect(
    (
      await request.post(
        `${process.env.FLUX_BROWSER_FIXTURE}/restore-member?workspace=${seedWorkspace.id}`,
      )
    ).status(),
  ).toBe(204);
  const foreignIdentity = await request.get(`${fixture.api}/api/v1/me`, {
    headers: { Authorization: `Bearer ${foreign.token}` },
  });
  expect(foreignIdentity.status()).toBe(200);
  const created = await request.post(`${fixture.api}/api/v1/workspaces`, {
    headers: { ...headers, "Idempotency-Key": "privacy-tenant-workspace" },
    data: { name: "Private HTTP workspace" },
  });
  expect(created.status()).toBe(201);
  const { workspace } = await created.json();
  const result = await request.post(`${fixture.api}/api/v1/workspaces/${workspace.id}/links`, {
    headers: { ...headers, "Idempotency-Key": "privacy-tenant-link" },
    data: { destination: "https://example.com/private-http", title: "Private HTTP title" },
  });
  expect(result.status()).toBe(201);
  const { link } = await result.json();
  const paths = [
    `/workspaces/${workspace.id}`,
    `/workspaces/${workspace.id}/links`,
    `/workspaces/${workspace.id}/links/${link.id}`,
    `/workspaces/${workspace.id}/members`,
    `/workspaces/${workspace.id}/invitations`,
  ];
  for (const path of paths) {
    const denied = await request.get(`${fixture.api}/api/v1${path}`, {
      headers: { Authorization: `Bearer ${foreign.token}` },
    });
    expect(denied.status()).toBe(404);
    expect(denied.headers()["cache-control"]).toBe("no-store");
    const body = await denied.text();
    expect(
      [
        workspace.id,
        workspace.name,
        link.id,
        link.destination,
        link.title,
        fixture.token,
        "local@example.test",
      ].some((value) => body.includes(value)),
    ).toBe(false);
  }
});
