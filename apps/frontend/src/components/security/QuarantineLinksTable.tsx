import React from 'react';
import { ShieldAlert, ShieldCheck, Ban, CheckCircle, ExternalLink } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface QuarantineLink {
  id: string;
  shortUrl: string;
  destinationUrl: string;
  threatType: 'phishing' | 'malware' | 'spam' | 'suspicious';
  provider: 'Google Safe Browsing' | 'VirusTotal' | 'Flux Guard';
  status: 'quarantined' | 'blocked' | 'released';
  detectedAt: string;
}

export interface QuarantineLinksTableProps {
  links: QuarantineLink[];
  onDisableLink: (id: string) => void;
  onReleaseLink: (id: string) => void;
  isLoading?: boolean;
  className?: string;
}

export function QuarantineLinksTable({
  links,
  onDisableLink,
  onReleaseLink,
  isLoading = false,
  className,
}: QuarantineLinksTableProps) {
  const getThreatBadgeVariant = (threat: QuarantineLink['threatType']) => {
    switch (threat) {
      case 'phishing':
      case 'malware':
        return 'rose';
      case 'spam':
        return 'amber';
      case 'suspicious':
      default:
        return 'zinc';
    }
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
          Quarantined &amp; Flagged Destinations
        </h3>
        <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
          Links intercepted and halted from redirection due to confirmed or suspected threat signatures.
        </p>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs text-zinc-700 dark:text-zinc-300">
          <thead className="border-b border-zinc-200 bg-zinc-50/75 text-[11px] font-semibold uppercase tracking-wider text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900/50 dark:text-zinc-400">
            <tr>
              <th className="px-4 py-3">Short URL</th>
              <th className="px-4 py-3">Flagged Destination</th>
              <th className="px-4 py-3">Threat Category</th>
              <th className="px-4 py-3">Detection Provider</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3 text-right">Moderator Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-100 dark:divide-zinc-900">
            {links.map((link) => (
              <tr
                key={link.id}
                className="hover:bg-zinc-50/60 transition-colors dark:hover:bg-zinc-900/40"
              >
                <td className="px-4 py-3 font-mono font-bold text-zinc-900 dark:text-zinc-100">
                  {link.shortUrl}
                </td>

                <td className="px-4 py-3 font-mono text-zinc-600 dark:text-zinc-400 max-w-xs truncate">
                  {link.destinationUrl}
                </td>

                <td className="px-4 py-3">
                  <Badge variant={getThreatBadgeVariant(link.threatType)} size="sm">
                    {link.threatType}
                  </Badge>
                </td>

                <td className="px-4 py-3 font-mono text-xs text-zinc-500">
                  {link.provider}
                </td>

                <td className="px-4 py-3">
                  <Badge
                    variant={
                      link.status === 'quarantined'
                        ? 'amber'
                        : link.status === 'blocked'
                        ? 'rose'
                        : 'emerald'
                    }
                    size="sm"
                    dot
                  >
                    {link.status}
                  </Badge>
                </td>

                <td className="px-4 py-3 text-right">
                  <div className="flex items-center justify-end gap-2">
                    {link.status !== 'blocked' && (
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => onDisableLink(link.id)}
                        leftIcon={<Ban className="h-3 w-3" />}
                      >
                        Force Disable
                      </Button>
                    )}
                    {link.status !== 'released' && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => onReleaseLink(link.id)}
                        leftIcon={<CheckCircle className="h-3 w-3" />}
                      >
                        Release
                      </Button>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default QuarantineLinksTable;
