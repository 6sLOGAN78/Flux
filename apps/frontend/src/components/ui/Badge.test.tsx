import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { Badge } from './Badge';

describe('Badge UI Primitive', () => {
  it('renders default zinc badge', () => {
    const html = renderToString(<Badge>Active</Badge>);

    expect(html).toContain('Active');
    expect(html).toContain('bg-zinc-100');
    expect(html).toContain('text-zinc-800');
  });

  it('renders emerald success badge', () => {
    const html = renderToString(<Badge variant="emerald">Live</Badge>);

    expect(html).toContain('Live');
    expect(html).toContain('bg-emerald-50');
    expect(html).toContain('text-emerald-700');
  });

  it('renders blue info badge', () => {
    const html = renderToString(<Badge variant="blue">Enterprise</Badge>);

    expect(html).toContain('Enterprise');
    expect(html).toContain('bg-blue-50');
    expect(html).toContain('text-blue-700');
  });

  it('renders badge with status indicator dot', () => {
    const html = renderToString(<Badge dot variant="emerald">Healthy</Badge>);

    expect(html).toContain('Healthy');
    expect(html).toContain('rounded-full');
  });

  it('renders outline badge variant', () => {
    const html = renderToString(<Badge variant="outline">Custom Domain</Badge>);

    expect(html).toContain('Custom Domain');
    expect(html).toContain('border-zinc-200');
  });
});
