import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { Modal } from './Modal';

describe('Modal UI Primitive', () => {
  it('renders modal content with backdrop when isOpen is true', () => {
    const html = renderToString(
      <Modal
        isOpen={true}
        onClose={() => {}}
        title="Create New Link"
        description="Configure your custom branded short link"
      >
        <div data-testid="modal-body">Modal Body Content</div>
      </Modal>
    );

    expect(html).toContain('Create New Link');
    expect(html).toContain('Configure your custom branded short link');
    expect(html).toContain('Modal Body Content');
    expect(html).toContain('backdrop-blur-xs');
  });

  it('renders footer actions when provided', () => {
    const html = renderToString(
      <Modal
        isOpen={true}
        onClose={() => {}}
        title="Confirm Deletion"
        footer={<button data-testid="confirm-btn">Confirm</button>}
      >
        <div>Are you sure?</div>
      </Modal>
    );

    expect(html).toContain('data-testid="confirm-btn"');
    expect(html).toContain('Confirm');
  });

  it('does not render dialog elements when isOpen is false', () => {
    const html = renderToString(
      <Modal
        isOpen={false}
        onClose={() => {}}
        title="Hidden Modal"
      >
        <div>Hidden content</div>
      </Modal>
    );

    expect(html).not.toContain('Hidden Modal');
    expect(html).not.toContain('Hidden content');
  });
});
