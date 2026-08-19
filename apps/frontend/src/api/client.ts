import { initClient } from '@ts-rest/core';
import { apiContract } from '@flux/openapi';
import { env } from '@/config/env';

let currentAuthToken: string | null = null;

export function setAuthToken(token: string | null) {
  currentAuthToken = token;
}

export function getAuthToken(): string | null {
  return currentAuthToken;
}

export const apiClient = initClient(apiContract, {
  baseUrl: env.VITE_API_URL || 'http://localhost:8080',
  baseHeaders: {},
  api: async (args) => {
    const headers = new Headers(args.headers);
    if (currentAuthToken && !headers.has('Authorization')) {
      headers.set('Authorization', `Bearer ${currentAuthToken}`);
    }
    return fetch(args.path, {
      method: args.method,
      headers,
      body: args.body,
      credentials: args.credentials,
      signal: args.signal,
    });
  },
});
