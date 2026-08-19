import React, { useState } from 'react';
import {
  Building2,
  Users,
  Plus,
  Mail,
  Shield,
  Check,
  CreditCard,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import {
  TeamMemberTable,
  TeamMember,
} from '@/components/settings/TeamMemberTable';
import { InviteMemberModal } from '@/components/settings/InviteMemberModal';
import { CreateWorkspaceModal } from '@/components/settings/CreateWorkspaceModal';

const INITIAL_MEMBERS: TeamMember[] = [
  {
    id: 'usr_1',
    name: 'Alex Vance',
    email: 'alex@acme.com',
    role: 'owner',
    status: 'active',
    joinedAt: '2026-01-10T00:00:00Z',
  },
  {
    id: 'usr_2',
    name: 'Sarah Connor',
    email: 'sarah@acme.com',
    role: 'admin',
    status: 'active',
    joinedAt: '2026-02-15T00:00:00Z',
  },
  {
    id: 'usr_3',
    name: 'Elena Rostova',
    email: 'elena@acme.com',
    role: 'member',
    status: 'pending',
    joinedAt: '2026-08-18T00:00:00Z',
  },
];

export function WorkspacesPage() {
  const [members, setMembers] = useState<TeamMember[]>(INITIAL_MEMBERS);
  const [isInviteOpen, setIsInviteOpen] = useState(false);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [workspaceName, setWorkspaceName] = useState('Acme Corp');
  const [workspaceSlug, setWorkspaceSlug] = useState('acme-corp');
  const [notice, setNotice] = useState<string | null>(null);

  const handleInvite = (data: {
    email: string;
    role: TeamMember['role'];
  }) => {
    const newMember: TeamMember = {
      id: `usr_${Date.now()}`,
      name: data.email.split('@')[0],
      email: data.email,
      role: data.role,
      status: 'pending',
      joinedAt: new Date().toISOString(),
    };
    setMembers((prev) => [...prev, newMember]);
    setIsInviteOpen(false);
    setNotice(`Invitation sent to ${data.email}`);
    setTimeout(() => setNotice(null), 3000);
  };

  const handleCreateWorkspace = (data: { name: string; slug: string }) => {
    setWorkspaceName(data.name);
    setWorkspaceSlug(data.slug);
    setIsCreateOpen(false);
    setNotice(`Switched to new workspace "${data.name}"`);
    setTimeout(() => setNotice(null), 3000);
  };

  const handleRemoveMember = (id: string) => {
    setMembers((prev) => prev.filter((m) => m.id !== id));
  };

  const handleChangeRole = (id: string, role: TeamMember['role']) => {
    setMembers((prev) =>
      prev.map((m) => (m.id === id ? { ...m, role } : m))
    );
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              Workspaces &amp; Team Settings
            </h1>
            <Badge variant="zinc" size="sm">
              Pro Plan
            </Badge>
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Manage organization members, role-based access control, and workspace identity.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="md"
            onClick={() => setIsCreateOpen(true)}
            leftIcon={<Building2 className="h-4 w-4" />}
          >
            New Workspace
          </Button>
          <Button
            variant="primary"
            size="md"
            onClick={() => setIsInviteOpen(true)}
            leftIcon={<Plus className="h-4 w-4" />}
          >
            Invite Member
          </Button>
        </div>
      </div>

      {notice && (
        <div className="flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-semibold text-emerald-800 dark:border-emerald-900/50 dark:bg-emerald-950/30 dark:text-emerald-300 animate-in fade-in">
          <Check className="h-4 w-4" />
          <span>{notice}</span>
        </div>
      )}

      {/* Active Workspace Info Card */}
      <div className="flex flex-col justify-between gap-4 rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950 sm:flex-row sm:items-center">
        <div className="flex items-center gap-4">
          <div className="flex h-12 w-12 items-center justify-center rounded-2xl border border-zinc-200 bg-zinc-900 font-mono text-base font-bold text-white shadow-sm dark:border-zinc-700 dark:bg-zinc-100 dark:text-zinc-900">
            {workspaceName.charAt(0)}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-base font-bold text-zinc-900 dark:text-zinc-100">
                {workspaceName}
              </h2>
              <Badge variant="emerald" size="sm" dot>
                Active
              </Badge>
            </div>
            <p className="font-mono text-xs text-zinc-400">
              flux.to/w/{workspaceSlug}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-6 text-xs">
          <div>
            <span className="text-zinc-400">Team Size:</span>
            <span className="ml-1.5 font-mono font-bold text-zinc-900 dark:text-zinc-100">
              {members.length} members
            </span>
          </div>
        </div>
      </div>

      {/* Team Member Table */}
      <TeamMemberTable
        members={members}
        onRemoveMember={handleRemoveMember}
        onChangeRole={handleChangeRole}
      />

      {/* Invite Member Modal */}
      <InviteMemberModal
        isOpen={isInviteOpen}
        onClose={() => setIsInviteOpen(false)}
        onSubmit={handleInvite}
      />

      {/* Create Workspace Modal */}
      <CreateWorkspaceModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        onSubmit={handleCreateWorkspace}
      />
    </div>
  );
}

export default WorkspacesPage;
