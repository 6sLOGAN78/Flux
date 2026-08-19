import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { CampaignsPage } from './CampaignsPage';
import { CampaignListTable, CampaignItem } from '@/components/campaigns/CampaignListTable';

describe('Campaigns & Multi-Channel Marketing Page', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockCampaigns: CampaignItem[] = [
    {
      id: 'cmp_1',
      name: 'Q3 Product Hunt Launch',
      channel: 'ProductHunt',
      utmCampaign: 'ph_launch_v2',
      totalClicks: 8420,
      conversions: 612,
      status: 'active',
      createdAt: '2026-08-18T10:00:00Z',
    },
    {
      id: 'cmp_2',
      name: 'Twitter Developer Ads',
      channel: 'Twitter',
      utmCampaign: 'dev_ads_q3',
      totalClicks: 3290,
      conversions: 184,
      status: 'active',
      createdAt: '2026-08-16T14:00:00Z',
    },
  ];

  it('renders CampaignListTable with campaign metrics and conversion rates', () => {
    const html = renderToString(
      <CampaignListTable campaigns={mockCampaigns} />
    );

    expect(html).toContain('Q3 Product Hunt Launch');
    expect(html).toContain('ProductHunt');
    expect(html).toContain('8,420');
    expect(html).toContain('612');
    expect(html).toContain('Twitter Developer Ads');
  });

  it('renders CampaignsPage with UTM builder tabs and creation modal', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <CampaignsPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Marketing Campaigns');
    expect(html).toContain('Visual UTM Builder');
  });
});
