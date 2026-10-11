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
  const [attempt, setAttempt] = useState(0);
  const current = review && review.session === sessionId ? review.value : undefined;

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
    if (isLoaded && isSignedIn && sessionId && token) {
      inspectInvitation(
        createAPI(getToken, { origin: window.location.origin }),
        token,
        request.signal,
      )
        .then((value) => {
          if (request.current()) setReview({ session: sessionId, value });
        })
        .catch((failure: unknown) => {
          if (request.current()) setError(invitationFailure(failure));
        });
    }
    return () => requests.current.clear();
  }, [isLoaded, isSignedIn, sessionId, token, getToken, attempt]);

  const accept = async () => {
    if (!current || busy || !sessionId) return;
    const request = requests.current.begin();
    setBusy(true);
    setError("");
    try {
      const workspace = await acceptInvitation(
        createAPI(getToken, { origin: window.location.origin }),
        token,
        request.signal,
      );
      if (!request.current()) return;
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
    <main className="auth-panel" aria-busy={busy}>
      <p className="wordmark">Flux</p>
      <h1>{current ? `Join ${current.workspaceName}` : "Workspace invitation"}</h1>
      {error && <p role="alert">{error}</p>}
      {!isLoaded && <p role="status">Loading sign-in…</p>}
      {isLoaded && !isSignedIn && token && (
        <>
          <p>Sign in with the invited, verified email address to review this invitation.</p>
          <button
            type="button"
            onClick={() =>
              clerk.openSignIn({
                forceRedirectUrl: "/invitations",
                signUpForceRedirectUrl: "/invitations",
              })
            }
          >
            Sign in
          </button>
        </>
      )}
      {isSignedIn && token && !current && !error && <p role="status">Loading invitation…</p>}
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
      {isSignedIn && error && token && (
        <button type="button" onClick={() => setAttempt((value) => value + 1)}>
          Retry invitation
        </button>
      )}
      {isSignedIn && error && token && (
        <button
          type="button"
          onClick={() =>
            clerk.signOut(() =>
              clerk.openSignIn({
                forceRedirectUrl: "/invitations",
                signUpForceRedirectUrl: "/invitations",
              }),
            )
          }
        >
          Switch account
        </button>
      )}
      <p>
        <a href="/workspaces">Choose a workspace</a>
      </p>
    </main>
  );
}
