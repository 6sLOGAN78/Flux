import React, { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  ArrowLeft,
  Copy,
  Check,
  ExternalLink,
  BarChart2,
  QrCode,
  Save,
  Globe2,
  Layers,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { Tabs } from '@/components/ui/Tabs';
import { QRStudioCanvas } from '@/components/qr/QRStudioCanvas';
import { useUpdateLink } from '@/hooks/useLinksQuery';

export function LinkDetailPage() {
  const { id } = useParams<{ id: string }>();

  const [destinationUrl, setDestinationUrl] = useState(
    'https://flux.to/blog/high-performance-edge-router-v2'
  );
  const [title, setTitle] = useState('V2 Launch Announcement');
  const [description, setDescription] = useState(
    'Public announcement for the sub-10ms Anycast link infrastructure.'
  );
  const [customCode, setCustomCode] = useState('v2-launch');
  const [category, setCategory] = useState('Marketing');
  const [activeTab, setActiveTab] = useState('settings');
  const [isCopied, setIsCopied] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);

  const updateLinkMutation = useUpdateLink();

  const shortUrl = `flux.to/${customCode}`;

  const handleCopy = () => {
    navigator.clipboard?.writeText(`https://${shortUrl}`);
    setIsCopied(true);
    setTimeout(() => setIsCopied(false), 2000);
  };

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    if (!id) return;

    updateLinkMutation.mutate(
      {
        id,
        body: {
          destinationUrl,
          title,
          description,
        },
      },
      {
        onSettled: () => {
          setSaveSuccess(true);
          setTimeout(() => setSaveSuccess(false), 2500);
        },
      }
    );
  };

  const tabs = [
    { id: 'settings', label: 'Settings' },
    { id: 'qr', label: 'QR Code Studio' },
  ];

  return (
    <div className="space-y-6">
      {/* Top Breadcrumb & Actions */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div className="flex items-center gap-3">
          <Link
            to="/links"
            className="rounded-lg border border-zinc-200 bg-white p-2 text-zinc-500 hover:bg-zinc-50 hover:text-zinc-900 dark:border-zinc-800 dark:bg-zinc-950 dark:hover:bg-zinc-900"
          >
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
                Link Settings
              </h1>
              <Badge variant="emerald" size="sm" dot>
                Live
              </Badge>
            </div>
            <p className="font-mono text-xs text-zinc-400">
              ID: {id || 'link_1'}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <div className="flex items-center rounded-lg border border-zinc-200 bg-white px-3 py-1.5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
            <span className="font-mono text-xs font-semibold text-zinc-900 dark:text-zinc-100">
              https://{shortUrl}
            </span>
            <button
              type="button"
              onClick={handleCopy}
              className="ml-2 rounded p-1 text-zinc-400 hover:text-zinc-700 dark:hover:text-zinc-300"
              title="Copy short link"
            >
              {isCopied ? (
                <Check className="h-3.5 w-3.5 text-emerald-600 dark:text-emerald-400" />
              ) : (
                <Copy className="h-3.5 w-3.5" />
              )}
            </button>
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

      {/* Tab Content */}
      {activeTab === 'settings' ? (
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
          {/* Main Edit Form */}
          <div className="lg:col-span-2">
            <form
              onSubmit={handleSave}
              className="space-y-5 rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950"
            >
              <Input
                label="Destination URL"
                type="url"
                required
                value={destinationUrl}
                onChange={(e) => setDestinationUrl(e.target.value)}
                description="The target web page visitors will be redirected to."
              />

              <Input
                label="Link Title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Internal reference title"
              />

              <div className="space-y-1.5">
                <label className="block text-xs font-medium text-zinc-700 dark:text-zinc-300">
                  Description
                </label>
                <textarea
                  rows={3}
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  className="w-full rounded-lg border border-zinc-200 bg-white p-3 text-xs text-zinc-900 placeholder:text-zinc-400 focus:border-zinc-400 focus:outline-none dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-100"
                  placeholder="Optional internal notes or campaign details"
                />
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Input
                  label="Short Slug"
                  prefix="flux.to/"
                  value={customCode}
                  onChange={(e) => setCustomCode(e.target.value)}
                />

                <div>
                  <label className="mb-1 block text-xs font-medium text-zinc-700 dark:text-zinc-300">
                    Category Tag
                  </label>
                  <select
                    value={category}
                    onChange={(e) => setCategory(e.target.value)}
                    className="h-9 w-full rounded-lg border border-zinc-200 bg-white px-3 text-xs text-zinc-900 focus:border-zinc-400 focus:outline-none dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-100"
                  >
                    <option value="Marketing">Marketing</option>
                    <option value="Documentation">Documentation</option>
                    <option value="Product">Product</option>
                    <option value="Social">Social</option>
                  </select>
                </div>
              </div>

              <div className="flex items-center justify-between border-t border-zinc-100 pt-4 dark:border-zinc-900">
                {saveSuccess ? (
                  <span className="flex items-center gap-1.5 text-xs font-medium text-emerald-600 dark:text-emerald-400">
                    <Check className="h-4 w-4" /> Changes saved successfully!
                  </span>
                ) : (
                  <span />
                )}

                <Button
                  type="submit"
                  variant="primary"
                  size="md"
                  isLoading={updateLinkMutation.isPending}
                  leftIcon={<Save className="h-4 w-4" />}
                >
                  Save Changes
                </Button>
              </div>
            </form>
          </div>

          {/* Quick Metrics Summary Sidebar */}
          <div className="space-y-4">
            <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
              <div className="text-xs font-semibold text-zinc-900 dark:text-zinc-100">
                Performance Overview
              </div>
              <div className="mt-4 space-y-3">
                <div className="flex items-center justify-between border-b border-zinc-100 pb-2 dark:border-zinc-900">
                  <span className="text-xs text-zinc-500">Total Clicks</span>
                  <span className="font-mono text-xs font-bold text-zinc-900 dark:text-zinc-100">
                    1,420
                  </span>
                </div>
                <div className="flex items-center justify-between border-b border-zinc-100 pb-2 dark:border-zinc-900">
                  <span className="text-xs text-zinc-500">Average Latency</span>
                  <span className="font-mono text-xs font-bold text-emerald-600 dark:text-emerald-400">
                    4.2 ms
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-xs text-zinc-500">Status</span>
                  <Badge variant="emerald" size="sm">
                    Healthy & Active
                  </Badge>
                </div>
              </div>
            </div>
          </div>
        </div>
      ) : (
        <QRStudioCanvas url={`https://${shortUrl}`} />
      )}
    </div>
  );
}

export default LinkDetailPage;
