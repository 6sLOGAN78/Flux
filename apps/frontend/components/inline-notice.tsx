"use client";

import { useEffect, useRef } from "react";

export function InlineNotice({
  children,
  blocking = false,
}: {
  children: string;
  blocking?: boolean;
}) {
  const notice = useRef<HTMLParagraphElement>(null);
  useEffect(() => {
    if (blocking) notice.current?.focus();
  }, [children, blocking]);
  return (
    <p
      ref={notice}
      tabIndex={blocking ? -1 : undefined}
      role={blocking ? "alert" : "status"}
      aria-atomic="true"
    >
      {children}
    </p>
  );
}
