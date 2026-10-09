import { expect } from "@playwright/test";
import { describe, test } from "node:test";
import { renderToStaticMarkup } from "react-dom/server";
import { createElement } from "react";
import { AppShell } from "../components/app-shell";
import {
  workspaceCapabilities,
  WorkspaceRequests,
  listenWorkspaceInvalidation,
  invalidateWorkspaces,
  resolveWorkspace,
} from "./workspace";
import { ApiError } from "./api";

describe("switching generations and signal-only invalidation", () => {
  test("aborts previous requests and rejects late success even when transport ignores abort", async () => {
    const requests = new WorkspaceRequests();
    const old = requests.begin();
    const next = requests.begin();
    expect(old.signal.aborted).toBe(true);
    expect(old.current()).toBe(false);
    expect(next.current()).toBe(true);
    requests.clear();
    expect(next.signal.aborted).toBe(true);
    expect(next.current()).toBe(false);
  });

  test("broadcasts only a fixed invalidation and never accepts workspace payloads", () => {
    const sent: unknown[] = [];
    const channel = new EventTarget() as EventTarget & { postMessage(value: unknown): void };
    channel.postMessage = (value) => sent.push(value);
    let cleared = 0;
    const stop = listenWorkspaceInvalidation(channel, () => cleared++);
    channel.dispatchEvent(
      new MessageEvent("message", { data: { workspaceId: "foreign", name: "private" } }),
    );
    expect(cleared).toBe(0);
    invalidateWorkspaces(channel);
    expect(sent).toEqual(["invalidate"]);
    channel.dispatchEvent(new MessageEvent("message", { data: "invalidate" }));
    expect(cleared).toBe(1);
    stop();
    channel.dispatchEvent(new MessageEvent("message", { data: "invalidate" }));
    expect(cleared).toBe(1);
  });

  test("a missing scoped resource confirms membership instead of revoking valid workspace", async () => {
    const workspace = {
      id: "00000000-0000-4000-8000-000000000001",
      name: "Growth",
      role: "viewer" as const,
    };
    const identity = {
      authenticated: true,
      user: { id: "00000000-0000-4000-8000-000000000002", email: "local@example.test" },
      workspaces: [workspace],
      lastWorkspace: workspace,
    };
    const paths: string[] = [];
    const api = {
      request: async (path: string) => {
        paths.push(path);
        if (path === "/me") return identity;
        throw new ApiError("request_failed", 404);
      },
    };
    const result = await resolveWorkspace(api as never, workspace.id, new AbortController().signal);
    expect(result.workspace).toEqual(workspace);
    expect(paths).toEqual(["/me", `/workspaces/${workspace.id}`, "/me"]);
    identity.workspaces = [];
    await expect(
      resolveWorkspace(api as never, workspace.id, new AbortController().signal),
    ).rejects.toMatchObject({ status: 404 });
  });
});

describe("workspace capability presentation", () => {
  test("closed matrix never grants unknown roles", () => {
    for (const role of ["owner", "admin", "member", "viewer", "unknown", "OWNER", ""]) {
      const valid = ["owner", "admin", "member", "viewer"].includes(role);
      expect(workspaceCapabilities(role)).toEqual({
        read: valid,
        create: valid && role !== "viewer",
        team: ["owner", "admin"].includes(role),
      });
    }
  });

  test("shell exposes Team only to administrators without a fabricated route", () => {
    for (const role of ["owner", "admin", "member", "viewer"] as const) {
      const html = renderToStaticMarkup(
        createElement(AppShell, {
          workspace: { id: "00000000-0000-4000-8000-000000000000", name: "Growth", role },
          children: "No links yet",
        }),
      );
      expect(html).toContain('aria-label="Workspace"');
      expect(html).toContain('aria-current="page"');
      expect(html).toContain("<main");
      expect(html.includes('id="team-availability"')).toBe(["owner", "admin"].includes(role));
      expect(html).not.toContain("/team");
    }
  });
});
