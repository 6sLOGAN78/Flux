import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { UTMBuilderStudio, buildSanitizedUtmUrl } from './UTMBuilderStudio';

describe('UTM Builder Studio', () => {
  it('serializes UTM parameters properly into destination URL', () => {
    const serialized = buildSanitizedUtmUrl({
      baseUrl: 'https://example.com/pricing',
      utmSource: 'twitter',
      utmMedium: 'social',
      utmCampaign: 'summer_2026',
      utmTerm: 'edge-redirects',
      utmContent: 'hero_cta',
    });

    expect(serialized).toContain('https://example.com/pricing');
    expect(serialized).toContain('utm_source=twitter');
    expect(serialized).toContain('utm_medium=social');
    expect(serialized).toContain('utm_campaign=summer_2026');
    expect(serialized).toContain('utm_term=edge-redirects');
    expect(serialized).toContain('utm_content=hero_cta');
  });

  it('renders UTMBuilderStudio with preset channels and real-time preview', () => {
    const html = renderToString(
      <UTMBuilderStudio
        initialBaseUrl="https://flux.to"
        onGenerateLink={() => {}}
      />
    );

    expect(html).toContain('Visual UTM Builder');
    expect(html).toContain('Google Ads');
    expect(html).toContain('Meta / FB');
    expect(html).toContain('Twitter / X');
    expect(html).toContain('LinkedIn');
    expect(html).toContain('Generated Target URL Preview');
    expect(html).toContain('Create Campaign Link');
  });
});
