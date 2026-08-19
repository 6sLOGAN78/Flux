import React, { useState } from 'react';
import { Building2, Layers } from 'lucide-react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';

export interface CreateWorkspaceModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: { name: string; slug: string }) => void;
  isLoading?: boolean;
}

export function CreateWorkspaceModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading = false,
}: CreateWorkspaceModalProps) {
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');

  const handleNameChange = (val: string) => {
    setName(val);
    const autoSlug = val
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-|-$/g, '');
    setSlug(autoSlug);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !slug.trim()) return;
    onSubmit({ name: name.trim(), slug: slug.trim() });
    setName('');
    setSlug('');
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Create Workspace"
      description="Workspaces provide isolated billing, custom domains, and team access control."
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
            disabled={!name.trim() || !slug.trim()}
          >
            Create Workspace
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Workspace Name"
          placeholder="e.g. Acme Corp Marketing"
          required
          autoFocus
          value={name}
          onChange={(e) => handleNameChange(e.target.value)}
          startIcon={<Building2 className="h-4 w-4" />}
        />

        <Input
          label="Workspace Slug"
          placeholder="acme-marketing"
          prefix="flux.to/w/"
          required
          value={slug}
          onChange={(e) => setSlug(e.target.value)}
          description="A unique URL slug identifying your workspace."
        />
      </form>
    </Modal>
  );
}

export default CreateWorkspaceModal;
