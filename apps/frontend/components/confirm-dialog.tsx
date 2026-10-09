"use client";

import { useEffect, useRef } from "react";

export function ConfirmDialog({
  onStay,
  onDiscard,
}: {
  onStay: () => void;
  onDiscard: () => void;
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
      <h2 id="discard-title">Discard unsaved changes?</h2>
      <p id="discard-description">Your unsaved changes will be lost when you switch workspaces.</p>
      <button ref={stay} type="button" onClick={onStay}>
        Stay
      </button>
      <button type="button" onClick={onDiscard}>
        Discard
      </button>
    </dialog>
  );
}
