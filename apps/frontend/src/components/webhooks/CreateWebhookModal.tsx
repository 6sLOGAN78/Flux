import React, { useState } from 'react';
import { Globe, Plus, Shield, Zap } from 'lucide-react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';

export interface CreateWebhookModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: { url: string; events: string[] }) => void;
  isLoading?: boolean;
}

const AVAILABLE_EVENTS = [
  { id: 'link.created', label: 'link.created', description: 'Triggered whenever a new short link is created' },
  { id: 'link.updated', label: 'link.updated', description: 'Triggered when link destination or routing rules change' },
  { id: 'click.recorded', label: 'click.recorded', description: 'Triggered on high-throughput link redirect clicks' },
  { id: 'conversion.recorded', label: 'conversion.recorded', description: 'Triggered upon conversion pixel / attribution match' },
  { id: 'domain.verified', label: 'domain.verified', description: 'Triggered when custom domain SSL is provisioned' },
];

export function CreateWebhookModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading = false,
}: CreateWebhookModalProps) {
  const [url, setUrl] = useState('');
  const [selectedEvents, setSelectedEvents] = useState<string[]>([
    'link.created',
    'click.recorded',
  ]);

  const toggleEvent = (eventId: string) => {
    setSelectedEvents((prev) =>
      prev.includes(eventId)
        ? prev.filter((e) => e !== eventId)
        : [...prev, eventId]
    );
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!url.trim() || selectedEvents.length === 0) return;
    onSubmit({
      url: url.trim(),
      events: selectedEvents,
    });
    setUrl('');
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Register Webhook Endpoint"
      description="Subscribe external services to asynchronous platform events via signed HTTP POST requests."
      footer={
        <>
          <Button variant="outline" size="sm" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            size="sm"
            onClick={handleSubmit}
            isLoading={isLoading}
            disabled={!url.trim() || selectedEvents.length === 0}
          >
            Register Webhook
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Endpoint URL (HTTPS)"
          placeholder="https://api.yourdomain.com/webhooks/flux"
          type="url"
          required
          autoFocus
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          startIcon={<Globe className="h-4 w-4" />}
          description="Must be a publicly accessible HTTPS endpoint."
        />

        <div className="space-y-2">
          <label className="block text-xs font-medium text-zinc-700 dark:text-zinc-300">
            Event Subscriptions
          </label>
          <div className="space-y-2">
            {AVAILABLE_EVENTS.map((evt) => {
              const isChecked = selectedEvents.includes(evt.id);
              return (
                <label
                  key={evt.id}
                  className="flex items-start gap-3 rounded-xl border border-zinc-200 bg-zinc-50/50 p-3 transition-colors hover:bg-zinc-100/50 dark:border-zinc-800 dark:bg-zinc-900/40 dark:hover:bg-zinc-900 cursor-pointer"
                >
                  <input
                    type="checkbox"
                    checked={isChecked}
                    onChange={() => toggleEvent(evt.id)}
                    className="mt-0.5 h-4 w-4 rounded border-zinc-300 text-zinc-900 focus:ring-zinc-900 dark:border-zinc-700 dark:bg-zinc-900"
                  />
                  <div>
                    <div className="font-mono text-xs font-bold text-zinc-900 dark:text-zinc-100">
                      {evt.label}
                    </div>
                    <div className="text-[11px] text-zinc-500 dark:text-zinc-400">
                      {evt.description}
                    </div>
                  </div>
                </label>
              );
            })}
          </div>
        </div>
      </form>
    </Modal>
  );
}

export default CreateWebhookModal;
