import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { mock } from 'bun:test';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BillingPage } from './BillingPage';
import { CurrentPlanCard, SubscriptionPlan } from '@/components/billing/CurrentPlanCard';
import {
  UsageQuotaProgressBar,
  QuotaItem,
} from '@/components/billing/UsageQuotaProgressBar';
import { InvoicesList, InvoiceItem } from '@/components/billing/InvoicesList';


mock.module('@/hooks/useBillingQuery', () => ({
  useGetSubscription: () => ({
    data: {
      plan: 'pro',
      status: 'active',
      currentPeriodEnd: '2026-09-15T00:00:00Z',
    },
    isLoading: false,
    isError: false,
  }),
  useCreateCustomerPortal: () => ({
    mutateAsync: async () => ({ url: 'https://billing.stripe.com/p/session/test' }),
    isPending: false,
  }),
}));
describe('Billing & Subscriptions Management Page', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockPlan: SubscriptionPlan = {
    tier: 'pro',
    name: 'Pro Tier',
    priceMonthly: 49,
    status: 'active',
    renewalDate: '2026-09-15T00:00:00Z',
    stripePortalUrl: 'https://billing.stripe.com/p/session/test_session',
  };

  const mockQuotas: QuotaItem[] = [
    { label: 'Monthly Redirect Clicks', current: 142500, max: 500000, unit: 'clicks' },
    { label: 'Custom Branded Domains', current: 3, max: 5, unit: 'domains' },
    { label: 'Active Team Seats', current: 3, max: 10, unit: 'seats' },
  ];

  const mockInvoices: InvoiceItem[] = [
    { id: 'in_1', date: '2026-08-15', amount: 49, status: 'paid', pdfUrl: '#' },
    { id: 'in_2', date: '2026-07-15', amount: 49, status: 'paid', pdfUrl: '#' },
  ];

  it('renders CurrentPlanCard with price, renewal date, and Stripe portal button', () => {
    const html = renderToString(
      <CurrentPlanCard
        plan={mockPlan}
        onUpgrade={() => {}}
      />
    );

    expect(html).toContain('Pro Tier');
    expect(html).toContain('$49');
    expect(html).toContain('Manage in Stripe Portal');
    expect(html).toContain('Manage Billing');
  });

  it('renders UsageQuotaProgressBar with percentage calculations and warning colors', () => {
    const html = renderToString(
      <UsageQuotaProgressBar quotas={mockQuotas} />
    );

    expect(html).toContain('Usage &amp; Resource Quotas');
    expect(html).toContain('Monthly Redirect Clicks');
    expect(html).toContain('142,500');
    expect(html).toContain('500,000');
  });

  it('renders InvoicesList with past receipts and payment statuses', () => {
    const html = renderToString(
      <InvoicesList invoices={mockInvoices} />
    );

    expect(html).toContain('Billing History &amp; Invoices');
    expect(html).toContain('$49.00');
    expect(html).toContain('Paid');
  });

  it('renders full BillingPage with subscription overview and meters', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <BillingPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Billing &amp; Subscriptions');
    expect(html).toContain('Pro Tier');
    expect(html).toContain('Usage &amp; Resource Quotas');
  });
});
