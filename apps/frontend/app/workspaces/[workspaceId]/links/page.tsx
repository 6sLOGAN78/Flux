"use client";

import { useParams } from "next/navigation";
import { LinkList } from "../../../../components/link-list";
import { WorkspaceSwitcher } from "../../../../components/workspace-switcher";
import { workspaceCapabilities } from "../../../../lib/workspace";

export default function LinksPage() {
  const { workspaceId } = useParams<{ workspaceId: string }>();
  return (
    <WorkspaceSwitcher key={workspaceId} workspaceId={workspaceId}>
      {(workspace, accessLost) => (
        <LinkList
          key={workspace.id}
          workspaceId={workspace.id}
          canCreate={workspaceCapabilities(workspace.role).create}
          accessLost={accessLost}
        />
      )}
    </WorkspaceSwitcher>
  );
}
