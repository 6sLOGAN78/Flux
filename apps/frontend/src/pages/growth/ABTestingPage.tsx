import React, { useState } from 'react';
import { Sliders, Plus, Trophy, Sparkles, Check, Layers } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import {
  VariantAllocationSlider,
  ABVariant,
} from '@/components/abtest/VariantAllocationSlider';
import {
  SignificanceScoreCard,
  VariantStats,
} from '@/components/abtest/SignificanceScoreCard';

const INITIAL_VARIANTS: ABVariant[] = [
  {
    id: 'var_a',
    name: 'Variant A (Control)',
    destinationUrl: 'https://flux.to/landing-v1',
    weight: 50,
  },
  {
    id: 'var_b',
    name: 'Variant B (New Pricing Hero)',
    destinationUrl: 'https://flux.to/landing-v2-pricing',
    weight: 50,
  },
];

export function ABTestingPage() {
  const [variants, setVariants] = useState<ABVariant[]>(INITIAL_VARIANTS);
  const [isPromoting, setIsPromoting] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const controlStats: VariantStats = {
    name: variants[0]?.name || 'Variant A',
    visitors: 2450,
    conversions: 122, // 4.98%
  };

  const challengerStats: VariantStats = {
    name: variants[1]?.name || 'Variant B',
    visitors: 2480,
    conversions: 198, // 7.98%
  };

  const handlePromoteWinner = () => {
    setIsPromoting(true);
    setTimeout(() => {
      // Set Variant B to 100%, others to 0%
      setVariants((prev) =>
        prev.map((v, idx) => ({
          ...v,
          weight: idx === 1 ? 100 : 0,
        }))
      );
      setIsPromoting(false);
      setSuccessMessage('Variant B promoted to 100% traffic successfully!');
      setTimeout(() => setSuccessMessage(null), 3000);
    }, 600);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              A/B Testing & Traffic Splitter
            </h1>
            <Badge variant="emerald" size="sm" dot>
              Live Experiment
            </Badge>
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Split traffic at the edge with zero client-side flicker and real-time statistical inference.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="md"
            leftIcon={<Plus className="h-4 w-4" />}
          >
            New Experiment
          </Button>
        </div>
      </div>

      {successMessage && (
        <div className="flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-semibold text-emerald-800 dark:border-emerald-900/50 dark:bg-emerald-950/30 dark:text-emerald-300 animate-in fade-in">
          <Check className="h-4 w-4" />
          <span>{successMessage}</span>
        </div>
      )}

      {/* Traffic Splitter Slider */}
      <VariantAllocationSlider
        variants={variants}
        onChangeVariants={setVariants}
      />

      {/* Statistical Significance Scorecard */}
      <SignificanceScoreCard
        controlVariant={controlStats}
        challengerVariant={challengerStats}
        onPromoteWinner={handlePromoteWinner}
        isLoading={isPromoting}
      />
    </div>
  );
}

export default ABTestingPage;
