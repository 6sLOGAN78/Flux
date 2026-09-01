import React, { useState, useEffect } from 'react';
import { Globe } from 'lucide-react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';

export interface DomainSetupModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (hostname: string) => void;
  isLoading?: boolean;
  error?: string | null;
}

export function DomainSetupModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading = false,
  error = null,
}: DomainSetupModalProps) {
  const [hostname, setHostname] = useState('');

  // Reset state when modal opens
  useEffect(() => {
    if (isOpen) {
      setHostname('');
    }
  }, [isOpen]);

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
          <Button variant="outline" size="sm" onClick={onClose} disabled={isLoading}>
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
        {error && (
          <div className="rounded-md bg-red-50 p-3 text-sm text-red-600 dark:bg-red-950/50 dark:text-red-400 border border-red-200 dark:border-red-900/50">
            {error}
          </div>
        )}
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
            DNS Verification
          </h4>
          <p className="mt-1 text-[11px] text-zinc-500 leading-relaxed dark:text-zinc-400">
            After adding your domain, you will receive a TXT verification token. 
            You must add this token to your DNS records to prove ownership before traffic can be routed.
          </p>
        </div>
      </form>
    </Modal>
  );
}

export default DomainSetupModal;
