import { expect } from "@playwright/test";
import { describe, test } from "node:test";
import { renderToStaticMarkup } from "react-dom/server";
import { createElement } from "react";
import { AppShell } from "../components/app-shell";
import { workspaceCapabilities } from "./workspace";

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
