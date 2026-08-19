import React from 'react';
import { Key, Trash2, Shield, Clock } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface ApiKeyItem {
  id: string;
  name: string;
  tokenPrefix: string;
  scopes: string[];
  createdAt: string;
  lastUsedAt?: string;
}

export interface ApiKeyTableProps {
  keys: ApiKeyItem[];
  onRevokeKey: (id: string) => void;
  isLoading?: boolean;
  className?: string;
}

export function ApiKeyTable({
  keys,
  onRevokeKey,
  isLoading = false,
  className,
}: ApiKeyTableProps) {
  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="border-b border-zinc-100 p-5 dark:border-zinc-900">
        <div className="flex items-center gap-2">
          <Key className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
          <h3 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
            Active REST API Keys
          </h3>
        </div>
        <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
          Secret tokens used to authenticate programmatic requests to the Flux API.
        </p>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs text-zinc-700 dark:text-zinc-300">
          <thead className="border-b border-zinc-200 bg-zinc-50/75 text-[11px] font-semibold uppercase tracking-wider text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900/50 dark:text-zinc-400">
            <tr>
              <th className="px-4 py-3">Key Name</th>
              <th className="px-4 py-3">Secret Prefix</th>
              <th className="px-4 py-3">Granular Scopes</th>
              <th className="px-4 py-3">Created</th>
              <th className="px-4 py-3">Last Active</th>
              <th className="px-4 py-3 text-right">Revoke</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-100 dark:divide-zinc-900">
            {keys.map((k) => (
              <tr
                key={k.id}
                className="hover:bg-zinc-50/60 transition-colors dark:hover:bg-zinc-900/40"
              >
                <td className="px-4 py-3 font-semibold text-zinc-900 dark:text-zinc-100">
                  {k.name}
                </td>
                <td className="px-4 py-3 font-mono text-zinc-600 dark:text-zinc-400">
                  <span className="rounded bg-zinc-100 px-1.5 py-0.5 text-[11px] dark:bg-zinc-800">
                    {k.tokenPrefix}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <div className="flex flex-wrap gap-1">
                    {k.scopes.map((s) => (
                      <Badge key={s} variant="zinc" size="sm" className="font-mono text-[10px]">
                        {s}
                      </Badge>
                    ))}
                  </div>
                </td>
                <td className="px-4 py-3 font-mono text-zinc-400 text-[11px]">
                  {new Date(k.createdAt).toLocaleDateString(undefined, {
                    month: 'short',
                    day: 'numeric',
                    year: 'numeric',
                  })}
                </td>
                <td className="px-4 py-3 font-mono text-zinc-400 text-[11px]">
                  {k.lastUsedAt
                    ? new Date(k.lastUsedAt).toLocaleDateString(undefined, {
                        month: 'short',
                        day: 'numeric',
                      })
                    : 'Never'}
                </td>
                <td className="px-4 py-3 text-right">
                  <button
                    type="button"
                    onClick={() => onRevokeKey(k.id)}
                    className="rounded-md p-1.5 text-zinc-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950/50 dark:hover:text-red-400"
                    title="Revoke API Key"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default ApiKeyTable;
