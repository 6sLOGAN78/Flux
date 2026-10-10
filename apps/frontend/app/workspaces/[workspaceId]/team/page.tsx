"use client";

import { useParams } from "next/navigation";
import { useState } from "react";
import { InvitationPanel } from "../../../../components/invitation-panel";
import { TeamList } from "../../../../components/team-list";
import { WorkspaceSwitcher } from "../../../../components/workspace-switcher";
import { workspaceCapabilities } from "../../../../lib/workspace";

export default function TeamPage() {
  const { workspaceId } = useParams<{ workspaceId: string }>();
  const [dirty, setDirty] = useState(false);
  return (
    <WorkspaceSwitcher key={workspaceId} workspaceId={workspaceId} section="Team" dirty={dirty}>
      {(workspace, accessLost, membershipChanged) => (
        <>
          <TeamList
            key={`${workspace.id}:${workspace.role}`}
            workspaceId={workspace.id}
            actorRole={workspace.role}
            canInspect={workspaceCapabilities(workspace.role).team}
            accessLost={accessLost}
            membershipChanged={membershipChanged}
          />
          {workspaceCapabilities(workspace.role).team && (
            <InvitationPanel
              key={`invites:${workspace.id}:${workspace.role}`}
              workspaceId={workspace.id}
              actorRole={workspace.role}
              accessLost={accessLost}
              onDirty={setDirty}
            />
          )}
        </>
      )}
    </WorkspaceSwitcher>
  );
}
