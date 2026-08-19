import React from 'react';
import { Smartphone, Monitor, Tablet } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface DeviceStat {
  label: string;
  value: number;
  percentage: number;
  color: string;
}

export interface DeviceDonutChartProps {
  devices: DeviceStat[];
  className?: string;
}

export function DeviceDonutChart({
  devices,
  className,
}: DeviceDonutChartProps) {
  return (
    <div
      className={cn(
        'space-y-4 rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="flex items-center justify-between border-b border-zinc-100 pb-4 dark:border-zinc-900">
        <div>
          <div className="flex items-center gap-2">
            <Smartphone className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              Devices & Platforms
            </h2>
          </div>
          <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
            Client user-agent distribution across hardware form factors.
          </p>
        </div>
      </div>

      <div className="space-y-3">
        {devices.map((dev) => (
          <div key={dev.label} className="space-y-1.5">
            <div className="flex items-center justify-between text-xs">
              <div className="flex items-center gap-2">
                <span
                  style={{ backgroundColor: dev.color }}
                  className="h-2.5 w-2.5 rounded-full"
                />
                <span className="font-medium text-zinc-900 dark:text-zinc-100">
                  {dev.label}
                </span>
              </div>

              <div className="flex items-center gap-2 font-mono">
                <span className="font-bold text-zinc-900 dark:text-zinc-100">
                  {dev.value.toLocaleString()}
                </span>
                <span className="text-zinc-400">({dev.percentage}%)</span>
              </div>
            </div>

            <div className="h-1.5 w-full overflow-hidden rounded-full bg-zinc-100 dark:bg-zinc-900">
              <div
                style={{ width: `${dev.percentage}%`, backgroundColor: dev.color }}
                className="h-full rounded-full transition-all duration-500"
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default DeviceDonutChart;
