import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { OverviewPage } from './OverviewPage';
import { MetricCardsGrid } from '@/components/dashboard/MetricCardsGrid';
import { RecentActivityFeed } from '@/components/dashboard/RecentActivityFeed';
import { HourlyClickSparkline } from '@/components/dashboard/HourlyClickSparkline';

describe('Overview Dashboard Page', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  it('renders KPI metric cards with clean borders and numbers', () => {
    const html = renderToString(
      <MetricCardsGrid
        metrics={{
          totalLinks: 12480,
          totalClicks24h: 382400,
          avgCtr: 4.8,
          activeDomains: 6,
        }}
      />
    );

    expect(html).toContain('Total Links');
    expect(html).toContain('12,480');
    expect(html).toContain('24h Clicks');
    expect(html).toContain('382.4K');
    expect(html).toContain('Avg CTR');
    expect(html).toContain('4.8%');
    expect(html).toContain('Active Domains');
  });

  it('renders RecentActivityFeed with links and click counts', () => {
    const activities = [
      {
        id: '1',
        shortCode: 'v2-launch',
        originalUrl: 'https://flux.to/blog/v2-launch',
        clicks: 1420,
        createdAt: '2 mins ago',
      },
      {
        id: '2',
        shortCode: 'docs-api',
        originalUrl: 'https://flux.to/docs/api',
        clicks: 890,
        createdAt: '15 mins ago',
      },
    ];

    const html = renderToString(
      <MemoryRouter>
        <RecentActivityFeed activities={activities} />
      </MemoryRouter>
    );

    expect(html).toContain('v2-launch');
    expect(html).toContain('1,420');
    expect(html).toContain('docs-api');
    expect(html).toContain('890');
  });

  it('renders HourlyClickSparkline with time series points', () => {
    const data = [
      { hour: '00:00', clicks: 120 },
      { hour: '04:00', clicks: 340 },
      { hour: '08:00', clicks: 890 },
      { hour: '12:00', clicks: 1420 },
      { hour: '16:00', clicks: 2100 },
      { hour: '20:00', clicks: 1800 },
    ];

    const html = renderToString(
      <HourlyClickSparkline data={data} title="24h Click Volume" />
    );

    expect(html).toContain('24h Click Volume');
  });

  it('renders OverviewPage with quick shortener bar and layout sections', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <OverviewPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Overview');
    expect(html).toContain('Shorten a link');
    expect(html).toContain('Recent Link Activity');
  });
});
