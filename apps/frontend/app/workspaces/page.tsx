"use client";

import { useAuth, useClerk } from "@clerk/nextjs";
import { ZIdentityResponse, type IdentityResponse } from "@flux/zod";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState, type FormEvent } from "react";
import { ApiError, createAPI } from "../../lib/api";
import { selectWorkspace } from "../../lib/workspace";

export default function WorkspacesPage() {
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
  const [changed, setChanged] = useState(false);
  const active = useRef<AbortController | null>(null);
  const currentSession = useRef(sessionId);
  const submitting = useRef(false);
  const heading = useRef<HTMLHeadingElement>(null);
  currentSession.current = sessionId;

  useEffect(() => {
    if (!isLoaded || !isSignedIn || !sessionId || closing) return;
    const controller = new AbortController();
    active.current = controller;
    setSnapshot(undefined);
    setSelection("");
    setError("");
    setExpired(false);
    const api = createAPI(getToken, { origin: window.location.origin });
    void api.request("/me", { signal: controller.signal }).then(
      (result) => {
        if (controller.signal.aborted) return;
        const parsed = ZIdentityResponse.safeParse(result);
        if (parsed.success) setSnapshot({ identity: parsed.data, session: sessionId });
        else setError("We couldn't load workspaces. Try again.");
      },
      (failure: unknown) => {
        if (controller.signal.aborted) return;
        if (failure instanceof ApiError && failure.code === "unauthenticated") setExpired(true);
        else setError("We couldn't load workspaces. Try again.");
      },
    );
    return () => {
      controller.abort();
      active.current?.abort();
    };
  }, [isLoaded, isSignedIn, sessionId, getToken, attempt, closing]);

  useEffect(() => {
    const refresh = () => {
      active.current?.abort();
      setSnapshot(undefined);
      setSelection("");
      setAttempt((value) => value + 1);
    };
    window.addEventListener("focus", refresh);
    return () => window.removeEventListener("focus", refresh);
  }, []);

  useEffect(() => {
    if (snapshot?.session === sessionId) heading.current?.focus();
  }, [snapshot, sessionId]);

  const open = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (submitting.current || !selection || snapshot?.session !== sessionId) return;
    submitting.current = true;
    setBusy(true);
    setError("");
    const controller = new AbortController();
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
    setSnapshot(undefined);
    setSelection("");
    setClosing(true);
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
            <p role="status">Your workspace access changed. Choose an available workspace.</p>
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
