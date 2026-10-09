"use client";

import { useParams } from "next/navigation";
import { EmptyState } from "../../../../components/empty-state";
import { WorkspaceSwitcher } from "../../../../components/workspace-switcher";
import { workspaceCapabilities } from "../../../../lib/workspace";

export default function LinksPage() {
  const { workspaceId } = useParams<{ workspaceId: string }>();
  return (
    <WorkspaceSwitcher key={workspaceId} workspaceId={workspaceId}>
      {(workspace) => (
        <>
          <p>Links</p>
          <EmptyState canCreate={workspaceCapabilities(workspace.role).create} />
        </>
      )}
    </WorkspaceSwitcher>
  );
}
