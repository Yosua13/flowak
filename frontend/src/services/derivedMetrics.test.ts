import { describe, expect, it } from 'vitest';
import { calculateDerivedMetrics } from './derivedMetrics';
import type { Module, WorkItem } from '../domain/types';

const module: Module = { id: 'm1', name: 'Main', nodes: [{ id: 'n1', type: 'process', label: 'Login', x: 0, y: 0, doc: {}, roles: { uiux: { readiness: 'done' }, frontend: { readiness: 'done' }, backend: { readiness: 'planned' } } }], edges: [], schemaVersion: 1 };
const item = (id: string, status: WorkItem['status'], due_date?: string): WorkItem => ({ id, key: id, project_id: 'p1', module_id: 'm1', node_id: 'n1', type: 'Task', title: id, priority: 'medium', status, reporter_id: 'u1', row_version: 1, created_at: '2026-01-01T00:00:00Z', due_date });

describe('calculateDerivedMetrics', () => {
  it('reconstructs cycle, blocked time, throughput, and rework from history', () => {
    const metrics = calculateDerivedMetrics([item('a', 'Done'), item('b', 'In Progress', '2026-01-02')], [
      { work_item_id: 'a', to_status: 'Backlog', created_at: '2026-01-01T00:00:00Z' },
      { work_item_id: 'a', from_status: 'Backlog', to_status: 'Blocked', created_at: '2026-01-01T02:00:00Z' },
      { work_item_id: 'a', from_status: 'Blocked', to_status: 'In Progress', created_at: '2026-01-01T05:00:00Z' },
      { work_item_id: 'a', from_status: 'In Progress', to_status: 'In Review', created_at: '2026-01-01T06:00:00Z' },
      { work_item_id: 'a', from_status: 'In Review', to_status: 'In Progress', created_at: '2026-01-01T07:00:00Z' },
      { work_item_id: 'a', from_status: 'In Progress', to_status: 'Done', created_at: '2026-01-01T10:00:00Z' },
    ], module, new Date('2026-01-03T00:00:00Z'));
    expect(metrics).toMatchObject({ throughput: 1, cycleTimeHours: 10, blockedTimeHours: 3, reviewRework: 1, overdue: 1 });
    expect(metrics.completeness).toBeCloseTo(2 / 3);
    expect(metrics.traceability).toBe(1);
  });
});
