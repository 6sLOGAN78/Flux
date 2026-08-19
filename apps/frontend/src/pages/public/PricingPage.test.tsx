import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { PricingPage } from './PricingPage';

describe('Public Pricing Page', () => {
  it('renders pricing headline and tier cards', () => {
    const html = renderToString(
      <MemoryRouter>
        <PricingPage />
      </MemoryRouter>
    );

    expect(html).toContain('Simple, predictable pricing');
    expect(html).toContain('Free');
    expect(html).toContain('Pro');
    expect(html).toContain('Enterprise');
    expect(html).toContain('Monthly');
    expect(html).toContain('Annual');
  });

  it('renders detailed feature comparison matrix', () => {
    const html = renderToString(
      <MemoryRouter>
        <PricingPage />
      </MemoryRouter>
    );

    expect(html).toContain('Feature Comparison');
    expect(html).toContain('Custom Domains');
    expect(html).toContain('Edge Latency SLA');
    expect(html).toContain('ClickHouse Retention');
  });
});
