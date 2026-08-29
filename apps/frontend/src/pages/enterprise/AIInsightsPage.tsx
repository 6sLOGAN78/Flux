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

// Empty state
const MOCK_CHART_DATA: CTRPredictionPoint[] = [];

const INITIAL_ANOMALIES: AnomalyEvent[] = [];

const INITIAL_TIPS: OptimizationTip[] = [];

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
