import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AttributionPage } from './AttributionPage';
import {
  ModelSelectorBar,
  AttributionModelType,
} from '@/components/attribution/ModelSelectorBar';
import {
  AttributionComparisonTable,
  CampaignAttributionItem,
} from '@/components/attribution/AttributionComparisonTable';
import {
  TouchpointTimelineFlow,
  TouchpointNode,
} from '@/components/attribution/TouchpointTimelineFlow';

describe('Multi-Touch Attribution Studio', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockCampaigns: CampaignAttributionItem[] = [
    {
      channel: 'Twitter Paid Ads',
      campaign: 'dev_growth_q3',
      touchpoints: 4120,
      conversions: 184.2,
      revenue: 46050,
      sharePercentage: 38.5,
    },
    {
      channel: 'Google Search CPC',
      campaign: 'brand_keywords',
      touchpoints: 3890,
      conversions: 152.8,
      revenue: 38200,
      sharePercentage: 31.9,
    },
    {
      channel: 'Product Hunt Launch',
      campaign: 'ph_v2_launch',
      touchpoints: 2100,
      conversions: 94.0,
      revenue: 23500,
      sharePercentage: 19.6,
    },
  ];

  const mockTimeline: TouchpointNode[] = [
    { channel: 'Twitter / X', type: 'first_touch', timestamp: 'Aug 10', weightPercentage: 40 },
    { channel: 'Google Search', type: 'middle_touch', timestamp: 'Aug 14', weightPercentage: 10 },
    { channel: 'Blog Post', type: 'middle_touch', timestamp: 'Aug 17', weightPercentage: 10 },
    { channel: 'Direct / Pricing', type: 'last_touch', timestamp: 'Aug 19', weightPercentage: 40 },
  ];

  it('renders ModelSelectorBar with 5 attribution models', () => {
    const html = renderToString(
      <ModelSelectorBar
        selectedModel="u_shaped"
        onSelectModel={() => {}}
      />
    );

    expect(html).toContain('First-Touch');
    expect(html).toContain('Last-Touch');
    expect(html).toContain('Linear');
    expect(html).toContain('Time-Decay');
    expect(html).toContain('Position-Based (U-Shaped)');
  });

  it('renders TouchpointTimelineFlow with journey milestones and weights', () => {
    const html = renderToString(
      <TouchpointTimelineFlow
        touchpoints={mockTimeline}
        model="u_shaped"
      />
    );

    expect(html).toContain('Customer Journey Attribution Path');
    expect(html).toContain('Twitter / X');
    expect(html).toContain('40%');
    expect(html).toContain('Direct / Pricing');
  });

  it('renders AttributionComparisonTable with attributed revenue and conversions', () => {
    const html = renderToString(
      <AttributionComparisonTable
        data={mockCampaigns}
        currency="$"
      />
    );

    expect(html).toContain('Twitter Paid Ads');
    expect(html).toContain('184.2');
    expect(html).toContain('$46,050');
    expect(html).toContain('Google Search CPC');
  });

  it('renders full AttributionPage with header, timeline, and model metrics', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <AttributionPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Multi-Touch Attribution Studio');
    expect(html).toContain('Position-Based (U-Shaped)');
    expect(html).toContain('Attributed Pipeline Revenue');
  });
});
