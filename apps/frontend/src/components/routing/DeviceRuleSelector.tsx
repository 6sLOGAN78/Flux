import React from 'react';
import { Smartphone, Monitor, Apple, ArrowRight, Laptop } from 'lucide-react';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface DeviceRule {
  id: string;
  deviceType: 'ios' | 'android' | 'desktop';
  destinationUrl: string;
}

export interface DeviceRuleSelectorProps {
  rules: DeviceRule[];
  onUpdateRule: (deviceType: 'ios' | 'android' | 'desktop', url: string) => void;
  isLoading?: boolean;
}

export function DeviceRuleSelector({
  rules,
  onUpdateRule,
  isLoading = false,
}: DeviceRuleSelectorProps) {
  const getRuleUrl = (type: 'ios' | 'android' | 'desktop') => {
    return rules.find((r) => r.deviceType === type)?.destinationUrl || '';
  };

  const devices = [
    {
      type: 'ios' as const,
      name: 'Apple iOS',
      subtitle: 'iPhone, iPad, & iPod touch',
      icon: <Smartphone className="h-5 w-5 text-zinc-900 dark:text-zinc-100" />,
      placeholder: 'https://apps.apple.com/app/id123456789 or myapp://',
    },
    {
      type: 'android' as const,
      name: 'Google Android',
      subtitle: 'Android smartphones & tablets',
      icon: <Smartphone className="h-5 w-5 text-zinc-900 dark:text-zinc-100" />,
      placeholder: 'https://play.google.com/store/apps/details?id=com.app or intent://',
    },
    {
      type: 'desktop' as const,
      name: 'Desktop & Other Browsers',
      subtitle: 'macOS, Windows, Linux, & fallback web',
      icon: <Monitor className="h-5 w-5 text-zinc-900 dark:text-zinc-100" />,
      placeholder: 'https://yourbrand.com/desktop-app',
    },
  ];

  return (
    <div className="space-y-6 rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
      <div>
        <div className="flex items-center gap-2">
          <Smartphone className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
          <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
            Device & OS Targeting
          </h2>
          <Badge variant="blue" size="sm">
            User-Agent Parsing
          </Badge>
        </div>
        <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
          Serve platform-tailored destination links based on the operating system of the requesting client.
        </p>
      </div>

      <div className="space-y-4">
        {devices.map((dev) => (
          <div
            key={dev.type}
            className="rounded-xl border border-zinc-200 bg-zinc-50/50 p-4 transition-all dark:border-zinc-800 dark:bg-zinc-900/40"
          >
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-3">
                <div className="flex h-8 w-8 items-center justify-center rounded-lg border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-900">
                  {dev.icon}
                </div>
                <div>
                  <h3 className="text-xs font-semibold text-zinc-900 dark:text-zinc-100">
                    {dev.name}
                  </h3>
                  <p className="text-[11px] text-zinc-400">{dev.subtitle}</p>
                </div>
              </div>
            </div>

            <Input
              placeholder={dev.placeholder}
              value={getRuleUrl(dev.type)}
              onChange={(e) => onUpdateRule(dev.type, e.target.value)}
              disabled={isLoading}
            />
          </div>
        ))}
      </div>
    </div>
  );
}

export default DeviceRuleSelector;
