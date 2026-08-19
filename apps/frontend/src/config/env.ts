import { z } from 'zod';

export const envSchema = z.object({
  VITE_API_URL: z.string().url({ message: 'VITE_API_URL must be a valid URL' }).default('http://localhost:8080'),
  VITE_CLERK_PUBLISHABLE_KEY: z.string().optional().default('pk_test_placeholder'),
  VITE_APP_URL: z.string().url({ message: 'VITE_APP_URL must be a valid URL' }).default('http://localhost:5173'),
  VITE_ENABLE_CLICKHOUSE_STREAM: z.string().optional().default('true'),
  VITE_ENABLE_AI_INSIGHTS: z.string().optional().default('true'),
});

export type EnvConfig = z.infer<typeof envSchema>;

export function validateEnv(rawEnv: Record<string, unknown> = {}): EnvConfig {
  const result = envSchema.safeParse(rawEnv);
  if (!result.success) {
    const errorDetails = result.error.format();
    console.error('❌ Environment validation failed:', JSON.stringify(errorDetails, null, 2));
    throw new Error(`Invalid environment variables: ${JSON.stringify(result.error.flatten().fieldErrors)}`);
  }
  return result.data;
}

export const env = (() => {
  try {
    const raw = typeof import.meta !== 'undefined' && import.meta.env ? import.meta.env : {};
    return validateEnv(raw);
  } catch {
    return envSchema.parse({
      VITE_API_URL: 'http://localhost:8080',
      VITE_CLERK_PUBLISHABLE_KEY: 'pk_test_placeholder',
      VITE_APP_URL: 'http://localhost:5173',
    });
  }
})();
