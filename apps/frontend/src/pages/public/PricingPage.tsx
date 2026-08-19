import React from 'react';
import { Link } from 'react-router-dom';
import { PublicHeader } from '@/components/public/PublicHeader';
import { PublicFooter } from '@/components/public/PublicFooter';
import { PricingMatrix } from '@/components/public/PricingMatrix';
import { Button } from '@/components/ui/Button';
import { ArrowRight, HelpCircle } from 'lucide-react';

export function PricingPage() {
  const faqs = [
    {
      q: 'How does the sub-10ms edge redirect SLA work?',
      a: 'Flux deploys redirect route rules to 300+ global edge Anycast nodes with local in-memory Redis caches. Inbound DNS lookups resolve to the geographically closest node, executing the 301/308 redirect in under 10 milliseconds.',
    },
    {
      q: 'Can I connect multiple custom domains on the Pro tier?',
      a: 'Yes, Pro includes up to 5 custom branded domains (e.g. go.yourcompany.com, link.brand.io) with automated TLS/SSL certificate issuance and verification.',
    },
    {
      q: 'How does ClickHouse analytics scale for high-volume campaigns?',
      a: 'Flux uses ClickHouse columnar storage for click event ingestion. Even with hundreds of millions of events, aggregation queries (filtering by UTM tag, country, or referrer) return in under 50ms.',
    },
    {
      q: 'Do you support SAML SSO and SCIM directory sync?',
      a: 'Enterprise plans support native SAML 2.0 and OIDC integrations for Okta, Azure AD (Entra ID), and Google Workspace, along with SCIM automated user de-provisioning.',
    },
  ];

  return (
    <div className="flex min-h-screen flex-col bg-white text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
      <PublicHeader />

      {/* Pricing Header */}
      <section className="px-4 pt-12 pb-8 sm:px-6 md:pt-16">
        <div className="mx-auto max-w-3xl text-center">
          <div className="inline-flex items-center gap-2 rounded-full border border-zinc-200 bg-zinc-50 px-3 py-1 text-xs font-medium text-zinc-700 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-300">
            Transparent SaaS & Enterprise Pricing
          </div>
          <h1 className="mt-4 text-3xl font-bold tracking-tight sm:text-5xl">
            Simple, predictable pricing
          </h1>
          <p className="mx-auto mt-3 max-w-lg text-sm text-zinc-500 dark:text-zinc-400">
            Scale from your first link to billions of monthly redirects with zero hidden fees.
          </p>
        </div>
      </section>

      {/* Pricing Matrix Section */}
      <section className="px-4 pb-20 sm:px-6">
        <PricingMatrix />
      </section>

      {/* FAQ Section */}
      <section className="border-t border-zinc-200 bg-zinc-50/50 px-4 py-16 dark:border-zinc-800 dark:bg-zinc-900/30 sm:px-6">
        <div className="mx-auto max-w-4xl space-y-8">
          <div className="text-center">
            <h2 className="text-xl font-bold tracking-tight sm:text-2xl">
              Frequently Asked Questions
            </h2>
            <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
              Everything you need to know about billing, limits, and edge performance.
            </p>
          </div>

          <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
            {faqs.map((faq, i) => (
              <div
                key={i}
                className="rounded-xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950"
              >
                <h3 className="text-xs font-semibold text-zinc-900 dark:text-zinc-100">
                  {faq.q}
                </h3>
                <p className="mt-2 text-xs text-zinc-500 leading-relaxed dark:text-zinc-400">
                  {faq.a}
                </p>
              </div>
            ))}
          </div>

          <div className="rounded-xl border border-zinc-200 bg-white p-6 text-center shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
            <h3 className="text-sm font-semibold">Have custom enterprise requirements?</h3>
            <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
              We offer dedicated VPC deployments, custom SLA contracts, and high-volume billing discounts.
            </p>
            <div className="mt-4 flex justify-center gap-3">
              <Link to="/sso">
                <Button variant="primary" size="sm">
                  Contact Enterprise Sales
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </section>

      <PublicFooter />
    </div>
  );
}

export default PricingPage;
