import React from 'react';
import { Globe2 } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface CountryStat {
  countryCode: string;
  countryName: string;
  clicks: number;
  percentage: number;
}

export interface GeographicChoroplethProps {
  countries: CountryStat[];
  className?: string;
}

export function GeographicChoropleth({
  countries,
  className,
}: GeographicChoroplethProps) {
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
            <Globe2 className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              Top Geographic Locations
            </h2>
          </div>
          <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
            Visitor traffic breakdown by ISO-3166 country origin.
          </p>
        </div>
      </div>

      <div className="space-y-3">
        {countries.map((country, idx) => (
          <div key={country.countryCode} className="space-y-1.5">
            <div className="flex items-center justify-between text-xs">
              <div className="flex items-center gap-2">
                <span className="font-mono text-[11px] font-bold text-zinc-400">
                  #{idx + 1}
                </span>
                <Badge variant="zinc" size="sm" className="font-mono">
                  {country.countryCode}
                </Badge>
                <span className="font-medium text-zinc-900 dark:text-zinc-100">
                  {country.countryName}
                </span>
              </div>

              <div className="flex items-center gap-2 font-mono">
                <span className="font-bold text-zinc-900 dark:text-zinc-100">
                  {country.clicks.toLocaleString()}
                </span>
                <span className="text-zinc-400">({country.percentage}%)</span>
              </div>
            </div>

            {/* Progress Bar */}
            <div className="h-1.5 w-full overflow-hidden rounded-full bg-zinc-100 dark:bg-zinc-900">
              <div
                style={{ width: `${country.percentage}%` }}
                className="h-full rounded-full bg-zinc-900 dark:bg-zinc-100 transition-all duration-500"
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default GeographicChoropleth;
