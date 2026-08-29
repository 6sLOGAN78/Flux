import React, { useState } from 'react';
import { X, Link2, Wand2 } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { getShortDomain } from '@/config/env';

export interface CreateLinkDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: {
    destinationUrl: string;
    customCode?: string;
    title?: string;
    category?: string;
  }) => void;
  isLoading?: boolean;
}

export function CreateLinkDrawer({
  isOpen,
  onClose,
  onSubmit,
  isLoading = false,
}: CreateLinkDrawerProps) {
  const [destinationUrl, setDestinationUrl] = useState('');
  const [title, setTitle] = useState('');
  const [customCode, setCustomCode] = useState('');
  const [category, setCategory] = useState('');

  if (!isOpen) return null;

  const handleGenerateRandomSlug = () => {
    const randomSlug = Math.random().toString(36).substring(2, 8);
    setCustomCode(randomSlug);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!destinationUrl) return;

    onSubmit({
      destinationUrl,
      title: title || undefined,
      customCode: customCode || undefined,
      category: category || undefined,
    });
  };

  return (
    <div className="fixed inset-0 z-50 flex justify-end bg-black/40 backdrop-blur-xs transition-opacity animate-in fade-in">
      <div className="relative h-full w-full max-w-md border-l border-zinc-200 bg-white p-6 shadow-2xl transition-transform animate-in slide-in-from-right dark:border-zinc-800 dark:bg-zinc-950">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-zinc-100 pb-4 dark:border-zinc-900">
          <div>
            <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              Create a new link
            </h2>
            <p className="text-xs text-zinc-500 dark:text-zinc-400">
              Configure your short slug, custom domain, and targeting.
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg p-1 text-zinc-400 hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-900 dark:hover:text-zinc-300"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="mt-6 flex h-[calc(100%-80px)] flex-col justify-between">
          <div className="space-y-4 overflow-y-auto pr-1">
            <Input
              label="Destination URL"
              placeholder="https://example.com/very-long-campaign-url"
              type="url"
              required
              value={destinationUrl}
              onChange={(e) => setDestinationUrl(e.target.value)}
              startIcon={<Link2 className="h-4 w-4" />}
            />

            <Input
              label="Link Title (Optional)"
              placeholder="e.g. Q3 Marketing Campaign"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />

            <div>
              <div className="mb-1 flex items-center justify-between">
                <label className="text-xs font-medium text-zinc-700 dark:text-zinc-300">
                  Short Slug
                </label>
                <button
                  type="button"
                  onClick={handleGenerateRandomSlug}
                  className="inline-flex items-center gap-1 text-[11px] font-medium text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-200"
                >
                  <Wand2 className="h-3 w-3" />
                  <span>Random</span>
                </button>
              </div>
              <Input
                prefix={`${getShortDomain()}/`}
                placeholder="custom-slug"
                value={customCode}
                onChange={(e) => setCustomCode(e.target.value)}
              />
            </div>

            <div>
              <label className="mb-1 block text-xs font-medium text-zinc-700 dark:text-zinc-300">
                Category / Tag
              </label>
              <select
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                className="h-9 w-full rounded-lg border border-zinc-200 bg-white px-3 text-xs text-zinc-900 transition-colors focus:border-zinc-400 focus:outline-none dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-100"
              >
                <option value="">None (Default)</option>
                <option value="Marketing">Marketing</option>
                <option value="Documentation">Documentation</option>
                <option value="Product">Product</option>
                <option value="Social">Social</option>
                <option value="Campaigns">Campaigns</option>
              </select>
            </div>
          </div>

          {/* Footer Actions */}
          <div className="flex items-center justify-end gap-3 border-t border-zinc-100 pt-4 dark:border-zinc-900">
            <Button type="button" variant="outline" size="md" onClick={onClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="md"
              isLoading={isLoading}
            >
              Create link
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default CreateLinkDrawer;
