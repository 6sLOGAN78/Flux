import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { DataTable } from './DataTable';

interface MockLinkItem {
  id: string;
  slug: string;
  url: string;
  clicks: number;
}

describe('DataTable UI Primitive', () => {
  const columns = [
    { key: 'slug', header: 'Short Link' },
    { key: 'url', header: 'Destination URL' },
    {
      key: 'clicks',
      header: 'Clicks',
      render: (item: MockLinkItem) => `${item.clicks.toLocaleString()} clicks`,
    },
  ];

  const sampleData: MockLinkItem[] = [
    { id: '1', slug: 'docs', url: 'https://flux.to/docs', clicks: 1420 },
    { id: '2', slug: 'blog', url: 'https://flux.to/blog', clicks: 830 },
  ];

  it('renders table headers and rows accurately', () => {
    const html = renderToString(
      <DataTable
        columns={columns}
        data={sampleData}
        keyExtractor={(item) => item.id}
      />
    );

    expect(html).toContain('Short Link');
    expect(html).toContain('Destination URL');
    expect(html).toContain('Clicks');
    expect(html).toContain('docs');
    expect(html).toContain('1,420 clicks');
    expect(html).toContain('blog');
    expect(html).toContain('830 clicks');
  });

  it('renders empty state when data is empty', () => {
    const html = renderToString(
      <DataTable
        columns={columns}
        data={[]}
        keyExtractor={(item) => item.id}
        emptyMessage="No links created yet"
      />
    );

    expect(html).toContain('No links created yet');
  });

  it('renders loading skeleton state when isLoading is true', () => {
    const html = renderToString(
      <DataTable
        columns={columns}
        data={[]}
        keyExtractor={(item) => item.id}
        isLoading={true}
      />
    );

    expect(html).toContain('animate-pulse');
  });
});
