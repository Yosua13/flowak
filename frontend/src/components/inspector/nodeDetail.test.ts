import { describe, expect, it } from 'vitest';
import type { Module, Node, WorkItem } from '../../domain/types';
import { isOverdue, nextPaths, nodeCompleteness, nodeWorkItems } from './nodeDetail';

const node: Node = {
  id: 'node-a', type: 'process', label: 'Validasi pengajuan', x: 0, y: 0,
  doc: { actor: 'Admin', trigger: 'Pengajuan masuk', process: 'Validasi data', outcome: 'Pengajuan tervalidasi' },
  roles: { uiux: { status: 'done' }, frontend: { status: 'review' }, backend: { status: 'planned' } },
};

const workItem: WorkItem = {
  id: 'wi-1', key: 'FLOW-1', project_id: 'project-1', module_id: 'module-1', node_id: 'node-a', type: 'Task', title: 'Tambahkan validasi', priority: 'high', status: 'In Progress', reporter_id: 'user-1', row_version: 1, due_date: '2026-09-01',
};

describe('node detail helpers', () => {
  it('calculates readiness from documented business and facet fields', () => {
    expect(nodeCompleteness(node)).toBe(100);
    expect(nodeCompleteness({ ...node, completeness: 42 })).toBe(42);
  });

  it('lists only outgoing next paths for the selected node', () => {
    const module: Module = { id: 'module-1', name: 'Modul', nodes: [node, { ...node, id: 'node-b', label: 'Berikutnya' }], edges: [{ id: 'edge-1', from: 'node-a', to: 'node-b', label: 'valid' }, { id: 'edge-2', from: 'node-b', to: 'node-a' }], schemaVersion: 1 };
    expect(nextPaths(node, module.edges, module.nodes)).toEqual([{ id: 'node-b', label: 'Berikutnya', edgeLabel: 'valid' }]);
  });

  it('filters work items by node and detects overdue open work', () => {
    expect(nodeWorkItems([workItem, { ...workItem, id: 'wi-2', node_id: 'other' }], node.id)).toEqual([workItem]);
    expect(isOverdue(workItem, new Date('2026-09-14T00:00:00'))).toBe(true);
    expect(isOverdue({ ...workItem, status: 'Done' }, new Date('2026-09-14T00:00:00'))).toBe(false);
  });
});
