"use client";

import { useAuth } from "@clerk/nextjs";
import { ZMembersResponse, type MembersResponse } from "@flux/zod";
import { useEffect, useRef, useState } from "react";
import { ApiError, createAPI } from "../lib/api";
import { WorkspaceRequests } from "../lib/workspace";

const roles = { owner: "Owner", admin: "Admin", member: "Member", viewer: "Viewer" };

export function TeamList({
  workspaceId,
  canInspect,
  accessLost,
}: {
  workspaceId: string;
  canInspect: boolean;
  accessLost: (status: number) => void;
}) {
  const { getToken } = useAuth();
  const requests = useRef(new WorkspaceRequests());
  const denied = useRef(accessLost);
  denied.current = accessLost;
  const [items, setItems] = useState<MembersResponse["items"]>();
  const [next, setNext] = useState<string | null>(null);
  const [after, setAfter] = useState<string>();
  const positions = useRef<(string | undefined)[]>([undefined]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(canInspect);
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    if (!canInspect) {
      requests.current.clear();
      setItems(undefined);
      return;
    }
    const request = requests.current.begin();
    setLoading(true);
    setError("");
    void createAPI(getToken, { origin: window.location.origin })
      .request(`/workspaces/${workspaceId}/members`, {
        signal: request.signal,
        query: after ? new URLSearchParams({ after }) : undefined,
      })
      .then((value) => {
        const response = ZMembersResponse.parse(value);
        if (!request.current()) return;
        if (response.items.some((item) => item.workspaceId !== workspaceId))
          throw new Error("Unexpected resource");
        if (!positions.current.includes(after)) positions.current.push(after);
        setItems(response.items);
        setNext(response.nextAfter);
      })
      .catch((failure: unknown) => {
        if (!request.current()) return;
        if (failure instanceof ApiError && [401, 403, 404].includes(failure.status)) {
          requests.current.clear();
          setItems(undefined);
          denied.current(failure.status);
        } else setError("We couldn't load team. Try again.");
      })
      .finally(() => {
        if (request.current()) setLoading(false);
      });
    return () => requests.current.clear();
  }, [workspaceId, canInspect, after, getToken, attempt]);
  if (!canInspect)
    return (
      <section aria-labelledby="team-title">
        <h2 id="team-title">Team</h2>
        <p>You do not have permission to view Team.</p>
      </section>
    );
  const previousIndex = positions.current.indexOf(after) - 1;
  return (
    <section aria-labelledby="team-title" aria-busy={loading}>
      <h2 id="team-title">Team</h2>
      <p>Inspect current workspace members and their roles.</p>
      {loading && <p role="status">Loading team…</p>}
      {error && <p role="alert">{error}</p>}
      <button type="button" disabled={loading} onClick={() => setAttempt((value) => value + 1)}>
        {error ? "Retry loading team" : "Reload team"}
      </button>
      {!loading && !error && items?.length === 0 && <p>No members on this page.</p>}
      {items && items.length > 0 && (
        <>
          <div className="desktop-team">
            <table aria-label="Team members">
              <thead>
                <tr>
                  <th scope="col">Identity</th>
                  <th scope="col">Role</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item) => (
                  <tr key={item.id}>
                    <td>{item.email}</td>
                    <td>{roles[item.role]}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <ul className="mobile-team" aria-label="Team members">
            {items.map((item) => (
              <li key={item.id}>
                <dl>
                  <dt>Identity</dt>
                  <dd>{item.email}</dd>
                  <dt>Role</dt>
                  <dd>{roles[item.role]}</dd>
                </dl>
              </li>
            ))}
          </ul>
        </>
      )}
      {items && (
        <nav aria-label="Team pages">
          <button
            type="button"
            disabled={loading || !!error || previousIndex < 0}
            onClick={() => {
              requests.current.clear();
              setAfter(positions.current[previousIndex]);
            }}
          >
            Previous page
          </button>
          <button
            type="button"
            disabled={loading || !!error || next === null}
            onClick={() => {
              if (next !== null) {
                requests.current.clear();
                setAfter(next);
              }
            }}
          >
            Next page
          </button>
          {!loading && !error && next === null && <p role="status">No more members.</p>}
        </nav>
      )}
      <style jsx>{`
        table { width: 100%; table-layout: fixed; border-collapse: collapse; }
        tr { height: 64px; }
        thead { background: var(--shell); }
        th, td { text-align: left; vertical-align: top; padding: var(--space-sm); overflow-wrap: anywhere; }
        .mobile-team { display: none; }
        dd { margin: 0 0 1rem; overflow-wrap: anywhere; }
        @media (max-width: 767px) { .desktop-team { display: none; } .mobile-team { display: block; padding-left: var(--space-lg); } }
      `}</style>
    </section>
  );
}
