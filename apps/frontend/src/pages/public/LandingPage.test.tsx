import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { LandingPage } from './LandingPage';

describe('Public Landing Page', () => {
  it('renders landing hero with concise headline and edge performance metrics', () => {
    const html = renderToString(
      <MemoryRouter>
        <LandingPage />
      </MemoryRouter>
    );

    expect(html).toContain('The Modern Link Infrastructure');
    expect(html).toContain('Sub-10ms Edge Redirects');
    expect(html).toContain('Start for free');
    expect(html).toContain('View Pricing');
  });

  it('renders interactive HeroRedirectSimulator component', () => {
    const html = renderToString(
      <MemoryRouter>
        <LandingPage />
      </MemoryRouter>
    );

    expect(html).toContain('Shorten Link');
    expect(html).toContain('flux.to/');
    expect(html).toContain('Edge Latency');
  });

  it('renders core feature highlights section', () => {
    const html = renderToString(
      <MemoryRouter>
        <LandingPage />
      </MemoryRouter>
    );

    expect(html).toContain('Global Edge Mesh');
    expect(html).toContain('Real-Time ClickHouse Analytics');
    expect(html).toContain('Dynamic QR Studio');
  });
});
