import React, { useState } from 'react';
import {
  Sparkles,
  Cpu,
  ShieldCheck,
  TrendingUp,
  Activity,
  Check,
} from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import {
  CTRPredictionChart,
  CTRPredictionPoint,
} from '@/components/ai/CTRPredictionChart';
import {
  AnomalyEventStream,
  AnomalyEvent,
} from '@/components/ai/AnomalyEventStream';
import {
  OptimizationTipsCard,
  OptimizationTip,
} from '@/components/ai/OptimizationTipsCard';

const MOCK_CHART_DATA: CTRPredictionPoint[] = [
  { hour: '00:00', actual: 3.2, predicted: 3.1 },
  { hour: '04:00', actual: 2.8, predicted: 2.9 },
  { hour: '08:00', actual: 4.5, predicted: 4.3 },
  { hour: '12:00', actual: 6.2, predicted: 6.0 },
  { hour: '16:00', actual: 5.8, predicted: 5.9 },
  { hour: '20:00', actual: null, predicted: 6.8 },
  { hour: '23:59', actual: null, predicted: 7.4 },
];

const INITIAL_ANOMALIES: AnomalyEvent[] = [
  {
    id: 'anom_1',
    type: 'traffic_spike',
    slug: 'summer-sale',
    zScore: 3.84,
    description: 'Sudden +480% referral surge from Hacker News frontpage.',
    timestamp: '2026-08-19T22:40:00Z',
  },
  {
    id: 'anom_2',
    type: 'bot_surge',
    slug: 'checkout-v2',
    zScore: 4.12,
    description: 'Anomalous automated scraper crawler detected from AS15169.',
    timestamp: '2026-08-19T22:35:00Z',
  },
  {
    id: 'anom_3',
    type: 'traffic_drop',
    slug: 'pricing-matrix',
    zScore: -2.94,
    description: 'Traffic dropped 65% below seasonal moving average expectation.',
    timestamp: '2026-08-19T21:15:00Z',
  },
];

const INITIAL_TIPS: OptimizationTip[] = [
  {
    id: 'tip_1',
    title: 'Optimal Distribution Window',
    description: 'Publishing Twitter/X short links at 14:00 UTC yields 2.4x higher conversion rate than average.',
    impact: 'high',
    actionLabel: 'Schedule Campaign',
  },
  {
    id: 'tip_2',
    title: 'Bot Protection Threshold',
    description: 'Enable Cloudflare Turnstile bot shield on /checkout-v2 to preserve downstream ad spend.',
    impact: 'high',
    actionLabel: 'Enable Bot Shield',
  },
  {
    id: 'tip_3',
    title: 'A/B Testing Variant Divergence',
    description: 'Variant B on smart routing rule #4 has reached 98.2% statistical confidence over Variant A.',
    impact: 'medium',
    actionLabel: 'Promote Variant',
  },
];

export function AIInsightsPage() {
  const [anomalies, setAnomalies] =
    useState<AnomalyEvent[]>(INITIAL_ANOMALIES);
  const [tips, setTips] = useState<OptimizationTip[]>(INITIAL_TIPS);
  const [notice, setNotice] = useState<string | null>(null);

  const handleResolveAnomaly = (id: string) => {
    setAnomalies((prev) => prev.filter((a) => a.id !== id));
    setNotice('Anomaly marked as resolved.');
    setTimeout(() => setNotice(null), 3000);
  };

  const handleTipAction = (id: string) => {
    const tip = tips.find((t) => t.id === id);
    if (!tip) return;
    setNotice(`Applied recommendation: ${tip.title}`);
    setTimeout(() => setNotice(null), 3000);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              Predictive AI Insights
            </h1>
            <Badge variant="blue" size="sm" dot>
              AI Engine Online
            </Badge>
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Transformer-driven CTR forecasting and real-time statistical anomaly monitoring.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Badge variant="zinc" size="sm" className="font-mono">
            Model: Flux-Transformer-v3
          </Badge>
        </div>
      </div>

      {notice && (
        <div className="flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-semibold text-emerald-800 dark:border-emerald-900/50 dark:bg-emerald-950/30 dark:text-emerald-300 animate-in fade-in">
          <Check className="h-4 w-4" />
          <span>{notice}</span>
        </div>
      )}

      {/* KPI Cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="text-xs text-zinc-500 dark:text-zinc-400">
            Forecasted CTR Growth
          </div>
          <div className="mt-2 font-mono text-2xl font-bold text-emerald-600 dark:text-emerald-400">
            +34.2%
          </div>
          <div className="mt-1 text-[11px] text-zinc-400">
            Next 7-day projected trajectory
          </div>
        </div>

        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="text-xs text-zinc-500 dark:text-zinc-400">
            Active Anomaly Outliers
          </div>
          <div className="mt-2 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            {anomalies.length}
          </div>
          <div className="mt-1 text-[11px] text-zinc-400">
            Statistical |Z| &gt; 2.58
          </div>
        </div>

        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="text-xs text-zinc-500 dark:text-zinc-400">
            Bot Filter Accuracy
          </div>
          <div className="mt-2 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            99.4%
          </div>
          <div className="mt-1 text-[11px] text-emerald-600 dark:text-emerald-400">
            Edge-level fingerprinting
          </div>
        </div>

        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="text-xs text-zinc-500 dark:text-zinc-400">
            AI Model Confidence
          </div>
          <div className="mt-2 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            96.8%
          </div>
          <div className="mt-1 text-[11px] text-zinc-400">
            Continuous Bayesian calibration
          </div>
        </div>
      </div>

      {/* Main CTR Prediction Chart */}
      <CTRPredictionChart data={MOCK_CHART_DATA} />

      {/* Anomaly Stream and Optimization Tips */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <AnomalyEventStream
          anomalies={anomalies}
          onResolveAnomaly={handleResolveAnomaly}
        />
        <OptimizationTipsCard tips={tips} onActionClick={handleTipAction} />
      </div>
    </div>
  );
}

export default AIInsightsPage;
