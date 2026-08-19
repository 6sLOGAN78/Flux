import React, { useState } from 'react';
import { Globe, Plus, Zap, Check, Shield } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import {
  WebhookEndpointList,
  WebhookEndpoint,
} from '@/components/webhooks/WebhookEndpointList';
import { CreateWebhookModal } from '@/components/webhooks/CreateWebhookModal';
import {
  WebhookDeliveryHistory,
  WebhookDeliveryItem,
} from '@/components/webhooks/WebhookDeliveryHistory';

const INITIAL_ENDPOINTS: WebhookEndpoint[] = [
  {
    id: 'wh_1',
    url: 'https://api.acme.com/webhooks/flux',
    events: ['link.created', 'click.recorded'],
    status: 'active',
    createdAt: '2026-08-10T00:00:00Z',
  },
];

const INITIAL_DELIVERIES: WebhookDeliveryItem[] = [
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

export function WebhooksPage() {
  const [endpoints, setEndpoints] = useState<WebhookEndpoint[]>(INITIAL_ENDPOINTS);
  const [deliveries, setDeliveries] =
    useState<WebhookDeliveryItem[]>(INITIAL_DELIVERIES);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);

  const handleAddEndpoint = (data: { url: string; events: string[] }) => {
    const newEndpoint: WebhookEndpoint = {
      id: `wh_${Date.now()}`,
      url: data.url,
      events: data.events,
      status: 'active',
      createdAt: new Date().toISOString(),
    };
    setEndpoints((prev) => [...prev, newEndpoint]);
    setIsModalOpen(false);
    setNotice(`Registered webhook endpoint for ${data.url}`);
    setTimeout(() => setNotice(null), 3000);
  };

  const handleDeleteEndpoint = (id: string) => {
    setEndpoints((prev) => prev.filter((ep) => ep.id !== id));
    setNotice('Webhook endpoint removed');
    setTimeout(() => setNotice(null), 3000);
  };

  const handleTestEndpoint = (id: string) => {
    const endpoint = endpoints.find((ep) => ep.id === id);
    if (!endpoint) return;

    const testDelivery: WebhookDeliveryItem = {
      id: `del_${Date.now()}`,
      eventId: `evt_test_${Math.random().toString(36).substring(2, 7)}`,
      event: endpoint.events[0] || 'ping',
      statusCode: 200,
      latencyMs: Math.floor(Math.random() * 40) + 15,
      timestamp: new Date().toISOString(),
      requestPayload: JSON.stringify({
        type: 'test_event',
        endpointId: id,
        timestamp: new Date().toISOString(),
      }),
    };

    setDeliveries((prev) => [testDelivery, ...prev]);
    setNotice(`Dispatched signed test ping to ${endpoint.url}`);
    setTimeout(() => setNotice(null), 3000);
  };

  const handleRetryDelivery = (id: string) => {
    const delivery = deliveries.find((d) => d.id === id);
    if (!delivery) return;

    const redelivered: WebhookDeliveryItem = {
      ...delivery,
      id: `del_${Date.now()}`,
      statusCode: 200,
      latencyMs: 28,
      timestamp: new Date().toISOString(),
    };

    setDeliveries((prev) => [redelivered, ...prev]);
    setNotice(`Re-delivered event ${delivery.eventId} (HTTP 200 OK)`);
    setTimeout(() => setNotice(null), 3000);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              Outbound Webhooks &amp; Event Bus
            </h1>
            <Badge variant="emerald" size="sm" dot>
              HMAC-SHA256 Signed
            </Badge>
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Dispatch asynchronous events to your backend microservices with automatic retries.
          </p>
        </div>

        <Button
          variant="primary"
          size="md"
          onClick={() => setIsModalOpen(true)}
          leftIcon={<Plus className="h-4 w-4" />}
        >
          Add Endpoint
        </Button>
      </div>

      {notice && (
        <div className="flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-semibold text-emerald-800 dark:border-emerald-900/50 dark:bg-emerald-950/30 dark:text-emerald-300 animate-in fade-in">
          <Check className="h-4 w-4" />
          <span>{notice}</span>
        </div>
      )}

      {/* Endpoints List */}
      <WebhookEndpointList
        endpoints={endpoints}
        onDeleteEndpoint={handleDeleteEndpoint}
        onTestEndpoint={handleTestEndpoint}
      />

      {/* Delivery Log */}
      <WebhookDeliveryHistory
        deliveries={deliveries}
        onRetryDelivery={handleRetryDelivery}
      />

      {/* Create Modal */}
      <CreateWebhookModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleAddEndpoint}
      />
    </div>
  );
}

export default WebhooksPage;
