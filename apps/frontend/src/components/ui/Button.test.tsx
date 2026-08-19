import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { Button } from './Button';

describe('Button UI Primitive', () => {
  it('renders primary button with Notion & Dub solid styling', () => {
    const html = renderToString(
      <Button variant="primary">Create Short Link</Button>
    );

    expect(html).toContain('Create Short Link');
    expect(html).toContain('bg-zinc-900');
    expect(html).toContain('text-white');
    expect(html).toContain('dark:bg-zinc-100');
    expect(html).toContain('dark:text-zinc-900');
  });

  it('renders outline button with subtle hairline borders', () => {
    const html = renderToString(
      <Button variant="outline">Cancel</Button>
    );

    expect(html).toContain('Cancel');
    expect(html).toContain('border-zinc-200');
    expect(html).toContain('dark:border-zinc-800');
  });

  it('renders secondary button variant', () => {
    const html = renderToString(
      <Button variant="secondary">Duplicate</Button>
    );

    expect(html).toContain('Duplicate');
    expect(html).toContain('bg-zinc-100');
    expect(html).toContain('dark:bg-zinc-800');
  });

  it('renders ghost button variant', () => {
    const html = renderToString(
      <Button variant="ghost">Settings</Button>
    );

    expect(html).toContain('Settings');
    expect(html).toContain('hover:bg-zinc-100');
    expect(html).toContain('dark:hover:bg-zinc-800');
  });

  it('renders destructive button variant with red background', () => {
    const html = renderToString(
      <Button variant="destructive">Delete Link</Button>
    );

    expect(html).toContain('Delete Link');
    expect(html).toContain('bg-red-600');
  });

  it('renders loading state with spinner and disables button', () => {
    const html = renderToString(
      <Button isLoading>Saving Changes</Button>
    );

    expect(html).toContain('Saving Changes');
    expect(html).toContain('animate-spin');
    expect(html).toContain('disabled');
  });

  it('renders with left and right icon accessories', () => {
    const html = renderToString(
      <Button
        leftIcon={<span data-testid="icon-left">←</span>}
        rightIcon={<span data-testid="icon-right">→</span>}
      >
        Navigate
      </Button>
    );

    expect(html).toContain('data-testid="icon-left"');
    expect(html).toContain('data-testid="icon-right"');
    expect(html).toContain('Navigate');
  });

  it('renders different size classes properly', () => {
    const smHtml = renderToString(<Button size="sm">Small</Button>);
    const lgHtml = renderToString(<Button size="lg">Large</Button>);
    const iconHtml = renderToString(<Button size="icon">✕</Button>);

    expect(smHtml).toContain('h-8');
    expect(lgHtml).toContain('h-10');
    expect(iconHtml).toContain('w-9');
  });
});
