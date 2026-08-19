import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { Input } from './Input';

describe('Input UI Primitive', () => {
  it('renders standard input field with label and placeholder', () => {
    const html = renderToString(
      <Input
        label="Destination URL"
        placeholder="https://example.com/long-url"
        defaultValue="https://flux.to"
      />
    );

    expect(html).toContain('Destination URL');
    expect(html).toContain('placeholder="https://example.com/long-url"');
    expect(html).toContain('value="https://flux.to"');
    expect(html).toContain('border-zinc-200');
    expect(html).toContain('dark:border-zinc-800');
  });

  it('renders helper description when provided', () => {
    const html = renderToString(
      <Input
        label="Short Slug"
        description="The unique key used for redirecting your link"
      />
    );

    expect(html).toContain('The unique key used for redirecting your link');
  });

  it('renders error message and applies destructive border styles', () => {
    const html = renderToString(
      <Input
        label="Work Email"
        error="Please provide a valid company email address"
      />
    );

    expect(html).toContain('Please provide a valid company email address');
    expect(html).toContain('border-red-500');
    expect(html).toContain('text-red-500');
  });

  it('renders leading prefix text or icons properly', () => {
    const html = renderToString(
      <Input
        label="Custom Back-half"
        prefix="flux.to/"
        startIcon={<span data-testid="prefix-icon">🔗</span>}
      />
    );

    expect(html).toContain('flux.to/');
    expect(html).toContain('data-testid="prefix-icon"');
  });

  it('renders disabled state styling', () => {
    const html = renderToString(
      <Input
        label="Domain"
        disabled
        value="default.domain.com"
      />
    );

    expect(html).toContain('disabled');
    expect(html).toContain('disabled:opacity-50');
  });
});
