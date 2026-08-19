import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { IntegrationsPage } from './IntegrationsPage';
import { NotificationsPage } from '@/pages/dashboard/NotificationsPage';
import {
  AppConnectorCard,
  IntegrationApp,
} from '@/components/integrations/AppConnectorCard';
import {
  NotificationDrawer,
  NotificationItem,
} from '@/components/notifications/NotificationDrawer';

describe('Integrations Directory & In-App Notification Center', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockApps: IntegrationApp[] = [
    {
      id: 'slack',
      name: 'Slack',
      category: 'chat',
      description: 'Receive real-time click spikes and daily summaries in Slack channels.',
      status: 'connected',
    },
    {
      id: 'zapier',
      name: 'Zapier',
      category: 'automation',
      description: 'Connect Flux events with 5,000+ business applications.',
      status: 'not_connected',
    },
  ];

  const mockNotifications: NotificationItem[] = [
    {
      id: 'notif_1',
      title: 'Traffic Spike Alert',
      message: 'Short link /summer-sale reached 10,000 clicks in the past 1 hour.',
      severity: 'alert',
      isRead: false,
      createdAt: '2026-08-19T22:30:00Z',
    },
    {
      id: 'notif_2',
      title: 'Custom Domain SSL Active',
      message: 'Certificate successfully provisioned for go.brand.com.',
      severity: 'info',
      isRead: true,
      createdAt: '2026-08-18T12:00:00Z',
    },
  ];

  it('renders AppConnectorCard with status badge and configure action', () => {
    const html = renderToString(
      <AppConnectorCard
        app={mockApps[0]}
        onToggleConnect={() => {}}
      />
    );

    expect(html).toContain('Slack');
    expect(html).toContain('Connected');
    expect(html).toContain('Configure');
  });

  it('renders NotificationDrawer with severity badges and mark all read button', () => {
    const html = renderToString(
      <NotificationDrawer
        isOpen={true}
        onClose={() => {}}
        notifications={mockNotifications}
        onMarkAllAsRead={() => {}}
        onMarkAsRead={() => {}}
      />
    );

    expect(html).toContain('Notifications');
    expect(html).toContain('Traffic Spike Alert');
    expect(html).toContain('Mark all as read');
  });

  it('renders full IntegrationsPage with category tabs and app directory', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <IntegrationsPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Integrations Directory');
    expect(html).toContain('Slack');
    expect(html).toContain('Zapier');
  });

  it('renders full NotificationsPage with notifications feed and unread counter', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <NotificationsPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Notification Center');
    expect(html).toContain('Traffic Spike Alert');
  });
});
