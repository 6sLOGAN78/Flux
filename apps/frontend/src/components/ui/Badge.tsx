import React from 'react';
import { cn } from '@/lib/utils';

export type BadgeVariant =
  | 'zinc'
  | 'default'
  | 'emerald'
  | 'blue'
  | 'amber'
  | 'rose'
  | 'purple'
  | 'outline';

export type BadgeSize = 'sm' | 'md' | 'lg';

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant;
  size?: BadgeSize;
  dot?: boolean;
}

const variantStyles: Record<BadgeVariant, string> = {
  zinc: 'bg-zinc-100 text-zinc-800 border-zinc-200 dark:bg-zinc-800/60 dark:text-zinc-300 dark:border-zinc-700/50',
  default:
    'bg-zinc-100 text-zinc-800 border-zinc-200 dark:bg-zinc-800/60 dark:text-zinc-300 dark:border-zinc-700/50',
  emerald:
    'bg-emerald-50 text-emerald-700 border-emerald-200/60 dark:bg-emerald-950/50 dark:text-emerald-300 dark:border-emerald-800/50',
  blue: 'bg-blue-50 text-blue-700 border-blue-200/60 dark:bg-blue-950/50 dark:text-blue-300 dark:border-blue-800/50',
  amber:
    'bg-amber-50 text-amber-700 border-amber-200/60 dark:bg-amber-950/50 dark:text-amber-300 dark:border-amber-800/50',
  rose: 'bg-rose-50 text-rose-700 border-rose-200/60 dark:bg-rose-950/50 dark:text-rose-300 dark:border-rose-800/50',
  purple:
    'bg-purple-50 text-purple-700 border-purple-200/60 dark:bg-purple-950/50 dark:text-purple-300 dark:border-purple-800/50',
  outline:
    'bg-transparent text-zinc-700 border-zinc-200 dark:text-zinc-300 dark:border-zinc-800',
};

const dotColors: Record<BadgeVariant, string> = {
  zinc: 'bg-zinc-400 dark:bg-zinc-500',
  default: 'bg-zinc-400 dark:bg-zinc-500',
  emerald: 'bg-emerald-500',
  blue: 'bg-blue-500',
  amber: 'bg-amber-500',
  rose: 'bg-rose-500',
  purple: 'bg-purple-500',
  outline: 'bg-zinc-400',
};

const sizeStyles: Record<BadgeSize, string> = {
  sm: 'text-[10px] px-1.5 py-0.25 gap-1',
  md: 'text-xs px-2 py-0.5 gap-1.5',
  lg: 'text-xs px-2.5 py-1 gap-1.5 font-medium',
};

export function Badge({
  className,
  variant = 'default',
  size = 'md',
  dot = false,
  children,
  ...props
}: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full border font-medium select-none',
        variantStyles[variant],
        sizeStyles[size],
        className
      )}
      {...props}
    >
      {dot && (
        <span
          className={cn(
            'h-1.5 w-1.5 rounded-full',
            dotColors[variant]
          )}
        />
      )}
      {children}
    </span>
  );
}

export default Badge;
