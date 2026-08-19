import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ApiKeysPage } from './ApiKeysPage';
import { ApiKeyTable, ApiKeyItem } from '@/components/developer/ApiKeyTable';
import { CreateApiKeyModal } from '@/components/developer/CreateApiKeyModal';
import {
  OAuthClientsCard,
  OAuthClientItem,
} from '@/components/developer/OAuthClientsCard';

describe('Developer API Keys & OAuth 2.0 Apps Manager', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockApiKeys: ApiKeyItem[] = [
    {
      id: 'key_1',
      name: 'Production Ingestion Worker',
      tokenPrefix: 'flx_live_a1b2...',
      scopes: ['links:read', 'links:write', 'analytics:read'],
      createdAt: '2026-08-01T00:00:00Z',
      lastUsedAt: '2026-08-19T22:00:00Z',
    },
    {
      id: 'key_2',
      name: 'Read-only Analytics Dashboard',
      tokenPrefix: 'flx_live_7c8d...',
      scopes: ['analytics:read'],
      createdAt: '2026-08-10T00:00:00Z',
    },
  ];

  const mockOAuthClients: OAuthClientItem[] = [
    {
      id: 'oauth_1',
      name: 'Slack Integration Bot',
      clientId: 'flux_client_99a8b7',
      redirectUris: ['https://slack.com/oauth/callback/flux'],
      createdAt: '2026-07-20T00:00:00Z',
    },
  ];

  it('renders ApiKeyTable with masked prefixes and granular scope badges', () => {
    const html = renderToString(
      <ApiKeyTable
        keys={mockApiKeys}
        onRevokeKey={() => {}}
      />
    );

    expect(html).toContain('Production Ingestion Worker');
    expect(html).toContain('flx_live_a1b2...');
    expect(html).toContain('links:write');
    expect(html).toContain('analytics:read');
  });

  it('renders CreateApiKeyModal with scope checklist and secret reveal view', () => {
    const html = renderToString(
      <CreateApiKeyModal
        isOpen={true}
        onClose={() => {}}
        onSubmit={() => {}}
        generatedSecret="flx_live_9f83a093be8173491028374a8d"
      />
    );

    expect(html).toContain('Save this secret key now');
    expect(html).toContain('flx_live_9f83a093be8173491028374a8d');
    expect(html).toContain('Copy API Key');
  });

  it('renders OAuthClientsCard with client_id and redirect URIs', () => {
    const html = renderToString(
      <OAuthClientsCard
        clients={mockOAuthClients}
        onRotateSecret={() => {}}
        onRegisterClient={() => {}}
      />
    );

    expect(html).toContain('OAuth 2.0 Client Applications');
    expect(html).toContain('Slack Integration Bot');
    expect(html).toContain('flux_client_99a8b7');
  });

  it('renders full ApiKeysPage with header, key table, and OAuth card', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <ApiKeysPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Developer API Keys &amp; OAuth 2.0');
    expect(html).toContain('Generate API Key');
    expect(html).toContain('Production Ingestion Worker');
  });
});
