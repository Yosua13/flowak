import { beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from './apiClient';
import { projectsService } from './projects';
import { queryClient, queryKeys } from './queryClient';

describe('project service boundary', () => {
  beforeEach(() => { queryClient.clear(); vi.restoreAllMocks(); });

  it('normalizes legacy JSON graph fields before they reach the UI adapter', async () => {
    vi.spyOn(apiClient, 'get').mockResolvedValue({ id: 'project-1', modules: [{ id: 'module-1', name: 'Main', schemaVersion: 1, nodes: '[{"id":"node-1"}]', edges: '[]' }] } as never);
    await expect(projectsService.detail('project-1')).resolves.toMatchObject({ modules: [{ id: 'module-1', nodes: [{ id: 'node-1' }], edges: [] }] });
    expect(queryClient.getQueryData(queryKeys.project('project-1'))).toMatchObject({ modules: [{ id: 'module-1' }] });
  });

  it('forwards cancellation signals when the active project changes', async () => {
    const signal = new AbortController().signal;
    const get = vi.spyOn(apiClient, 'get').mockResolvedValue({ id: 'project-1', modules: [] } as never);
    await projectsService.detail('project-1', signal);
    expect(get).toHaveBeenCalledWith('/projects/project-1', signal);
  });

  it('invalidates project cache after a server mutation', async () => {
    vi.spyOn(apiClient, 'post').mockResolvedValue({ project_id: 'project-2' } as never);
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries').mockResolvedValue();
    await projectsService.create('New project', 'Description');
    expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.projects });
  });
});
