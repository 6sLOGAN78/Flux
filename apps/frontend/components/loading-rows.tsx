"use client";

import { useEffect, useState } from "react";

export function LoadingRows({
  label,
  onCancel,
  onRetry,
  retryLabel,
}: {
  label: string;
  onCancel?: () => void;
  onRetry?: () => void;
  retryLabel?: string;
}) {
  const [delayed, setDelayed] = useState(false);
  useEffect(() => {
    const timer = window.setTimeout(() => setDelayed(true), 5000);
    return () => window.clearTimeout(timer);
  }, []);
  return (
    <div>
      <p role="status">{label}</p>
      <div className="loading-rows" inert aria-hidden="true">
        <div />
        <div />
        <div />
      </div>
      {delayed && (
        <>
          <p role="status">This is taking longer than expected. You can retry.</p>
          {onCancel && (
            <button type="button" className="button-secondary" onClick={onCancel}>
              Cancel loading
            </button>
          )}
          {onRetry && (
            <button type="button" className="button-secondary" onClick={onRetry}>
              {retryLabel}
            </button>
          )}
        </>
      )}
    </div>
  );
}
