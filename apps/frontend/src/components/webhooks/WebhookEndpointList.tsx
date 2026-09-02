import React from 'react';
import { Globe, Trash2, Power, PowerOff } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';
import { z } from 'zod';
import { ZWebhook } from '@flux/zod';

export type WebhookEndpoint = z.infer<typeof ZWebhook>;

export interface WebhookEndpointListProps {
  endpoints: WebhookEndpoint[];
  isLoading?: boolean;
  onDeleteEndpoint: (id: string) => void;
  onToggleEndpoint: (id: string, active: boolean) => void;
  selectedId: string | null;
  onSelect: (id: string) => void;
  className?: string;
}

export function WebhookEndpointList({
  endpoints,
  isLoading = false,
  onDeleteEndpoint,
  onToggleEndpoint,
  selectedId,
  onSelect,
  className,
}: WebhookEndpointListProps) {
  if (isLoading) {
    return (
      <div className={cn('rounded-2xl border border-zinc-200 bg-white p-6 dark:border-zinc-800 dark:bg-zinc-950', className)}>
        <p className="text-sm text-zinc-500">Loading webhooks...</p>
      </div>
    );
  }

  if (endpoints.length === 0) {
    return (
      <div className={cn('flex flex-col items-center justify-center rounded-2xl border border-dashed border-zinc-300 py-12 dark:border-zinc-800', className)}>
        <Globe className="h-8 w-8 text-zinc-400 mb-3" />
        <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">No webhooks configured</p>
        <p className="text-xs text-zinc-500 dark:text-zinc-400 mt-1">Add an endpoint to start receiving events.</p>
      </div>
    );
  }

  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="border-b border-zinc-100 pb-4 dark:border-zinc-900">
        <h3 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
          Configured Webhook Endpoints
        </h3>
        <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
          HTTP POST endpoints receiving signed HMAC-SHA256 event payloads.
        </p>
      </div>

      <div className="mt-4 space-y-4">
        {endpoints.map((ep) => (
          <div
            key={ep.id}
            onClick={() => onSelect(ep.id)}
            className={cn(
              "flex cursor-pointer flex-col justify-between gap-4 rounded-xl border p-4 transition-all sm:flex-row sm:items-center",
              selectedId === ep.id 
                ? "border-emerald-500 bg-emerald-50/50 dark:border-emerald-500/50 dark:bg-emerald-950/20" 
                : "border-zinc-200 bg-zinc-50/50 hover:border-zinc-300 dark:border-zinc-800 dark:bg-zinc-900/40 dark:hover:border-zinc-700"
            )}
          >
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Globe className="h-4 w-4 text-zinc-400" />
                <span className="font-mono text-xs font-bold text-zinc-900 dark:text-zinc-100 break-all">
                  {ep.endpoint_url}
                </span>
                <Badge
                  variant={ep.active ? 'emerald' : 'zinc'}
                  size="sm"
                  dot={ep.active}
                >
                  {ep.active ? 'Active' : 'Disabled'}
                </Badge>
              </div>

              <div className="flex flex-wrap gap-1.5">
                {ep.events.map((evt) => (
                  <Badge key={evt} variant="zinc" size="sm" className="font-mono text-[10px]">
                    {evt}
                  </Badge>
                ))}
              </div>
            </div>

            <div className="flex items-center gap-2" onClick={(e) => e.stopPropagation()}>
              <Button
                variant="outline"
                size="sm"
                onClick={() => onToggleEndpoint(ep.id, !ep.active)}
                leftIcon={ep.active ? <PowerOff className="h-3 w-3" /> : <Power className="h-3 w-3" />}
              >
                {ep.active ? 'Disable' : 'Enable'}
              </Button>
              <button
                type="button"
                onClick={() => onDeleteEndpoint(ep.id)}
                className="rounded-md p-1.5 text-zinc-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950/50 dark:hover:text-red-400"
                title="Delete Endpoint"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
