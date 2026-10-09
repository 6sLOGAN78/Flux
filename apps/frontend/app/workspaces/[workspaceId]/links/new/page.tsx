"use client";

import { useAuth } from "@clerk/nextjs";
import { ZLinkResponse, ZWorkspaceResponse } from "@flux/zod";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useRef, useState, type FormEvent } from "react";
import { WorkspaceSwitcher } from "../../../../../components/workspace-switcher";
import { ApiError, createAPI } from "../../../../../lib/api";
import {
  WorkspaceRequests,
  resolveWorkspace,
  workspaceCapabilities,
} from "../../../../../lib/workspace";

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

function LinkForm({
  workspaceId,
  onDirty,
  accessLost,
}: {
  workspaceId: string;
  onDirty: (value: boolean) => void;
  accessLost: (status: number) => void;
}) {
  const { getToken } = useAuth();
  const router = useRouter();
  const requests = useRef(new WorkspaceRequests());
  const denied = useRef(accessLost);
  denied.current = accessLost;
  const retry = useRef<{ destination: string; title: string; key: string } | undefined>(undefined);
  const submitting = useRef(false);
  const [destination, setDestination] = useState("");
  const [title, setTitle] = useState("");
  const [host, setHost] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [attempt, setAttempt] = useState(0);
  const [uncertain, setUncertain] = useState(false);
  const errorSummary = useRef<HTMLParagraphElement>(null);
  useEffect(() => {
    const request = requests.current.begin();
    void createAPI(getToken, { origin: window.location.origin })
      .request(`/workspaces/${workspaceId}`, { signal: request.signal })
      .then((value) => {
        if (!request.current()) return;
        const response = ZWorkspaceResponse.parse(value);
        if (!response.managedHost) throw new Error("Missing managed hostname");
        setHost(response.managedHost);
        setError("");
      })
      .catch(async (failure: unknown) => {
        if (!request.current()) return;
        if (failure instanceof ApiError && failure.status === 401) {
          denied.current(401);
          return;
        }
        if (failure instanceof ApiError && [403, 404].includes(failure.status)) {
          try {
            await resolveWorkspace(
              createAPI(getToken, { origin: window.location.origin }),
              workspaceId,
              request.signal,
            );
          } catch {
            if (request.current()) denied.current(404);
            return;
          }
        }
        if (request.current()) setError("Links could not be loaded. Try again.");
      });
    return () => {
      requests.current.clear();
    };
  }, [workspaceId, getToken, attempt, onDirty]);
  useEffect(() => () => onDirty(false), [onDirty]);
  useEffect(() => {
    onDirty(!!destination || !!title);
  }, [destination, title, onDirty]);
  useEffect(() => {
    if (error) errorSummary.current?.focus();
  }, [error]);

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (submitting.current || !host) return;
    if ([...title].length > 200) {
      setError("Enter a title of at most 200 characters.");
      return;
    }
    const draft =
      retry.current?.destination === destination && retry.current.title === title
        ? retry.current
        : { destination, title, key: crypto.randomUUID() };
    retry.current = draft;
    submitting.current = true;
    setBusy(true);
    setError("");
    const request = requests.current.begin();
    const api = createAPI(getToken, { origin: window.location.origin });
    try {
      const result = ZLinkResponse.parse(
        await api.request(`/workspaces/${workspaceId}/links`, {
          method: "POST",
          body: { destination: draft.destination, title: draft.title },
          idempotencyKey: draft.key,
          signal: request.signal,
        }),
      );
      if (!request.current()) return;
      if (result.link.workspaceId !== workspaceId) throw new Error("Unexpected workspace");
      onDirty(false);
      router.push(`/workspaces/${workspaceId}/links/${result.link.id}?created=1`);
    } catch (failure) {
      if (!request.current()) return;
      if (failure instanceof ApiError && failure.status === 401) {
        accessLost(401);
        return;
      }
      if (failure instanceof ApiError && [403, 404].includes(failure.status)) {
        try {
          await resolveWorkspace(api, workspaceId, request.signal);
        } catch {
          if (request.current()) {
            accessLost(404);
          }
          return;
        }
        if (!request.current()) return;
        setError("You don't have permission to do this. Ask a workspace owner or admin for help.");
      } else if (failure instanceof ApiError && failure.status === 400) {
        setUncertain(false);
        setError("Enter a public HTTP(S) destination and a title of at most 200 characters.");
      } else {
        setUncertain(true);
        setError("We couldn't confirm link creation. Retry to check the same request.");
      }
    } finally {
      submitting.current = false;
      if (request.current()) setBusy(false);
    }
  };
  return (
    <section aria-labelledby="new-link-title">
      <h2 id="new-link-title">Create link</h2>
      <p>Link management is available. Redirects and analytics are not available yet.</p>
      <p>Managed hostname: {host || "Loading…"}</p>
      {error && (
        <p ref={errorSummary} tabIndex={-1} role="alert">
          {error}
        </p>
      )}
      {!host && error && (
        <button type="button" onClick={() => setAttempt((value) => value + 1)}>
          Retry loading links
        </button>
      )}
      <form onSubmit={(event) => void submit(event)} aria-busy={busy}>
        <label htmlFor="link-destination">Destination URL</label>
        <input
          id="link-destination"
          type="url"
          required
          value={destination}
          disabled={busy}
          onChange={(event) => {
            setDestination(event.target.value);
            setUncertain(false);
          }}
        />
        <label htmlFor="link-title">Title</label>
        <input
          id="link-title"
          value={title}
          disabled={busy}
          onChange={(event) => {
            setTitle(event.target.value);
            setUncertain(false);
          }}
          aria-describedby="link-title-help"
        />
        <p id="link-title-help">
          Optional; at most 200 characters. A short key will be generated securely.
        </p>
        <button type="submit" disabled={busy || !host}>
          {busy ? "Creating link…" : uncertain ? "Retry creation" : "Create link"}
        </button>
      </form>
      <a href={`/workspaces/${workspaceId}/links`}>Back to links</a>
    </section>
  );
}
