import React from 'react';
import { Link } from 'react-router-dom';
import {
  Globe2,
  BarChart3,
  QrCode,
  GitFork,
  ShieldCheck,
  Zap,
  ArrowRight,
  Sparkles,
  Layers,
  CheckCircle2,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { PublicHeader } from '@/components/public/PublicHeader';
import { PublicFooter } from '@/components/public/PublicFooter';
import { HeroRedirectSimulator } from '@/components/public/HeroRedirectSimulator';

export function LandingPage() {
  const featureList = [
    {
      icon: <Globe2 className="h-5 w-5 text-zinc-900 dark:text-zinc-100" />,
      title: 'Global Edge Mesh',
      description:
        'Sub-10ms redirects delivered via 300+ edge points-of-presence with Anycast DNS and local in-memory caching.',
      badge: 'Sub-10ms',
    },
    {
      icon: <BarChart3 className="h-5 w-5 text-zinc-900 dark:text-zinc-100" />,
      title: 'Real-Time ClickHouse Analytics',
      description:
        'Query billions of click events in milliseconds. Slice by UTM tags, geographic city, OS, and referrers.',
      badge: 'OLAP Powered',
    },
    {
      icon: <QrCode className="h-5 w-5 text-zinc-900 dark:text-zinc-100" />,
      title: 'Dynamic QR Studio',
      description:
        'Generate high-resolution vector SVG and PNG QR codes with embedded brand logos, color gradients, and pixel styling.',
      badge: 'Vector SVG',
    },
    {
      icon: <GitFork className="h-5 w-5 text-zinc-900 dark:text-zinc-100" />,
      title: 'Traffic Splitter & A/B Routing',
      description:
        'Distribute inbound visitors across multiple target variants with deterministic hashing and conversion attribution.',
      badge: 'A/B Testing',
    },
    {
      icon: <Sparkles className="h-5 w-5 text-zinc-900 dark:text-zinc-100" />,
      title: 'Smart Deep Linking',
      description:
        'Route users seamlessly to native iOS Universal Links, Android App Links, or fall back gracefully to the web.',
      badge: 'iOS & Android',
    },
    {
      icon: <ShieldCheck className="h-5 w-5 text-zinc-900 dark:text-zinc-100" />,
      title: 'Enterprise RBAC & SAML SSO',
      description:
        'Okta, Azure AD, and Google Workspace SSO with SCIM automatic provisioning and audit event logging.',
      badge: 'SOC2 Ready',
    },
  ];

  const networkStats = [
    { label: 'Edge Locations', value: '300+' },
    { label: 'Average Latency', value: '< 8ms' },
    { label: 'Uptime SLA', value: '99.99%' },
    { label: 'Daily Redirects', value: '500M+' },
  ];

  return (
    <div className="flex min-h-screen flex-col bg-white text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
      <PublicHeader />

      {/* Hero Section */}
      <section className="relative overflow-hidden px-4 pt-12 pb-16 sm:px-6 md:pt-16 md:pb-24">
        <div className="mx-auto max-w-4xl text-center">
          {/* Eyebrow / Status Tag */}
          <div className="inline-flex items-center gap-2 rounded-full border border-zinc-200 bg-zinc-50 px-3 py-1 text-xs font-medium text-zinc-700 shadow-xs dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-300">
            <span className="flex h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse" />
            <span>Sub-10ms Edge Redirects</span>
            <span className="text-zinc-400">•</span>
            <span className="font-mono text-zinc-500">v1.0 Ready</span>
          </div>

          {/* Hero Headline (strictly max 2 lines desktop) */}
          <h1 className="mt-6 text-3xl font-bold tracking-tight sm:text-5xl md:text-6xl md:leading-[1.1]">
            The Modern Link Infrastructure
            <br />
            for High-Velocity Teams.
          </h1>

          {/* Subtext (concise, <= 20 words) */}
          <p className="mx-auto mt-4 max-w-xl text-sm text-zinc-500 dark:text-zinc-400 sm:text-base">
            Sub-10ms edge redirects, ClickHouse time-series analytics, and real-time attribution built for scale.
          </p>

          {/* CTAs */}
          <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
            <Link to="/sign-up">
              <Button size="lg" variant="primary" rightIcon={<ArrowRight className="h-4 w-4" />}>
                Start for free
              </Button>
            </Link>
            <Link to="/pricing">
              <Button size="lg" variant="outline">
                View Pricing
              </Button>
            </Link>
          </div>
        </div>

        {/* Hero Interactive Shortener Simulator */}
        <div className="mt-10">
          <HeroRedirectSimulator />
        </div>
      </section>

      {/* Network Stats Bar */}
      <section className="border-y border-zinc-200 bg-zinc-50/50 py-8 dark:border-zinc-800 dark:bg-zinc-900/30">
        <div className="mx-auto max-w-6xl px-4 sm:px-6">
          <div className="grid grid-cols-2 gap-6 text-center md:grid-cols-4">
            {networkStats.map((stat, i) => (
              <div key={i} className="space-y-1">
                <div className="font-mono text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100 sm:text-3xl">
                  {stat.value}
                </div>
                <div className="text-xs font-medium text-zinc-500 dark:text-zinc-400">
                  {stat.label}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Core Features Bento Grid */}
      <section className="px-4 py-16 sm:px-6 md:py-24">
        <div className="mx-auto max-w-6xl space-y-12">
          <div className="text-center">
            <h2 className="text-2xl font-bold tracking-tight sm:text-3xl">
              Engineered for Enterprise Reliability & Latency
            </h2>
            <p className="mx-auto mt-2 max-w-lg text-xs text-zinc-500 dark:text-zinc-400 sm:text-sm">
              Everything required to manage, route, and measure marketing links across web and mobile platforms.
            </p>
          </div>

          <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
            {featureList.map((feat, idx) => (
              <div
                key={idx}
                className="group relative rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs transition-all hover:border-zinc-300 dark:border-zinc-800 dark:bg-zinc-950 dark:hover:border-zinc-700"
              >
                <div className="flex items-center justify-between">
                  <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-zinc-100 dark:bg-zinc-900">
                    {feat.icon}
                  </div>
                  <Badge variant="zinc" size="sm">
                    {feat.badge}
                  </Badge>
                </div>
                <h3 className="mt-4 text-sm font-semibold text-zinc-900 dark:text-zinc-100">
                  {feat.title}
                </h3>
                <p className="mt-1.5 text-xs text-zinc-500 leading-relaxed dark:text-zinc-400">
                  {feat.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Bottom Conversion CTA Section */}
      <section className="border-t border-zinc-200 bg-zinc-50/50 px-4 py-16 dark:border-zinc-800 dark:bg-zinc-900/30 sm:px-6">
        <div className="mx-auto max-w-3xl text-center">
          <h2 className="text-2xl font-bold tracking-tight sm:text-3xl">
            Ready to upgrade your link infrastructure?
          </h2>
          <p className="mx-auto mt-2 max-w-md text-xs text-zinc-500 dark:text-zinc-400 sm:text-sm">
            Deploy in minutes with our open API, TypeScript client, and edge-native Anycast network.
          </p>
          <div className="mt-6 flex justify-center gap-3">
            <Link to="/sign-up">
              <Button size="md" variant="primary" rightIcon={<ArrowRight className="h-4 w-4" />}>
                Create free account
              </Button>
            </Link>
            <Link to="/pricing">
              <Button size="md" variant="outline">
                Compare Plans
              </Button>
            </Link>
          </div>
        </div>
      </section>

      <PublicFooter />
    </div>
  );
}

export default LandingPage;
