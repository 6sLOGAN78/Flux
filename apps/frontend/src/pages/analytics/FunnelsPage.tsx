import React, { useState } from 'react';
import { Filter, Plus, TrendingUp, DollarSign, Layers } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import {
  FunnelVisualizer,
  FunnelStepItem,
} from '@/components/funnels/FunnelVisualizer';
import {
  UnitEconomicsCards,
  UnitEconomicsData,
} from '@/components/revenue/UnitEconomicsCards';
import { CreateFunnelModal } from '@/components/funnels/CreateFunnelModal';

const INITIAL_STEPS: FunnelStepItem[] = [
  {
    id: 's1',
    name: '1. Ad Link Click',
    visitors: 10000,
    dropoffPercentage: 0,
    conversionRateFromStart: 100,
  },
  {
    id: 's2',
    name: '2. Landing Page View',
    visitors: 6500,
    dropoffPercentage: 35.0,
    conversionRateFromStart: 65.0,
  },
  {
    id: 's3',
    name: '3. Pricing Page Visit',
    visitors: 2600,
    dropoffPercentage: 60.0,
    conversionRateFromStart: 26.0,
  },
  {
    id: 's4',
    name: '4. Account Sign Up',
    visitors: 910,
    dropoffPercentage: 65.0,
    conversionRateFromStart: 9.1,
  },
  {
    id: 's5',
    name: '5. Paid Subscription',
    visitors: 431,
    dropoffPercentage: 52.6,
    conversionRateFromStart: 4.31,
  },
];

const INITIAL_ECONOMICS: UnitEconomicsData = {
  totalSpend: 24500,
  attributedRevenue: 119750,
  cac: 56.84,
  roas: 4.89,
  ltv: 240,
  ltvCacRatio: 4.22,
};

export function FunnelsPage() {
  const [steps, setSteps] = useState<FunnelStepItem[]>(INITIAL_STEPS);
  const [economics, setEconomics] =
    useState<UnitEconomicsData>(INITIAL_ECONOMICS);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [funnelTitle, setFunnelTitle] = useState(
    'SaaS Self-Serve Checkout Funnel'
  );

  const handleCreateFunnel = (data: { name: string; steps: string[] }) => {
    setFunnelTitle(data.name);
    const newSteps: FunnelStepItem[] = data.steps.map((name, idx) => {
      const visitors = Math.round(10000 / Math.pow(1.8, idx));
      const prevVisitors =
        idx === 0 ? visitors : Math.round(10000 / Math.pow(1.8, idx - 1));
      const dropoff =
        idx === 0 ? 0 : Number((((prevVisitors - visitors) / prevVisitors) * 100).toFixed(1));
      const cr = Number(((visitors / 10000) * 100).toFixed(1));

      return {
        id: `s_${idx + 1}`,
        name: `${idx + 1}. ${name}`,
        visitors,
        dropoffPercentage: dropoff,
        conversionRateFromStart: cr,
      };
    });

    setSteps(newSteps);
    setIsModalOpen(false);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              Conversion Funnels &amp; Unit Economics
            </h1>
            <Badge variant="emerald" size="sm" dot>
              ROAS Positive
            </Badge>
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            End-to-end attribution tracking from first ad impression down to lifetime value (LTV).
          </p>
        </div>

        <Button
          variant="primary"
          size="md"
          onClick={() => setIsModalOpen(true)}
          leftIcon={<Plus className="h-4 w-4" />}
        >
          New Funnel
        </Button>
      </div>

      {/* Unit Economics ROAS Dashboard */}
      <UnitEconomicsCards data={economics} />

      {/* Conversion Funnel Visualizer */}
      <FunnelVisualizer steps={steps} funnelName={funnelTitle} />

      {/* Create Funnel Modal */}
      <CreateFunnelModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleCreateFunnel}
      />
    </div>
  );
}

export default FunnelsPage;
