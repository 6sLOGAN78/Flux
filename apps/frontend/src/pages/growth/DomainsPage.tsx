import React, { useState } from 'react';
import { Plus, Globe } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { DNSVerificationCard } from '@/components/domains/DNSVerificationCard';
import { DomainSetupModal } from '@/components/domains/DomainSetupModal';
import { useGetDomains, useCreateDomain, useDeleteDomain } from '@/hooks/useDomainsQuery';

export function DomainsPage() {
  const { data: domains, isLoading, error, refetch } = useGetDomains();
  const createDomain = useCreateDomain();
  const deleteDomain = useDeleteDomain();
  
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);

  const handleAddDomain = async (hostname: string) => {
    setCreateError(null);
    try {
      await createDomain.mutateAsync({ hostname });
      setIsModalOpen(false);
    } catch (err: any) {
      setCreateError(err.message || 'Failed to add domain');
    }
  };

  const handleVerifyDNS = (id: string) => {
    // There is no explicit verify endpoint in the openapi contract (12H instructions).
    // DNS verification worker (12D) polls automatically. 
    // We just refetch the domains list to see if the status changed.
    refetch();
  };

  const handleDeleteDomain = async (id: string) => {
    if (window.confirm("Remove domain? This may stop traffic from this custom domain.")) {
      try {
        await deleteDomain.mutateAsync(id);
      } catch (err: any) {
        alert(err.message || 'Failed to delete domain');
      }
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              Custom Branded Domains
            </h1>
            <Badge variant="emerald" size="sm" dot>
              Anycast SSL Active
            </Badge>
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Connect custom subdomains and root domains with automatic Let's Encrypt TLS certificates.
          </p>
        </div>

        <Button
          variant="primary"
          size="md"
          onClick={() => setIsModalOpen(true)}
          leftIcon={<Plus className="h-4 w-4" />}
        >
          Add Domain
        </Button>
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="flex h-32 items-center justify-center rounded-2xl border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <p className="text-sm text-zinc-500">Loading domains...</p>
        </div>
      ) : error ? (
        <div className="flex h-32 items-center justify-center rounded-2xl border border-red-200 bg-red-50 p-6 dark:border-red-900/50 dark:bg-red-950/20">
          <p className="text-sm text-red-600 dark:text-red-400">Failed to load domains: {(error as Error).message}</p>
        </div>
      ) : domains?.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-zinc-200 bg-zinc-50 py-12 dark:border-zinc-800 dark:bg-zinc-900/50">
          <Globe className="mb-4 h-8 w-8 text-zinc-400 dark:text-zinc-600" />
          <h3 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">No custom domains yet</h3>
          <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
            Add a custom domain to create branded short links.
          </p>
          <Button
            variant="outline"
            size="sm"
            className="mt-6"
            onClick={() => setIsModalOpen(true)}
            leftIcon={<Plus className="h-4 w-4" />}
          >
            Add Domain
          </Button>
        </div>
      ) : (
        <div className="space-y-4">
          {domains?.map((dom) => (
            <DNSVerificationCard
              key={dom.id}
              domain={dom}
              onVerifyDNS={handleVerifyDNS}
              onDelete={handleDeleteDomain}
            />
          ))}
        </div>
      )}

      {/* Domain Setup Modal */}
      <DomainSetupModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleAddDomain}
        isLoading={createDomain.isPending}
        error={createError}
      />
    </div>
  );
}

export default DomainsPage;
