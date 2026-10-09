"use client";

import { useAuth } from "@clerk/nextjs";
import { ZLinksResponse, type LinksResponse } from "@flux/zod";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { ApiError, createAPI } from "../lib/api";
import {
  destinationSyntaxOK,
  invalidCursor,
  LinkPages,
  localLinkTime,
  normalizeLinkSearch,
} from "../lib/links";
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
  const [cursor, setCursor] = useState<string>();
  const [draftSearch, setDraftSearch] = useState("");
  const [search, setSearch] = useState("");
  const [lifecycle, setLifecycle] = useState("nondeleted");
  const pages = useRef(new LinkPages());
  const scope = JSON.stringify({ workspaceId, state: lifecycle, search });
  pages.current.bind(scope);
  useEffect(() => {
    const request = requests.current.begin();
    setLoading(true);
    setError("");
    void createAPI(getToken, { origin: window.location.origin })
      .request(`/workspaces/${workspaceId}/links`, {
        signal: request.signal,
        query: new URLSearchParams({
          search,
          state: lifecycle,
          ...(cursor === undefined ? {} : { cursor }),
        }),
      })
      .then((value) => {
        const response = ZLinksResponse.parse(value);
        if (!request.current()) return;
        if (response.items.some((item) => item.workspaceId !== workspaceId))
          throw new Error("Unexpected resource");
        pages.current.accept(scope, cursor, response.nextCursor);
        setItems(response.items);
      })
      .catch((failure: unknown) => {
        if (!request.current()) return;
        if (failure instanceof ApiError && [401, 403, 404].includes(failure.status)) {
          setItems(undefined);
          denied.current(failure.status);
        } else if (failure instanceof ApiError && failure.constraint === "CURSOR_INVALID") {
          setError(invalidCursor);
        } else setError("We couldn't load links. Try again.");
      })
      .finally(() => {
        if (request.current()) setLoading(false);
      });
    return () => requests.current.clear();
  }, [workspaceId, scope, cursor, getToken, attempt, search, lifecycle]);
  const commitFilters = (nextSearch: string, nextLifecycle: string) => {
    // Invalidate immediately; even a completion before the next effect is stale.
    requests.current.clear();
    setLoading(true);
    setError("");
    setSearch(normalizeLinkSearch(nextSearch));
    setLifecycle(nextLifecycle);
    setCursor(undefined);
    setAttempt((value) => value + 1);
  };
  const clearFilters = () => {
    setDraftSearch("");
    commitFilters("", "nondeleted");
  };
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
      <form
        onSubmit={(event) => {
          event.preventDefault();
          commitFilters(draftSearch, lifecycle);
        }}
      >
        <label htmlFor="link-search">Search links</label>
        <input
          id="link-search"
          type="search"
          value={draftSearch}
          aria-describedby="link-search-help"
          onChange={(event) => setDraftSearch(event.target.value)}
        />
        <p id="link-search-help">Search short keys, titles, or destinations</p>
        <button type="submit">Search links</button>
        <button
          type="button"
          onClick={() => {
            setDraftSearch("");
            commitFilters("", lifecycle);
          }}
        >
          Clear search
        </button>
        <label htmlFor="link-lifecycle">Lifecycle</label>
        <select
          id="link-lifecycle"
          value={lifecycle}
          onChange={(event) => {
            commitFilters(search, event.target.value);
          }}
        >
          <option value="nondeleted">All nondeleted</option>
          <option value="active">Active</option>
          <option value="disabled">Disabled</option>
          <option value="archived">Archived</option>
          <option value="deleted">Deleted</option>
        </select>
      </form>
      {loading && <p role="status">Loading links…</p>}
      {error && <p role="alert">{error}</p>}
      {error === invalidCursor && (
        <button
          type="button"
          disabled={loading}
          onClick={() => {
            setCursor(undefined);
            setAttempt((value) => value + 1);
          }}
        >
          Return to first page
        </button>
      )}
      <button type="button" disabled={loading} onClick={() => setAttempt((value) => value + 1)}>
        {error ? "Retry loading links" : "Reload links"}
      </button>
      {!loading &&
        !error &&
        items?.length === 0 &&
        !pages.current.canPrevious &&
        (search !== "" || lifecycle !== "nondeleted") && (
          <>
            <h3>
              {search === "" && lifecycle === "deleted" ? "No deleted links" : "No matching links"}
            </h3>
            <p>
              {search === "" && lifecycle === "deleted"
                ? "Deleted links appear here and can be restored."
                : "Try a different search or clear your filters."}
            </p>
            <button type="button" onClick={clearFilters}>
              Clear filters
            </button>
          </>
        )}
      {!loading &&
        !error &&
        items?.length === 0 &&
        !pages.current.canPrevious &&
        search === "" &&
        lifecycle === "nondeleted" && (
          <>
            <h3>No links yet</h3>
            <p>Create a managed-domain link to organize its destination and status.</p>
            {canCreate && (
              <Link href={`/workspaces/${workspaceId}/links/new`}>Create your first link</Link>
            )}
          </>
        )}
      {items && items.length > 0 && (
        <>
          <p>Newest links first. Showing up to 25 links.</p>
          {canCreate && <Link href={`/workspaces/${workspaceId}/links/new`}>Create link</Link>}
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
      {items && (
        <nav aria-label="Link pages">
          <button
            type="button"
            disabled={loading || !!error || !pages.current.canPrevious}
            onClick={() => setCursor(pages.current.previous)}
          >
            Previous page
          </button>
          <button
            type="button"
            disabled={loading || !!error || pages.current.next === null}
            onClick={() => {
              if (pages.current.next !== null) setCursor(pages.current.next);
            }}
          >
            Next page
          </button>
          {!loading && !error && pages.current.next === null && <p role="status">No more links.</p>}
        </nav>
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
