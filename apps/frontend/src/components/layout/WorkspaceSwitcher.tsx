import React, { useState, useRef, useEffect } from 'react';
import { ChevronDown, Check, Plus, Building2 } from 'lucide-react';

export interface Workspace {
  id: string;
  name: string;
  slug: string;
  plan?: string;
  logoUrl?: string;
}

export interface WorkspaceSwitcherProps {
  workspaces: Workspace[];
  activeWorkspaceId: string;
  onSelectWorkspace: (workspaceId: string) => void;
  onCreateWorkspace?: () => void;
}

export function WorkspaceSwitcher({
  workspaces,
  activeWorkspaceId,
  onSelectWorkspace,
  onCreateWorkspace,
}: WorkspaceSwitcherProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const activeWorkspace = workspaces.find((w) => w.id === activeWorkspaceId) || workspaces[0];

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  const getMonogram = (name?: string) => {
    if (!name) return 'F';
    return name
      .split(' ')
      .map((part) => part[0])
      .join('')
      .substring(0, 2)
      .toUpperCase();
  };

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex w-full items-center justify-between gap-2.5 rounded-lg border border-zinc-200 bg-white px-2.5 py-1.5 text-left text-xs font-medium text-zinc-900 shadow-2xs transition-colors hover:bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-100 dark:hover:bg-zinc-800/60"
        aria-expanded={isOpen}
      >
        <div className="flex min-w-0 items-center gap-2">
          <div className="flex h-5 w-5 shrink-0 items-center justify-center rounded-md bg-zinc-900 text-[10px] font-bold text-white dark:bg-zinc-100 dark:text-zinc-900">
            {getMonogram(activeWorkspace?.name)}
          </div>
          <div className="min-w-0 flex-1 truncate">
            <span className="block truncate font-medium">{activeWorkspace?.name || 'Select Workspace'}</span>
          </div>
        </div>
        <div className="flex items-center gap-1.5">
          {activeWorkspace?.plan && (
            <span className="rounded bg-zinc-100 px-1.5 py-0.5 text-[10px] font-medium text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400">
              {activeWorkspace.plan}
            </span>
          )}
          <ChevronDown className="h-3.5 w-3.5 text-zinc-400" />
        </div>
      </button>

      {isOpen && (
        <div className="absolute left-0 top-full z-50 mt-1.5 w-full min-w-[200px] rounded-lg border border-zinc-200 bg-white p-1 shadow-lg dark:border-zinc-800 dark:bg-zinc-900">
          <div className="px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-zinc-400">
            Workspaces
          </div>
          <div className="space-y-0.5">
            {workspaces.map((workspace) => {
              const isSelected = workspace.id === activeWorkspace?.id;
              return (
                <button
                  key={workspace.id}
                  type="button"
                  onClick={() => {
                    onSelectWorkspace(workspace.id);
                    setIsOpen(false);
                  }}
                  className={`flex w-full items-center justify-between rounded-md px-2 py-1.5 text-left text-xs transition-colors ${
                    isSelected
                      ? 'bg-zinc-100 font-medium text-zinc-900 dark:bg-zinc-800 dark:text-zinc-50'
                      : 'text-zinc-600 hover:bg-zinc-50 dark:text-zinc-400 dark:hover:bg-zinc-800/50'
                  }`}
                >
                  <div className="flex items-center gap-2 truncate">
                    <div className="flex h-5 w-5 shrink-0 items-center justify-center rounded-sm bg-zinc-200 text-[10px] font-semibold text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300">
                      {getMonogram(workspace.name)}
                    </div>
                    <span className="truncate">{workspace.name}</span>
                  </div>
                  {isSelected && <Check className="h-3.5 w-3.5 text-zinc-900 dark:text-zinc-100" />}
                </button>
              );
            })}
          </div>

          {onCreateWorkspace && (
            <>
              <div className="my-1 border-t border-zinc-100 dark:border-zinc-800" />
              <button
                type="button"
                onClick={() => {
                  onCreateWorkspace();
                  setIsOpen(false);
                }}
                className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-xs text-zinc-600 transition-colors hover:bg-zinc-50 dark:text-zinc-400 dark:hover:bg-zinc-800/50"
              >
                <Plus className="h-3.5 w-3.5" />
                <span>Create Workspace</span>
              </button>
            </>
          )}
        </div>
      )}
    </div>
  );
}

export default WorkspaceSwitcher;
