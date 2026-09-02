import React, { useState } from 'react';
import { Globe, Plus, Check, AlertCircle } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { WebhookEndpointList } from '@/components/webhooks/WebhookEndpointList';
import { CreateWebhookModal } from '@/components/webhooks/CreateWebhookModal';
import { WebhookDeliveryHistory } from '@/components/webhooks/WebhookDeliveryHistory';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import { useAuth } from '@clerk/clerk-react';

export function WebhooksPage() {
  const { orgId } = useAuth();
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const [selectedWebhookId, setSelectedWebhookId] = useState<string | null>(null);

  const { data: webhooks, isLoading, isError } = useQuery({
    queryKey: ['webhooks', orgId],
    queryFn: async () => {
      const { body, status } = await apiClient.getWebhooks();
      if (status !== 200) throw new Error('Failed to fetch webhooks');
      return body;
    },
  });

  const createMutation = useMutation({
    mutationFn: async (data: { url: string; events: string[] }) => {
      const { body, status } = await apiClient.createWebhook({
        body: { endpoint_url: data.url, events: data.events },
      });
      if (status !== 201) throw new Error('Failed to create webhook');
      return body;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['webhooks', orgId] });
      setIsModalOpen(false);
      setNotice(`Webhook created successfully. Secret: ${data.secret}`);
      // Don't auto-hide notice so they can copy the secret!
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      const { status } = await apiClient.deleteWebhook({
        params: { id },
      });
      if (status !== 204) throw new Error('Failed to delete webhook');
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks', orgId] });
      setNotice('Webhook endpoint removed');
      setTimeout(() => setNotice(null), 3000);
      setSelectedWebhookId(null);
    },
  });

  const toggleMutation = useMutation({
    mutationFn: async ({ id, active }: { id: string; active: boolean }) => {
      const { status } = await apiClient.updateWebhook({
        params: { id },
        body: { active },
      });
      if (status !== 200) throw new Error('Failed to update webhook');
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks', orgId] });
    },
  });

  const handleAddEndpoint = (data: { url: string; events: string[] }) => {
    createMutation.mutate(data);
  };

  const handleDeleteEndpoint = (id: string) => {
    if (confirm('Are you sure you want to delete this webhook? Deliveries will stop immediately.')) {
      deleteMutation.mutate(id);
    }
  };

  const handleToggleEndpoint = (id: string, active: boolean) => {
    toggleMutation.mutate({ id, active });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              Outbound Webhooks
            </h1>
            <Badge variant="emerald" size="sm" dot>
              HMAC-SHA256 Signed
            </Badge>
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Dispatch asynchronous events to your backend microservices with automatic retries.
          </p>
        </div>

        <Button
          variant="primary"
          size="md"
          onClick={() => setIsModalOpen(true)}
          leftIcon={<Plus className="h-4 w-4" />}
          disabled={createMutation.isPending}
        >
          Add Endpoint
        </Button>
      </div>

      {notice && (
        <div className="flex items-start gap-2 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-semibold text-emerald-800 dark:border-emerald-900/50 dark:bg-emerald-950/30 dark:text-emerald-300 animate-in fade-in">
          <Check className="h-4 w-4 mt-0.5" />
          <div className="flex-1">
            <p>{notice}</p>
            {notice.includes('Secret:') && (
              <p className="mt-1 text-emerald-600 dark:text-emerald-400">
                Please copy this secret now. It will not be shown again.
              </p>
            )}
          </div>
          <button onClick={() => setNotice(null)} className="text-emerald-500 hover:text-emerald-700">×</button>
        </div>
      )}

      {isError && (
        <div className="flex items-center gap-2 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-xs font-semibold text-rose-800">
          <AlertCircle className="h-4 w-4" />
          <span>Failed to load webhooks.</span>
        </div>
      )}

      <WebhookEndpointList
        endpoints={webhooks || []}
        isLoading={isLoading}
        onDeleteEndpoint={handleDeleteEndpoint}
        onToggleEndpoint={handleToggleEndpoint}
        selectedId={selectedWebhookId}
        onSelect={setSelectedWebhookId}
      />

      {selectedWebhookId && (
        <WebhookDeliveryHistory
          webhookId={selectedWebhookId}
        />
      )}

      <CreateWebhookModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleAddEndpoint}
      />
    </div>
  );
}

export default WebhooksPage;
