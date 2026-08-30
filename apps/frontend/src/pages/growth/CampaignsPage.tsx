import React, { useState } from 'react';
import { Sparkles, Plus, Layers, Filter } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Tabs } from '@/components/ui/Tabs';
import { UTMBuilderStudio } from '@/components/campaigns/UTMBuilderStudio';
import {
  CampaignListTable,
} from '@/components/campaigns/CampaignListTable';
import { useCreateLink } from '@/hooks/useLinksQuery';
import { useGetCampaigns, useCreateCampaign } from '@/hooks/useCampaignsQuery';
import { useAnalyticsCampaigns } from '@/hooks/useAnalyticsQuery';

export function CampaignsPage() {
  const [activeTab, setActiveTab] = useState('campaigns');

  const { data: campaignsData, isLoading: isCampLoading, isError: isCampError } = useGetCampaigns();
  const { data: analyticsData, isLoading: isAnaLoading } = useAnalyticsCampaigns();
  
  const createCampaignMutation = useCreateCampaign();
  const createLinkMutation = useCreateLink();

  const handleGenerateLink = async (data: {
    finalUrl: string;
    campaignName: string;
    customSlug?: string;
  }) => {
    // First create the campaign
    try {
      const camp = await createCampaignMutation.mutateAsync({
        name: data.campaignName,
        utm_campaign: data.campaignName, // Usually they use the name as utm_campaign here
      });

      // Then create the link using the new campaign
      createLinkMutation.mutate({
        destinationUrl: data.finalUrl,
        customCode: data.customSlug,
        title: data.campaignName,
        campaignId: camp.id,
      }, {
        onSettled: () => {
          setActiveTab('campaigns');
        }
      });
    } catch (e) {
      console.error(e);
    }
  };

  const campaignsList = (campaignsData || []).map((camp: any) => {
    const perf = analyticsData?.data?.find((p: any) => p.campaign_id === camp.id);
    return {
      ...camp,
      clicks: perf?.clicks || 0,
      unique_visitors: perf?.unique_visitors || 0,
    };
  });

  const isLoading = isCampLoading || isAnaLoading;
  const isError = isCampError;

  const tabs = [
    { id: 'campaigns', label: 'Active Campaigns', count: campaignsList.length },
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
        isLoading ? (
          <div className="text-zinc-500 p-4 text-center">Loading campaigns...</div>
        ) : isError ? (
          <div className="text-red-500 p-4 text-center">Error loading campaigns.</div>
        ) : (
          <CampaignListTable campaigns={campaignsList as any} />
        )
      ) : (
        <UTMBuilderStudio
          onGenerateLink={handleGenerateLink}
          isLoading={createCampaignMutation.isPending}
        />
      )}
    </div>
  );
}

export default CampaignsPage;
