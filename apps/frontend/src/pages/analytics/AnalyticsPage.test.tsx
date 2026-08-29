import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AnalyticsPage } from './AnalyticsPage';
import {
  TimeSeriesAreaChart,
  TimeSeriesPoint,
} from '@/components/analytics/TimeSeriesAreaChart';
import {
  GeographicChoropleth,
  CountryStat,
} from '@/components/analytics/GeographicChoropleth';
import {
  ReferrerBreakdownTable,
  ReferrerItem,
} from '@/components/analytics/ReferrerBreakdownTable';
import {
  DeviceDonutChart,
  DeviceStat,
} from '@/components/analytics/DeviceDonutChart';

describe('ClickHouse Analytics Explorer', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockTimeSeries: TimeSeriesPoint[] = [
    { timestamp: '00:00', clicks: 120, uniqueVisitors: 90 },
    { timestamp: '04:00', clicks: 240, uniqueVisitors: 190 },
    { timestamp: '08:00', clicks: 890, uniqueVisitors: 720 },
    { timestamp: '12:00', clicks: 1420, uniqueVisitors: 1100 },
    { timestamp: '16:00', clicks: 1180, uniqueVisitors: 940 },
    { timestamp: '20:00', clicks: 650, uniqueVisitors: 510 },
  ];

  const mockCountries: CountryStat[] = [
    { countryCode: 'US', countryName: 'United States', clicks: 14200, percentage: 48.2 },
    { countryCode: 'GB', countryName: 'United Kingdom', clicks: 6400, percentage: 21.7 },
    { countryCode: 'DE', countryName: 'Germany', clicks: 3100, percentage: 10.5 },
  ];

  const mockReferrers: ReferrerItem[] = [
    { domain: 'twitter.com', name: 'Twitter / X', clicks: 8200, percentage: 42.1 },
    { domain: 'google.com', name: 'Google Search', clicks: 5900, percentage: 30.3 },
    { domain: 'direct', name: 'Direct / Email', clicks: 3100, percentage: 15.9 },
  ];

  const mockDevices: DeviceStat[] = [
    { label: 'Mobile (iOS/Android)', value: 16400, percentage: 58.4, color: '#09090b' },
    { label: 'Desktop (macOS/Windows)', value: 10500, percentage: 37.4, color: '#2563eb' },
    { label: 'Tablet & Others', value: 1200, percentage: 4.2, color: '#10b981' },
  ];

  it('renders TimeSeriesAreaChart with time-range tabs and point totals', () => {
    const html = renderToString(
      <TimeSeriesAreaChart data={mockTimeSeries} />
    );

    expect(html).toContain('Click Volume Over Time');
    expect(html).toContain('24h');
    expect(html).toContain('7d');
    expect(html).toContain('30d');
  });

  it('renders GeographicChoropleth with top countries and percentage bars', () => {
    const html = renderToString(
      <GeographicChoropleth countries={mockCountries} />
    );

    expect(html).toContain('Top Geographic Locations');
    expect(html).toContain('United States');
    expect(html).toContain('48.2%');
    expect(html).toContain('United Kingdom');
  });

  it('renders ReferrerBreakdownTable with top domains and percentages', () => {
    const html = renderToString(
      <ReferrerBreakdownTable referrers={mockReferrers} />
    );

    expect(html).toContain('Traffic Referrers');
    expect(html).toContain('Twitter / X');
    expect(html).toContain('Google Search');
  });

  it('renders DeviceDonutChart with device breakdown segments', () => {
    const html = renderToString(
      <DeviceDonutChart devices={mockDevices} />
    );

    expect(html).toContain('Devices &amp; Platforms');
    expect(html).toContain('Mobile (iOS/Android)');
    expect(html).toContain('58.4%');
  });

  it('renders full AnalyticsPage with stream metrics and summary KPIs', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <AnalyticsPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Analytics Explorer');
    expect(html).toContain('ClickHouse Pipeline');
    expect(html).toContain('Total Clicks');
    expect(html).toContain('Stream Compression');
  });
});
