import { afterEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from './apiClient';
import { workItemsApi } from './workItems';
import type { WorkItem } from '../domain/types';

const item = (id: string): WorkItem => ({ id, key: `FLOW-${id}`, project_id: 'project-1', type: 'Task', title: id, priority: 'medium', status: 'Backlog', reporter_id: 'user-1', row_version: 1 });

afterEach(() => vi.restoreAllMocks());

describe('work item paging', () => {
  it('loads every cursor page in order without duplicate cards', async () => {
    const get = vi.spyOn(apiClient, 'get').mockResolvedValueOnce({ items: [item('1'), item('2')], next_cursor: 'next', sort: 'newest' }).mockResolvedValueOnce({ items: [item('2'), item('3')], next_cursor: '', sort: 'newest' });
    const result = await workItemsApi.listAll('project-1');
    expect(result.map((entry) => entry.id)).toEqual(['1', '2', '3']);
    expect(get.mock.calls[1][0]).toContain('cursor=next');
  });

  it('rejects a repeated cursor instead of looping forever', async () => {
    vi.spyOn(apiClient, 'get').mockResolvedValue({ items: [item('1')], next_cursor: 'same', sort: 'newest' });
    await expect(workItemsApi.listAll('project-1')).rejects.toThrow('Cursor work item berulang');
  });
});
