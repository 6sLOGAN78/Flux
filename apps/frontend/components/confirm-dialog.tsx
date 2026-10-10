"use client";

import { useEffect, useRef } from "react";

export function ConfirmDialog({
  onStay,
  onDiscard,
  roleChange,
}: {
  onStay: () => void;
  onDiscard: () => void;
  roleChange?: { email: string; removesManagement: boolean };
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
      <h2 id="discard-title">{roleChange ? "Change role?" : "Discard unsaved changes?"}</h2>
      {roleChange && <p>{roleChange.email}</p>}
      <p id="discard-description">
        {roleChange
          ? roleChange.removesManagement
            ? "This change removes the member's current management permissions."
            : "This change updates the member's role in this workspace."
          : "Your unsaved changes will be lost when you switch workspaces."}
      </p>
      <button ref={stay} type="button" onClick={onStay}>
        {roleChange ? "Keep current role" : "Stay"}
      </button>
      <button type="button" onClick={onDiscard}>
        {roleChange ? "Change role" : "Discard"}
      </button>
    </dialog>
  );
}
