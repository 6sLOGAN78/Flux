import React from 'react';
import { ArrowUpDown, ChevronLeft, ChevronRight, Inbox } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface Column<T> {
  key: string;
  header: React.ReactNode;
  render?: (item: T, index: number) => React.ReactNode;
  sortable?: boolean;
  align?: 'left' | 'center' | 'right';
  width?: string;
  className?: string;
}

export interface DataTablePagination {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
  totalItems?: number;
}

export interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  keyExtractor: (item: T, index: number) => string;
  isLoading?: boolean;
  emptyMessage?: string;
  emptyIcon?: React.ReactNode;
  onRowClick?: (item: T) => void;
  sortKey?: string;
  sortDirection?: 'asc' | 'desc';
  onSort?: (key: string) => void;
  pagination?: DataTablePagination;
  className?: string;
}

export function DataTable<T>({
  columns,
  data,
  keyExtractor,
  isLoading = false,
  emptyMessage = 'No records found',
  emptyIcon,
  onRowClick,
  sortKey,
  sortDirection,
  onSort,
  pagination,
  className,
}: DataTableProps<T>) {
  return (
    <div
      className={cn(
        'w-full overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs text-zinc-700 dark:text-zinc-300">
          <thead className="border-b border-zinc-200 bg-zinc-50/75 text-[11px] font-semibold uppercase tracking-wider text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900/50 dark:text-zinc-400">
            <tr>
              {columns.map((column) => {
                const isSorted = sortKey === column.key;
                return (
                  <th
                    key={column.key}
                    scope="col"
                    style={{ width: column.width }}
                    className={cn(
                      'px-4 py-3 select-none',
                      column.align === 'center' && 'text-center',
                      column.align === 'right' && 'text-right',
                      column.sortable && 'cursor-pointer hover:text-zinc-800 dark:hover:text-zinc-200',
                      column.className
                    )}
                    onClick={() => column.sortable && onSort?.(column.key)}
                  >
                    <div
                      className={cn(
                        'inline-flex items-center gap-1.5',
                        column.align === 'center' && 'justify-center',
                        column.align === 'right' && 'justify-end'
                      )}
                    >
                      <span>{column.header}</span>
                      {column.sortable && (
                        <ArrowUpDown
                          className={cn(
                            'h-3 w-3 transition-colors',
                            isSorted
                              ? 'text-zinc-900 dark:text-zinc-100'
                              : 'text-zinc-400 opacity-60'
                          )}
                        />
                      )}
                    </div>
                  </th>
                );
              })}
            </tr>
          </thead>

          <tbody className="divide-y divide-zinc-100 dark:divide-zinc-900">
            {isLoading ? (
              // Loading Skeleton
              Array.from({ length: 4 }).map((_, index) => (
                <tr key={`loading-skeleton-${index}`} className="animate-pulse">
                  {columns.map((col) => (
                    <td key={col.key} className="px-4 py-3.5">
                      <div className="h-4 w-3/4 rounded bg-zinc-200/70 dark:bg-zinc-800/70" />
                    </td>
                  ))}
                </tr>
              ))
            ) : data.length === 0 ? (
              // Empty State
              <tr>
                <td
                  colSpan={columns.length}
                  className="px-4 py-12 text-center text-zinc-400 dark:text-zinc-500"
                >
                  <div className="flex flex-col items-center justify-center gap-2">
                    {emptyIcon || <Inbox className="h-8 w-8 stroke-1 text-zinc-300 dark:text-zinc-600" />}
                    <p className="text-xs font-medium text-zinc-500 dark:text-zinc-400">
                      {emptyMessage}
                    </p>
                  </div>
                </td>
              </tr>
            ) : (
              // Rows
              data.map((item, index) => (
                <tr
                  key={keyExtractor(item, index)}
                  onClick={() => onRowClick?.(item)}
                  className={cn(
                    'group transition-colors hover:bg-zinc-50/70 dark:hover:bg-zinc-900/40',
                    onRowClick && 'cursor-pointer'
                  )}
                >
                  {columns.map((col) => (
                    <td
                      key={col.key}
                      className={cn(
                        'px-4 py-3 align-middle',
                        col.align === 'center' && 'text-center',
                        col.align === 'right' && 'text-right',
                        col.className
                      )}
                    >
                      {col.render
                        ? col.render(item, index)
                        : (item as Record<string, any>)[col.key]}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Optional Pagination Footer */}
      {pagination && (
        <div className="flex items-center justify-between border-t border-zinc-100 bg-zinc-50/50 px-4 py-2.5 text-xs text-zinc-500 dark:border-zinc-900 dark:bg-zinc-900/30 dark:text-zinc-400">
          <div>
            {pagination.totalItems !== undefined ? (
              <span>Total {pagination.totalItems} entries</span>
            ) : (
              <span>
                Page {pagination.page} of {pagination.totalPages}
              </span>
            )}
          </div>
          <div className="flex items-center gap-1">
            <button
              type="button"
              disabled={pagination.page <= 1}
              onClick={() => pagination.onPageChange(pagination.page - 1)}
              className="inline-flex h-7 w-7 items-center justify-center rounded-md border border-zinc-200 bg-white text-zinc-700 transition-colors hover:bg-zinc-50 disabled:opacity-40 disabled:cursor-not-allowed dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-300 dark:hover:bg-zinc-900"
              aria-label="Previous page"
            >
              <ChevronLeft className="h-3.5 w-3.5" />
            </button>
            <span className="px-2 font-mono text-[11px]">
              {pagination.page} / {pagination.totalPages}
            </span>
            <button
              type="button"
              disabled={pagination.page >= pagination.totalPages}
              onClick={() => pagination.onPageChange(pagination.page + 1)}
              className="inline-flex h-7 w-7 items-center justify-center rounded-md border border-zinc-200 bg-white text-zinc-700 transition-colors hover:bg-zinc-50 disabled:opacity-40 disabled:cursor-not-allowed dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-300 dark:hover:bg-zinc-900"
              aria-label="Next page"
            >
              <ChevronRight className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

export default DataTable;
