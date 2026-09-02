import React, { useState } from 'react';
import { Sparkles, Check, AlertCircle } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { CurrentPlanCard, SubscriptionPlan } from '@/components/billing/CurrentPlanCard';
import { UsageQuotaProgressBar, QuotaItem } from '@/components/billing/UsageQuotaProgressBar';
import { InvoicesList, InvoiceItem } from '@/components/billing/InvoicesList';
import { useGetSubscription, useCreateCustomerPortal } from '@/hooks/useBillingQuery';

export function BillingPage() {
  const { data: subscription, isLoading, isError, error } = useGetSubscription();
  const portalMutation = useCreateCustomerPortal();
  
  const [notice, setNotice] = useState<string | null>(null);
  const [errorNotice, setErrorNotice] = useState<string | null>(null);

  const handleManageBilling = async () => {
    try {
      setNotice('Redirecting to Stripe Customer Portal...');
      setErrorNotice(null);
      const res = await portalMutation.mutateAsync();
      if (res.url) {
        window.location.href = res.url;
      }
    } catch (e: any) {
      setNotice(null);
      setErrorNotice(e.message || 'Failed to open billing portal');
    }
  };

  if (isLoading) {
    return <div className="p-8 text-center text-zinc-500">Loading billing information...</div>;
  }

  if (isError) {
    return (
      <div className="p-8 text-center text-red-500">
        Error loading billing info: {error?.message}
      </div>
    );
  }

  // Fallback map backend state to frontend CurrentPlanCard format
  const plan: SubscriptionPlan = {
    tier: subscription?.plan as any || 'free',
    name: subscription?.plan === 'business' ? 'Business Tier' : subscription?.plan === 'pro' ? 'Pro Tier' : 'Free Tier',
    priceMonthly: subscription?.plan === 'business' ? 99 : subscription?.plan === 'pro' ? 29 : 0,
    status: subscription?.status as any || 'active',
    renewalDate: subscription?.currentPeriodEnd || new Date().toISOString(),
  };

  // Limits mapped according to backend rules (or we could fetch them explicitly, but for now we hardcode the display rules or leave the UsageQuota blank since we don't fetch active usages in this task).
  const quotas: QuotaItem[] = [
    {
       
      label: 'Short Links',
      current: 0, // Placeholder, since we just have the max
      max: subscription?.maxLinks || 1000,
      unit: 'links',
    },
    {
       
      label: 'Analytics Retention',
      current: subscription?.analyticsRetention || 7,
      max: subscription?.analyticsRetention || 7,
      unit: 'days',
    }
  ];

  const invoices: InvoiceItem[] = []; // Placeholder

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
      
      {errorNotice && (
        <div className="flex items-center gap-2 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-xs font-semibold text-red-800 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-300 animate-in fade-in">
          <AlertCircle className="h-4 w-4" />
          <span>{errorNotice}</span>
        </div>
      )}

      {/* Current Plan Overview */}
      <CurrentPlanCard 
        plan={plan} 
        onUpgrade={handleManageBilling} 
        isLoading={portalMutation.isPending} 
      />

      {/* Resource Quotas Progress Meter */}
      <UsageQuotaProgressBar quotas={quotas} />

      {/* Stripe Invoices List */}
      <InvoicesList invoices={invoices} />
    </div>
  );
}

export default BillingPage;
