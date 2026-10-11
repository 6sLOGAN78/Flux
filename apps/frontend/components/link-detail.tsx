"use client";

import { useAuth } from "@clerk/nextjs";
import { ZLinkResponse, type LinkResponse } from "@flux/zod";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { ApiError, createAPI } from "../lib/api";
import {
  destinationSyntaxOK,
  localLinkTime,
  LifecycleRequests,
  type LifecycleAction,
} from "../lib/links";
import { WorkspaceRequests, resolveWorkspace } from "../lib/workspace";
import { workspaceCapabilities } from "../lib/workspace";
import { ConfirmDialog } from "./confirm-dialog";

export function LinkDetail({
  workspaceId,
  linkId,
  accessLost,
}: {
  workspaceId: string;
  linkId: string;
  accessLost: (status: number) => void;
}) {
  const { getToken } = useAuth();
  const router = useRouter();
  const requests = useRef(new WorkspaceRequests());
  const denied = useRef(accessLost);
  denied.current = accessLost;
  const [link, setLink] = useState<LinkResponse["link"]>();
  const [error, setError] = useState("");
  const [copied, setCopied] = useState("");
  const [attempt, setAttempt] = useState(0);
  const [created, setCreated] = useState(false);
  const [canWrite, setCanWrite] = useState(false);
  const [busy, setBusy] = useState(false);
  const submitting = useRef(false);
  const [notice, setNotice] = useState("");
  const [requiresReload, setRequiresReload] = useState(false);
  const [reviewReady, setReviewReady] = useState(false);
  const [uncertain, setUncertain] = useState(false);
  const mutations = useRef(new LifecycleRequests());
  const retry = useRef<HTMLButtonElement>(null);
  const [confirm, setConfirm] = useState<"archive" | "delete">();
  const reload = useRef<HTMLButtonElement>(null);
  const focusAfterCommit = useRef(false);
  useEffect(() => {
    if (focusAfterCommit.current && !busy && !confirm) {
      focusAfterCommit.current = false;
      (uncertain ? retry.current : reload.current)?.focus();
    }
  }, [busy, confirm, uncertain, requiresReload]);
  useEffect(() => {
    setCreated(new URLSearchParams(window.location.search).get("created") === "1");
    const request = requests.current.begin();
    mutations.current.clear();
    setUncertain(false);
    setReviewReady(false);
    const api = createAPI(getToken, { origin: window.location.origin });
    setBusy(true);
    void api
      .request(`/workspaces/${workspaceId}/links/${linkId}`, { signal: request.signal })
      .then(async (value) => {
        const response = ZLinkResponse.parse(value);
        const membership = await resolveWorkspace(api, workspaceId, request.signal);
        if (!request.current()) return;
        if (response.link.workspaceId !== workspaceId || response.link.id !== linkId)
          throw new Error("Unexpected resource");
        setLink(response.link);
        setCanWrite(workspaceCapabilities(membership.workspace.role).create);
        setError("");
        setRequiresReload(false);
        setBusy(false);
        submitting.current = false;
      })
      .catch(async (failure: unknown) => {
        if (!request.current()) return;
        setBusy(false);
        if (failure instanceof ApiError && failure.status === 401) {
          denied.current(401);
        } else if (failure instanceof ApiError && [403, 404].includes(failure.status)) {
          setLink(undefined);
          try {
            await resolveWorkspace(api, workspaceId, request.signal);
          } catch {
            if (request.current()) denied.current(404);
            return;
          }
          if (request.current()) setError("Link not found");
        } else if (request.current()) setError("We couldn't load this link. Try again.");
      });
    return () => {
      requests.current.clear();
      mutations.current.clear();
    };
  }, [workspaceId, linkId, getToken, attempt, router]);
  const transition = async (action: LifecycleAction, identicalRetry = false) => {
    if (!link || !canWrite || submitting.current || (requiresReload && !identicalRetry)) return;
    const mutation = identicalRetry
      ? mutations.current.retry()
      : mutations.current.start(action, link.version);
    if (!mutation) return;
    submitting.current = true;
    setBusy(true);
    setConfirm(undefined);
    setNotice("");
    setError("");
    const request = requests.current.begin();
    const api = createAPI(getToken, { origin: window.location.origin });
    try {
      const value = await api.request(`/workspaces/${workspaceId}/links/${linkId}/transitions`, {
        method: "POST",
        body: mutation.body,
        idempotencyKey: mutation.key,
        signal: request.signal,
      });
      const result = ZLinkResponse.parse(value);
      if (!request.current()) return;
      if (result.link.workspaceId !== workspaceId || result.link.id !== linkId)
        throw new Error("Unexpected resource");
      setLink(result.link);
      mutations.current.clear();
      setUncertain(false);
      setRequiresReload(false);
      setNotice("Link updated.");
    } catch (failure) {
      if (!request.current()) return;
      if (failure instanceof ApiError && failure.status === 401) {
        setLink(undefined);
        denied.current(401);
      } else if (failure instanceof ApiError && [403, 404].includes(failure.status)) {
        mutations.current.clear();
        setUncertain(false);
        setCanWrite(false);
        try {
          const current = await resolveWorkspace(api, workspaceId, request.signal);
          if (!request.current()) return;
          setCanWrite(workspaceCapabilities(current.workspace.role).create);
          setError(
            failure.status === 403
              ? "You have view-only access. Ask an owner or admin to change your role."
              : "Link not found",
          );
          if (failure.status === 404) setLink(undefined);
        } catch {
          if (request.current()) {
            setLink(undefined);
            denied.current(404);
          }
        }
      } else {
        setRequiresReload(true);
        const versionConflict =
          failure instanceof ApiError && failure.constraint === "VERSION_CONFLICT";
        const networkUncertain =
          failure instanceof ApiError && failure.status === 0 && failure.code === "unavailable";
        setUncertain(networkUncertain);
        if (!networkUncertain) mutations.current.clear();
        setError(
          versionConflict
            ? "This link changed since you opened it. Review the latest details before trying again."
            : failure instanceof ApiError && failure.constraint === "PRECONDITION_REQUIRED"
              ? "We couldn't verify the latest version. Reload this link before trying again."
              : "We couldn't save this change. Try again.",
        );
        if (versionConflict) {
          try {
            const latest = ZLinkResponse.parse(
              await api.request(`/workspaces/${workspaceId}/links/${linkId}`, {
                signal: request.signal,
              }),
            );
            const membership = await resolveWorkspace(api, workspaceId, request.signal);
            if (!request.current()) return;
            if (latest.link.workspaceId !== workspaceId || latest.link.id !== linkId)
              throw new Error("Unexpected resource");
            setLink(latest.link);
            setCanWrite(workspaceCapabilities(membership.workspace.role).create);
            setReviewReady(true);
          } catch (refreshFailure) {
            if (!request.current()) return;
            if (refreshFailure instanceof ApiError && refreshFailure.status === 401) {
              setLink(undefined);
              denied.current(401);
            } else if (
              refreshFailure instanceof ApiError &&
              [403, 404].includes(refreshFailure.status)
            ) {
              setLink(undefined);
              try {
                const membership = await resolveWorkspace(api, workspaceId, request.signal);
                if (!request.current()) return;
                setCanWrite(workspaceCapabilities(membership.workspace.role).create);
                setError("Link not found");
              } catch {
                if (request.current()) denied.current(404);
              }
            }
          }
        }
      }
    } finally {
      if (request.current()) {
        submitting.current = false;
        setBusy(false);
        focusAfterCommit.current = true;
      }
    }
  };
  const copy = async () => {
    if (!link) return;
    const request = requests.current.begin();
    try {
      await navigator.clipboard.writeText(link.shortUrl);
      if (request.current()) setCopied("Short URL copied.");
    } catch {
      if (request.current()) setCopied("Couldn't copy. Select and copy the short URL below.");
    }
  };
  const timestamp = (value: string) => <time dateTime={value}>{localLinkTime(value)}</time>;
  return (
    <section aria-labelledby="link-detail-title" aria-busy={busy}>
      <h2 id="link-detail-title">Link details</h2>
      <p>Link management is available. Redirects and analytics are not available yet.</p>
      {error && <p role="alert">{error}</p>}
      {notice && <p role="status">{notice}</p>}
      {uncertain && (
        <>
          <p>The result is uncertain. Retry the identical change to check its committed result.</p>
          <button
            ref={retry}
            type="button"
            disabled={busy}
            onClick={() => {
              const pending = mutations.current.retry();
              if (pending) void transition(pending.body.action, true);
            }}
          >
            Retry identical change
          </button>
        </>
      )}
      {requiresReload && (
        <button
          ref={reload}
          type="button"
          disabled={busy}
          onClick={() => {
            setNotice("");
            mutations.current.clear();
            setUncertain(false);
            if (reviewReady) {
              focusAfterCommit.current = true;
              setRequiresReload(false);
              setReviewReady(false);
              setError("");
            } else setAttempt((value) => value + 1);
          }}
        >
          Review latest details
        </button>
      )}
      {error === "Link not found" && <p>Choose a link from your workspace.</p>}
      {error === "Your session ended. Sign in to continue." && <a href="/sign-in">Sign in</a>}
      {error === "We couldn't load this link. Try again." && (
        <button type="button" onClick={() => setAttempt((value) => value + 1)}>
          Reload link
        </button>
      )}
      {!link && !error && <p role="status">Loading link…</p>}
      {link && (
        <>
          {created && <p role="status">Link created.</p>}
          <dl>
            <dt>Short URL</dt>
            <dd>{link.shortUrl}</dd>
            <dt>Destination</dt>
            <dd>
              <span>{link.destination}</span>
              {destinationSyntaxOK(link.destination) && (
                <p>
                  <a href={link.destination} target="_blank" rel="noreferrer noopener">
                    Open destination (opens in a new tab)
                  </a>
                </p>
              )}
            </dd>
            <dt>Title</dt>
            <dd>{link.title || "Untitled link"}</dd>
            <dt>Creator</dt>
            <dd>{link.creator.email}</dd>
            <dt>Created</dt>
            <dd>{timestamp(link.createdAt)}</dd>
            <dt>Updated</dt>
            <dd>{timestamp(link.updatedAt)}</dd>
            <dt>Lifecycle</dt>
            <dd>{link.lifecycle}</dd>
            <dt>Effective state</dt>
            <dd>
              {link.lifecycle === "deleted"
                ? "Deleted"
                : link.suspension
                  ? "Suspended"
                  : link.lifecycle}
            </dd>
            <dt>Version</dt>
            <dd>{link.version}</dd>
          </dl>
          {link.suspension && (
            <p>
              Suspended by {link.suspension.actorId}: {link.suspension.reason} (
              {timestamp(link.suspension.at)})
            </p>
          )}
          {!canWrite && (
            <p>You have view-only access. Ask an owner or admin to change your role.</p>
          )}
          {canWrite && (
            <div aria-label="Lifecycle actions">
              {(link.lifecycle === "active"
                ? ["disable", "archive", "delete"]
                : link.lifecycle === "disabled"
                  ? ["enable", "archive", "delete"]
                  : link.lifecycle === "archived"
                    ? ["restore", "delete"]
                    : ["restore"]
              ).map((action) => (
                <button
                  key={action}
                  type="button"
                  disabled={busy || requiresReload || (action === "enable" && !!link.suspension)}
                  onClick={() => {
                    setNotice("");
                    if (action === "archive" || action === "delete") setConfirm(action);
                    else void transition(action as "enable" | "disable" | "restore");
                  }}
                >
                  {action[0]?.toUpperCase()}
                  {action.slice(1)} link
                </button>
              ))}
              {link.lifecycle === "disabled" && link.suspension && (
                <p>An owner or admin must clear suspension first.</p>
              )}
              {!requiresReload && (
                <button
                  ref={reload}
                  type="button"
                  disabled={busy}
                  onClick={() => {
                    setNotice("");
                    setAttempt((value) => value + 1);
                  }}
                >
                  Reload link
                </button>
              )}
            </div>
          )}
          <button type="button" disabled={busy} onClick={() => void copy()}>
            Copy short URL
          </button>
          {copied && <p role="status">{copied}</p>}
        </>
      )}
      <a href={`/workspaces/${workspaceId}/links`}>Back to links</a>
      {confirm && link && (
        <ConfirmDialog
          link={{ name: link.title || link.shortUrl, action: confirm }}
          onStay={() => setConfirm(undefined)}
          onDiscard={() => void transition(confirm)}
        />
      )}
    </section>
  );
}
