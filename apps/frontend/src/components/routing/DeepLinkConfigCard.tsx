import React, { useState } from 'react';
import { Sparkles, Smartphone, Save, Check } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';

export interface DeepLinkConfig {
  iosAppStoreId?: string;
  iosBundleId?: string;
  androidPackageName?: string;
  fallbackUrl: string;
}

export interface DeepLinkConfigCardProps {
  config?: DeepLinkConfig;
  onSave?: (config: DeepLinkConfig) => void;
  isLoading?: boolean;
}

export function DeepLinkConfigCard({
  config = {
    iosAppStoreId: '123456789',
    iosBundleId: 'com.flux.mobile',
    androidPackageName: 'com.flux.app',
    fallbackUrl: 'https://flux.to/download',
  },
  onSave,
  isLoading = false,
}: DeepLinkConfigCardProps) {
  const [iosAppStoreId, setIosAppStoreId] = useState(config.iosAppStoreId || '');
  const [iosBundleId, setIosBundleId] = useState(config.iosBundleId || '');
  const [androidPackageName, setAndroidPackageName] = useState(
    config.androidPackageName || ''
  );
  const [fallbackUrl, setFallbackUrl] = useState(config.fallbackUrl || '');
  const [saved, setSaved] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave?.({
      iosAppStoreId,
      iosBundleId,
      androidPackageName,
      fallbackUrl,
    });
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div className="space-y-6 rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
      <div>
        <div className="flex items-center gap-2">
          <Sparkles className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
          <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
            Mobile Deep Linking
          </h2>
          <Badge variant="emerald" size="sm">
            Universal & App Links
          </Badge>
        </div>
        <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
          Seamlessly open installed mobile apps or direct users to the App Store and Google Play.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-5">
        {/* iOS Section */}
        <div className="space-y-3 rounded-xl border border-zinc-100 bg-zinc-50/50 p-4 dark:border-zinc-900 dark:bg-zinc-900/30">
          <h3 className="text-xs font-semibold text-zinc-900 dark:text-zinc-100">
            Apple iOS Universal Links
          </h3>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <Input
              label="iOS Bundle ID"
              placeholder="com.yourbrand.app"
              value={iosBundleId}
              onChange={(e) => setIosBundleId(e.target.value)}
            />
            <Input
              label="Apple App Store ID"
              placeholder="e.g. 1548293021"
              value={iosAppStoreId}
              onChange={(e) => setIosAppStoreId(e.target.value)}
            />
          </div>
        </div>

        {/* Android Section */}
        <div className="space-y-3 rounded-xl border border-zinc-100 bg-zinc-50/50 p-4 dark:border-zinc-900 dark:bg-zinc-900/30">
          <h3 className="text-xs font-semibold text-zinc-900 dark:text-zinc-100">
            Android App Links & Intents
          </h3>
          <Input
            label="Android Package Name"
            placeholder="com.yourbrand.android"
            value={androidPackageName}
            onChange={(e) => setAndroidPackageName(e.target.value)}
          />
        </div>

        {/* Fallback URL */}
        <div>
          <Input
            label="Fallback Web URL"
            type="url"
            required
            placeholder="https://yourbrand.com/download"
            value={fallbackUrl}
            onChange={(e) => setFallbackUrl(e.target.value)}
            description="Used when the mobile app is not installed or when opened on desktop."
          />
        </div>

        <div className="flex items-center justify-between border-t border-zinc-100 pt-4 dark:border-zinc-900">
          {saved ? (
            <span className="flex items-center gap-1.5 text-xs font-medium text-emerald-600 dark:text-emerald-400">
              <Check className="h-4 w-4" /> Configuration saved!
            </span>
          ) : (
            <span />
          )}

          <Button
            type="submit"
            variant="primary"
            size="md"
            isLoading={isLoading}
            leftIcon={<Save className="h-4 w-4" />}
          >
            Save Deep Link Config
          </Button>
        </div>
      </form>
    </div>
  );
}

export default DeepLinkConfigCard;
