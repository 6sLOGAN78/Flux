"use client";

import { useAuth } from "@clerk/nextjs";
import { ZMemberResponse, ZMembersResponse, type MembersResponse } from "@flux/zod";
import { useEffect, useId, useRef, useState } from "react";
import { ApiError, createAPI } from "../lib/api";
import { WorkspaceRequests } from "../lib/workspace";
import { ConfirmDialog } from "./confirm-dialog";

const roles = { owner: "Owner", admin: "Admin", member: "Member", viewer: "Viewer" };

export function TeamList({
  workspaceId,
  actorRole,
  canInspect,
  accessLost,
}: {
  workspaceId: string;
  actorRole: MembersResponse["items"][number]["role"];
  canInspect: boolean;
  accessLost: (status: number) => void;
}) {
  const { getToken } = useAuth();
  const requests = useRef(new WorkspaceRequests());
  const mutations = useRef(new WorkspaceRequests());
  const submitting = useRef(false);
  const denied = useRef(accessLost);
  denied.current = accessLost;
  const [items, setItems] = useState<MembersResponse["items"]>();
  const [next, setNext] = useState<string | null>(null);
  const [after, setAfter] = useState<string>();
  const positions = useRef<(string | undefined)[]>([undefined]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(canInspect);
  const [attempt, setAttempt] = useState(0);
  const [pending, setPending] = useState<{
    member: MembersResponse["items"][number];
    role: MembersResponse["items"][number]["role"];
  }>();
  const [roleError, setRoleError] = useState("");
  const [committed, setCommitted] = useState(false);
  const [changing, setChanging] = useState(false);
  const changeRole = async () => {
    if (!pending || submitting.current) return;
    submitting.current = true;
    const selected = pending;
    setPending(undefined);
    setChanging(true);
    setCommitted(false);
    const request = mutations.current.begin();
    try {
      const response = ZMemberResponse.parse(
        await createAPI(getToken, { origin: window.location.origin }).request(
          `/workspaces/${workspaceId}/members/${selected.member.id}`,
          {
            method: "PATCH",
            body: { role: selected.role },
            idempotencyKey: crypto.randomUUID(),
            signal: request.signal,
          },
        ),
      );
      if (!request.current()) return;
      if (
        response.member.workspaceId !== workspaceId ||
        response.member.id !== selected.member.id ||
        response.member.role !== selected.role
      )
        throw new Error("Unexpected resource");
      if (response.actorRole !== actorRole) {
        requests.current.clear();
        setItems(undefined);
        denied.current(403);
        return;
      }
      setItems((value) =>
        value?.map((item) => (item.id === response.member.id ? response.member : item)),
      );
      setCommitted(true);
      window.dispatchEvent(new Event("focus"));
    } catch (failure: unknown) {
      if (!request.current()) return;
      if (failure instanceof ApiError && [401, 403, 404].includes(failure.status)) {
        requests.current.clear();
        setItems(undefined);
        denied.current(failure.status);
      } else
        setRoleError(
          failure instanceof ApiError && failure.constraint === "OWNER_REQUIRED"
            ? "Promote another owner first"
            : "We couldn't change the role. Reload team before trying again.",
        );
    } finally {
      submitting.current = false;
      if (request.current()) setChanging(false);
    }
  };
  useEffect(() => {
    if (!canInspect) {
      requests.current.clear();
      setItems(undefined);
      return;
    }
    const request = requests.current.begin();
    setPending(undefined);
    setRoleError("");
    setCommitted(false);
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
    return () => {
      requests.current.clear();
      mutations.current.clear();
    };
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
      {roleError && <p role="alert">{roleError}</p>}
      {committed && <p role="status">Role changed.</p>}
      {pending && (
        <ConfirmDialog
          roleChange={{
            email: pending.member.email,
            removesManagement:
              (pending.member.role === "owner" || pending.member.role === "admin") &&
              pending.role !== pending.member.role &&
              pending.role !== "owner",
          }}
          onStay={() => setPending(undefined)}
          onDiscard={() => void changeRole()}
        />
      )}
      <button
        type="button"
        disabled={loading || changing}
        onClick={() => setAttempt((value) => value + 1)}
      >
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
                  <th scope="col">Actions</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item) => (
                  <tr key={item.id}>
                    <td>{item.email}</td>
                    <td>{roles[item.role]}</td>
                    <td>
                      <RoleControl
                        key={`${item.id}:${item.role}`}
                        item={item}
                        actorRole={actorRole}
                        disabled={loading || changing || !!roleError || !!error}
                        onChange={(role) => setPending({ member: item, role })}
                      />
                    </td>
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
                <RoleControl
                  key={`${item.id}:${item.role}`}
                  item={item}
                  actorRole={actorRole}
                  disabled={loading || changing || !!roleError || !!error}
                  onChange={(role) => setPending({ member: item, role })}
                />
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

function RoleControl({
  item,
  actorRole,
  disabled,
  onChange,
}: {
  item: MembersResponse["items"][number];
  actorRole: MembersResponse["items"][number]["role"];
  disabled: boolean;
  onChange: (role: MembersResponse["items"][number]["role"]) => void;
}) {
  const id = useId();
  const [role, setRole] = useState(item.role);
  const options =
    actorRole === "owner"
      ? (["owner", "admin", "member", "viewer"] as const)
      : (["member", "viewer"] as const);
  if (actorRole !== "owner" && (item.role === "owner" || item.role === "admin"))
    return <span>Owner manages this role.</span>;
  return (
    <>
      <label htmlFor={id}>New role</label>
      <select
        id={id}
        value={role}
        disabled={disabled}
        onChange={(event) => {
          const value = event.target.value;
          if (value === "owner" || value === "admin" || value === "member" || value === "viewer")
            setRole(value);
        }}
      >
        {options.map((value) => (
          <option key={value} value={value}>
            {roles[value]}
          </option>
        ))}
      </select>
      <button
        type="button"
        disabled={disabled || role === item.role}
        onClick={() => onChange(role)}
      >
        Change role
      </button>
      {item.role === "owner" && <p>Promote another owner first</p>}
    </>
  );
}
