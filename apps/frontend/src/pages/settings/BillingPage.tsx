import React, { useState } from 'react';
import { CreditCard, Sparkles, Shield, Check } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import {
  CurrentPlanCard,
  SubscriptionPlan,
} from '@/components/billing/CurrentPlanCard';
import {
  UsageQuotaProgressBar,
  QuotaItem,
} from '@/components/billing/UsageQuotaProgressBar';
import { InvoicesList, InvoiceItem } from '@/components/billing/InvoicesList';

const INITIAL_PLAN: SubscriptionPlan = {
  tier: 'pro',
  name: 'Pro Tier',
  priceMonthly: 49,
  status: 'active',
  renewalDate: '2026-09-15T00:00:00Z',
  stripePortalUrl: 'https://billing.stripe.com/p/session/test_session',
};

const INITIAL_QUOTAS: QuotaItem[] = [
  {
    label: 'Monthly Redirect Clicks',
    current: 142500,
    max: 500000,
    unit: 'clicks',
  },
  {
    label: 'Custom Branded Domains',
    current: 3,
    max: 5,
    unit: 'domains',
  },
  {
    label: 'Active Team Seats',
    current: 3,
    max: 10,
    unit: 'seats',
  },
];

const INITIAL_INVOICES: InvoiceItem[] = [
  {
    id: 'inv_flux_aug2026',
    date: '2026-08-15',
    amount: 49.0,
    status: 'paid',
    pdfUrl: '#',
  },
  {
    id: 'inv_flux_jul2026',
    date: '2026-07-15',
    amount: 49.0,
    status: 'paid',
    pdfUrl: '#',
  },
  {
    id: 'inv_flux_jun2026',
    date: '2026-06-15',
    amount: 49.0,
    status: 'paid',
    pdfUrl: '#',
  },
];

export function BillingPage() {
  const [plan, setPlan] = useState<SubscriptionPlan>(INITIAL_PLAN);
  const [quotas, setQuotas] = useState<QuotaItem[]>(INITIAL_QUOTAS);
  const [invoices, setInvoices] = useState<InvoiceItem[]>(INITIAL_INVOICES);
  const [notice, setNotice] = useState<string | null>(null);

  const handleUpgrade = () => {
    setNotice('Initiating Stripe Checkout upgrade to Enterprise Tier...');
    setTimeout(() => {
      setPlan((prev) => ({
        ...prev,
        tier: 'enterprise',
        name: 'Enterprise Tier',
        priceMonthly: 199,
      }));
      setNotice('Successfully upgraded to Enterprise Tier!');
      setTimeout(() => setNotice(null), 3000);
    }, 1000);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              Billing &amp; Subscriptions
            </h1>
            <Badge variant="emerald" size="sm" dot>
              Stripe Connected
            </Badge>
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            View monthly usage quotas, manage payment methods, and download past invoices.
          </p>
        </div>
      </div>

      {notice && (
        <div className="flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-semibold text-emerald-800 dark:border-emerald-900/50 dark:bg-emerald-950/30 dark:text-emerald-300 animate-in fade-in">
          <Check className="h-4 w-4" />
          <span>{notice}</span>
        </div>
      )}

      {/* Current Plan Overview */}
      <CurrentPlanCard plan={plan} onUpgrade={handleUpgrade} />

      {/* Resource Quotas Progress Meter */}
      <UsageQuotaProgressBar quotas={quotas} />

      {/* Stripe Invoices List */}
      <InvoicesList invoices={invoices} />
    </div>
  );
}

export default BillingPage;
