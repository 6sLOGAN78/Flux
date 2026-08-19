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

const INITIAL_CAMPAIGNS: CampaignItem[] = [
  {
    id: 'cmp_1',
    name: 'Q3 Product Hunt Launch',
    channel: 'ProductHunt',
    utmCampaign: 'ph_launch_v2',
    totalClicks: 8420,
    conversions: 612,
    status: 'active',
    createdAt: '2026-08-18T10:00:00Z',
  },
  {
    id: 'cmp_2',
    name: 'Twitter Developer Ads',
    channel: 'Twitter',
    utmCampaign: 'dev_ads_q3',
    totalClicks: 3290,
    conversions: 184,
    status: 'active',
    createdAt: '2026-08-16T14:00:00Z',
  },
  {
    id: 'cmp_3',
    name: 'Monthly Newsletter - Issue #42',
    channel: 'Newsletter',
    utmCampaign: 'newsletter_august',
    totalClicks: 5120,
    conversions: 430,
    status: 'active',
    createdAt: '2026-08-12T08:30:00Z',
  },
];

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
