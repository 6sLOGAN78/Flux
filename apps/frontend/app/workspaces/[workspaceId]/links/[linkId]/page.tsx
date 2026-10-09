"use client";

import { useParams } from "next/navigation";
import { WorkspaceSwitcher } from "../../../../../components/workspace-switcher";
import { LinkDetail } from "../../../../../components/link-detail";

export default function LinkDetailPage() {
  const { workspaceId, linkId } = useParams<{ workspaceId: string; linkId: string }>();
  return (
    <WorkspaceSwitcher key={workspaceId} workspaceId={workspaceId}>
      {(_, accessLost) => (
        <LinkDetail
          key={linkId}
          workspaceId={workspaceId}
          linkId={linkId}
          accessLost={accessLost}
        />
      )}
    </WorkspaceSwitcher>
  );
}
