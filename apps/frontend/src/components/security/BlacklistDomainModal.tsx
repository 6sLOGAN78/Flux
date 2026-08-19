import React, { useState } from 'react';
import { Ban, ShieldAlert, Globe } from 'lucide-react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';

export interface BlacklistDomainModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: { domain: string; reason: string }) => void;
  isLoading?: boolean;
}

export function BlacklistDomainModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading = false,
}: BlacklistDomainModalProps) {
  const [domain, setDomain] = useState('');
  const [reason, setReason] = useState('Phishing / Credential Harvesting');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!domain.trim()) return;
    onSubmit({
      domain: domain.trim().toLowerCase().replace(/^https?:\/\//, ''),
      reason,
    });
    setDomain('');
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Blacklist Malicious Domain"
      description="Permanently prevent any user from shortening or redirecting to this root domain."
      footer={
        <>
          <Button variant="outline" size="sm" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            size="sm"
            onClick={handleSubmit}
            isLoading={isLoading}
            disabled={!domain.trim()}
          >
            Add to Blacklist
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Domain Name"
          placeholder="e.g. malicious-site.xyz"
          required
          autoFocus
          value={domain}
          onChange={(e) => setDomain(e.target.value)}
          startIcon={<Globe className="h-4 w-4" />}
          description="Enter root domain or hostname to block globally."
        />

        <div>
          <label className="mb-1 block text-xs font-medium text-zinc-700 dark:text-zinc-300">
            Block Reason
          </label>
          <select
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            className="h-9 w-full rounded-lg border border-zinc-200 bg-white px-3 text-xs text-zinc-900 focus:border-zinc-400 focus:outline-none dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-100"
          >
            <option value="Phishing / Credential Harvesting">Phishing / Credential Harvesting</option>
            <option value="Malware / Ransomware Payload Host">Malware / Ransomware Payload Host</option>
            <option value="Spam / Automated Bot Abuse">Spam / Automated Bot Abuse</option>
            <option value="Manual Administrative Quarantine">Manual Administrative Quarantine</option>
          </select>
        </div>
      </form>
    </Modal>
  );
}

export default BlacklistDomainModal;
