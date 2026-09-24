import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Module } from '../domain/types';
import { ApiError, apiClient } from './apiClient';
import { saveGraphOptimistically } from './graphMutation';
import { queryClient, queryKeys } from './queryClient';

const graph: Module = { id: 'module-1', name: 'Test', nodes: [], edges: [], schemaVersion: 1 };

describe('graph mutation', () => {
  beforeEach(() => { queryClient.clear(); vi.restoreAllMocks(); });

  it('invalidates the project and project-list cache after a save', async () => {
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries').mockResolvedValue();
    vi.spyOn(apiClient, 'put').mockResolvedValue({});
    await expect(saveGraphOptimistically(graph, 'project-1', vi.fn())).resolves.toBe('saved');
    expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.project('project-1') });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.projects });
  });

  it('rolls an optimistic graph edit back on server rejection', async () => {
    const rollback = vi.fn();
    vi.spyOn(apiClient, 'put').mockRejectedValue(new Error('rejected'));
    await expect(saveGraphOptimistically(graph, 'project-1', rollback)).resolves.toBe('failed');
    expect(rollback).toHaveBeenCalledOnce();
  });

  it('exposes conflict state and rolls back when the server returns 409', async () => {
    const rollback = vi.fn();
    vi.spyOn(apiClient, 'put').mockRejectedValue(new ApiError('conflict', 'Versi sudah berubah', 409));
    await expect(saveGraphOptimistically(graph, 'project-1', rollback)).resolves.toBe('conflict');
    expect(rollback).toHaveBeenCalledOnce();
  });

  it('sends tombstones with the graph save payload', async () => {
    const graphWithDeletes: Module = { ...graph, deletedNodes: [{ id: 'node-1', rowVersion: 2 }], deletedEdges: [{ id: 'edge-1', rowVersion: 3 }] };
    const put = vi.spyOn(apiClient, 'put').mockResolvedValue({});
    await saveGraphOptimistically(graphWithDeletes, 'project-1', vi.fn());
    expect(put).toHaveBeenCalledWith('/modules/module-1', expect.objectContaining({ deletedNodes: [{ id: 'node-1', rowVersion: 2 }], deletedEdges: [{ id: 'edge-1', rowVersion: 3 }] }), undefined);
  });
});
