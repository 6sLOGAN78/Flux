import React, { useState, useMemo } from 'react';
import {
  Sparkles,
  Link2,
  Copy,
  Check,
  Zap,
  Tag,
  ArrowRight,
  ExternalLink,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface UTMParams {
  baseUrl: string;
  utmSource?: string;
  utmMedium?: string;
  utmCampaign?: string;
  utmTerm?: string;
  utmContent?: string;
}

export function buildSanitizedUtmUrl({
  baseUrl,
  utmSource,
  utmMedium,
  utmCampaign,
  utmTerm,
  utmContent,
}: UTMParams): string {
  if (!baseUrl) return '';

  try {
    const parsedUrl = new URL(
      baseUrl.startsWith('http://') || baseUrl.startsWith('https://')
        ? baseUrl
        : `https://${baseUrl}`
    );

    if (utmSource) parsedUrl.searchParams.set('utm_source', utmSource.trim());
    if (utmMedium) parsedUrl.searchParams.set('utm_medium', utmMedium.trim());
    if (utmCampaign)
      parsedUrl.searchParams.set('utm_campaign', utmCampaign.trim());
    if (utmTerm) parsedUrl.searchParams.set('utm_term', utmTerm.trim());
    if (utmContent)
      parsedUrl.searchParams.set('utm_content', utmContent.trim());

    return parsedUrl.toString();
  } catch {
    // Fallback if URL parsing fails during live typing
    const params = new URLSearchParams();
    if (utmSource) params.set('utm_source', utmSource.trim());
    if (utmMedium) params.set('utm_medium', utmMedium.trim());
    if (utmCampaign) params.set('utm_campaign', utmCampaign.trim());
    if (utmTerm) params.set('utm_term', utmTerm.trim());
    if (utmContent) params.set('utm_content', utmContent.trim());

    const queryString = params.toString();
    return queryString ? `${baseUrl}?${queryString}` : baseUrl;
  }
}

export interface UTMBuilderStudioProps {
  initialBaseUrl?: string;
  onGenerateLink?: (data: {
    finalUrl: string;
    campaignName: string;
    customSlug?: string;
  }) => void;
  isLoading?: boolean;
  className?: string;
}

const CHANNEL_PRESETS = [
  { name: 'Google Ads', source: 'google', medium: 'cpc' },
  { name: 'Meta / FB', source: 'facebook', medium: 'paid_social' },
  { name: 'Twitter / X', source: 'twitter', medium: 'social' },
  { name: 'LinkedIn', source: 'linkedin', medium: 'sponsored' },
  { name: 'Newsletter', source: 'newsletter', medium: 'email' },
  { name: 'YouTube', source: 'youtube', medium: 'video' },
];

export function UTMBuilderStudio({
  initialBaseUrl = 'https://flux.to/pricing',
  onGenerateLink,
  isLoading = false,
  className,
}: UTMBuilderStudioProps) {
  const [baseUrl, setBaseUrl] = useState(initialBaseUrl);
  const [utmSource, setUtmSource] = useState('twitter');
  const [utmMedium, setUtmMedium] = useState('social');
  const [utmCampaign, setUtmCampaign] = useState('summer_2026');
  const [utmTerm, setUtmTerm] = useState('');
  const [utmContent, setUtmContent] = useState('');
  const [customSlug, setCustomSlug] = useState('summer-deal');
  const [isCopied, setIsCopied] = useState(false);

  const generatedUrl = useMemo(() => {
    return buildSanitizedUtmUrl({
      baseUrl,
      utmSource,
      utmMedium,
      utmCampaign,
      utmTerm,
      utmContent,
    });
  }, [baseUrl, utmSource, utmMedium, utmCampaign, utmTerm, utmContent]);

  const handleApplyPreset = (preset: { source: string; medium: string }) => {
    setUtmSource(preset.source);
    setUtmMedium(preset.medium);
  };

  const handleCopyUrl = () => {
    if (!generatedUrl) return;
    navigator.clipboard?.writeText(generatedUrl);
    setIsCopied(true);
    setTimeout(() => setIsCopied(false), 2000);
  };

  const handleCreateLink = (e: React.FormEvent) => {
    e.preventDefault();
    if (!generatedUrl) return;

    onGenerateLink?.({
      finalUrl: generatedUrl,
      campaignName: utmCampaign || 'Campaign',
      customSlug: customSlug || undefined,
    });
  };

  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="border-b border-zinc-100 pb-4 dark:border-zinc-900">
        <div className="flex items-center gap-2">
          <Sparkles className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
          <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
            Visual UTM Builder
          </h2>
          <Badge variant="blue" size="sm">
            Attribution Ready
          </Badge>
        </div>
        <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
          Build sanitized tracking URLs with preset advertising channels and campaign tagging.
        </p>
      </div>

      {/* Preset Channel Buttons */}
      <div className="mt-5">
        <label className="mb-2 block text-xs font-medium text-zinc-700 dark:text-zinc-300">
          Channel Presets
        </label>
        <div className="flex flex-wrap gap-2">
          {CHANNEL_PRESETS.map((preset) => {
            const isSelected =
              utmSource === preset.source && utmMedium === preset.medium;
            return (
              <button
                key={preset.name}
                type="button"
                onClick={() => handleApplyPreset(preset)}
                className={cn(
                  'rounded-lg border px-3 py-1.5 text-xs font-medium transition-colors',
                  isSelected
                    ? 'border-zinc-900 bg-zinc-900 text-white dark:border-zinc-100 dark:bg-zinc-100 dark:text-zinc-900 font-semibold'
                    : 'border-zinc-200 bg-white text-zinc-700 hover:bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-300 dark:hover:bg-zinc-800'
                )}
              >
                {preset.name}
              </button>
            );
          })}
        </div>
      </div>

      {/* Form Fields */}
      <form onSubmit={handleCreateLink} className="mt-6 space-y-4">
        <Input
          label="Destination Base URL"
          placeholder="https://yourbrand.com/landing-page"
          type="url"
          required
          value={baseUrl}
          onChange={(e) => setBaseUrl(e.target.value)}
          startIcon={<Link2 className="h-4 w-4" />}
        />

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <Input
            label="UTM Source"
            placeholder="e.g. google, twitter"
            required
            value={utmSource}
            onChange={(e) => setUtmSource(e.target.value)}
          />
          <Input
            label="UTM Medium"
            placeholder="e.g. cpc, social, email"
            required
            value={utmMedium}
            onChange={(e) => setUtmMedium(e.target.value)}
          />
          <Input
            label="UTM Campaign"
            placeholder="e.g. summer_sale, launch_v2"
            required
            value={utmCampaign}
            onChange={(e) => setUtmCampaign(e.target.value)}
          />
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Input
            label="UTM Term (Optional Keyword)"
            placeholder="e.g. edge-redirects"
            value={utmTerm}
            onChange={(e) => setUtmTerm(e.target.value)}
          />
          <Input
            label="UTM Content (Optional Creative ID)"
            placeholder="e.g. hero_banner_cta"
            value={utmContent}
            onChange={(e) => setUtmContent(e.target.value)}
          />
        </div>

        <div>
          <Input
            label="Custom Short Slug"
            prefix="flux.to/"
            placeholder="campaign-slug"
            value={customSlug}
            onChange={(e) => setCustomSlug(e.target.value)}
          />
        </div>

        {/* Real-time Generated URL Output Preview Box */}
        <div className="rounded-xl border border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-800 dark:bg-zinc-900/50">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-zinc-700 dark:text-zinc-300">
              Generated Target URL Preview
            </span>
            <button
              type="button"
              onClick={handleCopyUrl}
              className="inline-flex items-center gap-1 rounded-md border border-zinc-200 bg-white px-2 py-1 text-[11px] font-medium text-zinc-700 shadow-xs hover:bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-300 dark:hover:bg-zinc-900"
            >
              {isCopied ? (
                <>
                  <Check className="h-3 w-3 text-emerald-600 dark:text-emerald-400" />
                  <span className="text-emerald-600 dark:text-emerald-400">
                    Copied
                  </span>
                </>
              ) : (
                <>
                  <Copy className="h-3 w-3" />
                  <span>Copy URL</span>
                </>
              )}
            </button>
          </div>
          <p className="mt-2 break-all font-mono text-xs text-zinc-600 dark:text-zinc-400">
            {generatedUrl || 'Enter a valid base URL to generate preview...'}
          </p>
        </div>

        <div className="flex justify-end pt-2">
          <Button
            type="submit"
            variant="primary"
            size="md"
            isLoading={isLoading}
            leftIcon={<Zap className="h-3.5 w-3.5 fill-current" />}
          >
            Create Campaign Link
          </Button>
        </div>
      </form>
    </div>
  );
}

export default UTMBuilderStudio;
