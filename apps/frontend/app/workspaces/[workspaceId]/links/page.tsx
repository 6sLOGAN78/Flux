"use client";

import { useAuth } from "@clerk/nextjs";
import { ZWorkspaceResponse, type WorkspaceResponse } from "@flux/zod";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { EmptyState } from "../../../../components/empty-state";
import { AppShell } from "../../../../components/app-shell";
import { ApiError, createAPI } from "../../../../lib/api";
import { workspaceCapabilities } from "../../../../lib/workspace";

export default function LinksPage() {
  const { workspaceId } = useParams<{ workspaceId: string }>();
  const { isLoaded, isSignedIn, sessionId, getToken } = useAuth();
  const [state, setState] = useState<{
    workspace?: WorkspaceResponse["workspace"];
    session?: string;
    id?: string;
    error?: string;
  }>({});
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    if (!isLoaded || !isSignedIn || !sessionId) return;
    const controller = new AbortController();
    setState({});
    const api = createAPI(getToken, { origin: window.location.origin });
    void api.request(`/workspaces/${workspaceId}`, { signal: controller.signal }).then(
      (result) => {
        if (controller.signal.aborted) return;
        const parsed = ZWorkspaceResponse.safeParse(result);
        setState(
          parsed.success
            ? { workspace: parsed.data.workspace, session: sessionId, id: workspaceId }
            : { error: "Workspace could not be loaded. Try again." },
        );
      },
      (error: unknown) => {
        if (controller.signal.aborted) return;
        setState({
          error:
            error instanceof ApiError && [403, 404].includes(error.status)
              ? "You do not have access to this workspace."
              : "Workspace could not be loaded. Try again.",
        });
      },
    );
    return () => controller.abort();
  }, [isLoaded, isSignedIn, sessionId, getToken, workspaceId, attempt]);

  // Focus revalidation clears tenant content before a fresh authoritative read.
  useEffect(() => {
    const refresh = () => {
      setState({});
      setAttempt((value) => value + 1);
    };
    window.addEventListener("focus", refresh);
    return () => window.removeEventListener("focus", refresh);
  }, []);

  if (
    isLoaded &&
    isSignedIn &&
    state.workspace &&
    state.session === sessionId &&
    state.id === workspaceId
  ) {
    return (
      <AppShell workspace={state.workspace}>
        <p>Links</p>
        <EmptyState canCreate={workspaceCapabilities(state.workspace.role).create} />
      </AppShell>
    );
  }

  return (
    <main className="auth-panel">
      <p className="wordmark">Flux</p>
      {!isLoaded ? (
        <p role="status">Loading sign-in…</p>
      ) : !isSignedIn ? (
        <a href="/sign-in">Sign in to continue</a>
      ) : state.error ? (
        <>
          <h1>Links</h1>
          <p role="alert">{state.error}</p>
          <button type="button" onClick={() => setAttempt((value) => value + 1)}>
            Retry
          </button>
          <a href="/onboarding">Create workspace</a>
        </>
      ) : (
        <p role="status">Loading workspace…</p>
      )}
    </main>
  );
}
