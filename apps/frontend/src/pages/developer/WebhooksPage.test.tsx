import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { WebhooksPage } from './WebhooksPage';
import {
  WebhookEndpointList,
  WebhookEndpoint,
} from '@/components/webhooks/WebhookEndpointList';
import { CreateWebhookModal } from '@/components/webhooks/CreateWebhookModal';
import {
  WebhookDeliveryHistory,
  WebhookDeliveryItem,
} from '@/components/webhooks/WebhookDeliveryHistory';

describe('Outbound Webhooks Manager & Event Delivery Log', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockEndpoints: WebhookEndpoint[] = [
    {
      id: 'wh_1',
      url: 'https://api.acme.com/webhooks/flux',
      events: ['link.created', 'click.recorded'],
      status: 'active',
      createdAt: '2026-08-10T00:00:00Z',
    },
  ];

  const mockDeliveries: WebhookDeliveryItem[] = [
    {
      id: 'del_1',
      eventId: 'evt_99120',
      event: 'click.recorded',
      statusCode: 200,
      latencyMs: 34,
      timestamp: '2026-08-19T22:30:00Z',
      requestPayload: JSON.stringify({ slug: 'summer-launch', clicks: 1420 }),
      responseBody: '{"received": true}',
    },
    {
      id: 'del_2',
      eventId: 'evt_99121',
      event: 'link.created',
      statusCode: 500,
      latencyMs: 128,
      timestamp: '2026-08-19T22:28:00Z',
      requestPayload: JSON.stringify({ slug: 'promo-2026', url: 'https://acme.com/promo' }),
      responseBody: 'Internal Server Error',
    },
  ];

  it('renders WebhookEndpointList with HTTPS URL and event pills', () => {
    const html = renderToString(
      <WebhookEndpointList
        endpoints={mockEndpoints}
        onDeleteEndpoint={() => {}}
        onTestEndpoint={() => {}}
      />
    );

    expect(html).toContain('https://api.acme.com/webhooks/flux');
    expect(html).toContain('link.created');
    expect(html).toContain('click.recorded');
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
    expect(html).toContain('click.recorded');
  });

  it('renders WebhookDeliveryHistory with status codes, latency, and payloads', () => {
    const html = renderToString(
      <WebhookDeliveryHistory
        deliveries={mockDeliveries}
        onRetryDelivery={() => {}}
      />
    );

    expect(html).toContain('Event Delivery History');
    expect(html).toContain('200');
    expect(html).toContain('34 ms');
    expect(html).toContain('500');
    expect(html).toContain('evt_99120');
  });

  it('renders full WebhooksPage with endpoint manager and delivery history', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <WebhooksPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Outbound Webhooks &amp; Event Bus');
    expect(html).toContain('Add Endpoint');
    expect(html).toContain('https://api.acme.com/webhooks/flux');
  });
});
