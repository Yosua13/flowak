import { describe, expect, it } from 'vitest';
import type { WorkItem } from '../../domain/types';
import { columnMetrics, filterWorkItems, itemByKey } from './kanbanModel';

const base: WorkItem = { id: 'wi-1', key: 'FLOW-1', project_id: 'p-1', module_id: 'm-1', node_id: 'n-1', facet_key: 'backend', type: 'Bug', title: 'Perbaiki kontrak', priority: 'high', points: 5, status: 'Backlog', assignee_id: 'u-1', reporter_id: 'u-2', row_version: 1, due_date: '2026-09-10' };
const filters = { search: '', moduleId: 'all', assigneeId: 'all', type: 'all', facet: 'all', priority: 'all', includeCanceled: false };

describe('kanban model', () => {
  it('filters attributes and hides canceled work by default', () => {
    const items = [base, { ...base, id: 'wi-2', key: 'FLOW-2', status: 'Canceled' as const }, { ...base, id: 'wi-3', key: 'FLOW-3', type: 'Task' as const, title: 'Dokumentasi UI' }];
    expect(filterWorkItems(items, { ...filters, search: 'kontrak', type: 'Bug' }).map((item) => item.key)).toEqual(['FLOW-1']);
    expect(filterWorkItems(items, { ...filters, includeCanceled: true })).toHaveLength(3);
  });

  it('calculates count, points, and overdue cards per column', () => {
    expect(columnMetrics([base, { ...base, id: 'wi-2', key: 'FLOW-2', points: 3 }], 'Backlog', new Date('2026-09-14T00:00:00'))).toEqual({ count: 2, points: 8, overdue: 2 });
  });

  it('restores selection from a work item key', () => {
    expect(itemByKey([base], 'FLOW-1')).toEqual(base);
    expect(itemByKey([base], 'FLOW-404')).toBeUndefined();
  });
});
