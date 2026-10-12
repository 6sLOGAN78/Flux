import type { InputHTMLAttributes } from "react";

export function Field({
  label,
  hint,
  error,
  id,
  ...input
}: InputHTMLAttributes<HTMLInputElement> & {
  id: string;
  label: string;
  hint?: string;
  error?: string;
}) {
  const description = [hint && `${id}-help`, error && `${id}-error`].filter(Boolean).join(" ");
  return (
    <div className="field">
      <label htmlFor={id}>{label}</label>
      <input
        {...input}
        id={id}
        aria-invalid={error ? true : undefined}
        aria-describedby={description || undefined}
      />
      {hint && (
        <p id={`${id}-help`} className="field-help">
          {hint}
        </p>
      )}
      {error && (
        <p id={`${id}-error`} className="field-error">
          {error}
        </p>
      )}
    </div>
  );
}
