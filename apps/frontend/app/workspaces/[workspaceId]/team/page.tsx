"use client";

import { useParams } from "next/navigation";
import { TeamList } from "../../../../components/team-list";
import { WorkspaceSwitcher } from "../../../../components/workspace-switcher";
import { workspaceCapabilities } from "../../../../lib/workspace";

export default function TeamPage() {
  const { workspaceId } = useParams<{ workspaceId: string }>();
  return (
    <WorkspaceSwitcher key={workspaceId} workspaceId={workspaceId} section="Team">
      {(workspace, accessLost) => (
        <TeamList
          key={`${workspace.id}:${workspace.role}`}
          workspaceId={workspace.id}
          canInspect={workspaceCapabilities(workspace.role).team}
          accessLost={accessLost}
        />
      )}
    </WorkspaceSwitcher>
  );
}
