import React, { useState } from 'react';
import { ShieldCheck, ShieldAlert, Plus, Ban, Check } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import {
  ThreatStatsGrid,
  ThreatStats,
} from '@/components/security/ThreatStatsGrid';
import {
  QuarantineLinksTable,
  QuarantineLink,
} from '@/components/security/QuarantineLinksTable';
import { BlacklistDomainModal } from '@/components/security/BlacklistDomainModal';

const INITIAL_STATS: ThreatStats = {
  totalScanned: 248920,
  threatsBlocked: 14,
  reputationScore: 99.99,
  quarantineCount: 2,
};

const INITIAL_QUARANTINE: QuarantineLink[] = [];

export function SecurityScannerPage() {
  const [stats, setStats] = useState<ThreatStats>(INITIAL_STATS);
  const [links, setLinks] = useState<QuarantineLink[]>(INITIAL_QUARANTINE);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);

  const handleDisableLink = (id: string) => {
    setLinks((prev) =>
      prev.map((l) => (l.id === id ? { ...l, status: 'blocked' } : l))
    );
    setNotice('Short link forced into permanent block status across edge nodes.');
    setTimeout(() => setNotice(null), 3000);
  };

  const handleReleaseLink = (id: string) => {
    setLinks((prev) =>
      prev.map((l) => (l.id === id ? { ...l, status: 'released' } : l))
    );
    setNotice('Link released from quarantine (marked as false positive).');
    setTimeout(() => setNotice(null), 3000);
  };

  const handleBlacklistDomain = (data: { domain: string; reason: string }) => {
    setIsModalOpen(false);
    setNotice(`Domain ${data.domain} added to global abuse blocklist.`);
    setTimeout(() => setNotice(null), 3000);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              Security &amp; Abuse Scanner
            </h1>
            <Badge variant="emerald" size="sm" dot>
              Safe Browsing Active
            </Badge>
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Real-time reputation verification, automated malware sandboxing, and malicious domain quarantine.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="md"
            onClick={() => setIsModalOpen(true)}
            leftIcon={<Ban className="h-4 w-4" />}
          >
            Blacklist Domain
          </Button>
        </div>
      </div>

      {notice && (
        <div className="flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-semibold text-emerald-800 dark:border-emerald-900/50 dark:bg-emerald-950/30 dark:text-emerald-300 animate-in fade-in">
          <Check className="h-4 w-4" />
          <span>{notice}</span>
        </div>
      )}

      {/* Threat Stats Grid */}
      <ThreatStatsGrid stats={stats} />

      {/* Quarantine Links Table */}
      <QuarantineLinksTable
        links={links}
        onDisableLink={handleDisableLink}
        onReleaseLink={handleReleaseLink}
      />

      {/* Blacklist Modal */}
      <BlacklistDomainModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleBlacklistDomain}
      />
    </div>
  );
}

export default SecurityScannerPage;
