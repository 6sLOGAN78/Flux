import React, { useState } from 'react';
import { Key, Copy, Check, AlertTriangle, ShieldCheck } from 'lucide-react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';

export interface CreateApiKeyModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: { name: string; scopes: string[] }) => void;
  generatedSecret?: string;
  isLoading?: boolean;
}

const AVAILABLE_SCOPES = [
  { id: 'links:read', label: 'links:read', description: 'Read short links, slugs, and destinations' },
  { id: 'links:write', label: 'links:write', description: 'Create, update, and archive short links' },
  { id: 'analytics:read', label: 'analytics:read', description: 'Query time-series click and conversion metrics' },
  { id: 'domains:manage', label: 'domains:manage', description: 'Configure custom domains and verify SSL' },
];

export function CreateApiKeyModal({
  isOpen,
  onClose,
  onSubmit,
  generatedSecret,
  isLoading = false,
}: CreateApiKeyModalProps) {
  const [name, setName] = useState('');
  const [selectedScopes, setSelectedScopes] = useState<string[]>([
    'links:read',
    'links:write',
    'analytics:read',
  ]);
  const [copied, setCopied] = useState(false);

  const toggleScope = (scopeId: string) => {
    setSelectedScopes((prev) =>
      prev.includes(scopeId)
        ? prev.filter((s) => s !== scopeId)
        : [...prev, scopeId]
    );
  };

  const handleCopy = () => {
    if (generatedSecret) {
      navigator.clipboard.writeText(generatedSecret);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || selectedScopes.length === 0) return;
    onSubmit({
      name: name.trim(),
      scopes: selectedScopes,
    });
  };

  // One-time secret reveal modal view
  if (generatedSecret) {
    return (
      <Modal
        isOpen={isOpen}
        onClose={onClose}
        title="API Key Created"
        description="Save this secret key now; it will not be shown again."
        footer={
          <Button variant="primary" size="sm" onClick={onClose}>
            Done
          </Button>
        }
      >
        <div className="space-y-4">
          <div className="rounded-xl border border-amber-200 bg-amber-50/80 p-3 text-xs text-amber-800 dark:border-amber-900/50 dark:bg-amber-950/30 dark:text-amber-300">
            <div className="flex items-center gap-2 font-bold">
              <AlertTriangle className="h-4 w-4 shrink-0 text-amber-600" />
              <span>Save this secret key now</span>
            </div>
            <p className="mt-1 text-[11px] leading-relaxed opacity-90">
              For security reasons, this token will never be displayed in plaintext again.
            </p>
          </div>

          <div className="flex items-center gap-2 rounded-xl border border-zinc-200 bg-zinc-50 p-3 dark:border-zinc-800 dark:bg-zinc-900">
            <code className="flex-1 font-mono text-xs font-semibold text-zinc-900 break-all dark:text-zinc-100">
              {generatedSecret}
            </code>
            <Button
              variant="outline"
              size="sm"
              onClick={handleCopy}
              leftIcon={copied ? <Check className="h-3.5 w-3.5 text-emerald-600" /> : <Copy className="h-3.5 w-3.5" />}
            >
              {copied ? 'Copied' : 'Copy API Key'}
            </Button>
          </div>
        </div>
      </Modal>
    );
  }

  // Create form view
  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Create New API Key"
      description="Generate an authenticated token with granular permission scopes."
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
            disabled={!name.trim() || selectedScopes.length === 0}
          >
            Create API Key
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Key Name"
          placeholder="e.g. Production CI/CD Ingestion Engine"
          required
          autoFocus
          value={name}
          onChange={(e) => setName(e.target.value)}
          startIcon={<Key className="h-4 w-4" />}
        />

        <div className="space-y-2">
          <label className="block text-xs font-medium text-zinc-700 dark:text-zinc-300">
            Permission Scopes
          </label>
          <div className="space-y-2">
            {AVAILABLE_SCOPES.map((scope) => {
              const isChecked = selectedScopes.includes(scope.id);
              return (
                <label
                  key={scope.id}
                  className="flex items-start gap-3 rounded-xl border border-zinc-200 bg-zinc-50/50 p-3 transition-colors hover:bg-zinc-100/50 dark:border-zinc-800 dark:bg-zinc-900/40 dark:hover:bg-zinc-900 cursor-pointer"
                >
                  <input
                    type="checkbox"
                    checked={isChecked}
                    onChange={() => toggleScope(scope.id)}
                    className="mt-0.5 h-4 w-4 rounded border-zinc-300 text-zinc-900 focus:ring-zinc-900 dark:border-zinc-700 dark:bg-zinc-900"
                  />
                  <div>
                    <div className="font-mono text-xs font-bold text-zinc-900 dark:text-zinc-100">
                      {scope.label}
                    </div>
                    <div className="text-[11px] text-zinc-500 dark:text-zinc-400">
                      {scope.description}
                    </div>
                  </div>
                </label>
              );
            })}
          </div>
        </div>
      </form>
    </Modal>
  );
}

export default CreateApiKeyModal;
