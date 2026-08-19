import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { FunnelsPage } from './FunnelsPage';
import {
  FunnelVisualizer,
  FunnelStepItem,
} from '@/components/funnels/FunnelVisualizer';
import {
  UnitEconomicsCards,
  UnitEconomicsData,
} from '@/components/revenue/UnitEconomicsCards';
import { CreateFunnelModal } from '@/components/funnels/CreateFunnelModal';

describe('Conversion Funnels & Unit Economics ROAS Dashboard', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockSteps: FunnelStepItem[] = [
    { id: 's1', name: '1. Ad Link Click', visitors: 10000, dropoffPercentage: 0, conversionRateFromStart: 100 },
    { id: 's2', name: '2. Landing Page View', visitors: 6500, dropoffPercentage: 35.0, conversionRateFromStart: 65.0 },
    { id: 's3', name: '3. Pricing Page Visit', visitors: 2600, dropoffPercentage: 60.0, conversionRateFromStart: 26.0 },
    { id: 's4', name: '4. Account Sign Up', visitors: 910, dropoffPercentage: 65.0, conversionRateFromStart: 9.1 },
    { id: 's5', name: '5. Paid Subscription', visitors: 431, dropoffPercentage: 52.6, conversionRateFromStart: 4.31 },
  ];

  const mockEconomics: UnitEconomicsData = {
    totalSpend: 24500,
    attributedRevenue: 119750,
    cac: 56.84,
    roas: 4.89,
    ltv: 240,
    ltvCacRatio: 4.22,
  };

  it('renders FunnelVisualizer with sequential steps and drop-off rates', () => {
    const html = renderToString(
      <FunnelVisualizer steps={mockSteps} funnelName="SaaS Self-Serve Checkout Funnel" />
    );

    expect(html).toContain('SaaS Self-Serve Checkout Funnel');
    expect(html).toContain('1. Ad Link Click');
    expect(html).toContain('10,000');
    expect(html).toContain('5. Paid Subscription');
    expect(html).toContain('4.31% overall conversion');
  });

  it('renders UnitEconomicsCards with CAC, ROAS, and LTV:CAC health indicator', () => {
    const html = renderToString(
      <UnitEconomicsCards data={mockEconomics} />
    );

    expect(html).toContain('Customer Acquisition Cost (CAC)');
    expect(html).toContain('$56.84');
    expect(html).toContain('Return on Ad Spend (ROAS)');
    expect(html).toContain('4.89x');
    expect(html).toContain('LTV:CAC Ratio');
    expect(html).toContain('4.22x');
    expect(html).toContain('Healthy (&gt;3.0x)');
  });

  it('renders CreateFunnelModal with funnel name input and step builder', () => {
    const html = renderToString(
      <CreateFunnelModal
        isOpen={true}
        onClose={() => {}}
        onSubmit={() => {}}
      />
    );

    expect(html).toContain('Create Conversion Funnel');
    expect(html).toContain('Funnel Name');
    expect(html).toContain('Add Funnel Step');
  });

  it('renders full FunnelsPage with economics cards and funnel visualizer', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <FunnelsPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Conversion Funnels &amp; Unit Economics');
    expect(html).toContain('Customer Acquisition Cost (CAC)');
    expect(html).toContain('SaaS Self-Serve Checkout Funnel');
  });
});
