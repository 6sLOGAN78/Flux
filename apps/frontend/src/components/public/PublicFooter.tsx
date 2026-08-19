import React from 'react';
import { Link } from 'react-router-dom';

export function PublicFooter() {
  return (
    <footer className="border-t border-zinc-200 bg-zinc-50/50 dark:border-zinc-800 dark:bg-zinc-950/50">
      <div className="mx-auto max-w-6xl px-4 py-12 sm:px-6">
        <div className="grid grid-cols-2 gap-8 md:grid-cols-4">
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <span className="flex h-5 w-5 items-center justify-center rounded-md bg-zinc-900 text-xs font-bold text-white dark:bg-zinc-100 dark:text-zinc-900">
                F
              </span>
              <span className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
                Flux
              </span>
            </div>
            <p className="text-xs text-zinc-500 dark:text-zinc-400">
              High-performance link infrastructure, ClickHouse analytics, and dynamic QR Studio.
            </p>
            <div className="flex items-center gap-2 pt-1">
              <span className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
              <span className="text-[11px] font-medium text-zinc-600 dark:text-zinc-400">
                Edge Mesh 99.99% Uptime
              </span>
            </div>
          </div>

          <div>
            <div className="text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
              Product
            </div>
            <ul className="mt-3 space-y-2 text-xs">
              <li>
                <Link to="/" className="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100">
                  Global Edge Redirects
                </Link>
              </li>
              <li>
                <Link to="/pricing" className="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100">
                  Pricing Matrix
                </Link>
              </li>
              <li>
                <Link to="/sign-in" className="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100">
                  Dynamic QR Studio
                </Link>
              </li>
              <li>
                <Link to="/sign-in" className="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100">
                  Traffic Splitter
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <div className="text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
              Enterprise
            </div>
            <ul className="mt-3 space-y-2 text-xs">
              <li>
                <Link to="/sso" className="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100">
                  SAML 2.0 & SCIM
                </Link>
              </li>
              <li>
                <Link to="/pricing" className="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100">
                  Uptime SLA
                </Link>
              </li>
              <li>
                <Link to="/pricing" className="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100">
                  Custom Contracts
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <div className="text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
              Developers
            </div>
            <ul className="mt-3 space-y-2 text-xs">
              <li>
                <a href="https://github.com/flux-platform/flux" target="_blank" rel="noreferrer" className="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100">
                  OpenAPI Spec
                </a>
              </li>
              <li>
                <a href="https://github.com/flux-platform/flux" target="_blank" rel="noreferrer" className="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100">
                  TypeScript SDK
                </a>
              </li>
              <li>
                <a href="https://github.com/flux-platform/flux" target="_blank" rel="noreferrer" className="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100">
                  Go Router Engine
                </a>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-12 flex flex-col items-center justify-between gap-4 border-t border-zinc-200 pt-6 dark:border-zinc-800 sm:flex-row">
          <p className="text-[11px] text-zinc-500 dark:text-zinc-400">
            © {new Date().getFullYear()} Flux Inc. All rights reserved. Notion & Dub minimalist design.
          </p>
          <div className="flex items-center gap-4 text-[11px] text-zinc-500 dark:text-zinc-400">
            <span>Privacy</span>
            <span>Terms</span>
            <span>Security</span>
          </div>
        </div>
      </div>
    </footer>
  );
}

export default PublicFooter;
