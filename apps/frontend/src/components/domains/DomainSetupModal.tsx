import React, { useState } from 'react';
import { Globe, ArrowRight, ShieldCheck } from 'lucide-react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';

export interface DomainSetupModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (hostname: string) => void;
  isLoading?: boolean;
}

export function DomainSetupModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading = false,
}: DomainSetupModalProps) {
  const [hostname, setHostname] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!hostname.trim()) return;
    onSubmit(hostname.trim());
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Add Custom Domain"
      description="Brand your short links with your own custom root domain or subdomain."
      footer={
        <>
          <Button variant="outline" size="sm" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            size="sm"
            onClick={handleSubmit}
            isLoading={isLoading}
            disabled={!hostname.trim()}
          >
            Add Domain
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Domain Hostname"
          placeholder="e.g. links.yourbrand.com or go.acme.co"
          required
          autoFocus
          value={hostname}
          onChange={(e) => setHostname(e.target.value)}
          startIcon={<Globe className="h-4 w-4" />}
          description="Enter a subdomain or root domain managed by your DNS provider (Cloudflare, Route53, Vercel, etc.)."
        />

        <div className="rounded-xl border border-zinc-200 bg-zinc-50/75 p-4 dark:border-zinc-800 dark:bg-zinc-900/50">
          <h4 className="text-xs font-semibold text-zinc-900 dark:text-zinc-100">
            CNAME Record Configuration
          </h4>
          <p className="mt-1 text-[11px] text-zinc-500 leading-relaxed dark:text-zinc-400">
            After adding your domain, create a CNAME DNS record pointing to{' '}
            <code className="rounded bg-zinc-200 px-1 py-0.5 font-mono text-zinc-800 dark:bg-zinc-800 dark:text-zinc-200">
              cname.flux.to
            </code>
            . Automated ACME TLS/SSL certificates will be provisioned in under 60 seconds.
          </p>
        </div>
      </form>
    </Modal>
  );
}

export default DomainSetupModal;
