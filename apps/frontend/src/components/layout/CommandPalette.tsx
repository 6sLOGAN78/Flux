import React, { useState, useEffect, useRef } from 'react';
import { Search, Link as LinkIcon, BarChart2, Split, QrCode, Route, Sparkles, X, ArrowRight } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

export interface CommandAction {
  id: string;
  title: string;
  category: string;
  shortcut?: string;
  icon?: React.ReactNode;
  perform?: () => void;
  href?: string;
}

export interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  actions?: CommandAction[];
}

const DEFAULT_ACTIONS: CommandAction[] = [
  { id: 'nav_links', title: 'Go to Links', category: 'Navigation', shortcut: 'G L', href: '/links', icon: <LinkIcon className="h-4 w-4" /> },
  { id: 'nav_analytics', title: 'Go to Analytics', category: 'Navigation', shortcut: 'G A', href: '/analytics', icon: <BarChart2 className="h-4 w-4" /> },
  { id: 'nav_routing', title: 'Go to Smart Routing', category: 'Navigation', shortcut: 'G R', href: '/routing', icon: <Route className="h-4 w-4" /> },
  { id: 'nav_split', title: 'Go to Traffic Splitter', category: 'Navigation', shortcut: 'G S', href: '/traffic-splitter', icon: <Split className="h-4 w-4" /> },
  { id: 'nav_qr', title: 'Go to QR Studio', category: 'Navigation', shortcut: 'G Q', href: '/qr-studio', icon: <QrCode className="h-4 w-4" /> },
  { id: 'act_new_link', title: 'Create Short Link', category: 'Actions', shortcut: 'C', href: '/links/new', icon: <Sparkles className="h-4 w-4" /> },
];

export function CommandPalette({ isOpen, onClose, actions = DEFAULT_ACTIONS }: CommandPaletteProps) {
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  const allActions = actions.length > 0 ? actions : DEFAULT_ACTIONS;

  const filteredActions = allActions.filter((action) =>
    action.title.toLowerCase().includes(query.toLowerCase()) ||
    action.category.toLowerCase().includes(query.toLowerCase())
  );

  useEffect(() => {
    if (isOpen) {
      setQuery('');
      setSelectedIndex(0);
      setTimeout(() => inputRef.current?.focus(), 50);
    }
  }, [isOpen]);

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (!isOpen) return;

      if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
      } else if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSelectedIndex((prev) => (prev + 1) % Math.max(1, filteredActions.length));
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSelectedIndex((prev) => (prev - 1 + filteredActions.length) % Math.max(1, filteredActions.length));
      } else if (e.key === 'Enter') {
        e.preventDefault();
        const selected = filteredActions[selectedIndex];
        if (selected) {
          if (selected.perform) {
            selected.perform();
          } else if (selected.href) {
            navigate(selected.href);
          }
          onClose();
        }
      }
    }

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, selectedIndex, filteredActions, navigate, onClose]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-20 p-4">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/40 backdrop-blur-xs transition-opacity"
        onClick={onClose}
      />

      {/* Modal Dialog */}
      <div className="relative w-full max-w-xl rounded-xl border border-zinc-200 bg-white shadow-2xl overflow-hidden dark:border-zinc-800 dark:bg-zinc-900">
        {/* Search Header */}
        <div className="flex items-center border-b border-zinc-200 px-3.5 py-3 dark:border-zinc-800">
          <Search className="h-4 w-4 text-zinc-400 shrink-0" />
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setSelectedIndex(0);
            }}
            placeholder="Search commands, links, or actions..."
            className="ml-2.5 flex-1 bg-transparent text-sm text-zinc-900 placeholder:text-zinc-400 focus:outline-hidden dark:text-zinc-50"
          />
          <div className="flex items-center gap-1">
            <kbd className="rounded border border-zinc-200 bg-zinc-100 px-1.5 py-0.5 text-[10px] font-mono text-zinc-500 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-400">
              ESC
            </kbd>
          </div>
        </div>

        {/* Action List */}
        <div className="max-h-72 overflow-y-auto p-1.5 space-y-0.5">
          {filteredActions.length === 0 ? (
            <div className="py-8 text-center text-xs text-zinc-500 dark:text-zinc-400">
              No matching commands or links found.
            </div>
          ) : (
            filteredActions.map((action, index) => {
              const isSelected = index === selectedIndex;
              return (
                <button
                  key={action.id}
                  type="button"
                  onClick={() => {
                    if (action.perform) action.perform();
                    else if (action.href) navigate(action.href);
                    onClose();
                  }}
                  onMouseEnter={() => setSelectedIndex(index)}
                  className={`flex w-full items-center justify-between rounded-lg px-2.5 py-2 text-left text-xs transition-colors ${
                    isSelected
                      ? 'bg-zinc-100 text-zinc-900 dark:bg-zinc-800 dark:text-zinc-50'
                      : 'text-zinc-600 hover:bg-zinc-50 dark:text-zinc-400 dark:hover:bg-zinc-800/40'
                  }`}
                >
                  <div className="flex items-center gap-2.5 min-w-0">
                    <span className="text-zinc-400">{action.icon || <ArrowRight className="h-3.5 w-3.5" />}</span>
                    <span className="truncate font-medium">{action.title}</span>
                    <span className="text-[10px] text-zinc-400 font-normal">({action.category})</span>
                  </div>
                  {action.shortcut && (
                    <kbd className="rounded bg-zinc-200/60 px-1.5 py-0.5 text-[10px] font-mono text-zinc-600 dark:bg-zinc-700/60 dark:text-zinc-300">
                      {action.shortcut}
                    </kbd>
                  )}
                </button>
              );
            })
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between border-t border-zinc-100 bg-zinc-50/60 px-3.5 py-2 text-[11px] text-zinc-400 dark:border-zinc-800 dark:bg-zinc-900/60">
          <div className="flex items-center gap-3">
            <span>Use ↑↓ to navigate</span>
            <span>↵ to select</span>
          </div>
          <span>Flux Quick Command</span>
        </div>
      </div>
    </div>
  );
}

export default CommandPalette;
