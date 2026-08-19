import React, { useState } from 'react';
import { Plus, Trash2, Filter } from 'lucide-react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';

export interface CreateFunnelModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: { name: string; steps: string[] }) => void;
  isLoading?: boolean;
}

export function CreateFunnelModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading = false,
}: CreateFunnelModalProps) {
  const [funnelName, setFunnelName] = useState('');
  const [steps, setSteps] = useState<string[]>([
    'Ad Link Click',
    'Landing Page View',
    'Sign Up Completed',
    'Subscription Checkout',
  ]);

  const handleAddStep = () => {
    setSteps((prev) => [...prev, `Step ${prev.length + 1}`]);
  };

  const handleStepChange = (index: number, value: string) => {
    setSteps((prev) => prev.map((s, i) => (i === index ? value : s)));
  };

  const handleRemoveStep = (index: number) => {
    if (steps.length <= 2) return;
    setSteps((prev) => prev.filter((_, i) => i !== index));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!funnelName.trim()) return;
    onSubmit({
      name: funnelName.trim(),
      steps: steps.filter((s) => s.trim().length > 0),
    });
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Create Conversion Funnel"
      description="Track step-by-step conversion rates and friction drop-offs across your growth loops."
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
            disabled={!funnelName.trim()}
          >
            Create Funnel
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Funnel Name"
          placeholder="e.g. Q3 Paid Search Acquisition Funnel"
          required
          autoFocus
          value={funnelName}
          onChange={(e) => setFunnelName(e.target.value)}
        />

        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <label className="text-xs font-medium text-zinc-700 dark:text-zinc-300">
              Funnel Milestones / Steps
            </label>
            <button
              type="button"
              onClick={handleAddStep}
              className="inline-flex items-center gap-1 text-[11px] font-medium text-zinc-900 hover:underline dark:text-zinc-100"
            >
              <Plus className="h-3 w-3" />
              <span>Add Funnel Step</span>
            </button>
          </div>

          <div className="space-y-2">
            {steps.map((step, idx) => (
              <div key={idx} className="flex items-center gap-2">
                <span className="flex h-5 w-5 items-center justify-center rounded bg-zinc-100 font-mono text-[10px] font-bold text-zinc-500 dark:bg-zinc-800">
                  {idx + 1}
                </span>
                <input
                  type="text"
                  value={step}
                  onChange={(e) => handleStepChange(idx, e.target.value)}
                  className="h-8 flex-1 rounded-lg border border-zinc-200 bg-white px-3 text-xs text-zinc-900 focus:border-zinc-400 focus:outline-none dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-100"
                  placeholder={`Step ${idx + 1} milestone`}
                  required
                />
                {steps.length > 2 && (
                  <button
                    type="button"
                    onClick={() => handleRemoveStep(idx)}
                    className="p-1 text-zinc-400 hover:text-red-600 dark:hover:text-red-400"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                )}
              </div>
            ))}
          </div>
        </div>
      </form>
    </Modal>
  );
}

export default CreateFunnelModal;
