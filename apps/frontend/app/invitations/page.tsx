"use client";

import { useAuth, useClerk } from "@clerk/nextjs";
import type { InvitationInspectionResponse } from "@flux/zod";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { createAPI } from "../../lib/api";
import {
  acceptInvitation,
  forgetInvitation,
  inspectInvitation,
  invitationArrival,
  invitationFailure,
} from "../../lib/invitation";
import { WorkspaceRequests } from "../../lib/workspace";

export default function InvitationPage() {
  const { isLoaded, isSignedIn, sessionId, getToken } = useAuth();
  const clerk = useClerk();
  const router = useRouter();
  const requests = useRef(new WorkspaceRequests());
  const [token, setToken] = useState("");
  const [review, setReview] = useState<{
    session: string;
    value: InvitationInspectionResponse["invitation"];
  }>();
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [switching, setSwitching] = useState(false);
  const [switchError, setSwitchError] = useState("");
  const closing = useRef(false);
  const activeSession = useRef(sessionId);
  activeSession.current = sessionId;
  const [attempt, setAttempt] = useState(0);
  const current =
    !switching && isSignedIn && review && review.session === sessionId ? review.value : undefined;

  useEffect(() => {
    setToken(invitationArrival());
  }, []);
  useEffect(() => {
    setReview(undefined);
    setError("");
    setBusy(false);
    const request = requests.current.begin();
    if (isLoaded && !token)
      setError("This invitation is no longer valid. Ask an owner or admin for a new invitation.");
    if (isLoaded && isSignedIn && sessionId && token && !closing.current && !switchError) {
      inspectInvitation(
        createAPI(getToken, { origin: window.location.origin }),
        token,
        request.signal,
      )
        .then((value) => {
          if (request.current() && !closing.current && activeSession.current === sessionId)
            setReview({ session: sessionId, value });
        })
        .catch((failure: unknown) => {
          if (request.current() && !closing.current && activeSession.current === sessionId)
            setError(invitationFailure(failure));
        });
    }
    return () => requests.current.clear();
  }, [isLoaded, isSignedIn, sessionId, token, getToken, attempt, switching, switchError]);

  const openSignIn = () =>
    clerk.openSignIn({ forceRedirectUrl: "/invitations", signUpForceRedirectUrl: "/invitations" });
  const switchAccount = async () => {
    if (closing.current) return;
    closing.current = true;
    requests.current.clear();
    setReview(undefined);
    setError("");
    setSwitchError("");
    setBusy(false);
    setSwitching(true);
    try {
      await clerk.signOut(() => openSignIn());
    } catch {
      setSwitchError("Sign-in is temporarily unavailable. Try again shortly.");
    } finally {
      closing.current = false;
      setSwitching(false);
    }
  };

  const accept = async () => {
    if (!current || busy || !sessionId || closing.current) return;
    const request = requests.current.begin();
    setBusy(true);
    setError("");
    try {
      const workspace = await acceptInvitation(
        createAPI(getToken, { origin: window.location.origin }),
        token,
        request.signal,
      );
      if (!request.current() || closing.current || activeSession.current !== sessionId) return;
      forgetInvitation();
      setToken("");
      setReview(undefined);
      router.replace(`/workspaces/${workspace.id}/links`);
    } catch (failure) {
      if (request.current()) {
        setReview(undefined);
        setError(invitationFailure(failure));
      }
    } finally {
      if (request.current()) setBusy(false);
    }
  };

  return (
    <main className="auth-panel" aria-busy={busy || switching}>
      <p className="wordmark">Flux</p>
      <h1>{current ? `Join ${current.workspaceName}` : "Workspace invitation"}</h1>
      {(error || switchError) && <p role="alert">{switchError || error}</p>}
      {switchError && (
        <button type="button" onClick={switchAccount}>
          Retry sign-out
        </button>
      )}
      {switching && <p role="status">Switching account…</p>}
      {!isLoaded && !switching && <p role="status">Loading sign-in…</p>}
      {isLoaded && !isSignedIn && token && !switching && (
        <>
          <p>Sign in with the invited, verified email address to review this invitation.</p>
          <button type="button" onClick={openSignIn}>
            Sign in
          </button>
        </>
      )}
      {isSignedIn && token && !current && !error && !switchError && !switching && (
        <p role="status">Loading invitation…</p>
      )}
      {current && (
        <>
          <p>
            Role:{" "}
            {current.role === "admin" ? "Admin" : current.role === "member" ? "Member" : "Viewer"}
          </p>
          <p>
            Expires:{" "}
            <time dateTime={current.expiresAt}>{new Date(current.expiresAt).toLocaleString()}</time>
          </p>
          {current.status === "Accepted" && <p role="status">Invitation already accepted.</p>}
          <button type="button" disabled={busy} onClick={accept}>
            {busy
              ? "Accepting invitation…"
              : current.status === "Accepted"
                ? "Open workspace"
                : "Accept invitation"}
          </button>
        </>
      )}
      {isSignedIn && error && token && !switching && !switchError && (
        <button type="button" onClick={() => setAttempt((value) => value + 1)}>
          Reload invitation
        </button>
      )}
      {isSignedIn && token && !switching && (
        <button type="button" disabled={busy} onClick={switchAccount}>
          Switch account
        </button>
      )}
      <p>
        <a href="/workspaces">Choose a workspace</a>
      </p>
    </main>
  );
}
