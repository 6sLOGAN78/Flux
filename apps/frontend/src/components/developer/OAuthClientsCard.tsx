import React from 'react';
import { ShieldCheck, Plus, RefreshCw, Globe } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface OAuthClientItem {
  id: string;
  name: string;
  clientId: string;
  redirectUris: string[];
  createdAt: string;
}

export interface OAuthClientsCardProps {
  clients: OAuthClientItem[];
  onRotateSecret: (id: string) => void;
  onRegisterClient: () => void;
  className?: string;
}

export function OAuthClientsCard({
  clients,
  onRotateSecret,
  onRegisterClient,
  className,
}: OAuthClientsCardProps) {
  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="flex flex-col justify-between gap-4 border-b border-zinc-100 pb-4 dark:border-zinc-900 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <ShieldCheck className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h3 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              OAuth 2.0 Client Applications
            </h3>
          </div>
          <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
            Allow third-party applications to authenticate users using OAuth 2.0 PKCE.
          </p>
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={onRegisterClient}
          leftIcon={<Plus className="h-3.5 w-3.5" />}
        >
          Register App
        </Button>
      </div>

      <div className="mt-4 space-y-4">
        {clients.map((c) => (
          <div
            key={c.id}
            className="flex flex-col justify-between gap-4 rounded-xl border border-zinc-200 bg-zinc-50/50 p-4 dark:border-zinc-800 dark:bg-zinc-900/40 sm:flex-row sm:items-center"
          >
            <div>
              <div className="flex items-center gap-2">
                <span className="text-xs font-bold text-zinc-900 dark:text-zinc-100">
                  {c.name}
                </span>
                <Badge variant="emerald" size="sm" dot>
                  Enabled
                </Badge>
              </div>

              <div className="mt-2 flex flex-wrap items-center gap-2 font-mono text-[11px]">
                <span className="text-zinc-400">client_id:</span>
                <span className="rounded bg-zinc-200/70 px-1.5 py-0.5 font-bold text-zinc-800 dark:bg-zinc-800 dark:text-zinc-200">
                  {c.clientId}
                </span>
              </div>

              <div className="mt-2 flex items-center gap-1.5 text-[11px] text-zinc-400">
                <Globe className="h-3.5 w-3.5" />
                <span>Redirect URIs: {c.redirectUris.join(', ')}</span>
              </div>
            </div>

            <div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => onRotateSecret(c.id)}
                leftIcon={<RefreshCw className="h-3 w-3" />}
              >
                Rotate Secret
              </Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default OAuthClientsCard;
