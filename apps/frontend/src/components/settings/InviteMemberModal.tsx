import React, { useState } from 'react';
import { Mail, Shield, UserPlus } from 'lucide-react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { TeamMember } from './TeamMemberTable';

export interface InviteMemberModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: { email: string; role: TeamMember['role'] }) => void;
  isLoading?: boolean;
}

export function InviteMemberModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading = false,
}: InviteMemberModalProps) {
  const [email, setEmail] = useState('');
  const [role, setRole] = useState<TeamMember['role']>('member');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim()) return;
    onSubmit({ email: email.trim(), role });
    setEmail('');
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Invite Team Member"
      description="Grant collaborators access to manage links, campaigns, and domains."
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
            disabled={!email.trim()}
          >
            Send Invitation
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Email Address"
          placeholder="colleague@yourcompany.com"
          type="email"
          required
          autoFocus
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          startIcon={<Mail className="h-4 w-4" />}
        />

        <div>
          <label className="mb-1 block text-xs font-medium text-zinc-700 dark:text-zinc-300">
            Workspace Role
          </label>
          <select
            value={role}
            onChange={(e) => setRole(e.target.value as TeamMember['role'])}
            className="h-9 w-full rounded-lg border border-zinc-200 bg-white px-3 text-xs text-zinc-900 focus:border-zinc-400 focus:outline-none dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-100"
          >
            <option value="admin">Admin (Full write & configuration access)</option>
            <option value="member">Member (Create & edit links, campaigns)</option>
            <option value="viewer">Viewer (Read-only analytics and metrics)</option>
          </select>
          <p className="mt-1 text-[11px] text-zinc-400">
            {role === 'admin' && 'Can manage domains, API keys, and invite members.'}
            {role === 'member' && 'Can shorten URLs, create QR codes, and run A/B tests.'}
            {role === 'viewer' && 'Can view analytics dashboards and read link data.'}
          </p>
        </div>
      </form>
    </Modal>
  );
}

export default InviteMemberModal;
