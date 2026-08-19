import React, { useState, useMemo } from 'react';
import {
  Layers,
  Sparkles,
  DollarSign,
  TrendingUp,
  GitMerge,
  PieChart,
} from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import {
  ModelSelectorBar,
  AttributionModelType,
} from '@/components/attribution/ModelSelectorBar';
import {
  TouchpointTimelineFlow,
  TouchpointNode,
} from '@/components/attribution/TouchpointTimelineFlow';
import {
  AttributionComparisonTable,
  CampaignAttributionItem,
} from '@/components/attribution/AttributionComparisonTable';

const BASE_CAMPAIGNS = [
  {
    channel: 'Twitter Paid Ads',
    campaign: 'dev_growth_q3',
    touchpoints: 4120,
    rawConversions: 431,
  },
  {
    channel: 'Google Search CPC',
    campaign: 'brand_keywords',
    touchpoints: 3890,
    rawConversions: 431,
  },
  {
    channel: 'Product Hunt Launch',
    campaign: 'ph_v2_launch',
    touchpoints: 2100,
    rawConversions: 431,
  },
  {
    channel: 'Newsletter & Blog',
    campaign: 'august_digest',
    touchpoints: 1450,
    rawConversions: 431,
  },
];

export function AttributionPage() {
  const [selectedModel, setSelectedModel] =
    useState<AttributionModelType>('u_shaped');

  const campaignData: CampaignAttributionItem[] = useMemo(() => {
    // Model weight distributions
    const weights: Record<AttributionModelType, number[]> = {
      first_touch: [0.55, 0.25, 0.15, 0.05],
      last_touch: [0.2, 0.5, 0.15, 0.15],
      linear: [0.3, 0.3, 0.25, 0.15],
      time_decay: [0.22, 0.45, 0.18, 0.15],
      u_shaped: [0.385, 0.319, 0.196, 0.1],
    };

    const currentWeights = weights[selectedModel];
    const totalPipeline = 119750;
    const totalConversions = 431.0;

    return BASE_CAMPAIGNS.map((c, idx) => {
      const share = currentWeights[idx];
      const revenue = Math.round(totalPipeline * share);
      const conversions = Number((totalConversions * share).toFixed(1));
      const sharePercentage = Number((share * 100).toFixed(1));

      return {
        channel: c.channel,
        campaign: c.campaign,
        touchpoints: c.touchpoints,
        conversions,
        revenue,
        sharePercentage,
      };
    });
  }, [selectedModel]);

  const timelineNodes: TouchpointNode[] = useMemo(() => {
    switch (selectedModel) {
      case 'first_touch':
        return [
          { channel: 'Twitter / X', type: 'first_touch', timestamp: 'Aug 10', weightPercentage: 100 },
          { channel: 'Google Search', type: 'middle_touch', timestamp: 'Aug 14', weightPercentage: 0 },
          { channel: 'Blog Post', type: 'middle_touch', timestamp: 'Aug 17', weightPercentage: 0 },
          { channel: 'Direct / Pricing', type: 'last_touch', timestamp: 'Aug 19', weightPercentage: 0 },
        ];
      case 'last_touch':
        return [
          { channel: 'Twitter / X', type: 'first_touch', timestamp: 'Aug 10', weightPercentage: 0 },
          { channel: 'Google Search', type: 'middle_touch', timestamp: 'Aug 14', weightPercentage: 0 },
          { channel: 'Blog Post', type: 'middle_touch', timestamp: 'Aug 17', weightPercentage: 0 },
          { channel: 'Direct / Pricing', type: 'last_touch', timestamp: 'Aug 19', weightPercentage: 100 },
        ];
      case 'linear':
        return [
          { channel: 'Twitter / X', type: 'first_touch', timestamp: 'Aug 10', weightPercentage: 25 },
          { channel: 'Google Search', type: 'middle_touch', timestamp: 'Aug 14', weightPercentage: 25 },
          { channel: 'Blog Post', type: 'middle_touch', timestamp: 'Aug 17', weightPercentage: 25 },
          { channel: 'Direct / Pricing', type: 'last_touch', timestamp: 'Aug 19', weightPercentage: 25 },
        ];
      case 'time_decay':
        return [
          { channel: 'Twitter / X', type: 'first_touch', timestamp: 'Aug 10', weightPercentage: 10 },
          { channel: 'Google Search', type: 'middle_touch', timestamp: 'Aug 14', weightPercentage: 20 },
          { channel: 'Blog Post', type: 'middle_touch', timestamp: 'Aug 17', weightPercentage: 30 },
          { channel: 'Direct / Pricing', type: 'last_touch', timestamp: 'Aug 19', weightPercentage: 40 },
        ];
      case 'u_shaped':
      default:
        return [
          { channel: 'Twitter / X', type: 'first_touch', timestamp: 'Aug 10', weightPercentage: 40 },
          { channel: 'Google Search', type: 'middle_touch', timestamp: 'Aug 14', weightPercentage: 10 },
          { channel: 'Blog Post', type: 'middle_touch', timestamp: 'Aug 17', weightPercentage: 10 },
          { channel: 'Direct / Pricing', type: 'last_touch', timestamp: 'Aug 19', weightPercentage: 40 },
        ];
    }
  }, [selectedModel]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              Multi-Touch Attribution Studio
            </h1>
            <Badge variant="blue" size="sm" dot>
              Enterprise
            </Badge>
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Compare algorithmic conversion models and trace full customer conversion journeys.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Badge variant="zinc" size="sm" className="font-mono">
            Model: Position-Based (U-Shaped)
          </Badge>
        </div>
      </div>

      {/* KPI Stats */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="text-xs text-zinc-500 dark:text-zinc-400">
            Attributed Pipeline Revenue
          </div>
          <div className="mt-2 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            $119,750
          </div>
          <div className="mt-1 text-[11px] font-medium text-emerald-600 dark:text-emerald-400">
            +24.8% pipeline growth
          </div>
        </div>

        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="text-xs text-zinc-500 dark:text-zinc-400">
            Attributed Conversions
          </div>
          <div className="mt-2 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            431.0
          </div>
          <div className="mt-1 text-[11px] text-zinc-400">
            Across 11,560 touchpoints
          </div>
        </div>

        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="text-xs text-zinc-500 dark:text-zinc-400">
            Avg Touchpoints / Journey
          </div>
          <div className="mt-2 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            3.8
          </div>
          <div className="mt-1 text-[11px] text-zinc-400">
            Multi-channel interaction path
          </div>
        </div>

        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="text-xs text-zinc-500 dark:text-zinc-400">
            Top Acquisition Channel
          </div>
          <div className="mt-2 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100 truncate">
            Twitter / X
          </div>
          <div className="mt-1 text-[11px] text-zinc-400">
            38.5% total pipeline share
          </div>
        </div>
      </div>

      {/* Model Selector Bar */}
      <ModelSelectorBar
        selectedModel={selectedModel}
        onSelectModel={setSelectedModel}
      />

      {/* Touchpoint Timeline Flow */}
      <TouchpointTimelineFlow
        touchpoints={timelineNodes}
        model={selectedModel}
      />

      {/* Attribution Performance Table */}
      <AttributionComparisonTable data={campaignData} />
    </div>
  );
}

export default AttributionPage;
