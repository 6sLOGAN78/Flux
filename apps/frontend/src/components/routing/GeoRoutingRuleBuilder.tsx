import React, { useState } from 'react';
import { Globe2, Plus, Trash2, ArrowRight, CornerDownRight } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface GeoRule {
  id: string;
  countryCode: string;
  countryName: string;
  destinationUrl: string;
}

export interface GeoRoutingRuleBuilderProps {
  rules: GeoRule[];
  onAddRule: (rule: Omit<GeoRule, 'id'>) => void;
  onDeleteRule: (id: string) => void;
  fallbackUrl?: string;
  isLoading?: boolean;
}

const COUNTRY_OPTIONS = [
  { code: 'US', name: 'United States' },
  { code: 'GB', name: 'United Kingdom' },
  { code: 'DE', name: 'Germany' },
  { code: 'FR', name: 'France' },
  { code: 'JP', name: 'Japan' },
  { code: 'CA', name: 'Canada' },
  { code: 'AU', name: 'Australia' },
  { code: 'BR', name: 'Brazil' },
  { code: 'IN', name: 'India' },
];

export function GeoRoutingRuleBuilder({
  rules,
  onAddRule,
  onDeleteRule,
  fallbackUrl = 'https://flux.to/global-default',
  isLoading = false,
}: GeoRoutingRuleBuilderProps) {
  const [selectedCountry, setSelectedCountry] = useState('GB');
  const [targetUrl, setTargetUrl] = useState('');

  const handleAdd = (e: React.FormEvent) => {
    e.preventDefault();
    if (!targetUrl.trim()) return;

    const countryObj = COUNTRY_OPTIONS.find((c) => c.code === selectedCountry);
    onAddRule({
      countryCode: selectedCountry,
      countryName: countryObj ? countryObj.name : selectedCountry,
      destinationUrl: targetUrl.trim(),
    });

    setTargetUrl('');
  };

  return (
    <div className="space-y-6 rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
      <div>
        <div className="flex items-center gap-2">
          <Globe2 className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
          <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
            Geographic Country Targeting
          </h2>
          <Badge variant="emerald" size="sm">
            Edge IP Geolocation
          </Badge>
        </div>
        <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
          Route visitors to localized store pages or regional mirrors based on the requester's country code.
        </p>
      </div>

      {/* Add New Rule Form */}
      <form
        onSubmit={handleAdd}
        className="flex flex-col gap-3 rounded-xl border border-zinc-100 bg-zinc-50/70 p-4 dark:border-zinc-900 dark:bg-zinc-900/40 sm:flex-row sm:items-end"
      >
        <div className="w-full sm:w-48">
          <label className="mb-1 block text-xs font-medium text-zinc-700 dark:text-zinc-300">
            Select Country
          </label>
          <select
            value={selectedCountry}
            onChange={(e) => setSelectedCountry(e.target.value)}
            className="h-9 w-full rounded-lg border border-zinc-200 bg-white px-3 text-xs text-zinc-900 focus:border-zinc-400 focus:outline-none dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-100"
          >
            {COUNTRY_OPTIONS.map((c) => (
              <option key={c.code} value={c.code}>
                {c.name} ({c.code})
              </option>
            ))}
          </select>
        </div>

        <div className="flex-1">
          <Input
            label="Target Regional URL"
            placeholder="https://uk.brand.com/products"
            type="url"
            required
            value={targetUrl}
            onChange={(e) => setTargetUrl(e.target.value)}
          />
        </div>

        <Button
          type="submit"
          variant="primary"
          size="md"
          isLoading={isLoading}
          leftIcon={<Plus className="h-4 w-4" />}
        >
          Add Country Rule
        </Button>
      </form>

      {/* Active Rules List */}
      <div className="space-y-3">
        <div className="text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
          Active Geo Routing Hierarchy
        </div>

        {rules.length === 0 ? (
          <div className="rounded-xl border border-dashed border-zinc-200 p-6 text-center text-xs text-zinc-400 dark:border-zinc-800">
            No country-specific rules configured yet. All visitors will use the fallback URL below.
          </div>
        ) : (
          <div className="divide-y divide-zinc-100 rounded-xl border border-zinc-200 bg-white dark:divide-zinc-900 dark:border-zinc-800 dark:bg-zinc-950">
            {rules.map((rule, idx) => (
              <div
                key={rule.id}
                className="flex items-center justify-between p-3.5 transition-colors hover:bg-zinc-50/60 dark:hover:bg-zinc-900/40"
              >
                <div className="flex items-center gap-3">
                  <span className="flex h-5 w-5 items-center justify-center rounded bg-zinc-100 font-mono text-[10px] font-bold text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300">
                    {idx + 1}
                  </span>
                  <Badge variant="zinc" size="sm" className="font-mono">
                    {rule.countryCode}
                  </Badge>
                  <span className="text-xs font-semibold text-zinc-900 dark:text-zinc-100">
                    {rule.countryName}
                  </span>
                  <ArrowRight className="h-3 w-3 text-zinc-400" />
                  <span className="font-mono text-xs text-zinc-600 dark:text-zinc-400 truncate max-w-xs sm:max-w-md">
                    {rule.destinationUrl}
                  </span>
                </div>

                <button
                  type="button"
                  onClick={() => onDeleteRule(rule.id)}
                  className="rounded-md p-1.5 text-zinc-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950/50 dark:hover:text-red-400"
                  title="Delete rule"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              </div>
            ))}
          </div>
        )}

        {/* Global Fallback Row */}
        <div className="flex items-center gap-3 rounded-xl border border-zinc-200 bg-zinc-50/50 p-3 text-xs text-zinc-600 dark:border-zinc-800 dark:bg-zinc-900/30 dark:text-zinc-400">
          <CornerDownRight className="h-4 w-4 text-zinc-400 shrink-0" />
          <span className="font-semibold text-zinc-700 dark:text-zinc-300">
            Fallback (All other countries):
          </span>
          <span className="font-mono text-zinc-500 truncate">{fallbackUrl}</span>
        </div>
      </div>
    </div>
  );
}

export default GeoRoutingRuleBuilder;
