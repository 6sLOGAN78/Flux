import React, { useState } from 'react';
import { Sparkles, Plus, Layers, Filter } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Tabs } from '@/components/ui/Tabs';
import { UTMBuilderStudio } from '@/components/campaigns/UTMBuilderStudio';
import {
  CampaignListTable,
  CampaignItem,
} from '@/components/campaigns/CampaignListTable';
import { useCreateLink } from '@/hooks/useLinksQuery';

// Empty state for new accounts
const INITIAL_CAMPAIGNS: CampaignItem[] = [];

export function CampaignsPage() {
  const [campaigns, setCampaigns] = useState<CampaignItem[]>(INITIAL_CAMPAIGNS);
  const [activeTab, setActiveTab] = useState('campaigns');

  const createLinkMutation = useCreateLink();

  const handleGenerateLink = (data: {
    finalUrl: string;
    campaignName: string;
    customSlug?: string;
  }) => {
    createLinkMutation.mutate(
      {
        destinationUrl: data.finalUrl,
        customCode: data.customSlug,
        title: data.campaignName,
      },
      {
        onSettled: () => {
          const newCampaign: CampaignItem = {
            id: `cmp_${Date.now()}`,
            name: data.campaignName,
            channel: 'Multi-Channel',
            utmCampaign: data.customSlug || 'custom_campaign',
            totalClicks: 0,
            conversions: 0,
            status: 'active',
            createdAt: new Date().toISOString(),
          };
          setCampaigns((prev) => [newCampaign, ...prev]);
          setActiveTab('campaigns');
        },
      }
    );
  };

  const tabs = [
    { id: 'campaigns', label: 'Active Campaigns', count: campaigns.length },
    { id: 'builder', label: 'Visual UTM Builder' },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
            Marketing Campaigns
          </h1>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Multi-channel UTM tracking, conversion attribution, and link tagging.
          </p>
        </div>

        <Button
          variant="primary"
          size="md"
          onClick={() => setActiveTab('builder')}
          leftIcon={<Plus className="h-4 w-4" />}
        >
          New Campaign
        </Button>
      </div>

      {/* Tabs */}
      <Tabs
        tabs={tabs}
        activeTab={activeTab}
        onChange={setActiveTab}
        variant="underline"
      />

      {/* Content */}
      {activeTab === 'campaigns' ? (
        <CampaignListTable campaigns={campaigns} />
      ) : (
        <UTMBuilderStudio
          onGenerateLink={handleGenerateLink}
          isLoading={createLinkMutation.isPending}
        />
      )}
    </div>
  );
}

export default CampaignsPage;
