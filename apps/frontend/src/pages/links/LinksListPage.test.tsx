import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { LinksListPage } from './LinksListPage';
import { LinksTable, LinkItem } from '@/components/links/LinksTable';
import { CreateLinkDrawer } from '@/components/links/CreateLinkDrawer';
import { BulkCategorizeModal } from '@/components/links/BulkCategorizeModal';

describe('Links Management Hub', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockLinks: LinkItem[] = [
    {
      id: 'link_1',
      shortCode: 'v2-launch',
      destinationUrl: 'https://flux.to/blog/v2-launch',
      title: 'V2 Launch Announcement',
      clicks: 1420,
      createdAt: '2026-08-19T10:00:00Z',
      category: 'Marketing',
      domain: 'flux.to',
    },
    {
      id: 'link_2',
      shortCode: 'api-reference',
      destinationUrl: 'https://flux.to/docs/api',
      title: 'Developer API Reference',
      clicks: 890,
      createdAt: '2026-08-18T14:30:00Z',
      category: 'Documentation',
      domain: 'flux.to',
    },
  ];

  it('renders LinksTable with Dub-style link rows, favicons, and click badges', () => {
    const html = renderToString(
      <MemoryRouter>
        <LinksTable
          links={mockLinks}
          selectedLinkIds={[]}
          onToggleSelect={() => {}}
          onSelectAll={() => {}}
        />
      </MemoryRouter>
    );

    expect(html).toContain('v2-launch');
    expect(html).toContain('https://flux.to/blog/v2-launch');
    expect(html).toContain('1,420');
    expect(html).toContain('Marketing');
    expect(html).toContain('api-reference');
    expect(html).toContain('Documentation');
  });

  it('renders CreateLinkDrawer with destination URL and custom slug inputs', () => {
    const html = renderToString(
      <CreateLinkDrawer
        isOpen={true}
        onClose={() => {}}
        onSubmit={() => {}}
      />
    );

    expect(html).toContain('Create a new link');
    expect(html).toContain('Destination URL');
    expect(html).toContain('Short Slug');
    expect(html).toContain('Create link');
  });

  it('renders BulkCategorizeModal with category selection', () => {
    const html = renderToString(
      <BulkCategorizeModal
        isOpen={true}
        selectedCount={2}
        onClose={() => {}}
        onCategorize={() => {}}
      />
    );

    expect(html).toContain('Categorize 2 links');
    expect(html).toContain('Category Name');
    expect(html).toContain('Apply Category');
  });

  it('renders full LinksListPage with search bar, filter tabs, and action button', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <LinksListPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Links');
    expect(html).toContain('Search links by title, slug, or URL...');
    expect(html).toContain('Create Link');
  });
});
