import React from 'react';
import { NavLink, Link } from 'react-router-dom';
import {
  Link2,
  BarChart3,
  Globe2,
  GitFork,
  QrCode,
  Layers,
  Settings,
  ShieldCheck,
  CreditCard,
  Webhook,
  Sliders,
  Sparkles,
} from 'lucide-react';
import { WorkspaceSwitcher, Workspace } from './WorkspaceSwitcher';
import { useAuth } from '@/components/auth/AuthContext';

export interface NavItem {
  name: string;
  href: string;
  icon: React.ReactNode;
  badge?: string;
}

export interface SidebarProps {
  workspaces?: Workspace[];
  activeWorkspaceId?: string;
  onSelectWorkspace?: (id: string) => void;
  onCloseMobile?: () => void;
}

const DEFAULT_WORKSPACES: Workspace[] = [
  { id: 'ws_default', name: 'Acme Corp', slug: 'acme', plan: 'Enterprise' },
  { id: 'ws_personal', name: 'Personal', slug: 'personal', plan: 'Pro' },
];

const MAIN_NAV_ITEMS: NavItem[] = [
  { name: 'Overview', href: '/dashboard', icon: <Layers className="h-4 w-4" /> },
  { name: 'Links', href: '/links', icon: <Link2 className="h-4 w-4" /> },
  { name: 'Analytics', href: '/analytics', icon: <BarChart3 className="h-4 w-4" /> },
  { name: 'Campaigns', href: '/campaigns', icon: <Sparkles className="h-4 w-4" /> },
  { name: 'Smart Routing', href: '/routing', icon: <Globe2 className="h-4 w-4" /> },
  { name: 'Traffic Splitter', href: '/traffic-splitter', icon: <GitFork className="h-4 w-4" /> },
  { name: 'QR Studio', href: '/qr-studio', icon: <QrCode className="h-4 w-4" /> },
];

const SETTINGS_NAV_ITEMS: NavItem[] = [
  { name: 'Domains', href: '/settings/domains', icon: <Globe2 className="h-4 w-4" /> },
  { name: 'Webhooks & API', href: '/settings/webhooks', icon: <Webhook className="h-4 w-4" /> },
  { name: 'Billing', href: '/settings/billing', icon: <CreditCard className="h-4 w-4" /> },
  { name: 'Team & RBAC', href: '/settings/team', icon: <ShieldCheck className="h-4 w-4" /> },
  { name: 'Settings', href: '/settings', icon: <Settings className="h-4 w-4" /> },
];

export function Sidebar({
  workspaces,
  activeWorkspaceId = 'ws_default',
  onSelectWorkspace = () => {},
  onCloseMobile,
}: SidebarProps) {
  const { user } = useAuth();

  const userWorkspaceName = user?.workspaceName || 'Development Workspace';

  const defaultWorkspaces: Workspace[] = [
    { id: 'ws_default', name: userWorkspaceName, slug: userWorkspaceName.toLowerCase().replace(/\s+/g, '-'), plan: 'Pro' },
    { id: 'ws_personal', name: 'Personal Workspace', slug: 'personal', plan: 'Free' },
  ];

  const currentWorkspaces = workspaces && workspaces.length > 0 ? workspaces : defaultWorkspaces;

  return (
    <aside className="flex h-full w-64 flex-col border-r border-zinc-200 bg-zinc-50/50 dark:border-zinc-800 dark:bg-zinc-950/50">
      {/* Brand & Workspace Switcher */}
      <div className="p-3">
        <div className="mb-3 flex items-center justify-between px-1">
          <Link to="/" className="flex items-center gap-2 font-semibold text-zinc-900 dark:text-zinc-100">
            <span className="flex h-5 w-5 items-center justify-center rounded-md bg-zinc-900 text-xs font-bold text-white dark:bg-zinc-100 dark:text-zinc-900">
              F
            </span>
            <span className="text-sm tracking-tight">Flux</span>
          </Link>
          <span className="rounded bg-zinc-200/60 px-1.5 py-0.5 text-[10px] font-mono text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400">
            v1.0
          </span>
        </div>

        <WorkspaceSwitcher
          workspaces={currentWorkspaces}
          activeWorkspaceId={activeWorkspaceId}
          onSelectWorkspace={onSelectWorkspace}
        />
      </div>

      {/* Navigation Links */}
      <div className="flex-1 overflow-y-auto px-3 py-2 space-y-6">
        <div>
          <div className="px-2 pb-1.5 text-[10px] font-semibold uppercase tracking-wider text-zinc-400">
            Core Ops
          </div>
          <nav className="space-y-0.5">
            {MAIN_NAV_ITEMS.map((item) => (
              <NavLink
                key={item.href}
                to={item.href}
                onClick={onCloseMobile}
                className={({ isActive }) =>
                  `flex items-center justify-between rounded-lg px-2.5 py-1.5 text-xs font-medium transition-colors ${
                    isActive
                      ? 'bg-zinc-200/70 text-zinc-900 dark:bg-zinc-800 dark:text-zinc-50 font-semibold'
                      : 'text-zinc-600 hover:bg-zinc-100 hover:text-zinc-900 dark:text-zinc-400 dark:hover:bg-zinc-900 dark:hover:text-zinc-100'
                  }`
                }
              >
                <div className="flex items-center gap-2.5">
                  {item.icon}
                  <span>{item.name}</span>
                </div>
                {item.badge && (
                  <span className="rounded bg-emerald-100 px-1.5 py-0.5 text-[9px] font-semibold text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300">
                    {item.badge}
                  </span>
                )}
              </NavLink>
            ))}
          </nav>
        </div>

        <div>
          <div className="px-2 pb-1.5 text-[10px] font-semibold uppercase tracking-wider text-zinc-400">
            Settings & Admin
          </div>
          <nav className="space-y-0.5">
            {SETTINGS_NAV_ITEMS.map((item) => (
              <NavLink
                key={item.href}
                to={item.href}
                onClick={onCloseMobile}
                className={({ isActive }) =>
                  `flex items-center justify-between rounded-lg px-2.5 py-1.5 text-xs font-medium transition-colors ${
                    isActive
                      ? 'bg-zinc-200/70 text-zinc-900 dark:bg-zinc-800 dark:text-zinc-50 font-semibold'
                      : 'text-zinc-600 hover:bg-zinc-100 hover:text-zinc-900 dark:text-zinc-400 dark:hover:bg-zinc-900 dark:hover:text-zinc-100'
                  }`
                }
              >
                <div className="flex items-center gap-2.5">
                  {item.icon}
                  <span>{item.name}</span>
                </div>
              </NavLink>
            ))}
          </nav>
        </div>
      </div>

      {/* Footer / System Status */}
      <div className="border-t border-zinc-200 p-3 dark:border-zinc-800">
        <div className="flex items-center justify-between rounded-lg bg-zinc-100/70 px-2.5 py-2 text-[11px] text-zinc-600 dark:bg-zinc-900/70 dark:text-zinc-400">
          <div className="flex items-center gap-2">
            <span className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
            <span>Edge Mesh Active</span>
          </div>
          <span className="font-mono text-[10px] text-zinc-400">99.99%</span>
        </div>
      </div>
    </aside>
  );
}

export default Sidebar;
