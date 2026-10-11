"use client";

import { useEffect, useRef } from "react";

export function ConfirmDialog({
  onStay,
  onDiscard,
  roleChange,
  invitation,
  link,
}: {
  onStay: () => void;
  onDiscard: () => void;
  roleChange?: { email: string; removesManagement: boolean };
  invitation?: { email: string; resend: boolean; uncertain: boolean };
  link?: { name: string; action: "archive" | "delete" };
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  const stay = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    const previous = document.activeElement;
    const element = dialog.current;
    element?.showModal();
    stay.current?.focus();
    return () => {
      element?.close();
      if (previous instanceof HTMLElement && previous.isConnected) previous.focus();
    };
  }, []);
  return (
    <dialog
      ref={dialog}
      aria-labelledby="discard-title"
      aria-describedby="discard-description"
      onCancel={(event) => {
        event.preventDefault();
        onStay();
      }}
    >
      <h2 id="discard-title">
        {link
          ? link.action === "delete"
            ? "Delete link?"
            : "Archive link?"
          : invitation
            ? invitation.resend
              ? "Resend invitation?"
              : "Revoke invitation?"
            : roleChange
              ? "Change role?"
              : "Discard unsaved changes?"}
      </h2>
      {invitation && <p>{invitation.email}</p>}
      {link && <p>{link.name}</p>}
      {roleChange && <p>{roleChange.email}</p>}
      <p id="discard-description">
        {link
          ? link.action === "delete"
            ? "This link moves to Deleted. Its short key stays reserved. You can restore it as disabled."
            : "Restoring this link will leave it disabled."
          : invitation
            ? invitation.resend
              ? "A new invitation will replace the previous one."
              : "The recipient will no longer be able to accept this invitation."
            : roleChange
              ? roleChange.removesManagement
                ? "This change removes the member's current management permissions."
                : "This change updates the member's role in this workspace."
              : "Your unsaved changes will be lost when you switch workspaces."}
      </p>
      {invitation?.resend && invitation.uncertain && (
        <p>
          The previous message may have arrived. Resending invalidates its invitation and queues a
          new message.
        </p>
      )}
      <button ref={stay} type="button" onClick={onStay}>
        {link
          ? "Keep link"
          : invitation
            ? "Keep invitation"
            : roleChange
              ? "Keep current role"
              : "Stay"}
      </button>
      <button type="button" onClick={onDiscard}>
        {link
          ? link.action === "delete"
            ? "Delete link"
            : "Archive link"
          : invitation
            ? invitation.resend
              ? "Resend invitation"
              : "Revoke invitation"
            : roleChange
              ? "Change role"
              : "Discard"}
      </button>
    </dialog>
  );
}
