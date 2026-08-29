import { z } from 'zod';

export const envSchema = z.object({
  VITE_API_URL: z.string().url().default('http://localhost:8080'),
  VITE_CLERK_PUBLISHABLE_KEY: z.string().optional().default('pk_test_placeholder'),
  VITE_APP_URL: z.string().url().default('http://localhost:3000'),
  VITE_SHORT_DOMAIN: z.string().default('localhost:3000'),
  VITE_ENABLE_CLICKHOUSE_STREAM: z.string().optional().default('true'),
  VITE_ENABLE_AI_INSIGHTS: z.string().optional().default('true'),
});

export type EnvConfig = z.infer<typeof envSchema>;

export const env = (() => {
  const raw = {
    VITE_API_URL: import.meta.env.VITE_API_URL,
    VITE_CLERK_PUBLISHABLE_KEY: import.meta.env.VITE_CLERK_PUBLISHABLE_KEY,
    VITE_APP_URL: import.meta.env.VITE_APP_URL,
    VITE_SHORT_DOMAIN: import.meta.env.VITE_SHORT_DOMAIN,
  };
  
  const result = envSchema.safeParse(raw);
  if (result.success) {
    return result.data;
  }
  
  // Fallback to defaults
  return envSchema.parse({
    VITE_API_URL: 'http://localhost:8080',
    VITE_CLERK_PUBLISHABLE_KEY: 'pk_test_placeholder',
    VITE_APP_URL: 'http://localhost:3000',
    VITE_SHORT_DOMAIN: 'localhost:3000',
  });
})();

export function getShortDomain(): string {
  return env.VITE_SHORT_DOMAIN || 'localhost:3000';
}
