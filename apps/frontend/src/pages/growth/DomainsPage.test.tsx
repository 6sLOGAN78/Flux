import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { DomainsPage } from './DomainsPage';
import { DomainSetupModal } from '@/components/domains/DomainSetupModal';
import { DNSVerificationCard, CustomDomainItem } from '@/components/domains/DNSVerificationCard';
import { SSLStatusBadge } from '@/components/domains/SSLStatusBadge';

describe('Custom Branded Domains & DNS Checker', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockDomain: CustomDomainItem = {
    id: 'dom_1',
    hostname: 'go.brand.com',
    status: 'verified',
    sslStatus: 'active',
    cnameTarget: 'cname.flux.to',
    txtVerificationKey: '_flux-challenge.go.brand.com',
    txtVerificationValue: 'flux-vld-8849204924',
    rootRedirectUrl: 'https://brand.com',
    clicksRouted: 49200,
    createdAt: '2026-08-15T12:00:00Z',
  };

  it('renders SSLStatusBadge with active, pending, and error states', () => {
    const activeHtml = renderToString(<SSLStatusBadge status="active" />);
    expect(activeHtml).toContain('SSL Active');

    const pendingHtml = renderToString(<SSLStatusBadge status="pending" />);
    expect(pendingHtml).toContain('SSL Pending');

    const errorHtml = renderToString(<SSLStatusBadge status="error" />);
    expect(errorHtml).toContain('SSL Error');
  });

  it('renders DNSVerificationCard with DNS records and verify action', () => {
    const html = renderToString(
      <DNSVerificationCard
        domain={mockDomain}
        onVerifyDNS={() => {}}
        onDelete={() => {}}
      />
    );

    expect(html).toContain('go.brand.com');
    expect(html).toContain('cname.flux.to');
    expect(html).toContain('SSL Active');
    expect(html).toContain('Verify DNS');
  });

  it('renders DomainSetupModal with DNS CNAME and TXT guidance', () => {
    const html = renderToString(
      <DomainSetupModal
        isOpen={true}
        onClose={() => {}}
        onSubmit={() => {}}
      />
    );

    expect(html).toContain('Add Custom Domain');
    expect(html).toContain('CNAME Record');
    expect(html).toContain('cname.flux.to');
  });

  it('renders full DomainsPage with header and domains list', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <DomainsPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Custom Branded Domains');
    expect(html).toContain('Add Domain');
  });
});
