import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { AppLayout } from './AppLayout';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import { WorkspaceSwitcher } from './WorkspaceSwitcher';
import { CommandPalette } from './CommandPalette';

describe('AppLayout & Navigation Components', () => {
  it('renders AppLayout with sidebar, header, and main content area', () => {
    const html = renderToString(
      <MemoryRouter initialEntries={['/dashboard']}>
        <AppLayout>
          <div data-testid="dashboard-content">Dashboard Metrics Overview</div>
        </AppLayout>
      </MemoryRouter>
    );

    expect(html).toContain('Dashboard Metrics Overview');
    expect(html).toContain('Flux');
    expect(html).toContain('Links');
    expect(html).toContain('Analytics');
    expect(html).toContain('Settings');
  });

  it('renders Sidebar with active nav item highlight for current route', () => {
    const html = renderToString(
      <MemoryRouter initialEntries={['/links']}>
        <Sidebar />
      </MemoryRouter>
    );

    expect(html).toContain('Links');
    expect(html).toContain('Analytics');
    expect(html).toContain('Campaigns');
    expect(html).toContain('QR Studio');
    // Active class or attribute check
    expect(html).toContain('href="/links"');
  });

  it('renders WorkspaceSwitcher with active workspace and selector trigger', () => {
    const html = renderToString(
      <WorkspaceSwitcher
        workspaces={[
          { id: 'ws_1', name: 'Acme Corp', slug: 'acme', plan: 'Enterprise' },
          { id: 'ws_2', name: 'Personal Project', slug: 'personal', plan: 'Free' },
        ]}
        activeWorkspaceId="ws_1"
        onSelectWorkspace={() => {}}
      />
    );

    expect(html).toContain('Acme Corp');
    expect(html).toContain('Enterprise');
  });

  it('renders Header with breadcrumbs and Cmd+K search button', () => {
    const html = renderToString(
      <MemoryRouter initialEntries={['/analytics']}>
        <Header onOpenCommandPalette={() => {}} />
      </MemoryRouter>
    );

    expect(html).toContain('⌘K');
    expect(html).toContain('Search');
  });

  it('renders CommandPalette when open with search input and action groups', () => {
    const html = renderToString(
      <MemoryRouter>
        <CommandPalette
          isOpen={true}
          onClose={() => {}}
          actions={[
            { id: 'act_1', title: 'Create Short Link', category: 'Links', shortcut: 'C' },
            { id: 'act_2', title: 'View Analytics', category: 'Navigation', shortcut: 'G A' },
          ]}
        />
      </MemoryRouter>
    );

    expect(html).toContain('Search commands, links, or actions...');
    expect(html).toContain('Create Short Link');
    expect(html).toContain('View Analytics');
    expect(html).toContain('ESC');
  });

  it('does not render CommandPalette dialog when isOpen is false', () => {
    const html = renderToString(
      <MemoryRouter>
        <CommandPalette
          isOpen={false}
          onClose={() => {}}
          actions={[]}
        />
      </MemoryRouter>
    );

    expect(html).not.toContain('Search commands, links, or actions...');
  });
});
