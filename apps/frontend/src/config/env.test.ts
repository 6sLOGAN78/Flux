import { describe, expect, it } from 'bun:test';
import { envSchema, validateEnv } from './env';

describe('Frontend Environment Validator', () => {
  it('successfully validates and returns default values when minimal input provided', () => {
    const parsed = validateEnv({
      VITE_API_URL: 'http://localhost:8080',
      VITE_APP_URL: 'http://localhost:5173',
    });

    expect(parsed.VITE_API_URL).toBe('http://localhost:8080');
    expect(parsed.VITE_APP_URL).toBe('http://localhost:5173');
    expect(parsed.VITE_ENABLE_CLICKHOUSE_STREAM).toBe('true');
    expect(parsed.VITE_ENABLE_AI_INSIGHTS).toBe('true');
  });

  it('correctly accepts custom environment variables', () => {
    const custom = {
      VITE_API_URL: 'https://api.flux.dev',
      VITE_CLERK_PUBLISHABLE_KEY: 'pk_live_custom123',
      VITE_APP_URL: 'https://app.flux.dev',
      VITE_ENABLE_CLICKHOUSE_STREAM: 'false',
      VITE_ENABLE_AI_INSIGHTS: 'false',
    };

    const parsed = validateEnv(custom);
    expect(parsed.VITE_API_URL).toBe('https://api.flux.dev');
    expect(parsed.VITE_CLERK_PUBLISHABLE_KEY).toBe('pk_live_custom123');
    expect(parsed.VITE_APP_URL).toBe('https://app.flux.dev');
    expect(parsed.VITE_ENABLE_CLICKHOUSE_STREAM).toBe('false');
    expect(parsed.VITE_ENABLE_AI_INSIGHTS).toBe('false');
  });

  it('rejects invalid API URL with validation error', () => {
    expect(() => {
      validateEnv({
        VITE_API_URL: 'not-a-valid-url',
        VITE_APP_URL: 'http://localhost:5173',
      });
    }).toThrow();
  });

  it('rejects invalid App URL with validation error', () => {
    expect(() => {
      validateEnv({
        VITE_API_URL: 'http://localhost:8080',
        VITE_APP_URL: 'invalid-app-url',
      });
    }).toThrow();
  });
});
