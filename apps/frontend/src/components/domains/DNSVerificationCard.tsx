import React, { useState } from 'react';
import {
  Globe,
  RefreshCw,
  Trash2,
  Copy,
  Check,
  CheckCircle2,
  Clock,
  XCircle,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';
import type { CustomDomain } from '@flux/zod';

export interface DNSVerificationCardProps {
  domain: CustomDomain;
  onVerifyDNS?: (id: string) => void;
  onDelete: (id: string) => void;
  isVerifying?: boolean;
  className?: string;
}

function StatusBadge({ status }: { status: string }) {
  switch (status) {
    case 'active':
      return (
        <span className="inline-flex items-center gap-1 font-sans text-xs font-semibold text-emerald-600 dark:text-emerald-400">
          <CheckCircle2 className="h-3.5 w-3.5" /> Active
        </span>
      );
    case 'verifying':
      return (
        <span className="inline-flex items-center gap-1 font-sans text-xs font-semibold text-amber-600 dark:text-amber-400">
          <RefreshCw className="h-3.5 w-3.5 animate-spin" /> Verifying
        </span>
      );
    case 'pending':
      return (
        <span className="inline-flex items-center gap-1 font-sans text-xs font-semibold text-blue-600 dark:text-blue-400">
          <Clock className="h-3.5 w-3.5" /> Pending
        </span>
      );
    case 'failed':
      return (
        <span className="inline-flex items-center gap-1 font-sans text-xs font-semibold text-red-600 dark:text-red-400">
          <XCircle className="h-3.5 w-3.5" /> Failed
        </span>
      );
    default:
      return (
        <span className="inline-flex items-center gap-1 font-sans text-xs font-semibold text-zinc-500">
          {status}
        </span>
      );
  }
}

export function DNSVerificationCard({
  domain,
  onVerifyDNS,
  onDelete,
  isVerifying = false,
  className,
}: DNSVerificationCardProps) {
  const [copiedField, setCopiedField] = useState<string | null>(null);

  const handleCopy = (field: string, text: string) => {
    navigator.clipboard?.writeText(text);
    setCopiedField(field);
    setTimeout(() => setCopiedField(null), 2000);
  };

  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="flex flex-col justify-between gap-3 border-b border-zinc-100 pb-4 dark:border-zinc-900 sm:flex-row sm:items-center">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-xl border border-zinc-200 bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900">
            <Globe className="h-4 w-4 text-zinc-700 dark:text-zinc-300" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="font-mono text-sm font-bold text-zinc-900 dark:text-zinc-100">
                {domain.hostname}
              </h3>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {onVerifyDNS && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => onVerifyDNS(domain.id)}
              isLoading={isVerifying}
              leftIcon={<RefreshCw className="h-3.5 w-3.5" />}
            >
              Refresh
            </Button>
          )}
          <button
            type="button"
            onClick={() => onDelete(domain.id)}
            className="rounded-lg p-2 text-zinc-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950/40 dark:hover:text-red-400"
            title="Remove domain"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      </div>

      {domain.status !== 'active' && domain.verification_token && (
        <div className="mt-4 space-y-3">
          <div className="text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
            Required DNS Verification
          </div>
          <div className="overflow-x-auto rounded-xl border border-zinc-200 bg-zinc-50/50 dark:border-zinc-800 dark:bg-zinc-900/30">
            <table className="w-full text-left text-xs">
              <thead className="border-b border-zinc-200 text-[11px] font-semibold text-zinc-500 dark:border-zinc-800 dark:text-zinc-400">
                <tr>
                  <th className="px-4 py-2.5">Type</th>
                  <th className="px-4 py-2.5">Name / Host</th>
                  <th className="px-4 py-2.5">Value / Target</th>
                  <th className="px-4 py-2.5 text-right">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-100 font-mono text-[11px] dark:divide-zinc-900">
                <tr>
                  <td className="px-4 py-2.5 font-bold text-zinc-900 dark:text-zinc-100">
                    TXT
                  </td>
                  <td className="px-4 py-2.5 text-zinc-600 dark:text-zinc-400 truncate max-w-[150px]">
                    {domain.hostname}
                  </td>
                  <td className="px-4 py-2.5">
                    <div className="flex items-center gap-2 text-zinc-900 dark:text-zinc-100">
                      <span className="truncate max-w-xs">{domain.verification_token}</span>
                      <button
                        type="button"
                        onClick={() => handleCopy('txt', domain.verification_token as string)}
                        className="text-zinc-400 hover:text-zinc-700 dark:hover:text-zinc-200 shrink-0"
                      >
                        {copiedField === 'txt' ? (
                          <Check className="h-3 w-3 text-emerald-600 dark:text-emerald-400" />
                        ) : (
                          <Copy className="h-3 w-3" />
                        )}
                      </button>
                    </div>
                  </td>
                  <td className="px-4 py-2.5 text-right">
                    <StatusBadge status={domain.status} />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      )}
      
      {domain.status === 'active' && (
        <div className="mt-4 flex items-center gap-2 rounded-lg bg-emerald-50 px-4 py-3 text-sm text-emerald-800 dark:bg-emerald-950/30 dark:text-emerald-400">
          <CheckCircle2 className="h-4 w-4 text-emerald-600 dark:text-emerald-500" />
          <p>
            This domain is active and verified. DNS and Let's Encrypt TLS certificates are automatically managed.
          </p>
        </div>
      )}
    </div>
  );
}

export default DNSVerificationCard;
