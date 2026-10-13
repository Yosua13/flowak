import { describe, expect, it, vi } from 'vitest';
import React from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { BoardFilters, Modal } from './KanbanView';
import type { KanbanFilters } from './kanbanModel';

describe('KanbanView components', () => {
  it('renders BoardFilters with compact search and filter controls', () => {
    const filters: KanbanFilters = {
      search: 'login',
      moduleId: 'all',
      assigneeId: 'all',
      type: 'all',
      facet: 'all',
      priority: 'all',
      includeCanceled: false,
    };
    const setFilters = vi.fn();
    const modules = [{ id: 'm1', name: 'Auth' }];
    const members = [{ id: 'u1', name: 'Yosua' }];

    const markup = renderToStaticMarkup(
      <BoardFilters
        filters={filters}
        setFilters={setFilters}
        modules={modules}
        members={members}
      />
    );

    // Search input exists with compact width and placeholder
    expect(markup).toContain('Cari task...');
    expect(markup).toContain('value="login"');
    // Reset button appears when active filters exist
    expect(markup).toContain('Reset');
    // Filter options are present
    expect(markup).toContain('Semua modul');
    expect(markup).toContain('Semua assignee');
    expect(markup).toContain('Canceled');
  });

  it('renders Modal with dialog role and title', () => {
    const onClose = vi.fn();
    const markup = renderToStaticMarkup(
      <Modal title="P08EC2091-1 · Bug" onClose={onClose}>
        <div data-testid="modal-body">Konten Modal Detail</div>
      </Modal>
    );

    expect(markup).toContain('role="dialog"');
    expect(markup).toContain('aria-modal="true"');
    expect(markup).toContain('P08EC2091-1 · Bug');
    expect(markup).toContain('Konten Modal Detail');
  });
});
