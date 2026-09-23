import type { Module } from '../domain/types';
import { ApiError, apiClient } from './apiClient';
import { invalidateProjectCache } from './queryClient';

export type SaveStatus = 'idle' | 'saving' | 'saved' | 'offline' | 'failed' | 'conflict';

export interface GraphSaveResponse { nodes?: Module['nodes']; edges?: Module['edges']; }

export function saveGraph(module: Module, signal?: AbortSignal) {
  return apiClient.put<GraphSaveResponse>(`/modules/${module.id}`, {
    nodes: module.nodes,
    edges: module.edges,
    deletedNodes: module.deletedNodes || [],
    deletedEdges: module.deletedEdges || [],
  }, signal);
}

export async function saveGraphOptimistically(module: Module, projectId: string | null, rollback: () => void, signal?: AbortSignal): Promise<SaveStatus> {
  try {
    await saveGraph(module, signal);
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
