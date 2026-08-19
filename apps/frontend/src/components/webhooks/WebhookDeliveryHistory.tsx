import React, { useState } from 'react';
import { RefreshCw, Code, CheckCircle, AlertCircle, Clock } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface WebhookDeliveryItem {
  id: string;
  eventId: string;
  event: string;
  statusCode: number;
  latencyMs: number;
  timestamp: string;
  requestPayload: string;
  responseBody?: string;
}

export interface WebhookDeliveryHistoryProps {
  deliveries: WebhookDeliveryItem[];
  onRetryDelivery: (id: string) => void;
  isLoading?: boolean;
  className?: string;
}

export function WebhookDeliveryHistory({
  deliveries,
  onRetryDelivery,
  isLoading = false,
  className,
}: WebhookDeliveryHistoryProps) {
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const toggleExpand = (id: string) => {
    setExpandedId((prev) => (prev === id ? null : id));
  };

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
          Real-time log of dispatched webhook deliveries, HTTP response codes, and latency trace.
        </p>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs text-zinc-700 dark:text-zinc-300">
          <thead className="border-b border-zinc-200 bg-zinc-50/75 text-[11px] font-semibold uppercase tracking-wider text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900/50 dark:text-zinc-400">
            <tr>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Event Type</th>
              <th className="px-4 py-3">Event ID</th>
              <th className="px-4 py-3 text-right">Latency</th>
              <th className="px-4 py-3">Timestamp</th>
              <th className="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-100 dark:divide-zinc-900">
            {deliveries.map((del) => {
              const isSuccess = del.statusCode >= 200 && del.statusCode < 300;
              const isExpanded = expandedId === del.id;

              return (
                <React.Fragment key={del.id}>
                  <tr className="hover:bg-zinc-50/60 transition-colors dark:hover:bg-zinc-900/40">
                    <td className="px-4 py-3">
                      <Badge
                        variant={isSuccess ? 'emerald' : 'rose'}
                        size="sm"
                        className="font-mono"
                      >
                        {del.statusCode}
                      </Badge>
                    </td>

                    <td className="px-4 py-3 font-mono font-semibold text-zinc-900 dark:text-zinc-100">
                      {del.event}
                    </td>

                    <td className="px-4 py-3 font-mono text-zinc-500">
                      {del.eventId}
                    </td>

                    <td className="px-4 py-3 text-right font-mono font-medium text-zinc-600 dark:text-zinc-400">
                      {`${del.latencyMs} ms`}
                    </td>

                    <td className="px-4 py-3 font-mono text-zinc-400 text-[11px]">
                      {new Date(del.timestamp).toLocaleTimeString()}
                    </td>

                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => toggleExpand(del.id)}
                          className="text-[11px]"
                        >
                          {isExpanded ? 'Hide Payload' : 'Inspect'}
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => onRetryDelivery(del.id)}
                          leftIcon={<RefreshCw className="h-3 w-3" />}
                        >
                          Retry
                        </Button>
                      </div>
                    </td>
                  </tr>

                  {isExpanded && (
                    <tr className="bg-zinc-50/75 dark:bg-zinc-900/50">
                      <td colSpan={6} className="px-6 py-4">
                        <div className="space-y-2">
                          <div className="text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">
                            JSON Request Payload
                          </div>
                          <pre className="overflow-x-auto rounded-xl bg-zinc-900 p-4 font-mono text-xs text-zinc-100">
                            {JSON.stringify(JSON.parse(del.requestPayload), null, 2)}
                          </pre>
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

export default WebhookDeliveryHistory;
