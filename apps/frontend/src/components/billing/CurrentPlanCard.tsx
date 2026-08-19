import React from 'react';
import { CreditCard, Sparkles, ExternalLink, ArrowRight, ShieldCheck } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface SubscriptionPlan {
  tier: 'free' | 'pro' | 'enterprise';
  name: string;
  priceMonthly: number;
  status: 'active' | 'past_due' | 'canceled';
  renewalDate: string;
  stripePortalUrl?: string;
}

export interface CurrentPlanCardProps {
  plan: SubscriptionPlan;
  onUpgrade: () => void;
  isLoading?: boolean;
  className?: string;
}

export function CurrentPlanCard({
  plan,
  onUpgrade,
  isLoading = false,
  className,
}: CurrentPlanCardProps) {
  const handleOpenStripePortal = () => {
    if (plan.stripePortalUrl) {
      window.open(plan.stripePortalUrl, '_blank', 'noreferrer');
    }
  };

  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <CreditCard className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              Subscription Plan
            </h2>
            <Badge variant="emerald" size="sm" dot>
              {plan.status === 'active' ? 'Active' : plan.status}
            </Badge>
          </div>
          <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
            Automated metered billing managed securely via Stripe Billing.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleOpenStripePortal}
            rightIcon={<ExternalLink className="h-3.5 w-3.5" />}
          >
            Manage in Stripe Portal
          </Button>
          <Button
            variant="primary"
            size="sm"
            onClick={onUpgrade}
            leftIcon={<Sparkles className="h-3.5 w-3.5" />}
          >
            Upgrade Plan
          </Button>
        </div>
      </div>

      <div className="mt-6 flex flex-col justify-between gap-4 rounded-xl border border-zinc-100 bg-zinc-50/50 p-4 dark:border-zinc-900 dark:bg-zinc-900/30 sm:flex-row sm:items-center">
        <div>
          <div className="text-xs font-semibold text-zinc-500 uppercase tracking-wider">
            Current Tier
          </div>
          <div className="mt-1 flex items-baseline gap-1.5">
            <span className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
              {plan.name}
            </span>
            <span className="font-mono text-xs text-zinc-500">
              {`$${plan.priceMonthly}/month`}
            </span>
          </div>
        </div>

        <div className="text-left sm:text-right">
          <div className="text-xs text-zinc-400">Next Billing Renewal</div>
          <div className="mt-1 font-mono text-xs font-semibold text-zinc-900 dark:text-zinc-100">
            {new Date(plan.renewalDate).toLocaleDateString(undefined, {
              month: 'long',
              day: 'numeric',
              year: 'numeric',
            })}
          </div>
        </div>
      </div>
    </div>
  );
}

export default CurrentPlanCard;
