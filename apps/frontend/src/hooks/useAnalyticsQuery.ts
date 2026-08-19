import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/api/client';

export const analyticsQueryKeys = {
  all: ['analytics'] as const,
  summary: () => ['analytics', 'summary'] as const,
  linkMetrics: (id: string) => ['analytics', 'link', id] as const,
  streamMetrics: () => ['analytics', 'stream'] as const,
};

function extractErrorMessage(body: unknown, fallback: string): string {
  if (body && typeof body === 'object' && 'error' in body) {
    return String((body as Record<string, unknown>).error);
  }
  return fallback;
}

export function useAnalyticsSummary() {
  return useQuery({
    queryKey: analyticsQueryKeys.summary(),
    queryFn: async () => {
      const response = await apiClient.getAnalyticsSummary();
      if (response.status !== 200) {
        throw new Error('Failed to fetch analytics summary');
      }
      return response.body;
    },
  });
}

export function useLinkMetrics(id: string, enabled: boolean = true) {
  return useQuery({
    queryKey: analyticsQueryKeys.linkMetrics(id),
    queryFn: async () => {
      const response = await apiClient.getLinkMetrics({
        params: { id },
      });
      if (response.status !== 200) {
        throw new Error(extractErrorMessage(response.body, 'Failed to fetch link metrics'));
      }
      return response.body;
    },
    enabled: Boolean(id) && enabled,
  });
}

export function useStreamMetrics() {
  return useQuery({
    queryKey: analyticsQueryKeys.streamMetrics(),
    queryFn: async () => {
      const response = await apiClient.getAnalyticsStreamMetrics();
      if (response.status !== 200) {
        throw new Error('Failed to fetch stream metrics');
      }
      return response.body;
    },
    refetchInterval: 5000,
  });
}
