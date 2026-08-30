import React, { useState, useMemo } from 'react';
import {
  Search,
  Plus,
  Filter,
  Tag,
  Trash2,
  X,
  SlidersHorizontal,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Tabs } from '@/components/ui/Tabs';
import { LinksTable, LinkItem } from '@/components/links/LinksTable';
import { CreateLinkDrawer } from '@/components/links/CreateLinkDrawer';
import { BulkCategorizeModal } from '@/components/links/BulkCategorizeModal';
import { useCreateLink, useBulkCategorize, useGetLinks } from '@/hooks/useLinksQuery';
import { getShortDomain } from '@/config/env';


export function LinksListPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const { data: queryData, isLoading, refetch } = useGetLinks({ search: searchQuery });
  const rawLinks = queryData?.data || [];
  const links: LinkItem[] = rawLinks.map((r: any) => ({
    id: r.id,
    shortCode: r.shortCode || r.short_code,
    destinationUrl: r.destinationUrl || r.destination_url,
    title: r.title,
    clicks: r.clicks || 0,
    createdAt: r.createdAt || r.created_at,
    category: r.category || 'all',
    domain: r.domain || getShortDomain(),
  }));
  const [activeCategory, setActiveCategory] = useState('all');
  const [selectedLinkIds, setSelectedLinkIds] = useState<string[]>([]);
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  const [isBulkModalOpen, setIsBulkModalOpen] = useState(false);

  const createLinkMutation = useCreateLink();
  const bulkCategorizeMutation = useBulkCategorize();

  const categories = useMemo(() => {
    return [
      { id: 'all', label: 'All Links', count: links.length },
      { id: 'Marketing', label: 'Marketing' },
      { id: 'Documentation', label: 'Documentation' },
      { id: 'Social', label: 'Social' },
    ];
  }, [links]);

  const filteredLinks = useMemo(() => {
    return links.filter((link) => {
      const matchesSearch =
        searchQuery === '' ||
        link.shortCode.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (link.title && link.title.toLowerCase().includes(searchQuery.toLowerCase())) ||
        link.destinationUrl.toLowerCase().includes(searchQuery.toLowerCase());

      const matchesCategory =
        activeCategory === 'all' || link.category === activeCategory;

      return matchesSearch && matchesCategory;
    });
  }, [links, searchQuery, activeCategory]);

  const handleToggleSelect = (id: string) => {
    setSelectedLinkIds((prev) =>
      prev.includes(id) ? prev.filter((item) => item !== id) : [...prev, id]
    );
  };

  const handleSelectAll = () => {
    if (selectedLinkIds.length === filteredLinks.length) {
      setSelectedLinkIds([]);
    } else {
      setSelectedLinkIds(filteredLinks.map((l) => l.id));
    }
  };

  const handleCreateLink = (data: {
    destinationUrl: string;
    customCode?: string;
    title?: string;
    category?: string;
    campaignId?: string;
    utmSource?: string;
    utmMedium?: string;
    utmCampaign?: string;
  }) => {
    createLinkMutation.mutate(
      {
        destinationUrl: data.destinationUrl,
        customCode: data.customCode,
        title: data.title,
        campaignId: data.campaignId,
        utmSource: data.utmSource,
        utmMedium: data.utmMedium,
        utmCampaign: data.utmCampaign,
      },
      {
        onSuccess: (res: any) => {
          setIsDrawerOpen(false);
          // Wait for the query cache to refetch to update UI automatically
          refetch();
        },
        onError: (error) => {
          console.error("Failed to create link:", error);
        },
      }
    );
  };

  const handleBulkCategorize = (newCategory: string) => {
    bulkCategorizeMutation.mutate(
      {
        linkIds: selectedLinkIds,
        categoryId: newCategory,
      },
      {
        onSettled: () => {
          setSelectedLinkIds([]);
          setIsBulkModalOpen(false);
        },
      }
    );
  };

  const handleDeleteSelected = () => {
    setSelectedLinkIds([]);
  };

  const handleDeleteSingle = (id: string) => {
    setSelectedLinkIds((prev) => prev.filter((item) => item !== id));
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
            Links
          </h1>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Create, brand, and measure your redirects across domains.
          </p>
        </div>

        <Button
          variant="primary"
          size="md"
          onClick={() => setIsDrawerOpen(true)}
          leftIcon={<Plus className="h-4 w-4" />}
        >
          Create Link
        </Button>
      </div>

      {/* Filter and Search Bar */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-1 items-center rounded-lg border border-zinc-200 bg-white px-3 py-1.5 text-xs transition-colors focus-within:border-zinc-400 focus-within:ring-2 focus-within:ring-zinc-900/10 dark:border-zinc-800 dark:bg-zinc-950 dark:focus-within:border-zinc-600">
          <Search className="mr-2 h-4 w-4 shrink-0 text-zinc-400" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search links by title, slug, or URL..."
            className="w-full bg-transparent text-xs text-zinc-900 placeholder:text-zinc-400 focus:outline-none dark:text-zinc-100"
          />
          <kbd className="hidden sm:inline-block rounded bg-zinc-100 px-1.5 py-0.5 font-mono text-[10px] text-zinc-500 dark:bg-zinc-800">
            /
          </kbd>
        </div>

        <Tabs
          tabs={categories}
          activeTab={activeCategory}
          onChange={setActiveCategory}
          variant="pills"
        />
      </div>

      {/* Bulk Action Toolbar */}
      {selectedLinkIds.length > 0 && (
        <div className="flex items-center justify-between rounded-xl border border-zinc-200 bg-zinc-900 px-4 py-2.5 text-xs text-white shadow-lg dark:border-zinc-700 dark:bg-zinc-100 dark:text-zinc-900 animate-in fade-in slide-in-from-bottom-2">
          <div className="flex items-center gap-2">
            <span className="font-semibold">{selectedLinkIds.length}</span>
            <span>links selected</span>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setIsBulkModalOpen(true)}
              className="inline-flex items-center gap-1 rounded-lg bg-zinc-800 px-3 py-1 text-xs font-medium text-white hover:bg-zinc-700 dark:bg-zinc-200 dark:text-zinc-900 dark:hover:bg-zinc-300"
            >
              <Tag className="h-3 w-3" />
              <span>Categorize</span>
            </button>
            <button
              type="button"
              onClick={handleDeleteSelected}
              className="inline-flex items-center gap-1 rounded-lg bg-red-600 px-3 py-1 text-xs font-medium text-white hover:bg-red-700"
            >
              <Trash2 className="h-3 w-3" />
              <span>Delete</span>
            </button>
            <button
              type="button"
              onClick={() => setSelectedLinkIds([])}
              className="rounded-lg p-1 text-zinc-400 hover:text-white dark:hover:text-zinc-900"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        </div>
      )}

      {/* Links List Table */}
      <LinksTable
        links={filteredLinks}
        selectedLinkIds={selectedLinkIds}
        onToggleSelect={handleToggleSelect}
        onSelectAll={handleSelectAll}
        onDeleteLink={handleDeleteSingle}
      />

      {/* Create Link Drawer */}
      <CreateLinkDrawer
        isOpen={isDrawerOpen}
        onClose={() => setIsDrawerOpen(false)}
        onSubmit={handleCreateLink}
        isLoading={createLinkMutation.isPending}
        error={createLinkMutation.error ? createLinkMutation.error.message || "Failed to create link. Please try again." : null}
      />

      {/* Bulk Categorize Modal */}
      <BulkCategorizeModal
        isOpen={isBulkModalOpen}
        selectedCount={selectedLinkIds.length}
        onClose={() => setIsBulkModalOpen(false)}
        onCategorize={handleBulkCategorize}
        isLoading={bulkCategorizeMutation.isPending}
      />
    </div>
  );
}

export default LinksListPage;
