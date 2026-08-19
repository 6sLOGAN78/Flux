import React from 'react';
import { Tag, Edit2, Trash2, Link2, Plus } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';

export interface CategoryItem {
  id: string;
  name: string;
  color: string;
  description?: string;
  linkCount: number;
}

export interface CategoryGridProps {
  categories: CategoryItem[];
  onEdit: (category: CategoryItem) => void;
  onDelete: (id: string) => void;
  isLoading?: boolean;
}

export function CategoryGrid({
  categories,
  onEdit,
  onDelete,
  isLoading = false,
}: CategoryGridProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <div
            key={i}
            className="h-36 rounded-xl border border-zinc-200 bg-white p-5 animate-pulse dark:border-zinc-800 dark:bg-zinc-950"
          />
        ))}
      </div>
    );
  }

  if (categories.length === 0) {
    return (
      <div className="rounded-xl border border-zinc-200 bg-white p-12 text-center text-xs text-zinc-400 dark:border-zinc-800 dark:bg-zinc-950">
        No categories found. Create a new category to organize your links.
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {categories.map((cat) => (
        <div
          key={cat.id}
          className="group relative flex flex-col justify-between rounded-xl border border-zinc-200 bg-white p-5 shadow-xs transition-all hover:border-zinc-300 dark:border-zinc-800 dark:bg-zinc-950 dark:hover:border-zinc-700"
        >
          <div>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span
                  style={{ backgroundColor: cat.color }}
                  className="h-3 w-3 rounded-full shadow-xs"
                />
                <h3 className="text-xs font-semibold text-zinc-900 dark:text-zinc-100">
                  {cat.name}
                </h3>
              </div>

              <span className="rounded-full bg-zinc-100 px-2 py-0.5 font-mono text-[10px] font-medium text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300">
                {`${cat.linkCount} links`}
              </span>
            </div>

            {cat.description && (
              <p className="mt-2 text-xs text-zinc-500 leading-relaxed dark:text-zinc-400">
                {cat.description}
              </p>
            )}
          </div>

          <div className="mt-4 flex items-center justify-end gap-2 border-t border-zinc-100 pt-3 dark:border-zinc-900">
            <button
              type="button"
              onClick={() => onEdit(cat)}
              className="rounded-md p-1.5 text-zinc-400 hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-900 dark:hover:text-zinc-300"
              title="Edit category"
            >
              <Edit2 className="h-3.5 w-3.5" />
            </button>
            <button
              type="button"
              onClick={() => onDelete(cat.id)}
              className="rounded-md p-1.5 text-zinc-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950/50 dark:hover:text-red-400"
              title="Delete category"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}

export default CategoryGrid;
