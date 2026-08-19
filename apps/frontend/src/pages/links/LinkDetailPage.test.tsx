import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { LinkDetailPage } from './LinkDetailPage';
import { CategoriesPage } from './CategoriesPage';
import { QRStudioCanvas } from '@/components/qr/QRStudioCanvas';
import { CategoryGrid, CategoryItem } from '@/components/categories/CategoryGrid';

describe('Link Detail & QR Studio', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  it('renders QRStudioCanvas with color controls and SVG/PNG export buttons', () => {
    const html = renderToString(
      <QRStudioCanvas
        url="https://flux.to/v2-launch"
        initialFgColor="#09090b"
        initialBgColor="#ffffff"
      />
    );

    expect(html).toContain('QR Code Studio');
    expect(html).toContain('Foreground Color');
    expect(html).toContain('Background Color');
    expect(html).toContain('Download PNG');
    expect(html).toContain('Download SVG');
  });

  it('renders CategoryGrid with category cards and link count badges', () => {
    const categories: CategoryItem[] = [
      {
        id: 'cat_1',
        name: 'Marketing Campaigns',
        color: '#10b981',
        description: 'Public marketing & social media campaigns',
        linkCount: 24,
      },
      {
        id: 'cat_2',
        name: 'API Documentation',
        color: '#3b82f6',
        description: 'Developer portal & SDK reference links',
        linkCount: 12,
      },
    ];

    const html = renderToString(
      <CategoryGrid
        categories={categories}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    );

    expect(html).toContain('Marketing Campaigns');
    expect(html).toContain('24 links');
    expect(html).toContain('API Documentation');
    expect(html).toContain('12 links');
  });

  it('renders LinkDetailPage with link settings form and save button', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter initialEntries={['/links/link_1']}>
          <Routes>
            <Route path="/links/:id" element={<LinkDetailPage />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Link Settings');
    expect(html).toContain('Destination URL');
    expect(html).toContain('Save Changes');
  });

  it('renders CategoriesPage with header, create button, and grid', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <CategoriesPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Categories');
    expect(html).toContain('New Category');
  });
});
