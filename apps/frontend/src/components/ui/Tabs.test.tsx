import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { Tabs } from './Tabs';

describe('Tabs UI Primitive', () => {
  it('renders tab triggers with active state indicator', () => {
    const tabs = [
      { id: 'general', label: 'General' },
      { id: 'analytics', label: 'Analytics', count: 12 },
      { id: 'domains', label: 'Domains' },
    ];

    const html = renderToString(
      <Tabs
        tabs={tabs}
        activeTab="general"
        onChange={() => {}}
      />
    );

    expect(html).toContain('General');
    expect(html).toContain('Analytics');
    expect(html).toContain('12');
    expect(html).toContain('Domains');
  });

  it('renders pill style variant properly', () => {
    const tabs = [
      { id: '7d', label: 'Last 7 Days' },
      { id: '30d', label: 'Last 30 Days' },
    ];

    const html = renderToString(
      <Tabs
        tabs={tabs}
        activeTab="7d"
        variant="pills"
        onChange={() => {}}
      />
    );

    expect(html).toContain('Last 7 Days');
    expect(html).toContain('Last 30 Days');
  });
});
