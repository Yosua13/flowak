import type { ID, Module } from '../domain/types';
import { apiClient } from './apiClient';
import { invalidateProjectCache, queryClient, queryKeys } from './queryClient';

export interface ProjectItem {
  id: string; name: string; description: string; owner_id: string;
  status: 'active' | 'archived'; created_at: string;
}
interface ProjectResponse extends ProjectItem { modules: Array<Module & { project_id?: ID; nodes: Module['nodes'] | string; edges: Module['edges'] | string }>; }

export function normalizeModule(module: ProjectResponse['modules'][number]): Module {
  return {
    ...module,
    nodes: typeof module.nodes === 'string' ? JSON.parse(module.nodes) : module.nodes,
    edges: typeof module.edges === 'string' ? JSON.parse(module.edges) : module.edges,
  };
}

export const projectsService = {
  async list(status: 'active' | 'archived' = 'active', signal?: AbortSignal) {
    const result = await apiClient.get<ProjectItem[]>(`/projects${status === 'archived' ? '?status=archived' : ''}`, signal);
    queryClient.setQueryData([...queryKeys.projects, status], result);
    return result;
  },
  async detail(id: ID, signal?: AbortSignal) {
    const result = await apiClient.get<ProjectResponse>(`/projects/${id}`, signal);
    const normalized = { ...result, modules: result.modules.map(normalizeModule) };
    queryClient.setQueryData(queryKeys.project(id), normalized);
    return normalized;
  },
  async create(name: string, description: string) {
    const result = await apiClient.post<{ project_id: ID }>('/projects', { name, description });
    await invalidateProjectCache();
    return result;
  },
  async archive(id: ID) { await apiClient.delete(`/projects/${id}`); await invalidateProjectCache(id); },
  async restore(id: ID) { await apiClient.post(`/projects/${id}/restore`); await invalidateProjectCache(id); },
  async listModules(projectId: ID, signal?: AbortSignal) {
    return (await apiClient.get<ProjectResponse['modules']>(`/projects/${projectId}/modules`, signal)).map(normalizeModule);
  },
  async createModule(projectId: ID, module: Pick<Module, 'name' | 'description'> & Partial<Pick<Module, 'nodes' | 'edges'>>) {
    const result = await apiClient.post<{ module_id: ID }>(`/projects/${projectId}/modules`, module);
    await invalidateProjectCache(projectId);
    return result;
  },
  async renameModule(id: ID, name: string, description?: string) { await apiClient.put(`/modules/${id}`, { name, description }); },
  async deleteModule(id: ID) { await apiClient.delete(`/modules/${id}`); },
};
