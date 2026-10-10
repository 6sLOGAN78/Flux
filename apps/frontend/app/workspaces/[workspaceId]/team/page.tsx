"use client";

import { useParams } from "next/navigation";
import { TeamList } from "../../../../components/team-list";
import { WorkspaceSwitcher } from "../../../../components/workspace-switcher";
import { workspaceCapabilities } from "../../../../lib/workspace";

export default function TeamPage() {
  const { workspaceId } = useParams<{ workspaceId: string }>();
  return (
    <WorkspaceSwitcher key={workspaceId} workspaceId={workspaceId} section="Team">
      {(workspace, accessLost, membershipChanged) => (
        <TeamList
          key={`${workspace.id}:${workspace.role}`}
          workspaceId={workspace.id}
          actorRole={workspace.role}
          canInspect={workspaceCapabilities(workspace.role).team}
          accessLost={accessLost}
          membershipChanged={membershipChanged}
        />
      )}
    </WorkspaceSwitcher>
  );
}
