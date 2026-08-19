import React from 'react';
import { Link, NavLink } from 'react-router-dom';
import { Button } from '@/components/ui/Button';

export function PublicHeader() {
  return (
    <header className="sticky top-0 z-40 w-full border-b border-zinc-200/80 bg-white/80 backdrop-blur-md dark:border-zinc-800/80 dark:bg-zinc-950/80">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4 sm:px-6">
        {/* Brand */}
        <div className="flex items-center gap-6">
          <Link to="/" className="flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-md bg-zinc-900 text-xs font-bold text-white dark:bg-zinc-100 dark:text-zinc-900">
              F
            </span>
            <span className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              Flux
            </span>
          </Link>

          <nav className="hidden items-center gap-5 md:flex">
            <NavLink
              to="/"
              className={({ isActive }) =>
                `text-xs font-medium transition-colors ${
                  isActive
                    ? 'text-zinc-900 dark:text-zinc-100 font-semibold'
                    : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100'
                }`
              }
            >
              Product
            </NavLink>
            <NavLink
              to="/pricing"
              className={({ isActive }) =>
                `text-xs font-medium transition-colors ${
                  isActive
                    ? 'text-zinc-900 dark:text-zinc-100 font-semibold'
                    : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100'
                }`
              }
            >
              Pricing
            </NavLink>
            <a
              href="https://github.com/flux-platform/flux"
              target="_blank"
              rel="noreferrer"
              className="text-xs font-medium text-zinc-600 transition-colors hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100"
            >
              GitHub
            </a>
          </nav>
        </div>

        {/* Action CTAs */}
        <div className="flex items-center gap-3">
          <Link to="/sign-in">
            <Button variant="ghost" size="sm">
              Sign In
            </Button>
          </Link>
          <Link to="/sign-up">
            <Button variant="primary" size="sm">
              Start for free
            </Button>
          </Link>
        </div>
      </div>
    </header>
  );
}

export default PublicHeader;
