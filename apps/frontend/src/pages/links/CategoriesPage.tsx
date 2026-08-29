import React, { useState } from 'react';
import { Plus, Tag, Palette } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';
import { CategoryGrid, CategoryItem } from '@/components/categories/CategoryGrid';

const INITIAL_CATEGORIES: CategoryItem[] = [];

const COLOR_PRESETS = [
  '#10b981', // emerald
  '#3b82f6', // blue
  '#8b5cf6', // violet
  '#f59e0b', // amber
  '#ef4444', // red
  '#09090b', // zinc
];

export function CategoriesPage() {
  const [categories, setCategories] = useState<CategoryItem[]>(INITIAL_CATEGORIES);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingCategory, setEditingCategory] = useState<CategoryItem | null>(null);

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [color, setColor] = useState('#10b981');

  const handleOpenCreate = () => {
    setEditingCategory(null);
    setName('');
    setDescription('');
    setColor('#10b981');
    setIsModalOpen(true);
  };

  const handleOpenEdit = (cat: CategoryItem) => {
    setEditingCategory(cat);
    setName(cat.name);
    setDescription(cat.description || '');
    setColor(cat.color);
    setIsModalOpen(true);
  };

  const handleDelete = (id: string) => {
    setCategories((prev) => prev.filter((c) => c.id !== id));
  };

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    if (editingCategory) {
      setCategories((prev) =>
        prev.map((c) =>
          c.id === editingCategory.id
            ? { ...c, name: name.trim(), description, color }
            : c
        )
      );
    } else {
      const newCat: CategoryItem = {
        id: `cat_${Date.now()}`,
        name: name.trim(),
        description,
        color,
        linkCount: 0,
      };
      setCategories((prev) => [newCat, ...prev]);
    }

    setIsModalOpen(false);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
            Categories
          </h1>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Organize and segment your links with custom tags and color swatches.
          </p>
        </div>

        <Button
          variant="primary"
          size="md"
          onClick={handleOpenCreate}
          leftIcon={<Plus className="h-4 w-4" />}
        >
          New Category
        </Button>
      </div>

      {/* Grid */}
      <CategoryGrid
        categories={categories}
        onEdit={handleOpenEdit}
        onDelete={handleDelete}
      />

      {/* Create / Edit Category Modal */}
      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title={editingCategory ? 'Edit Category' : 'Create Category'}
        description="Categorize links for streamlined filtering and team RBAC."
        footer={
          <>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setIsModalOpen(false)}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              size="sm"
              onClick={handleSave}
              disabled={!name.trim()}
            >
              {editingCategory ? 'Save Changes' : 'Create Category'}
            </Button>
          </>
        }
      >
        <form onSubmit={handleSave} className="space-y-4">
          <Input
            label="Category Name"
            placeholder="e.g. Paid Ads, Investor Deck, Podcasts"
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
          />

          <Input
            label="Description (Optional)"
            placeholder="Brief explanation of this link category"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />

          <div>
            <label className="mb-1.5 block text-xs font-medium text-zinc-700 dark:text-zinc-300">
              Color Swatch
            </label>
            <div className="flex items-center gap-2">
              {COLOR_PRESETS.map((c) => (
                <button
                  key={c}
                  type="button"
                  onClick={() => setColor(c)}
                  style={{ backgroundColor: c }}
                  className={`h-7 w-7 rounded-full transition-transform ${
                    color === c ? 'scale-110 ring-2 ring-zinc-900 ring-offset-2 dark:ring-zinc-100' : 'hover:scale-105'
                  }`}
                />
              ))}
            </div>
          </div>
        </form>
      </Modal>
    </div>
  );
}

export default CategoriesPage;
