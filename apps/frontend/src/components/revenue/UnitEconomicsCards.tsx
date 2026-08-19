import React from 'react';
import { DollarSign, TrendingUp, Sparkles, Percent, ShieldCheck } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface UnitEconomicsData {
  totalSpend: number;
  attributedRevenue: number;
  cac: number;
  roas: number;
  ltv: number;
  ltvCacRatio: number;
}

export interface UnitEconomicsCardsProps {
  data: UnitEconomicsData;
  className?: string;
}

export function UnitEconomicsCards({
  data,
  className,
}: UnitEconomicsCardsProps) {
  const isLtvCacHealthy = data.ltvCacRatio >= 3.0;

  return (
    <div className={cn('grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4', className)}>
      {/* Spend & Attributed Revenue */}
      <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
        <div className="flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400">
          <span>Ad Spend &amp; Attributed Rev</span>
          <DollarSign className="h-4 w-4 text-zinc-400" />
        </div>
        <div className="mt-2 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
          {`$${data.attributedRevenue.toLocaleString()}`}
        </div>
        <div className="mt-1 text-[11px] text-zinc-400 font-mono">
          {`$${data.totalSpend.toLocaleString()} total ad spend`}
        </div>
      </div>

      {/* Customer Acquisition Cost (CAC) */}
      <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
        <div className="flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400">
          <span>Customer Acquisition Cost (CAC)</span>
          <TrendingUp className="h-4 w-4 text-zinc-400" />
        </div>
        <div className="mt-2 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
          {`$${data.cac.toFixed(2)}`}
        </div>
        <div className="mt-1 text-[11px] font-medium text-emerald-600 dark:text-emerald-400">
          -14.2% reduction this month
        </div>
      </div>

      {/* Return on Ad Spend (ROAS) */}
      <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
        <div className="flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400">
          <span>Return on Ad Spend (ROAS)</span>
          <Sparkles className="h-4 w-4 text-zinc-400" />
        </div>
        <div className="mt-2 font-mono text-2xl font-bold text-emerald-600 dark:text-emerald-400">
          {`${data.roas.toFixed(2)}x`}
        </div>
        <div className="mt-1 text-[11px] text-zinc-400 font-mono">
          Gross margin positive
        </div>
      </div>

      {/* LTV:CAC Ratio */}
      <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
        <div className="flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400">
          <span>LTV:CAC Ratio</span>
          <ShieldCheck className="h-4 w-4 text-zinc-400" />
        </div>
        <div className="mt-2 flex items-baseline gap-2">
          <span className="font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            {`${data.ltvCacRatio.toFixed(2)}x`}
          </span>
          <Badge
            variant={isLtvCacHealthy ? 'emerald' : 'amber'}
            size="sm"
            dot
          >
            {isLtvCacHealthy ? 'Healthy (>3.0x)' : 'Attention Needed'}
          </Badge>
        </div>
        <div className="mt-1 text-[11px] text-zinc-400 font-mono">
          {`LTV: $${data.ltv.toFixed(2)} / customer`}
        </div>
      </div>
    </div>
  );
}

export default UnitEconomicsCards;
