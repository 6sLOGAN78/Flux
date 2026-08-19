import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { Check, Minus, Zap, ArrowRight, ShieldCheck, Sparkles } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export function PricingMatrix() {
  const [billingInterval, setBillingInterval] = useState<'monthly' | 'annual'>('annual');

  const tiers = [
    {
      id: 'tier_free',
      name: 'Free',
      badge: 'Starter',
      description: 'Essential link shortening and click tracking for developers and personal projects.',
      price: { monthly: '$0', annual: '$0' },
      period: 'forever',
      cta: 'Start for Free',
      ctaVariant: 'outline' as const,
      popular: false,
      features: [
        '1,000 links / month',
        '10,000 monthly redirects',
        '1 custom branded domain',
        '7-day analytics retention',
        'Standard QR code generation',
        'Community Discord support',
      ],
    },
    {
      id: 'tier_pro',
      name: 'Pro',
      badge: 'Most Popular',
      description: 'High-speed link infrastructure, ClickHouse analytics, and dynamic QR Studio for growing brands.',
      price: { monthly: '$24', annual: '$19' },
      period: '/ month, billed annually',
      cta: 'Start 14-Day Free Trial',
      ctaVariant: 'primary' as const,
      popular: true,
      features: [
        '50,000 links / month',
        '500,000 monthly redirects',
        '5 custom branded domains',
        '365-day ClickHouse analytics',
        'Dynamic QR Studio & vector SVG export',
        'Smart Routing (Geo, OS, Device)',
        'Traffic Splitter (A/B testing)',
        'Webhooks & 600 req/min API access',
        'Priority email support',
      ],
    },
    {
      id: 'tier_enterprise',
      name: 'Enterprise',
      badge: 'Scale & Security',
      description: 'Dedicated multi-region Anycast edge mesh, SAML SSO, SCIM, and 99.99% uptime SLA.',
      price: { monthly: '$199', annual: '$159' },
      period: '/ month, billed annually',
      cta: 'Talk to Sales',
      ctaVariant: 'outline' as const,
      popular: false,
      features: [
        'Unlimited short links',
        '10M+ monthly redirects included',
        'Unlimited custom domains with auto-SSL',
        'Multi-year ClickHouse data warehouse retention',
        'Multi-touch marketing attribution engine',
        'SAML 2.0 / Okta / Azure AD SSO & SCIM',
        'Tenant isolation & granular RBAC',
        'Sub-10ms global edge Anycast SLA',
        'Dedicated Slack channel & 99.99% uptime SLA',
      ],
    },
  ];

  const comparisonRows = [
    {
      category: 'Volume & Capacity',
      items: [
        { name: 'Monthly Short Links', free: '1,000', pro: '50,000', enterprise: 'Unlimited' },
        { name: 'Monthly Edge Redirects', free: '10,000', pro: '500,000', enterprise: '10M+ (Custom tier)' },
        { name: 'Custom Domains', free: '1 domain', pro: '5 domains', enterprise: 'Unlimited' },
        { name: 'Team Members & Seats', free: '1 member', pro: '5 seats', enterprise: 'Unlimited' },
      ],
    },
    {
      category: 'Analytics & Intelligence',
      items: [
        { name: 'ClickHouse Retention', free: '7 days', pro: '365 days', enterprise: 'Multi-year cold tier' },
        { name: 'Real-Time Geo & Device Breakdown', free: true, pro: true, enterprise: true },
        { name: 'UTM Builder & Tagging Hub', free: true, pro: true, enterprise: true },
        { name: 'Multi-Touch Attribution Modeling', free: false, pro: false, enterprise: true },
        { name: 'Raw Click Event Streaming', free: false, pro: true, enterprise: true },
      ],
    },
    {
      category: 'Routing & Optimization',
      items: [
        { name: 'Smart Geo & OS Deep Linking', free: false, pro: true, enterprise: true },
        { name: 'Traffic Splitter (A/B Distribution)', free: false, pro: true, enterprise: true },
        { name: 'Dynamic QR Studio (Color & Logo)', free: 'Basic PNG', pro: 'Full Custom + SVG', enterprise: 'Custom Branding' },
        { name: 'OpenGraph Meta Dynamic Generator', free: false, pro: true, enterprise: true },
      ],
    },
    {
      category: 'Security & Enterprise Infrastructure',
      items: [
        { name: 'Edge Latency SLA', free: false, pro: true, enterprise: true },
        { name: 'SAML 2.0 / Okta / SCIM Provisioning', free: false, pro: false, enterprise: true },
        { name: 'Granular Tenant RBAC & Audit Logs', free: false, pro: false, enterprise: true },
        { name: 'Malware & Phishing Threat Shield', free: true, pro: true, enterprise: true },
        { name: 'Uptime SLA Guarantee', free: 'Best Effort', pro: '99.9%', enterprise: '99.99%' },
      ],
    },
  ];

  return (
    <div className="mx-auto w-full max-w-6xl space-y-16">
      {/* Billing Switcher Toggle */}
      <div className="flex flex-col items-center justify-center gap-3">
        <div className="inline-flex items-center rounded-full border border-zinc-200 bg-zinc-100/80 p-1 dark:border-zinc-800 dark:bg-zinc-900/80">
          <button
            type="button"
            onClick={() => setBillingInterval('monthly')}
            className={cn(
              'rounded-full px-4 py-1.5 text-xs font-medium transition-all',
              billingInterval === 'monthly'
                ? 'bg-white text-zinc-900 shadow-xs dark:bg-zinc-800 dark:text-zinc-100 font-semibold'
                : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200'
            )}
          >
            Monthly
          </button>
          <button
            type="button"
            onClick={() => setBillingInterval('annual')}
            className={cn(
              'inline-flex items-center gap-1.5 rounded-full px-4 py-1.5 text-xs font-medium transition-all',
              billingInterval === 'annual'
                ? 'bg-white text-zinc-900 shadow-xs dark:bg-zinc-800 dark:text-zinc-100 font-semibold'
                : 'text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200'
            )}
          >
            <span>Annual</span>
            <span className="rounded-full bg-emerald-100 px-1.5 py-0.2 text-[10px] font-bold text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300">
              Save 20%
            </span>
          </button>
        </div>
      </div>

      {/* Tier Cards Grid */}
      <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
        {tiers.map((tier) => {
          const currentPrice =
            billingInterval === 'annual'
              ? tier.price.annual
              : tier.price.monthly;

          return (
            <div
              key={tier.id}
              className={cn(
                'relative flex flex-col justify-between rounded-2xl border p-6 transition-all duration-200',
                tier.popular
                  ? 'border-zinc-900 bg-white shadow-xl shadow-zinc-200/50 dark:border-zinc-100 dark:bg-zinc-950 dark:shadow-none'
                  : 'border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-950/60'
              )}
            >
              {tier.popular && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                  <Badge variant="zinc" size="sm" className="font-semibold shadow-xs">
                    {tier.badge}
                  </Badge>
                </div>
              )}

              <div>
                <div className="flex items-center justify-between">
                  <h3 className="text-base font-semibold text-zinc-900 dark:text-zinc-100">
                    {tier.name}
                  </h3>
                  {!tier.popular && (
                    <span className="text-[11px] font-mono text-zinc-400">
                      {tier.badge}
                    </span>
                  )}
                </div>

                <p className="mt-2 min-h-[36px] text-xs text-zinc-500 dark:text-zinc-400">
                  {tier.description}
                </p>

                <div className="mt-4 flex items-baseline gap-1.5 border-b border-zinc-100 pb-4 dark:border-zinc-900">
                  <span className="font-mono text-3xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
                    {currentPrice}
                  </span>
                  <span className="text-xs text-zinc-400">
                    {tier.price.monthly === '$0' ? 'forever' : '/ month'}
                  </span>
                </div>

                <div className="mt-6 space-y-2.5">
                  <div className="text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
                    What's included
                  </div>
                  {tier.features.map((feat, idx) => (
                    <div key={idx} className="flex items-center gap-2 text-xs text-zinc-700 dark:text-zinc-300">
                      <Check className="h-3.5 w-3.5 shrink-0 text-emerald-600 dark:text-emerald-400" />
                      <span>{feat}</span>
                    </div>
                  ))}
                </div>
              </div>

              <div className="mt-8">
                <Link to="/sign-up" className="block w-full">
                  <Button
                    variant={tier.ctaVariant}
                    size="md"
                    className="w-full justify-center font-semibold"
                  >
                    {tier.cta}
                  </Button>
                </Link>
              </div>
            </div>
          );
        })}
      </div>

      {/* Feature Comparison Matrix Section */}
      <div className="space-y-6 pt-12">
        <div className="text-center">
          <h3 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
            Feature Comparison
          </h3>
          <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
            Comprehensive breakdown of infrastructure capacities across each subscription tier.
          </p>
        </div>

        <div className="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead>
                <tr className="border-b border-zinc-200 bg-zinc-50/75 dark:border-zinc-800 dark:bg-zinc-900/50">
                  <th className="w-2/5 p-4 text-xs font-semibold uppercase tracking-wider text-zinc-500 dark:text-zinc-400">
                    Capability
                  </th>
                  <th className="w-1/5 p-4 text-center text-xs font-semibold text-zinc-900 dark:text-zinc-100">
                    Free
                  </th>
                  <th className="w-1/5 p-4 text-center text-xs font-semibold text-zinc-900 dark:text-zinc-100">
                    Pro
                  </th>
                  <th className="w-1/5 p-4 text-center text-xs font-semibold text-zinc-900 dark:text-zinc-100">
                    Enterprise
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-100 dark:divide-zinc-900">
                {comparisonRows.map((cat, catIdx) => (
                  <React.Fragment key={catIdx}>
                    <tr className="bg-zinc-50/50 dark:bg-zinc-900/30">
                      <td
                        colSpan={4}
                        className="px-4 py-2 font-mono text-[11px] font-bold uppercase tracking-wider text-zinc-500 dark:text-zinc-400"
                      >
                        {cat.category}
                      </td>
                    </tr>
                    {cat.items.map((item, itemIdx) => (
                      <tr key={itemIdx} className="hover:bg-zinc-50/60 dark:hover:bg-zinc-900/40">
                        <td className="px-4 py-3 font-medium text-zinc-700 dark:text-zinc-300">
                          {item.name}
                        </td>
                        <td className="px-4 py-3 text-center">
                          {typeof item.free === 'boolean' ? (
                            item.free ? (
                              <Check className="mx-auto h-4 w-4 text-emerald-600 dark:text-emerald-400" />
                            ) : (
                              <Minus className="mx-auto h-4 w-4 text-zinc-300 dark:text-zinc-700" />
                            )
                          ) : (
                            <span className="text-zinc-600 dark:text-zinc-400">{item.free}</span>
                          )}
                        </td>
                        <td className="px-4 py-3 text-center font-medium">
                          {typeof item.pro === 'boolean' ? (
                            item.pro ? (
                              <Check className="mx-auto h-4 w-4 text-emerald-600 dark:text-emerald-400" />
                            ) : (
                              <Minus className="mx-auto h-4 w-4 text-zinc-300 dark:text-zinc-700" />
                            )
                          ) : (
                            <span className="text-zinc-900 dark:text-zinc-100">{item.pro}</span>
                          )}
                        </td>
                        <td className="px-4 py-3 text-center font-medium">
                          {typeof item.enterprise === 'boolean' ? (
                            item.enterprise ? (
                              <Check className="mx-auto h-4 w-4 text-emerald-600 dark:text-emerald-400" />
                            ) : (
                              <Minus className="mx-auto h-4 w-4 text-zinc-300 dark:text-zinc-700" />
                            )
                          ) : (
                            <span className="text-zinc-900 dark:text-zinc-100 font-semibold">{item.enterprise}</span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </React.Fragment>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}

export default PricingMatrix;
