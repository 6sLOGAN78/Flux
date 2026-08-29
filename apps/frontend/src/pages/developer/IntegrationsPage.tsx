import React, { useState, useMemo } from 'react';
import { Layers, Plus, Check, Search } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { Input } from '@/components/ui/Input';
import {
  AppConnectorCard,
  IntegrationApp,
} from '@/components/integrations/AppConnectorCard';

const INITIAL_APPS: IntegrationApp[] = [];

export function IntegrationsPage() {
  const [apps, setApps] = useState<IntegrationApp[]>(INITIAL_APPS);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [notice, setNotice] = useState<string | null>(null);

  const categories = [
    { id: 'all', label: 'All Apps' },
    { id: 'chat', label: 'Communication' },
    { id: 'automation', label: 'Automation' },
    { id: 'analytics', label: 'Analytics' },
    { id: 'crm', label: 'CRM & Growth' },
  ];

  const filteredApps = useMemo(() => {
    return apps.filter((app) => {
      const matchesCategory =
        selectedCategory === 'all' || app.category === selectedCategory;
      const matchesSearch =
        app.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        app.description.toLowerCase().includes(searchQuery.toLowerCase());
      return matchesCategory && matchesSearch;
    });
  }, [apps, selectedCategory, searchQuery]);

  const handleToggleConnect = (id: string) => {
    setApps((prev) =>
      prev.map((app) => {
        if (app.id === id) {
          const nextStatus =
            app.status === 'connected' ? 'not_connected' : 'connected';
          setNotice(
            nextStatus === 'connected'
              ? `Connected ${app.name} integration successfully.`
              : `Disconnected ${app.name} integration.`
          );
          setTimeout(() => setNotice(null), 3000);
          return { ...app, status: nextStatus };
        }
        return app;
      })
    );
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              Integrations Directory
            </h1>
            <Badge variant="zinc" size="sm">
              Ecosystem
            </Badge>
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Connect Flux with your favorite chat, CRM, and automation workflows.
          </p>
        </div>
      </div>

      {notice && (
        <div className="flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-semibold text-emerald-800 dark:border-emerald-900/50 dark:bg-emerald-950/30 dark:text-emerald-300 animate-in fade-in">
          <Check className="h-4 w-4" />
          <span>{notice}</span>
        </div>
      )}

      {/* Filter and Search Bar */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div className="flex flex-wrap gap-1.5">
          {categories.map((c) => (
            <button
              key={c.id}
              type="button"
              onClick={() => setSelectedCategory(c.id)}
              className={`rounded-lg px-3 py-1.5 text-xs font-medium transition-colors ${
                selectedCategory === c.id
                  ? 'bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900'
                  : 'bg-zinc-100 text-zinc-600 hover:bg-zinc-200/70 dark:bg-zinc-900 dark:text-zinc-400 dark:hover:bg-zinc-800'
              }`}
            >
              {c.label}
            </button>
          ))}
        </div>

        <div className="w-full sm:w-64">
          <Input
            placeholder="Search integrations..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            startIcon={<Search className="h-4 w-4 text-zinc-400" />}
          />
        </div>
      </div>

      {/* Apps Grid */}
      <div className="grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-3">
        {filteredApps.map((app) => (
          <AppConnectorCard
            key={app.id}
            app={app}
            onToggleConnect={handleToggleConnect}
          />
        ))}
      </div>
    </div>
  );
}

export default IntegrationsPage;
