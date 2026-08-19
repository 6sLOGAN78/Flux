import React from 'react';
import { useLocation, Link } from 'react-router-dom';
import { Search, Menu, Bell, User } from 'lucide-react';

export interface HeaderProps {
  onOpenCommandPalette: () => void;
  onToggleMobileSidebar?: () => void;
}

export function Header({ onOpenCommandPalette, onToggleMobileSidebar }: HeaderProps) {
  const location = useLocation();

  const getBreadcrumbTitle = (pathname: string) => {
    const segments = pathname.split('/').filter(Boolean);
    if (segments.length === 0 || segments[0] === 'dashboard') return 'Overview';
    const first = segments[0];
    return first
      .split('-')
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ');
  };

  return (
    <header className="sticky top-0 z-30 flex h-14 w-full items-center justify-between border-b border-zinc-200 bg-white/80 px-4 backdrop-blur-md dark:border-zinc-800 dark:bg-zinc-950/80 sm:px-6">
      {/* Left: Mobile trigger & Breadcrumbs */}
      <div className="flex items-center gap-3">
        {onToggleMobileSidebar && (
          <button
            type="button"
            onClick={onToggleMobileSidebar}
            className="flex h-8 w-8 items-center justify-center rounded-md border border-zinc-200 text-zinc-600 hover:bg-zinc-100 dark:border-zinc-800 dark:text-zinc-400 dark:hover:bg-zinc-900 md:hidden"
            aria-label="Toggle menu"
          >
            <Menu className="h-4 w-4" />
          </button>
        )}
        <div className="flex items-center gap-1.5 text-xs font-medium text-zinc-500 dark:text-zinc-400">
          <Link to="/dashboard" className="hover:text-zinc-900 dark:hover:text-zinc-100">
            Flux
          </Link>
          <span>/</span>
          <span className="text-zinc-900 dark:text-zinc-100">{getBreadcrumbTitle(location.pathname)}</span>
        </div>
      </div>

      {/* Right: Cmd+K search trigger & Actions */}
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={onOpenCommandPalette}
          className="flex items-center gap-2 rounded-lg border border-zinc-200 bg-zinc-50/80 px-2.5 py-1.5 text-xs text-zinc-500 transition-colors hover:border-zinc-300 hover:bg-zinc-100 dark:border-zinc-800 dark:bg-zinc-900/80 dark:text-zinc-400 dark:hover:border-zinc-700 dark:hover:bg-zinc-900"
        >
          <Search className="h-3.5 w-3.5 text-zinc-400" />
          <span className="hidden sm:inline">Search or jump to...</span>
          <span className="inline sm:hidden">Search</span>
          <kbd className="hidden sm:inline-flex items-center rounded border border-zinc-200 bg-white px-1.5 py-0.5 text-[10px] font-mono text-zinc-500 dark:border-zinc-700 dark:bg-zinc-850 dark:text-zinc-400">
            ⌘K
          </kbd>
        </button>

        <button
          type="button"
          aria-label="Notifications"
          className="flex h-8 w-8 items-center justify-center rounded-lg border border-zinc-200 text-zinc-500 transition-colors hover:bg-zinc-100 dark:border-zinc-800 dark:text-zinc-400 dark:hover:bg-zinc-900"
        >
          <Bell className="h-3.5 w-3.5" />
        </button>
      </div>
    </header>
  );
}

export default Header;
