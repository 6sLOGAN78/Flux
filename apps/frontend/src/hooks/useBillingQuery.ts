import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import { useAuth as useClerkAuth } from '@clerk/clerk-react';

function extractErrorMessage(body: unknown, fallback: string): string {
  if (body && typeof body === 'object' && 'error' in body) {
    return String((body as Record<string, unknown>).error);
  }
  return fallback;
}

export const billingQueryKeys = {
  all: (orgId?: string | null) => ['billing', orgId] as const,
  subscription: (orgId?: string | null) => [...billingQueryKeys.all(orgId), 'subscription'] as const,
};

export function useGetSubscription() {
  const { orgId } = useClerkAuth();
  
  return useQuery({
    queryKey: billingQueryKeys.subscription(orgId),
    queryFn: async () => {
      const response = await apiClient.getSubscription();
      if (response.status !== 200) {
        throw new Error(extractErrorMessage(response.body, 'Failed to fetch subscription'));
      }
      return response.body;
    },
    staleTime: 5 * 60 * 1000,
  });
}

export function useCreateCustomerPortal() {
  return useMutation({
    mutationFn: async () => {
      const response = await apiClient.createCustomerPortal({
        body: {},
      });
      if (response.status !== 200) {
        throw new Error(extractErrorMessage(response.body, 'Failed to create Customer Portal session'));
      }
      return response.body;
    },
  });
}
