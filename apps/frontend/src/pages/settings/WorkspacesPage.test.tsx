import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { WorkspacesPage } from './WorkspacesPage';
import {
  TeamMemberTable,
  TeamMember,
} from '@/components/settings/TeamMemberTable';
import { InviteMemberModal } from '@/components/settings/InviteMemberModal';
import { CreateWorkspaceModal } from '@/components/settings/CreateWorkspaceModal';

describe('Workspaces & Team RBAC Management Page', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockMembers: TeamMember[] = [
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

  it('renders TeamMemberTable with role badges and active statuses', () => {
    const html = renderToString(
      <TeamMemberTable
        members={mockMembers}
        onRemoveMember={() => {}}
        onChangeRole={() => {}}
      />
    );

    expect(html).toContain('Alex Vance');
    expect(html).toContain('alex@acme.com');
    expect(html).toContain('Owner');
    expect(html).toContain('Sarah Connor');
    expect(html).toContain('Admin');
    expect(html).toContain('Pending');
  });

  it('renders InviteMemberModal with email input and role select', () => {
    const html = renderToString(
      <InviteMemberModal
        isOpen={true}
        onClose={() => {}}
        onSubmit={() => {}}
      />
    );

    expect(html).toContain('Invite Team Member');
    expect(html).toContain('Email Address');
    expect(html).toContain('Workspace Role');
    expect(html).toContain('Send Invitation');
  });

  it('renders CreateWorkspaceModal with workspace name and slug inputs', () => {
    const html = renderToString(
      <CreateWorkspaceModal
        isOpen={true}
        onClose={() => {}}
        onSubmit={() => {}}
      />
    );

    expect(html).toContain('Create Workspace');
    expect(html).toContain('Workspace Name');
    expect(html).toContain('Workspace Slug');
  });

  it('renders full WorkspacesPage with workspace details and members', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <WorkspacesPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Workspaces &amp; Team Settings');
    expect(html).toContain('Team Members');
    expect(html).toContain('Invite Member');
  });
});
