import type { WorkspaceResponse } from "@flux/zod";
import type { ReactNode } from "react";
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
  return (
    <>
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
      <main className="auth-panel" aria-label={section}>
        {children}
      </main>
    </>
  );
}
