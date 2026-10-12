"use client";

import type { WorkspaceResponse } from "@flux/zod";
import { usePathname } from "next/navigation";
import { useEffect, useRef, type ReactNode } from "react";
import { workspaceCapabilities } from "../lib/workspace";

export function AppShell({
  workspace,
  children,
  switcher,
  section = "Links",
}: {
  workspace: WorkspaceResponse["workspace"];
  children: ReactNode;
  switcher?: ReactNode;
  section?: "Links" | "Team";
}) {
  const capabilities = workspaceCapabilities(workspace.role);
  const pathname = usePathname();
  const main = useRef<HTMLElement>(null);
  useEffect(() => {
    const heading = main.current?.querySelector<HTMLElement>("h2");
    if (heading) {
      heading.tabIndex = -1;
      heading.focus();
    }
  }, [pathname]);
  return (
    <>
      <a className="skip-link" href="#workspace-main">
        Skip to main content
      </a>
      <header className="auth-panel">
        <p className="wordmark">Flux</p>
        <h1>{workspace.name}</h1>
        {switcher}
        <nav aria-label="Workspace">
          <a
            href={`/workspaces/${workspace.id}/links`}
            aria-current={section === "Links" ? "page" : undefined}
          >
            Links
          </a>
          {capabilities.team && (
            <a
              href={`/workspaces/${workspace.id}/team`}
              aria-current={section === "Team" ? "page" : undefined}
            >
              Team
            </a>
          )}
          <a href="/">Account</a>
        </nav>
      </header>
      <main
        ref={main}
        id="workspace-main"
        tabIndex={-1}
        className="auth-panel"
        aria-label={section}
      >
        {children}
      </main>
    </>
  );
}
