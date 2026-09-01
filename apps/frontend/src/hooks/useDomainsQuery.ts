import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import type { CreateDomainInput, CustomDomain } from '@flux/zod';
import { useAuth as useClerkAuth } from '@clerk/clerk-react';

export const domainsQueryKeys = {
  all: (orgId: string | null | undefined) => ['domains', orgId] as const,
  detail: (orgId: string | null | undefined, id: string) => ['domains', 'detail', orgId, id] as const,
};

function extractErrorMessage(body: unknown, fallback: string): string {
  if (body && typeof body === 'object' && 'error' in body) {
    return String((body as Record<string, unknown>).error);
  }
  return fallback;
}

export function useGetDomains() {
  const { orgId } = useClerkAuth();
  return useQuery({
    queryKey: domainsQueryKeys.all(orgId),
    queryFn: async () => {
      const response = await apiClient.getDomains();
      if (response.status !== 200) {
        throw new Error(extractErrorMessage(response.body, 'Failed to fetch domains'));
      }
      return response.body.data;
    },
  });
}

export function useCreateDomain() {
  const queryClient = useQueryClient();
  const { orgId } = useClerkAuth();
  return useMutation({
    mutationFn: async (input: CreateDomainInput) => {
      const response = await apiClient.createDomain({
        body: input,
      });
      if (response.status !== 201) {
        throw new Error(extractErrorMessage(response.body, 'Failed to create custom domain'));
      }
      return response.body;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: domainsQueryKeys.all(orgId) });
    },
  });
}

export function useDeleteDomain() {
  const queryClient = useQueryClient();
  const { orgId } = useClerkAuth();
  return useMutation({
    mutationFn: async (id: string) => {
      const response = await apiClient.deleteDomain({
        params: { id },
        body: undefined,
      });
      if (response.status !== 204) {
        throw new Error('Failed to delete custom domain');
      }
      return true;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: domainsQueryKeys.all(orgId) });
    },
  });
}
