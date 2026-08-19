import React from 'react';
import { cn } from '@/lib/utils';

export interface TabItem {
  id: string;
  label: string;
  count?: number | string;
  icon?: React.ReactNode;
  disabled?: boolean;
}

export type TabsVariant = 'underline' | 'pills' | 'segments';

export interface TabsProps {
  tabs: TabItem[];
  activeTab: string;
  onChange: (id: string) => void;
  variant?: TabsVariant;
  className?: string;
  fullWidth?: boolean;
}

export function Tabs({
  tabs,
  activeTab,
  onChange,
  variant = 'underline',
  className,
  fullWidth = false,
}: TabsProps) {
  if (variant === 'pills') {
    return (
      <div
        role="tablist"
        className={cn(
          'inline-flex items-center gap-1 rounded-lg bg-zinc-100/80 p-1 dark:bg-zinc-900/80',
          fullWidth && 'w-full flex',
          className
        )}
      >
        {tabs.map((tab) => {
          const isActive = tab.id === activeTab;
          return (
            <button
              key={tab.id}
              role="tab"
              type="button"
              aria-selected={isActive}
              disabled={tab.disabled}
              onClick={() => !tab.disabled && onChange(tab.id)}
              className={cn(
                'inline-flex items-center justify-center gap-1.5 rounded-md px-3 py-1 text-xs font-medium transition-all duration-150',
                'disabled:cursor-not-allowed disabled:opacity-50',
                isActive
                  ? 'bg-white text-zinc-900 shadow-xs dark:bg-zinc-800 dark:text-zinc-100 font-semibold'
                  : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200',
                fullWidth && 'flex-1'
              )}
            >
              {tab.icon && (
                <span className="flex items-center">{tab.icon}</span>
              )}
              <span>{tab.label}</span>
              {tab.count !== undefined && (
                <span
                  className={cn(
                    'rounded-full px-1.5 py-0.2 text-[10px] font-mono',
                    isActive
                      ? 'bg-zinc-100 text-zinc-800 dark:bg-zinc-700 dark:text-zinc-200'
                      : 'bg-zinc-200/60 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400'
                  )}
                >
                  {tab.count}
                </span>
              )}
            </button>
          );
        })}
      </div>
    );
  }

  return (
    <div
      role="tablist"
      className={cn(
        'flex items-center border-b border-zinc-200 dark:border-zinc-800',
        fullWidth && 'w-full justify-between',
        className
      )}
    >
      {tabs.map((tab) => {
        const isActive = tab.id === activeTab;
        return (
          <button
            key={tab.id}
            role="tab"
            type="button"
            aria-selected={isActive}
            disabled={tab.disabled}
            onClick={() => !tab.disabled && onChange(tab.id)}
            className={cn(
              'relative inline-flex items-center gap-2 px-3.5 py-2 text-xs font-medium transition-colors',
              'disabled:cursor-not-allowed disabled:opacity-50',
              isActive
                ? 'text-zinc-900 font-semibold dark:text-zinc-100'
                : 'text-zinc-500 hover:text-zinc-800 dark:text-zinc-400 dark:hover:text-zinc-200',
              fullWidth && 'flex-1 justify-center'
            )}
          >
            {tab.icon && (
              <span className="flex items-center">{tab.icon}</span>
            )}
            <span>{tab.label}</span>
            {tab.count !== undefined && (
              <span
                className={cn(
                  'rounded-full px-1.5 py-0.2 text-[10px] font-mono',
                  isActive
                    ? 'bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900'
                    : 'bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400'
                )}
              >
                {tab.count}
              </span>
            )}

            {/* Subtle active underline indicator */}
            {isActive && (
              <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-zinc-900 dark:bg-zinc-100" />
            )}
          </button>
        );
      })}
    </div>
  );
}

export default Tabs;
