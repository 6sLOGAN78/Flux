import React, { useState } from 'react';
import { Globe, Plus, ShieldCheck, RefreshCw, Check } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import {
  DNSVerificationCard,
  CustomDomainItem,
} from '@/components/domains/DNSVerificationCard';
import { DomainSetupModal } from '@/components/domains/DomainSetupModal';

const INITIAL_DOMAINS: CustomDomainItem[] = [
  {
    id: 'dom_1',
    hostname: 'go.brand.com',
    status: 'verified',
    sslStatus: 'active',
    cnameTarget: 'cname.flux.to',
    txtVerificationKey: '_flux-challenge.go.brand.com',
    txtVerificationValue: 'flux-vld-994810293',
    rootRedirectUrl: 'https://brand.com',
    clicksRouted: 49200,
    createdAt: '2026-08-15T12:00:00Z',
  },
  {
    id: 'dom_2',
    hostname: 'links.acmecorp.io',
    status: 'verified',
    sslStatus: 'active',
    cnameTarget: 'cname.flux.to',
    txtVerificationKey: '_flux-challenge.links.acmecorp.io',
    txtVerificationValue: 'flux-vld-318491829',
    rootRedirectUrl: 'https://acmecorp.io',
    clicksRouted: 18450,
    createdAt: '2026-08-17T09:30:00Z',
  },
];

export function DomainsPage() {
  const [domains, setDomains] = useState<CustomDomainItem[]>(INITIAL_DOMAINS);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [verifyingId, setVerifyingId] = useState<string | null>(null);

  const handleAddDomain = (hostname: string) => {
    const newDomain: CustomDomainItem = {
      id: `dom_${Date.now()}`,
      hostname,
      status: 'verified',
      sslStatus: 'active',
      cnameTarget: 'cname.flux.to',
      txtVerificationKey: `_flux-challenge.${hostname}`,
      txtVerificationValue: `flux-vld-${Math.random().toString(36).substring(2, 10)}`,
      clicksRouted: 0,
      createdAt: new Date().toISOString(),
    };
    setDomains((prev) => [newDomain, ...prev]);
    setIsModalOpen(false);
  };

  const handleVerifyDNS = (id: string) => {
    setVerifyingId(id);
    setTimeout(() => {
      setVerifyingId(null);
    }, 800);
  };

  const handleDeleteDomain = (id: string) => {
    setDomains((prev) => prev.filter((d) => d.id !== id));
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

      {/* Domain Cards List */}
      <div className="space-y-4">
        {domains.map((dom) => (
          <DNSVerificationCard
            key={dom.id}
            domain={dom}
            onVerifyDNS={handleVerifyDNS}
            onDelete={handleDeleteDomain}
            isVerifying={verifyingId === dom.id}
          />
        ))}
      </div>

      {/* Domain Setup Modal */}
      <DomainSetupModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleAddDomain}
      />
    </div>
  );
}

export default DomainsPage;
