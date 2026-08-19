import React from 'react';
import {
  MessageSquare,
  Zap,
  Layers,
  BarChart,
  CheckCircle,
  ExternalLink,
  Settings,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface IntegrationApp {
  id: string;
  name: string;
  category: 'chat' | 'automation' | 'crm' | 'analytics';
  description: string;
  status: 'connected' | 'not_connected';
  installedAt?: string;
}

export interface AppConnectorCardProps {
  app: IntegrationApp;
  onToggleConnect: (id: string) => void;
  isLoading?: boolean;
  className?: string;
}

export function AppConnectorCard({
  app,
  onToggleConnect,
  isLoading = false,
  className,
}: AppConnectorCardProps) {
  const isConnected = app.status === 'connected';

  const getCategoryIcon = () => {
    switch (app.category) {
      case 'chat':
        return <MessageSquare className="h-5 w-5 text-zinc-900 dark:text-zinc-100" />;
      case 'automation':
        return <Zap className="h-5 w-5 text-zinc-900 dark:text-zinc-100" />;
      case 'analytics':
        return <BarChart className="h-5 w-5 text-zinc-900 dark:text-zinc-100" />;
      case 'crm':
      default:
        return <Layers className="h-5 w-5 text-zinc-900 dark:text-zinc-100" />;
    }
  };

  return (
    <div
      className={cn(
        'flex flex-col justify-between rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs transition-all hover:border-zinc-300 dark:border-zinc-800 dark:bg-zinc-950 dark:hover:border-zinc-700',
        className
      )}
    >
      <div>
        <div className="flex items-start justify-between">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl border border-zinc-100 bg-zinc-50 shadow-2xs dark:border-zinc-800 dark:bg-zinc-900">
            {getCategoryIcon()}
          </div>
          <Badge
            variant={isConnected ? 'emerald' : 'zinc'}
            size="sm"
            dot={isConnected}
          >
            {isConnected ? 'Connected' : 'Available'}
          </Badge>
        </div>

        <div className="mt-4">
          <h3 className="text-sm font-bold text-zinc-900 dark:text-zinc-100">
            {app.name}
          </h3>
          <p className="mt-1 text-xs leading-relaxed text-zinc-500 dark:text-zinc-400">
            {app.description}
          </p>
        </div>
      </div>

      <div className="mt-6 flex items-center justify-between border-t border-zinc-100 pt-4 dark:border-zinc-900">
        <span className="font-mono text-[11px] capitalize text-zinc-400">
          {app.category}
        </span>

        <Button
          variant={isConnected ? 'outline' : 'primary'}
          size="sm"
          onClick={() => onToggleConnect(app.id)}
          isLoading={isLoading}
          leftIcon={isConnected ? <Settings className="h-3.5 w-3.5" /> : undefined}
        >
          {isConnected ? 'Configure' : 'Connect'}
        </Button>
      </div>
    </div>
  );
}

export default AppConnectorCard;
