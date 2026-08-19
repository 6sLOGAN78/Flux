import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SecurityScannerPage } from './SecurityScannerPage';
import {
  QuarantineLinksTable,
  QuarantineLink,
} from '@/components/security/QuarantineLinksTable';
import {
  ThreatStatsGrid,
  ThreatStats,
} from '@/components/security/ThreatStatsGrid';
import { BlacklistDomainModal } from '@/components/security/BlacklistDomainModal';

describe('Abuse Detection & Automated Malware Security Scanner', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockStats: ThreatStats = {
    totalScanned: 248920,
    threatsBlocked: 14,
    reputationScore: 99.99,
    quarantineCount: 2,
  };

  const mockQuarantine: QuarantineLink[] = [
    {
      id: 'quar_1',
      shortUrl: 'flux.to/promo-free',
      destinationUrl: 'http://fake-banking-login.cc/auth',
      threatType: 'phishing',
      provider: 'Google Safe Browsing',
      status: 'quarantined',
      detectedAt: '2026-08-19T22:30:00Z',
    },
    {
      id: 'quar_2',
      shortUrl: 'flux.to/installer',
      destinationUrl: 'http://trojan-download.xyz/payload.exe',
      threatType: 'malware',
      provider: 'VirusTotal',
      status: 'blocked',
      detectedAt: '2026-08-18T14:00:00Z',
    },
  ];

  it('renders ThreatStatsGrid with scanner throughput and threat stats', () => {
    const html = renderToString(
      <ThreatStatsGrid stats={mockStats} />
    );

    expect(html).toContain('Scanned URLs');
    expect(html).toContain('248,920');
    expect(html).toContain('Threats Neutralized');
    expect(html).toContain('14');
    expect(html).toContain('99.99%');
  });

  it('renders QuarantineLinksTable with threat badges and quarantine actions', () => {
    const html = renderToString(
      <QuarantineLinksTable
        links={mockQuarantine}
        onDisableLink={() => {}}
        onReleaseLink={() => {}}
      />
    );

    expect(html).toContain('Quarantined &amp; Flagged Destinations');
    expect(html).toContain('flux.to/promo-free');
    expect(html).toContain('fake-banking-login.cc');
    expect(html).toContain('phishing');
    expect(html).toContain('Google Safe Browsing');
    expect(html).toContain('Force Disable');
  });

  it('renders BlacklistDomainModal with root domain input and reasons', () => {
    const html = renderToString(
      <BlacklistDomainModal
        isOpen={true}
        onClose={() => {}}
        onSubmit={() => {}}
      />
    );

    expect(html).toContain('Blacklist Malicious Domain');
    expect(html).toContain('Domain Name');
    expect(html).toContain('Block Reason');
  });

  it('renders full SecurityScannerPage with stats, table, and header', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <SecurityScannerPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Security &amp; Abuse Scanner');
    expect(html).toContain('Scanned URLs');
    expect(html).toContain('Quarantined &amp; Flagged Destinations');
  });
});
