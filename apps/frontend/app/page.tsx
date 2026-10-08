"use client";

import { ClerkFailed, ClerkLoaded, ClerkLoading, useAuth, useClerk } from "@clerk/nextjs";
import { ZIdentityResponse, type IdentityResponse } from "@flux/zod";
import { useEffect, useRef, useState } from "react";
import { ApiError, createAPI } from "../lib/api";

type AccountState =
  | { kind: "loading" }
  | { kind: "expired" }
  | { kind: "unavailable" }
  | { kind: "signed-out" }
  | { kind: "ready"; identity: IdentityResponse; sessionId: string };

function Account() {
  const { isLoaded, isSignedIn, sessionId, getToken } = useAuth();
  const { signOut } = useClerk();
  const [state, setState] = useState<AccountState>({ kind: "loading" });
  const [attempt, setAttempt] = useState(0);
  const [signingOut, setSigningOut] = useState(false);
  const active = useRef<AbortController | null>(null);
  const closingSession = useRef(false);

  useEffect(() => {
    if (!isLoaded || !isSignedIn || !sessionId || signingOut || closingSession.current) return;
    const controller = new AbortController();
    active.current = controller;
    setState({ kind: "loading" });
    const api = createAPI(getToken, { origin: window.location.origin });
    void api.request("/me", { signal: controller.signal }).then(
      (response) => {
        if (controller.signal.aborted) return;
        const parsed = ZIdentityResponse.safeParse(response);
        setState(
          parsed.success
            ? { kind: "ready", identity: parsed.data, sessionId }
            : { kind: "unavailable" },
        );
      },
      (error: unknown) => {
        if (controller.signal.aborted) return;
        setState({
          kind:
            error instanceof ApiError && error.code === "unauthenticated"
              ? "expired"
              : "unavailable",
        });
      },
    );
    return () => {
      controller.abort();
      if (active.current === controller) active.current = null;
    };
  }, [isLoaded, isSignedIn, sessionId, getToken, attempt, signingOut]);

  const logout = async () => {
    closingSession.current = true;
    active.current?.abort();
    setState({ kind: "signed-out" });
    setSigningOut(true);
    try {
      await signOut();
    } catch {
      setState({ kind: "unavailable" });
    } finally {
      setSigningOut(false);
    }
  };

  if (!isLoaded) return <p role="status">Loading sign-in…</p>;
  if (!isSignedIn || state.kind === "signed-out") {
    return (
      <>
        <h1>Sign in to Flux</h1>
        <p>Sign in to continue.</p>
        <a href="/sign-in">Sign in</a>
      </>
    );
  }
  if (state.kind === "expired") {
    return (
      <>
        <h1>Sign in to Flux</h1>
        <p role="alert">Your session ended. Sign in to continue.</p>
        <a href="/sign-in">Sign in</a>
      </>
    );
  }
  if (state.kind === "unavailable") {
    return (
      <>
        <h1>Your account</h1>
        <p role="alert">Sign-in is temporarily unavailable. Try again shortly.</p>
        <button
          type="button"
          disabled={signingOut}
          onClick={() =>
            closingSession.current ? void logout() : setAttempt((value) => value + 1)
          }
        >
          Retry
        </button>
      </>
    );
  }
  if (state.kind === "loading" || (state.kind === "ready" && state.sessionId !== sessionId))
    return (
      <>
        <h1>Your account</h1>
        <p role="status">Loading your account…</p>
      </>
    );
  return (
    <>
      <h1>Your account</h1>
      <p>{state.identity.user.email}</p>
      <button type="button" disabled={signingOut} onClick={() => void logout()}>
        Sign out
      </button>
    </>
  );
}

export default function HomePage() {
  return (
    <main className="auth-panel">
      <p className="wordmark">Flux</p>
      <ClerkLoading>
        <p role="status">Loading sign-in…</p>
      </ClerkLoading>
      <ClerkFailed>
        <h1>Sign in to Flux</h1>
        <p role="alert">Sign-in is temporarily unavailable. Try again shortly.</p>
        <a href="/sign-in">Retry sign-in</a>
      </ClerkFailed>
      <ClerkLoaded>
        <Account />
      </ClerkLoaded>
    </main>
  );
}
