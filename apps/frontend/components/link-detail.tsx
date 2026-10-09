"use client";

import { useAuth } from "@clerk/nextjs";
import { ZLinkResponse, type LinkResponse } from "@flux/zod";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { ApiError, createAPI } from "../lib/api";
import { destinationSyntaxOK, localLinkTime } from "../lib/links";
import { WorkspaceRequests, resolveWorkspace } from "../lib/workspace";

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
  useEffect(() => {
    setCreated(new URLSearchParams(window.location.search).get("created") === "1");
    const request = requests.current.begin();
    const api = createAPI(getToken, { origin: window.location.origin });
    void api
      .request(`/workspaces/${workspaceId}/links/${linkId}`, { signal: request.signal })
      .then((value) => {
        const response = ZLinkResponse.parse(value);
        if (!request.current()) return;
        if (response.link.workspaceId !== workspaceId || response.link.id !== linkId)
          throw new Error("Unexpected resource");
        setLink(response.link);
        setError("");
      })
      .catch(async (failure: unknown) => {
        if (!request.current()) return;
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
    return () => requests.current.clear();
  }, [workspaceId, linkId, getToken, attempt, router]);
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
    <section aria-labelledby="link-detail-title">
      <h2 id="link-detail-title">Link details</h2>
      <p>Link management is available. Redirects and analytics are not available yet.</p>
      {error && <p role="alert">{error}</p>}
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
          <button type="button" onClick={() => void copy()}>
            Copy short URL
          </button>
          {copied && <p role="status">{copied}</p>}
        </>
      )}
      <a href={`/workspaces/${workspaceId}/links`}>Back to links</a>
    </section>
  );
}
