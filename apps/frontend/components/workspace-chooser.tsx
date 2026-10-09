"use client";

import { useAuth, useClerk } from "@clerk/nextjs";
import { ZIdentityResponse, type IdentityResponse } from "@flux/zod";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState, type FormEvent } from "react";
import { ApiError, createAPI } from "../lib/api";
import {
  selectWorkspace,
  invalidateWorkspaces,
  listenWorkspaceInvalidation,
  workspaceChannel,
  WorkspaceRequests,
} from "../lib/workspace";

export function WorkspaceChooser({ accessChanged = false }: { accessChanged?: boolean }) {
  const { isLoaded, isSignedIn, sessionId, getToken } = useAuth();
  const { signOut } = useClerk();
  const router = useRouter();
  const [snapshot, setSnapshot] = useState<{ identity: IdentityResponse; session: string }>();
  const [selection, setSelection] = useState("");
  const [attempt, setAttempt] = useState(0);
  const [error, setError] = useState("");
  const [expired, setExpired] = useState(false);
  const [busy, setBusy] = useState(false);
  const [closing, setClosing] = useState(false);
  const [changed, setChanged] = useState(accessChanged);
  const requests = useRef(new WorkspaceRequests());
  const channel = useRef<BroadcastChannel | null>(null);
  const active = useRef<AbortController | null>(null);
  const currentSession = useRef(sessionId);
  const submitting = useRef(false);
  const heading = useRef<HTMLHeadingElement>(null);
  currentSession.current = sessionId;

  useEffect(() => {
    if (new URLSearchParams(window.location.search).get("access") === "changed") setChanged(true);
  }, []);

  useEffect(() => {
    if (!isLoaded) return;
    if (!isSignedIn || !sessionId || closing) {
      requests.current.clear();
      active.current?.abort();
      setSnapshot(undefined);
      setSelection("");
      return;
    }
    const request = requests.current.begin();
    setError("");
    setExpired(false);
    const api = createAPI(getToken, { origin: window.location.origin });
    void api.request("/me", { signal: request.signal }).then(
      (result) => {
        if (!request.current()) return;
        const parsed = ZIdentityResponse.safeParse(result);
        if (parsed.success) setSnapshot({ identity: parsed.data, session: sessionId });
        else setError("We couldn't load workspaces. Try again.");
      },
      (failure: unknown) => {
        if (!request.current()) return;
        if (failure instanceof ApiError && failure.code === "unauthenticated") {
          setSnapshot(undefined);
          setSelection("");
          setExpired(true);
        } else setError("We couldn't load workspaces. Try again.");
      },
    );
    return () => {
      requests.current.clear();
      active.current?.abort();
    };
  }, [isLoaded, isSignedIn, sessionId, getToken, attempt, closing]);

  useEffect(() => {
    const broadcast =
      typeof BroadcastChannel === "undefined" ? null : new BroadcastChannel(workspaceChannel);
    channel.current = broadcast;
    const stop = broadcast
      ? listenWorkspaceInvalidation(broadcast, () => {
          requests.current.clear();
          active.current?.abort();
          setSnapshot(undefined);
          setSelection("");
          setChanged(true);
          setAttempt((value) => value + 1);
        })
      : () => {};
    const refresh = () => {
      requests.current.clear();
      active.current?.abort();
      setAttempt((value) => value + 1);
    };
    window.addEventListener("focus", refresh);
    return () => {
      stop();
      broadcast?.close();
      channel.current = null;
      window.removeEventListener("focus", refresh);
    };
  }, []);

  useEffect(() => {
    if (snapshot?.session === sessionId) heading.current?.focus();
  }, [snapshot, sessionId]);

  const open = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (submitting.current || !selection || !snapshot || snapshot.session !== sessionId) return;
    submitting.current = true;
    setBusy(true);
    setError("");
    const controller = new AbortController();
    requests.current.clear();
    active.current?.abort();
    active.current = controller;
    const session = sessionId;
    try {
      const workspace = await selectWorkspace(
        createAPI(getToken, { origin: window.location.origin }),
        selection,
        controller.signal,
      );
      if (controller.signal.aborted || currentSession.current !== session) return;
      setSnapshot(undefined);
      setSelection("");
      invalidateWorkspaces(channel.current);
      router.push(`/workspaces/${workspace.id}/links`);
    } catch (failure) {
      if (controller.signal.aborted || currentSession.current !== session) return;
      if (failure instanceof ApiError && [403, 404].includes(failure.status)) {
        setSnapshot(undefined);
        setSelection("");
        setChanged(true);
        setAttempt((value) => value + 1);
      } else if (failure instanceof ApiError && failure.code === "unauthenticated") {
        setSnapshot(undefined);
        setExpired(true);
      } else setError("Workspace could not be opened. Try again.");
    } finally {
      submitting.current = false;
      setBusy(false);
    }
  };

  const logout = async () => {
    active.current?.abort();
    requests.current.clear();
    setSnapshot(undefined);
    setSelection("");
    setClosing(true);
    invalidateWorkspaces(channel.current);
    try {
      await signOut();
    } catch {
      setError("Sign-in is temporarily unavailable. Try again shortly.");
    }
  };

  const identity =
    snapshot && snapshot.session === sessionId && !closing ? snapshot.identity : undefined;
  return (
    <main className="auth-panel" aria-busy={!identity && !error && !expired}>
      <p className="wordmark">Flux</p>
      {!isLoaded ? (
        <p role="status">Loading sign-in…</p>
      ) : !isSignedIn || expired || (closing && !error) ? (
        <>
          <h1>Sign in to Flux</h1>
          {expired && <p role="alert">Your session ended. Sign in to continue.</p>}
          <a href="/sign-in">Sign in</a>
        </>
      ) : (
        <>
          <h1 ref={heading} tabIndex={-1}>
            Choose a workspace
          </h1>
          {changed && (
            <>
              <p role="status">Your workspace access changed. Choose an available workspace.</p>
              <p>You do not have access to this workspace.</p>
            </>
          )}
          {error && <p role="alert">{error}</p>}
          {identity ? (
            <>
              <p>{identity.user.email}</p>
              {identity.workspaces.length ? (
                <form onSubmit={(event) => void open(event)} aria-busy={busy}>
                  <label htmlFor="workspace-choice">Workspace</label>
                  <select
                    id="workspace-choice"
                    required
                    value={selection}
                    disabled={busy}
                    onChange={(event) => setSelection(event.target.value)}
                  >
                    <option value="">Choose a workspace</option>
                    {identity.workspaces.map((workspace) => (
                      <option key={workspace.id} value={workspace.id}>
                        {workspace.name} ({workspace.role})
                      </option>
                    ))}
                  </select>
                  <button type="submit" disabled={busy}>
                    {busy ? "Opening workspace…" : "Open workspace"}
                  </button>
                </form>
              ) : (
                <p>You don't have a workspace yet. Create one or accept an invitation.</p>
              )}
              <a href="/onboarding">Create workspace</a>
              <button type="button" onClick={() => void logout()}>
                Sign out
              </button>
            </>
          ) : error ? (
            <button
              type="button"
              onClick={() => (closing ? void logout() : setAttempt((value) => value + 1))}
            >
              Retry loading workspaces
            </button>
          ) : (
            <p role="status">Loading workspaces…</p>
          )}
        </>
      )}
    </main>
  );
}
