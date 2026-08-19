import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import type { CreateLinkInput, UpdateLinkInput, BulkCategorizeInput } from '@flux/zod';

export const linksQueryKeys = {
  all: ['links'] as const,
  detail: (shortCode: string) => ['links', 'detail', shortCode] as const,
  metrics: (id: string) => ['links', 'metrics', id] as const,
};

function extractErrorMessage(body: unknown, fallback: string): string {
  if (body && typeof body === 'object' && 'error' in body) {
    return String((body as Record<string, unknown>).error);
  }
  return fallback;
}

export function useGetLink(shortCode: string, enabled: boolean = true) {
  return useQuery({
    queryKey: linksQueryKeys.detail(shortCode),
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
      queryClient.invalidateQueries({ queryKey: linksQueryKeys.all });
    },
  });
}

export function useUpdateLink() {
  const queryClient = useQueryClient();
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
      queryClient.invalidateQueries({ queryKey: linksQueryKeys.all });
      if (data && typeof data === 'object' && 'shortCode' in data) {
        queryClient.invalidateQueries({ queryKey: linksQueryKeys.detail(String(data.shortCode)) });
      }
    },
  });
}

export function useBulkCategorize() {
  const queryClient = useQueryClient();
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
      queryClient.invalidateQueries({ queryKey: linksQueryKeys.all });
    },
  });
}
