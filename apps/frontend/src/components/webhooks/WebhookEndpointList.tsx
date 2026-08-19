import React from 'react';
import { Globe, Trash2, Zap, Shield, CheckCircle } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface WebhookEndpoint {
  id: string;
  url: string;
  events: string[];
  status: 'active' | 'disabled' | 'failing';
  createdAt: string;
}

export interface WebhookEndpointListProps {
  endpoints: WebhookEndpoint[];
  onDeleteEndpoint: (id: string) => void;
  onTestEndpoint: (id: string) => void;
  isLoading?: boolean;
  className?: string;
}

export function WebhookEndpointList({
  endpoints,
  onDeleteEndpoint,
  onTestEndpoint,
  isLoading = false,
  className,
}: WebhookEndpointListProps) {
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
            className="flex flex-col justify-between gap-4 rounded-xl border border-zinc-200 bg-zinc-50/50 p-4 transition-all dark:border-zinc-800 dark:bg-zinc-900/40 sm:flex-row sm:items-center"
          >
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Globe className="h-4 w-4 text-zinc-400" />
                <span className="font-mono text-xs font-bold text-zinc-900 dark:text-zinc-100 break-all">
                  {ep.url}
                </span>
                <Badge
                  variant={ep.status === 'active' ? 'emerald' : 'rose'}
                  size="sm"
                  dot
                >
                  {ep.status === 'active' ? 'Active' : ep.status}
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

            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => onTestEndpoint(ep.id)}
                leftIcon={<Zap className="h-3 w-3" />}
              >
                Send Test Event
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

export default WebhookEndpointList;
