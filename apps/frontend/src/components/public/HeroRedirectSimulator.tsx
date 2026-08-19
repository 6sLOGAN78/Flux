import React, { useState } from 'react';
import {
  Link2,
  Copy,
  Check,
  Zap,
  ArrowRight,
  ExternalLink,
  ShieldCheck,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';

export function HeroRedirectSimulator() {
  const [longUrl, setLongUrl] = useState(
    'https://github.com/flux-platform/core-infrastructure-edge-router'
  );
  const [customSlug, setCustomSlug] = useState('edge-v2');
  const [isCopied, setIsCopied] = useState(false);
  const [latencyMs, setLatencyMs] = useState(4.2);
  const [simulatedRedirects, setSimulatedRedirects] = useState(14820);
  const [isSimulating, setIsSimulating] = useState(false);

  const shortUrl = `flux.to/${customSlug || 'fast'}`;

  const handleCopy = () => {
    navigator.clipboard?.writeText(`https://${shortUrl}`);
    setIsCopied(true);
    setTimeout(() => setIsCopied(false), 2000);
  };

  const handleSimulateRedirect = () => {
    setIsSimulating(true);
    // Simulate real-time edge DNS + Redis route lookup
    setTimeout(() => {
      const randomLatency = (3.5 + Math.random() * 2.8).toFixed(1);
      setLatencyMs(parseFloat(randomLatency));
      setSimulatedRedirects((prev) => prev + 1);
      setIsSimulating(false);
    }, 280);
  };

  return (
    <div className="mx-auto w-full max-w-2xl rounded-2xl border border-zinc-200 bg-white p-4 shadow-xl shadow-zinc-200/40 transition-all dark:border-zinc-800 dark:bg-zinc-950 dark:shadow-none sm:p-6">
      {/* Top Simulator Status Bar */}
      <div className="mb-4 flex items-center justify-between border-b border-zinc-100 pb-3 dark:border-zinc-900">
        <div className="flex items-center gap-2">
          <span className="flex h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
          <span className="text-xs font-semibold text-zinc-900 dark:text-zinc-100">
            Interactive Edge Shortener Simulator
          </span>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="emerald" size="sm" dot>
            Global Mesh Active
          </Badge>
          <span className="font-mono text-[11px] text-zinc-400">
            TLS 1.3 / HTTP/3
          </span>
        </div>
      </div>

      {/* Simulator Form Controls */}
      <div className="space-y-3">
        <div>
          <label className="mb-1 block text-left text-xs font-medium text-zinc-600 dark:text-zinc-400">
            Destination URL
          </label>
          <div className="flex items-center rounded-lg border border-zinc-200 bg-zinc-50/50 px-3 py-2 text-xs transition-colors focus-within:border-zinc-400 focus-within:bg-white focus-within:ring-2 focus-within:ring-zinc-900/10 dark:border-zinc-800 dark:bg-zinc-900/40 dark:focus-within:border-zinc-600 dark:focus-within:bg-zinc-950">
            <Link2 className="mr-2 h-4 w-4 shrink-0 text-zinc-400 dark:text-zinc-500" />
            <input
              type="url"
              value={longUrl}
              onChange={(e) => setLongUrl(e.target.value)}
              placeholder="https://yourbrand.com/deep/campaign/path"
              className="w-full bg-transparent font-mono text-xs text-zinc-900 placeholder:text-zinc-400 focus:outline-none dark:text-zinc-100 dark:placeholder:text-zinc-600"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div>
            <label className="mb-1 block text-left text-xs font-medium text-zinc-600 dark:text-zinc-400">
              Custom Domain & Slug
            </label>
            <div className="flex items-center rounded-lg border border-zinc-200 bg-zinc-50/50 px-3 py-2 text-xs transition-colors focus-within:border-zinc-400 focus-within:bg-white focus-within:ring-2 focus-within:ring-zinc-900/10 dark:border-zinc-800 dark:bg-zinc-900/40 dark:focus-within:border-zinc-600 dark:focus-within:bg-zinc-950">
              <span className="select-none font-mono text-xs font-medium text-zinc-400 dark:text-zinc-500">
                flux.to/
              </span>
              <input
                type="text"
                value={customSlug}
                onChange={(e) => setCustomSlug(e.target.value)}
                placeholder="custom-slug"
                className="w-full bg-transparent font-mono text-xs font-medium text-zinc-900 focus:outline-none dark:text-zinc-100"
              />
            </div>
          </div>

          <div className="flex items-end">
            <Button
              variant="primary"
              className="w-full"
              onClick={handleSimulateRedirect}
              isLoading={isSimulating}
              leftIcon={<Zap className="h-3.5 w-3.5 fill-current" />}
            >
              Shorten Link
            </Button>
          </div>
        </div>
      </div>

      {/* Live Edge Result Preview Banner */}
      <div className="mt-4 rounded-xl border border-zinc-100 bg-zinc-50 p-3.5 dark:border-zinc-900 dark:bg-zinc-900/50">
        <div className="flex flex-col items-start justify-between gap-3 sm:flex-row sm:items-center">
          <div className="text-left">
            <div className="flex items-center gap-2">
              <span className="font-mono text-sm font-semibold text-zinc-900 dark:text-zinc-100">
                https://{shortUrl}
              </span>
              <button
                type="button"
                onClick={handleCopy}
                className="inline-flex items-center gap-1 rounded-md border border-zinc-200 bg-white px-2 py-1 text-[11px] font-medium text-zinc-700 shadow-xs transition-colors hover:bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-300 dark:hover:bg-zinc-900"
              >
                {isCopied ? (
                  <>
                    <Check className="h-3 w-3 text-emerald-600 dark:text-emerald-400" />
                    <span className="text-emerald-600 dark:text-emerald-400">
                      Copied
                    </span>
                  </>
                ) : (
                  <>
                    <Copy className="h-3 w-3" />
                    <span>Copy</span>
                  </>
                )}
              </button>
            </div>
            <p className="mt-0.5 truncate font-mono text-[11px] text-zinc-400 dark:text-zinc-500">
              Target: {longUrl}
            </p>
          </div>

          <div className="flex items-center gap-4 text-right">
            <div>
              <div className="text-[10px] uppercase tracking-wider text-zinc-400">
                Edge Latency
              </div>
              <div className="font-mono text-sm font-bold text-emerald-600 dark:text-emerald-400">
                {latencyMs} ms
              </div>
            </div>
            <div className="border-l border-zinc-200 pl-4 dark:border-zinc-800">
              <div className="text-[10px] uppercase tracking-wider text-zinc-400">
                Redirects
              </div>
              <div className="font-mono text-sm font-semibold text-zinc-700 dark:text-zinc-300">
                {simulatedRedirects.toLocaleString()}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default HeroRedirectSimulator;
