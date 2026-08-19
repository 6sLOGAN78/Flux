import React from 'react';
import { Sliders, Plus, Trash2, Link2, Percent } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface ABVariant {
  id: string;
  name: string;
  destinationUrl: string;
  weight: number;
}

export function calculateNormalizedWeights(
  variants: ABVariant[],
  changedId: string,
  newWeight: number
): ABVariant[] {
  const boundedWeight = Math.max(0, Math.min(100, newWeight));
  const otherVariants = variants.filter((v) => v.id !== changedId);

  if (otherVariants.length === 0) {
    return [{ ...variants[0], weight: 100 }];
  }

  const remainingWeight = 100 - boundedWeight;
  const currentOtherTotal = otherVariants.reduce((sum, v) => sum + v.weight, 0);

  const updatedOthers = otherVariants.map((v) => {
    if (currentOtherTotal === 0) {
      return {
        ...v,
        weight: Math.round(remainingWeight / otherVariants.length),
      };
    }
    const ratio = v.weight / currentOtherTotal;
    return {
      ...v,
      weight: Math.round(remainingWeight * ratio),
    };
  });

  // Adjust rounding errors so sum is exactly 100
  const total =
    boundedWeight + updatedOthers.reduce((sum, v) => sum + v.weight, 0);
  const diff = 100 - total;
  if (diff !== 0 && updatedOthers.length > 0) {
    updatedOthers[0].weight += diff;
  }

  return variants.map((v) => {
    if (v.id === changedId) return { ...v, weight: boundedWeight };
    const found = updatedOthers.find((o) => o.id === v.id);
    return found ? found : v;
  });
}

const VARIANT_COLORS = [
  'bg-zinc-900 dark:bg-zinc-100',
  'bg-blue-600',
  'bg-emerald-600',
  'bg-purple-600',
  'bg-amber-600',
];

export interface VariantAllocationSliderProps {
  variants: ABVariant[];
  onChangeVariants: (variants: ABVariant[]) => void;
  isLoading?: boolean;
  className?: string;
}

export function VariantAllocationSlider({
  variants,
  onChangeVariants,
  isLoading = false,
  className,
}: VariantAllocationSliderProps) {
  const handleWeightChange = (id: string, newWeight: number) => {
    const updated = calculateNormalizedWeights(variants, id, newWeight);
    onChangeVariants(updated);
  };

  const handleUrlChange = (id: string, newUrl: string) => {
    onChangeVariants(
      variants.map((v) => (v.id === id ? { ...v, destinationUrl: newUrl } : v))
    );
  };

  const handleNameChange = (id: string, newName: string) => {
    onChangeVariants(
      variants.map((v) => (v.id === id ? { ...v, name: newName } : v))
    );
  };

  const handleAddVariant = () => {
    const nextLetter = String.fromCharCode(65 + variants.length);
    const newVariant: ABVariant = {
      id: `var_${Date.now()}`,
      name: `Variant ${nextLetter}`,
      destinationUrl: 'https://flux.to/new-page',
      weight: 0,
    };
    const updated = calculateNormalizedWeights(
      [...variants, newVariant],
      newVariant.id,
      Math.round(100 / (variants.length + 1))
    );
    onChangeVariants(updated);
  };

  const handleDeleteVariant = (id: string) => {
    if (variants.length <= 2) return; // Keep at least 2 variants
    const remaining = variants.filter((v) => v.id !== id);
    const updated = calculateNormalizedWeights(
      remaining,
      remaining[0].id,
      remaining[0].weight + (variants.find((v) => v.id === id)?.weight || 0)
    );
    onChangeVariants(updated);
  };

  return (
    <div
      className={cn(
        'space-y-6 rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="flex items-center justify-between border-b border-zinc-100 pb-4 dark:border-zinc-900">
        <div>
          <div className="flex items-center gap-2">
            <Sliders className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              Traffic Split Allocation
            </h2>
            <Badge variant="zinc" size="sm">
              Total 100%
            </Badge>
          </div>
          <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
            Adjust incoming traffic percentage weights across your destination variants.
          </p>
        </div>

        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={handleAddVariant}
          leftIcon={<Plus className="h-3.5 w-3.5" />}
        >
          Add Variant
        </Button>
      </div>

      {/* Visual Weight Progress Bar */}
      <div className="flex h-3 w-full overflow-hidden rounded-full bg-zinc-100 dark:bg-zinc-900">
        {variants.map((v, idx) => (
          <div
            key={v.id}
            style={{ width: `${v.weight}%` }}
            className={cn(
              'h-full transition-all duration-300',
              VARIANT_COLORS[idx % VARIANT_COLORS.length]
            )}
            title={`${v.name}: ${v.weight}%`}
          />
        ))}
      </div>

      {/* Variant Cards */}
      <div className="space-y-4">
        {variants.map((variant, idx) => (
          <div
            key={variant.id}
            className="rounded-xl border border-zinc-200 bg-zinc-50/50 p-4 transition-all dark:border-zinc-800 dark:bg-zinc-900/40"
          >
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <span
                  className={cn(
                    'h-2.5 w-2.5 rounded-full',
                    VARIANT_COLORS[idx % VARIANT_COLORS.length]
                  )}
                />
                <input
                  type="text"
                  value={variant.name}
                  onChange={(e) => handleNameChange(variant.id, e.target.value)}
                  className="bg-transparent text-xs font-semibold text-zinc-900 focus:outline-none dark:text-zinc-100"
                />
              </div>

              <div className="flex items-center gap-3">
                <span className="font-mono text-xs font-bold text-zinc-900 dark:text-zinc-100">
                  {variant.weight}%
                </span>
                {variants.length > 2 && (
                  <button
                    type="button"
                    onClick={() => handleDeleteVariant(variant.id)}
                    className="rounded p-1 text-zinc-400 hover:text-red-600 dark:hover:text-red-400"
                    title="Remove variant"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                )}
              </div>
            </div>

            <div className="space-y-3">
              <Input
                placeholder="https://flux.to/variant-destination"
                type="url"
                required
                value={variant.destinationUrl}
                onChange={(e) => handleUrlChange(variant.id, e.target.value)}
              />

              <div className="flex items-center gap-3">
                <input
                  type="range"
                  min="0"
                  max="100"
                  value={variant.weight}
                  onChange={(e) =>
                    handleWeightChange(variant.id, Number(e.target.value))
                  }
                  className="h-1.5 w-full cursor-pointer appearance-none rounded-lg bg-zinc-200 accent-zinc-900 dark:bg-zinc-700 dark:accent-zinc-100"
                />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default VariantAllocationSlider;
