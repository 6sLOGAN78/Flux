"use client";

import { useAuth } from "@clerk/nextjs";
import { ZWorkspaceResponse } from "@flux/zod";
import { useRouter } from "next/navigation";
import { type FormEvent, useRef, useState } from "react";
import { createAPI } from "../../lib/api";

export default function OnboardingPage() {
  const { isLoaded, isSignedIn, sessionId, getToken } = useAuth();
  const router = useRouter();
  const [name, setName] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const submitting = useRef(false);
  const activeSession = useRef(sessionId);
  activeSession.current = sessionId;

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (submitting.current || !sessionId) return;
    const trimmed = name.trim();
    if ([...trimmed].length < 1 || [...trimmed].length > 100) {
      setError("Enter a workspace name of 1–100 characters.");
      return;
    }
    submitting.current = true;
    setBusy(true);
    setError("");
    const session = sessionId;
    try {
      // The retry key and canonical draft survive reloads and uncertain delivery;
      // changed content gets a new key, while unchanged retries reuse the ledger.
      const storageKey = `flux.workspace-bootstrap.${session}`;
      const stored = sessionStorage.getItem(storageKey);
      const previous = stored ? (JSON.parse(stored) as { name?: string; key?: string }) : null;
      const key =
        previous?.name === trimmed && /^[A-Za-z0-9_-]{16,128}$/.test(previous.key ?? "")
          ? (previous.key as string)
          : crypto.randomUUID();
      sessionStorage.setItem(storageKey, JSON.stringify({ name: trimmed, key }));
      const api = createAPI(getToken, { origin: window.location.origin });
      const result = ZWorkspaceResponse.parse(
        await api.request("/workspaces", {
          method: "POST",
          body: { name: trimmed },
          idempotencyKey: key,
        }),
      );
      if (activeSession.current !== session) return;
      sessionStorage.removeItem(storageKey);
      router.push(`/workspaces/${result.workspace.id}/links`);
    } catch {
      if (activeSession.current === session) setError("Workspace could not be created. Try again.");
    } finally {
      submitting.current = false;
      setBusy(false);
    }
  };

  return (
    <main className="auth-panel">
      <p className="wordmark">Flux</p>
      {!isLoaded ? (
        <p role="status">Loading sign-in…</p>
      ) : !isSignedIn ? (
        <>
          <h1>Sign in to Flux</h1>
          <a href="/sign-in">Sign in to continue</a>
        </>
      ) : (
        <>
          <h1>Create your workspace</h1>
          <form onSubmit={(event) => void submit(event)} aria-busy={busy}>
            <label htmlFor="workspace-name">Workspace name</label>
            <p id="workspace-help">Choose a name your team will recognize.</p>
            <input
              id="workspace-name"
              name="name"
              required
              value={name}
              disabled={busy}
              aria-describedby="workspace-help workspace-error"
              aria-invalid={!!error}
              onChange={(event) => setName(event.target.value)}
            />
            <p id="workspace-error" role={error ? "alert" : undefined}>
              {error}
            </p>
            <button type="submit" disabled={busy}>
              {busy ? "Creating workspace…" : "Create workspace"}
            </button>
          </form>
        </>
      )}
    </main>
  );
}
