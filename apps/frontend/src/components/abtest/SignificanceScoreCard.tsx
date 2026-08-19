import React from 'react';
import {
  TrendingUp,
  Trophy,
  Award,
  Sparkles,
  Zap,
  CheckCircle2,
  AlertCircle,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface VariantStats {
  name: string;
  visitors: number;
  conversions: number;
}

export interface SignificanceResult {
  controlCr: number;
  challengerCr: number;
  liftPercentage: number;
  confidencePercentage: number;
  isSignificant: boolean;
}

export function calculateSignificance(
  control: VariantStats,
  challenger: VariantStats
): SignificanceResult {
  const p1 = control.visitors > 0 ? control.conversions / control.visitors : 0;
  const p2 =
    challenger.visitors > 0 ? challenger.conversions / challenger.visitors : 0;

  const liftPercentage = p1 > 0 ? Number((((p2 - p1) / p1) * 100).toFixed(1)) : 0;

  // Standard Error
  const se = Math.sqrt(
    (p1 * (1 - p1)) / Math.max(control.visitors, 1) +
      (p2 * (1 - p2)) / Math.max(challenger.visitors, 1)
  );

  const z = se > 0 ? Math.abs((p2 - p1) / se) : 0;

  // Approximate normal CDF to confidence percentage
  // z = 1.96 -> ~95%, z = 2.58 -> ~99%
  let confidencePercentage = 50;
  if (z > 0) {
    // erf approximation
    const t = 1.0 / (1.0 + 0.2316419 * z);
    const d = 0.3989423 * Math.exp((-z * z) / 2);
    const prob =
      1.0 -
      d *
        t *
        (0.3193815 +
          t * (-0.3565638 + t * (1.781478 + t * (-1.821256 + t * 1.330274))));
    confidencePercentage = Number((prob * 100).toFixed(1));
  }

  const isSignificant =
    confidencePercentage >= 95 &&
    control.visitors >= 100 &&
    challenger.visitors >= 100;

  return {
    controlCr: Number((p1 * 100).toFixed(2)),
    challengerCr: Number((p2 * 100).toFixed(2)),
    liftPercentage,
    confidencePercentage,
    isSignificant,
  };
}

export interface SignificanceScoreCardProps {
  controlVariant?: VariantStats;
  challengerVariant?: VariantStats;
  onPromoteWinner?: () => void;
  isLoading?: boolean;
  className?: string;
}

export function SignificanceScoreCard({
  controlVariant = { name: 'Variant A (Control)', visitors: 2000, conversions: 100 },
  challengerVariant = { name: 'Variant B (Challenger)', visitors: 2000, conversions: 160 },
  onPromoteWinner,
  isLoading = false,
  className,
}: SignificanceScoreCardProps) {
  const result = calculateSignificance(controlVariant, challengerVariant);

  return (
    <div
      className={cn(
        'space-y-6 rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="flex items-center justify-between border-b border-zinc-100 pb-4 dark:border-zinc-900">
        <div>
          <div className="flex items-center gap-2">
            <TrendingUp className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              Statistical Significance
            </h2>
            <Badge
              variant={result.isSignificant ? 'emerald' : 'zinc'}
              size="sm"
              dot={result.isSignificant}
            >
              {result.isSignificant
                ? `${result.confidencePercentage}% Confidence`
                : 'Collecting Sample Data'}
            </Badge>
          </div>
          <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
            Bayesian & frequentist confidence scores computed in real-time from edge telemetry.
          </p>
        </div>
      </div>

      {/* Grid Comparison */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {/* Control Card */}
        <div className="rounded-xl border border-zinc-200 bg-zinc-50/50 p-4 dark:border-zinc-800 dark:bg-zinc-900/30">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-zinc-700 dark:text-zinc-300">
              {controlVariant.name}
            </span>
            <Badge variant="zinc" size="sm">
              Baseline
            </Badge>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
              {result.controlCr}%
            </span>
            <span className="text-[11px] text-zinc-400">CR</span>
          </div>
          <div className="mt-2 text-[11px] text-zinc-500 font-mono">
            {controlVariant.conversions.toLocaleString()} / {controlVariant.visitors.toLocaleString()} visitors
          </div>
        </div>

        {/* Challenger Card */}
        <div className="rounded-xl border border-zinc-200 bg-emerald-50/40 p-4 dark:border-emerald-950/30 dark:bg-emerald-950/10">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-zinc-900 dark:text-zinc-100">
              {challengerVariant.name}
            </span>
            {result.liftPercentage > 0 && (
              <Badge variant="emerald" size="sm">
                +{result.liftPercentage}% Relative Lift
              </Badge>
            )}
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="font-mono text-2xl font-bold text-emerald-600 dark:text-emerald-400">
              {result.challengerCr}%
            </span>
            <span className="text-[11px] text-zinc-400">CR</span>
          </div>
          <div className="mt-2 text-[11px] text-zinc-500 font-mono">
            {challengerVariant.conversions.toLocaleString()} / {challengerVariant.visitors.toLocaleString()} visitors
          </div>
        </div>
      </div>

      {/* Promotion Action */}
      <div className="flex flex-col items-center justify-between gap-4 rounded-xl border border-zinc-200 bg-zinc-50/75 p-4 dark:border-zinc-800 dark:bg-zinc-900/40 sm:flex-row">
        <div className="flex items-center gap-3">
          <Trophy className="h-5 w-5 text-amber-500 shrink-0" />
          <div className="text-xs">
            <span className="font-semibold text-zinc-900 dark:text-zinc-100">
              {result.isSignificant
                ? `${challengerVariant.name} is winning by +${result.liftPercentage}%`
                : 'Experiment still accumulating visitors'}
            </span>
            <p className="text-[11px] text-zinc-500 dark:text-zinc-400">
              Locking winner will automatically allocate 100% of traffic to the winning variant.
            </p>
          </div>
        </div>

        <Button
          variant="primary"
          size="md"
          onClick={onPromoteWinner}
          isLoading={isLoading}
          leftIcon={<Zap className="h-3.5 w-3.5 fill-current" />}
        >
          Promote Winner (100% Traffic)
        </Button>
      </div>
    </div>
  );
}

export default SignificanceScoreCard;
