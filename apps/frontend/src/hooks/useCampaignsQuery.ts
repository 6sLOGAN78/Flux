import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import type { CreateCampaignInput } from '@flux/zod';
import { useAuth as useClerkAuth } from '@clerk/clerk-react';

export const campaignsQueryKeys = {
  all: (orgId: string | null | undefined) => ['campaigns', orgId] as const,
  detail: (orgId: string | null | undefined, id: string) => ['campaigns', 'detail', orgId, id] as const,
};

function extractErrorMessage(body: unknown, fallback: string): string {
  if (body && typeof body === 'object' && 'error' in body) {
    return String((body as Record<string, unknown>).error);
  }
  return fallback;
}

export function useGetCampaigns() {
  const { orgId } = useClerkAuth();
  return useQuery({
    queryKey: campaignsQueryKeys.all(orgId),
    queryFn: async () => {
      const response = await apiClient.getCampaigns();
      if (response.status !== 200) {
        throw new Error(extractErrorMessage(response.body, 'Failed to fetch campaigns'));
      }
      return response.body;
    },
  });
}

export function useCreateCampaign() {
  const queryClient = useQueryClient();
  const { orgId } = useClerkAuth();
  return useMutation({
    mutationFn: async (input: CreateCampaignInput) => {
      const response = await apiClient.createCampaign({
        body: input,
      });
      if (response.status !== 201) {
        throw new Error(extractErrorMessage(response.body, 'Failed to create campaign'));
      }
      return response.body;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: campaignsQueryKeys.all(orgId) });
    },
  });
}

export function useUpdateCampaign() {
  const queryClient = useQueryClient();
  const { orgId } = useClerkAuth();
  return useMutation({
    mutationFn: async ({ id, body }: { id: string; body: Partial<CreateCampaignInput> }) => {
      const response = await apiClient.updateCampaign({
        params: { id },
        body,
      });
      if (response.status !== 200) {
        throw new Error(extractErrorMessage(response.body, 'Failed to update campaign'));
      }
      return response.body;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: campaignsQueryKeys.all(orgId) });
      if (data && typeof data === 'object' && 'id' in data) {
        queryClient.invalidateQueries({ queryKey: campaignsQueryKeys.detail(orgId, String(data.id)) });
      }
    },
  });
}

export function useDeleteCampaign() {
  const queryClient = useQueryClient();
  const { orgId } = useClerkAuth();
  return useMutation({
    mutationFn: async (id: string) => {
      const response = await apiClient.deleteCampaign({
        params: { id },
      });
      if (response.status !== 204) {
        throw new Error('Failed to delete campaign');
      }
      return true;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: campaignsQueryKeys.all(orgId) });
    },
  });
}
