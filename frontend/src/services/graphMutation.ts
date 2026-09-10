import type { Module } from '../domain/types';
import { ApiError, apiClient } from './apiClient';
import { invalidateProjectCache } from './queryClient';

export type SaveStatus = 'idle' | 'saving' | 'saved' | 'offline' | 'failed' | 'conflict';

export async function saveGraphOptimistically(module: Module, projectId: string | null, rollback: () => void, signal?: AbortSignal): Promise<SaveStatus> {
  try {
    await apiClient.put(`/modules/${module.id}`, { nodes: module.nodes, edges: module.edges }, signal);
    await invalidateProjectCache(projectId ?? undefined);
    return 'saved';
  } catch (error) {
    rollback();
    const normalized = error instanceof ApiError ? error : new ApiError('unknown_error', 'Gagal menyimpan graph.');
    if (normalized.code === 'conflict') return 'conflict';
    if (normalized.code === 'network_error') return 'offline';
    return 'failed';
  }
}
