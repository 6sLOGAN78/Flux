import React, { useState } from 'react';
import { AlertCircle } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import { useAuth } from '@clerk/clerk-react';

export interface WebhookDeliveryHistoryProps {
  webhookId: string;
  className?: string;
}

export function WebhookDeliveryHistory({
  webhookId,
  className,
}: WebhookDeliveryHistoryProps) {
  const { orgId } = useAuth();
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const { data: deliveries, isLoading, isError } = useQuery({
    queryKey: ['webhooks', orgId, webhookId, 'deliveries'],
    queryFn: async () => {
      const { body, status } = await apiClient.getWebhookDeliveries({
        params: { id: webhookId },
      });
      if (status !== 200) throw new Error('Failed to fetch deliveries');
      return body;
    },
    refetchInterval: 10000, // auto refresh every 10s
  });

  const toggleExpand = (id: string) => {
    setExpandedId((prev) => (prev === id ? null : id));
  };

  if (isLoading) {
    return (
      <div className={cn('rounded-2xl border border-zinc-200 bg-white p-6 dark:border-zinc-800 dark:bg-zinc-950', className)}>
        <p className="text-sm text-zinc-500">Loading deliveries...</p>
      </div>
    );
  }

  if (isError) {
    return (
      <div className={cn('rounded-2xl border border-rose-200 bg-rose-50 p-6 dark:bg-rose-950/20', className)}>
        <p className="text-sm text-rose-600 flex items-center gap-2"><AlertCircle className="w-4 h-4" /> Failed to load deliveries</p>
      </div>
    );
  }

  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="border-b border-zinc-100 p-5 dark:border-zinc-900">
        <h3 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
          Event Delivery History
        </h3>
        <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
          Real-time log of dispatched webhook deliveries, HTTP response codes, and status.
        </p>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs text-zinc-700 dark:text-zinc-300">
          <thead className="border-b border-zinc-200 bg-zinc-50/75 text-[11px] font-semibold uppercase tracking-wider text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900/50 dark:text-zinc-400">
            <tr>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">HTTP Code</th>
              <th className="px-4 py-3">Event ID</th>
              <th className="px-4 py-3">Attempts</th>
              <th className="px-4 py-3">Timestamp</th>
              <th className="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-100 dark:divide-zinc-900">
            {deliveries?.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-zinc-500">
                  No deliveries recorded yet.
                </td>
              </tr>
            )}
            
            {deliveries?.map((del) => {
              const isExpanded = expandedId === del.id;
              
              let statusBadge: React.ReactNode;
              if (del.status === 'success') {
                statusBadge = <Badge variant="emerald" size="sm">SUCCESS</Badge>;
              } else if (del.status === 'retrying') {
                statusBadge = <Badge variant="amber" size="sm">RETRYING</Badge>;
              } else if (del.status === 'dead_letter') {
                statusBadge = <Badge variant="rose" size="sm">DEAD LETTER</Badge>;
              } else {
                statusBadge = <Badge variant="zinc" size="sm">{del.status.toUpperCase()}</Badge>;
              }

              return (
                <React.Fragment key={del.id}>
                  <tr className="hover:bg-zinc-50/60 transition-colors dark:hover:bg-zinc-900/40">
                    <td className="px-4 py-3 font-mono">
                      {statusBadge}
                    </td>

                    <td className="px-4 py-3 font-mono font-semibold text-zinc-900 dark:text-zinc-100">
                      {del.response_status || '-'}
                    </td>

                    <td className="px-4 py-3 font-mono text-zinc-500 break-all">
                      {del.event_id}
                    </td>

                    <td className="px-4 py-3 font-mono text-zinc-600 dark:text-zinc-400">
                      {del.attempt_count}
                    </td>

                    <td className="px-4 py-3 font-mono text-zinc-400 text-[11px]">
                      {new Date(del.created_at).toLocaleString()}
                    </td>

                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => toggleExpand(del.id)}
                          className="text-[11px]"
                        >
                          {isExpanded ? 'Hide Payload' : 'Inspect Payload'}
                        </Button>
                      </div>
                    </td>
                  </tr>

                  {isExpanded && (
                    <tr className="bg-zinc-50/75 dark:bg-zinc-900/50">
                      <td colSpan={6} className="px-6 py-4">
                        <div className="space-y-4">
                          {del.last_error && (
                            <div className="text-rose-600 text-xs font-mono bg-rose-50 p-2 rounded-md dark:bg-rose-950/30">
                              Error: {del.last_error}
                            </div>
                          )}
                          {del.next_attempt_at && del.status === 'retrying' && (
                            <div className="text-amber-600 text-xs font-mono bg-amber-50 p-2 rounded-md dark:bg-amber-950/30">
                              Next attempt: {new Date(del.next_attempt_at).toLocaleString()}
                            </div>
                          )}
                          <div>
                            <div className="text-[11px] font-semibold text-zinc-500 uppercase tracking-wider mb-2">
                              JSON Payload
                            </div>
                            <pre className="overflow-x-auto rounded-xl bg-zinc-900 p-4 font-mono text-xs text-zinc-100">
                              {del.payload ? JSON.stringify(del.payload, null, 2) : 'No payload recorded'}
                            </pre>
                          </div>
                        </div>
                      </td>
                    </tr>
                  )}
                </React.Fragment>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
