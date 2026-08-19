import React, { useState } from 'react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';

export interface BulkCategorizeModalProps {
  isOpen: boolean;
  selectedCount: number;
  onClose: () => void;
  onCategorize: (category: string) => void;
  isLoading?: boolean;
}

export function BulkCategorizeModal({
  isOpen,
  selectedCount,
  onClose,
  onCategorize,
  isLoading = false,
}: BulkCategorizeModalProps) {
  const [categoryName, setCategoryName] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!categoryName.trim()) return;
    onCategorize(categoryName.trim());
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={`Categorize ${selectedCount} links`}
      description="Apply a common category tag to all selected short links."
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
            disabled={!categoryName.trim()}
          >
            Apply Category
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-3">
        <Input
          label="Category Name"
          placeholder="e.g. Marketing, Changelog, Partner"
          value={categoryName}
          onChange={(e) => setCategoryName(e.target.value)}
          autoFocus
          required
        />
      </form>
    </Modal>
  );
}

export default BulkCategorizeModal;
