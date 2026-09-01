import { describe, expect, it, mock } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { DomainsPage } from './DomainsPage';
import { DNSVerificationCard } from '@/components/domains/DNSVerificationCard';
import { DomainSetupModal } from '@/components/domains/DomainSetupModal';
import type { CustomDomain } from '@flux/zod';

const mockDomainActive: CustomDomain = {
  id: 'uuid-1',
  hostname: 'link.acme.com',
  status: 'active',
};

const mockDomainPending: CustomDomain = {
  id: 'uuid-2',
  hostname: 'pending.acme.com',
  verification_token: 'flux-verify=test1234',
  status: 'pending',
};

describe('Domains Management Page', () => {
  it('renders DNSVerificationCard for active domain', () => {
    const html = renderToString(
      <DNSVerificationCard
        domain={mockDomainActive}
        onDelete={() => {}}
      />
    );
    expect(html).toContain('link.acme.com');
    expect(html).toContain('active and verified');
    expect(html).not.toContain('Required DNS Verification');
  });

  it('renders DNSVerificationCard for pending domain with instructions', () => {
    const html = renderToString(
      <DNSVerificationCard
        domain={mockDomainPending}
        onVerifyDNS={() => {}}
        onDelete={() => {}}
      />
    );
    expect(html).toContain('pending.acme.com');
    expect(html).toContain('Pending');
    expect(html).toContain('Required DNS Verification');
    expect(html).toContain('flux-verify=test1234');
    expect(html).toContain('TXT');
  });

  it('renders DomainSetupModal properly', () => {
    const html = renderToString(
      <DomainSetupModal
        isOpen={true}
        onClose={() => {}}
        onSubmit={() => {}}
      />
    );
    expect(html).toContain('Add Custom Domain');
    expect(html).toContain('Domain Hostname');
    expect(html).toContain('DNS Verification');
  });

  it('renders DomainSetupModal with error', () => {
    const html = renderToString(
      <DomainSetupModal
        isOpen={true}
        onClose={() => {}}
        onSubmit={() => {}}
        error="Invalid hostname"
      />
    );
    expect(html).toContain('Invalid hostname');
  });
});
