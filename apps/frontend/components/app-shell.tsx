import type { WorkspaceResponse } from "@flux/zod";
import type { ReactNode } from "react";
import { workspaceCapabilities } from "../lib/workspace";

export function AppShell({
  workspace,
  children,
}: {
  workspace: WorkspaceResponse["workspace"];
  children: ReactNode;
}) {
  const capabilities = workspaceCapabilities(workspace.role);
  return (
    <>
      <header className="auth-panel">
        <p className="wordmark">Flux</p>
        <h1>{workspace.name}</h1>
        <nav aria-label="Workspace">
          <a href={`/workspaces/${workspace.id}/links`} aria-current="page">
            Links
          </a>
          {capabilities.team && (
            <span>
              <button type="button" disabled aria-describedby="team-availability">
                Team
              </button>
              <span id="team-availability">
                Team management will be available in a later update.
              </span>
            </span>
          )}
          <a href="/">Account</a>
        </nav>
      </header>
      <main className="auth-panel" aria-label="Links">
        {children}
      </main>
    </>
  );
}
