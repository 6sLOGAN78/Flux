import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import { useAuth as useClerkAuth } from '@clerk/clerk-react';

export const analyticsQueryKeys = {
  all: ['analytics'] as const,
  summary: (orgId: string | null | undefined, from?: string, to?: string) => ['analytics', 'summary', orgId, from, to] as const,
  timeseries: (orgId: string | null | undefined, from?: string, to?: string, interval?: string) => ['analytics', 'timeseries', orgId, from, to, interval] as const,
  topLinks: (orgId: string | null | undefined, from?: string, to?: string) => ['analytics', 'topLinks', orgId, from, to] as const,
  referrers: (orgId: string | null | undefined, from?: string, to?: string) => ['analytics', 'referrers', orgId, from, to] as const,
  linkMetrics: (orgId: string | null | undefined, id: string) => ['analytics', 'link', orgId, id] as const,
  streamMetrics: (orgId: string | null | undefined) => ['analytics', 'stream', orgId] as const,
};

function extractErrorMessage(body: unknown, fallback: string): string {
  if (body && typeof body === 'object' && 'error' in body) {
    return String((body as Record<string, unknown>).error);
  }
  return fallback;
}

export function useAnalyticsSummary(from?: string, to?: string) {
  const { orgId } = useClerkAuth();
  return useQuery({
    queryKey: analyticsQueryKeys.summary(orgId, from, to),
    queryFn: async () => {
      const response = await apiClient.getAnalyticsSummary({
        query: { from, to }
      });
      if (response.status !== 200) {
        throw new Error('Failed to fetch analytics summary');
      }
      return response.body;
    },
  });
}

export function useAnalyticsTimeseries(from?: string, to?: string, interval?: string) {
  const { orgId } = useClerkAuth();
  return useQuery({
    queryKey: analyticsQueryKeys.timeseries(orgId, from, to, interval),
    queryFn: async () => {
      const response = await apiClient.getAnalyticsTimeseries({
        query: { from, to, interval }
      });
      if (response.status !== 200) {
        throw new Error('Failed to fetch analytics timeseries');
      }
      return response.body;
    },
  });
}

export function useAnalyticsTopLinks(from?: string, to?: string) {
  const { orgId } = useClerkAuth();
  return useQuery({
    queryKey: analyticsQueryKeys.topLinks(orgId, from, to),
    queryFn: async () => {
      const response = await apiClient.getAnalyticsTopLinks({
        query: { from, to }
      });
      if (response.status !== 200) {
        throw new Error('Failed to fetch top links');
      }
      return response.body;
    },
  });
}

export function useAnalyticsReferrers(from?: string, to?: string) {
  const { orgId } = useClerkAuth();
  return useQuery({
    queryKey: analyticsQueryKeys.referrers(orgId, from, to),
    queryFn: async () => {
      const response = await apiClient.getAnalyticsReferrers({
        query: { from, to }
      });
      if (response.status !== 200) {
        throw new Error('Failed to fetch referrers');
      }
      return response.body;
    },
  });
}

export function useLinkMetrics(id: string, enabled: boolean = true) {
  const { orgId } = useClerkAuth();
  return useQuery({
    queryKey: analyticsQueryKeys.linkMetrics(orgId, id),
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
  const { orgId } = useClerkAuth();
  return useQuery({
    queryKey: analyticsQueryKeys.streamMetrics(orgId),
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

export function useAnalyticsCampaigns(from?: string, to?: string) {
  const { orgId } = useClerkAuth();
  return useQuery({
    queryKey: ['analytics', 'campaigns', orgId, from, to] as const,
    queryFn: async () => {
      const response = await apiClient.getAnalyticsCampaigns({
        query: { from, to }
      });
      if (response.status !== 200) {
        throw new Error('Failed to fetch campaign analytics');
      }
      return response.body;
    },
  });
}

export function useAnalyticsUTM(dimension: string, from?: string, to?: string) {
  const { orgId } = useClerkAuth();
  return useQuery({
    queryKey: ['analytics', 'utm', dimension, orgId, from, to] as const,
    queryFn: async () => {
      const response = await apiClient.getAnalyticsUTM({
        query: { dimension, from, to }
      });
      if (response.status !== 200) {
        throw new Error('Failed to fetch UTM analytics');
      }
      return response.body;
    },
    enabled: Boolean(dimension),
  });
}
