import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import type { CreateLinkInput, UpdateLinkInput, BulkCategorizeInput } from '@flux/zod';
import { useAuth as useClerkAuth } from '@clerk/clerk-react';

export const linksQueryKeys = {
  all: (orgId: string | null | undefined) => ['links', orgId] as const,
  detail: (orgId: string | null | undefined, shortCode: string) => ['links', 'detail', orgId, shortCode] as const,
  metrics: (orgId: string | null | undefined, id: string) => ['links', 'metrics', orgId, id] as const,
};

function extractErrorMessage(body: unknown, fallback: string): string {
  if (body && typeof body === 'object' && 'error' in body) {
    return String((body as Record<string, unknown>).error);
  }
  return fallback;
}

export function useGetLink(shortCode: string, enabled: boolean = true) {
  const { orgId } = useClerkAuth();
  return useQuery({
    queryKey: linksQueryKeys.detail(orgId, shortCode),
    queryFn: async () => {
      const response = await apiClient.getLink({
        params: { shortCode },
      });
      if (response.status !== 200) {
        throw new Error(extractErrorMessage(response.body, 'Failed to fetch link'));
      }
      return response.body;
    },
    enabled: Boolean(shortCode) && enabled,
  });
}

export function useCreateLink() {
  const queryClient = useQueryClient();
  const { orgId } = useClerkAuth();
  return useMutation({
    mutationFn: async (input: CreateLinkInput) => {
      const response = await apiClient.createLink({
        body: input,
      });
      if (response.status !== 201) {
        throw new Error(extractErrorMessage(response.body, 'Failed to create link'));
      }
      return response.body;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: linksQueryKeys.all(orgId) });
    },
  });
}

export function useUpdateLink() {
  const queryClient = useQueryClient();
  const { orgId } = useClerkAuth();
  return useMutation({
    mutationFn: async ({ id, body }: { id: string; body: UpdateLinkInput }) => {
      const response = await apiClient.updateLink({
        params: { id },
        body,
      });
      if (response.status !== 200) {
        throw new Error(extractErrorMessage(response.body, 'Failed to update link'));
      }
      return response.body;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: linksQueryKeys.all(orgId) });
      if (data && typeof data === 'object' && 'shortCode' in data) {
        queryClient.invalidateQueries({ queryKey: linksQueryKeys.detail(orgId, String(data.shortCode)) });
      }
    },
  });
}

export function useBulkCategorize() {
  const queryClient = useQueryClient();
  const { orgId } = useClerkAuth();
  return useMutation({
    mutationFn: async (body: BulkCategorizeInput) => {
      const response = await apiClient.bulkCategorizeLinks({
        body,
      });
      if (response.status !== 200) {
        throw new Error('Failed to bulk categorize links');
      }
      return response.body;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: linksQueryKeys.all(orgId) });
    },
  });
}

export function useGetLinks(params?: { page?: string, limit?: string, search?: string }) {
  const { orgId } = useClerkAuth();
  return useQuery({
    queryKey: [...linksQueryKeys.all(orgId), params],
    queryFn: async () => {
      // @ts-ignore
      const response = await apiClient.getLinks({
        query: params || {},
      });
      if (response.status !== 200) {
        throw new Error(extractErrorMessage(response.body, 'Failed to fetch links'));
      }
      return response.body;
    },
  });
}
