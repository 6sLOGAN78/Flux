"use client";

import { useParams } from "next/navigation";
import { useState } from "react";
import { workspaceCapabilities } from "../../../../../lib/workspace";
import { WorkspaceSwitcher } from "../../../../../components/workspace-switcher";
import { LinkForm } from "../../../../../components/link-form";

export default function NewLinkPage() {
  const { workspaceId } = useParams<{ workspaceId: string }>();
  const [dirty, setDirty] = useState(false);
  return (
    <WorkspaceSwitcher key={workspaceId} workspaceId={workspaceId} dirty={dirty}>
      {(workspace, accessLost) =>
        workspaceCapabilities(workspace.role).create ? (
          <LinkForm workspaceId={workspaceId} onDirty={setDirty} accessLost={accessLost} />
        ) : (
          <>
            <p role="alert">
              You don't have permission to do this. Ask a workspace owner or admin for help.
            </p>
            <a href={`/workspaces/${workspaceId}/links`}>Back to links</a>
          </>
        )
      }
    </WorkspaceSwitcher>
  );
}
