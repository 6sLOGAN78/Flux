import { describe, expect, it, mock } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

mock.module('@clerk/clerk-react', () => ({
  useAuth: () => ({ orgId: 'org_123' }),
}));

import { WebhooksPage } from './WebhooksPage';
import {
  WebhookEndpointList,
  WebhookEndpoint,
} from '@/components/webhooks/WebhookEndpointList';
import { CreateWebhookModal } from '@/components/webhooks/CreateWebhookModal';
import { WebhookDeliveryHistory } from '@/components/webhooks/WebhookDeliveryHistory';

describe('Outbound Webhooks Manager & Event Delivery Log', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockEndpoints: WebhookEndpoint[] = [
    {
      id: 'wh_1',
      workspace_id: 'org_123',
      endpoint_url: 'https://api.acme.com/webhooks/flux',
      events: ['link.redirect'],
      active: true,
      created_at: '2026-08-10T00:00:00Z',
      updated_at: '2026-08-10T00:00:00Z',
    },
  ];

  it('renders WebhookEndpointList with HTTPS URL and event pills', () => {
    const html = renderToString(
      <WebhookEndpointList
        endpoints={mockEndpoints}
        onDeleteEndpoint={() => {}}
        onToggleEndpoint={() => {}}
        onSelect={() => {}}
        selectedId={null}
      />
    );

    expect(html).toContain('https://api.acme.com/webhooks/flux');
    expect(html).toContain('link.redirect');
    expect(html).toContain('Active');
  });

  it('renders CreateWebhookModal with URL input and event trigger checkboxes', () => {
    const html = renderToString(
      <CreateWebhookModal
        isOpen={true}
        onClose={() => {}}
        onSubmit={() => {}}
      />
    );

    expect(html).toContain('Register Webhook Endpoint');
    expect(html).toContain('Endpoint URL (HTTPS)');
    expect(html).toContain('Event Subscriptions');
    expect(html).toContain('link.redirect');
  });

  it('renders WebhookDeliveryHistory fetching component', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <WebhookDeliveryHistory webhookId="wh_1" />
      </QueryClientProvider>
    );

    expect(html).toContain('Loading deliveries');
  });

  it('renders full WebhooksPage with endpoint manager', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <WebhooksPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Outbound Webhooks');
    expect(html).toContain('Add Endpoint');
  });
});
