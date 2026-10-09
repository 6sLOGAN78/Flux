"use client";

import { useAuth } from "@clerk/nextjs";
import { ZLinksResponse, type LinksResponse } from "@flux/zod";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { ApiError, createAPI } from "../lib/api";
import { destinationSyntaxOK, localLinkTime } from "../lib/links";
import { WorkspaceRequests } from "../lib/workspace";

export function LinkList({
  workspaceId,
  canCreate,
  accessLost,
}: {
  workspaceId: string;
  canCreate: boolean;
  accessLost: (status: number) => void;
}) {
  const { getToken } = useAuth();
  const requests = useRef(new WorkspaceRequests());
  const denied = useRef(accessLost);
  denied.current = accessLost;
  const [items, setItems] = useState<LinksResponse["items"]>();
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    const request = requests.current.begin();
    setLoading(true);
    setError("");
    void createAPI(getToken, { origin: window.location.origin })
      .request(`/workspaces/${workspaceId}/links`, { signal: request.signal })
      .then((value) => {
        const response = ZLinksResponse.parse(value);
        if (!request.current()) return;
        if (response.items.some((item) => item.workspaceId !== workspaceId))
          throw new Error("Unexpected resource");
        setItems(response.items);
      })
      .catch((failure: unknown) => {
        if (!request.current()) return;
        if (failure instanceof ApiError && [401, 403, 404].includes(failure.status)) {
          setItems(undefined);
          denied.current(failure.status);
        } else setError("We couldn't load links. Try again.");
      })
      .finally(() => {
        if (request.current()) setLoading(false);
      });
    return () => requests.current.clear();
  }, [workspaceId, getToken, attempt]);
  const title = (item: LinksResponse["items"][number]) => (
    <Link href={`/workspaces/${workspaceId}/links/${item.id}`}>
      {item.title || "Untitled link"}
    </Link>
  );
  const destination = (item: LinksResponse["items"][number]) =>
    destinationSyntaxOK(item.destination) ? (
      <a href={item.destination} target="_blank" rel="noreferrer noopener">
        {item.destination}
      </a>
    ) : (
      <span>{item.destination}</span>
    );
  const state = (item: LinksResponse["items"][number]) =>
    item.suspension
      ? "Suspended"
      : { active: "Active", disabled: "Disabled", archived: "Archived", deleted: "Deleted" }[
          item.lifecycle
        ];
  const time = (item: LinksResponse["items"][number]) => (
    <time dateTime={item.createdAt}>{localLinkTime(item.createdAt)}</time>
  );
  return (
    <section aria-labelledby="links-library-title" aria-busy={loading}>
      <h2 id="links-library-title">Links</h2>
      <p>Link management is available. Redirects and analytics are not available yet.</p>
      {!canCreate && <p>You have view-only access. Ask an owner or admin to change your role.</p>}
      {loading && <p role="status">Loading links…</p>}
      {error && <p role="alert">{error}</p>}
      <button type="button" disabled={loading} onClick={() => setAttempt((value) => value + 1)}>
        {error ? "Retry loading links" : "Reload links"}
      </button>
      {items?.length === 0 && (
        <>
          <h3>No links yet</h3>
          <p>No links have been created in this workspace.</p>
        </>
      )}
      {items && items.length > 0 && (
        <>
          <p>Newest links first. Showing up to 25 links.</p>
          <div className="desktop-library">
            <table aria-label="Links library">
              <thead>
                <tr>
                  <th scope="col">Title</th>
                  <th scope="col">Short URL</th>
                  <th scope="col">Destination</th>
                  <th scope="col">Created</th>
                  <th scope="col">State</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item) => (
                  <tr key={item.id}>
                    <td>{title(item)}</td>
                    <td>{item.shortUrl}</td>
                    <td>{destination(item)}</td>
                    <td>{time(item)}</td>
                    <td>{state(item)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <ul className="mobile-library" aria-label="Links library">
            {items.map((item) => (
              <li key={item.id}>
                <dl>
                  <dt>Title</dt>
                  <dd>{title(item)}</dd>
                  <dt>Short URL</dt>
                  <dd>{item.shortUrl}</dd>
                  <dt>Destination</dt>
                  <dd>{destination(item)}</dd>
                  <dt>Created</dt>
                  <dd>{time(item)}</dd>
                  <dt>State</dt>
                  <dd>{state(item)}</dd>
                </dl>
              </li>
            ))}
          </ul>
        </>
      )}
      <style jsx>{`
        table { width: 100%; table-layout: fixed; border-collapse: collapse; }
        th, td { text-align: left; vertical-align: top; padding: 0.5rem; overflow-wrap: anywhere; }
        .mobile-library { display: none; }
        dd { margin: 0 0 1rem; overflow-wrap: anywhere; }
        @media (max-width: 640px) {
          .desktop-library { display: none; }
          .mobile-library { display: block; padding-left: 1.25rem; }
        }
      `}</style>
    </section>
  );
}
