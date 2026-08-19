import { describe, expect, it, beforeEach } from 'bun:test';
import { apiClient, setAuthToken, getAuthToken } from './client';

describe('ts-rest API Client & Auth Interceptor', () => {
  beforeEach(() => {
    setAuthToken(null);
  });

  it('manages auth token state correctly', () => {
    expect(getAuthToken()).toBeNull();
    setAuthToken('test-jwt-token-xyz');
    expect(getAuthToken()).toBe('test-jwt-token-xyz');
    setAuthToken(null);
    expect(getAuthToken()).toBeNull();
  });

  it('initializes apiClient with required contract routes', () => {
    expect(apiClient).toBeDefined();
    expect(typeof apiClient.getHealth).toBe('function');
    expect(typeof apiClient.createLink).toBe('function');
    expect(typeof apiClient.getLink).toBe('function');
    expect(typeof apiClient.updateLink).toBe('function');
    expect(typeof apiClient.bulkCategorizeLinks).toBe('function');
    expect(typeof apiClient.getAnalyticsSummary).toBe('function');
    expect(typeof apiClient.getLinkMetrics).toBe('function');
    expect(typeof apiClient.getAnalyticsStreamMetrics).toBe('function');
  });
});
