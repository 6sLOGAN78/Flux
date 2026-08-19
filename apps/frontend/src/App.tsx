import React from 'react';

export function App() {
  return (
    <div className="flex min-h-[100dvh] flex-col items-center justify-center p-6 text-center">
      <div className="mx-auto max-w-lg rounded-xl border border-zinc-200 bg-white p-8 shadow-xs dark:border-zinc-800 dark:bg-zinc-900">
        <div className="inline-flex items-center gap-2 rounded-full border border-zinc-200 bg-zinc-50 px-3 py-1 text-xs font-medium text-zinc-700 dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-300">
          <span className="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
          Flux Platform • Notion & Dub Minimalist UI
        </div>
        <h1 className="mt-4 text-2xl font-semibold tracking-tight text-zinc-900 dark:text-zinc-50">
          High-Performance URL & Attribution Infrastructure
        </h1>
        <p className="mt-2 text-sm text-zinc-500 dark:text-zinc-400">
          Sub-10ms edge redirects, ClickHouse time-series analytics, and multi-touch marketing attribution.
        </p>
        <div className="mt-6 flex items-center justify-center gap-3">
          <button className="rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200">
            Open Dashboard
          </button>
          <button className="rounded-lg border border-zinc-200 bg-transparent px-4 py-2 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-100 dark:border-zinc-800 dark:text-zinc-300 dark:hover:bg-zinc-800">
            Documentation
          </button>
        </div>
      </div>
    </div>
  );
}

export default App;
