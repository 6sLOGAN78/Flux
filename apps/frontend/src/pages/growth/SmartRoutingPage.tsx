import React, { useState } from 'react';
import { Globe2, Smartphone, Sparkles, Layers, ArrowRight, Zap } from 'lucide-react';
import { Tabs } from '@/components/ui/Tabs';
import { Badge } from '@/components/ui/Badge';
import {
  GeoRoutingRuleBuilder,
  GeoRule,
} from '@/components/routing/GeoRoutingRuleBuilder';
import {
  DeviceRuleSelector,
  DeviceRule,
} from '@/components/routing/DeviceRuleSelector';
import {
  DeepLinkConfigCard,
  DeepLinkConfig,
} from '@/components/routing/DeepLinkConfigCard';

const INITIAL_GEO_RULES: GeoRule[] = [
  {
    id: 'geo_1',
    countryCode: 'GB',
    countryName: 'United Kingdom',
    destinationUrl: 'https://uk.store.acme.com',
  },
  {
    id: 'geo_2',
    countryCode: 'DE',
    countryName: 'Germany',
    destinationUrl: 'https://de.store.acme.com',
  },
  {
    id: 'geo_3',
    countryCode: 'JP',
    countryName: 'Japan',
    destinationUrl: 'https://jp.store.acme.com',
  },
];

const INITIAL_DEVICE_RULES: DeviceRule[] = [
  {
    id: 'dev_1',
    deviceType: 'ios',
    destinationUrl: 'https://apps.apple.com/app/id1548293021',
  },
  {
    id: 'dev_2',
    deviceType: 'android',
    destinationUrl: 'https://play.google.com/store/apps/details?id=com.acme.app',
  },
  {
    id: 'dev_3',
    deviceType: 'desktop',
    destinationUrl: 'https://acme.com/web-app',
  },
];

export function SmartRoutingPage() {
  const [activeTab, setActiveTab] = useState('geo');
  const [geoRules, setGeoRules] = useState<GeoRule[]>(INITIAL_GEO_RULES);
  const [deviceRules, setDeviceRules] =
    useState<DeviceRule[]>(INITIAL_DEVICE_RULES);

  const handleAddGeoRule = (rule: Omit<GeoRule, 'id'>) => {
    const newRule: GeoRule = {
      id: `geo_${Date.now()}`,
      ...rule,
    };
    setGeoRules((prev) => [...prev, newRule]);
  };

  const handleDeleteGeoRule = (id: string) => {
    setGeoRules((prev) => prev.filter((r) => r.id !== id));
  };

  const handleUpdateDeviceRule = (
    deviceType: 'ios' | 'android' | 'desktop',
    url: string
  ) => {
    setDeviceRules((prev) =>
      prev.map((r) =>
        r.deviceType === deviceType ? { ...r, destinationUrl: url } : r
      )
    );
  };

  const tabs = [
    { id: 'geo', label: 'Geo Targeting', count: geoRules.length },
    { id: 'device', label: 'Device & OS', count: deviceRules.length },
    { id: 'deeplink', label: 'Universal Deep Linking' },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
            Smart Routing & Deep Linking
          </h1>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Dynamically route visitors by geographic country code, mobile OS, and Universal Links.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Badge variant="emerald" size="sm" dot>
            Edge Routing Active
          </Badge>
        </div>
      </div>

      {/* Visual Flow Diagram */}
      <div className="rounded-2xl border border-zinc-200 bg-zinc-50/75 p-4 dark:border-zinc-800 dark:bg-zinc-900/40">
        <div className="flex flex-col items-center justify-between gap-3 text-xs sm:flex-row">
          <div className="flex items-center gap-2 rounded-lg border border-zinc-200 bg-white px-3 py-2 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
            <span className="h-2 w-2 rounded-full bg-blue-500" />
            <span className="font-semibold text-zinc-900 dark:text-zinc-100">Inbound Request</span>
          </div>

          <ArrowRight className="h-4 w-4 text-zinc-400" />

          <div className="flex items-center gap-2 rounded-lg border border-zinc-200 bg-white px-3 py-2 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
            <Globe2 className="h-4 w-4 text-emerald-600" />
            <span>Country Resolver</span>
          </div>

          <ArrowRight className="h-4 w-4 text-zinc-400" />

          <div className="flex items-center gap-2 rounded-lg border border-zinc-200 bg-white px-3 py-2 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
            <Smartphone className="h-4 w-4 text-purple-600" />
            <span>Device & App Link</span>
          </div>

          <ArrowRight className="h-4 w-4 text-zinc-400" />

          <div className="flex items-center gap-2 rounded-lg border border-zinc-200 bg-white px-3 py-2 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
            <Zap className="h-4 w-4 text-amber-500 fill-current" />
            <span className="font-semibold text-zinc-900 dark:text-zinc-100">Target Redirect</span>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <Tabs
        tabs={tabs}
        activeTab={activeTab}
        onChange={setActiveTab}
        variant="underline"
      />

      {/* Tab Panels */}
      {activeTab === 'geo' && (
        <GeoRoutingRuleBuilder
          rules={geoRules}
          onAddRule={handleAddGeoRule}
          onDeleteRule={handleDeleteGeoRule}
        />
      )}

      {activeTab === 'device' && (
        <DeviceRuleSelector
          rules={deviceRules}
          onUpdateRule={handleUpdateDeviceRule}
        />
      )}

      {activeTab === 'deeplink' && <DeepLinkConfigCard />}
    </div>
  );
}

export default SmartRoutingPage;
