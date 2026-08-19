import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ABTestingPage } from './ABTestingPage';
import {
  VariantAllocationSlider,
  ABVariant,
  calculateNormalizedWeights,
} from '@/components/abtest/VariantAllocationSlider';
import {
  SignificanceScoreCard,
  calculateSignificance,
} from '@/components/abtest/SignificanceScoreCard';

describe('A/B Traffic Splitter & Statistical Significance', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  it('normalizes variant weights when adjusting allocation', () => {
    const variants: ABVariant[] = [
      { id: 'var_a', name: 'Variant A (Control)', destinationUrl: 'https://flux.to/a', weight: 60 },
      { id: 'var_b', name: 'Variant B (Challenger)', destinationUrl: 'https://flux.to/b', weight: 40 },
    ];

    const updated = calculateNormalizedWeights(variants, 'var_a', 70);
    expect(updated[0].weight).toBe(70);
    expect(updated[1].weight).toBe(30);
  });

  it('computes statistical significance confidence correctly', () => {
    const control = { visitors: 1000, conversions: 50 }; // 5%
    const challenger = { visitors: 1000, conversions: 80 }; // 8%

    const result = calculateSignificance(control, challenger);
    expect(result.liftPercentage).toBe(60); // (8 - 5) / 5 * 100 = 60%
    expect(result.isSignificant).toBe(true);
    expect(result.confidencePercentage).toBeGreaterThan(95);
  });

  it('renders VariantAllocationSlider with percentage weights and URL inputs', () => {
    const variants: ABVariant[] = [
      { id: 'var_a', name: 'Variant A (Control)', destinationUrl: 'https://flux.to/original', weight: 50 },
      { id: 'var_b', name: 'Variant B (New Copy)', destinationUrl: 'https://flux.to/redesign', weight: 50 },
    ];

    const html = renderToString(
      <VariantAllocationSlider
        variants={variants}
        onChangeVariants={() => {}}
      />
    );

    expect(html).toContain('Traffic Split Allocation');
    expect(html).toContain('Variant A (Control)');
    expect(html).toContain('Variant B (New Copy)');
    expect(html).toContain('50%');
  });

  it('renders SignificanceScoreCard with confidence badge and Promote Winner button', () => {
    const html = renderToString(
      <SignificanceScoreCard
        controlVariant={{ name: 'Control A', visitors: 2000, conversions: 100 }}
        challengerVariant={{ name: 'Challenger B', visitors: 2000, conversions: 160 }}
        onPromoteWinner={() => {}}
      />
    );

    expect(html).toContain('Statistical Significance');
    expect(html).toContain('Relative Lift');
    expect(html).toContain('Promote Winner (100% Traffic)');
  });

  it('renders full ABTestingPage with experiment switcher and controls', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <ABTestingPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('A/B Testing &amp; Traffic Splitter');
    expect(html).toContain('Traffic Split Allocation');
    expect(html).toContain('Statistical Significance');
  });
});
