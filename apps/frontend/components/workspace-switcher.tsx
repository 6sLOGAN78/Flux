"use client";

import { useAuth, useClerk } from "@clerk/nextjs";
import type { WorkspaceResponse } from "@flux/zod";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState, type FormEvent, type ReactNode } from "react";
import { ApiError, createAPI } from "../lib/api";
import {
  WorkspaceRequests,
  invalidateWorkspaces,
  listenWorkspaceInvalidation,
  resolveWorkspace,
  selectWorkspace,
  workspaceChannel,
} from "../lib/workspace";
import { AppShell } from "./app-shell";
import { WorkspaceChooser } from "./workspace-chooser";
import { ConfirmDialog } from "./confirm-dialog";

type Workspace = WorkspaceResponse["workspace"];

// Own the scoped subtree: clearing this snapshot unmounts all tenant state.
export function WorkspaceSwitcher({
  workspaceId,
  children,
  dirty = false,
  section = "Links",
}: {
  workspaceId: string;
  children: (
    workspace: Workspace,
    accessLost: (status: number) => void,
    membershipChanged: () => void,
  ) => ReactNode;
  dirty?: boolean;
  section?: "Links" | "Team";
}) {
  const { isLoaded, isSignedIn, sessionId, getToken } = useAuth();
  const { signOut } = useClerk();
  const router = useRouter();
  const requests = useRef(new WorkspaceRequests());
  const channel = useRef<BroadcastChannel | null>(null);
  const [snapshot, setSnapshot] = useState<{
    workspace: Workspace;
    workspaces: Workspace[];
    session: string;
  }>();
  const [selection, setSelection] = useState("");
  const [attempt, setAttempt] = useState(0);
  const [chooser, setChooser] = useState(false);
  const [changed, setChanged] = useState(false);
  const [closing, setClosing] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [confirmSwitch, setConfirmSwitch] = useState(false);
  const submitting = useRef(false);

  const scrub = () => {
    requests.current.clear();
    setSnapshot(undefined);
    setSelection("");
    setError("");
  };

  useEffect(() => {
    if (!isLoaded) return;
    if (!isSignedIn || !sessionId || closing) {
      scrub();
      return;
    }
    if (chooser) return;
    const request = requests.current.begin();
    void resolveWorkspace(
      createAPI(getToken, { origin: window.location.origin }),
      workspaceId,
      request.signal,
    ).then(
      (result) => {
        if (!request.current()) return;
        setSnapshot({ ...result, session: sessionId });
        setError("");
      },
      (failure: unknown) => {
        if (!request.current()) return;
        if (failure instanceof ApiError && failure.code === "unauthenticated") {
          scrub();
          setClosing(true);
          invalidateWorkspaces(channel.current);
        } else if (failure instanceof ApiError && [403, 404].includes(failure.status)) {
          scrub();
          setChanged(true);
          setChooser(true);
          invalidateWorkspaces(channel.current);
          router.replace("/workspaces?access=changed");
        } else setError("Workspace could not be loaded. Try again.");
      },
    );
    return () => requests.current.clear();
  }, [isLoaded, isSignedIn, sessionId, getToken, workspaceId, attempt, chooser, closing]);

  useEffect(() => {
    const broadcast =
      typeof BroadcastChannel === "undefined" ? null : new BroadcastChannel(workspaceChannel);
    channel.current = broadcast;
    const stop = broadcast
      ? listenWorkspaceInvalidation(broadcast, () => {
          scrub();
          setChanged(true);
          setChooser(true);
          router.replace("/workspaces?access=changed");
        })
      : () => {};
    const focus = () => {
      requests.current.clear();
      setAttempt((value) => value + 1);
    };
    window.addEventListener("focus", focus);
    return () => {
      stop();
      broadcast?.close();
      channel.current = null;
      requests.current.clear();
      window.removeEventListener("focus", focus);
    };
  }, []);

  const open = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (dirty) {
      setConfirmSwitch(true);
      return;
    }
    await openWorkspace();
  };

  const openWorkspace = async () => {
    if (submitting.current || !selection || !snapshot || snapshot.session !== sessionId) return;
    const target = selection;
    submitting.current = true;
    scrub();
    setBusy(true);
    const request = requests.current.begin();
    try {
      const workspace = await selectWorkspace(
        createAPI(getToken, { origin: window.location.origin }),
        target,
        request.signal,
      );
      if (!request.current()) return;
      invalidateWorkspaces(channel.current);
      router.push(`/workspaces/${workspace.id}/links`);
    } catch (failure) {
      if (!request.current()) return;
      if (failure instanceof ApiError && failure.code === "unauthenticated") setClosing(true);
      else {
        setChooser(true);
        setChanged(failure instanceof ApiError && [403, 404].includes(failure.status));
      }
    } finally {
      submitting.current = false;
      if (request.current()) setBusy(false);
    }
  };

  const logout = async () => {
    scrub();
    setClosing(true);
    invalidateWorkspaces(channel.current);
    try {
      await signOut();
    } catch {
      setError("Sign-in is temporarily unavailable. Try again shortly.");
    }
  };

  if (isLoaded && (!isSignedIn || closing))
    return (
      <main className="auth-panel">
        <h1>Sign in to Flux</h1>
        {error && (
          <>
            <p role="alert">{error}</p>
            <button type="button" onClick={() => void logout()}>
              Retry sign out
            </button>
          </>
        )}
        <a href="/sign-in">Sign in</a>
      </main>
    );
  if (chooser) return <WorkspaceChooser accessChanged={changed} />;
  const workspace =
    snapshot && snapshot.session === sessionId && snapshot.workspace.id === workspaceId
      ? snapshot.workspace
      : undefined;
  if (workspace && snapshot)
    return (
      <AppShell
        workspace={workspace}
        section={section}
        switcher={
          <>
            <form onSubmit={(event) => void open(event)}>
              <label htmlFor="switch-workspace">Switch workspace</label>
              <select
                id="switch-workspace"
                value={selection}
                onChange={(event) => setSelection(event.target.value)}
                required
              >
                <option value="">Choose a workspace</option>
                {snapshot.workspaces
                  .filter((item) => item.id !== workspace.id)
                  .map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name} ({item.role})
                    </option>
                  ))}
              </select>
              <button type="submit">Open workspace</button>
            </form>
            <button type="button" onClick={() => void logout()}>
              Sign out
            </button>
          </>
        }
      >
        {error && (
          <>
            <p role="alert">{error}</p>
            <button type="button" onClick={() => setAttempt((value) => value + 1)}>
              Retry
            </button>
          </>
        )}
        {confirmSwitch && (
          <ConfirmDialog
            onStay={() => setConfirmSwitch(false)}
            onDiscard={() => {
              setConfirmSwitch(false);
              void openWorkspace();
            }}
          />
        )}
        {children(
          workspace,
          (status) => {
            scrub();
            if (status === 403 && section === "Team") {
              setAttempt((value) => value + 1);
              return;
            }
            invalidateWorkspaces(channel.current);
            if (status === 401) setClosing(true);
            else {
              setChanged(true);
              setChooser(true);
              router.replace("/workspaces?access=changed");
            }
          },
          () => invalidateWorkspaces(channel.current),
        )}
      </AppShell>
    );
  return (
    <main className="auth-panel" aria-busy={busy || !error}>
      {error ? (
        <>
          <p role="alert">{error}</p>
          <button type="button" onClick={() => setAttempt((value) => value + 1)}>
            Retry
          </button>
        </>
      ) : (
        <p role="status">Loading workspace…</p>
      )}
    </main>
  );
}
