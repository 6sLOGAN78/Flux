import React from 'react';
import { User, Trash2, Mail, Shield, ShieldCheck } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface TeamMember {
  id: string;
  name: string;
  email: string;
  role: 'owner' | 'admin' | 'member' | 'viewer';
  status: 'active' | 'pending';
  joinedAt: string;
}

export interface TeamMemberTableProps {
  members: TeamMember[];
  onRemoveMember: (id: string) => void;
  onChangeRole: (id: string, role: TeamMember['role']) => void;
  isLoading?: boolean;
  className?: string;
}

export function TeamMemberTable({
  members,
  onRemoveMember,
  onChangeRole,
  isLoading = false,
  className,
}: TeamMemberTableProps) {
  const getRoleBadgeVariant = (role: TeamMember['role']) => {
    switch (role) {
      case 'owner':
        return 'zinc';
      case 'admin':
        return 'blue';
      case 'member':
        return 'zinc';
      case 'viewer':
      default:
        return 'zinc';
    }
  };

  const formatRoleLabel = (role: TeamMember['role']) => {
    switch (role) {
      case 'owner':
        return 'Owner';
      case 'admin':
        return 'Admin';
      case 'member':
        return 'Member';
      case 'viewer':
        return 'Viewer';
    }
  };

  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="border-b border-zinc-100 p-5 dark:border-zinc-900">
        <h3 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
          Team Members
        </h3>
        <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
          Manage workspace members, roles, and granular organization permissions.
        </p>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs text-zinc-700 dark:text-zinc-300">
          <thead className="border-b border-zinc-200 bg-zinc-50/75 text-[11px] font-semibold uppercase tracking-wider text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900/50 dark:text-zinc-400">
            <tr>
              <th className="px-4 py-3">User</th>
              <th className="px-4 py-3">Role</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Joined</th>
              <th className="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-100 dark:divide-zinc-900">
            {members.map((m) => {
              const isOwner = m.role === 'owner';
              const initial = m.name ? m.name.charAt(0).toUpperCase() : 'U';

              return (
                <tr
                  key={m.id}
                  className="hover:bg-zinc-50/60 transition-colors dark:hover:bg-zinc-900/40"
                >
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <div className="flex h-7 w-7 items-center justify-center rounded-full bg-zinc-900 font-mono text-xs font-bold text-white dark:bg-zinc-100 dark:text-zinc-900">
                        {initial}
                      </div>
                      <div>
                        <div className="font-semibold text-zinc-900 dark:text-zinc-100">
                          {m.name}
                        </div>
                        <div className="font-mono text-[11px] text-zinc-400">
                          {m.email}
                        </div>
                      </div>
                    </div>
                  </td>

                  <td className="px-4 py-3">
                    <Badge variant={getRoleBadgeVariant(m.role)} size="sm">
                      {formatRoleLabel(m.role)}
                    </Badge>
                  </td>

                  <td className="px-4 py-3">
                    <Badge
                      variant={m.status === 'active' ? 'emerald' : 'amber'}
                      size="sm"
                      dot
                    >
                      {m.status === 'active' ? 'Active' : 'Pending'}
                    </Badge>
                  </td>

                  <td className="px-4 py-3 font-mono text-zinc-400 text-[11px]">
                    {new Date(m.joinedAt).toLocaleDateString(undefined, {
                      month: 'short',
                      day: 'numeric',
                      year: 'numeric',
                    })}
                  </td>

                  <td className="px-4 py-3 text-right">
                    {!isOwner && (
                      <button
                        type="button"
                        onClick={() => onRemoveMember(m.id)}
                        className="rounded-md p-1.5 text-zinc-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950/50 dark:hover:text-red-400"
                        title="Remove member"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default TeamMemberTable;
