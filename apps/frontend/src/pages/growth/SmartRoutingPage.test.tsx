import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SmartRoutingPage } from './SmartRoutingPage';
import { GeoRoutingRuleBuilder, GeoRule } from '@/components/routing/GeoRoutingRuleBuilder';
import { DeviceRuleSelector, DeviceRule } from '@/components/routing/DeviceRuleSelector';
import { DeepLinkConfigCard, DeepLinkConfig } from '@/components/routing/DeepLinkConfigCard';

describe('Smart Routing & Mobile Deep Linking Page', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  it('renders GeoRoutingRuleBuilder with country condition sets and destinations', () => {
    const rules: GeoRule[] = [
      {
        id: 'r_1',
        countryCode: 'GB',
        countryName: 'United Kingdom',
        destinationUrl: 'https://uk.store.com/checkout',
      },
      {
        id: 'r_2',
        countryCode: 'DE',
        countryName: 'Germany',
        destinationUrl: 'https://de.store.com/checkout',
      },
    ];

    const html = renderToString(
      <GeoRoutingRuleBuilder
        rules={rules}
        onAddRule={() => {}}
        onDeleteRule={() => {}}
      />
    );

    expect(html).toContain('Geographic Country Targeting');
    expect(html).toContain('United Kingdom');
    expect(html).toContain('https://uk.store.com/checkout');
    expect(html).toContain('Germany');
    expect(html).toContain('https://de.store.com/checkout');
    expect(html).toContain('Add Country Rule');
  });

  it('renders DeviceRuleSelector with OS targeting configurations', () => {
    const rules: DeviceRule[] = [
      {
        id: 'd_1',
        deviceType: 'ios',
        destinationUrl: 'https://apps.apple.com/app/id123456789',
      },
      {
        id: 'd_2',
        deviceType: 'android',
        destinationUrl: 'https://play.google.com/store/apps/details?id=com.flux.app',
      },
    ];

    const html = renderToString(
      <DeviceRuleSelector
        rules={rules}
        onUpdateRule={() => {}}
      />
    );

    expect(html).toContain('Device &amp; OS Targeting');
    expect(html).toContain('Apple iOS');
    expect(html).toContain('Google Android');
  });

  it('renders DeepLinkConfigCard with iOS and Android package identifiers', () => {
    const config: DeepLinkConfig = {
      iosAppStoreId: '123456789',
      iosBundleId: 'com.flux.mobile',
      androidPackageName: 'com.flux.app',
      fallbackUrl: 'https://flux.to/download',
    };

    const html = renderToString(
      <DeepLinkConfigCard
        config={config}
        onSave={() => {}}
      />
    );

    expect(html).toContain('Mobile Deep Linking');
    expect(html).toContain('iOS Bundle ID');
    expect(html).toContain('Android Package Name');
    expect(html).toContain('Fallback Web URL');
  });

  it('renders full SmartRoutingPage with tabs and rules overview', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <SmartRoutingPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Smart Routing');
    expect(html).toContain('Geo Targeting');
    expect(html).toContain('Device &amp; OS');
    expect(html).toContain('Universal Deep Linking');
  });
});
