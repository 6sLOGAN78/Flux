import React from 'react';
import { cn } from '@/lib/utils';

export interface InputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'prefix'> {
  label?: string;
  description?: string;
  error?: string;
  prefix?: React.ReactNode;
  suffix?: React.ReactNode;
  startIcon?: React.ReactNode;
  endIcon?: React.ReactNode;
  wrapperClassName?: string;
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  (
    {
      className,
      wrapperClassName,
      label,
      description,
      error,
      prefix,
      suffix,
      startIcon,
      endIcon,
      id,
      disabled,
      ...props
    },
    ref
  ) => {
    const generatedId = React.useId();
    const inputId = id || generatedId;

    return (
      <div className={cn('w-full space-y-1.5', wrapperClassName)}>
        {label && (
          <label
            htmlFor={inputId}
            className="block text-xs font-medium text-zinc-700 dark:text-zinc-300"
          >
            {label}
          </label>
        )}

        <div
          className={cn(
            'group flex h-9 w-full items-center rounded-lg border bg-white px-3 text-xs transition-colors dark:bg-zinc-950',
            error
              ? 'border-red-500 focus-within:border-red-500 focus-within:ring-2 focus-within:ring-red-500/10 dark:border-red-500 dark:focus-within:border-red-500'
              : 'border-zinc-200 focus-within:border-zinc-400 focus-within:ring-2 focus-within:ring-zinc-900/10 dark:border-zinc-800 dark:focus-within:border-zinc-600 dark:focus-within:ring-zinc-100/10',
            disabled && 'cursor-not-allowed opacity-50 bg-zinc-50 dark:bg-zinc-900/50'
          )}
        >
          {startIcon && (
            <div className="mr-2 flex items-center text-zinc-400 dark:text-zinc-500">
              {startIcon}
            </div>
          )}

          {prefix && (
            <span className="mr-1 select-none font-mono text-xs text-zinc-400 dark:text-zinc-500">
              {prefix}
            </span>
          )}

          <input
            id={inputId}
            ref={ref}
            disabled={disabled}
            className={cn(
              'w-full flex-1 bg-transparent py-1 text-xs text-zinc-900 placeholder:text-zinc-400 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50 dark:text-zinc-100 dark:placeholder:text-zinc-600',
              className
            )}
            {...props}
          />

          {suffix && (
            <span className="ml-1 select-none font-mono text-xs text-zinc-400 dark:text-zinc-500">
              {suffix}
            </span>
          )}

          {endIcon && (
            <div className="ml-2 flex items-center text-zinc-400 dark:text-zinc-500">
              {endIcon}
            </div>
          )}
        </div>

        {description && !error && (
          <p className="text-[11px] text-zinc-500 dark:text-zinc-400">
            {description}
          </p>
        )}

        {error && (
          <p className="text-[11px] font-medium text-red-500 dark:text-red-400">
            {error}
          </p>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';

export default Input;
