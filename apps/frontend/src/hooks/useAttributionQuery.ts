import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import { useAuth as useClerkAuth } from '@clerk/clerk-react';

export const attributionQueryKeys = {
  all: ['attribution'] as const,
  model: (orgId: string | null | undefined, from?: string, to?: string, model?: string) => 
    ['attribution', orgId, from, to, model] as const,
};

export function useAttributionQuery(from?: string, to?: string, model: string = 'linear') {
  const { orgId } = useClerkAuth();
  
  return useQuery({
    queryKey: attributionQueryKeys.model(orgId, from, to, model),
    queryFn: async () => {
      const response = await apiClient.getAnalyticsAttribution({
        query: { from, to, model }
      });
      if (response.status !== 200) {
        throw new Error('Failed to fetch attribution');
      }
      return response.body;
    },
  });
}
