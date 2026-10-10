"use client";

import { useAuth } from "@clerk/nextjs";
import type { InvitationsResponse } from "@flux/zod";
import { useEffect, useRef, useState, type FormEvent } from "react";
import { ApiError, createAPI } from "../lib/api";
import { WorkspaceRequests } from "../lib/workspace";
import { invitationPage, queuedInvitation, invitationStatus, invitationTime } from "../lib/team";

export function InvitationPanel({
  workspaceId,
  actorRole,
  accessLost,
  onDirty,
}: {
  workspaceId: string;
  actorRole: string;
  accessLost: (status: number) => void;
  onDirty: (dirty: boolean) => void;
}) {
  const { getToken } = useAuth();
  const reads = useRef(new WorkspaceRequests());
  const writes = useRef(new WorkspaceRequests());
  const denied = useRef(accessLost);
  denied.current = accessLost;
  const submitting = useRef(false);
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("member");
  const [items, setItems] = useState<InvitationsResponse["items"]>([]);
  const [after, setAfter] = useState<string>();
  const [next, setNext] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [queued, setQueued] = useState(false);
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    onDirty(email !== "" || role !== "member");
  }, [email, role, onDirty]);
  useEffect(
    () => () => {
      reads.current.clear();
      writes.current.clear();
      onDirty(false);
    },
    [onDirty],
  );
  useEffect(() => {
    const request = reads.current.begin();
    setLoading(true);
    setError("");
    void createAPI(getToken, { origin: window.location.origin })
      .request(`/workspaces/${workspaceId}/invitations`, {
        signal: request.signal,
        query: after ? new URLSearchParams({ after }) : undefined,
      })
      .then((value) => {
        const response = invitationPage(value, workspaceId);
        if (!request.current()) return;
        setItems(response.items);
        setNext(response.nextAfter);
      })
      .catch((failure: unknown) => {
        if (!request.current()) return;
        if (failure instanceof ApiError && [401, 403, 404].includes(failure.status)) {
          setItems([]);
          denied.current(failure.status);
        } else setError("We couldn't load invitations. Try again.");
      })
      .finally(() => {
        if (request.current()) setLoading(false);
      });
    return () => reads.current.clear();
  }, [getToken, workspaceId, after, attempt]);
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (submitting.current) return;
    submitting.current = true;
    setBusy(true);
    setQueued(false);
    setError("");
    const request = writes.current.begin();
    try {
      const value = await createAPI(getToken, { origin: window.location.origin }).request(
        `/workspaces/${workspaceId}/invitations`,
        {
          method: "POST",
          body: { email, role },
          signal: request.signal,
          idempotencyKey: crypto.randomUUID(),
        },
      );
      const invitation = queuedInvitation(value, workspaceId);
      if (!request.current()) return;
      setItems([invitation]);
      setNext(null);
      setAfter(undefined);
      setEmail("");
      setRole("member");
      setQueued(true);
      setAttempt((value) => value + 1);
    } catch (failure: unknown) {
      if (!request.current()) return;
      if (failure instanceof ApiError && [401, 403, 404].includes(failure.status)) {
        setItems([]);
        denied.current(failure.status);
      } else
        setError(
          failure instanceof ApiError && failure.constraint === "INVITATION_PENDING"
            ? "An invitation is already pending for this email. Resend or revoke it."
            : "We couldn't queue the invitation. Reload invitations before trying again.",
        );
    } finally {
      submitting.current = false;
      if (request.current()) setBusy(false);
    }
  };
  return (
    <section aria-label="Invitations">
      <h2>Invitations</h2>
      <form aria-label="Invite member" onSubmit={submit}>
        <label htmlFor="invite-email">Email address</label>
        <input
          id="invite-email"
          type="email"
          maxLength={254}
          required
          value={email}
          disabled={busy}
          onChange={(event) => setEmail(event.target.value)}
        />
        <label htmlFor="invite-role">Role</label>
        <select
          id="invite-role"
          value={role}
          disabled={busy}
          onChange={(event) => setRole(event.target.value)}
        >
          {actorRole === "owner" && <option value="admin">Admin</option>}
          <option value="member">Member</option>
          <option value="viewer">Viewer</option>
        </select>
        <button type="submit" disabled={busy}>
          {busy ? "Queuing invitation…" : "Invite member"}
        </button>
      </form>
      {queued && <p role="status">Invitation queued</p>}
      {error && <p role="alert">{error}</p>}
      {loading ? (
        <p>Loading invitations…</p>
      ) : items.length === 0 ? (
        <p>No pending invitations. Invite a member to collaborate in this workspace.</p>
      ) : (
        <ul aria-label="Workspace invitations">
          {items.map((item) => (
            <li key={item.id}>
              {item.email} · {item.role} · Delivery status: {invitationStatus(item.status)} ·
              Expires <time dateTime={item.expiresAt}>{invitationTime(item.expiresAt)}</time>
              {item.status === "Failed" && (
                <p>Invitation delivery failed. Resend the invitation to try again.</p>
              )}
            </li>
          ))}
        </ul>
      )}
      <button
        type="button"
        disabled={busy || loading}
        onClick={() => {
          setAfter(undefined);
          setAttempt((value) => value + 1);
        }}
      >
        Reload invitations
      </button>
      {after && (
        <button type="button" disabled={loading} onClick={() => setAfter(undefined)}>
          First page
        </button>
      )}
      {next && (
        <button type="button" disabled={loading} onClick={() => setAfter(next)}>
          Next invitations
        </button>
      )}
    </section>
  );
}
